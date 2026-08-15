package commands

import (
	"context"
	"fmt"

	"daemon/cli/output"
	coreContext "daemon/core/context"
	"daemon/core/reasoning"
	"github.com/spf13/cobra"
)

var askJSONFlag bool

var askCmd = &cobra.Command{
	Use:   "ask <query>",
	Short: "Perform evidence-grounded AI reasoning on project state",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		gs := rt.Container.ResolveGraphStore()
		ms := rt.Container.ResolveMemoryStore()
		ce := coreContext.NewContextEngine(gs, ms)
		engReasoner := reasoning.NewEngineeringReasoner(ce)

		res, err := engReasoner.Reason(context.Background(), query)

		if askJSONFlag {
			output.RenderJSON("ask", res, err)
			return
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("\n=========================================")
		fmt.Println("DAEMON EVIDENCE-GROUNDED REASONING")
		fmt.Println("=========================================")
		fmt.Printf("Query:            %s\n", query)
		fmt.Printf("Daemon Confidence: %.0f%%\n", res.Confidence*100)
		fmt.Printf("Insufficient:     %t\n\n", res.InsufficientContext)
		fmt.Println(res.Answer)

		if len(res.Facts) > 0 {
			fmt.Println("\nObserved Facts:")
			for _, f := range res.Facts {
				fmt.Printf("  • [FACT] %s (Evidence IDs: %v)\n", f.Statement, f.EvidenceIDs)
			}
		}

		if len(res.Inferences) > 0 {
			fmt.Println("\nDerived Inferences:")
			for _, inf := range res.Inferences {
				fmt.Printf("  • [INFERENCE] %s (Confidence: %.0f%%)\n", inf.Statement, inf.Confidence*100)
			}
		}

		if len(res.Unknowns) > 0 {
			fmt.Println("\nIdentified Context Unknowns:")
			for _, u := range res.Unknowns {
				fmt.Printf("  • [UNKNOWN] %s\n", u)
			}
		}
	},
}

func init() {
	askCmd.Flags().BoolVar(&askJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(askCmd)
}
