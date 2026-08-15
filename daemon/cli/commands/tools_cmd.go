package commands

import (
	"fmt"
	"os"

	"daemon/cli/output"
	"daemon/core/instruments"
	gobuild "daemon/core/instruments/adapters/build/go"
	staticgo "daemon/core/instruments/adapters/static/go"
	staticjs "daemon/core/instruments/adapters/static/javascript"
	staticread "daemon/core/instruments/adapters/static"
	gotest "daemon/core/instruments/adapters/testing/go"
	"github.com/spf13/cobra"
)

var toolsJSONFlag bool

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Explore and inspect registered engineering instruments",
}

var toolsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered engineering instruments in the runtime",
	Run: func(cmd *cobra.Command, args []string) {
		reg := instruments.NewInstrumentRegistry()
		_ = reg.Register(gobuild.NewGoBuildInstrument())
		_ = reg.Register(gotest.NewGoTestInstrument())
		_ = reg.Register(staticgo.NewGoLeakInstrument())
		_ = reg.Register(staticjs.NewJSBugsInstrument())
		_ = reg.Register(staticread.NewReadFileInstrument())
		list := reg.List()

		if toolsJSONFlag {
			var serialized []map[string]string
			for _, t := range list {
				serialized = append(serialized, map[string]string{
					"id":          t.Identity().ID,
					"name":        t.Identity().Name,
					"version":     t.Identity().Version,
					"description": t.Identity().Description,
					"category":    string(t.Identity().Category),
					"vendor":      t.Identity().Vendor,
				})
			}
			output.RenderJSON("tools.list", serialized, nil)
			return
		}

		fmt.Println("=== REGISTERED DAEMON ENGINEERING INSTRUMENTS ===")
		for _, t := range list {
			fmt.Printf("  - %s (%s) | Vendor: %s | Category: %s\n", t.Identity().Name, t.Identity().ID, t.Identity().Vendor, t.Identity().Category)
			fmt.Printf("    Description: %s\n\n", t.Identity().Description)
		}
	},
}

var toolsInspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Inspect the details of a specific instrument",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		reg := instruments.NewInstrumentRegistry()
		_ = reg.Register(gobuild.NewGoBuildInstrument())
		_ = reg.Register(gotest.NewGoTestInstrument())
		_ = reg.Register(staticgo.NewGoLeakInstrument())
		_ = reg.Register(staticjs.NewJSBugsInstrument())
		_ = reg.Register(staticread.NewReadFileInstrument())
		t := reg.FindByID(id)
		if t == nil {
			err := fmt.Errorf("instrument '%s' not found", id)
			if toolsJSONFlag {
				output.RenderJSON("tools.inspect", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}

		if toolsJSONFlag {
			serialized := map[string]interface{}{
				"id":          t.Identity().ID,
				"name":        t.Identity().Name,
				"version":     t.Identity().Version,
				"description": t.Identity().Description,
				"category":    t.Identity().Category,
				"vendor":      t.Identity().Vendor,
				"source_url":  t.Identity().SourceURL,
			}
			output.RenderJSON("tools.inspect", serialized, nil)
			return
		}

		fmt.Printf("Instrument ID:    %s\n", t.Identity().ID)
		fmt.Printf("Name:             %s\n", t.Identity().Name)
		fmt.Printf("Version:          %s\n", t.Identity().Version)
		fmt.Printf("Description:      %s\n", t.Identity().Description)
		fmt.Printf("Category:         %s\n", t.Identity().Category)
		fmt.Printf("Vendor:           %s\n", t.Identity().Vendor)
		fmt.Printf("Source URL:       %s\n", t.Identity().SourceURL)
	},
}

func init() {
	toolsCmd.PersistentFlags().BoolVar(&toolsJSONFlag, "json", false, "Output machine-readable JSON")
	toolsCmd.AddCommand(toolsListCmd)
	toolsCmd.AddCommand(toolsInspectCmd)
	rootCmd.AddCommand(toolsCmd)
}
