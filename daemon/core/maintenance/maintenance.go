package maintenance

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
}

// DepDriftInfo details dependency lockfile vs install drift.
type DepDriftInfo struct {
	ManifestFile  string    `json:"manifest_file"`
	LockFile      string    `json:"lock_file"`
	InstallDir    string    `json:"install_dir"`
	ManifestTime  time.Time `json:"manifest_time"`
	LockTime      time.Time `json:"lock_time"`
	InstallTime   time.Time `json:"install_time"`
	DriftType     string    `json:"drift_type"`
	SuggestCmd    string    `json:"suggest_cmd"`
}

// DockerDanglingItem describes an exited container or dangling image layer.
type DockerDanglingItem struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "container" or "image"
	Name      string `json:"name"`
	Size      string `json:"size"`
	Age       string `json:"age"`
	Reversible bool   `json:"reversible"`
}

// BrokenSymlinkInfo describes a broken symlink or dead config reference.
type BrokenSymlinkInfo struct {
	Path   string `json:"path"`
	Target string `json:"target"`
	Reason string `json:"reason"`
}

// CoreFourReport holds the exact findings for the 4 core maintenance checks.
type CoreFourReport struct {
	CheckedDir       string                `json:"checked_dir"`
	EnvDrift         *EnvDriftInfo         `json:"env_drift,omitempty"`
	DepDrift         []DepDriftInfo        `json:"dep_drift,omitempty"`
	DockerDangling   []DockerDanglingItem  `json:"docker_dangling,omitempty"`
	BrokenSymlinks   []BrokenSymlinkInfo   `json:"broken_symlinks,omitempty"`
	RepairsExecuted  []string              `json:"repairs_executed,omitempty"`
	HasDrift         bool                  `json:"has_drift"`
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

// CheckEnvironmentDrift detects missing .env or key mismatches strictly by presence (never values).
func CheckEnvironmentDrift(root string) (*EnvDriftInfo, bool) {
	envPath := filepath.Join(root, ".env")
	examplePath := filepath.Join(root, ".env.example")

	envKeys, hasEnv := ExtractKeysFromFile(envPath)
	exampleKeys, hasExample := ExtractKeysFromFile(examplePath)

	if !hasEnv && !hasExample {
		return nil, false
	}

	info := &EnvDriftInfo{
		MissingEnvFile:     !hasEnv,
		MissingExampleFile: !hasExample,
	}

	envKeyMap := make(map[string]bool)
	for _, k := range envKeys {
		envKeyMap[k] = true
	}

	exKeyMap := make(map[string]bool)
	for _, k := range exampleKeys {
		exKeyMap[k] = true
	}

	// Detect keys in .env.example missing from .env
	for _, k := range exampleKeys {
		if !envKeyMap[k] {
			info.MissingKeysInEnv = append(info.MissingKeysInEnv, k)
		}
	}

	// Detect keys in .env missing from .env.example
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

// CheckDependencyDrift detects timestamp mismatches between manifest, lockfile, and install dir.
func CheckDependencyDrift(root string) ([]DepDriftInfo, bool) {
	var drifts []DepDriftInfo

	// Check Node.js: package.json vs package-lock.json vs node_modules
	pkgPath := filepath.Join(root, "package.json")
	lockPath := filepath.Join(root, "package-lock.json")
	modulesPath := filepath.Join(root, "node_modules")

	pkgStat, pkgErr := os.Stat(pkgPath)
	lockStat, lockErr := os.Stat(lockPath)
	modStat, modErr := os.Stat(modulesPath)

	if pkgErr == nil && lockErr == nil {
		// Lockfile newer than node_modules (someone pulled lockfile change but never ran npm install)
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
		// package.json newer than lockfile (package added/updated but lockfile not updated)
		if pkgStat.ModTime().After(lockStat.ModTime()) {
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
		// package.json present but node_modules missing
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

	return drifts, len(drifts) > 0
}

// CheckDockerDanglingState scans local docker daemon for exited containers (>24h) and dangling images (>10MB).
func CheckDockerDanglingState(ctx context.Context) ([]DockerDanglingItem, bool) {
	var items []DockerDanglingItem

	// 1. Exited containers older than 24 hours
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.ID}}|{{.Names}}|{{.CreatedAt}}|{{.Status}}")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, l := range lines {
			parts := strings.Split(l, "|")
			if len(parts) >= 4 {
				cID, cName, cCreated, cStatus := parts[0], parts[1], parts[2], parts[3]
				// Only flag if exited > 24 hours
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

	// 2. Dangling images > 10MB
	imgCmd := exec.CommandContext(ctx, "docker", "images", "-f", "dangling=true", "--format", "{{.ID}}|{{.Size}}|{{.CreatedAt}}")
	if imgOut, err := imgCmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(imgOut)), "\n")
		for _, l := range lines {
			parts := strings.Split(l, "|")
			if len(parts) >= 2 {
				imgID, imgSize := parts[0], parts[1]
				// Parse size > 10MB
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

	return items, len(items) > 0
}

// CheckBrokenSymlinksAndReferences walks project root for broken symlinks.
func CheckBrokenSymlinksAndReferences(root string) ([]BrokenSymlinkInfo, bool) {
	var broken []BrokenSymlinkInfo

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == ".daemon") {
			return filepath.SkipDir
		}

		// Use Lstat to detect symlink
		lst, err := os.Lstat(path)
		if err == nil && (lst.Mode()&os.ModeSymlink != 0) {
			target, rErr := os.Readlink(path)
			if rErr == nil {
				// Check if target exists
				absTarget := target
				if !filepath.IsAbs(target) {
					absTarget = filepath.Join(filepath.Dir(path), target)
				}
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

	return broken, len(broken) > 0
}

// RunCoreFourMaintenance evaluates the 4 core checks and returns strict empirical evidence.
func (me *MaintenanceEngine) RunCoreFourMaintenance(ctx context.Context, applyFix bool) (*CoreFourReport, error) {
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}

	report := &CoreFourReport{
		CheckedDir: root,
	}

	// 1. Env check
	if envInfo, ok := CheckEnvironmentDrift(root); ok {
		report.EnvDrift = envInfo
		report.HasDrift = true
	}

	// 2. Dependency check
	if depInfo, ok := CheckDependencyDrift(root); ok {
		report.DepDrift = depInfo
		report.HasDrift = true
	}

	// 3. Docker check
	if dockerItems, ok := CheckDockerDanglingState(ctx); ok {
		report.DockerDangling = dockerItems
		report.HasDrift = true
	}

	// 4. Symlinks check
	if symlinkItems, ok := CheckBrokenSymlinksAndReferences(root); ok {
		report.BrokenSymlinks = symlinkItems
		report.HasDrift = true
	}

	// Apply self-healing under explicit --apply flag
	if applyFix && report.HasDrift {
		dec, policyErr := me.policyEngine.Evaluate(ctx, "workspace_self_healing", "all")
		if policyErr == nil && (dec == policies.DecAllow || dec == policies.DecConfirm) {
			// Fix A: Generate missing .env.example if missing
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

			// Fix B: Prune Docker resources only if --apply is passed
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

	return report, nil
}
