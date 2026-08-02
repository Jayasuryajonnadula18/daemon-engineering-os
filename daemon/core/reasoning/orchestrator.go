package reasoning

import (
	"context"
	"fmt"
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

// EngineeringOrchestrator coordinates semantic task mapping.
type EngineeringOrchestrator struct {
	contextBuilder *ContextBuilder
	modelRouter    *ModelRouter
}

// NewEngineeringOrchestrator instantiates an EngineeringOrchestrator.
func NewEngineeringOrchestrator(cb *ContextBuilder, mr *ModelRouter) *EngineeringOrchestrator {
	return &EngineeringOrchestrator{
		contextBuilder: cb,
		modelRouter:    mr,
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

	var nodes []ExecutionNode
	var edges []ExecutionEdge
	var risks []string

	// DYNAMIC DAG COMPILATION: Use real workspace services from optCtx
	nodeCounter := 1

	// Step 1: Add workspace discovery node
	step1ID := fmt.Sprintf("node-%d", nodeCounter)
	nodes = append(nodes, ExecutionNode{
		ID:       step1ID,
		TaskName: fmt.Sprintf("Verify workspace context & compilation (%d services tracked)", len(optCtx.Services)),
		Status:   "Completed",
	})
	nodeCounter++

	// Step 2: Add dynamic service-specific tasks if services exist
	if len(optCtx.Services) > 0 {
		var prevID string = step1ID

		for _, s := range optCtx.Services {
			sID := fmt.Sprintf("node-%d", nodeCounter)
			taskLabel := fmt.Sprintf("Verify %s service dependencies & port binding", s.Name)
			if domain == "Deployment" {
				taskLabel = fmt.Sprintf("Build & deploy %s container image", s.Name)
			}
			nodes = append(nodes, ExecutionNode{
				ID:       sID,
				TaskName: taskLabel,
				Status:   "Pending",
			})
			edges = append(edges, ExecutionEdge{From: prevID, To: sID})
			prevID = sID
			nodeCounter++
		}

		// Step 3: Add final verification node
		finalID := fmt.Sprintf("node-%d", nodeCounter)
		nodes = append(nodes, ExecutionNode{
			ID:       finalID,
			TaskName: fmt.Sprintf("Verify telemetry probes & gateway routing (%s strategy)", domain),
			Status:   "Pending",
		})
		edges = append(edges, ExecutionEdge{From: prevID, To: finalID})
	} else {
		// Fallback for empty graph
		step2ID := fmt.Sprintf("node-%d", nodeCounter)
		nodes = append(nodes, ExecutionNode{
			ID:       step2ID,
			TaskName: "Scan project workspace directory and run 'daemon sync'",
			Status:   "Pending",
		})
		edges = append(edges, ExecutionEdge{From: step1ID, To: step2ID})
	}

	// Dynamic Risks
	if domain == "Deployment" {
		risks = append(risks, "Live deployment risk: requires interactive stdin confirmation for staging/prod target")
		if len(optCtx.Services) > 1 {
			risks = append(risks, fmt.Sprintf("Coupling risk: updating %d microservices simultaneously", len(optCtx.Services)))
		}
	} else {
		risks = append(risks, fmt.Sprintf("Workspace drift risk: %d tracked dependencies across project manifests", len(optCtx.Dependencies)))
	}

	if len(optCtx.Incidents) > 0 {
		risks = append(risks, "Active incident alert: workspace health compromised")
	}

	return &OrchestrationPlan{
		Intent:           intent,
		Graph:            ExecutionGraph{Nodes: nodes, Edges: edges},
		Confidence:       92,
		Risks:            risks,
		RollbackStrategy: "Stash modified local configurations, revert fix snapshot via 'daemon fix --rollback'",
		RequiresApproval: domain == "Deployment" || domain == "Security",
		ModelUsed:        recModel.ModelName,
		Domain:           domain,
	}, nil
}
