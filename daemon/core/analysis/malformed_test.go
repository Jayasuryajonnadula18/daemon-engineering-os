package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestMalformedSource_SyntaxError verifies a syntax error in one file does not crash the pipeline.
func TestMalformedSource_SyntaxError(t *testing.T) {
	dir := t.TempDir()

	// Valid file
	_ = os.WriteFile(filepath.Join(dir, "good.go"), []byte(`package main
func main() {}
`), 0644)

	// Intentionally malformed file — missing closing brace
	_ = os.WriteFile(filepath.Join(dir, "bad.go"), []byte(`package main
func broken( {
	this is not valid go
`), 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("pipeline must not crash on syntax error: %v", err)
	}
	if res == nil {
		t.Fatal("got nil result")
	}
	// Pipeline must still return a valid result object
	if res.Timestamp.IsZero() {
		t.Fatal("result timestamp must not be zero")
	}
}

// TestMalformedSource_BinaryFile verifies a binary file does not crash the analysis pipeline.
func TestMalformedSource_BinaryFile(t *testing.T) {
	dir := t.TempDir()

	// Binary data masquerading as .go file
	binary := []byte{0x00, 0x01, 0xFF, 0xFE, 0xDE, 0xAD, 0xBE, 0xEF}
	_ = os.WriteFile(filepath.Join(dir, "binary.go"), binary, 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("pipeline must not crash on binary .go file: %v", err)
	}
	if res == nil {
		t.Fatal("got nil result")
	}
}

// TestMalformedSource_EmptyFile verifies an empty file is handled gracefully.
func TestMalformedSource_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "empty.go"), []byte{}, 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatalf("pipeline must not crash on empty file: %v", err)
	}
	if res == nil {
		t.Fatal("got nil result")
	}
}

// TestMalformedSource_HugeFile verifies a very large file does not cause OOM or crash.
func TestMalformedSource_HugeFile(t *testing.T) {
	dir := t.TempDir()

	// Generate a large but parseable Go file
	large := []byte("package main\n\nfunc main() {\n")
	for i := 0; i < 10000; i++ {
		large = append(large, []byte("\t_ = 0\n")...)
	}
	large = append(large, '}')
	_ = os.WriteFile(filepath.Join(dir, "large.go"), large, 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatalf("pipeline must not crash on large file: %v", err)
	}
	if res == nil {
		t.Fatal("got nil result")
	}
}

// TestMalformedSource_InvalidUnicode verifies unusual Unicode in source does not crash analyzers.
func TestMalformedSource_InvalidUnicode(t *testing.T) {
	dir := t.TempDir()

	// Go file with unusual Unicode in comments/strings
	code := `package main
// Ñoño Ångström 日本語 العربية — valid Unicode in comment
func main() {
	s := "日本語"
	_ = s
}`
	_ = os.WriteFile(filepath.Join(dir, "unicode.go"), []byte(code), 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, false)
	if err != nil {
		t.Fatalf("pipeline must not crash on Unicode source: %v", err)
	}
	if res == nil {
		t.Fatal("got nil result")
	}
}
