package debug

import (
	"strings"
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
	intentLower := strings.ToLower(intent)

	// Rule 1: leak / growth pattern
	if strings.Contains(intentLower, "leak") || strings.Contains(intentLower, "growth") || strings.Contains(intentLower, "memory") {
		list = append(list, Hypothesis{
			ID:        "hyp-leak-http-body",
			Statement: "Unclosed HTTP response bodies are retaining network connection and memory resources.",
			Confidence: HypothesisConfidence{
				Score:   0.5,
				Ceiling: 1.0,
				Method:  "deterministic",
			},
			Conclusion: ConclusionInconclusive,
			CreatedAt:  time.Now(),
		})
		list = append(list, Hypothesis{
			ID:        "hyp-leak-goroutine",
			Statement: "Goroutines spawned without termination channels or context cancellation are leaking memory.",
			Confidence: HypothesisConfidence{
				Score:   0.5,
				Ceiling: 1.0,
				Method:  "deterministic",
			},
			Conclusion: ConclusionInconclusive,
			CreatedAt:  time.Now(),
		})
	}

	// Rule 2: build failure / crash
	if strings.Contains(intentLower, "build") || strings.Contains(intentLower, "compile") || strings.Contains(intentLower, "crash") || strings.Contains(intentLower, "panic") {
		list = append(list, Hypothesis{
			ID:        "hyp-build-syntax",
			Statement: "A syntax error, undefined symbol, or runtime exception/crash is present in the codebase.",
			Confidence: HypothesisConfidence{
				Score:   0.6,
				Ceiling: 1.0,
				Method:  "deterministic",
			},
			Conclusion: ConclusionInconclusive,
			CreatedAt:  time.Now(),
		})
	}

	// Rule 3: test failure
	if strings.Contains(intentLower, "test") || strings.Contains(intentLower, "fail") {
		list = append(list, Hypothesis{
			ID:        "hyp-test-regression",
			Statement: "A code change in a recently modified module has caused unit tests to fail.",
			Confidence: HypothesisConfidence{
				Score:   0.5,
				Ceiling: 1.0,
				Method:  "deterministic",
			},
			Conclusion: ConclusionInconclusive,
			CreatedAt:  time.Now(),
		})
	}

	// Generic fallback
	if len(list) == 0 {
		list = append(list, Hypothesis{
			ID:        "hyp-generic-regression",
			Statement: "A functional regression was introduced by a recent commit change.",
			Confidence: HypothesisConfidence{
				Score:   0.3,
				Ceiling: 1.0,
				Method:  "deterministic",
			},
			Conclusion: ConclusionInconclusive,
			CreatedAt:  time.Now(),
		})
	}

	return list
}
