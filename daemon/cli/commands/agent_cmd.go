package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"daemon/cli/output"
	"daemon/core/agent"
	"daemon/core/policies"
	"daemon/core/reasoning"
	"daemon/core/resource"
	"github.com/spf13/cobra"
)

var (
	agentJSONFlag  bool
	agentDryRun    bool
	agentMaxSteps  int
	agentTimeout   int
	agentMaxTools  int
	agentVerbose   bool
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run and manage autonomous agent sessions",
}

var agentRunCmd = &cobra.Command{
	Use:   "run <intent>",
	Short: "Start a new agent session to execute an engineering task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		intent := args[0]
		dbPath := filepath.Join(".daemon", "daemon.db")

		sessionStore, err := agent.NewSessionStore(dbPath)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.run", nil, err)
				return
			}
			fmt.Printf("Error opening session database: %v\n", err)
			os.Exit(1)
		}
		defer sessionStore.Close()

		// Resolve policy engine and resource governor from global container
		var pe policies.PolicyEngine
		var gov *resource.ResourceGovernor
		if rt != nil && rt.Container != nil {
			pe = rt.Container.ResolvePolicyEngine()
			if manager := rt.Container.ResolveResourceManager(); manager != nil {
				// Use type assertion or casting if necessary, but resource manager interface fits
			}
		}

		router := reasoning.NewModelRouter(false)
		deterministicEngine := reasoning.NewDeterministicReasoningEngine()
		llmEngine := reasoning.NewLLMReasoningEngine(router)
		hybridEngine := reasoning.NewHybridReasoningEngine(deterministicEngine, llmEngine, true)
		runtime := agent.NewAgentRuntimeWithInstruments(pe, gov, sessionStore, nil, nil, hybridEngine)
		sessionID := fmt.Sprintf("sess-%d", time.Now().UnixNano())

		// Configure custom budget from CLI flags
		sess := agent.AgentSession{
			ID:        sessionID,
			Intent:    intent,
			State:     string(agent.StateIdle),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Budget: agent.AgentBudget{
				MaxIterations:        agentMaxSteps,
				MaxToolCalls:         agentMaxTools,
				MaxDuration:          float64(agentTimeout),
				MaxRiskLevel:         "high",
				MaxParallelTools:     1,
				MaxRetries:           3,
				MaxExperiments:       5,
				MaxWrites:            5,
				MaxNetworkOperations: 5,
			},
		}
		_ = sessionStore.SaveSession(sess)

		if agentDryRun {
			fmt.Printf("Dry-run requested. Session %s prepared with budget: %+v\n", sessionID, sess.Budget)
			return
		}

		// Run loop (using deterministic/offline reasoning mode to remain local-friendly)
		resSess, err := runtime.RunLoop(context.Background(), sessionID, intent, false)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.run", nil, err)
				return
			}
			fmt.Printf("Agent runtime execution failure: %v\n", err)
			os.Exit(1)
		}

		if agentJSONFlag {
			output.RenderJSON("agent.run", resSess, nil)
			return
		}

		fmt.Println("=== DAEMON AGENT RUN LOOP EXITED ===")
		fmt.Printf("Session ID:   %s\n", resSess.ID)
		fmt.Printf("Final State:  %s\n", resSess.State)
		fmt.Printf("Result:       %s\n", resSess.FinalResult)
		if resSess.Failure != "" {
			fmt.Printf("Failure:      %s\n", resSess.Failure)
		}
	},
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active and completed agent sessions",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := filepath.Join(".daemon", "daemon.db")
		sessionStore, err := agent.NewSessionStore(dbPath)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.list", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}
		defer sessionStore.Close()

		list, err := sessionStore.ListSessions()
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.list", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}

		if agentJSONFlag {
			output.RenderJSON("agent.list", list, nil)
			return
		}

		fmt.Println("=== ACTIVE DAEMON AGENT SESSIONS ===")
		for _, s := range list {
			fmt.Printf("  [%s] State: %s | Intent: %q\n", s.ID, s.State, s.Intent)
		}
	},
}

var agentInspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Inspect detail of a specific agent session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		dbPath := filepath.Join(".daemon", "daemon.db")
		sessionStore, err := agent.NewSessionStore(dbPath)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.inspect", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}
		defer sessionStore.Close()

		s, err := sessionStore.GetSession(id)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.inspect", nil, err)
				return
			}
			fmt.Printf("Session %s not found\n", id)
			os.Exit(1)
		}

		if agentJSONFlag {
			output.RenderJSON("agent.inspect", s, nil)
			return
		}

		fmt.Printf("Session ID:         %s\n", s.ID)
		fmt.Printf("Intent:             %s\n", s.Intent)
		fmt.Printf("State:              %s\n", s.State)
		fmt.Printf("Plan Reference:     %s\n", s.PlanRef)
		fmt.Printf("Final Result:       %s\n", s.FinalResult)
		fmt.Printf("Failure Rationale:  %s\n", s.Failure)
		fmt.Printf("Created At:         %s\n", s.CreatedAt.Format(time.RFC3339))
	},
}

var agentResumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume execution of a suspended or interrupted agent session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		dbPath := filepath.Join(".daemon", "daemon.db")
		sessionStore, err := agent.NewSessionStore(dbPath)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.resume", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}
		defer sessionStore.Close()

		s, err := sessionStore.GetSession(id)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.resume", nil, err)
				return
			}
			fmt.Printf("Session %s not found\n", id)
			os.Exit(1)
		}

		// Transition back to running state
		s.State = string(agent.StatePlanning)
		s.UpdatedAt = time.Now()
		_ = sessionStore.SaveSession(*s)

		var pe policies.PolicyEngine
		var gov *resource.ResourceGovernor
		if rt != nil && rt.Container != nil {
			pe = rt.Container.ResolvePolicyEngine()
		}

		router := reasoning.NewModelRouter(false)
		deterministicEngine := reasoning.NewDeterministicReasoningEngine()
		llmEngine := reasoning.NewLLMReasoningEngine(router)
		hybridEngine := reasoning.NewHybridReasoningEngine(deterministicEngine, llmEngine, true)
		runtime := agent.NewAgentRuntimeWithInstruments(pe, gov, sessionStore, nil, nil, hybridEngine)
		resSess, err := runtime.RunLoop(context.Background(), s.ID, s.Intent, false)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.resume", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}

		if agentJSONFlag {
			output.RenderJSON("agent.resume", resSess, nil)
			return
		}
		fmt.Printf("Resumed Session %s completed: %s\n", resSess.ID, resSess.FinalResult)
	},
}

var agentCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Transition active agent session to CANCELLED state",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		dbPath := filepath.Join(".daemon", "daemon.db")
		sessionStore, err := agent.NewSessionStore(dbPath)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.cancel", nil, err)
				return
			}
			fmt.Println(err)
			os.Exit(1)
		}
		defer sessionStore.Close()

		s, err := sessionStore.GetSession(id)
		if err != nil {
			if agentJSONFlag {
				output.RenderJSON("agent.cancel", nil, err)
				return
			}
			fmt.Printf("Session %s not found\n", id)
			os.Exit(1)
		}

		s.State = string(agent.StateCancelled)
		s.CancellationState = true
		s.UpdatedAt = time.Now()
		_ = sessionStore.SaveSession(*s)

		if agentJSONFlag {
			output.RenderJSON("agent.cancel", s, nil)
			return
		}
		fmt.Printf("Successfully cancelled session %s\n", id)
	},
}

func init() {
	agentRunCmd.Flags().BoolVar(&agentDryRun, "dry-run", false, "Prepare session budget without starting execution loop")
	agentRunCmd.Flags().IntVar(&agentMaxSteps, "max-steps", 10, "Maximum agent iterations allowed")
	agentRunCmd.Flags().IntVar(&agentTimeout, "timeout", 60, "Timeout threshold in seconds")
	agentRunCmd.Flags().IntVar(&agentMaxTools, "max-tools", 15, "Maximum tool requests permitted")
	agentRunCmd.Flags().BoolVar(&agentVerbose, "verbose", false, "Emit detailed verbose run traces")

	agentCmd.PersistentFlags().BoolVar(&agentJSONFlag, "json", false, "Output machine-readable JSON format")
	agentCmd.AddCommand(agentRunCmd)
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentInspectCmd)
	agentCmd.AddCommand(agentResumeCmd)
	agentCmd.AddCommand(agentCancelCmd)

	rootCmd.AddCommand(agentCmd)
}
