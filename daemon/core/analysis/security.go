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

type SecurityAnalyzer struct{}

func NewSecurityAnalyzer() *SecurityAnalyzer {
	return &SecurityAnalyzer{}
}

func (sa *SecurityAnalyzer) AnalyzeDirectory(dirPath string) ([]Finding, error) {
	var findings []Finding

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			fileFindings, _ := sa.AnalyzeGoFile(path)
			findings = append(findings, fileFindings...)
		}
		return nil
	})

	return findings, nil
}

func (sa *SecurityAnalyzer) AnalyzeGoFile(filePath string) ([]Finding, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	relPath, _ := filepath.Rel(".", filePath)

	secretKeywords := []string{"password", "token", "secret", "private_key", "api_key"}

	ast.Inspect(node, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					nameLower := strings.ToLower(ident.Name)
					for _, kw := range secretKeywords {
						if strings.Contains(nameLower, kw) {
							findings = append(findings, Finding{
								ID:              fmt.Sprintf("find-sec-%d", time.Now().UnixNano()),
								Category:        CategorySecurity,
								Severity:        SeverityHigh,
								FactType:        FactTypeFact,
								Title:           "Potential Hardcoded Secret / Credential",
								Description:     fmt.Sprintf("Hardcoded credential variable '%s' matching pattern '%s' detected in %s", ident.Name, kw, relPath),
								EvidenceIDs:     []string{fmt.Sprintf("ev-sec-%s", relPath)},
								AffectedFiles:   []string{relPath},
								Confidence:      0.85,
								DetectionMethod: "ast_security_analyzer",
								SuggestedActions: []string{
									"Move sensitive credentials out of source code into environment variables or secret store",
								},
								AutoFixAvailable: true,
								FirstObserved:    time.Now(),
								LastObserved:     time.Now(),
							})
							break
						}
					}
				}
			}
		}
		return true
	})

	return findings, nil
}
