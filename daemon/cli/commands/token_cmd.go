package commands

import (
	"fmt"

	"daemon/core/security"

	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Display or initialize OS Keyring master token for authentication",
	Long:  `Retrieve the high-entropy master secret token securely stored in your OS Keyring (Windows Credential Manager / macOS Keychain / Linux Secret Service).`,
	Run: func(cmd *cobra.Command, args []string) {
		secret, err := security.GetOrGenerateMasterSecret()
		if err != nil {
			fmt.Printf("Error accessing OS Keyring: %v\n", err)
			return
		}
		fmt.Println("=== Daemon Master Secret Token (OS Keyring) ===")
		fmt.Printf("Service: %s\n", security.ServiceName)
		fmt.Printf("User:    %s\n", security.SecretUser)
		fmt.Printf("Token:   %s\n\n", secret)
		fmt.Println("Set in environment: $env:DAEMON_PASSWORD = \"" + secret + "\"")
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}
