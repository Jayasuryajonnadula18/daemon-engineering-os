package workflow

import (
	"context"
	"testing"

	"daemon/core/memory"
)

func TestIntelligenceEngine_ConfigurableThresholds(t *testing.T) {
	thresholds := ConfigurableWorkflowThresholds{
		MinOccurrences:           3,
		AutomationMinOccurrences: 5,
		MinSuccessRate:           0.80,
	}

	ie := NewIntelligenceEngine(nil, &thresholds)

	recs := make([]memory.KnowledgeRecord, 6)
	for i := 0; i < 6; i++ {
		recs[i] = memory.KnowledgeRecord{
			ID:             "rec",
			ErrorSignature: "RESTART_LOOP",
		}
	}

	opps, err := ie.AnalyzePatterns(context.Background(), recs)
	if err != nil {
		t.Fatalf("AnalyzePatterns failed: %v", err)
	}

	if len(opps) != 1 {
		t.Fatalf("expected 1 opportunity candidate, got %d", len(opps))
	}
	if opps[0].OpportunityScore != "HIGH" {
		t.Fatalf("expected HIGH opportunity score for 6 occurrences (>= 5 threshold), got %s", opps[0].OpportunityScore)
	}
}

func TestPredictionEngine_BaselineScoring(t *testing.T) {
	pe := NewPredictionEngine()
	preds, err := pe.PredictNextActions(context.Background(), "POST /orders route updated")
	if err != nil {
		t.Fatalf("PredictNextActions failed: %v", err)
	}

	if len(preds) != 3 {
		t.Fatalf("expected 3 predictions, got %d", len(preds))
	}
	if preds[0].Probability != 0.94 || preds[0].Action != "Run Integration Tests" {
		t.Fatalf("top prediction mismatch: %v", preds[0])
	}
}
