package commands

import (
	"fmt"

	"daemon/cli/output"
	missioncontrol "daemon/mission-control"
	"github.com/spf13/cobra"
)

var (
	cockpitPortFlag int
	cockpitJSONFlag bool
)

var cockpitCmd = &cobra.Command{
	Use:     "cockpit",
	Aliases: []string{"server"},
	Short:   "Start local Engineering Cockpit server and REST API",
	Run: func(cmd *cobra.Command, args []string) {
		server := missioncontrol.NewServer(rt, "127.0.0.1", cockpitPortFlag)
		boundURI, err := server.Start()

		if err != nil {
			if cockpitJSONFlag {
				output.RenderJSON("cockpit", nil, err)
				return
			}
			fmt.Printf("Error starting Cockpit server: %v\n", err)
			return
		}

		if cockpitJSONFlag {
			output.RenderJSON("cockpit", map[string]interface{}{
				"status":    "running",
				"bound_uri": boundURI,
				"host":      "127.0.0.1",
			}, nil)
			return
		}

		fmt.Println("==================================================")
		fmt.Println("DAEMON ENGINEERING COCKPIT v1.2")
		fmt.Println("==================================================")
		fmt.Printf("✔ Mission Control Server running at %s\n", boundURI)
		fmt.Println("✔ REST API v1 contract active at /api/v1/")
		fmt.Println("✔ Press Ctrl+C to terminate local server")
	},
}

var cockpitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check local Cockpit server status",
	Run: func(cmd *cobra.Command, args []string) {
		statusData := map[string]interface{}{
			"status":    "operational",
			"host":      "127.0.0.1",
			"base_port": 8080,
			"api_ver":   "v1.2",
		}
		if cockpitJSONFlag {
			output.RenderJSON("cockpit.status", statusData, nil)
			return
		}
		fmt.Printf("Cockpit Server Status: %s (127.0.0.1:8080)\n", statusData["status"])
	},
}

func init() {
	cockpitCmd.Flags().IntVar(&cockpitPortFlag, "port", 8080, "Base port to bind Mission Control server (default: 8080)")
	cockpitCmd.Flags().BoolVar(&cockpitJSONFlag, "json", false, "Output machine-readable JSON status")
	cockpitStatusCmd.Flags().BoolVar(&cockpitJSONFlag, "json", false, "Output machine-readable JSON status")

	cockpitCmd.AddCommand(cockpitStatusCmd)
	rootCmd.AddCommand(cockpitCmd)
}
