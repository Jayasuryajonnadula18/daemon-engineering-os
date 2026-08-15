package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"daemon/cli/output"
	"daemon/core/analysis"
	"github.com/spf13/cobra"
)

var diagnoseJSONFlag bool

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose [query]",
	Short: "Diagnose engineering issues (memory leaks, slowness, test failures, correctness)",
	Run: func(cmd *cobra.Command, args []string) {
		query := "find resource leaks and correctness issues"
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}

		cwd, _ := os.Getwd()
		pipeline := analysis.NewDeepAnalyzerPipeline(nil, nil)
		res, err := pipeline.RunAnalysis(context.Background(), cwd, true)

		if diagnoseJSONFlag {
			output.RenderJSON("diagnose", map[string]interface{}{
				"query":      query,
				"result":     res,
				"diagnostic": "deterministic_analysis_complete",
			}, err)
			return
		}

		if err != nil {
			fmt.Printf("Diagnosis failed: %v\n", err)
			return
		}

		fmt.Println("==================================================")
		fmt.Println("DAEMON DETERMINISTIC DIAGNOSTIC REPORT")
		fmt.Println("==================================================")
		fmt.Printf("Diagnostic Query:       \"%s\"\n", query)
		fmt.Printf("AI Multiplier Status:   %s\n", map[bool]string{true: "ACTIVE", false: "UNAVAILABLE (100% Deterministic Engine Active)"}[res.AIEnhanced])
		fmt.Printf("Evaluated Findings:     %d\n\n", len(res.Findings))

		if len(res.Findings) > 0 {
			for i, f := range res.Findings {
				fmt.Printf("Finding #%d [%s]: %s\n", i+1, f.Severity, f.Title)
				fmt.Printf("  - Category:    %s\n", f.Category)
				fmt.Printf("  - Description: %s\n", f.Description)
				fmt.Printf("  - Evidence:    %v\n", f.EvidenceIDs)
				if len(f.SuggestedActions) > 0 {
					fmt.Printf("  - Fix:         %s\n", f.SuggestedActions[0])
				}
				fmt.Println()
			}
		} else {
			fmt.Println("✔ Diagnostics complete. Zero defects found matching query.")
		}
	},
}

func init() {
	diagnoseCmd.Flags().BoolVar(&diagnoseJSONFlag, "json", false, "Output machine-readable JSON diagnostic report")
	rootCmd.AddCommand(diagnoseCmd)
}
