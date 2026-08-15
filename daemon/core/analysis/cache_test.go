package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestProject(t *testing.T) string {
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCache_MemoryDB(t *testing.T) {
	cache, err := NewAnalysisCache(":memory:")
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer cache.Close()
	stats, err := cache.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries, got %d", stats.TotalEntries)
	}
}

func TestCache_UnchangedFileIsSkipped(t *testing.T) {
	dir := setupTestProject(t)
	cache, _ := NewAnalysisCache(":memory:")
	defer cache.Close()

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	pipeline.SetCache(cache)

	res1, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if res1.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss on first run, got %d", res1.CacheMisses)
	}
	if res1.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits on first run, got %d", res1.CacheHits)
	}

	res2, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if res2.CacheMisses != 0 {
		t.Errorf("Expected 0 cache misses on second run, got %d", res2.CacheMisses)
	}
	if res2.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit on second run, got %d", res2.CacheHits)
	}
}

func TestCache_ChangedFileIsReanalyzed(t *testing.T) {
	dir := setupTestProject(t)
	cache, _ := NewAnalysisCache(":memory:")
	defer cache.Close()

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	pipeline.SetCache(cache)

	// First run
	pipeline.RunAnalysis(context.Background(), dir, false)

	// Modify file
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte("package main\n\nfunc main() { fmt.Println(1) }\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Second run
	res2, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if res2.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss after file change, got %d", res2.CacheMisses)
	}
	if res2.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits after file change, got %d", res2.CacheHits)
	}
}

func TestCache_NewAnalyzerVersionInvalidates(t *testing.T) {
	dir := setupTestProject(t)
	cache, _ := NewAnalysisCache(":memory:")
	defer cache.Close()

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	pipeline.SetCache(cache)

	// First run
	pipeline.RunAnalysis(context.Background(), dir, false)

	// Manually update the DB to simulate old version
	file := filepath.Join(dir, "main.go")
	
	_, currentHash, _, _ := cache.IsStale(file, AnalyzerVersion, "dep-v1", "twin-v1")
	
	err := cache.SetEntry(CacheEntry{
		FilePath:        file,
		FileHash:        currentHash,
		AnalyzerVersion: "0.9.0", // old version
		DependencyState: "dep-v1",
		TwinVersion:     "twin-v1",
		AnalyzedAt:      time.Now(),
		Stale:           false,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Second run
	res2, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if res2.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss due to version change, got %d", res2.CacheMisses)
	}
}

func TestCache_DependencyStateChangeInvalidates(t *testing.T) {
	dir := setupTestProject(t)
	cache, _ := NewAnalysisCache(":memory:")
	defer cache.Close()

	// First query: verify we can manually trigger invalidation by passing different dep states
	file := filepath.Join(dir, "main.go")

	// Set initial cached state
	_ = cache.SetEntry(CacheEntry{
		FilePath:        file,
		FileHash:        "somehash",
		AnalyzerVersion: AnalyzerVersion,
		DependencyState: "dep-v1",
		TwinVersion:     "twin-v1",
		AnalyzedAt:      time.Now(),
		Stale:           false,
	})

	// Check with different dependency state
	stale, _, _, _ := cache.IsStale(file, AnalyzerVersion, "dep-v2", "twin-v1")
	if !stale {
		t.Errorf("Expected stale=true due to dependency change")
	}

	// Check with different twin version
	staleTwin, _, _, _ := cache.IsStale(file, AnalyzerVersion, "dep-v1", "twin-v2")
	if !staleTwin {
		t.Errorf("Expected stale=true due to twin version change")
	}
}
