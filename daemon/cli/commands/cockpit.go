package commands

import (
	"fmt"
	"os"

	"daemon/cli/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var cockpitCmd = &cobra.Command{
	Use:   "cockpit",
	Short: "Launch the Engineering Cockpit (keyboard-first TUI)",
	Long:  `Open the full-screen Engineering Cockpit — the keyboard-first interface for searching, inspecting, automating, and operating the entire engineering workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(tui.NewModel(rt), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting Engineering Cockpit: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cockpitCmd)
}
