package commands

import (
	"context"
	"fmt"
	"strings"

	"daemon/core/twin"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Contextually search the live Engineering Twin models",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		fmt.Printf("Searching the Engineering Twin for '%s'...\n\n", query)

		t := twin.NewTwinModel(rt.Container.ResolveGraphStore())
		results, err := t.Search(context.Background(), query)
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("No contextual results matched in the active Twin model.")
			return
		}

		for _, r := range results {
			fmt.Printf("[%s] %s\n", strings.ToUpper(r.Type), r.Name)
			fmt.Printf("  Details: %s\n\n", r.Context)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}


