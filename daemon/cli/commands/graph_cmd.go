package commands

import (
	"fmt"
	"strings"

	engContext "daemon/core/context"

	"github.com/spf13/cobra"
)

var graphFilter string

var graphCmd = &cobra.Command{
	Use:   "graph [kind]",
	Short: "Explore the Engineering Twin and Knowledge Graph",
	Long:  `Traverse and visualize the Engineering Twin nodes and edges including dependency, service, API, infrastructure, impact, and execution graphs.`,
	Run: func(cmd *cobra.Command, args []string) {
		kind := "all"
		if len(args) > 0 {
			kind = strings.ToLower(args[0])
		}

		fmt.Println("=== DAEMON KNOWLEDGE GRAPH EXPLORER ===")
		fmt.Printf("Graph Kind:  %s\n", strings.ToUpper(kind))
		if graphFilter != "" {
			fmt.Printf("Filter:      %s\n", graphFilter)
		}
		fmt.Println()

		ctx := cmd.Context()
		ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		engCtx, err := ce.BuildContext(ctx)
		if err != nil {
			fmt.Printf("Error building Engineering Context: %v\n", err)
			return
		}

		if len(engCtx.Services) == 0 {
			fmt.Println("  ⚠ Knowledge Graph is empty. Run 'daemon workspace sync' to populate the Twin.")
			return
		}

		fmt.Printf("  Engineering Twin Nodes (%d services):\n", len(engCtx.Services))
		for i, svc := range engCtx.Services {
			fmt.Printf("    [%d] %-30s → Status: %s\n", i+1, svc.Name, svc.Status)
		}

		fmt.Printf("\n  Dependencies (%d tracked):\n", len(engCtx.Dependencies))
		for i, dep := range engCtx.Dependencies {
			fmt.Printf("    [%d] %-30s → %s\n", i+1, dep.Name, dep.Version)
		}
	},
}

func init() {
	graphCmd.Flags().StringVar(&graphFilter, "filter", "", "Filter graph nodes by name or kind")

	// Register subcommands for graph kinds
	graphKinds := []string{"services", "dependencies", "apis", "databases", "infrastructure", "impact", "execution"}
	for _, kind := range graphKinds {
		k := kind
		graphCmd.AddCommand(&cobra.Command{
			Use:   k,
			Short: fmt.Sprintf("Explore %s graph", k),
			Run:   graphCmd.Run,
		})
	}

	rootCmd.AddCommand(graphCmd)
}
