package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"daemon/cli/commands"
	"daemon/core/events"
	"daemon/core/graph"
	"daemon/core/policies"
	"daemon/core/reasoning"
	"daemon/core/runtime"
	"daemon/core/workflow"
	"daemon/sdk/plugin"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to resolve current working directory: %v\n", err)
		os.Exit(1)
	}

	daemonDir := filepath.Join(cwd, ".daemon")
	_ = os.MkdirAll(daemonDir, 0755)
	dbPath := filepath.Join(daemonDir, "daemon.db")

	// Instantiate concrete stores & core engines
	dbStore, err := graph.NewSQLiteStore(dbPath)
	if err != nil {
		fmt.Printf("Failed to open SQLite database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = dbStore.Close()
	}()

	eb := events.NewMemoryEventBus()
	we := workflow.NewMemoryWorkflowEngine(eb)
	re := reasoning.NewMemoryReasoningEngine("")
	pe := policies.NewMemoryPolicyEngine(false)
	reg := plugin.NewMemoryCapabilityRegistry()
	sched := runtime.NewMemoryScheduler()
	res := runtime.NewMemoryResourceManager()

	// Bind interfaces via Service Container dependency injection
	container := runtime.NewDefaultContainer(
		dbStore,
		dbStore, // SQLiteStore implements both GraphStore and MemoryStore
		we,
		re,
		reg,
		pe,
		sched,
		eb,
		res,
	)

	rt := runtime.NewRuntime(container)

	ctx := context.Background()
	if err := rt.Initialize(ctx); err != nil {
		fmt.Printf("Failed to initialize Daemon Runtime: %v\n", err)
		os.Exit(1)
	}

	if err := rt.Start(ctx); err != nil {
		fmt.Printf("Failed to start Daemon Runtime: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = rt.Shutdown(ctx)
	}()

	// Execute Cobra CLI routing
	commands.Execute(rt)
}

