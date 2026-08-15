package evolution

import (
	"context"
	"testing"
	"time"
)

// TestEvolutionSafety_CannotModifyLayerOnePolicy verifies the Evolution Engine
// cannot promote patterns that claim policy ceiling mutations.
func TestEvolutionSafety_CannotModifyLayerOnePolicy(t *testing.T) {
	ledger, err := NewFixLedger(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory ledger: %v", err)
	}
	engine := NewEvolutionEngine(ledger, DefaultPromotionConfig())
	ctx := context.Background()

	_, err = engine.RecordOutcome(ctx, OutcomeRecord{
		ActionID:         "ev-safety-1",
		PatternID:        "pat-policy-mutate",
		VerifiedSuccess:  false, // not verified
		RootCause:        "attempted policy ceiling mutation",
		FixSummary:       "raise max_shell_ops ceiling to unlimited",
		EnvironmentScope: ScopeProject,
		Timestamp:        time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	patterns := engine.GetPatterns()
	for _, p := range patterns {
		if p.PatternID == "pat-policy-mutate" {
			if p.Status == StatusPromotable || p.Status == StatusTrusted {
				t.Fatalf("evolution engine promoted an unverified policy-mutation pattern to %s", p.Status)
			}
		}
	}
}

// TestEvolutionSafety_OnlyVerifiedOutcomesPromote verifies that patterns require
// verified successes (not just executions) before confidence increases meaningfully.
func TestEvolutionSafety_OnlyVerifiedOutcomesPromote(t *testing.T) {
	ledger, err := NewFixLedger(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory ledger: %v", err)
	}

	cfg := DefaultPromotionConfig()
	engine := NewEvolutionEngine(ledger, cfg)
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		_, _ = engine.RecordOutcome(ctx, OutcomeRecord{
			ActionID:         "unverified-act",
			PatternID:        "pat-unverified",
			VerifiedSuccess:  false,
			RootCause:        "execution failed verification",
			FixSummary:       "attempted fix",
			EnvironmentScope: ScopeGeneric,
			Timestamp:        time.Now(),
		})
	}

	patterns := engine.GetPatterns()
	for _, p := range patterns {
		if p.PatternID == "pat-unverified" {
			if p.Status == StatusPromotable || p.Status == StatusTrusted {
				t.Fatalf("unverified pattern was promoted to %s — evolution safety violated", p.Status)
			}
		}
	}
}

// TestEvolutionSafety_SecretNeverInLedger verifies Fix Ledger redacts secrets
// before persisting to SQLite.
func TestEvolutionSafety_SecretNeverInLedger(t *testing.T) {
	ledger, err := NewFixLedger(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory ledger: %v", err)
	}
	engine := NewEvolutionEngine(ledger, DefaultPromotionConfig())
	ctx := context.Background()

	sensitiveText := "connection string: password=SUPERSECRET_12345 host=db"
	_, _ = engine.RecordOutcome(ctx, OutcomeRecord{
		ActionID:         "sec-test-1",
		PatternID:        "pat-db-conn",
		VerifiedSuccess:  false,
		RootCause:        sensitiveText,
		FixSummary:       "Reconnect with rotated credentials",
		EnvironmentScope: ScopeProject,
		Timestamp:        time.Now(),
	})

	entries, err := ledger.GetEntries()
	if err != nil {
		t.Fatalf("failed to get entries: %v", err)
	}

	for _, entry := range entries {
		if entry.RootCause == sensitiveText {
			t.Fatalf("raw secret persisted to fix ledger without redaction")
		}
	}
}
