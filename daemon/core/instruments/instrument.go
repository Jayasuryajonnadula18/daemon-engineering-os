package instruments

import (
	"context"
	"time"
)

type InstrumentCategory string

const (
	CategoryDebugging     InstrumentCategory = "debugging"
	CategoryProfiling     InstrumentCategory = "profiling"
	CategoryTracing       InstrumentCategory = "tracing"
	CategoryTesting       InstrumentCategory = "testing"
	CategoryStatic        InstrumentCategory = "static"
	CategoryBuild         InstrumentCategory = "build"
	CategoryDependency    InstrumentCategory = "dependency"
	CategoryRuntime       InstrumentCategory = "runtime"
	CategoryDatabase      InstrumentCategory = "database"
	CategoryObservability InstrumentCategory = "observability"
)

type InstrumentIdentity struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Version        string             `json:"version"`
	Vendor         string             `json:"vendor"`
	Category       InstrumentCategory `json:"category"`
	Description    string             `json:"description"`
	License        string             `json:"license"`
	SourceURL      string             `json:"source_url"`
	Installed      bool               `json:"installed"`
	ExecutablePath string             `json:"executable_path"`
}

type Environment struct {
	ProjectDir string            `json:"project_dir"`
	EnvVars    map[string]string `json:"env_vars"`
}

type DetectionResult struct {
	Compatible bool   `json:"compatible"`
	Reason     string `json:"reason"`
}

type HealthResult struct {
	Status string `json:"status"` // "AVAILABLE", "DEGRADED", "UNAVAILABLE", "TOOL_HEALTH_UNKNOWN"
	Reason string `json:"reason"`
}

type InstrumentRequest struct {
	Capability Capability        `json:"capability"`
	Args       []string          `json:"args"`
	Target     string            `json:"target"`
	TimeoutSec int               `json:"timeout_sec"`
	Metadata   map[string]string `json:"metadata"`
}

type ToolRequest struct {
	Executable string            `json:"executable"`
	Args       []string          `json:"args"`
	Dir        string            `json:"dir"`
	Env        []string          `json:"env"`
	ReadOnly   bool              `json:"read_only"`
	Metadata   map[string]string `json:"metadata"`
}

type ToolResult struct {
	InstrumentID string            `json:"instrument_id"`
	ExitCode     int               `json:"exit_code"`
	Duration     time.Duration     `json:"duration"`
	Success      bool              `json:"success"`
	Stdout       string            `json:"stdout"`
	Stderr       string            `json:"stderr"`
	Artifacts    []ArtifactRef     `json:"artifacts"`
	Metadata     map[string]string `json:"metadata"`
}

type ArtifactRef struct {
	ID                    string    `json:"id"`
	Type                  string    `json:"type"`
	Size                  int64     `json:"size"`
	Timestamp             time.Time `json:"timestamp"`
	Path                  string    `json:"path"`
	PrivacyClassification string    `json:"privacy_classification"`
}

// EngineeringInstrument represents a unified interface for external tools and native analyzers
type EngineeringInstrument interface {
	Identity() InstrumentIdentity
	Capabilities() []Capability
	Detect(ctx context.Context, env Environment) DetectionResult
	Health(ctx context.Context) HealthResult
	BuildRequest(ctx context.Context, request InstrumentRequest) (ToolRequest, error)
	Execute(ctx context.Context, request ToolRequest) (ToolResult, error)
	Normalize(ctx context.Context, result ToolResult) ([]Evidence, error)
}
