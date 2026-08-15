package maintenance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDependencyDrift_IgnoresMtimeOnlyChangesWhenContentIsUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	lockPath := filepath.Join(tmpDir, "package-lock.json")

	if err := os.WriteFile(pkgPath, []byte(`{"name":"demo"}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(lockPath, []byte(`{"name":"demo"}`), 0o644); err != nil {
		t.Fatalf("write package-lock.json: %v", err)
	}

	if drifts, ok := CheckDependencyDrift(tmpDir); ok || len(drifts) != 0 {
		t.Fatalf("expected first pass to be clean, got drift=%v ok=%v", drifts, ok)
	}

	now := time.Now()
	if err := os.Chtimes(pkgPath, now, now); err != nil {
		t.Fatalf("touch package.json: %v", err)
	}
	if err := os.Chtimes(lockPath, now.Add(-5*time.Minute), now.Add(-5*time.Minute)); err != nil {
		t.Fatalf("touch package-lock.json: %v", err)
	}

	if drifts, ok := CheckDependencyDrift(tmpDir); ok || len(drifts) != 0 {
		t.Fatalf("expected no drift when only mtimes changed and content stayed identical, got drift=%v ok=%v", drifts, ok)
	}
}

func TestDependencyDrift_GitCheckoutLockfile_ZeroFalsePositives(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	lockPath := filepath.Join(tmpDir, "package-lock.json")

	content := []byte(`{"name":"test-repo","version":"1.0.0"}`)
	if err := os.WriteFile(pkgPath, content, 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(lockPath, content, 0o644); err != nil {
		t.Fatalf("write package-lock.json: %v", err)
	}

	// 1. Initial pass records hash state
	drifts, hasDrift := CheckDependencyDrift(tmpDir)
	if hasDrift || len(drifts) > 0 {
		t.Fatalf("expected initial pass clean, got drifts=%v", drifts)
	}

	// 2. Simulate git checkout / git restore on package-lock.json (mtime changes to now, content identical)
	now := time.Now().Add(10 * time.Minute)
	if err := os.Chtimes(lockPath, now, now); err != nil {
		t.Fatalf("touch lockfile: %v", err)
	}
	// Rewrite exact same content to simulate git checkout file replacement
	if err := os.WriteFile(lockPath, content, 0o644); err != nil {
		t.Fatalf("rewrite lockfile: %v", err)
	}

	// 3. Verify zero false positive drift results
	drifts, hasDrift = CheckDependencyDrift(tmpDir)
	if hasDrift || len(drifts) > 0 {
		t.Fatalf("git checkout lockfile false positive detected: drifts=%v", drifts)
	}
}

