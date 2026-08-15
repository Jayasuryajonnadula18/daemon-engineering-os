package evaluation

import (
	"context"
	"testing"
)

func TestIntelligenceEvaluator_50ScenarioBenchmark(t *testing.T) {
	scenarios := Generate50BenchmarkScenarios()
	if len(scenarios) != 50 {
		t.Fatalf("expected 50 benchmark scenarios, got %d", len(scenarios))
	}

	evaluator := NewIntelligenceEvaluator(nil)
	report, err := evaluator.EvaluateBenchmark(context.Background(), scenarios)
	if err != nil {
		t.Fatalf("unexpected error evaluating benchmark: %v", err)
	}

	if report.TotalCases != 50 {
		t.Fatalf("expected total cases to be 50, got %d", report.TotalCases)
	}
	if report.OverallEvalScore < 90.0 {
		t.Fatalf("expected overall eval score >= 90%%, got %.2f%%", report.OverallEvalScore)
	}
}
