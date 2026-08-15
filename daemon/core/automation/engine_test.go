package automation

import (
	"context"
	"testing"

	"daemon/core/capabilities"
	"daemon/core/policies"
)

func TestAutomationEngineExecutesRegisteredCapability(t *testing.T) {
	reg := capabilities.NewRegistry()
	if err := reg.Register(capabilities.Capability{
		Name:          "restart_container",
		Description:   "Restart a local container",
		Preconditions: []string{"docker available"},
		Inputs:        []string{"container"},
		Risk:          capabilities.RiskLow,
		Execution:     "docker restart <container>",
		Verification:  "container health check",
		Rollback:      "docker restart <container>",
	}); err != nil {
		t.Fatalf("register capability: %v", err)
	}

	engine := NewAutomationEngineWithPolicy(nil, nil, policies.NewMemoryPolicyEngine(false), reg)
	result, err := engine.ExecuteCapability(context.Background(), "restart_container", map[string]string{"container": "orders-api"})
	if err != nil {
		t.Fatalf("execute capability: %v", err)
	}
	if result.Name != "restart_container" {
		t.Fatalf("expected registered capability to execute, got %s", result.Name)
	}
}

func TestAutomationEngineRejectsUnregisteredCapability(t *testing.T) {
	engine := NewAutomationEngineWithPolicy(nil, nil, policies.NewMemoryPolicyEngine(false), capabilities.NewRegistry())
	if _, err := engine.ExecuteCapability(context.Background(), "unknown_capability", nil); err == nil {
		t.Fatal("expected unregistered capability to be rejected")
	}
}

func TestMemoryPolicyEngineRejectsHighRiskProductionCapability(t *testing.T) {
	pe := policies.NewMemoryPolicyEngine(false)
	decision, err := pe.EvaluateCapability(context.Background(), "restart_container", "production", string(capabilities.RiskHigh))
	if err != nil {
		t.Fatalf("evaluate capability: %v", err)
	}
	if decision == policies.DecAllow {
		t.Fatal("expected high-risk production capability to be blocked from auto-approval")
	}
}
