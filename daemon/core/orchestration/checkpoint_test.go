package orchestration

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCheckpointStore_SaveAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_checkpoints.db")

	store, err := NewCheckpointStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create CheckpointStore: %v", err)
	}
	defer store.Close()

	now := time.Now()
	cp := NodeCheckpoint{
		ExecutionID:     "exec-101",
		DAGID:           "dag-101",
		NodeID:          "node-1-build",
		Attempt:         1,
		Status:          NodeVerified,
		InputHash:       "hash-abc",
		StartedAt:       now,
		CompletedAt:     &now,
		OutputRef:       "out-1",
		VerificationRef: "ver-1",
	}

	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	list, err := store.GetCheckpoints("exec-101")
	if err != nil {
		t.Fatalf("failed to get checkpoints: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 checkpoint record, got %d", len(list))
	}
	if list[0].NodeID != "node-1-build" || list[0].Status != NodeVerified {
		t.Fatalf("checkpoint mismatch: %v", list[0])
	}
}

func TestRecoveryEngine_NonReversibleHandling(t *testing.T) {
	re := NewRecoveryEngine()
	nonRevNode := DAGNode{
		ID:             "node-deploy",
		CapabilityName: "prod_deploy",
		Inputs:         map[string]string{"target": "k8s"},
		Reversible:     false,
	}

	strat := re.EvaluateRecovery(nonRevNode, nil)
	if strat.CanRollback {
		t.Fatalf("expected non-reversible node to NOT allow rollback")
	}
	if strat.Class != FailureNonReversible {
		t.Fatalf("expected FailureNonReversible class, got %s", strat.Class)
	}
}
