package commands

import (
	"context"
	"fmt"
	"strings"

	"daemon/cli/output"
	"daemon/core/orchestration"
	"github.com/spf13/cobra"
)

var impactJSONFlag bool

var impactCmd = &cobra.Command{
	Use:   "impact [node]",
	Short: "Analyze the dependency impact of changing a specific component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeTarget := strings.ToLower(strings.TrimSpace(args[0]))

		impactEng := orchestration.NewImpactEngine(nil)
		analysis, err := impactEng.AnalyzeImpact(context.Background(), nodeTarget)

		if impactJSONFlag {
			output.RenderJSON("impact", analysis, err)
			return
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("=== DEPENDENCY IMPACT ANALYSIS: %s ===\n\n", strings.ToUpper(nodeTarget))
		fmt.Printf("Target Entity:        %s\n", analysis.TargetEntity)
		fmt.Printf("Blast Radius Score:   %.0f/100\n", analysis.BlastRadiusScore)
		fmt.Printf("Calculated Risk:      %s\n", analysis.RiskLevel)
		fmt.Printf("Affected Services:    %v\n", analysis.AffectedServices)
		fmt.Printf("Single Points Fail:   %v\n", analysis.SinglePointsOfFailure)
	},
}

func init() {
	impactCmd.Flags().BoolVar(&impactJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(impactCmd)
}
