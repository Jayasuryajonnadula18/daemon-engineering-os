package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixSaveAndRollbackSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "sample_config.json")

	originalContent := `{"version": "1.0", "status": "original"}`
	modifiedContent := `{"version": "1.0", "status": "modified-by-fix"}`

	// 1. Write original file
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write original file: %v", err)
	}

	// 2. Save snapshot of original state
	backupMap := map[string]string{
		testFile: originalContent,
	}
	if err := SaveFixSnapshot("test-target", backupMap); err != nil {
		t.Fatalf("SaveFixSnapshot failed: %v", err)
	}

	// 3. Mutate file (simulating fix execution)
	if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to write modified file: %v", err)
	}

	// Verify file is mutated
	current, _ := os.ReadFile(testFile)
	if string(current) != modifiedContent {
		t.Fatalf("Expected file to be modified, got %s", string(current))
	}

	// 4. Perform rollback
	snap, err := RestoreFixSnapshot()
	if err != nil {
		t.Fatalf("RestoreFixSnapshot failed: %v", err)
	}

	if snap.Target != "test-target" {
		t.Errorf("Expected snapshot target 'test-target', got '%s'", snap.Target)
	}

	// 5. Verify file is 100% restored to original content
	restored, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restored) != originalContent {
		t.Fatalf("Rollback failed! Expected %s, got %s", originalContent, string(restored))
	}
}

func TestFixEngine_ThreeRealFixes_ApplyVerifyRollback(t *testing.T) {
	tempDir := t.TempDir()

	// Fix 1: Environment template fix
	envFile := filepath.Join(tempDir, ".env")
	exFile := filepath.Join(tempDir, ".env.example")
	_ = os.WriteFile(envFile, []byte("API_KEY=secret123\nPORT=8080"), 0644)

	// Fix 2: Dependency lockfile sync fix
	pkgFile := filepath.Join(tempDir, "package.json")
	pkgOriginal := `{"name":"demo","devDependencies":{"eslint":"7.0.0"}}`
	_ = os.WriteFile(pkgFile, []byte(pkgOriginal), 0644)

	// Fix 3: Broken symlink cleanup fix
	linkFile := filepath.Join(tempDir, "dead_link.txt")
	_ = os.Symlink("missing_target.txt", linkFile)

	// Execute Apply for all 3 fixes with pre-fix snapshot
	backupMap := map[string]string{
		envFile:  "API_KEY=secret123\nPORT=8080",
		pkgFile:  pkgOriginal,
		linkFile: "dead_link.txt (symlink)",
	}
	if err := SaveFixSnapshot("multi-fix-target", backupMap); err != nil {
		t.Fatalf("SaveFixSnapshot failed: %v", err)
	}

	// 1. Apply Fix 1: Write .env.example
	_ = os.WriteFile(exFile, []byte("API_KEY=\nPORT=\n"), 0644)

	// 2. Apply Fix 2: Update package.json
	pkgUpdated := `{"name":"demo","devDependencies":{"eslint":"^8.50.0"}}`
	_ = os.WriteFile(pkgFile, []byte(pkgUpdated), 0644)

	// 3. Apply Fix 3: Remove broken symlink
	_ = os.Remove(linkFile)

	// Verification Phase
	if _, err := os.Stat(exFile); err != nil {
		t.Fatalf("Fix 1 verification failed: .env.example missing")
	}
	pkgData, _ := os.ReadFile(pkgFile)
	if string(pkgData) != pkgUpdated {
		t.Fatalf("Fix 2 verification failed: package.json not updated")
	}
	if _, err := os.Stat(linkFile); !os.IsNotExist(err) {
		t.Fatalf("Fix 3 verification failed: broken symlink not removed")
	}

	// Execute Rollback
	snap, err := RestoreFixSnapshot()
	if err != nil || snap == nil {
		t.Fatalf("RestoreFixSnapshot failed: %v", err)
	}

	// Verify Rollback
	pkgRestored, _ := os.ReadFile(pkgFile)
	if string(pkgRestored) != pkgOriginal {
		t.Fatalf("Rollback failed for Fix 2: expected %s, got %s", pkgOriginal, string(pkgRestored))
	}
}

