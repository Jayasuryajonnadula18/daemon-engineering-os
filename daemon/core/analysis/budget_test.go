package analysis

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestFiles(t *testing.T, count int) string {
	dir := t.TempDir()
	for i := 0; i < count; i++ {
		path := filepath.Join(dir, "file_"+string(rune('a'+i))+".go")
		err := os.WriteFile(path, []byte("package main\nfunc main() {}\n"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestBudget_MaxFilesRespected(t *testing.T) {
	dir := createTestFiles(t, 20)
	
	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	budget := DefaultAnalysisBudget()
	budget.MaxFiles = 5
	pipeline.SetBudget(budget)

	ctx := context.Background()
	result, err := pipeline.RunAnalysis(ctx, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPartial {
		t.Errorf("expected status %s, got %s", StatusPartial, result.Status)
	}
	if result.AnalyzedFiles > 5 {
		t.Errorf("expected AnalyzedFiles <= 5, got %d", result.AnalyzedFiles)
	}
}

func TestBudget_MaxDurationRespected(t *testing.T) {
	dir := createTestFiles(t, 5)
	
	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	budget := DefaultAnalysisBudget()
	budget.MaxDuration = 1 * time.Nanosecond
	pipeline.SetBudget(budget)

	ctx := context.Background()
	result, err := pipeline.RunAnalysis(ctx, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPartial {
		t.Errorf("expected status %s, got %s", StatusPartial, result.Status)
	}
}

func TestBudget_NoCrashOnBudgetExceeded(t *testing.T) {
	dir := createTestFiles(t, 10)
	
	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	budget := DefaultAnalysisBudget()
	budget.MaxFiles = 2
	pipeline.SetBudget(budget)

	ctx := context.Background()
	result, err := pipeline.RunAnalysis(ctx, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify valid JSON structure
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestBudget_DefaultBudgetCompletes(t *testing.T) {
	dir := createTestFiles(t, 5)
	
	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	
	ctx := context.Background()
	result, err := pipeline.RunAnalysis(ctx, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s. Reason: %s", StatusCompleted, result.Status, result.StatusReason)
	}
}
