package commands

import (
	"fmt"
	"strings"

	"daemon/cli/output"
	"daemon/core/orchestration"

	"github.com/spf13/cobra"
)

var planJSONFlag bool

var planCmd = &cobra.Command{
	Use:   "plan [intent]",
	Short: "Generate deterministic dry-run plans from developer intent",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		intentText := strings.Join(args, " ")

		compiler := orchestration.NewDAGCompiler()
		dag, err := compiler.Compile(orchestration.ExecutionIntent{
			Objective: intentText,
			Targets:   []string{"workspace"},
		})

		if planJSONFlag {
			output.RenderJSON("plan", dag, err)
			return
		}

		if err != nil {
			fmt.Printf("Orchestrator Error: %v\n", err)
			return
		}

		fmt.Println("=== DAEMON ENGINEERING ORCHESTRATOR ===")
		fmt.Printf("Developer Intent: %s\n", dag.Intent)
		fmt.Printf("Plan Hash:        %s\n", dag.Freshness.PlanHash)
		fmt.Printf("DAG State:        %s\n\n", dag.State)

		fmt.Println("DAG Execution Graph Nodes:")
		for _, n := range dag.Nodes {
			fmt.Printf("  * [%s] Capability: %s (Status: %s, Risk: %s)\n", n.ID, n.CapabilityName, n.Status, n.RiskLevel)
		}
	},
}

func init() {
	planCmd.Flags().BoolVar(&planJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(planCmd)
}
