package commands

import (
	"context"
	"fmt"
	"strings"

	engContext "daemon/core/context"
	"daemon/core/reasoning"

	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan [intent]",
	Short: "Generate deterministic dry-run plans from developer intent",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		intent := strings.Join(args, " ")

		ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		cb := reasoning.NewContextBuilder(ce)
		mr := reasoning.NewModelRouter(false)
		orch := reasoning.NewEngineeringOrchestrator(cb, mr)

		plan, err := orch.Orchestrate(context.Background(), intent)
		if err != nil {
			fmt.Printf("Orchestrator Error: %v\n", err)
			return
		}

		fmt.Println("=== DAEMON ENGINEERING ORCHESTRATOR ===")
		fmt.Printf("Developer Intent: %s\n", plan.Intent)
		fmt.Printf("Orchestrator:     %s Domain Orchestrator\n", plan.Domain)
		fmt.Printf("Model Selected:   %s\n", plan.ModelUsed)
		fmt.Printf("Confidence level: %d%%\n", plan.Confidence)
		fmt.Printf("Requires Approval: %t\n\n", plan.RequiresApproval)

		fmt.Println("DAG Execution Graph Nodes:")
		for _, n := range plan.Graph.Nodes {
			fmt.Printf("  * [%s] %s (Status: %s)\n", n.ID, n.TaskName, n.Status)
		}

		fmt.Println("\nDAG Dependency Edges (Precedence Rules):")
		for _, e := range plan.Graph.Edges {
			fmt.Printf("  - [%s] ──> [%s]\n", e.From, e.To)
		}

		fmt.Println("\nPredicted Risks:")
		for _, r := range plan.Risks {
			fmt.Printf("  ⚠ %s\n", r)
		}

		fmt.Printf("\nRollback Strategy:\n  %s\n", plan.RollbackStrategy)
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
