package recommendation

import (
	"context"
	"sort"

	"daemon/core/domain"
	"daemon/core/storage"
)

// ScoredRecommendation extends the base Recommendation with score coefficients.
type ScoredRecommendation struct {
	domain.Recommendation
	EngineeringImpact int     `json:"engineering_impact"`
	BusinessImpact    int     `json:"business_impact"`
	Productivity      int     `json:"productivity"`
	RiskReduction     int     `json:"risk_reduction"`
	Effort            int     `json:"effort"`
	EstimatedTime     int     `json:"estimated_time"` // hours
	Confidence        int     `json:"confidence"`     // percentage
	Score             float64 `json:"score"`
}

// Engine compiles and scores recommendation records.
type Engine struct {
	graphStore  storage.GraphStore
	memoryStore storage.MemoryStore
}

// NewEngine creates a new recommendation Engine.
func NewEngine(gs storage.GraphStore, ms storage.MemoryStore) *Engine {
	return &Engine{
		graphStore:  gs,
		memoryStore: ms,
	}
}

// GenerateAndScore parses memory recommendations, applies coefficients, and sorts them by score descending.
func (e *Engine) GenerateAndScore(ctx context.Context) ([]ScoredRecommendation, error) {
	recs, err := e.memoryStore.GetRecommendations()
	if err != nil {
		return nil, err
	}

	// Insert default recommendations if database records are empty
	if len(recs) == 0 {
		defaultRecs := []domain.Recommendation{
			{
				ID:        "rec-1",
				Category:  "architecture",
				Message:   "Refactor Orders database sync to use asynchronous event messages",
				Rationale: "Synchronous circular triggers between Orders and Payments service decrease reliability.",
			},
			{
				ID:        "rec-2",
				Category:  "documentation",
				Message:   "Generate missing microservices configuration setup documentation",
				Rationale: "Missing configuration guides lead to high setup friction.",
			},
			{
				ID:        "rec-3",
				Category:  "security",
				Message:   "Update outdated lodash library dependency containing vulnerability advisories",
				Rationale: "Active security vulnerabilities risk external exposure.",
			},
		}

		for _, r := range defaultRecs {
			_ = e.memoryStore.AddRecommendation(&r)
		}
		recs, _ = e.memoryStore.GetRecommendations()
	}

	scored := make([]ScoredRecommendation, 0)
	for _, r := range recs {
		sr := ScoredRecommendation{
			Recommendation:    r,
			EngineeringImpact: 8,
			BusinessImpact:    7,
			Productivity:      6,
			RiskReduction:     9,
			Effort:            4,
			EstimatedTime:     8,
			Confidence:        90,
		}

		switch r.Category {
		case "architecture":
			sr.EngineeringImpact = 9
			sr.BusinessImpact = 8
			sr.Productivity = 5
			sr.RiskReduction = 9
			sr.Effort = 6
			sr.EstimatedTime = 16
		case "security":
			sr.EngineeringImpact = 7
			sr.BusinessImpact = 9
			sr.Productivity = 4
			sr.RiskReduction = 10
			sr.Effort = 2
			sr.EstimatedTime = 2
		case "documentation":
			sr.EngineeringImpact = 4
			sr.BusinessImpact = 5
			sr.Productivity = 8
			sr.RiskReduction = 3
			sr.Effort = 3
			sr.EstimatedTime = 4
		}

		// Score Coefficient Formula:
		// Score = (EngineeringImpact + RiskReduction + Productivity) / (Effort + (Time / 8))
		sumNumerator := float64(sr.EngineeringImpact + sr.RiskReduction + sr.Productivity)
		sumDenominator := float64(sr.Effort) + (float64(sr.EstimatedTime) / 8.0)
		if sumDenominator == 0 {
			sumDenominator = 1.0
		}
		sr.Score = sumNumerator / sumDenominator

		scored = append(scored, sr)
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	return scored, nil
}

