package debug

import "math"

type RootCause struct {
	Statement          string   `json:"statement"`
	EvidenceIDs        []string `json:"evidence_ids"`
	Confidence         float64  `json:"confidence"`
	VerificationStatus string   `json:"verification_status"` // VERIFIED, PARTIALLY_VERIFIED, UNVERIFIED, REFUTED
}

// Configurable confidence weights (stored or defaulted)
type ConfidenceConfig struct {
	EvidenceStrengthWeight   float64 `json:"evidence_strength_weight"`
	SourceReliabilityWeight  float64 `json:"source_reliability_weight"`
	FreshnessWeight          float64 `json:"freshness_weight"`
	AgreementWeight          float64 `json:"agreement_weight"`
	IndependenceWeight       float64 `json:"independence_weight"`
	ContradictionPenalty     float64 `json:"contradiction_penalty"`
	VerificationBonus        float64 `json:"verification_bonus"`
	HistoricalSuccessWeight  float64 `json:"historical_success_weight"`
}

func DefaultConfidenceConfig() ConfidenceConfig {
	return ConfidenceConfig{
		EvidenceStrengthWeight:  0.2,
		SourceReliabilityWeight: 0.2,
		FreshnessWeight:         0.1,
		AgreementWeight:         0.15,
		IndependenceWeight:      0.15,
		ContradictionPenalty:    0.3,
		VerificationBonus:       0.2,
		HistoricalSuccessWeight: 0.1,
	}
}

// RankRootCauses orders root cause candidates based on evaluated evidence
func RankRootCauses(causes []RootCause, config ConfidenceConfig) []RootCause {
	// Simple sorting or ranking can be performed here.
	// For this Phase 1 kernel, we will expose the RankRootCauses function to order by confidence descending.
	for i := 0; i < len(causes); i++ {
		for j := i + 1; j < len(causes); j++ {
			if causes[j].Confidence > causes[i].Confidence {
				causes[i], causes[j] = causes[j], causes[i]
			}
		}
	}
	return causes
}

// CalculateConfidence implements the dynamic score calculation logic, yielding
// a rich HypothesisConfidence descriptor rather than a raw float64.
func CalculateConfidence(
	strength, reliability, freshness, agreement, independence, histSuccess float64,
	hasContradiction bool,
	verifiedStatus string,
	cfg ConfidenceConfig,
	supporting []string,
	contradicting []string,
) HypothesisConfidence {
	score := strength*cfg.EvidenceStrengthWeight +
		reliability*cfg.SourceReliabilityWeight +
		freshness*cfg.FreshnessWeight +
		agreement*cfg.AgreementWeight +
		independence*cfg.IndependenceWeight +
		histSuccess*cfg.HistoricalSuccessWeight

	if hasContradiction {
		score -= cfg.ContradictionPenalty
	}

	if verifiedStatus == "VERIFIED" {
		score += cfg.VerificationBonus
	}

	// Clamp the score between 0.0 and 1.0
	score = math.Max(0.0, math.Min(1.0, score))

	// Hard ceilings when evidence is weak
	if strength < 0.3 && score > 0.4 {
		score = 0.4
	}

	// Calculate a ceiling based on evidence breadth/variety
	supportingCount := len(supporting)
	contradictingCount := len(contradicting)
	
	ceiling := 1.0
	if supportingCount == 0 {
		ceiling = 0.2
	} else if supportingCount == 1 {
		ceiling = 0.9
	}
	
	if score > ceiling {
		score = ceiling
	}

	return HypothesisConfidence{
		Score:              score,
		SupportingCount:    supportingCount,
		ContradictingCount: contradictingCount,
		VerifiedBy:         supporting,
		Ceiling:            ceiling,
		Method:             "deterministic",
	}
}

