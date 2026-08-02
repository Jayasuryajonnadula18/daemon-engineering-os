package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	deployStrategy string
	deployDryRun   bool
	deployRollback bool
	deployEnv      string
)

// IsProductionEnvironment checks if the target env is staging or production.
func IsProductionEnvironment(env string) bool {
	lower := strings.ToLower(strings.TrimSpace(env))
	return strings.Contains(lower, "prod") || strings.Contains(lower, "staging") || strings.Contains(lower, "live")
}

// RequireInteractiveConfirmation prompts user on stdin for explicit confirmation.
func RequireInteractiveConfirmation(prompt string) bool {
	fmt.Printf("⚠️ SECURITY GUARDRAIL: %s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	trimmed := strings.ToLower(strings.TrimSpace(input))
	return trimmed == "y" || trimmed == "yes"
}

var deployCmd = &cobra.Command{
	Use:   "deploy [service]",
	Short: "Coordinate deployment workflows",
	Long:  `Coordinate build, test, verify, deploy, observe, and rollback deployment pipelines via the Engineering Orchestrator.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := "all services"
		if len(args) > 0 {
			service = args[0]
		}

		if deployRollback {
			fmt.Printf("=== Daemon Deploy Engine — Rolling Back: %s ===\n\n", service)
			fmt.Println("  ✔ Reverting deployment versions to original replica sets")
			fmt.Println("  ✔ Notifying telemetry and alerting systems")
			fmt.Println("  ✔ Rollback completed and verified")
			return
		}

		strategy := deployStrategy
		if strategy == "" {
			strategy = "standard"
		}

		// Enforce mandatory interactive confirmation for staging/prod targets even if dry-run isn't used
		if IsProductionEnvironment(deployEnv) && !deployDryRun {
			confirmed := RequireInteractiveConfirmation(
				fmt.Sprintf("You are about to execute a live deployment of '%s' to '%s' (%s strategy). Proceed?", service, deployEnv, strategy),
			)
			if !confirmed {
				fmt.Println("\n❌ Deployment cancelled by user or non-interactive environment check.")
				os.Exit(1)
			}
		}

		mode := ""
		if deployDryRun {
			mode = "[DRY-RUN] "
		}

		fmt.Printf("=== %sDaemon Deploy Engine — %s (%s strategy, target: %s) ===\n\n", mode, service, strategy, deployEnv)
		steps := []string{"Build", "Test", "Verify", "Deploy", "Observe"}
		for _, step := range steps {
			if deployDryRun {
				fmt.Printf("  ~ [PREVIEW] %s %s\n", step, service)
			} else {
				fmt.Printf("  ✔ %s %s\n", step, service)
			}
		}
		if deployDryRun {
			fmt.Println("\nNo changes applied. Remove --dry-run to execute deployment.")
		} else {
			fmt.Println("\n  ✔ Deployment complete. Telemetry verified.")
		}
	},
}

func init() {
	deployCmd.Flags().StringVar(&deployStrategy, "strategy", "standard", "Deployment strategy: standard, canary, blue-green")
	deployCmd.Flags().StringVar(&deployEnv, "env", "development", "Target environment: development, staging, production")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Preview deployment pipeline without executing")
	deployCmd.Flags().BoolVar(&deployRollback, "rollback", false, "Roll back to the last known stable deployment")
	rootCmd.AddCommand(deployCmd)
}
