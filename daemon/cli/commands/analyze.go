package commands

import (
	"context"
	"fmt"
	"os"

	"daemon/cli/output"
	"daemon/core/analysis"
	"github.com/spf13/cobra"
)

var (
	analyzeDeepFlag    bool
	analyzeChangedFlag bool
	analyzeJSONFlag    bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Run deep program analysis (AST, call graphs, resource lifetime, correctness)",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		pipeline := analysis.NewDeepAnalyzerPipeline(nil, nil)
		res, err := pipeline.RunAnalysis(context.Background(), cwd, analyzeDeepFlag)

		if analyzeJSONFlag {
			output.RenderJSON("analyze", res, err)
			return
		}

		if err != nil {
			fmt.Printf("Analysis failed: %v\n", err)
			return
		}

		fmt.Println("==================================================")
		fmt.Println("DAEMON DEEP ENGINEERING ANALYSIS REPORT")
		fmt.Println("==================================================")
		fmt.Printf("Analyzed Source Files:  %d\n", res.AnalyzedFiles)
		fmt.Printf("AI Enhancement Active:  %t\n", res.AIEnhanced)
		fmt.Printf("Overall Confidence:     %.1f%%\n", res.Confidence*100)
		fmt.Printf("Extracted Findings:     %d\n\n", len(res.Findings))

		if len(res.Findings) > 0 {
			fmt.Println("Discovered Defects & Findings:")
			for _, f := range res.Findings {
				fmt.Printf("  [%s] %s (%s)\n", f.Severity, f.Title, f.Category)
				fmt.Printf("    -> %s\n", f.Description)
				if len(f.SuggestedActions) > 0 {
					fmt.Printf("    -> Fix: %s\n", f.SuggestedActions[0])
				}
				fmt.Println()
			}
		} else {
			fmt.Println("✔ Zero critical resource leaks or correctness defects detected.")
		}

		fmt.Println("Analyzer Pipeline Status:")
		for _, s := range res.AnalyzerStatus {
			fmt.Printf("  - %-26s [%s] %s\n", s.Name, s.RunTime, s.Message)
		}
	},
}

var analyzeReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate summary report of engineering findings",
	Run: func(cmd *cobra.Command, args []string) {
		analyzeCmd.Run(cmd, args)
	},
}

func init() {
	analyzeCmd.Flags().BoolVar(&analyzeDeepFlag, "deep", false, "Run progressive 20-phase deep program analysis")
	analyzeCmd.Flags().BoolVar(&analyzeChangedFlag, "changed", false, "Run incremental analysis on git changed files only")
	analyzeCmd.Flags().BoolVar(&analyzeJSONFlag, "json", false, "Output machine-readable JSON analysis report")
	analyzeReportCmd.Flags().BoolVar(&analyzeJSONFlag, "json", false, "Output machine-readable JSON analysis report")

	analyzeCmd.AddCommand(analyzeReportCmd)
	rootCmd.AddCommand(analyzeCmd)
}
