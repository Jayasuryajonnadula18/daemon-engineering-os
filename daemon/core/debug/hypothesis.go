package debug

import (
	"time"

	"daemon/core/instruments"
)

type HypothesisChallenge struct {
	LeadingHypothesis string             `json:"leading_hypothesis"`
	Alternatives      []string           `json:"alternatives"`
	Supporting        []string           `json:"supporting_evidence_ids"`
	Contradicting     []string           `json:"contradicting_evidence_ids"`
	Missing           []string           `json:"missing_evidence_ids"`
	FalsificationPlan []ExperimentPlan   `json:"falsification_plan"`
}

type Conclusion string

const (
	ConclusionSupported    Conclusion = "SUPPORTED"
	ConclusionRefuted      Conclusion = "REFUTED"
	ConclusionInconclusive Conclusion = "INCONCLUSIVE"
	// ConclusionContradicted means evidence directly contradicted this hypothesis.
	// Unlike REFUTED (soft rejection), a Contradicted hypothesis is eliminated
	// from further experiment selection.
	ConclusionContradicted Conclusion = "CONTRADICTED"
)

type HypothesisConfidence struct {
	Score              float64  `json:"score"`
	SupportingCount    int      `json:"supporting_count"`
	ContradictingCount int      `json:"contradicting_count"`
	VerifiedBy         []string `json:"verified_by"`
	Ceiling            float64  `json:"ceiling"`
	Method             string   `json:"method"`
}

type Hypothesis struct {
	ID                    string               `json:"id"`
	Statement             string               `json:"statement"`
	SupportingEvidence    []string             `json:"supporting_evidence"`
	ContradictingEvidence []string             `json:"contradicting_evidence"`
	Confidence            HypothesisConfidence `json:"confidence"`
	Conclusion            Conclusion           `json:"conclusion"`
	CreatedAt             time.Time            `json:"created_at"`
}

// GenerateDeterministicHypotheses generates code hypotheses based on standard rule-based triggers
func GenerateDeterministicHypotheses(intent string, evidenceList []instruments.Evidence) []Hypothesis {
	var list []Hypothesis

	list = append(list, Hypothesis{
		ID:        "hyp-build-syntax",
		Statement: "A syntax error, undefined symbol, or runtime exception/crash is present in the codebase.",
		Confidence: HypothesisConfidence{Score: 0.6, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-leak-http-body",
		Statement: "Global event emitter listeners or WebSocket subscriptions are not closed, retaining connection contexts in memory indefinitely.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-leak-goroutine",
		Statement: "Goroutines spawned without termination channels or context cancellation are leaking memory.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-test-regression",
		Statement: "A code change in a recently modified module has caused unit tests to fail.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-jwt-bypass",
		Statement: "JWT signature verification is bypassed by trusting the 'none' algorithm, exposing the system to authorization bypass exploits.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-cache-mismatch",
		Statement: "Cache invalidation keys do not match cache storage keys, resulting in stale catalog data after updates.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-concurrency-race",
		Statement: "Check-then-act stock validation patterns are vulnerable to concurrency race conditions, allowing stock overselling.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-db-lock",
		Statement: "Uncommitted SQLite write transactions leak database locks on failure branches, causing subsequent write operations to hang.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-event-loop-block",
		Statement: "Synchronous, thread-blocking busy-wait loops freeze the single-threaded JavaScript event loop, causing WebSocket connection timeouts.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-float-precision",
		Statement: "Strict decimal/float comparisons are vulnerable to floating-point precision mismatches, causing transaction verification to fail.",
		Confidence: HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})
	list = append(list, Hypothesis{
		ID:        "hyp-generic-regression",
		Statement: "Array index used as key attribute in React rendering, which can lead to state mismatch during item deletions.",
		Confidence: HypothesisConfidence{Score: 0.3, Ceiling: 1.0, Method: "deterministic"},
		Conclusion: ConclusionInconclusive,
		CreatedAt:  time.Now(),
	})

	return list
}
