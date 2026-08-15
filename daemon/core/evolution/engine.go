package evolution

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type OutcomeRecord struct {
	ActionID         string     `json:"action_id"`
	PatternID        string     `json:"pattern_id"`
	VerifiedSuccess  bool       `json:"verified_success"`
	VerificationRef  string     `json:"verification_ref"`
	RootCause        string     `json:"root_cause,omitempty"`
	FixSummary       string     `json:"fix_summary,omitempty"`
	EnvironmentScope ScopeLevel `json:"environment_scope"`
	Timestamp        time.Time  `json:"timestamp"`
}

type PatternRecord struct {
	PatternID         string        `json:"pattern_id"`
	Occurrences       int           `json:"occurrences"`
	VerifiedSuccesses int           `json:"verified_successes"`
	Failures          int           `json:"failures"`
	Confidence        float64       `json:"confidence"`
	Status            PatternStatus `json:"status"`
	Scope             ScopeLevel    `json:"scope"`
	LastEvaluated     time.Time     `json:"last_evaluated"`
}

type EvolutionEngine struct {
	mu       sync.RWMutex
	ledger   *FixLedger
	config   PromotionConfig
	patterns map[string]*PatternRecord
}

func NewEvolutionEngine(ledger *FixLedger, cfg PromotionConfig) *EvolutionEngine {
	return &EvolutionEngine{
		ledger:   ledger,
		config:   cfg,
		patterns: make(map[string]*PatternRecord),
	}
}

// RecordOutcome processes a verified execution outcome and evaluates pattern promotion/demotion.
func (ee *EvolutionEngine) RecordOutcome(ctx context.Context, outcome OutcomeRecord) (*PatternRecord, error) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	pattern, exists := ee.patterns[outcome.PatternID]
	if !exists {
		pattern = &PatternRecord{
			PatternID: outcome.PatternID,
			Status:    StatusObserved,
			Scope:     outcome.EnvironmentScope,
		}
		ee.patterns[outcome.PatternID] = pattern
	}

	pattern.Occurrences++
	if outcome.VerifiedSuccess {
		pattern.VerifiedSuccesses++
	} else {
		pattern.Failures++
		if ee.ledger != nil && outcome.RootCause != "" {
			_ = ee.ledger.RecordFix(FixLedgerEntry{
				ActionID:           outcome.ActionID,
				PatternID:          outcome.PatternID,
				ErrorSignature:     "EXECUTION_FAILURE",
				RootCause:          outcome.RootCause,
				FixSummary:         outcome.FixSummary,
				VerificationResult: outcome.VerificationRef,
				Environment:        string(outcome.EnvironmentScope),
				Timestamp:          time.Now(),
			})
		}
	}

	// Calculate confidence score weighted by scope
	successRate := 0.0
	if pattern.Occurrences > 0 {
		successRate = float64(pattern.VerifiedSuccesses) / float64(pattern.Occurrences)
	}
	scopeWeight := CalculateScopeWeight(pattern.Scope)
	pattern.Confidence = successRate * (1.0 + (scopeWeight - 1.0)*0.1)
	if pattern.Confidence > 1.0 {
		pattern.Confidence = 1.0
	}

	// Evaluate Lifecycle State Machine
	pattern.Status = ee.evaluateLifecycle(pattern)
	pattern.LastEvaluated = time.Now()

	return pattern, nil
}

func (ee *EvolutionEngine) evaluateLifecycle(p *PatternRecord) PatternStatus {
	successRate := 0.0
	if p.Occurrences > 0 {
		successRate = float64(p.VerifiedSuccesses) / float64(p.Occurrences)
	}

	// Failure Demotion Gate
	if p.Failures >= 2 && successRate < ee.config.FailureDemotionThreshold {
		if p.Status == StatusTrusted {
			return StatusReview
		}
		return StatusDegraded
	}

	// Promotion Gates
	if p.Occurrences >= ee.config.MinOccurrences && successRate >= ee.config.MinVerifiedSuccessRate {
		return StatusTrusted
	}
	if p.Occurrences >= 5 && successRate >= 0.80 {
		return StatusPromotable
	}
	if p.Occurrences >= 3 {
		return StatusCandidate
	}

	return StatusObserved
}

func (ee *EvolutionEngine) RecordFeedback(patternID string, feedback FeedbackType) (*PatternRecord, error) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	p, exists := ee.patterns[patternID]
	if !exists {
		return nil, fmt.Errorf("pattern %s not found", patternID)
	}

	switch feedback {
	case FeedbackAccepted:
		p.Confidence += 0.05
	case FeedbackRejected:
		p.Confidence -= 0.10
		p.Failures++
	case FeedbackModified:
		p.Confidence += 0.02
	}

	if p.Confidence > 1.0 {
		p.Confidence = 1.0
	}
	if p.Confidence < 0.0 {
		p.Confidence = 0.0
	}

	p.Status = ee.evaluateLifecycle(p)
	p.LastEvaluated = time.Now()

	return p, nil
}

func (ee *EvolutionEngine) GetPatterns() []*PatternRecord {
	ee.mu.RLock()
	defer ee.mu.RUnlock()

	var list []*PatternRecord
	for _, p := range ee.patterns {
		list = append(list, p)
	}
	return list
}
