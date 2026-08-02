package maintenance

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	engContext "daemon/core/context"
	"daemon/core/domain"
	"daemon/core/policies"
	"daemon/core/storage"
)

type CompleteSpecDummyGraphStore struct {
	storage.GraphStore
}

func (d *CompleteSpecDummyGraphStore) GetServices() ([]domain.Service, error) {
	return nil, nil
}
func (d *CompleteSpecDummyGraphStore) GetDependencies() ([]domain.Dependency, error) {
	return nil, nil
}

type CompleteSpecDummyMemoryStore struct {
	storage.MemoryStore
}

func (d *CompleteSpecDummyMemoryStore) GetIncidents() ([]domain.Incident, error) {
	return nil, nil
}
func (d *CompleteSpecDummyMemoryStore) GetRecommendations() ([]domain.Recommendation, error) {
	return nil, nil
}
func (d *CompleteSpecDummyMemoryStore) GetDeployments() ([]domain.Deployment, error) {
	return nil, nil
}

func helperNewEngine() *MaintenanceEngine {
	ce := engContext.NewContextEngine(&CompleteSpecDummyGraphStore{}, &CompleteSpecDummyMemoryStore{})
	pe := policies.NewMemoryPolicyEngine(false)
	return NewMaintenanceEngine(ce, pe)
}

// -----------------------------------------------------------------------------
// SECTION 1: .ENV CHECK SCENARIOS (1.1 - 1.12)
// -----------------------------------------------------------------------------

func TestSpec_1_1_EnvExampleExistsEnvMissing(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=5000\nDB_HOST=localhost\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if !ok || info == nil || !info.MissingEnvFile || len(info.MissingKeysInEnv) != 2 {
		t.Fatalf("1.1 failed: expected MissingEnvFile = true with 2 missing keys")
	}
}

func TestSpec_1_2_EnvMatchesExample(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\nDB_HOST=localhost\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=\nDB_HOST=\n"), 0644)

	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.2 failed: expected silent pass when .env matches .env.example")
	}
}

func TestSpec_1_3_EnvMissingKeysOnly(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("KEY1=a\nKEY2=b\nKEY3=c\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("KEY1=\nKEY2=\nKEY3=\nKEY4=\nKEY5=\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if !ok || info == nil || len(info.MissingKeysInEnv) != 2 {
		t.Fatalf("1.3 failed: expected 2 missing keys (KEY4, KEY5)")
	}
}

func TestSpec_1_4_EnvHasExtraKeys(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("KEY1=a\nKEY2=b\nEXTRA=secret\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("KEY1=\nKEY2=\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if ok && info != nil && len(info.MissingKeysInEnv) > 0 {
		t.Errorf("1.4 failed: extra keys in .env should not trigger missing keys error")
	}
}

func TestSpec_1_5_NoExampleFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if ok && info != nil && len(info.MissingKeysInEnv) > 0 {
		t.Errorf("1.5 failed: missing .env.example should not guess keys")
	}
}

func TestSpec_1_6_ExampleFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte(""), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if ok && info != nil && len(info.MissingKeysInEnv) > 0 {
		t.Errorf("1.6 failed: empty .env.example expects 0 missing keys")
	}
}

func TestSpec_1_7_EnvEmptyExampleHasKeys(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("K1=\nK2=\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if !ok || info == nil || len(info.MissingKeysInEnv) != 2 {
		t.Fatalf("1.7 failed: empty .env with example containing 2 keys must flag all 2 missing")
	}
}

func TestSpec_1_8_MultipleEnvFilesSkippedNote(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("K1=1\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.local"), []byte("K1=local\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("K1=\n"), 0644)

	// .env.local skipped gracefully in v1
	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.8 failed: .env matches .env.example, .env.local should be skipped without false positive")
	}
}

func TestSpec_1_9_MonorepoSubfolderIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	apiDir := filepath.Join(tmpDir, "apps", "api")
	webDir := filepath.Join(tmpDir, "apps", "web")
	_ = os.MkdirAll(apiDir, 0755)
	_ = os.MkdirAll(webDir, 0755)

	_ = os.WriteFile(filepath.Join(apiDir, ".env.example"), []byte("API_KEY=\n"), 0644)
	_ = os.WriteFile(filepath.Join(webDir, ".env.example"), []byte("WEB_KEY=\n"), 0644)

	apiInfo, apiOk := CheckEnvironmentDrift(apiDir)
	webInfo, webOk := CheckEnvironmentDrift(webDir)

	if !apiOk || apiInfo == nil || len(apiInfo.MissingKeysInEnv) != 1 || apiInfo.MissingKeysInEnv[0] != "API_KEY" {
		t.Errorf("1.9 failed: subfolder api must be checked independently")
	}
	if !webOk || webInfo == nil || len(webInfo.MissingKeysInEnv) != 1 || webInfo.MissingKeysInEnv[0] != "WEB_KEY" {
		t.Errorf("1.9 failed: subfolder web must be checked independently")
	}
}

func TestSpec_1_10_EnvSyntaxErrorHandled(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("NORMAL_KEY=val\nMALFORMED_LINE_WITHOUT_EQUALS\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("NORMAL_KEY=\n"), 0644)

	// Malformed line without equals is ignored safely without crashing
	keys, ok := ExtractKeysFromFile(filepath.Join(tmpDir, ".env"))
	if !ok || len(keys) != 1 || keys[0] != "NORMAL_KEY" {
		t.Fatalf("1.10 failed: malformed line should be skipped cleanly, parsed %v", keys)
	}
}

func TestSpec_1_11_InlineCommentParsing(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("STRIPE_KEY=sk_test_123 # comment\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("STRIPE_KEY= # comment\n"), 0644)

	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.11 failed: inline comments must be stripped during key extraction")
	}
}

func TestSpec_1_12_ValueDifferencesIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=8080\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=3000\n"), 0644)

	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.12 failed: value differences must be explicitly ignored")
	}
}

// -----------------------------------------------------------------------------
// SECTION 2: DEPENDENCY DRIFT CHECK SCENARIOS (2.1 - 2.12)
// -----------------------------------------------------------------------------

func TestSpec_2_1_LockfileMatchesModules(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	lockPath := filepath.Join(tmpDir, "package-lock.json")
	modPath := filepath.Join(tmpDir, "node_modules")

	_ = os.WriteFile(pkgPath, []byte("{}"), 0644)
	_ = os.WriteFile(lockPath, []byte("{}"), 0644)
	_ = os.Mkdir(modPath, 0755)

	now := time.Now()
	_ = os.Chtimes(pkgPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(lockPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(modPath, now, now)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if ok && len(drifts) > 0 {
		t.Errorf("2.1 failed: matching lockfile and node_modules must not flag drift")
	}
}

func TestSpec_2_2_LockfileNewerThanModules(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	lockPath := filepath.Join(tmpDir, "package-lock.json")
	modPath := filepath.Join(tmpDir, "node_modules")

	_ = os.WriteFile(pkgPath, []byte("{}"), 0644)
	_ = os.WriteFile(lockPath, []byte("{}"), 0644)
	_ = os.Mkdir(modPath, 0755)

	now := time.Now()
	_ = os.Chtimes(modPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(lockPath, now, now)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 || drifts[0].SuggestCmd != "npm install" {
		t.Fatalf("2.2 failed: lockfile newer than install dir must suggest npm install")
	}
}

func TestSpec_2_3_ModulesMissingEntirely(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 {
		t.Fatalf("2.3 failed: missing node_modules must flag drift")
	}
}

func TestSpec_2_4_PackageJsonChangedLockfileNotRegenerated(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	lockPath := filepath.Join(tmpDir, "package-lock.json")

	_ = os.WriteFile(pkgPath, []byte("{}"), 0644)
	_ = os.WriteFile(lockPath, []byte("{}"), 0644)

	now := time.Now()
	_ = os.Chtimes(lockPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(pkgPath, now, now)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 {
		t.Fatalf("2.4 failed: package.json newer than lockfile must flag lockfile out of sync")
	}
}

func TestSpec_2_5_GoModNewerThanSum(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")
	goSumPath := filepath.Join(tmpDir, "go.sum")

	_ = os.WriteFile(goModPath, []byte("module test"), 0644)
	_ = os.WriteFile(goSumPath, []byte(""), 0644)

	now := time.Now()
	_ = os.Chtimes(goSumPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(goModPath, now, now)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 || drifts[0].SuggestCmd != "go mod download" {
		t.Fatalf("2.5 failed: go.mod newer than go.sum must suggest go mod download")
	}
}

func TestSpec_2_9_NoPackageManagerFiles(t *testing.T) {
	tmpDir := t.TempDir()

	drifts, ok := CheckDependencyDrift(tmpDir)
	if ok || len(drifts) > 0 {
		t.Errorf("2.9 failed: static project without package files must not flag dependency drift")
	}
}

// -----------------------------------------------------------------------------
// SECTION 3: DANGLING DOCKER STATE (3.1 - 3.10)
// -----------------------------------------------------------------------------

func TestSpec_3_1_DockerNotInstalledGracefulSkip(t *testing.T) {
	ctx := context.Background()
	// CheckDockerDanglingState executes local docker command and gracefully returns empty list if daemon unavailable
	items, _, _ := CheckDockerDanglingState(ctx)
	// Must never panic or crash
	if items == nil {
		items = []DockerDanglingItem{}
	}
}

// -----------------------------------------------------------------------------
// SECTION 4: BROKEN SYMLINKS & DEAD REFERENCES (4.1 - 4.9)
// -----------------------------------------------------------------------------

func TestSpec_4_1_ValidSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	_ = os.WriteFile(targetFile, []byte("hello"), 0644)
	_ = os.Symlink("target.txt", linkFile)

	broken, _, ok := CheckBrokenSymlinksAndReferences(tmpDir)
	if ok || len(broken) > 0 {
		t.Errorf("4.1 failed: valid symlink must not be flagged")
	}
}

func TestSpec_4_2_BrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	_ = os.WriteFile(targetFile, []byte("hello"), 0644)
	_ = os.Symlink("target.txt", linkFile)
	_ = os.Remove(targetFile)

	broken, _, ok := CheckBrokenSymlinksAndReferences(tmpDir)
	if !ok || len(broken) == 0 || broken[0].Target != "target.txt" {
		t.Fatalf("4.2 failed: broken symlink pointing to missing target must be flagged")
	}
}

func TestSpec_4_7_NodeModulesSymlinkExcluded(t *testing.T) {
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "node_modules", "pkg")
	_ = os.MkdirAll(modDir, 0755)
	_ = os.Symlink("missing.txt", filepath.Join(modDir, "broken.link"))

	broken, _, ok := CheckBrokenSymlinksAndReferences(tmpDir)
	if ok && len(broken) > 0 {
		for _, b := range broken {
			if strings.Contains(b.Path, "node_modules") {
				t.Errorf("4.7 failed: node_modules symlinks must be excluded from scan")
			}
		}
	}
}

func TestSpec_4_8_GitDirectorySymlinkExcluded(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0755)
	_ = os.Symlink("missing.txt", filepath.Join(gitDir, "broken.link"))

	broken, _, ok := CheckBrokenSymlinksAndReferences(tmpDir)
	if ok && len(broken) > 0 {
		for _, b := range broken {
			if strings.Contains(b.Path, ".git") {
				t.Errorf("4.8 failed: .git directory symlinks must be excluded from scan")
			}
		}
	}
}

// -----------------------------------------------------------------------------
// SECTION 5 & 6 & 7 & 8: CROSS-CUTTING, EVIDENCE, JSON & GUARDRAILS (5.1 - 8.4)
// -----------------------------------------------------------------------------

func TestSpec_5_1_RunTwiceIdenticalOutput(t *testing.T) {
	ctx := context.Background()
	me := helperNewEngine()

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("K1=\n"), 0644)

	rep1, _ := me.RunCoreFourMaintenance(ctx, false)
	rep2, _ := me.RunCoreFourMaintenance(ctx, false)

	if rep1.HasDrift != rep2.HasDrift || len(rep1.EnvDrift.MissingKeysInEnv) != len(rep2.EnvDrift.MissingKeysInEnv) {
		t.Errorf("5.1 failed: idempotency failure between consecutive runs")
	}
}

func TestSpec_6_4_EmptyDirectoryClean(t *testing.T) {
	ctx := context.Background()
	me := helperNewEngine()

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	rep, err := me.RunCoreFourMaintenance(ctx, false)
	if err != nil {
		t.Fatalf("6.4 failed: unexpected error in empty directory: %v", err)
	}
	if rep.HasDrift {
		t.Errorf("6.4 failed: empty directory must report HasDrift = false")
	}
}

func TestSpec_7_4_JsonOutputSchema(t *testing.T) {
	ctx := context.Background()
	me := helperNewEngine()

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("K1=\n"), 0644)

	rep, _ := me.RunCoreFourMaintenance(ctx, false)
	data, err := json.Marshal(rep)
	if err != nil || len(data) == 0 {
		t.Fatalf("7.4 failed: json schema output serialization failed")
	}
	if !strings.Contains(string(data), "\"has_drift\":true") {
		t.Errorf("7.4 failed: json payload missing has_drift field")
	}
}

func TestSpec_8_1_ZeroWritesWithoutApply(t *testing.T) {
	ctx := context.Background()
	me := helperNewEngine()

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)

	rep, _ := me.RunCoreFourMaintenance(ctx, false)
	if len(rep.RepairsExecuted) > 0 {
		t.Errorf("8.1 failed: zero writes guardrail violated without --apply flag!")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".env.example")); err == nil {
		t.Errorf("8.1 failed: .env.example must NOT be created without --apply flag")
	}
}

func TestSpec_8_2_ApplyOnlyTouchesFlagged(t *testing.T) {
	ctx := context.Background()
	me := helperNewEngine()

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)

	rep, _ := me.RunCoreFourMaintenance(ctx, true)
	if len(rep.RepairsExecuted) != 1 {
		t.Errorf("8.2 failed: --apply should only repair flagged item (.env.example auto-gen)")
	}
}
