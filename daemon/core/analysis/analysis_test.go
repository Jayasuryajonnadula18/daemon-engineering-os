package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDeepAnalyzerPipeline_LLMUnavailableDeterministicRun(t *testing.T) {
	tempDir := t.TempDir()

	// Create test Go file with unclosed HTTP body, ignored error, goroutine spawn, and hardcoded credential
	code := `package main
import (
	"net/http"
	"os"
)
func main() {
	password := "secret123=" // Security trigger
	_ = password
	go func() {}()           // Concurrency trigger
	resp, err := http.Get("http://localhost:8080") // Resource leak trigger
	if err != nil { return }
	_ = os.Getenv("PORT")    // Ignored error trigger
	_ = resp
}`

	err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(code), 0644)
	if err != nil {
		t.Fatalf("failed to create test go file: %v", err)
	}

	pipeline := NewDeepAnalyzerPipeline(nil, nil) // LLM reasoner is nil (unavailable)
	res, err := pipeline.RunAnalysis(context.Background(), tempDir, true)
	if err != nil {
		t.Fatalf("unexpected error during analysis: %v", err)
	}

	if res.AIEnhanced {
		t.Fatalf("expected AIEnhanced to be false when LLM is unavailable")
	}

	if len(res.Findings) == 0 {
		t.Fatalf("expected deterministic findings across analyzers, got 0")
	}

	categories := make(map[FindingCategory]bool)
	for _, f := range res.Findings {
		categories[f.Category] = true
	}

	if !categories[CategoryResourceLifecycle] {
		t.Fatalf("expected resource lifecycle finding")
	}
	if !categories[CategoryErrorHandling] {
		t.Fatalf("expected error handling finding")
	}
	if !categories[CategoryConcurrency] {
		t.Fatalf("expected concurrency finding")
	}
	if !categories[CategorySecurity] {
		t.Fatalf("expected security finding")
	}
}
