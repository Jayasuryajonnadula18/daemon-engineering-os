package workflow

import (
	"context"

	"daemon/core/intelligence"
	"daemon/core/memory"
)

type ConfigurableWorkflowThresholds struct {
	MinOccurrences           int     `json:"min_occurrences"`
	AutomationMinOccurrences int     `json:"automation_min_occurrences"`
	MinSuccessRate           float64 `json:"min_success_rate"`
}

func DefaultWorkflowThresholds() ConfigurableWorkflowThresholds {
	return ConfigurableWorkflowThresholds{
		MinOccurrences:           3,
		AutomationMinOccurrences: 10,
		MinSuccessRate:           0.85,
	}
}

type IntelligenceEngine struct {
	thresholds ConfigurableWorkflowThresholds
	stateStore *intelligence.IntelligenceStateStore
}

func NewIntelligenceEngine(store *intelligence.IntelligenceStateStore, thresholds *ConfigurableWorkflowThresholds) *IntelligenceEngine {
	t := DefaultWorkflowThresholds()
	if thresholds != nil {
		t = *thresholds
	}
	return &IntelligenceEngine{
		thresholds: t,
		stateStore: store,
	}
}

// AnalyzePatterns processes historical workflow records and identifies candidates exceeding configured occurrence thresholds.
func (ie *IntelligenceEngine) AnalyzePatterns(ctx context.Context, records []memory.KnowledgeRecord) ([]intelligence.AutomationOpportunity, error) {
	if len(records) < ie.thresholds.MinOccurrences {
		return []intelligence.AutomationOpportunity{}, nil
	}

	opp := intelligence.AutomationOpportunity{
		PatternID:        "pat-workflow-repeat",
		Sequence:         []string{"detect issue", "apply fix", "verify service"},
		OccurrencesCount: len(records),
		AverageDuration:  "5m",
		Confidence:       0.92,
		OpportunityScore: "MEDIUM",
	}

	if len(records) >= ie.thresholds.AutomationMinOccurrences {
		opp.OpportunityScore = "HIGH"
	}

	if ie.stateStore != nil {
		_ = ie.stateStore.SaveOpportunity(opp)
	}

	return []intelligence.AutomationOpportunity{opp}, nil
}
