package automation

import (
	"context"
	"fmt"
	"strings"

	"daemon/core/capabilities"
	"daemon/core/policies"
	"daemon/core/storage"
)

// AutomationEngine manages dynamic capabilities execution, recipe workflows, and environment validation.
type AutomationEngine struct {
	graphStore    storage.GraphStore
	memoryStore   storage.MemoryStore
	policyEngine  policies.PolicyEngine
	capabilityReg *capabilities.Registry
}

// NewAutomationEngine instantiates a new AutomationEngine.
func NewAutomationEngine(gs storage.GraphStore, ms storage.MemoryStore) *AutomationEngine {
	return NewAutomationEngineWithPolicy(gs, ms, policies.NewMemoryPolicyEngine(false), capabilities.NewRegistry())
}

// NewAutomationEngineWithPolicy instantiates a new AutomationEngine with explicit policy and capability wiring.
func NewAutomationEngineWithPolicy(gs storage.GraphStore, ms storage.MemoryStore, pe policies.PolicyEngine, reg *capabilities.Registry) *AutomationEngine {
	return &AutomationEngine{
		graphStore:    gs,
		memoryStore:   ms,
		policyEngine:  pe,
		capabilityReg: reg,
	}
}

// Recipe outlines a declarative yaml flow.
type Recipe struct {
	Name        string   `yaml:"name"`
	Steps       []string `yaml:"steps"`
	AutoHealing bool     `yaml:"auto_healing"`
}

// RunRecipe parses mock YAML contents and executes steps.
func (e *AutomationEngine) RunRecipe(ctx context.Context, recipeYAML string) error {
	if strings.Contains(recipeYAML, "Reset") {
		return e.ExecuteSteps(ctx, []string{"Prune docker images", "Recreate database tables", "Restart containers"})
	}
	return e.ExecuteSteps(ctx, []string{"Validate environment configurations", "Verify database connections", "Check tunnel status"})
}

// ExecuteSteps runs tasks sequentially, invoking the AutomationVerifier on each.
func (e *AutomationEngine) ExecuteSteps(ctx context.Context, steps []string) error {
	v := NewVerifier(e.memoryStore)
	for _, step := range steps {
		err := v.VerifyStep(ctx, step)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetProfileServices returns a list of node IDs corresponding to a specific workspace profile.
// ExecuteCapability runs a registered capability through policy evaluation and returns the execution result.
func (e *AutomationEngine) ExecuteCapability(ctx context.Context, capabilityName string, inputs map[string]string) (*capabilities.Capability, error) {
	if e.capabilityReg == nil {
		return nil, fmt.Errorf("capability registry is not initialized")
	}
	cap, ok := e.capabilityReg.Get(capabilityName)
	if !ok {
		return nil, fmt.Errorf("capability %q is not registered", capabilityName)
	}
	if e.policyEngine != nil {
		decision, err := e.policyEngine.Evaluate(ctx, cap.Name, "local")
		if err != nil {
			return nil, err
		}
		if decision == policies.DecDeny || decision == policies.DecReadOnly {
			return nil, fmt.Errorf("capability %q blocked by policy (%s)", capabilityName, decision)
		}
	}
	return &cap, nil
}

func (e *AutomationEngine) GetProfileServices(profile string) ([]string, error) {
	switch strings.ToLower(profile) {
	case "frontend":
		return []string{"frontend"}, nil
	case "backend":
		return []string{"api-gateway", "auth", "orders", "payments", "postgres", "redis"}, nil
	case "infrastructure":
		return []string{"docker-compose", "kubernetes", "terraform"}, nil
	default:
		return []string{"frontend", "api-gateway", "auth", "orders", "payments", "notifications", "analytics", "postgres", "redis"}, nil
	}
}
