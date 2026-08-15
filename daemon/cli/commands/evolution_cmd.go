package commands

import (
	"fmt"

	"daemon/cli/output"
	"daemon/core/evolution"
	"github.com/spf13/cobra"
)

var evolutionJSONFlag bool

var evolutionCmd = &cobra.Command{
	Use:   "evolution",
	Short: "Inspect continuous evolution patterns, confidence scores, and Fix Ledger entries",
	Run: func(cmd *cobra.Command, args []string) {
		ledger, _ := evolution.NewFixLedger(":memory:")
		engine := evolution.NewEvolutionEngine(ledger, evolution.DefaultPromotionConfig())

		patterns := engine.GetPatterns()
		entries, _ := ledger.GetEntries()

		data := map[string]interface{}{
			"patterns":    patterns,
			"fix_ledger":  entries,
			"total_rules": len(patterns),
		}

		if evolutionJSONFlag {
			output.RenderJSON("evolution", data, nil)
			return
		}

		fmt.Println("==================================================")
		fmt.Println("DAEMON CONTINUOUS EVOLUTION & FIX LEDGER")
		fmt.Println("==================================================")
		fmt.Printf("Active Evolution Patterns: %d\n", len(patterns))
		fmt.Printf("Fix Ledger Entries:       %d\n", len(entries))

		if len(patterns) == 0 {
			fmt.Println("\nNo learned patterns accumulated yet. Execute workflows to build experience.")
		}
	},
}

func init() {
	evolutionCmd.Flags().BoolVar(&evolutionJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(evolutionCmd)
}
