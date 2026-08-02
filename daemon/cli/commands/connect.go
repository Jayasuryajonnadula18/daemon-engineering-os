package commands

import (
	"context"
	"fmt"
	"strings"

	"daemon/core/integrations"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [provider]",
	Short: "Establish an integration session connection to a developer service provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := strings.ToLower(args[0])
		im := integrations.NewIntegrationManager(rt.Container.ResolveGraphStore())
		c, err := im.GetConnector(provider)
		if err != nil {
			fmt.Printf("Error resolving connector for provider '%s': %v\n", provider, err)
			return
		}

		fmt.Printf("Connecting to %s...\n", provider)
		err = c.Connect(context.Background())
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			return
		}

		fmt.Printf("✔ Successfully connected to %s.\n", provider)
	},
}

var disconnectCmd = &cobra.Command{
	Use:   "disconnect [provider]",
	Short: "Disconnect an active integration session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := strings.ToLower(args[0])
		im := integrations.NewIntegrationManager(rt.Container.ResolveGraphStore())
		c, err := im.GetConnector(provider)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		_ = c.Disconnect(context.Background())
		fmt.Printf("✔ Disconnected from %s.\n", provider)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(disconnectCmd)
}
