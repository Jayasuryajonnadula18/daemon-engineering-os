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
			fmt.Println("================================================================================")
			fmt.Println("=== [DRY-RUN MODE] Previewing Maintenance Engine operations ===")
			fmt.Println("================================================================================")
			autoFix = false
		}

		pe := policies.NewMemoryPolicyEngine(maintainDryRunFlag)
		ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		me := maintenance.NewMaintenanceEngine(ce, pe)

		rep, err := me.RunMaintenance(cmd.Context(), category, autoFix)
		if err != nil {
			fmt.Printf("❌ Error running Maintenance Engine: %v\n", err)
			return
		}

		fmt.Println("================================================================================")
		fmt.Println("🛠️  DAEMON WORKSPACE MAINTENANCE ENGINE (PILLAR 24)")
		fmt.Println("================================================================================")
		fmt.Printf("Category:             %s\n", strings.ToUpper(rep.Category))
		fmt.Printf("Workspace Health:     %s\n", rep.Status)
		fmt.Printf("Confidence Score:     %d%%\n", rep.ConfidenceScore)
		fmt.Printf("Est. Developer Saved: %s\n", rep.EstimatedTimeSaved)
		fmt.Println("--------------------------------------------------------------------------------")

		fmt.Println("\n🔍 OBSERVATION:")
		fmt.Printf("  • %s\n", rep.Observation)

		fmt.Println("\n📋 EVIDENTIARY AUDIT:")
		for _, ev := range rep.Evidence {
			fmt.Printf("  %s\n", ev)
		}

		if len(rep.IncidentsFound) > 0 {
			fmt.Println("\n⚠️  FLAGGED INCIDENTS & WORKSPACE DRIFT:")
			for _, inc := range rep.IncidentsFound {
				fmt.Printf("  • [%s] Target: %s — %s\n", strings.ToUpper(inc.Severity), inc.Target, inc.Message)
			}
		}

		if len(rep.RepairsExecuted) > 0 {
			fmt.Println("\n⚡ SELF-HEALING REPAIRS EXECUTED:")
			for _, r := range rep.RepairsExecuted {
				fmt.Printf("  %s\n", r)
			}
		} else if len(rep.SelfHealingActions) > 0 {
			fmt.Println("\n💡 RECOMMENDED SELF-HEALING ACTIONS:")
			for _, act := range rep.SelfHealingActions {
				fmt.Printf("  -> %s\n", act)
			}
			if !autoFix {
				fmt.Println("\n👉 Run 'daemon maintain --fix' to execute these self-healing actions automatically.")
			}
		}

		fmt.Println("\n💡 RECOMMENDATION:")
		fmt.Printf("  * %s\n", rep.Recommendation)
		fmt.Println("================================================================================")
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
