package reasoning

import (
	"context"
	"strings"
)

// ExecutionNode represents a task node in the Directed Acyclic Graph (DAG).
type ExecutionNode struct {
	ID       string `json:"id"`
	TaskName string `json:"task_name"`
	Status   string `json:"status"` // Pending, Running, Completed, Failed
}

// ExecutionEdge defines execution dependencies between DAG nodes.
type ExecutionEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ExecutionGraph represents the compiled DAG blueprint.
type ExecutionGraph struct {
	Nodes []ExecutionNode `json:"nodes"`
	Edges []ExecutionEdge `json:"edges"`
}

// OrchestrationPlan holds the plan metadata, risks, and DAG.
type OrchestrationPlan struct {
	Intent           string         `json:"intent"`
	Graph            ExecutionGraph `json:"graph"`
	Confidence       int            `json:"confidence"`
	Risks            []string       `json:"risks"`
	RollbackStrategy string         `json:"rollback_strategy"`
	RequiresApproval bool           `json:"requires_approval"`
	ModelUsed        string         `json:"model_used"`
	Domain           string         `json:"domain"`
}

// Specialized Orchestrators
type DeploymentOrchestrator struct{}
type ArchitectureOrchestrator struct{}
type WorkspaceOrchestrator struct{}
type InfrastructureOrchestrator struct{}
type IncidentResponseOrchestrator struct{}
type MaintenanceOrchestrator struct{}
type SecurityOrchestrator struct{}
type PerformanceOrchestrator struct{}
type MigrationOrchestrator struct{}
type DocumentationOrchestrator struct{}

// EngineeringOrchestrator coordinates semantic task mapping.
type EngineeringOrchestrator struct {
	contextBuilder *ContextBuilder
	modelRouter    *ModelRouter

	DeployOrch   *DeploymentOrchestrator
	ArchOrch     *ArchitectureOrchestrator
	WorkOrch     *WorkspaceOrchestrator
	InfraOrch    *InfrastructureOrchestrator
	IncidentOrch *IncidentResponseOrchestrator
	MaintOrch    *MaintenanceOrchestrator
	SecOrch      *SecurityOrchestrator
	PerfOrch     *PerformanceOrchestrator
	MigrOrch     *MigrationOrchestrator
	DocOrch      *DocumentationOrchestrator
}

// NewEngineeringOrchestrator instantiates an EngineeringOrchestrator.
func NewEngineeringOrchestrator(cb *ContextBuilder, mr *ModelRouter) *EngineeringOrchestrator {
	return &EngineeringOrchestrator{
		contextBuilder: cb,
		modelRouter:    mr,
		DeployOrch:     &DeploymentOrchestrator{},
		ArchOrch:       &ArchitectureOrchestrator{},
		WorkOrch:       &WorkspaceOrchestrator{},
		InfraOrch:      &InfrastructureOrchestrator{},
		IncidentOrch:   &IncidentResponseOrchestrator{},
		MaintOrch:      &MaintenanceOrchestrator{},
		SecOrch:        &SecurityOrchestrator{},
		PerfOrch:       &PerformanceOrchestrator{},
		MigrOrch:       &MigrationOrchestrator{},
		DocOrch:        &DocumentationOrchestrator{},
	}
}

// Orchestrate compiles intent and semantic context into a DAG Execution Graph.
func (eo *EngineeringOrchestrator) Orchestrate(ctx context.Context, intent string) (*OrchestrationPlan, error) {
	optCtx, err := eo.contextBuilder.BuildOptimizedContext(ctx, intent)
	if err != nil {
		return nil, err
	}

	intentLower := strings.ToLower(intent)
	domain := "Workspace"
	taskType := "planning"

	if strings.Contains(intentLower, "deploy") || strings.Contains(intentLower, "production") {
		domain = "Deployment"
	} else if strings.Contains(intentLower, "arch") || strings.Contains(intentLower, "design") {
		domain = "Architecture"
	} else if strings.Contains(intentLower, "infra") || strings.Contains(intentLower, "cloud") {
		domain = "Infrastructure"
	} else if strings.Contains(intentLower, "security") {
		domain = "Security"
	}

	recModel := eo.modelRouter.RouteTask(taskType, 2500)

	plan := &OrchestrationPlan{
		Intent:           intent,
		Confidence:       94,
		RequiresApproval: false,
		ModelUsed:        recModel.ModelName,
		Domain:           domain,
		RollbackStrategy: "Stash modified local configurations, clean workspace caches, restart Docker stack.",
	}

	// Compile DAG Nodes and Edges based on domain
	switch domain {
	case "Deployment":
		plan.RequiresApproval = true
		plan.Graph = ExecutionGraph{
			Nodes: []ExecutionNode{
				{ID: "node-1", TaskName: "Verify workspace compilation", Status: "Pending"},
				{ID: "node-2", TaskName: "Run test suites in parallel", Status: "Pending"},
				{ID: "node-3", TaskName: "Build docker base images", Status: "Pending"},
				{ID: "node-4", TaskName: "Apply schema migrations check", Status: "Pending"},
				{ID: "node-5", TaskName: "Upgrade replica set deployment", Status: "Pending"},
				{ID: "node-6", TaskName: "Verify gateway telemetry response", Status: "Pending"},
			},
			Edges: []ExecutionEdge{
				{From: "node-1", To: "node-2"},
				{From: "node-1", To: "node-3"},
				{From: "node-2", To: "node-4"},
				{From: "node-3", To: "node-5"},
				{From: "node-4", To: "node-5"},
				{From: "node-5", To: "node-6"},
			},
		}
		plan.Risks = []string{
			"Parallel builds memory exhaustion",
			"Postgres schema locks during migration checks",
		}
		plan.RollbackStrategy = "Revert deployment versions to original replica sets, notify telemetry alerts."

	case "Security":
		plan.RequiresApproval = true
		plan.Graph = ExecutionGraph{
			Nodes: []ExecutionNode{
				{ID: "node-1", TaskName: "Scan secrets references", Status: "Pending"},
				{ID: "node-2", TaskName: "Validate config environment variables", Status: "Pending"},
				{ID: "node-3", TaskName: "Enforce GITHUB_PAT rotation", Status: "Pending"},
			},
			Edges: []ExecutionEdge{
				{From: "node-1", To: "node-3"},
				{From: "node-2", To: "node-3"},
			},
		}
		plan.Risks = []string{
			"Access tokens drift keys verification failures",
		}

	default:
		// Default Workspace domain DAG
		plan.Graph = ExecutionGraph{
			Nodes: []ExecutionNode{
				{ID: "node-1", TaskName: "Scan workspace topology", Status: "Pending"},
				{ID: "node-2", TaskName: "Resolve dependency graph drift", Status: "Pending"},
				{ID: "node-3", TaskName: "Validate compose containers", Status: "Pending"},
			},
			Edges: []ExecutionEdge{
				{From: "node-1", To: "node-2"},
				{From: "node-1", To: "node-3"},
			},
		}
	}

	if len(optCtx.Incidents) > 0 {
		plan.Confidence -= 12
		plan.Risks = append(plan.Risks, "Active context alert: workspace health compromised.")
	}

	return plan, nil
}
