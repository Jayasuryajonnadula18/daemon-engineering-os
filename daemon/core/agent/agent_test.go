package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"daemon/core/instruments"
	gobuild "daemon/core/instruments/adapters/build/go"
	"daemon/core/policies"
)

func TestAgent_Registry(t *testing.T) {
	reg := instruments.NewInstrumentRegistry()
	_ = reg.Register(gobuild.NewGoBuildInstrument())
	list := reg.List()
	if len(list) == 0 {
		t.Fatal("expected registered instruments, got 0")
	}

	inst := reg.FindByID("go-build")
	if inst == nil {
		t.Fatal("expected 'go-build' instrument to exist in registry")
	}
}

func TestAgent_BudgetExceeded(t *testing.T) {
	dbPath := ":memory:"
	store, err := NewSessionStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create SessionStore: %v", err)
	}
	defer store.Close()

	pe := policies.NewMemoryPolicyEngine(false)
	runtime := NewAgentRuntime(pe, nil, nil, store)

	sess := AgentSession{
		ID:        "test-sess-budget",
		Intent:    "inspect workspace",
		State:     string(StateIdle),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Budget: AgentBudget{
			MaxIterations: 1, // very low to trigger budget exceed
			MaxDuration:   60.0,
			MaxToolCalls:  5,
		},
	}
	_ = store.SaveSession(sess)

	res, err := runtime.RunLoop(context.Background(), sess.ID, sess.Intent, false)
	if err != nil {
		t.Fatalf("RunLoop failed: %v", err)
	}

	if res.State != string(StateBudgetExceeded) {
		t.Errorf("expected state %s, got %s", StateBudgetExceeded, res.State)
	}
}

func TestAgent_Persistence(t *testing.T) {
	dbPath := "test_agent_sessions.db"
	defer os.Remove(dbPath)

	store, err := NewSessionStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}
	defer store.Close()

	sess := AgentSession{
		ID:        "sess-123",
		Intent:    "run diagnostics",
		State:     string(StateIdle),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Budget: AgentBudget{
			MaxIterations: 10,
			MaxDuration:   30.0,
		},
	}

	err = store.SaveSession(sess)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	retrieved, err := store.GetSession("sess-123")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if retrieved.ID != sess.ID || retrieved.Intent != sess.Intent {
		t.Errorf("retrieved session does not match saved one")
	}
}
