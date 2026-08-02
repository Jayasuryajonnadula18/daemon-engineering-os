package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var whyCmd = &cobra.Command{
	Use:   "why [node]",
	Short: "Explain purpose, dependencies, consumers, and business/engineering impact of a node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetNode := strings.ToLower(strings.TrimSpace(args[0]))

		fmt.Printf("=== DAEMON EXPLAIN NODE: %s ===\n\n", strings.ToUpper(targetNode))

		gs := rt.Container.ResolveGraphStore()
		re := rt.Container.ResolveReasoningEngine()

		allNodes, err := gs.GetAllNodes()
		if err != nil {
			fmt.Printf("Error querying Knowledge Graph: %v\n", err)
			return
		}

		allEdges, _ := gs.GetEdges()

		// Find target node in Knowledge Graph
		foundName := targetNode
		foundType := "module"
		foundPath := ""
		nodeMatched := false

		for _, n := range allNodes {
			if strings.Contains(strings.ToLower(n.ID), targetNode) || strings.Contains(strings.ToLower(n.Name), targetNode) {
				foundName = n.Name
				foundType = n.Type
				foundPath = n.Path
				nodeMatched = true
				break
			}
		}

		// Traverse real graph edges
		var consumers []string
		var dependencies []string

		for _, e := range allEdges {
			toLower := strings.ToLower(e.ToID)
			fromLower := strings.ToLower(e.FromID)

			if strings.Contains(toLower, targetNode) {
				consumers = append(consumers, e.FromID+" ("+e.FromType+")")
			}
			if strings.Contains(fromLower, targetNode) {
				dependencies = append(dependencies, e.ToID+" ("+e.ToType+")")
			}
		}

		// Calculate impact & reasoning
		riskScore := "Low"
		if len(consumers) >= 3 {
			riskScore = "Critical (High degree centrality / Single Point of Failure)"
		} else if len(consumers) >= 1 {
			riskScore = "Medium (Active downstream consumers)"
		}

		promptContext := fmt.Sprintf("Node: %s (Type: %s, Path: %s). Consumers: %s. Dependencies: %s.",
			foundName, foundType, foundPath, strings.Join(consumers, ", "), strings.Join(dependencies, ", "))

		explanation, _ := re.Explain(context.Background(), foundName, promptContext)

		fmt.Printf("Matched Node:       %s (Type: %s)\n", foundName, foundType)
		if foundPath != "" {
			fmt.Printf("Workspace Path:     %s\n", foundPath)
		}
		fmt.Printf("Direct Consumers:   %s\n", strings.Join(consumers, ", "))
		if len(consumers) == 0 {
			fmt.Println("  (No inbound consumer edges registered)")
		}
		fmt.Printf("Dependencies:       %s\n", strings.Join(dependencies, ", "))
		if len(dependencies) == 0 {
			fmt.Println("  (No outbound dependency edges registered)")
		}
		fmt.Printf("Calculated Risk:    %s\n", riskScore)

		fmt.Println("\nAI Architecture Reasoning:")
		fmt.Printf("  • Explanation:     %s\n", explanation)

		if !nodeMatched {
			fmt.Printf("\n💡 Tip: '%s' was not explicitly indexed. Run 'daemon sync' to re-index all workspace nodes.\n", targetNode)
		}
	},
}

func init() {
	rootCmd.AddCommand(whyCmd)
}
