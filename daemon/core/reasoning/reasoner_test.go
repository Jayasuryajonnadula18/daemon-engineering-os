package reasoning

import (
	"context"
	"testing"
)

func TestCalculateDaemonConfidence_HardCapOnWeakEvidence(t *testing.T) {
	// Single evidence item must be capped at 0.70 (70%) regardless of raw score
	scoreSingle := CalculateDaemonConfidence(1, 0.95, 0.95, 0.95)
	if scoreSingle > 0.70 {
		t.Fatalf("expected hard cap <= 0.70 on single evidence item, got %.2f", scoreSingle)
	}

	// Multiple strong evidence items calculate uncapped confidence
	scoreMulti := CalculateDaemonConfidence(4, 0.95, 0.95, 0.95)
	if scoreMulti <= 0.70 {
		t.Fatalf("expected uncapped confidence > 0.70 for multiple evidence sources, got %.2f", scoreMulti)
	}
}

func TestEngineeringReasoner_InsufficientContextFallback(t *testing.T) {
	reasoner := NewEngineeringReasoner(nil)
	res, err := reasoner.Reason(context.Background(), "why is checkout failing")
	if err != nil {
		t.Fatalf("unexpected reasoning error: %v", err)
	}

	if !res.InsufficientContext {
		t.Fatalf("expected InsufficientContext to be true when no context engine is wired")
	}
	if res.Confidence != 0.0 {
		t.Fatalf("expected confidence 0.0 on insufficient context, got %.2f", res.Confidence)
	}
}
