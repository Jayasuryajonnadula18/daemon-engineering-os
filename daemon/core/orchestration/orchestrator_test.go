package orchestration

import (
	"context"
	"testing"
)

func TestOrchestrator_DryRun(t *testing.T) {
	orch := NewOrchestrator(nil, nil)
	intent := ExecutionIntent{
		Objective: "restart orders service",
		Targets:   []string{"service-orders"},
	}

	res, err := orch.ExecuteIntent(context.Background(), intent, true, "")
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	if !res.DryRun {
		t.Fatalf("expected DryRun flag to be true")
	}
	if res.FinalState != StateCompleted {
		t.Fatalf("expected final state StateCompleted, got %s", res.FinalState)
	}
	if len(res.Waves) != 3 {
		t.Fatalf("expected 3 scheduled waves, got %d", len(res.Waves))
	}
}

func TestOrchestrator_CooperativeCancellation(t *testing.T) {
	orch := NewOrchestrator(nil, nil)
	execID := "exec-test-cancel"

	orch.CancelExecution(execID)

	intent := ExecutionIntent{
		Objective: "cancel orders service deploy",
		Targets:   []string{"service-orders"},
	}

	res, err := orch.ExecuteIntent(context.Background(), intent, false, execID)
	if err != nil {
		t.Fatalf("ExecuteIntent failed: %v", err)
	}

	if res.FinalState != StateRolledBack {
		t.Fatalf("expected final state StateRolledBack after cancellation, got %s", res.FinalState)
	}
}
