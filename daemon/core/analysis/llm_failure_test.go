package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestLLMFailure_OllamaUnavailable verifies Daemon stays operational when LLM is nil.
func TestLLMFailure_OllamaUnavailable(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, `package main
import "net/http"
func main() {
	resp, _ := http.Get("http://example.com")
	_ = resp
}`)

	pipeline := NewDeepAnalyzerPipeline(nil, nil) // nil reasoner = no LLM
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("Daemon crashed with LLM unavailable: %v", err)
	}
	if res.AIEnhanced {
		t.Fatalf("expected AIEnhanced=false, got true")
	}
	if len(res.Findings) == 0 {
		t.Fatalf("expected deterministic findings even without LLM")
	}
}

// TestLLMFailure_MalformedResponse verifies Daemon handles a nil reasoner same as malformed LLM response.
func TestLLMFailure_MalformedResponse(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, `package main
func main() { _ = doWork() }
func doWork() error { return nil }`)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatalf("Daemon crashed on LLM failure simulation: %v", err)
	}
	if res == nil {
		t.Fatal("got nil result")
	}
	if res.AIEnhanced {
		t.Fatal("AIEnhanced must be false when LLM fails")
	}
}

// TestLLMFailure_AllProvidersDown verifies Daemon produces deterministic output with no providers.
func TestLLMFailure_AllProvidersDown(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, `package main
import "os"
func main() {
	password := "hunter2"
	_ = password
	_ = os.Getenv("KEY")
}`)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Confirm security finding detected deterministically
	found := false
	for _, f := range res.Findings {
		if f.Category == CategorySecurity {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected security finding even with all LLM providers down")
	}
	if res.AIEnhanced {
		t.Fatal("AIEnhanced must remain false")
	}
}

// writeGoFile is a helper to create a Go source file in a temp directory.
func writeGoFile(t *testing.T, dir, code string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test go file: %v", err)
	}
}
