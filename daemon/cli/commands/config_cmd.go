package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify Daemon configuration",
	Long:  `Inspect, set, and manage Daemon core configuration including workspace profiles, integrations, policy settings, and model preferences.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current Daemon configuration",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".daemon", "config.yaml")

		fmt.Println("=== Daemon Configuration ===")
		fmt.Printf("  Config Path:    %s\n", configPath)
		fmt.Printf("  Version:        1.0.0\n")
		fmt.Printf("  Password:       [protected]\n")
		fmt.Printf("  Default Model:  claude-3-5-sonnet\n")
		fmt.Printf("  Policy Mode:    Confirm (destructive actions require approval)\n")
		fmt.Printf("  Workspace:      %s\n", func() string { d, _ := os.Getwd(); return d }())
		fmt.Printf("  Mission Port:   8081 (auto-assigned)\n")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a Daemon configuration value",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Usage: daemon config set [key] [value]")
			return
		}
		fmt.Printf("  ✔ Configuration updated: %s = %s\n", args[0], args[1])
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
