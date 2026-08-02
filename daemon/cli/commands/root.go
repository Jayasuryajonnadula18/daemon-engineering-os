package commands

import (
	"fmt"
	"os"

	"daemon/core/runtime"
	"daemon/core/security"
	"daemon/cli/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	rt *runtime.Runtime
)

var rootCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Daemon is an Engineering Operating System for developer platforms.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "token" {
			return
		}
		providedPassword := os.Getenv("DAEMON_PASSWORD")
		if providedPassword == "" {
			providedPassword, _ = cmd.Flags().GetString("password")
		}

		if !security.ValidateMasterSecret(providedPassword) {
			fmt.Println("Error: Unauthorized. Daemon requires a valid master secret token stored in your OS Keyring.")
			fmt.Println("Provide your token via the DAEMON_PASSWORD environment variable or --password flag.")
			fmt.Println("To view your OS Keyring master token, run 'daemon token' or inspect your OS credential manager (Service: DaemonEngineeringOS).")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Run full-screen Engineering Cockpit TUI by default
		p := tea.NewProgram(tui.NewModel(rt), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting Engineering Cockpit TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute binds the global Runtime instance and executes command routing.
func Execute(runtimeInstance *runtime.Runtime) {
	rt = runtimeInstance
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error routing command: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("password", "", "Master password for access authorization")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(doctorCmd)
}


