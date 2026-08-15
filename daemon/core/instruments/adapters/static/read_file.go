package staticread

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	"daemon/core/instruments"
)

type ReadFileInstrument struct {
	ident instruments.InstrumentIdentity
}

func NewReadFileInstrument() *ReadFileInstrument {
	return &ReadFileInstrument{
		ident: instruments.InstrumentIdentity{
			ID:          "read_file",
			Name:        "read_file",
			Version:     "1.0.0",
			Vendor:      "Daemon Core",
			Category:    instruments.CategoryStatic,
			Description: "Reads and outputs file contents safely",
			Installed:   true,
		},
	}
}

func (r *ReadFileInstrument) Identity() instruments.InstrumentIdentity {
	return r.ident
}

func (r *ReadFileInstrument) Capabilities() []instruments.Capability {
	return []instruments.Capability{instruments.CapStaticAnalysis}
}

func (r *ReadFileInstrument) Detect(ctx context.Context, env instruments.Environment) instruments.DetectionResult {
	return instruments.DetectionResult{Compatible: true, Reason: "Always compatible"}
}

func (r *ReadFileInstrument) Health(ctx context.Context) instruments.HealthResult {
	return instruments.HealthResult{Status: "AVAILABLE", Reason: "Ready"}
}

func (r *ReadFileInstrument) BuildRequest(ctx context.Context, request instruments.InstrumentRequest) (instruments.ToolRequest, error) {
	return instruments.ToolRequest{
		Executable: "read_file",
		Args:       request.Args,
		Dir:        request.Target,
		ReadOnly:   true,
	}, nil
}

func (r *ReadFileInstrument) Execute(ctx context.Context, request instruments.ToolRequest) (instruments.ToolResult, error) {
	if len(request.Args) == 0 {
		return instruments.ToolResult{Success: false, Stdout: "No file specified"}, nil
	}
	filePath := filepath.Join(request.Dir, request.Args[0])
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return instruments.ToolResult{Success: false, Stdout: err.Error()}, nil
	}
	return instruments.ToolResult{
		InstrumentID: "read_file",
		Success:      true,
		Stdout:       string(data),
	}, nil
}

func (r *ReadFileInstrument) Normalize(ctx context.Context, result instruments.ToolResult) ([]instruments.Evidence, error) {
	var evs []instruments.Evidence
	if result.Success {
		evs = append(evs, instruments.Evidence{
			ID:         "ev-read-file",
			Type:       instruments.EvidenceSourceCode,
			Source:     "read_file",
			Statement:  "FACT: Read file content successfully.",
			ObservedAt: time.Now(),
		})
	}
	return evs, nil
}
