package orchestration

import (
	"context"
	"testing"
)

func TestImpactEngine_BlastRadiusCalculation(t *testing.T) {
	ie := NewImpactEngine(nil)

	impact, err := ie.AnalyzeImpact(context.Background(), "orders")
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	if impact.TargetEntity != "orders" {
		t.Fatalf("expected TargetEntity 'orders', got %s", impact.TargetEntity)
	}
	if impact.BlastRadiusScore == 0.0 {
		t.Fatalf("expected non-zero BlastRadiusScore")
	}
	if len(impact.SinglePointsOfFailure) == 0 {
		t.Fatalf("expected single points of failure analysis")
	}
}
