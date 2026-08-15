package commands

import (
	"fmt"

	"daemon/cli/output"
	"daemon/core/resource"
	"github.com/spf13/cobra"
)

var resourceJSONFlag bool

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Inspect host hardware metrics, resource tier, and Governor budget decisions",
	Run: func(cmd *cobra.Command, args []string) {
		profiler := resource.NewProfiler()
		gov := resource.NewResourceGovernor(profiler, resource.DefaultResourceConfig())

		metrics := profiler.GetMetrics()
		tier := gov.CalculateTier(metrics)
		dec := gov.Evaluate("background_analysis", false)

		data := map[string]interface{}{
			"hardware":          metrics,
			"resource_tier":     tier,
			"governor_decision": dec,
			"model_tier":        gov.SelectModelTier(),
		}

		if resourceJSONFlag {
			output.RenderJSON("resource", data, nil)
			return
		}

		fmt.Println("==================================================")
		fmt.Println("DAEMON RESOURCE GOVERNOR & HARDWARE ADAPTATION")
		fmt.Println("==================================================")
		fmt.Printf("Active Resource Tier:   %s\n", tier)
		fmt.Printf("CPU Cores:              %d\n", metrics.CPUCores)
		fmt.Printf("CPU Utilization:        %.1f%%\n", metrics.CPUUtilization*100)
		fmt.Printf("Available Memory:       %d MB / %d MB\n", metrics.AvailableMemoryMB, metrics.TotalMemoryMB)
		fmt.Printf("GPU Available:          %t\n", metrics.GPUAvailable)
		fmt.Printf("Free Disk:              %d GB\n", metrics.FreeDiskGB)
		fmt.Printf("Model Selection Tier:   %s\n", gov.SelectModelTier())
		fmt.Printf("\nGovernor Decision:      %s (%s)\n", dec.Decision, dec.Reason)
	},
}

func init() {
	resourceCmd.Flags().BoolVar(&resourceJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(resourceCmd)
}
