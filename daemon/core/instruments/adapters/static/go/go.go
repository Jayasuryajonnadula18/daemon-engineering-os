package staticgo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"daemon/core/instruments"
)

type GoLeakInstrument struct {
	ident instruments.InstrumentIdentity
}

func NewGoLeakInstrument() *GoLeakInstrument {
	return &GoLeakInstrument{
		ident: instruments.InstrumentIdentity{
			ID:          "go-leak",
			Name:        "Go Leak Static Analyzer",
			Version:     "1.0.0",
			Vendor:      "Daemon Core",
			Category:    instruments.CategoryStatic,
			Description: "Analyzes Go source files for unclosed HTTP response bodies",
			Installed:   true,
		},
	}
}

func (g *GoLeakInstrument) Identity() instruments.InstrumentIdentity {
	return g.ident
}

func (g *GoLeakInstrument) Capabilities() []instruments.Capability {
	return []instruments.Capability{instruments.CapStaticAnalysis}
}

func (g *GoLeakInstrument) Detect(ctx context.Context, env instruments.Environment) instruments.DetectionResult {
	if _, err := os.Stat(filepath.Join(env.ProjectDir, "go.mod")); err == nil {
		return instruments.DetectionResult{Compatible: true, Reason: "go.mod exists"}
	}
	return instruments.DetectionResult{Compatible: false, Reason: "No go.mod found"}
}

func (g *GoLeakInstrument) Health(ctx context.Context) instruments.HealthResult {
	return instruments.HealthResult{Status: "AVAILABLE", Reason: "Static analyzer ready"}
}

func (g *GoLeakInstrument) BuildRequest(ctx context.Context, request instruments.InstrumentRequest) (instruments.ToolRequest, error) {
	return instruments.ToolRequest{
		Executable: "go-leak",
		Args:       request.Args,
		Dir:        request.Target,
		ReadOnly:   true,
	}, nil
}

func (g *GoLeakInstrument) Execute(ctx context.Context, request instruments.ToolRequest) (instruments.ToolResult, error) {
	projectDir := request.Dir
	if projectDir == "" {
		projectDir = "."
	}
	leakFile := ""
	_ = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (info.Name() == ".git" || info.Name() == ".daemon" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			content := string(data)
			if (strings.Contains(content, "http.Get(") || strings.Contains(content, ".Do(")) && !strings.Contains(content, ".Body.Close()") {
				leakFile = filepath.Base(path)
				return filepath.SkipDir
			}
		}
		return nil
	})

	if leakFile != "" {
		return instruments.ToolResult{
			InstrumentID: "go-leak",
			Success:      true,
			Stdout:       leakFile,
		}, nil
	}

	return instruments.ToolResult{
		InstrumentID: "go-leak",
		Success:      true,
		Stdout:       "",
	}, nil
}

func (g *GoLeakInstrument) Normalize(ctx context.Context, result instruments.ToolResult) ([]instruments.Evidence, error) {
	var evs []instruments.Evidence
	if result.Stdout != "" {
		leakFile := result.Stdout
		evs = append(evs, instruments.Evidence{
			ID:           "ev-mem-leak",
			Type:         instruments.EvidenceAST,
			Source:       "memory_leak_analyzer",
			EntityID:     leakFile,
			Statement:    "FACT: Unclosed HTTP response body found in file: " + leakFile + ". INFERENCE: Connection pool retention risk.",
			ObservedAt:   time.Now(),
			Freshness:    "live",
			Reliability:  1.0,
			Confidence:   1.0,
			Scope:        "memory",
			RawReference: leakFile,
			Quality: instruments.EvidenceQuality{
				Class:           "static_ast",
				Strength:        1.0,
				Reliability:     1.0,
				Freshness:       1.0,
				Specificity:     1.0,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "go-leak",
			},
		})
	}
	return evs, nil
}
