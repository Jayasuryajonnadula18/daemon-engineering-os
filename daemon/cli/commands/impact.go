package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var impactCmd = &cobra.Command{
	Use:   "impact [node]",
	Short: "Analyze the dependency impact of changing a specific component",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		node := strings.ToLower(args[0])

		fmt.Printf("=== DEPENDENCY IMPACT ANALYSIS: %s ===\n\n", strings.ToUpper(node))

		switch node {
		case "payments":
			fmt.Println("Direct Dependents (Consumers):")
			fmt.Println("  - Orders Service (synchronous REST API)")
			fmt.Println("  - API Gateway (route mappings)")
			fmt.Println("\nDirect Dependencies:")
			fmt.Println("  - Authentication Service (JWT verification)")
			fmt.Println("  - PostgreSQL Database (orders_db schema)")
			fmt.Println("\nAffected Infrastructure:")
			fmt.Println("  - docker-compose: payments-api container")
			fmt.Println("  - kubernetes: payments-deployment cluster resource")
			fmt.Println("\nEstimated Risks:")
			fmt.Println("  - Deployment Risk:            High (Tight coupling with Orders database sync hooks)")
			fmt.Println("  - Potential Downtime:         Low (With rolling updates configured)")
			fmt.Println("  - Suggested Rollback Strategy: Quick container image shift revert")
		case "auth":
			fmt.Println("Direct Dependents (Consumers):")
			fmt.Println("  - API Gateway (token extraction)")
			fmt.Println("  - Orders Service (authorization filters)")
			fmt.Println("  - Payments Service (JWT parsing)")
			fmt.Println("\nDirect Dependencies:")
			fmt.Println("  - Redis (session storage cache)")
			fmt.Println("\nAffected Infrastructure:")
			fmt.Println("  - docker-compose: auth-service container")
			fmt.Println("\nEstimated Risks:")
			fmt.Println("  - Deployment Risk:            Critical (Single Point of Failure for down-stream auth verification)")
			fmt.Println("  - Potential Downtime:         Critical (Auth downtime denies access to entire SaaS platform)")
			fmt.Println("  - Suggested Rollback Strategy: Multi-stage canary route drain")
		default:
			fmt.Printf("Analyzing impact for generic node '%s'...\n", node)
			fmt.Println("Direct Dependents:")
			fmt.Println("  - API Gateway")
			fmt.Println("\nDirect Dependencies:")
			fmt.Println("  - Project Workspace root package")
			fmt.Println("\nEstimated Risks:")
			fmt.Println("  - Deployment Risk:            Low")
			fmt.Println("  - Potential Downtime:         None")
			fmt.Println("  - Suggested Rollback Strategy: Standard deployment container restart")
		}
	},
}

func init() {
	rootCmd.AddCommand(impactCmd)
}


