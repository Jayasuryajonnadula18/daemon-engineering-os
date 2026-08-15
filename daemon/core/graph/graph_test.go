package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKnowledgeGraph_TraversalAndExports(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite store: %v", err)
	}
	defer store.Close()

	_ = store.AddNode("service", "web", "Web Frontend", nil)
	_ = store.AddNode("service", "orders-api", "Orders API", nil)
	_ = store.AddNode("database", "postgres", "PostgreSQL", nil)

	_ = store.AddEdge("service", "web", "service", "orders-api", "calls")
	_ = store.AddEdge("service", "orders-api", "database", "postgres", "reads")

	kg := NewKnowledgeGraph(store)

	// Test Downstream Impact of web
	impact, err := kg.FindDownstreamImpact("web")
	if err != nil {
		t.Fatalf("FindDownstreamImpact failed: %v", err)
	}
	if len(impact) < 2 {
		t.Fatalf("expected at least 2 impacted entities from web, got %v", impact)
	}

	// Test Upstream Dependencies of postgres
	deps, err := kg.FindUpstreamDependencies("postgres")
	if err != nil {
		t.Fatalf("FindUpstreamDependencies failed: %v", err)
	}
	if len(deps) < 2 {
		t.Fatalf("expected at least 2 upstream dependencies for postgres, got %v", deps)
	}

	// Test Export JSON
	if err := kg.ExportJSON(tmpDir); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".daemon", "graph.json")); err != nil {
		t.Fatalf("graph.json missing after export")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".daemon", "context.json")); err != nil {
		t.Fatalf("context.json missing after export")
	}
}
