package commands

import (
	"fmt"
	"net"
	"os"

	"daemon/mission-control"

	"github.com/spf13/cobra"
)

var missionControlCmd = &cobra.Command{
	Use:     "mission",
	Aliases: []string{"dashboard"},
	Short:   "Launch Mission Control (web dashboard)",
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		var listener net.Listener
		var err error

		addr := net.JoinHostPort("127.0.0.1", port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			for p := 8081; p <= 8095; p++ {
				port = fmt.Sprintf("%d", p)
				addr = net.JoinHostPort("127.0.0.1", port)
				listener, err = net.Listen("tcp", addr)
				if err == nil {
					break
				}
			}
		}

		if err != nil {
			fmt.Printf("Error starting Mission Control server: no free ports found\n")
			os.Exit(1)
		}
		_ = listener.Close()

		fmt.Printf("✔ Starting Mission Control Web Server on http://%s...\n", addr)

		server := missioncontrol.NewServer(rt, addr)
		if err := server.Start(); err != nil {
			fmt.Printf("Error starting Mission Control server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(missionControlCmd)
	// 'daemon dashboard' remains a valid backwards-compatible alias via Aliases field above
}
