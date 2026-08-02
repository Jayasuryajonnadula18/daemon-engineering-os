package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect engineering environment and project context",
	Run: func(cmd *cobra.Command, args []string) {
		gs := rt.Container.ResolveGraphStore()
		ms := rt.Container.ResolveMemoryStore()

		fmt.Println("=== DAEMON ENGINEERING INSPECT REPORT ===")

		projects, err := gs.GetNodes("project")
		if err != nil || len(projects) == 0 {
			fmt.Println("No initialized project found. Please run 'daemon init' first.")
			os.Exit(1)
		}

		proj := projects[0]
		fmt.Printf("Project Name:      %s\n", proj.Name)
		fmt.Printf("Language Stack:    %s\n", proj.Type)

		services, _ := gs.GetServices()
		fmt.Println("\nDiscovered Services:")
		if len(services) > 0 {
			for _, s := range services {
				fmt.Printf("  - %s (Port: %d) -> status: %s\n", s.Name, s.Port, s.Status)
			}
		} else {
			fmt.Println("  No active backend services discovered.")
		}

		dependencies, _ := gs.GetDependencies()
		fmt.Printf("\nDependencies:      %d indexed packages\n", len(dependencies))

		incidents, _ := ms.GetIncidents()
		recs, _ := ms.GetRecommendations()

		healthScore := 100 - len(incidents)*15
		if healthScore < 0 {
			healthScore = 0
		}

		fmt.Printf("\nEngineering Health: %d%%\n", healthScore)
		fmt.Printf("Active Incidents:   %d\n", len(incidents))
		fmt.Printf("Recommendations:    %d available\n", len(recs))
	},
}


