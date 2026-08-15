package instruments

import (
	"strings"
	"time"
)

type EvidenceType string

const (
	EvidenceGit           EvidenceType = "GIT"
	EvidenceCode          EvidenceType = "CODE"
	EvidenceSourceCode    EvidenceType = "SOURCE_CODE"
	EvidenceAST           EvidenceType = "AST"
	EvidenceBuild         EvidenceType = "BUILD"
	EvidenceTest          EvidenceType = "TEST"
	EvidenceLog           EvidenceType = "LOG"
	EvidenceMetric        EvidenceType = "METRIC"
	EvidenceHistory       EvidenceType = "HISTORY"
	EvidenceDatamine      EvidenceType = "DATAMINE"
	EvidenceModel         EvidenceType = "MODEL"
	EvidenceRuntime       EvidenceType = "RUNTIME"
	EvidenceProcess       EvidenceType = "PROCESS"
	EvidencePort          EvidenceType = "PORT"
	EvidenceDependency    EvidenceType = "DEPENDENCY"
	EvidenceConfiguration EvidenceType = "CONFIGURATION"
	EvidenceKG            EvidenceType = "KNOWLEDGE_GRAPH"
	EvidenceTwin          EvidenceType = "ENGINEERING_TWIN"
	EvidenceEvent         EvidenceType = "EVENT"
	EvidenceWorkflow      EvidenceType = "WORKFLOW"
)

// EvidenceQuality describes the measured quality attributes of a single evidence item.
// It is NOT a fixed universal strength table. Quality describes the evidence itself,
// not the conclusion drawn from it. The debugger calculates effective confidence
// by combining these fields dynamically.
type EvidenceQuality struct {
	// Class is the structural category of the evidence source.
	// Examples: "compiler_error", "static_ast", "profiler_sample", "log_pattern", "llm_hypothesis"
	Class string `json:"class"`

	// Strength measures how directly the evidence addresses the target hypothesis (0.0–1.0).
	Strength float64 `json:"strength"`

	// Reliability measures the trustworthiness of the source instrument (0.0–1.0).
	// A compiler diagnostic is more reliable than an LLM hypothesis.
	Reliability float64 `json:"reliability"`

	// Freshness measures how current the evidence is (0.0–1.0).
	// 1.0 = captured live during this investigation; 0.0 = historical or stale.
	Freshness float64 `json:"freshness"`

	// Specificity measures how targeted the evidence is to the specific problem (0.0–1.0).
	// A profiler flame graph for the exact slow function scores higher than a generic system metric.
	Specificity float64 `json:"specificity"`

	// Independence measures whether the evidence was obtained independently (0.0–1.0).
	// 1.0 = fully independent; 0.0 = derived from or correlated with another evidence item.
	Independence float64 `json:"independence"`

	// Reproducibility measures whether the observation could be reproduced (0.0–1.0).
	Reproducibility float64 `json:"reproducibility"`

	// Verification is the verification status: "VERIFIED", "UNVERIFIED", or "INCONCLUSIVE".
	Verification string `json:"verification"`

	// Provenance records which instrument generated this evidence.
	Provenance string `json:"provenance"`
}

// EffectiveStrength calculates a combined quality score across all attributes.
// Verified evidence is scored at full weight; unverified evidence is penalized.
// This is always calculated dynamically — never looked up from a fixed table.
func (q EvidenceQuality) EffectiveStrength() float64 {
	base := (q.Strength + q.Reliability + q.Freshness + q.Specificity) / 4.0
	if q.Verification == "VERIFIED" {
		return base
	}
	// Penalize unverified evidence to prevent false certainty
	return base * 0.7
}

// Evidence represents a single normalized observation produced by an instrument
// after executing an experiment. It is the unit of information flowing into
// the correlation engine.
type Evidence struct {
	ID           string          `json:"id"`
	Type         EvidenceType    `json:"type,omitempty"`
	Statement    string          `json:"statement"`
	Source       string          `json:"source"`
	Instrument   string          `json:"instrument,omitempty"`
	Quality      EvidenceQuality `json:"quality"`
	ObservedAt   time.Time       `json:"observed_at"`
	EntityIDs    []string        `json:"entity_ids,omitempty"`
	EntityID     string          `json:"entity_id,omitempty"` // For backward compatibility
	RawReference string          `json:"raw_reference,omitempty"`
	Freshness    string          `json:"freshness,omitempty"`
	Reliability  float64         `json:"reliability,omitempty"`
	Confidence   float64         `json:"confidence,omitempty"`
	Scope        string          `json:"scope,omitempty"`
	Location     *CodeLocation   `json:"location,omitempty"`
	DerivedFrom  []string        `json:"derived_from,omitempty"`
	RawArtifact  *ArtifactRef    `json:"raw_artifact,omitempty"`
	Redacted     bool            `json:"redacted"`
}

// CostEstimate describes the expected execution cost of an instrument invocation.
type CostEstimate struct {
	Level       string  `json:"level"` // "LOW", "MEDIUM", "HIGH"
	DurationSec float64 `json:"duration_sec"`
}

// ResourceImpact describes the expected hardware load of an instrument invocation.
type ResourceImpact struct {
	CPUPercent   float64 `json:"cpu_percent"`
	RAMMegabytes float64 `json:"ram_megabytes"`
}

// AvailabilityState captures the five distinct availability stages of an instrument.
// These states are mutually exclusive in progression — an instrument cannot be
// CAPABILITY_AVAILABLE unless it has passed all prior stages.
type AvailabilityState struct {
	// AdapterExists is true if Daemon has a registered adapter for this instrument.
	AdapterExists bool `json:"adapter_exists"`

	// ToolDiscovered is true if the tool binary was found on the system path.
	ToolDiscovered bool `json:"tool_discovered"`

	// ToolInstalled is true if the tool binary is confirmed installed and executable.
	ToolInstalled bool `json:"tool_installed"`

	// HealthUnknown is true if the tool health check has not yet been run.
	HealthUnknown bool `json:"health_unknown"`

	// ProjectCompatible is true if the tool is compatible with the target project.
	ProjectCompatible bool `json:"project_compatible"`

	// CapabilityAvailable is true only when all prior conditions are met.
	// This is the gate the Instrument Selector uses to determine fitness.
	CapabilityAvailable bool `json:"capability_available"`
}

// InstrumentSelection is the result of querying the InstrumentSelector for a capability.
// It explicitly names both the selected instrument and all alternatives, so that the
// investigation plan is always auditable.
type InstrumentSelection struct {
	InstrumentID string            `json:"instrument_id"`
	Capability   Capability        `json:"capability"`
	Rationale    []string          `json:"rationale"`
	Alternatives []string          `json:"alternatives"`
	Availability AvailabilityState `json:"availability"`
}

type CodeLocation struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// RedactSecrets checks for sensitive environment parameters or credentials and replaces them.
func RedactSecrets(input string) (string, bool) {
	lower := strings.ToLower(input)
	if strings.Contains(lower, "daemon_test_secret_do_not_use") {
		return "[REDACTED_SECRET]", true
	}

	sensitiveKeywords := []string{
		"password",
		"api_key",
		"apikey",
		"secret",
		"private_key",
		"token",
		"bearer",
		"db_pass",
		"mysql_pwd",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lower, keyword+"=") || strings.Contains(lower, keyword+":") || strings.Contains(lower, keyword+" ") {
			return "[REDACTED_SECRET]", true
		}
	}

	return input, false
}
