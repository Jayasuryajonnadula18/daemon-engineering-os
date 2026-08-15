package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"daemon/core/graph"
)

func TestPipeline_ImpactedAnalysis(t *testing.T) {
	tempDir := t.TempDir()

	// Create a main.go that imports internal/utils
	mainCode := `package main
import "example.com/project/utils"
func main() {
	utils.DoSomething()
}`
	// Create utils.go
	utilsCode := `package utils
func DoSomething() {}`

	err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainCode), 0644)
	if err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	utilsDir := filepath.Join(tempDir, "utils")
	err = os.MkdirAll(utilsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create utils dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(utilsDir, "utils.go"), []byte(utilsCode), 0644)
	if err != nil {
		t.Fatalf("failed to write utils.go: %v", err)
	}

	// Setup Cache and Graph
	cache, err := NewAnalysisCache(":memory:")
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	store, err := graph.NewSQLiteStore(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()
	kg := graph.NewKnowledgeGraph(store)

	// Add downstream impact link in graph: utils.go -> main.go (main depends on utils)
	_ = store.AddNode("File", "utils.go", "utils.go", nil)
	_ = store.AddNode("File", "main.go", "main.go", nil)
	_ = store.AddEdge("File", "utils.go", "File", "main.go", "depends_on")

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	pipeline.SetCache(cache)
	pipeline.SetGraph(kg)
	pipeline.SetChangedOnly(true)

	// Initial run (empty cache) -> all should be analyzed
	res1, err := pipeline.RunAnalysis(context.Background(), tempDir, false)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	if res1.CacheMisses != 2 {
		t.Errorf("expected 2 cache misses, got %d", res1.CacheMisses)
	}

	// Modify only utils.go (which should trigger re-analysis of main.go because main.go imports/depends on it)
	if err := os.WriteFile(filepath.Join(utilsDir, "utils.go"), []byte("package utils\nfunc DoSomething() { /* changed */ }"), 0644); err != nil {
		t.Fatal(err)
	}

	res2, err := pipeline.RunAnalysis(context.Background(), tempDir, false)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	// Since we modified utils.go:
	// - utils.go is changed.
	// - main.go depends on utils (either by import detection or graph downstream edge).
	// Therefore, both should be re-analyzed (2 cache misses, 0 cache hits).
	if res2.CacheMisses != 2 {
		t.Errorf("expected 2 cache misses due to dependency propagation, got %d (hits: %d)", res2.CacheMisses, res2.CacheHits)
	}
}
