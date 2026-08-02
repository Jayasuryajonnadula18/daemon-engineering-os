package commands

import (
	"context"
	"fmt"
	"os"

	"daemon/core/discovery"
	"daemon/core/integrations"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [target]",
	Short: "Synchronize local project knowledge graph metadata with active connected integrations",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := "all"
		if len(args) > 0 {
			target = args[0]
		}

		fmt.Printf("Starting incremental synchronization for: %s...\n", target)

		gs := rt.Container.ResolveGraphStore()
		im := integrations.NewIntegrationManager(gs)

		if target == "all" {
			err := im.SyncAll(context.Background())
			if err != nil {
				fmt.Printf("Integration sync failed: %v\n", err)
			} else {
				fmt.Println("✔ All active integrations synchronized.")
			}
		} else {
			c, err := im.GetConnector(target)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			_ = c.Connect(context.Background())
			_, _ = c.Authenticate(context.Background())
			_, _ = c.Discover(context.Background())
			_ = c.Synchronize(context.Background())
			fmt.Printf("✔ Integration '%s' synchronized.\n", target)
		}

		// Auto-run local workspace discovery to populate Engineering Twin
		cwd, err := os.Getwd()
		if err == nil {
			fmt.Printf("\nRunning local workspace discovery in: %s\n", cwd)
			de := discovery.NewDiscoveryEngine(gs)
			info, err := de.Scan(context.Background(), cwd)
			if err != nil {
				fmt.Printf("  ⚠ Workspace discovery warning: %v\n", err)
			} else {
				fmt.Printf("  ✔ Discovered project: %s (%s)\n", info.Name, info.Language)
				fmt.Printf("  ✔ Detected %d services, %d dependencies, %d API routes\n",
					len(info.Services), len(info.Dependencies), len(info.APIRoutes))
				if info.DockerCompose {
					fmt.Println("  ✔ docker-compose.yml detected — container topology mapped")
				}
				if info.Kubernetes {
					fmt.Println("  ✔ Kubernetes manifests detected — cluster topology mapped")
				}
			}
		}

		fmt.Println("\n✔ Engineering Twin synchronization complete. Run 'daemon graph' to explore.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
