package analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var lifetimeSeq uint64

type ResourceLifetimeAnalyzer struct{}

func NewResourceLifetimeAnalyzer() *ResourceLifetimeAnalyzer {
	return &ResourceLifetimeAnalyzer{}
}

func (rla *ResourceLifetimeAnalyzer) AnalyzeDirectory(dirPath string) ([]Finding, error) {
	var findings []Finding

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			fileFindings, _ := rla.AnalyzeGoFile(path)
			findings = append(findings, fileFindings...)
		}
		return nil
	})

	return findings, nil
}

func (rla *ResourceLifetimeAnalyzer) AnalyzeGoFile(filePath string) ([]Finding, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	relPath, _ := filepath.Rel(".", filePath)

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Get" || sel.Sel.Name == "Do" || sel.Sel.Name == "Post" {
				if ident, ok := sel.X.(*ast.Ident); ok && (ident.Name == "http" || strings.Contains(ident.Name, "client")) {
					findings = append(findings, Finding{
						ID:              fmt.Sprintf("find-http-body-%d", atomic.AddUint64(&lifetimeSeq, 1)),
						Category:        CategoryResourceLifecycle,
						Severity:        SeverityHigh,
						FactType:        FactTypeFact,
						Title:           "Potential Unclosed HTTP Response Body",
						Description:     fmt.Sprintf("HTTP %s call in %s may not close resp.Body, causing socket/goroutine leaks", sel.Sel.Name, relPath),
						EvidenceIDs:     []string{fmt.Sprintf("ev-ast-http-%s", relPath)},
						AffectedFiles:   []string{relPath},
						Confidence:      0.85,
						DetectionMethod: "ast_static_analysis",
						SuggestedActions: []string{
							"Add defer resp.Body.Close() immediately after error check",
						},
						AutoFixAvailable: true,
						FirstObserved:    time.Now(),
						LastObserved:     time.Now(),
					})
				}
			}
		}
		return true
	})

	return findings, nil
}
