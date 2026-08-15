package evolution

type PatternStatus string

const (
	StatusObserved   PatternStatus = "OBSERVED"
	StatusCandidate  PatternStatus = "CANDIDATE"
	StatusPromotable PatternStatus = "PROMOTABLE"
	StatusTrusted    PatternStatus = "TRUSTED"
	StatusReview     PatternStatus = "REVIEW"
	StatusDegraded   PatternStatus = "DEGRADED"
	StatusDeprecated PatternStatus = "DEPRECATED"
)

type PromotionConfig struct {
	MinOccurrences          int     `json:"min_occurrences"`
	MinVerifiedSuccessRate  float64 `json:"min_verified_success_rate"`
	FailureDemotionThreshold float64 `json:"failure_demotion_threshold"`
}

func DefaultPromotionConfig() PromotionConfig {
	return PromotionConfig{
		MinOccurrences:         10,
		MinVerifiedSuccessRate: 0.85,
		FailureDemotionThreshold: 0.60,
	}
}
