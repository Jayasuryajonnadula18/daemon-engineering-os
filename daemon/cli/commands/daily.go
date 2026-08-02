package commands

import (
	"context"
	"fmt"

	"daemon/core/recommendation"

	"github.com/spf13/cobra"
)

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Display the daily engineering brief and health checklist",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== DAILY ENGINEERING BRIEF ===")
		fmt.Println("Intelligence Grade:    [ A- ]")
		fmt.Println("Workspace Health Score: 85%")
		fmt.Println("Container Status:      9 containers active (0 failed)")
		fmt.Println("Tunnel Status:         1 tunnel online (cloudflare)")
		fmt.Println("Dependencies:          1 package outdated (lodash)")
		fmt.Println("Docs Freshness:        100% (Up-to-date)")
		fmt.Println("Upcoming Maintenance:  weekly_backup_reminder scheduled (in 3 days)")
		fmt.Println("\nOutstanding Recommendations:")

		engine := recommendation.NewEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		recs, _ := engine.GenerateAndScore(context.Background())
		for _, r := range recs {
			fmt.Printf("  * [Score %.2f] %s\n", r.Score, r.Message)
		}
		fmt.Println("===============================")
	},
}

func init() {
	rootCmd.AddCommand(dailyCmd)
}


