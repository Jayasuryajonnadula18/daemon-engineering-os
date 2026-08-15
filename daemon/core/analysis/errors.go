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

var correctnessSeq uint64

type CorrectnessAnalyzer struct{}

func NewCorrectnessAnalyzer() *CorrectnessAnalyzer {
	return &CorrectnessAnalyzer{}
}

func (ca *CorrectnessAnalyzer) AnalyzeDirectory(dirPath string) ([]Finding, error) {
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

func (ca *CorrectnessAnalyzer) AnalyzeGoFile(filePath string) ([]Finding, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	relPath, _ := filepath.Rel(".", filePath)

	ast.Inspect(node, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
					findings = append(findings, Finding{
						ID:               fmt.Sprintf("find-ignored-err-%d", atomic.AddUint64(&correctnessSeq, 1)),
						Category:         CategoryErrorHandling,
						Severity:         SeverityMedium,
						FactType:         FactTypeFact,
						Title:            "Ignored Function Return Value / Error",
						Description:      fmt.Sprintf("Return value ignored using blank identifier '_' in %s", relPath),
						EvidenceIDs:      []string{fmt.Sprintf("ev-ast-ignored-%s", relPath)},
						AffectedFiles:    []string{relPath},
						Confidence:       0.80,
						DetectionMethod:  "ast_static_analysis",
						SuggestedActions: []string{
							"Check return error explicitly instead of swallowing",
						},
						AutoFixAvailable: false,
						FirstObserved:    time.Now(),
						LastObserved:     time.Now(),
					})
					break
				}
			}
		}
		return true
	})

	return findings, nil
}
