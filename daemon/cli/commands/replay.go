package commands

import (
	"fmt"
	"strings"
	"time"

	engContext "daemon/core/context"
	"daemon/core/replay"

	"github.com/spf13/cobra"
)

var (
	sinceFlag      string
	repositoryFlag string
)

func runReplay(cmd *cobra.Command, args []string) {
	sub := cmd.Name()
	if sub == "replay" {
		sub = ""
		if len(args) > 0 {
			sub = strings.ToLower(args[0])
		}
	}

	duration := 24 * time.Hour
	filterType := ""

	switch sub {
	case "today":
		duration = 12 * time.Hour
	case "yesterday":
		duration = 36 * time.Hour
	case "deployment":
		filterType = "deploy"
	case "incident":
		filterType = "incident"
	case "workspace":
		filterType = "workspace"
	}

	if sinceFlag != "" {
		d, err := time.ParseDuration(sinceFlag)
		if err == nil {
			duration = d
		}
	}

	ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
	re := replay.NewReplayEngine(rt.Container.ResolveEventBus(), ce)
	eventsList, err := re.ReplaySession(duration, repositoryFlag, filterType)
	if err != nil {
		fmt.Printf("Replay Error: %v\n", err)
		return
	}

	fmt.Println("=== DAEMON ENGINEERING REPLAY TIMELINE ===")
	for _, ev := range eventsList {
		fmt.Printf("[%s] Operation: %s\n", ev.Timestamp.Format("15:04:05"), ev.Title)
		fmt.Printf("  Detail: %s\n", ev.Description)
		fmt.Printf("  Impact: %s\n\n", ev.Impact)
	}
}

var replayCmd = &cobra.Command{
	Use:   "replay [subcommand]",
	Short: "Replay engineering history and execution graphs",
	Long:  `Reconstruct and playback engineering workspace activity sessions. Subcommands: today, yesterday, deployment, incident, workspace.`,
	Run:   runReplay,
}

func init() {
	replayCmd.Flags().StringVar(&sinceFlag, "since", "", "Filter events within a relative timeframe (e.g. 2h, 45m)")
	replayCmd.Flags().StringVar(&repositoryFlag, "repository", "", "Filter events matching repository name")

	// Named subcommands
	for _, sub := range []string{"today", "yesterday", "deployment", "incident", "workspace"} {
		s := sub
		replayCmd.AddCommand(&cobra.Command{
			Use:   s,
			Short: fmt.Sprintf("Replay %s engineering events", s),
			Run:   runReplay,
		})
	}

	rootCmd.AddCommand(replayCmd)
}
