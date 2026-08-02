package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	daemonVersion   = "1.0.0"
	daemonBuild     = "RC3"
	daemonCodename  = "Engineering OS"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display Daemon version and build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Daemon Engineering OS\n")
		fmt.Printf("  Version:    %s (%s)\n", daemonVersion, daemonBuild)
		fmt.Printf("  Codename:   %s\n", daemonCodename)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  Platform:   Engineering Operating System\n")
		fmt.Printf("\nPillars Active: 24 (Engineering Context, Twin, Knowledge Graph, Automation,\n")
		fmt.Printf("  Integration, Workspace, Cockpit, Mission Control, SDK, Search, Memory,\n")
		fmt.Printf("  Timeline, Recommendation, Architecture Intelligence, Risk Engine,\n")
		fmt.Printf("  Context Engine, Workflow Engine, Advisor, Replay, Orchestrator,\n")
		fmt.Printf("  Policy Engine, Fix Engine, Deploy Engine, Maintenance Engine)\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
