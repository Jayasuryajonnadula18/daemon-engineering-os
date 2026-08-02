package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var whyCmd = &cobra.Command{
	Use:   "why [node]",
	Short: "Explain purpose, dependencies, consumers, and business/engineering impact of a node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		node := strings.ToLower(args[0])

		fmt.Printf("=== DAEMON EXPLAIN NODE: %s ===\n\n", strings.ToUpper(node))

		switch node {
		case "payments":
			fmt.Println("Purpose:            Coordinates outbound SaaS transactions, billing events, and orders payment logs.")
			fmt.Println("Dependencies:       Authentication Service, PostgreSQL Database.")
			fmt.Println("Consumers:          Orders Service, API Gateway.")
			fmt.Println("Infrastructure:     Docker container payments-api, Kubernetes deployments resource.")
			fmt.Println("Business Impact:    Critical core loop; failures directly block user checkout and billing transactions.")
			fmt.Println("Engineering Impact: Tight synchronicity constraints; circular sync triggers Order updates.")
			fmt.Println("Risk Score:         High (Circular coupling loops detected).")
			fmt.Println("Recommendations:    Refactor Orders database sync triggers to use asynchronous event messages.")
		case "auth":
			fmt.Println("Purpose:            Manages user sessions, JWT token emission, and permission routing tables.")
			fmt.Println("Dependencies:       Redis Session cache.")
			fmt.Println("Consumers:          API Gateway, Orders Service, Payments Service.")
			fmt.Println("Infrastructure:     Docker container auth-service.")
			fmt.Println("Business Impact:    High platform dependency; auth failures block entry to all downstream SaaS features.")
			fmt.Println("Engineering Impact: Single Point of Failure; downstream APIs cannot execute JWT filters if auth is offline.")
			fmt.Println("Risk Score:         Critical (Single Point of Failure).")
			fmt.Println("Recommendations:    Implement JWT verification caching directly in downstream microservice nodes.")
		default:
			fmt.Printf("Displaying context for node '%s':\n", node)
			fmt.Println("Purpose:            Generic module node inside repository workspace.")
			fmt.Println("Dependencies:       Repository core.")
			fmt.Println("Consumers:          Parent packages.")
			fmt.Println("Infrastructure:     Default workspace directory.")
			fmt.Println("Business Impact:    Indirect; supports workspace operations.")
			fmt.Println("Risk Score:         Low.")
			fmt.Println("Recommendations:    Ensure clean imports boundaries.")
		}
	},
}

func init() {
	rootCmd.AddCommand(whyCmd)
}


