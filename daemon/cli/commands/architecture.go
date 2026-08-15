package commands

import (
	"context"
	"fmt"

	"daemon/cli/output"
	"daemon/core/architecture"

	"github.com/spf13/cobra"
)

var architectureJSONFlag bool

var architectureCmd = &cobra.Command{
	Use:   "architecture",
	Short: "Generate complete engineering architecture report",
	Run: func(cmd *cobra.Command, args []string) {
		engine := architecture.NewEngine(rt.Container.ResolveGraphStore())
		report, err := engine.Analyze(context.Background())

		if architectureJSONFlag {
			output.RenderJSON("architecture", report, err)
			return
		}

		if err != nil {
			fmt.Printf("Error analyzing architecture: %v\n", err)
			return
		}

		fmt.Println("=== DAEMON ARCHITECTURE REPORT ===")
		fmt.Printf("Detected Style:     %s\n", report.Style)
		fmt.Printf("Architecture Score: %d%%\n", report.ArchitectureScore)
		fmt.Printf("Coupling Index:     %d%%\n", report.CouplingScore)
		fmt.Printf("Cohesion Index:     %d%%\n", report.CohesionScore)
		fmt.Printf("Scalability Index:  %d%%\n", report.ScalabilityScore)
		fmt.Printf("Reliability Index:  %d%%\n", report.ReliabilityScore)
		fmt.Printf("Complexity Index:   %d%%\n", report.ComplexityScore)
		fmt.Println("----------------------------------")

		if len(report.Issues) > 0 {
			fmt.Println("Structural Bottlenecks/Issues:")
			for _, iss := range report.Issues {
				fmt.Printf("  - %s\n", iss)
			}
		}

		if len(report.Recommendations) > 0 {
			fmt.Println("\nRecommended Improvements:")
			for _, rec := range report.Recommendations {
				fmt.Printf("  * %s\n", rec)
			}
		}
	},
}

func init() {
	architectureCmd.Flags().BoolVar(&architectureJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(architectureCmd)
}


