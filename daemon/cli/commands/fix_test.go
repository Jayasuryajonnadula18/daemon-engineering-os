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
