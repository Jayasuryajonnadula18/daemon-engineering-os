package analysis

import (
	"time"
)

type FindingCategory string

const (
	CategoryMemory           FindingCategory = "memory"
	CategoryPerformance      FindingCategory = "performance"
	CategoryConcurrency      FindingCategory = "concurrency"
	CategoryCorrectness      FindingCategory = "correctness"
	CategorySecurity         FindingCategory = "security"
	CategoryReliability      FindingCategory = "reliability"
	CategoryErrorHandling    FindingCategory = "error_handling"
	CategoryResourceLifecycle FindingCategory = "resource_lifecycle"
	CategoryDependency       FindingCategory = "dependency"
	CategoryArchitecture     FindingCategory = "architecture"
	CategoryAPI              FindingCategory = "api"
	CategoryTesting          FindingCategory = "testing"
	CategoryConfiguration    FindingCategory = "configuration"
	CategoryObservability    FindingCategory = "observability"
	CategoryMaintainability  FindingCategory = "maintainability"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

type FactType string

const (
	FactTypeFact       FactType = "FACT"
	FactTypeInference  FactType = "INFERENCE"
	FactTypeHypothesis FactType = "HYPOTHESIS"
)

type Finding struct {
	ID               string          `json:"id"`
	Category         FindingCategory `json:"category"`
	Severity         Severity        `json:"severity"`
	FactType         FactType        `json:"fact_type"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	EvidenceIDs      []string        `json:"evidence_ids"`
	Files            []string        `json:"files"`
	AffectedFiles    []string        `json:"affected_files,omitempty"`
	Symbols          []string        `json:"symbols"`
	Services         []string        `json:"services"`
	DetectionMethod  string          `json:"detection_method"`
	Confidence       float64         `json:"confidence"`
	AutoFixAvailable bool            `json:"auto_fix_available"`
	SuggestedActions []string        `json:"suggested_actions"`
	FirstObserved    time.Time       `json:"first_observed"`
	LastObserved     time.Time       `json:"last_observed"`
}

type AnalyzerStatus struct {
	Name      string    `json:"name"`
	Available bool      `json:"available"`
	RunTime   string    `json:"run_time"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type AnalysisResult struct {
	Findings        []Finding        `json:"findings"`
	Evidence        []map[string]any `json:"evidence"`
	Recommendations []string         `json:"recommendations"`
	Confidence      float64          `json:"confidence"`
	AIEnhanced      bool             `json:"ai_enhanced"`
	AnalyzerStatus  []AnalyzerStatus `json:"analyzer_status"`
	AnalyzedFiles   int              `json:"analyzed_files"`
	CacheHits       int              `json:"cache_hits"`
	CacheMisses     int              `json:"cache_misses"`
	Timestamp       time.Time        `json:"timestamp"`
	Status          AnalysisStatus   `json:"status"`
	StatusReason    string           `json:"status_reason,omitempty"`
	FilesRemaining  int              `json:"files_remaining,omitempty"`
}
