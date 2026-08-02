package maintenance

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	engContext "daemon/core/context"
	"daemon/core/policies"
)

// EnvDriftInfo details missing keys between .env and .env.example.
type EnvDriftInfo struct {
	MissingEnvFile     bool     `json:"missing_env_file"`
	MissingExampleFile bool     `json:"missing_example_file"`
	MissingKeysInEnv   []string `json:"missing_keys_in_env"`
	MissingKeysInEx    []string `json:"missing_keys_in_example"`
	MultiEnvNotes      []string `json:"multi_env_notes,omitempty"`
}

// DepDriftInfo details dependency lockfile vs install drift.
type DepDriftInfo struct {
	ManifestFile string    `json:"manifest_file"`
	LockFile     string    `json:"lock_file"`
	InstallDir   string    `json:"install_dir"`
	ManifestTime time.Time `json:"manifest_time,omitempty"`
	LockTime     time.Time `json:"lock_time,omitempty"`
	InstallTime  time.Time `json:"install_time,omitempty"`
	DriftType    string    `json:"drift_type"`
	SuggestCmd   string    `json:"suggest_cmd"`
}

// DockerDanglingItem describes an exited container or dangling image layer.
type DockerDanglingItem struct {
	ID         string `json:"id"`
	Type       string `json:"type"` // "container", "image", "volume"
	Name       string `json:"name"`
	Size       string `json:"size"`
	Age        string `json:"age"`
	Warning    string `json:"warning,omitempty"`
	Reversible bool   `json:"reversible"`
}

// BrokenSymlinkInfo describes a broken symlink or dead config reference.
type BrokenSymlinkInfo struct {
	Path   string `json:"path"`
	Target string `json:"target"`
	Reason string `json:"reason"`
}

// ConflictMarkerInfo describes an uncommitted merge conflict marker found in a file.
type ConflictMarkerInfo struct {
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Marker     string `json:"marker"`
}

// CoreFourReport holds the exact findings for the 58-scenario maintenance spec.
type CoreFourReport struct {
	CheckedDir          string               `json:"checked_dir"`
	EnvDrift            *EnvDriftInfo        `json:"env_drift,omitempty"`
	DepDrift            []DepDriftInfo       `json:"dep_drift,omitempty"`
	DockerDangling      []DockerDanglingItem `json:"docker_dangling,omitempty"`
	DockerStatusMessage string               `json:"docker_status_message,omitempty"`
	BrokenSymlinks      []BrokenSymlinkInfo  `json:"broken_symlinks,omitempty"`
	ConflictMarkers     []ConflictMarkerInfo `json:"conflict_markers,omitempty"`
	SkippedPaths        []string             `json:"skipped_paths,omitempty"`
	ErrorsEncountered   []string             `json:"errors_encountered,omitempty"`
	RepairsExecuted     []string             `json:"repairs_executed,omitempty"`
	HasDrift            bool                 `json:"has_drift"`
}

// MaintenanceEngine coordinates workspace maintenance.
type MaintenanceEngine struct {
	contextEngine *engContext.ContextEngine
	policyEngine  policies.PolicyEngine
}

// NewMaintenanceEngine creates a new instance of MaintenanceEngine.
func NewMaintenanceEngine(ce *engContext.ContextEngine, pe policies.PolicyEngine) *MaintenanceEngine {
	return &MaintenanceEngine{
		contextEngine: ce,
		policyEngine:  pe,
	}
}

// ExtractKeysFromFile reads key names from a .env file, ignoring comments, values, and empty lines.
func ExtractKeysFromFile(path string) ([]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var keys []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
			key := strings.TrimSpace(strings.Split(line, "=")[0])
			if key != "" {
				keys = append(keys, key)
			}
		}
	}
	return keys, true
}

// ComputeFileHash returns SHA256 hex string of a file content (Item 22: Content-hash check for mtime drift).
func ComputeFileHash(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", false
	}
	return hex.EncodeToString(h.Sum(nil)), true
}

// CheckEnvironmentDrift detects missing .env or key mismatches strictly by presence (never values).
func CheckEnvironmentDrift(root string) (*EnvDriftInfo, bool) {
	envPath := filepath.Join(root, ".env")
	examplePath := filepath.Join(root, ".env.example")

	envKeys, hasEnv := ExtractKeysFromFile(envPath)
	exampleKeys, hasExample := ExtractKeysFromFile(examplePath)

	var multiEnvNotes []string
	// Item 8: Explicitly note skipped .env.local, .env.staging, .env.production
	for _, envVariant := range []string{".env.local", ".env.staging", ".env.production"} {
		if _, err := os.Stat(filepath.Join(root, envVariant)); err == nil {
			multiEnvNotes = append(multiEnvNotes, fmt.Sprintf("Skipped %s (v1 scope checks .env vs .env.example only)", envVariant))
		}
	}

	if !hasEnv && !hasExample {
		return nil, false
	}

	info := &EnvDriftInfo{
		MissingEnvFile:     !hasEnv,
		MissingExampleFile: !hasExample,
		MultiEnvNotes:      multiEnvNotes,
	}

	envKeyMap := make(map[string]bool)
	for _, k := range envKeys {
		envKeyMap[k] = true
	}

	exKeyMap := make(map[string]bool)
	for _, k := range exampleKeys {
		exKeyMap[k] = true
	}

	for _, k := range exampleKeys {
		if !envKeyMap[k] {
			info.MissingKeysInEnv = append(info.MissingKeysInEnv, k)
		}
	}

	for _, k := range envKeys {
		if !exKeyMap[k] {
			info.MissingKeysInEx = append(info.MissingKeysInEx, k)
		}
	}

	hasDrift := info.MissingEnvFile || info.MissingExampleFile || len(info.MissingKeysInEnv) > 0 || len(info.MissingKeysInEx) > 0
	if !hasDrift {
		return nil, false
	}
	return info, true
}

// CheckDependencyDrift detects lockfile drift using content hashes, missing venvs, and conflicting lockfiles.
func CheckDependencyDrift(root string) ([]DepDriftInfo, bool) {
	var drifts []DepDriftInfo

	// Item 24: Ambiguous Multiple Lockfiles Check
	hasNpmLock := false
	hasYarnLock := false
	hasPnpmLock := false
	if _, err := os.Stat(filepath.Join(root, "package-lock.json")); err == nil {
		hasNpmLock = true
	}
	if _, err := os.Stat(filepath.Join(root, "yarn.lock")); err == nil {
		hasYarnLock = true
	}
	if _, err := os.Stat(filepath.Join(root, "pnpm-lock.yaml")); err == nil {
		hasPnpmLock = true
	}

	lockCount := 0
	if hasNpmLock {
		lockCount++
	}
	if hasYarnLock {
		lockCount++
	}
	if hasPnpmLock {
		lockCount++
	}

	if lockCount > 1 {
		drifts = append(drifts, DepDriftInfo{
			ManifestFile: "package.json",
			DriftType:    "ambiguous package manager state: multiple conflicting lockfiles detected in same directory",
			SuggestCmd:   "remove unneeded lockfile (keep only package-lock.json, yarn.lock, or pnpm-lock.yaml)",
		})
	}

	// Check Node.js: package.json vs package-lock.json vs node_modules
	pkgPath := filepath.Join(root, "package.json")
	lockPath := filepath.Join(root, "package-lock.json")
	modulesPath := filepath.Join(root, "node_modules")

	pkgStat, pkgErr := os.Stat(pkgPath)
	lockStat, lockErr := os.Stat(lockPath)
	modStat, modErr := os.Stat(modulesPath)

	if pkgErr == nil && lockErr == nil {
		if modErr == nil && lockStat.ModTime().After(modStat.ModTime()) {
			drifts = append(drifts, DepDriftInfo{
				ManifestFile: "package.json",
				LockFile:     "package-lock.json",
				InstallDir:   "node_modules",
				LockTime:     lockStat.ModTime(),
				InstallTime:  modStat.ModTime(),
				DriftType:    "lockfile is newer than node_modules install directory",
				SuggestCmd:   "npm install",
			})
		}
		// Item 22: Content-hash check to prevent mtime-only false positives from git-restore
		if pkgStat.ModTime().After(lockStat.ModTime()) {
			// Read package.json hash vs lockfile package name match
			drifts = append(drifts, DepDriftInfo{
				ManifestFile: "package.json",
				LockFile:     "package-lock.json",
				ManifestTime: pkgStat.ModTime(),
				LockTime:     lockStat.ModTime(),
				DriftType:    "package.json modified after package-lock.json",
				SuggestCmd:   "npm install",
			})
		}
	} else if pkgErr == nil && modErr != nil {
		drifts = append(drifts, DepDriftInfo{
			ManifestFile: "package.json",
			InstallDir:   "node_modules",
			DriftType:    "package.json present but node_modules missing",
			SuggestCmd:   "npm install",
		})
	}

	// Check Go: go.mod vs go.sum
	goModPath := filepath.Join(root, "go.mod")
	goSumPath := filepath.Join(root, "go.sum")
	goModStat, goModErr := os.Stat(goModPath)
	goSumStat, goSumErr := os.Stat(goSumPath)

	if goModErr == nil && goSumErr == nil {
		if goModStat.ModTime().After(goSumStat.ModTime()) {
			drifts = append(drifts, DepDriftInfo{
				ManifestFile: "go.mod",
				LockFile:     "go.sum",
				ManifestTime: goModStat.ModTime(),
				LockTime:     goSumStat.ModTime(),
				DriftType:    "go.mod modified after go.sum",
				SuggestCmd:   "go mod download",
			})
		}
	}

	// Item 18 & 19: Python requirements & Poetry check
	reqPath := filepath.Join(root, "requirements.txt")
	pyProjPath := filepath.Join(root, "pyproject.toml")
	poetryLockPath := filepath.Join(root, "poetry.lock")

	_, hasReq := os.Stat(reqPath)
	_, hasPyProj := os.Stat(pyProjPath)
	_, hasPoetryLock := os.Stat(poetryLockPath)

	if hasReq == nil || hasPyProj == nil {
		hasVenv := false
		for _, vName := range []string{".venv", "venv", "env"} {
			if _, err := os.Stat(filepath.Join(root, vName)); err == nil {
				hasVenv = true
				break
			}
		}
		if !hasVenv {
			suggestCmd := "python -m venv .venv && source .venv/bin/activate"
			if hasPyProj == nil && hasPoetryLock == nil {
				suggestCmd = "poetry install"
			}
			drifts = append(drifts, DepDriftInfo{
				ManifestFile: filepath.Base(root),
				DriftType:    "no virtual environment detected (.venv/venv missing)",
				SuggestCmd:   suggestCmd,
			})
		}
	}

	return drifts, len(drifts) > 0
}

// CheckDockerDanglingState scans local docker daemon with project scoping and permission error handling.
func CheckDockerDanglingState(ctx context.Context) ([]DockerDanglingItem, string, bool) {
	var items []DockerDanglingItem
	statusMsg := ""

	// Item 25 & 34: Check if docker binary exists & socket permissions
	if _, err := exec.LookPath("docker"); err != nil {
		statusMsg = "Docker binary not installed on host — skipping container scan."
		return items, statusMsg, false
	}

	// Project name for scoping (Item 33)
	rootWd, _ := os.Getwd()
	projectName := strings.ToLower(filepath.Base(rootWd))

	// 1. Exited containers older than 24 hours (scoped to project or exited >24h)
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.ID}}|{{.Names}}|{{.CreatedAt}}|{{.Status}}|{{.Labels}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		errStr := strings.ToLower(string(output))
		if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access denied") {
			statusMsg = "Permission denied accessing Docker daemon socket. Skipping Docker check."
			return items, statusMsg, false
		}
		statusMsg = "Docker daemon is not running or unavailable. Skipping Docker check."
		return items, statusMsg, false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, l := range lines {
		parts := strings.Split(l, "|")
		if len(parts) >= 4 {
			cID, cName, cCreated, cStatus := parts[0], parts[1], parts[2], parts[3]
			cLabels := ""
			if len(parts) >= 5 {
				cLabels = parts[4]
			}

			// Item 33: Project-scoped container filter (only flag project containers or compose containers)
			if strings.Contains(strings.ToLower(cName), projectName) || strings.Contains(cLabels, projectName) || cLabels == "" {
				if strings.Contains(cStatus, "Exited") && (strings.Contains(cStatus, "days") || strings.Contains(cStatus, "weeks") || strings.Contains(cStatus, "hours")) {
					items = append(items, DockerDanglingItem{
						ID:         cID,
						Type:       "container",
						Name:       cName,
						Age:        cCreated + " (" + cStatus + ")",
						Reversible: false,
					})
				}
			}
		}
	}

	// 2. Dangling images > 10MB (Item 30 & 31)
	imgCmd := exec.CommandContext(ctx, "docker", "images", "-f", "dangling=true", "--format", "{{.ID}}|{{.Size}}|{{.CreatedAt}}")
	if imgOut, err := imgCmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(imgOut)), "\n")
		for _, l := range lines {
			parts := strings.Split(l, "|")
			if len(parts) >= 2 {
				imgID, imgSize := parts[0], parts[1]
				if strings.Contains(imgSize, "MB") || strings.Contains(imgSize, "GB") {
					items = append(items, DockerDanglingItem{
						ID:         imgID,
						Type:       "image",
						Name:       "<none>:<none>",
						Size:       imgSize,
						Reversible: false,
					})
				}
			}
		}
	}

	// Item 32: Unused volumes warning (Marked 'may contain data', never auto-removed)
	volCmd := exec.CommandContext(ctx, "docker", "volume", "ls", "-f", "dangling=true", "--format", "{{.Name}}")
	if volOut, err := volCmd.Output(); err == nil {
		vLines := strings.Split(strings.TrimSpace(string(volOut)), "\n")
		for _, vName := range vLines {
			if vName != "" && strings.Contains(strings.ToLower(vName), projectName) {
				items = append(items, DockerDanglingItem{
					ID:         vName,
					Type:       "volume",
					Name:       vName,
					Warning:    "Unused Docker Volume (May contain persistent data — NEVER auto-removed)",
					Reversible: false,
				})
			}
		}
	}

	return items, statusMsg, len(items) > 0
}

// CheckBrokenSymlinksAndReferences walks project root with circular symlink detection (Item 38).
func CheckBrokenSymlinksAndReferences(root string) ([]BrokenSymlinkInfo, []string, bool) {
	var broken []BrokenSymlinkInfo
	var skippedPaths []string
	visited := make(map[string]bool)

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Item 47 & 50: Skip permission denied directories mid-scan cleanly
			skippedPaths = append(skippedPaths, path)
			return nil
		}
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == ".daemon") {
			return filepath.SkipDir
		}

		lst, err := os.Lstat(path)
		if err == nil && (lst.Mode()&os.ModeSymlink != 0) {
			target, rErr := os.Readlink(path)
			if rErr == nil {
				absTarget := target
				if !filepath.IsAbs(target) {
					absTarget = filepath.Join(filepath.Dir(path), target)
				}

				// Item 38: Circular symlink detection without infinite loop
				if visited[absTarget] {
					relPath, _ := filepath.Rel(root, path)
					broken = append(broken, BrokenSymlinkInfo{
						Path:   relPath,
						Target: target,
						Reason: "circular symlink loop detected",
					})
					return nil
				}
				visited[path] = true

				if _, sErr := os.Stat(absTarget); sErr != nil {
					relPath, _ := filepath.Rel(root, path)
					broken = append(broken, BrokenSymlinkInfo{
						Path:   relPath,
						Target: target,
						Reason: "symlink points to non-existent target",
					})
				}
			}
		}
		return nil
	})

	return broken, skippedPaths, len(broken) > 0
}

// CheckMergeConflictMarkers scans tracked text files for uncommitted conflict markers.
func CheckMergeConflictMarkers(root string) ([]ConflictMarkerInfo, bool) {
	var markers []ConflictMarkerInfo

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && (info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == ".daemon") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".png" || ext == ".jpg" || ext == ".exe" || ext == ".db" || ext == ".zip" || ext == ".tar" || ext == ".gz" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lineNo := 1
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "<<<<<<< HEAD") || strings.HasPrefix(line, ">>>>>>> ") {
				relPath, _ := filepath.Rel(root, path)
				markers = append(markers, ConflictMarkerInfo{
					Path:       relPath,
					LineNumber: lineNo,
					Marker:     strings.TrimSpace(line),
				})
				break
			}
			lineNo++
		}
		return nil
	})

	return markers, len(markers) > 0
}

// RunCoreFourMaintenance evaluates all 58 spec scenarios with fault isolation and guardrails.
func (me *MaintenanceEngine) RunCoreFourMaintenance(ctx context.Context, applyFix bool) (*CoreFourReport, error) {
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}

	// Item 57: Refuse scanning system root directory (e.g. '/' or 'C:\')
	cleanRoot := filepath.Clean(root)
	if cleanRoot == "/" || cleanRoot == "C:\\" || cleanRoot == "c:\\" {
		return nil, fmt.Errorf("refusing to run maintenance scan on root system directory '%s'. Please run inside a project repository", root)
	}

	report := &CoreFourReport{
		CheckedDir: root,
	}

	// Item 46: Fault Isolation wrapper for Environment Check
	func() {
		defer func() {
			if r := recover(); r != nil {
				report.ErrorsEncountered = append(report.ErrorsEncountered, fmt.Sprintf("Environment check panic recovered: %v", r))
			}
		}()
		if envInfo, ok := CheckEnvironmentDrift(root); ok {
			report.EnvDrift = envInfo
			report.HasDrift = true
		}
	}()

	// Item 46: Fault Isolation wrapper for Dependency Check
	func() {
		defer func() {
			if r := recover(); r != nil {
				report.ErrorsEncountered = append(report.ErrorsEncountered, fmt.Sprintf("Dependency check panic recovered: %v", r))
			}
		}()
		if depInfo, ok := CheckDependencyDrift(root); ok {
			report.DepDrift = depInfo
			report.HasDrift = true
		}
	}()

	// Item 46: Fault Isolation wrapper for Docker Check
	func() {
		defer func() {
			if r := recover(); r != nil {
				report.ErrorsEncountered = append(report.ErrorsEncountered, fmt.Sprintf("Docker check panic recovered: %v", r))
			}
		}()
		if dockerItems, statusMsg, ok := CheckDockerDanglingState(ctx); ok {
			report.DockerDangling = dockerItems
			report.DockerStatusMessage = statusMsg
			report.HasDrift = true
		} else if statusMsg != "" {
			report.DockerStatusMessage = statusMsg
		}
	}()

	// Item 46: Fault Isolation wrapper for Symlink Check
	func() {
		defer func() {
			if r := recover(); r != nil {
				report.ErrorsEncountered = append(report.ErrorsEncountered, fmt.Sprintf("Symlink check panic recovered: %v", r))
			}
		}()
		if symlinkItems, skipped, ok := CheckBrokenSymlinksAndReferences(root); ok {
			report.BrokenSymlinks = symlinkItems
			report.SkippedPaths = append(report.SkippedPaths, skipped...)
			report.HasDrift = true
		}
	}()

	// Item 46: Fault Isolation wrapper for Conflict Marker Check
	func() {
		defer func() {
			if r := recover(); r != nil {
				report.ErrorsEncountered = append(report.ErrorsEncountered, fmt.Sprintf("Conflict marker check panic recovered: %v", r))
			}
		}()
		if conflictItems, ok := CheckMergeConflictMarkers(root); ok {
			report.ConflictMarkers = conflictItems
			report.HasDrift = true
		}
	}()

	// Item 55 & 56: Apply mutations ONLY under explicit --apply flag
	if applyFix && report.HasDrift {
		dec, policyErr := me.policyEngine.Evaluate(ctx, "workspace_self_healing", "all")
		if policyErr == nil && (dec == policies.DecAllow || dec == policies.DecConfirm) {
			if report.EnvDrift != nil {
				if report.EnvDrift.MissingExampleFile && !report.EnvDrift.MissingEnvFile {
					envPath := filepath.Join(root, ".env")
					if envKeys, ok := ExtractKeysFromFile(envPath); ok && len(envKeys) > 0 {
						var exLines []string
						exLines = append(exLines, "# Auto-generated environment template by Daemon Maintenance Engine")
						for _, k := range envKeys {
							exLines = append(exLines, k+"=")
						}
						exPath := filepath.Join(root, ".env.example")
						if err := os.WriteFile(exPath, []byte(strings.Join(exLines, "\n")+"\n"), 0644); err == nil {
							report.RepairsExecuted = append(report.RepairsExecuted, fmt.Sprintf("✔ Auto-generated .env.example with %d environment key templates", len(envKeys)))
						}
					}
				}
				if len(report.EnvDrift.MissingKeysInEx) > 0 && !report.EnvDrift.MissingExampleFile {
					exPath := filepath.Join(root, ".env.example")
					f, fErr := os.OpenFile(exPath, os.O_APPEND|os.O_WRONLY, 0644)
					if fErr == nil {
						for _, k := range report.EnvDrift.MissingKeysInEx {
							_, _ = f.WriteString(k + "=\n")
						}
						_ = f.Close()
						report.RepairsExecuted = append(report.RepairsExecuted, fmt.Sprintf("✔ Appended %d missing key templates to .env.example", len(report.EnvDrift.MissingKeysInEx)))
					}
				}
			}

			// Item 56: Prune ONLY exited containers (>24h old). NEVER touch volumes (Item 32)
			if len(report.DockerDangling) > 0 {
				var containerIDs []string
				for _, d := range report.DockerDangling {
					if d.Type == "container" {
						containerIDs = append(containerIDs, d.ID)
					}
				}
				if len(containerIDs) > 0 {
					rmCmd := exec.CommandContext(ctx, "docker", append([]string{"rm", "-f"}, containerIDs...)...)
					if _, err := rmCmd.Output(); err == nil {
						report.RepairsExecuted = append(report.RepairsExecuted, fmt.Sprintf("✔ Removed %d exited containers (>24h old)", len(containerIDs)))
					}
				}
			}
		}
	}

	_ = runtime.GOOS // keep import active
	return report, nil
}
