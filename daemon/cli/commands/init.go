package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"daemon/core/discovery"
	"daemon/core/events"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Daemon in the current project repository",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error resolving working directory: %v\n", err)
			os.Exit(1)
		}

		daemonDir := filepath.Join(cwd, ".daemon")
		if err := os.MkdirAll(daemonDir, 0755); err != nil {
			fmt.Printf("Error creating .daemon workspace folder: %v\n", err)
			os.Exit(1)
		}

		configPath := filepath.Join(daemonDir, "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configJSON := `{
  "ai": {
    "provider": "claude",
    "model": "claude-3-5-sonnet"
  },
  "policies": {
    "read_only": false
  }
}`
			_ = os.WriteFile(configPath, []byte(configJSON), 0644)
		}

		fmt.Println("✔ Initializing Daemon Runtime Environment...")

		gs := rt.Container.ResolveGraphStore()
		eb := rt.Container.ResolveEventBus()

		scanner := discovery.NewDiscoveryEngine(gs)
		info, err := scanner.Scan(context.Background(), cwd)
		if err != nil {
			fmt.Printf("Error parsing repository stack: %v\n", err)
			os.Exit(1)
		}

		eb.Publish(events.Event{
			Type:      "RepositoryScanned",
			Payload:   info,
			Timestamp: time.Now(),
		})

		fmt.Printf("✔ Repository detected: %s\n", info.Name)
		if info.Framework != "" {
			fmt.Printf("✔ Framework identified: %s\n", info.Framework)
		}
		if len(info.Services) > 0 {
			fmt.Printf("✔ Services mapped: %d active services\n", len(info.Services))
		}
		fmt.Printf("✔ Dependencies indexed: %d packages\n", len(info.Dependencies))
		fmt.Println("✔ Engineering Knowledge Graph created successfully in '.daemon/graph.db'")
		fmt.Println("✔ Daemon initialized.")
	},
}


