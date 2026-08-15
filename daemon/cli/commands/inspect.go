package commands

import (
	"fmt"
	"os"

	"daemon/cli/output"
	"github.com/spf13/cobra"
)

var inspectJSONFlag bool

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect engineering environment and project context",
	Run: func(cmd *cobra.Command, args []string) {
		gs := rt.Container.ResolveGraphStore()
		ms := rt.Container.ResolveMemoryStore()

		projects, err := gs.GetNodes("project")
		if err != nil || len(projects) == 0 {
			if inspectJSONFlag {
				output.RenderJSON("inspect", nil, fmt.Errorf("no initialized project found"))
				return
			}
			fmt.Println("No initialized project found. Please run 'daemon init' first.")
			os.Exit(1)
		}

		proj := projects[0]
		services, _ := gs.GetServices()
		dependencies, _ := gs.GetDependencies()
		incidents, _ := ms.GetIncidents()
		recs, _ := ms.GetRecommendations()

		healthScore := 100 - len(incidents)*15
		if healthScore < 0 {
			healthScore = 0
		}

		data := map[string]interface{}{
			"project_name":     proj.Name,
			"language_stack":   proj.Type,
			"services":         services,
			"dependencies":     len(dependencies),
			"health_score":     healthScore,
			"active_incidents": len(incidents),
			"recommendations":  len(recs),
		}

		if inspectJSONFlag {
			output.RenderJSON("inspect", data, nil)
			return
		}

		fmt.Println("=== DAEMON ENGINEERING INSPECT REPORT ===")
		fmt.Printf("Project Name:      %s\n", proj.Name)
		fmt.Printf("Language Stack:    %s\n", proj.Type)
		fmt.Println("\nDiscovered Services:")
		if len(services) > 0 {
			for _, s := range services {
				fmt.Printf("  - %s (Port: %d) -> status: %s\n", s.Name, s.Port, s.Status)
			}
		} else {
			fmt.Println("  No active backend services discovered.")
		}

		fmt.Printf("\nDependencies:      %d indexed packages\n", len(dependencies))
		fmt.Printf("\nEngineering Health: %d%%\n", healthScore)
		fmt.Printf("Active Incidents:   %d\n", len(incidents))
		fmt.Printf("Recommendations:    %d available\n", len(recs))
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(inspectCmd)
}


