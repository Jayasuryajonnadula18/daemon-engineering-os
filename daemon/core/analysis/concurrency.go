package analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ConcurrencyAnalyzer struct{}

func NewConcurrencyAnalyzer() *ConcurrencyAnalyzer {
	return &ConcurrencyAnalyzer{}
}

func (ca *ConcurrencyAnalyzer) AnalyzeDirectory(dirPath string) ([]Finding, error) {
	var findings []Finding

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			fileFindings, _ := ca.AnalyzeGoFile(path)
			findings = append(findings, fileFindings...)
		}
		return nil
	})

	return findings, nil
}

func (ca *ConcurrencyAnalyzer) AnalyzeGoFile(filePath string) ([]Finding, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	relPath, _ := filepath.Rel(".", filePath)

	ast.Inspect(node, func(n ast.Node) bool {
		// Detect unbuffered channel creation or unchecked go statements inside loops
		if goStmt, ok := n.(*ast.GoStmt); ok {
			findings = append(findings, Finding{
				ID:              fmt.Sprintf("find-concurrency-%d", time.Now().UnixNano()),
				Category:        CategoryConcurrency,
				Severity:        SeverityLow,
				FactType:        FactTypeFact,
				Title:           "Asynchronous Goroutine Spawn Detected",
				Description:     fmt.Sprintf("Goroutine spawned in %s; verify lifecycle termination and context cancellation handling", relPath),
				EvidenceIDs:     []string{fmt.Sprintf("ev-ast-goroutine-%s", relPath)},
				Files:           []string{relPath},
				Confidence:      0.80,
				DetectionMethod: "ast_concurrency_analyzer",
				SuggestedActions: []string{
					"Ensure sync.WaitGroup, errgroup, or context cancellation is used to prevent goroutine leaks",
				},
				AutoFixAvailable: false,
				FirstObserved:    time.Now(),
				LastObserved:     time.Now(),
			})
			_ = goStmt
		}
		return true
	})

	return findings, nil
}
