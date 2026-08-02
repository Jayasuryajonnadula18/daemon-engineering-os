package maintenance

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	engContext "daemon/core/context"
	"daemon/core/domain"
	"daemon/core/policies"
	"daemon/core/storage"
)

type SpecDummyGraphStore struct {
	storage.GraphStore
}

func (d *SpecDummyGraphStore) GetServices() ([]domain.Service, error) {
	return nil, nil
}
func (d *SpecDummyGraphStore) GetDependencies() ([]domain.Dependency, error) {
	return nil, nil
}

type SpecDummyMemoryStore struct {
	storage.MemoryStore
}

func (d *SpecDummyMemoryStore) GetIncidents() ([]domain.Incident, error) {
	return nil, nil
}
func (d *SpecDummyMemoryStore) GetRecommendations() ([]domain.Recommendation, error) {
	return nil, nil
}
func (d *SpecDummyMemoryStore) GetDeployments() ([]domain.Deployment, error) {
	return nil, nil
}

// -----------------------------------------------------------------------------
// 1. .ENV CHECK SCENARIOS (1.1 - 1.12)
// -----------------------------------------------------------------------------

func TestEnvCheck_1_1_ExampleExistsEnvMissing(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=5000\nDB_HOST=localhost\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if !ok || info == nil {
		t.Fatalf("1.1: expected env drift flagged when .env missing")
	}
	if !info.MissingEnvFile {
		t.Errorf("1.1: expected MissingEnvFile = true")
	}
	if len(info.MissingKeysInEnv) != 2 {
		t.Errorf("1.1: expected 2 missing keys, got %d", len(info.MissingKeysInEnv))
	}
}

func TestEnvCheck_1_2_EnvMatchesExample(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\nDB_HOST=localhost\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=\nDB_HOST=\n"), 0644)

	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.2: expected silent pass when all keys present")
	}
}

func TestEnvCheck_1_3_EnvMissingKeysOnly(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("KEY1=a\nKEY2=b\nKEY3=c\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("KEY1=\nKEY2=\nKEY3=\nKEY4=\nKEY5=\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if !ok || info == nil {
		t.Fatalf("1.3: expected drift when 2 keys missing")
	}
	if len(info.MissingKeysInEnv) != 2 {
		t.Errorf("1.3: expected 2 missing keys (KEY4, KEY5), got %d (%v)", len(info.MissingKeysInEnv), info.MissingKeysInEnv)
	}
}

func TestEnvCheck_1_4_EnvHasExtraKeys(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("KEY1=a\nKEY2=b\nEXTRA_KEY=secret\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("KEY1=\nKEY2=\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if !ok {
		// Extra keys in .env do not trigger missing keys in .env
		if info != nil && len(info.MissingKeysInEnv) > 0 {
			t.Errorf("1.4: extra keys in .env should not trigger missing keys in env")
		}
	}
}

func TestEnvCheck_1_5_NoExampleFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if ok && info != nil && info.MissingKeysInEnv != nil {
		t.Errorf("1.5: no .env.example should not guess missing keys")
	}
}

func TestEnvCheck_1_6_ExampleFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte(""), 0644)

	info, ok := CheckEnvironmentDrift(tmpDir)
	if ok && info != nil && len(info.MissingKeysInEnv) > 0 {
		t.Errorf("1.6: empty .env.example expects 0 keys")
	}
}

func TestEnvCheck_1_11_InlineCommentParsing(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("STRIPE_KEY=sk_test_123 # inline comment\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("STRIPE_KEY= # stripe key comment\n"), 0644)

	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.11: STRIPE_KEY should match correctly ignoring comments")
	}
}

func TestEnvCheck_1_12_ValueDifferencesIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=8080\nNODE_ENV=production\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=3000\nNODE_ENV=development\n"), 0644)

	_, ok := CheckEnvironmentDrift(tmpDir)
	if ok {
		t.Errorf("1.12: value differences must be explicitly out of scope and ignored")
	}
}

// -----------------------------------------------------------------------------
// 2. DEPENDENCY DRIFT CHECK SCENARIOS (2.1 - 2.12)
// -----------------------------------------------------------------------------

func TestDepCheck_2_2_LockfileNewerThanModules(t *testing.T) {
	tmpDir := t.TempDir()

	pkgPath := filepath.Join(tmpDir, "package.json")
	lockPath := filepath.Join(tmpDir, "package-lock.json")
	modPath := filepath.Join(tmpDir, "node_modules")

	_ = os.WriteFile(pkgPath, []byte("{}"), 0644)
	_ = os.WriteFile(lockPath, []byte("{}"), 0644)
	_ = os.Mkdir(modPath, 0755)

	// Set lockfile mtime newer than node_modules
	now := time.Now()
	_ = os.Chtimes(modPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(lockPath, now, now)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 {
		t.Fatalf("2.2: expected drift when lockfile newer than node_modules")
	}
	if drifts[0].SuggestCmd != "npm install" {
		t.Errorf("2.2: expected suggest cmd 'npm install', got %s", drifts[0].SuggestCmd)
	}
}

func TestDepCheck_2_3_ModulesMissingEntirely(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 {
		t.Fatalf("2.3: expected drift when package.json present but node_modules missing")
	}
}

func TestDepCheck_2_5_GoModNewerThanSum(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")
	goSumPath := filepath.Join(tmpDir, "go.sum")

	_ = os.WriteFile(goModPath, []byte("module test"), 0644)
	_ = os.WriteFile(goSumPath, []byte(""), 0644)

	now := time.Now()
	_ = os.Chtimes(goSumPath, now.Add(-10*time.Minute), now.Add(-10*time.Minute))
	_ = os.Chtimes(goModPath, now, now)

	drifts, ok := CheckDependencyDrift(tmpDir)
	if !ok || len(drifts) == 0 {
		t.Fatalf("2.5: expected drift when go.mod newer than go.sum")
	}
	if drifts[0].SuggestCmd != "go mod download" {
		t.Errorf("2.5: expected suggest cmd 'go mod download', got %s", drifts[0].SuggestCmd)
	}
}

// -----------------------------------------------------------------------------
// 4. BROKEN SYMLINKS SCENARIOS (4.1 - 4.9)
// -----------------------------------------------------------------------------

func TestBrokenSymlinks_4_1_ValidSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	_ = os.WriteFile(targetFile, []byte("hello"), 0644)
	_ = os.Symlink("target.txt", linkFile)

	broken, ok := CheckBrokenSymlinksAndReferences(tmpDir)
	if ok || len(broken) > 0 {
		t.Errorf("4.1: valid symlink should not be flagged as broken")
	}
}

func TestBrokenSymlinks_4_2_BrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	_ = os.WriteFile(targetFile, []byte("hello"), 0644)
	_ = os.Symlink("target.txt", linkFile)
	_ = os.Remove(targetFile) // target is deleted

	broken, ok := CheckBrokenSymlinksAndReferences(tmpDir)
	if !ok || len(broken) == 0 {
		t.Fatalf("4.2: broken symlink pointing to deleted target must be flagged")
	}
	if broken[0].Target != "target.txt" {
		t.Errorf("4.2: expected target target.txt, got %s", broken[0].Target)
	}
}

// -----------------------------------------------------------------------------
// 5 & 8. CROSS-CUTTING IDEMPOTENCY & GUARDRAIL SCENARIOS
// -----------------------------------------------------------------------------

func TestCrossCutting_5_1_Idempotency(t *testing.T) {
	ctx := context.Background()
	ce := engContext.NewContextEngine(&SpecDummyGraphStore{}, &SpecDummyMemoryStore{})
	pe := policies.NewMemoryPolicyEngine(false)
	me := NewMaintenanceEngine(ce, pe)

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("PORT=5000\n"), 0644)

	rep1, _ := me.RunCoreFourMaintenance(ctx, false)
	rep2, _ := me.RunCoreFourMaintenance(ctx, false)

	if rep1.HasDrift != rep2.HasDrift {
		t.Errorf("5.1: idempotency failure — HasDrift mismatch between consecutive runs")
	}
}

func TestCrossCutting_8_1_ZeroWritesWithoutApply(t *testing.T) {
	ctx := context.Background()
	ce := engContext.NewContextEngine(&SpecDummyGraphStore{}, &SpecDummyMemoryStore{})
	pe := policies.NewMemoryPolicyEngine(false)
	me := NewMaintenanceEngine(ce, pe)

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\n"), 0644)

	// Run without apply
	rep, _ := me.RunCoreFourMaintenance(ctx, false)
	if len(rep.RepairsExecuted) > 0 {
		t.Errorf("8.1: zero writes guardrail violated — repairs executed without --apply flag!")
	}

	// Verify .env.example was NOT created
	if _, err := os.Stat(filepath.Join(tmpDir, ".env.example")); err == nil {
		t.Errorf("8.1: .env.example should NOT be created without --apply flag")
	}
}
