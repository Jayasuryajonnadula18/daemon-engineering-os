package debug

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)


func TestKernel_StateTransitions(t *testing.T) {
	inv := NewInvestigation("test-transition", "checkout is slow", DefaultDebugBudget())

	// Triage to EvidenceCollection should pass
	err := inv.Transition(StateEvidenceCollection)
	if err != nil {
		t.Fatalf("expected triage to evidence collection transition to pass, got: %v", err)
	}

	// EvidenceCollection to Localization should pass
	err = inv.Transition(StateLocalization)
	if err != nil {
		t.Fatalf("expected evidence collection to localization transition to pass, got: %v", err)
	}

	// Localization to HypothesisTesting should pass
	err = inv.Transition(StateHypothesisTesting)
	if err != nil {
		t.Fatalf("expected localization to hypothesis testing transition to pass, got: %v", err)
	}

	// HypothesisTesting to Verification should pass
	err = inv.Transition(StateVerification)
	if err != nil {
		t.Fatalf("expected hypothesis testing to verification transition to pass, got: %v", err)
	}

	// Verification to RootCauseIdentified should pass
	err = inv.Transition(StateRootCauseFound)
	if err != nil {
		t.Fatalf("expected verification to root cause identified transition to pass, got: %v", err)
	}

	// Illegal transition: RootCauseFound to Triage should fail
	err = inv.Transition(StateTriage)
	if err == nil {
		t.Fatal("expected root cause found to triage transition to fail, but it passed")
	}
}

func TestKernel_BudgetEnforcement(t *testing.T) {
	budget := DebugBudget{
		MaxDurationSeconds: 1.0,
		MaxIterations:      2,
		MaxFiles:           10,
		MaxEvidenceItems:   5,
		MaxExperiments:     3,
		MaxTests:           2,
		MaxAIRequests:      2,
		MaxContextTokens:   100,
	}

	inv := NewInvestigation("test-budget", "build is failing", budget)

	// Increment iterations to exceed MaxIterations limit
	inv.Iterations = 3
	isExceeded := inv.CheckBudget(10 * time.Millisecond)
	if !isExceeded {
		t.Fatal("expected iterations budget limit to be exceeded")
	}
	if inv.Status != StateInsufficient || inv.Reason != "investigation_budget_exhausted" {
		t.Errorf("expected status %s and reason 'investigation_budget_exhausted', got state=%s reason=%s", StateInsufficient, inv.Status, inv.Reason)
	}
}

func TestKernel_Correlation(t *testing.T) {
	evList := []Evidence{
		{
			ID:        "ev1",
			Type:      EvidenceAST,
			Source:    "lifetime_analyzer",
			EntityID:  "checkout/service.go",
			Statement: "Unclosed HTTP body found in file",
			Scope:     "file",
		},
		{
			ID:        "ev2",
			Type:      EvidenceRuntime,
			Source:    "port_monitor",
			EntityID:  "checkout/service.go",
			Statement: "Active socket leak detected in runtime",
			Scope:     "file",
		},
	}

	groups := Correlate(evList)
	if len(groups) == 0 {
		t.Fatal("expected correlated groups, got 0")
	}

	foundIndependent := false
	for _, g := range groups {
		if g.Type == CorrelationIndependent {
			foundIndependent = true
		}
	}

	if !foundIndependent {
		t.Error("expected independent confirmation correlation type between AST and Runtime evidence")
	}
}

func TestKernel_DebuggerRun(t *testing.T) {
	// Use a clean, isolated temp workspace to avoid detecting issues in the
	// live Daemon repo. The test intent is to verify the debugger runs to
	// completion and persists its result correctly.
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module testapp\ngo 1.20\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	dbPath := ":memory:"
	store, err := NewDebugStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create debug store: %v", err)
	}
	defer store.Close()

	debugger := NewDebugger(store, nil)
	inv, err := debugger.RunInvestigation(context.Background(), "test-run-123", "memory leak", tmp, false, false, false)
	if err != nil {
		t.Fatalf("RunInvestigation failed: %v", err)
	}

	// A clean workspace with a minimal main.go has no detectable memory leak patterns.
	// The debugger must NOT manufacture a root cause (No False Certainty rule).
	if inv.Status == StateRootCauseFound && len(inv.RootCauses) == 0 {
		t.Error("ROOT_CAUSE_IDENTIFIED with no root causes is a false certainty violation")
	}
	for _, rc := range inv.RootCauses {
		if rc.VerificationStatus != "VERIFIED" {
			t.Errorf("root cause %q has unverified status %q — No False Certainty violation", rc.Statement, rc.VerificationStatus)
		}
	}

	// Verify persistence: whatever status was reached, it must be stored correctly
	retrieved, err := store.GetInvestigation("test-run-123")
	if err != nil {
		t.Fatalf("failed to retrieve investigation: %v", err)
	}
	if retrieved.Problem != inv.Problem || retrieved.Status != inv.Status {
		t.Error("retrieved investigation does not match saved one")
	}
}

