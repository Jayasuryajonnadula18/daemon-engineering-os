package evolution

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestEvolutionEngine_PromotionAndDemotionLifecycle(t *testing.T) {
	ledger, err := NewFixLedger(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory ledger: %v", err)
	}

	cfg := PromotionConfig{
		MinOccurrences:         10,
		MinVerifiedSuccessRate: 0.85,
		FailureDemotionThreshold: 0.60,
	}
	engine := NewEvolutionEngine(ledger, cfg)

	ctx := context.Background()

	// 1. Initial Outcome -> OBSERVED
	p1, _ := engine.RecordOutcome(ctx, OutcomeRecord{
		ActionID:         "act-1",
		PatternID:        "pat-restart-build",
		VerifiedSuccess:  true,
		EnvironmentScope: ScopeProject,
		Timestamp:        time.Now(),
	})
	if p1.Status != StatusObserved {
		t.Fatalf("expected status OBSERVED on 1st outcome, got %s", p1.Status)
	}

	// 2. Simulate 9 more successful outcomes (total 10 successes, 10 occurrences) -> TRUSTED
	for i := 2; i <= 10; i++ {
		p1, _ = engine.RecordOutcome(ctx, OutcomeRecord{
			ActionID:         "act-iter",
			PatternID:        "pat-restart-build",
			VerifiedSuccess:  true,
			EnvironmentScope: ScopeProject,
			Timestamp:        time.Now(),
		})
	}
	if p1.Status != StatusTrusted {
		t.Fatalf("expected status TRUSTED after 10 successful occurrences, got %s", p1.Status)
	}

	// 3. Simulate repeated failures -> REVIEW / DEGRADED
	for i := 1; i <= 8; i++ {
		p1, _ = engine.RecordOutcome(ctx, OutcomeRecord{
			ActionID:         fmt.Sprintf("act-fail-%d", i),
			PatternID:        "pat-restart-build",
			VerifiedSuccess:  false,
			RootCause:        "Connection refused to db password=secret123",
			FixSummary:       "Restart database container",
			EnvironmentScope: ScopeProject,
			Timestamp:        time.Now(),
		})
	}

	if p1.Status != StatusReview && p1.Status != StatusDegraded {
		t.Fatalf("expected status REVIEW or DEGRADED after repeated failures, got %s", p1.Status)
	}

	// 4. Verify Fix Ledger secret redaction
	entries, err := ledger.GetEntries()
	if err != nil {
		t.Fatalf("failed to retrieve fix ledger entries: %v", err)
	}
	if len(entries) != 8 {
		t.Fatalf("expected 8 fix ledger entries, got %d", len(entries))
	}
	if entries[0].RootCause != "[REDACTED_SECRET]" {
		t.Fatalf("expected secret to be redacted, got '%s'", entries[0].RootCause)
	}
}
