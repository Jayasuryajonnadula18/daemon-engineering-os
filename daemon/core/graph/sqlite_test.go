package graph

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func TestSQLiteStore_BasicNodesAndEdges(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "graph.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// 1. Add Node
	err = store.AddNode("service", "api-1", "API Gateway", map[string]string{"port": "8080"})
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	// 2. Add Secondary Node
	err = store.AddNode("database", "db-1", "PostgreSQL Core", map[string]string{"port": "5432"})
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	// 3. Add Edge
	err = store.AddEdge("service", "api-1", "database", "db-1", "depends_on")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// 4. Query All Nodes
	nodes, err := store.GetAllNodes()
	if err != nil {
		t.Fatalf("GetAllNodes failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}

	// 5. Query Edges
	edges, err := store.GetEdges()
	if err != nil {
		t.Fatalf("GetEdges failed: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(edges))
	}
}

func TestSQLiteStore_CircularEdgesAndOrphans(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "graph_edge.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Add circular edges (Service A -> Service B -> Service A)
	_ = store.AddNode("service", "srv-a", "Service A", nil)
	_ = store.AddNode("service", "srv-b", "Service B", nil)

	_ = store.AddEdge("service", "srv-a", "service", "srv-b", "calls")
	_ = store.AddEdge("service", "srv-b", "service", "srv-a", "calls")

	edges, err := store.GetEdges()
	if err != nil {
		t.Fatalf("GetEdges failed: %v", err)
	}
	if len(edges) != 2 {
		t.Errorf("Expected 2 circular edges, got %d", len(edges))
	}

	// Add orphan edge pointing to non-existent target node
	_ = store.AddEdge("service", "srv-a", "service", "non-existent-node", "depends_on")
	edges, _ = store.GetEdges()
	if len(edges) != 3 {
		t.Errorf("Expected 3 edges including orphan, got %d", len(edges))
	}
}

func TestSQLiteStore_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "graph_concurrent.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	var wg sync.WaitGroup
	workers := 10

	// Concurrent node writers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			nodeID := fmt.Sprintf("node-%d", workerID)
			_ = store.AddNode("service", nodeID, fmt.Sprintf("Worker Node %d", workerID), nil)
			_, _ = store.GetAllNodes()
		}(i)
	}

	wg.Wait()

	nodes, err := store.GetAllNodes()
	if err != nil {
		t.Fatalf("GetAllNodes after concurrent writes failed: %v", err)
	}
	if len(nodes) != workers {
		t.Errorf("Expected %d nodes after concurrent writes, got %d", workers, len(nodes))
	}
}
