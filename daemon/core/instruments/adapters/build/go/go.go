package gobuild

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"daemon/core/instruments"
)

type GoBuildInstrument struct {
	ident instruments.InstrumentIdentity
}

func NewGoBuildInstrument() *GoBuildInstrument {
	return &GoBuildInstrument{
		ident: instruments.InstrumentIdentity{
			ID:          "go-build",
			Name:        "Go Compiler",
			Version:     "1.20+",
			Vendor:      "Go Authors",
			Category:    instruments.CategoryBuild,
			Description: "Compiles Go source files and packages",
			License:     "BSD-style",
			SourceURL:   "https://go.dev",
		},
	}
}

func (g *GoBuildInstrument) Identity() instruments.InstrumentIdentity {
	// Dynamically discover if installed
	g.ident.Installed = instruments.IsBinaryInstalled("go")
	if g.ident.Installed {
		g.ident.ExecutablePath = "go"
	}
	return g.ident
}

func (g *GoBuildInstrument) Capabilities() []instruments.Capability {
	return []instruments.Capability{instruments.CapBuild}
}

func (g *GoBuildInstrument) Detect(ctx context.Context, env instruments.Environment) instruments.DetectionResult {
	if _, err := os.Stat(filepath.Join(env.ProjectDir, "go.mod")); err == nil {
		return instruments.DetectionResult{Compatible: true, Reason: "go.mod exists"}
	}
	return instruments.DetectionResult{Compatible: false, Reason: "No go.mod found"}
}

func (g *GoBuildInstrument) Health(ctx context.Context) instruments.HealthResult {
	if instruments.IsBinaryInstalled("go") {
		return instruments.HealthResult{Status: "AVAILABLE", Reason: "Go binary found on PATH"}
	}
	return instruments.HealthResult{Status: "UNAVAILABLE", Reason: "Go binary not installed"}
}

func (g *GoBuildInstrument) BuildRequest(ctx context.Context, request instruments.InstrumentRequest) (instruments.ToolRequest, error) {
	return instruments.ToolRequest{
		Executable: "go",
		Args:       append([]string{"build"}, request.Args...),
		Dir:        request.Target,
		ReadOnly:   false,
	}, nil
}

func (g *GoBuildInstrument) Execute(ctx context.Context, request instruments.ToolRequest) (instruments.ToolResult, error) {
	// Actual execution is managed centrally by the InstrumentExecutor to respect Layer 1 + Governor.
	// We implement this as a fallback self-execution.
	executor := instruments.NewInstrumentExecutor(nil, nil)
	return executor.ExecuteRequest(ctx, instruments.CapBuild, request, request.Dir, "project", true)
}

func (g *GoBuildInstrument) Normalize(ctx context.Context, result instruments.ToolResult) ([]instruments.Evidence, error) {
	var evs []instruments.Evidence

	if !result.Success {
		evs = append(evs, instruments.Evidence{
			ID:        "ev-gobuild-error",
			Statement: "FACT: Compiler failed to build Go project. " + result.Stdout,
			Source:    "compiler",
			Instrument: "go-build",
			ObservedAt: time.Now(),
			Quality: instruments.EvidenceQuality{
				Class:           "compiler_error",
				Strength:        1.0,
				Reliability:     1.0,
				Freshness:       1.0,
				Specificity:     1.0,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "go-build",
			},
		})
	}

	return evs, nil
}
