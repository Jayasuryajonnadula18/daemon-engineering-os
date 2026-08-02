package commands

import (
	"context"
	"fmt"
	"strings"

	"daemon/core/advisor"
	engContext "daemon/core/context"

	"github.com/spf13/cobra"
)

var (
	serviceAdvFlag    string
	repositoryAdvFlag string
)

func runAdvise(cmd *cobra.Command, args []string) {
	// Determine category from subcommand name or positional arg
	sub := cmd.Name()
	if sub == "advise" {
		sub = ""
		if len(args) > 0 {
			sub = strings.ToLower(args[0])
		}
	}

	ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
	ae := advisor.NewAdvisorEngine(ce)
	report, err := ae.Advise(context.Background(), sub, serviceAdvFlag, repositoryAdvFlag)
	if err != nil {
		fmt.Printf("Error generating advice: %v\n", err)
		return
	}

	fmt.Printf("Engineering Health Score: %d%%\n", report.HealthScore)
	fmt.Printf("Workspace Status:       %s\n", report.Status)
	fmt.Printf("Confidence Score:       %d%%\n", report.ConfidenceScore)
	if report.ProductivityImprovement != "" {
		fmt.Printf("Est. Productivity:      %s\n", report.ProductivityImprovement)
	}

	fmt.Println("\nFindings:")
	for _, f := range report.Findings {
		fmt.Printf("  ✓ %s\n", f)
	}

	fmt.Println("\nRisks:")
	for _, r := range report.Risks {
		fmt.Printf("  ⚠ %s\n", r)
	}

	fmt.Println("\nRecommendations:")
	for _, rec := range report.Recommendations {
		fmt.Printf("  * %s\n", rec)
	}
}

var adviseCmd = &cobra.Command{
	Use:   "advise [category]",
	Short: "Receive AI-powered engineering guidance",
	Long:  `Query context-aware workspace scorecards, health scores, risks, and evidence-based recommendations. Categories: workspace, architecture, deployment, dependencies, security, maintenance, performance, incident, daily.`,
	Run:   runAdvise,
}

func init() {
	adviseCmd.Flags().StringVar(&serviceAdvFlag, "service", "", "Filter recommendations to a specific service")
	adviseCmd.Flags().StringVar(&repositoryAdvFlag, "repository", "", "Filter recommendations to a specific repository")

	// Register named subcommands for each advice category
	for _, cat := range []string{"workspace", "architecture", "deployment", "dependencies", "security", "maintenance", "performance", "incident", "daily"} {
		c := cat
		sub := &cobra.Command{
			Use:   c,
			Short: fmt.Sprintf("Receive %s engineering advice", c),
			Run:   runAdvise,
		}
		adviseCmd.AddCommand(sub)
	}

	rootCmd.AddCommand(adviseCmd)
}
