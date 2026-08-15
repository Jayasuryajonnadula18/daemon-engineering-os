package commands

import (
	"context"
	"fmt"

	"daemon/cli/output"
	engContext "daemon/core/context"
	"daemon/core/reasoning"
	"daemon/core/reasoning/evaluation"
	"github.com/spf13/cobra"
)

var evalJSONFlag bool

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Run 50-scenario intelligence benchmark suite evaluating Daemon vs generic LLM baselines",
	Run: func(cmd *cobra.Command, args []string) {
		ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		engReasoner := reasoning.NewEngineeringReasoner(ce)
		evaluator := evaluation.NewIntelligenceEvaluator(engReasoner)

		scenarios := evaluation.Generate50BenchmarkScenarios()
		report, err := evaluator.EvaluateBenchmark(context.Background(), scenarios)

		if evalJSONFlag {
			output.RenderJSON("eval", report, err)
			return
		}

		if err != nil {
			fmt.Printf("Evaluation failed: %v\n", err)
			return
		}

		fmt.Println("==================================================")
		fmt.Println("DAEMON v1.0 INTELLIGENCE BENCHMARK REPORT (50 SCENARIOS)")
		fmt.Println("==================================================")
		fmt.Printf("Total Benchmark Cases:  %d\n", report.TotalCases)
		fmt.Printf("Passed Cases:           %d\n", report.PassedCases)
		fmt.Printf("Overall Eval Score:     %.1f%%\n\n", report.OverallEvalScore)

		fmt.Println("Category Breakdown:")
		for _, cat := range report.CategoryBreakdown {
			fmt.Printf("  - %-32s %2d/%2d passed (%.1f%%)\n", cat.Category, cat.PassedCases, cat.TotalCases, cat.Score)
		}

		fmt.Println("\nDaemon vs Generic LLM Advantage:")
		fmt.Printf("  %s\n", report.DaemonVsLLMAdvantage)
	},
}

func init() {
	evalCmd.Flags().BoolVar(&evalJSONFlag, "json", false, "Output machine-readable JSON evaluation report")
	rootCmd.AddCommand(evalCmd)
}
