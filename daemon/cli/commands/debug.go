package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"daemon/cli/output"
	"daemon/core/debug"
	"daemon/core/reasoning"
	"daemon/core/storage"
	"github.com/spf13/cobra"
)

var (
	debugJSONFlag    bool
	debugNoLLMFlag   bool
	debugDeepFlag    bool
	debugChangedFlag bool
	debugVerboseFlag bool
)

var debugCmd = &cobra.Command{
	Use:   "debug [problem]",
	Short: "Autonomously investigate software bugs, crashes, regressions, or performance bottlenecks",
	Run: func(cmd *cobra.Command, args []string) {
		problem := "application is slow or failing"
		if len(args) > 0 {
			problem = args[0]
		}

		// Redact secrets in the input problem
		problem = debug.RedactSecrets(problem)

		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}

		var store *debug.DebugStore
		dbStore := rt.Container.ResolveGraphStore()
		if dbProv, ok := dbStore.(storage.DatabaseProvider); ok {
			var err error
			store, err = debug.NewDebugStoreFromDB(dbProv.DB())
			if err != nil {
				if debugJSONFlag {
					output.RenderJSON("debug", nil, err)
					return
				}
				fmt.Printf("Failed to open debug database from shared connection: %v\n", err)
				os.Exit(1)
			}
		} else {
			dbPath := filepath.Join(cwd, ".daemon", "daemon.db")
			var err error
			store, err = debug.NewDebugStore(dbPath)
			if err != nil {
				if debugJSONFlag {
					output.RenderJSON("debug", nil, err)
					return
				}
				fmt.Printf("Failed to open debug database: %v\n", err)
				os.Exit(1)
			}
		}
		defer store.Close()

		// Deduplication check
		dup, err := store.FindDuplicate(problem)
		if err == nil && dup != nil && time.Since(dup.StartedAt) < 1*time.Minute {
			// Reuse recent duplicate investigation to save resources
			if debugJSONFlag {
				out, err := debug.RenderJSON(dup, nil)
				if err == nil {
					fmt.Println(out)
				}
				return
			}
			fmt.Println("Found active duplicate investigation. Returning cached diagnosis:")
			fmt.Println(debug.RenderStdout(dup))
			return
		}

		debugger := debug.NewDebugger(store, nil)
		if !debugNoLLMFlag {
			router := reasoning.NewModelRouter(false)
			deterministicEngine := reasoning.NewDeterministicReasoningEngine()
			llmEngine := reasoning.NewLLMReasoningEngine(router)
			hybridEngine := reasoning.NewHybridReasoningEngine(deterministicEngine, llmEngine, true)
			debugger.SetReasoningEngine(hybridEngine)
		}
		invID := fmt.Sprintf("dbg-%d", time.Now().UnixNano())

		// Execute progressive investigation engine
		res, err := debugger.RunInvestigation(context.Background(), invID, problem, cwd, debugDeepFlag, debugChangedFlag, !debugNoLLMFlag)
		if err != nil {
			if debugJSONFlag {
				output.RenderJSON("debug", nil, err)
				return
			}
			fmt.Printf("Debug investigation failed: %v\n", err)
			os.Exit(1)
		}

		if debugJSONFlag {
			out, err := debug.RenderJSON(res, nil)
			if err != nil {
				output.RenderJSON("debug", nil, err)
				return
			}
			fmt.Println(out)
			return
		}

		fmt.Println(debug.RenderStdout(res))
	},
}

func init() {
	debugCmd.Flags().BoolVar(&debugJSONFlag, "json", false, "Output machine-readable investigation report")
	debugCmd.Flags().BoolVar(&debugNoLLMFlag, "no-llm", false, "100% deterministic investigation mode")
	debugCmd.Flags().BoolVar(&debugDeepFlag, "deep", false, "Expand investigation budget constraint parameters")
	debugCmd.Flags().BoolVar(&debugChangedFlag, "changed", false, "Prioritize files and components changed recently")
	debugCmd.Flags().BoolVar(&debugVerboseFlag, "verbose", false, "Detailed progressive trace step logging")

	rootCmd.AddCommand(debugCmd)
}
