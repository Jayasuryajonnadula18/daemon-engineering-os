package discovery

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoveryEngine_MultiStackAndFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested multi-stack project: Go api + Node web + Rust tool
	apiDir := filepath.Join(tmpDir, "api")
	webDir := filepath.Join(tmpDir, "web")
	toolDir := filepath.Join(tmpDir, "tool")
	_ = os.MkdirAll(apiDir, 0755)
	_ = os.MkdirAll(webDir, 0755)
	_ = os.MkdirAll(toolDir, 0755)

	_ = os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte("module api"), 0644)
	_ = os.WriteFile(filepath.Join(webDir, "package.json"), []byte(`{"name":"web","dependencies":{"next":"13.0.0"}}`), 0644)
	_ = os.WriteFile(filepath.Join(toolDir, "Cargo.toml"), []byte("[package]\nname=\"tool\""), 0644)

	de := NewDiscoveryEngine(nil)
	info, err := de.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !strings.Contains(info.Language, "Go") || !strings.Contains(info.Language, "JavaScript") || !strings.Contains(info.Language, "Rust") {
		t.Fatalf("expected multi-stack language detection for Go, JS, Rust, got: %s", info.Language)
	}
}

func TestDiscoveryEngine_UnknownCustomFallback(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Custom Project"), 0644)

	de := NewDiscoveryEngine(nil)
	info, err := de.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if info.Language != "Unknown / Custom" {
		t.Fatalf("expected unknown project to degrade gracefully to 'Unknown / Custom', got: %s", info.Language)
	}
}

func TestDiscoveryEngine_MonorepoDetection(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "pnpm-workspace.yaml"), []byte("packages:\n  - 'packages/*'"), 0644)

	de := NewDiscoveryEngine(nil)
	info, err := de.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !info.Monorepo {
		t.Fatalf("expected pnpm-workspace.yaml to flag monorepo = true")
	}
}
