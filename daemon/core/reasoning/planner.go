package reasoning

import (
	"context"
	"strings"
)

// ExecutionPlan specifies deterministic steps, risks, and rollbacks for approved execution.
type ExecutionPlan struct {
	Intent           string   `json:"intent"`
	Steps            []string `json:"steps"`
	Risks            []string `json:"risks"`
	RollbackStrategy string   `json:"rollback_strategy"`
	Confidence       int      `json:"confidence"`
	RequiresApproval bool     `json:"requires_approval"`
	ModelUsed        string   `json:"model_used"`
}

// EngineeringPlanner reasons over intent and context windows to construct plans.
type EngineeringPlanner struct {
	contextBuilder *ContextBuilder
	modelRouter    *ModelRouter
}

// NewEngineeringPlanner instantiates an EngineeringPlanner.
func NewEngineeringPlanner(cb *ContextBuilder, mr *ModelRouter) *EngineeringPlanner {
	return &EngineeringPlanner{
		contextBuilder: cb,
		modelRouter:    mr,
	}
}

// GeneratePlan builds a structured dry-run blueprint without executing modifying operations.
func (ep *EngineeringPlanner) GeneratePlan(ctx context.Context, intent string) (*ExecutionPlan, error) {
	// 1. Gather optimized context window
	optCtx, err := ep.contextBuilder.BuildOptimizedContext(ctx, intent)
	if err != nil {
		return nil, err
	}

	// 2. Select reasoning model dynamically
	taskType := "planning"
	if strings.Contains(strings.ToLower(intent), "document") {
		taskType = "documentation"
	}
	recModel := ep.modelRouter.RouteTask(taskType, 2000)

	// 3. Construct plan structure deterministically
	plan := &ExecutionPlan{
		Intent:           intent,
		Confidence:       95,
		RequiresApproval: false,
		ModelUsed:        recModel.ModelName,
		RollbackStrategy: "Stash modified workspace state and restart original container instances.",
	}

	intentLower := strings.ToLower(intent)
	switch {
	case strings.Contains(intentLower, "deploy") || strings.Contains(intentLower, "production"):
		plan.RequiresApproval = true
		plan.Confidence = 92
		plan.Steps = []string{
			"Verify local workspace code builds and compiles cleanly.",
			"Execute unit test suite (ensure 100% green).",
			"Compile production assets and target Docker base container images.",
			"Check remote staging cluster resources.",
			"Perform rolling upgrade deploy rollout.",
			"Verify remote health checks parameters.",
		}
		plan.Risks = []string{
			"Redis cache latency spike during rollout initialization.",
			"Stale connection pools to postgres DB endpoints.",
		}
		plan.RollbackStrategy = "Restore DB snapshots, revert cluster deployment versions to original replica sets."

	case strings.Contains(intentLower, "sync") || strings.Contains(intentLower, "drift"):
		plan.Confidence = 88
		plan.Steps = []string{
			"Query remote staging configurations.",
			"Compare schema tables and variable mappings to identify configuration drifts.",
			"Generate diff delta manifest files.",
			"Inject credentials safely without persistent logs database storage.",
			"Apply configuration updates safely.",
		}
		plan.Risks = []string{
			"Active schemas drifts check locks database writers.",
		}

	default:
		// General workflow plan
		plan.Steps = []string{
			"Validate active local development containers status.",
			"Run standard repository stacks checks.",
			"Refreshed workspace internal recommendations metadata.",
		}
		plan.Risks = []string{
			"None identified.",
		}
	}

	// Adjust properties based on active context elements (e.g. incidents size)
	if len(optCtx.Incidents) > 0 {
		plan.Confidence -= 10
		plan.Risks = append(plan.Risks, "Active incidents registered in workspace.")
	}

	return plan, nil
}
