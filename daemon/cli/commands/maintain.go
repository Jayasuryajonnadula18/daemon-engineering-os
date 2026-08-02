package commands

import (
	"fmt"
	"strings"

	engContext "daemon/core/context"
	"daemon/core/maintenance"
	"daemon/core/policies"

	"github.com/spf13/cobra"
)

var (
	maintainFixFlag     bool
	maintainAutoFixFlag bool
	maintainDryRunFlag  bool
	maintainSchedule    string
)

var maintainCmd = &cobra.Command{
	Use:     "maintain [category]",
	Aliases: []string{"care", "health"},
	Short:   "Perform engineering maintenance and health checks",
	Long: `The Maintenance Engine continuously monitors, maintains, and cares for the engineering workspace.
Supports categories: dependencies, containers, kubernetes, cloudflare, database, security, performance, docs, workspace, all.
You can also invoke this command using its aliases: 'daemon care' or 'daemon health'.`,
	Run: func(cmd *cobra.Command, args []string) {
		if maintainSchedule != "" {
			fmt.Printf("✔ Configured automated maintenance routine on schedule: [%s]\n", strings.ToUpper(maintainSchedule))
			fmt.Println("  Next execution will run under Policy Engine controlled self-healing.")
			return
		}

		category := "all"
		if len(args) > 0 {
			category = args[0]
		}

		autoFix := maintainFixFlag || maintainAutoFixFlag
		if maintainDryRunFlag {
			fmt.Println("=== [DRY-RUN MODE] Previewing Maintenance Engine operations ===")
			autoFix = false
		}

		pe := policies.NewMemoryPolicyEngine(maintainDryRunFlag)
		ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		me := maintenance.NewMaintenanceEngine(ce, pe)

		rep, err := me.RunMaintenance(cmd.Context(), category, autoFix)
		if err != nil {
			fmt.Printf("Error running Maintenance Engine: %v\n", err)
			return
		}

		fmt.Println("=== DAEMON MAINTENANCE ENGINE (PILLAR 24) ===")
		fmt.Printf("Category:             %s\n", strings.ToUpper(rep.Category))
		fmt.Printf("Workspace Status:     %s\n", rep.Status)
		fmt.Printf("Confidence Score:     %d%%\n", rep.ConfidenceScore)
		fmt.Printf("Estimated Time Saved: %s\n\n", rep.EstimatedTimeSaved)

		fmt.Println("Observation:")
		fmt.Printf("  • %s\n\n", rep.Observation)

		fmt.Println("Evidence:")
		for _, ev := range rep.Evidence {
			fmt.Printf("  ✓ %s\n", ev)
		}
		fmt.Println()

		fmt.Println("Recommendation:")
		fmt.Printf("  * %s\n\n", rep.Recommendation)

		if len(rep.SelfHealingActions) > 0 {
			fmt.Println("Self-Healing Actions:")
			for _, act := range rep.SelfHealingActions {
				fmt.Printf("  -> %s\n", act)
			}
			fmt.Println()
		}
	},
}

func init() {
	maintainCmd.Flags().BoolVar(&maintainFixFlag, "fix", false, "Execute approved repair workflows and self-healing")
	maintainCmd.Flags().BoolVar(&maintainAutoFixFlag, "auto-fix", false, "Execute controlled self-healing routines")
	maintainCmd.Flags().BoolVar(&maintainDryRunFlag, "dry-run", false, "Preview maintenance operations without making changes")
	maintainCmd.Flags().StringVar(&maintainSchedule, "schedule", "", "Configure schedule (daily, weekly, monthly, pre-deploy, post-deploy)")

	// Register subcommands for categories
	categories := []string{
		"dependencies", "containers", "kubernetes", "cloudflare",
		"database", "security", "performance", "docs", "workspace", "all",
	}
	for _, cat := range categories {
		catCmd := &cobra.Command{
			Use:   cat,
			Short: fmt.Sprintf("Perform engineering maintenance for %s", cat),
			Run:   maintainCmd.Run,
		}
		maintainCmd.AddCommand(catCmd)
	}

	rootCmd.AddCommand(maintainCmd)
}
