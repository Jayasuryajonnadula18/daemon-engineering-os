package automation

import (
	"context"
	"strings"

	"daemon/core/storage"
)

// AutomationEngine manages dynamic capabilities execution, recipe workflows, and environment validation.
type AutomationEngine struct {
	graphStore  storage.GraphStore
	memoryStore storage.MemoryStore
}

// NewAutomationEngine instantiates a new AutomationEngine.
func NewAutomationEngine(gs storage.GraphStore, ms storage.MemoryStore) *AutomationEngine {
	return &AutomationEngine{
		graphStore:  gs,
		memoryStore: ms,
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
