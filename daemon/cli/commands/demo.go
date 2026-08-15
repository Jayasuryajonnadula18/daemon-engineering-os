package commands

import (
	"fmt"
	"os"
	"time"

	"daemon/core/domain"
	"daemon/core/events"
	"daemon/cli/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Load a complete SaaS microservices demo project and launch the Cockpit TUI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("✔ Initializing Fictional SaaS Platform Demo Environment...")
		time.Sleep(300 * time.Millisecond)

		gs := rt.Container.ResolveGraphStore()
		ms := rt.Container.ResolveMemoryStore()
		eb := rt.Container.ResolveEventBus()

		// 1. Clear database
		_ = gs.Clear()

		// 2. Load SaaS nodes
		nodes := []struct {
			Type, ID, Name string
			Props          map[string]string
		}{
			{"project", "saas-core", "SaaS Platform", map[string]string{"language": "TypeScript", "framework": "Next.js", "package_manager": "pnpm"}},
			{"service", "frontend", "Frontend Gateway SPA", map[string]string{"port": "3000", "status": "running"}},
			{"service", "api-gateway", "API Gateway", map[string]string{"port": "80", "status": "running"}},
			{"service", "auth", "Authentication Service", map[string]string{"port": "5001", "status": "running"}},
			{"service", "orders", "Orders Service API", map[string]string{"port": "5002", "status": "running", "depends_on": `["auth", "payments"]`}},
			{"service", "payments", "Payments Service API", map[string]string{"port": "5003", "status": "running", "depends_on": `["auth", "orders"]`}},
			{"service", "notifications", "Notifications Worker", map[string]string{"port": "0", "status": "running"}},
			{"service", "analytics", "Analytics Engine", map[string]string{"port": "5004", "status": "running"}},
			{"database", "postgres", "PostgreSQL database", map[string]string{"port": "5432", "status": "running"}},
			{"database", "redis", "Redis Session cache", map[string]string{"port": "6379", "status": "running"}},
			{"infrastructure", "docker-compose", "Docker Compose layout", map[string]string{"path": "./docker-compose.yml"}},
			{"infrastructure", "kubernetes", "K8s cluster resources", map[string]string{"path": "./k8s"}},
			{"infrastructure", "terraform", "Terraform Cloud resources", map[string]string{"path": "./terraform"}},
			{"infrastructure", "github-actions", "GitHub Actions CI/CD", map[string]string{"path": "./.github/workflows"}},
			{"infrastructure", "aws-cloud", "AWS Cloud Infrastructure", map[string]string{"provider": "aws"}},
		}

		for _, n := range nodes {
			_ = gs.AddNode(n.Type, n.ID, n.Name, n.Props)
		}

		// Connect relationships
		_ = gs.AddEdge("project", "saas-core", "service", "frontend", "contains")
		_ = gs.AddEdge("project", "saas-core", "service", "api-gateway", "contains")
		_ = gs.AddEdge("project", "saas-core", "service", "auth", "contains")
		_ = gs.AddEdge("project", "saas-core", "service", "orders", "contains")
		_ = gs.AddEdge("project", "saas-core", "service", "payments", "contains")
		_ = gs.AddEdge("project", "saas-core", "service", "notifications", "contains")
		_ = gs.AddEdge("project", "saas-core", "service", "analytics", "contains")
		_ = gs.AddEdge("project", "saas-core", "infrastructure", "github-actions", "contains")
		_ = gs.AddEdge("project", "saas-core", "infrastructure", "aws-cloud", "contains")

		_ = gs.AddEdge("service", "orders", "database", "postgres", "queries")
		_ = gs.AddEdge("service", "payments", "database", "postgres", "queries")
		_ = gs.AddEdge("service", "auth", "database", "redis", "caches")

		// 3. Load historical incidents and recommendations
		incidents := []domain.Incident{
			{ID: "inc-1", Message: "Single Point of Failure: API Gateway is not configured with cluster replicas", Severity: "critical", Resolved: false, DetectedAt: time.Now().Add(-2 * time.Hour)},
			{ID: "inc-2", Message: "Circular Dependency: Orders Service invokes Payments Service which loops back to Orders database hooks", Severity: "warning", Resolved: false, DetectedAt: time.Now().Add(-1 * time.Hour)},
			{ID: "inc-3", Message: "Outdated dependency: lodash (v4.17.21) contains 2 active security vulnerability advisories", Severity: "info", Resolved: false, DetectedAt: time.Now().Add(-30 * time.Minute)},
		}

		for _, inc := range incidents {
			_ = ms.AddIncident(&inc)
		}

		recs := []domain.Recommendation{
			{ID: "rec-1", Category: "security", Message: "Update outdated lodash library dependency containing vulnerability advisories", Rationale: "Active security vulnerabilities risk external exposure."},
			{ID: "rec-2", Category: "architecture", Message: "Refactor Orders database sync to use asynchronous event messages", Rationale: "Synchronous circular triggers between Orders and Payments service decrease reliability."},
			{ID: "rec-3", Category: "documentation", Message: "Generate missing microservices configuration setup documentation", Rationale: "Missing configuration guides lead to high setup friction."},
		}

		for _, r := range recs {
			_ = ms.AddRecommendation(&r)
		}

		// 4. Populate Timeline Events
		eventsList := []struct {
			Type    string
			Payload interface{}
		}{
			{"RepositoryInitialized", "SaaS Core Workspace detected"},
			{"DiscoveryCompleted", "SaaS Platform discovery finished"},
			{"IncidentDetected", "Circular Coupling detected: Orders -> Payments -> Orders"},
			{"IncidentDetected", "SPOF Warning: single instance API Gateway"},
			{"DeploymentStarted", "Deploying SaaS core version v1.2.0"},
			{"DeploymentCompleted", "SaaS core v1.2.0 deployed successfully"},
		}

		for _, e := range eventsList {
			eb.Publish(events.Event{
				Type:      e.Type,
				Payload:   map[string]any{"data": e.Payload},
				Timestamp: time.Now().Add(-15 * time.Minute),
			})
		}

		fmt.Println("✔ SaaS platform simulation loaded successfully.")
		time.Sleep(100 * time.Millisecond)

		// Launch Cockpit TUI
		p := tea.NewProgram(tui.NewModel(rt), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting Engineering Cockpit TUI in demo mode: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
}


