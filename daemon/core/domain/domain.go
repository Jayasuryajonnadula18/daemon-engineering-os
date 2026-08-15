package domain

import "time"

// Provenance represents origin, freshness, confidence, and scope metadata for an observation.
type Provenance struct {
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
	Freshness  string    `json:"freshness"`  // e.g., "live", "cached", "stale"
	Confidence float64   `json:"confidence"` // 0.0 - 1.0
	Scope      string    `json:"scope"`      // e.g., "project", "global", "module"
}

// Repository represents a version control repository.
type Repository struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Branch    string    `json:"branch"`
	URL       string    `json:"url"`
	IsClean   bool      `json:"is_clean"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project represents the software project housed inside a repository.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Framework   string    `json:"framework"`
	Language    string    `json:"language"`
	Path        string    `json:"path"`
	Runtime     string    `json:"runtime"`
	PkgManager  string    `json:"pkg_manager"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Workspace represents a developer's workspace configuration.
type Workspace struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// Module represents a logical component or package in the system.
type Module struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"` // e.g., "frontend", "backend", "library"
	Path    string   `json:"path"`
	Imports []string `json:"imports"`
}

// Service represents a running or runnable system service.
type Service struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Port      int      `json:"port"`
	Status    string   `json:"status"` // e.g., "running", "stopped", "unknown"
	DependsOn []string `json:"depends_on"`
}

// API represents an exposed API endpoint or route.
type API struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Service string `json:"service"`
}

// Dependency represents a package dependency.
type Dependency struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Type       string `json:"type"` // e.g., "direct", "dev"
	IsOutdated bool   `json:"is_outdated"`
}

// Deployment represents a deployment run execution.
type Deployment struct {
	ID        string    `json:"id"`
	Env       string    `json:"env"` // e.g., "production", "staging", "local"
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// Incident represents an issue or failure state observed in the repository.
type Incident struct {
	ID         string    `json:"id"`
	Message    string    `json:"message"`
	Severity   string    `json:"severity"` // e.g., "critical", "warning", "info"
	Resolved   bool      `json:"resolved"`
	DetectedAt time.Time `json:"detected_at"`
	ResolvedAt time.Time `json:"resolved_at"`
}

// Recommendation represents an intelligence action proposed to the developer.
type Recommendation struct {
	ID        string `json:"id"`
	Category  string `json:"category"` // e.g., "security", "config", "dependency"
	Message   string `json:"message"`
	Fixable   bool   `json:"fixable"`
	Rationale string `json:"rationale"`
}

// Workflow represents a sequence of steps orchestrated by the workflow engine.
type Workflow struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"` // e.g., "pending", "running", "succeeded", "failed"
	Steps     []string  `json:"steps"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// Policy represents a rule evaluated by the policy engine before running any task.
type Policy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Action      string `json:"action"`   // e.g., "install", "deploy"
	Decision    string `json:"decision"` // e.g., "allow", "deny", "confirm"
	Description string `json:"description"`
}

// Plugin represents an installed extension or tool adapter.
type Plugin struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// Capability represents an operation type.
type Capability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EngineeringReport represents a summary metrics check of the system.
type EngineeringReport struct {
	HealthScore      int                    `json:"health_score"`
	ReadinessScore   int                    `json:"readiness_score"`
	TechDebtCategory string                 `json:"tech_debt_category"`
	Vulnerabilities  int                    `json:"vulnerabilities"`
	Recommendations  []Recommendation       `json:"recommendations"`
	Metrics          map[string]interface{} `json:"metrics"`
}

// Runbook represents a guide to solve common incidents.
type Runbook struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	IncidentPattern string   `json:"incident_pattern"`
	Steps           []string `json:"steps"`
}

// EdgeRecord represents a relational link in the Knowledge Graph.
type EdgeRecord struct {
	FromType string `json:"from_type"`
	FromID   string `json:"from_id"`
	ToType   string `json:"to_type"`
	ToID     string `json:"to_id"`
	Relation string `json:"relation"`
}

// Fact represents an observed fact used during evaluation or CLI questioning
type Fact struct {
	Statement   string   `json:"statement"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type Inference struct {
	Statement  string  `json:"statement"`
	Confidence float64 `json:"confidence"`
}

// ReasoningResult encapsulates the output of the reasoning engine
type ReasoningResult struct {
	Answer              string      `json:"answer"`
	Confidence          float64     `json:"confidence"`
	InsufficientContext bool        `json:"insufficient_context"`
	Facts               []Fact      `json:"facts"`
	Inferences          []Inference `json:"inferences"`
	Unknowns            []string    `json:"unknowns"`
}

