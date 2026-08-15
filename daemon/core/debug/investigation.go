package debug

import (
	"fmt"
	"time"

	"daemon/core/analysis"
	"daemon/core/instruments"
)

type InvestigationState string

type VerificationSummary struct {
	Status             string   `json:"status"` // VERIFIED, PARTIALLY_VERIFIED, UNVERIFIED, REFUTED
	VerificationMethod string   `json:"verification_method"`
	EvidenceIDs        []string `json:"evidence_ids"`
}

const (
	StateTriage             InvestigationState = "TRIAGE"
	StateEvidenceCollection InvestigationState = "EVIDENCE_COLLECTION"
	StateLocalization       InvestigationState = "LOCALIZATION"
	StateHypothesisTesting  InvestigationState = "HYPOTHESIS_TESTING"
	StateVerification       InvestigationState = "VERIFICATION"
	StateRootCauseFound     InvestigationState = "ROOT_CAUSE_IDENTIFIED"
	StateInsufficient       InvestigationState = "INSUFFICIENT_CONTEXT"
	StateNoIssue            InvestigationState = "NO_ISSUE_FOUND"
	StateFailed             InvestigationState = "FAILED"
	StateCancelled          InvestigationState = "CANCELLED"
)

// AllowedTransitions defines the legal state transition mappings
var AllowedTransitions = map[InvestigationState][]InvestigationState{
	StateTriage: {
		StateEvidenceCollection,
		StateFailed,
		StateCancelled,
	},
	StateEvidenceCollection: {
		StateLocalization,
		StateFailed,
		StateCancelled,
		StateInsufficient,
	},
	StateLocalization: {
		StateHypothesisTesting,
		StateFailed,
		StateCancelled,
		StateInsufficient,
	},
	StateHypothesisTesting: {
		StateVerification,
		StateFailed,
		StateCancelled,
		StateInsufficient,
		StateNoIssue,
	},
	StateVerification: {
		StateRootCauseFound,
		StateFailed,
		StateCancelled,
		StateInsufficient,
		StateNoIssue,
		StateHypothesisTesting, // can loop back for next experiment
	},
	StateRootCauseFound:     {},
	StateInsufficient:       {},
	StateNoIssue:           {},
	StateFailed:             {},
	StateCancelled:          {},
}

type DebugBudget struct {
	MaxDurationSeconds float64 `json:"max_duration_seconds"`
	MaxIterations      int     `json:"max_iterations"`
	MaxFiles           int     `json:"max_files"`
	MaxEvidenceItems   int     `json:"max_evidence_items"`
	MaxExperiments     int     `json:"max_experiments"`
	MaxTests           int     `json:"max_tests"`
	MaxAIRequests      int     `json:"max_ai_requests"`
	MaxContextTokens   int     `json:"max_context_tokens"`
}

func DefaultDebugBudget() DebugBudget {
	return DebugBudget{
		MaxDurationSeconds: 60.0,
		MaxIterations:      15,
		MaxFiles:           100,
		MaxEvidenceItems:   50,
		MaxExperiments:     10,
		MaxTests:           5,
		MaxAIRequests:      5,
		MaxContextTokens:   8192,
	}
}

type Investigation struct {
	ID                  string               `json:"id"`
	Problem             string               `json:"problem"`
	Status              InvestigationState   `json:"status"`
	Reason              string               `json:"reason,omitempty"`
	StartedAt           time.Time            `json:"started_at"`
	CompletedAt         *time.Time           `json:"completed_at,omitempty"`
	AIEnhanced          bool                 `json:"ai_enhanced"`
	InsufficientContext bool                 `json:"insufficient_context"`
	Findings            []analysis.Finding   `json:"findings"`
	Hypotheses          []Hypothesis         `json:"hypotheses"`
	Experiments         []ExperimentResult   `json:"experiments"`
	Evidence            []instruments.Evidence `json:"evidence"`
	RootCauses          []RootCause          `json:"root_causes"`
	Recommendations     []string             `json:"recommendations"`
	Verification        VerificationSummary  `json:"verification"`
	Confidence          float64              `json:"confidence"`
	Budget              DebugBudget          `json:"budget"`
	Iterations          int                  `json:"iterations"`
	FilesInspected      int                  `json:"files_inspected"`
	TestsExecuted       int                  `json:"tests_executed"`
	AIRequestsCount     int                  `json:"ai_requests_count"`
	DurationMs          int64                `json:"duration_ms"`
}

func NewInvestigation(id, problem string, budget DebugBudget) *Investigation {
	return &Investigation{
		ID:         id,
		Problem:    RedactSecrets(problem),
		Status:     StateTriage,
		StartedAt:  time.Now(),
		Findings:   []analysis.Finding{},
		Hypotheses: []Hypothesis{},
		Evidence:   []instruments.Evidence{},
		RootCauses: []RootCause{},
		Budget:     budget,
	}
}

// Transition performs state transition checks and updates the status
func (inv *Investigation) Transition(to InvestigationState) error {
	if inv.Status == to {
		return nil
	}
	allowed, ok := AllowedTransitions[inv.Status]
	if !ok {
		return fmt.Errorf("invalid current state: %s", inv.Status)
	}
	for _, state := range allowed {
		if state == to {
			inv.Status = to
			return nil
		}
	}
	return fmt.Errorf("illegal state transition from %s to %s", inv.Status, to)
}

// CheckBudget verifies if any limit of the DebugBudget has been exceeded
func (inv *Investigation) CheckBudget(elapsed time.Duration) bool {
	if elapsed.Seconds() >= inv.Budget.MaxDurationSeconds {
		inv.Status = StateInsufficient
		inv.Reason = "investigation_budget_exhausted"
		inv.InsufficientContext = true
		return true
	}
	if inv.Iterations >= inv.Budget.MaxIterations {
		inv.Status = StateInsufficient
		inv.Reason = "investigation_budget_exhausted"
		inv.InsufficientContext = true
		return true
	}
	if len(inv.Evidence) >= inv.Budget.MaxEvidenceItems {
		inv.Status = StateInsufficient
		inv.Reason = "investigation_budget_exhausted"
		inv.InsufficientContext = true
		return true
	}
	if len(inv.Experiments) >= inv.Budget.MaxExperiments {
		inv.Status = StateInsufficient
		inv.Reason = "investigation_budget_exhausted"
		inv.InsufficientContext = true
		return true
	}
	if inv.TestsExecuted >= inv.Budget.MaxTests {
		inv.Status = StateInsufficient
		inv.Reason = "investigation_budget_exhausted"
		inv.InsufficientContext = true
		return true
	}
	if inv.AIRequestsCount >= inv.Budget.MaxAIRequests {
		inv.Status = StateInsufficient
		inv.Reason = "investigation_budget_exhausted"
		inv.InsufficientContext = true
		return true
	}
	return false
}
