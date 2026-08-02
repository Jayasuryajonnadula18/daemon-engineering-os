package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"daemon/core/graph"
)

func TestDiscoveryScan_EmptyRepo(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_graph.db")
	gs, err := graph.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteStore: %v", err)
	}
	defer gs.Close()

	de := NewDiscoveryEngine(gs)
	info, err := de.Scan(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("Scan failed on empty directory: %v", err)
	}

	if info == nil {
		t.Fatalf("Expected non-nil ProjectInfo")
	}

	if len(info.Dependencies) != 0 {
		t.Errorf("Expected 0 dependencies for empty directory, got %d", len(info.Dependencies))
	}
}

func TestDiscoveryScan_MonorepoAndManifests(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_graph.db")
	gs, err := graph.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteStore: %v", err)
	}
	defer gs.Close()

	// Create nested monorepo structure
	apiDir := filepath.Join(tempDir, "services", "api")
	webDir := filepath.Join(tempDir, "services", "web")
	ignoredDir := filepath.Join(tempDir, "node_modules", "some-pkg")

	_ = os.MkdirAll(apiDir, 0755)
	_ = os.MkdirAll(webDir, 0755)
	_ = os.MkdirAll(ignoredDir, 0755)

	// Add package.json to apiDir
	pkgJsonContent := `{
		"name": "api-service",
		"dependencies": {
			"express": "^4.18.2",
			"cors": "^2.8.5"
		},
		"devDependencies": {
			"typescript": "^5.0.0"
		}
	}`
	_ = os.WriteFile(filepath.Join(apiDir, "package.json"), []byte(pkgJsonContent), 0644)

	// Add docker-compose.yml at root
	dockerComposeContent := "version: '3.8'\nservices:\n  api:\n    build: .\n"
	_ = os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(dockerComposeContent), 0644)

	// Add .env file
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("PORT=5000\n"), 0644)

	// Add ignored package.json in node_modules
	_ = os.WriteFile(filepath.Join(ignoredDir, "package.json"), []byte(`{"name":"ignored"}`), 0644)

	de := NewDiscoveryEngine(gs)
	info, err := de.Scan(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("Scan failed on monorepo: %v", err)
	}

	if !info.DockerCompose {
		t.Errorf("Expected DockerCompose to be true")
	}

	if len(info.EnvFiles) == 0 {
		t.Errorf("Expected .env file to be detected")
	}

	// Ensure node_modules was skipped
	for _, dep := range info.Dependencies {
		if dep.Name == "ignored" {
			t.Errorf("Discovery engine should skip node_modules directory!")
		}
	}
}

func TestDiscoveryScan_CorruptManifest(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_graph.db")
	gs, err := graph.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteStore: %v", err)
	}
	defer gs.Close()

	// Write invalid/malformed JSON in package.json
	corruptContent := `{ "name": "bad-json", "dependencies": { "express": `
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(corruptContent), 0644)

	de := NewDiscoveryEngine(gs)
	// Scan should not panic or fail catastrophically on malformed manifests
	info, err := de.Scan(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("Scan should gracefully handle corrupt manifest, got error: %v", err)
	}

	if info == nil {
		t.Fatalf("Expected non-nil ProjectInfo even with corrupt manifest")
	}
}
