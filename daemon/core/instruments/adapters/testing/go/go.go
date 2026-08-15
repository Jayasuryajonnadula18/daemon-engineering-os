package gotest

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"daemon/core/instruments"
)

type GoTestInstrument struct {
	ident instruments.InstrumentIdentity
}

func NewGoTestInstrument() *GoTestInstrument {
	return &GoTestInstrument{
		ident: instruments.InstrumentIdentity{
			ID:          "go-test",
			Name:        "Go Test Runner",
			Version:     "1.20+",
			Vendor:      "Go Authors",
			Category:    instruments.CategoryTesting,
			Description: "Executes Go unit and integration tests",
			License:     "BSD-style",
			SourceURL:   "https://go.dev",
		},
	}
}

func (g *GoTestInstrument) Identity() instruments.InstrumentIdentity {
	g.ident.Installed = instruments.IsBinaryInstalled("go")
	if g.ident.Installed {
		g.ident.ExecutablePath = "go"
	}
	return g.ident
}

func (g *GoTestInstrument) Capabilities() []instruments.Capability {
	return []instruments.Capability{instruments.CapUnitTesting}
}

func (g *GoTestInstrument) Detect(ctx context.Context, env instruments.Environment) instruments.DetectionResult {
	if _, err := os.Stat(filepath.Join(env.ProjectDir, "go.mod")); err == nil {
		return instruments.DetectionResult{Compatible: true, Reason: "go.mod exists"}
	}
	return instruments.DetectionResult{Compatible: false, Reason: "No go.mod found"}
}

func (g *GoTestInstrument) Health(ctx context.Context) instruments.HealthResult {
	if instruments.IsBinaryInstalled("go") {
		return instruments.HealthResult{Status: "AVAILABLE", Reason: "Go binary found on PATH"}
	}
	return instruments.HealthResult{Status: "UNAVAILABLE", Reason: "Go binary not installed"}
}

func (g *GoTestInstrument) BuildRequest(ctx context.Context, request instruments.InstrumentRequest) (instruments.ToolRequest, error) {
	return instruments.ToolRequest{
		Executable: "go",
		Args:       append([]string{"test"}, request.Args...),
		Dir:        request.Target,
		ReadOnly:   false,
	}, nil
}

func (g *GoTestInstrument) Execute(ctx context.Context, request instruments.ToolRequest) (instruments.ToolResult, error) {
	executor := instruments.NewInstrumentExecutor(nil, nil)
	return executor.ExecuteRequest(ctx, instruments.CapUnitTesting, request, request.Dir, "project", true)
}

func (g *GoTestInstrument) Normalize(ctx context.Context, result instruments.ToolResult) ([]instruments.Evidence, error) {
	var evs []instruments.Evidence

	if !result.Success {
		evs = append(evs, instruments.Evidence{
			ID:        "ev-gotest-failure",
			Statement: "FACT: Go unit tests failed to run successfully. Output: " + result.Stdout,
			Source:    "test_runner",
			Instrument: "go-test",
			ObservedAt: time.Now(),
			Quality: instruments.EvidenceQuality{
				Class:           "test_failure",
				Strength:        1.0,
				Reliability:     1.0,
				Freshness:       1.0,
				Specificity:     0.9,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "go-test",
			},
		})
	}

	return evs, nil
}
