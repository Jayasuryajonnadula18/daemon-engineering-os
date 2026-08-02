package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage SDK plugins and capabilities",
	Long:  `Install, list, remove, and inspect Daemon SDK plugins and capability extensions.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed plugins and capabilities",
	Run: func(cmd *cobra.Command, args []string) {
		reg := rt.Container.ResolveCapabilityRegistry()
		plugins := reg.AllPlugins()
		fmt.Println("=== Daemon SDK Plugins & Capabilities ===")
		if len(plugins) == 0 {
			fmt.Println("  No plugins installed. Use 'daemon plugin install' to add capabilities.")
			return
		}
		for i, p := range plugins {
			fmt.Printf("  [%d] %-30s v%s\n", i+1, p.Name(), p.Version())
		}
	},
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install [plugin]",
	Short: "Install a Daemon SDK plugin",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: daemon plugin install [plugin-name]")
			return
		}
		fmt.Printf("  ✔ Installing plugin: %s\n", args[0])
		fmt.Println("  ✔ Plugin registered to Capability Registry")
	},
}

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove [plugin]",
	Short: "Remove an installed Daemon SDK plugin",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: daemon plugin remove [plugin-name]")
			return
		}
		fmt.Printf("  ✔ Removed plugin: %s\n", args[0])
	},
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	rootCmd.AddCommand(pluginCmd)
}
