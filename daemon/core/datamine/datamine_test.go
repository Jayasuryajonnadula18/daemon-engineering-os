package datamine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestClassifyCommitSubject_NoTestingFallback(t *testing.T) {
	// Unmatched commit subject must return empty string (never fallback to "testing")
	cat := ClassifyCommitSubject("xyz123 unusual commit message without keywords")
	if cat != "" {
		t.Fatalf("expected unmatched commit to return empty category, got %s", cat)
	}
}

func TestMineCommits_CategoryDistributionAndExcludedLogging(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo with sample commits
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", tmpDir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Tester",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Tester",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		if err := cmd.Run(); err != nil {
			t.Fatalf("git error (%v): %v", args, err)
		}
	}

	run("init")
	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("1"), 0644)
	run("add", ".")
	run("commit", "-m", "feat: add initial feature module")

	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("2"), 0644)
	run("commit", "-am", "fix: repair broken null check in parser")

	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("3"), 0644)
	run("commit", "-am", "refactor: restructure internal graph solver")

	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("4"), 0644)
	run("commit", "-am", "docs: update security policy and architecture spec")

	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("5"), 0644)
	run("commit", "-am", "chore: bump dependencies")

	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("6"), 0644)
	run("commit", "-am", "wip standalone experiment string")

	report, err := MineCommits(context.Background(), tmpDir, 10)
	if err != nil {
		t.Fatalf("mine commits failed: %v", err)
	}

	if report.ExcludedCount == 0 || len(report.ExcludedLog) == 0 {
		t.Fatalf("expected non-zero excluded count and log, got count=%d", report.ExcludedCount)
	}

	if report.CategoryCounts[CategoryFeature] != 1 || report.CategoryCounts[CategoryFix] != 1 {
		t.Fatalf("unexpected category count distribution: %+v", report.CategoryCounts)
	}

	if !report.IsBalancedSample {
		t.Fatalf("expected balanced category distribution, got %+v", report.CategoryCounts)
	}
}
