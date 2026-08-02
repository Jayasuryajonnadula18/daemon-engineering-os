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
		nodeTarget := strings.ToLower(strings.TrimSpace(args[0]))

		fmt.Printf("=== DEPENDENCY IMPACT ANALYSIS: %s ===\n\n", strings.ToUpper(nodeTarget))

		gs := rt.Container.ResolveGraphStore()
		allEdges, err := gs.GetEdges()
		if err != nil {
			fmt.Printf("Error accessing Knowledge Graph: %v\n", err)
			return
		}

		allNodes, _ := gs.GetAllNodes()

		// Real Graph Edge Traversal
		var directDependents []string // Node is target (other nodes call/depend on target)
		var directDependencies []string // Target calls/depends on other nodes

		for _, e := range allEdges {
			fromLower := strings.ToLower(e.FromID)
			toLower := strings.ToLower(e.ToID)

			if strings.Contains(toLower, nodeTarget) || strings.Contains(strings.ToLower(e.ToType), nodeTarget) {
				directDependents = append(directDependents, fmt.Sprintf("%s (%s -> %s)", e.FromID, e.FromType, e.Relation))
			}
			if strings.Contains(fromLower, nodeTarget) || strings.Contains(strings.ToLower(e.FromType), nodeTarget) {
				directDependencies = append(directDependencies, fmt.Sprintf("%s (%s -> %s)", e.ToID, e.ToType, e.Relation))
			}
		}

		// Calculate Risk Level based on inbound dependent count
		riskLevel := "Low"
		switch {
		case len(directDependents) >= 3:
			riskLevel = "Critical (Single Point of Failure for down-stream consumers)"
		case len(directDependents) == 2:
			riskLevel = "High (Tight coupling with multiple service consumers)"
		case len(directDependents) == 1:
			riskLevel = "Medium (Downstream impact on 1 consumer)"
		}

		fmt.Println("Direct Dependents (Consumers / Inbound Edges):")
		if len(directDependents) > 0 {
			for _, dep := range directDependents {
				fmt.Printf("  • %s\n", dep)
			}
		} else {
			fmt.Println("  (None detected — standalone component or leaf node)")
		}

		fmt.Println("\nDirect Dependencies (Outbound Edges):")
		if len(directDependencies) > 0 {
			for _, dep := range directDependencies {
				fmt.Printf("  • %s\n", dep)
			}
		} else {
			fmt.Println("  (No outbound graph edges registered for this node)")
		}

		fmt.Println("\nKnowledge Graph Context:")
		fmt.Printf("  • Total Tracked Graph Nodes: %d\n", len(allNodes))
		fmt.Printf("  • Total Relational Edges:    %d\n", len(allEdges))

		fmt.Println("\nDynamic Risk Assessment:")
		fmt.Printf("  • Calculated Impact Risk:    %s\n", riskLevel)
		fmt.Printf("  • Inbound Degree Centrality: %d\n", len(directDependents))
		fmt.Println("  • Rollback Strategy:         Atomic image/container snapshot revert via 'daemon fix --rollback'")
	},
}

func init() {
	rootCmd.AddCommand(impactCmd)
}
