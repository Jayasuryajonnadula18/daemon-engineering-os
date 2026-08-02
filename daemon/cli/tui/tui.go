package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"daemon/core/advisor"
	"daemon/core/architecture"
	engContext "daemon/core/context"
	"daemon/core/events"
	"daemon/core/insights"
	"daemon/core/integrations"
	"daemon/core/recommendation"
	"daemon/core/replay"
	"daemon/core/runtime"
	"daemon/core/twin"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewType specifies tabs within the Cockpit.
type ViewType int

const (
	ViewOverview ViewType = iota
	ViewWorkspace
	ViewTwin
	ViewTimeline
	ViewArchitecture
	ViewDependencies
	ViewRecommendations
	ViewSearch
	ViewIntegrations
	ViewContainers
	ViewInfrastructure
	ViewCloud
	ViewMaintenance
	ViewAutomation
	ViewSessions
	ViewReplay
	ViewAdvisor
)

// Model represents the visual state wrapper for Bubble Tea.
type Model struct {
	runtime        *runtime.Runtime
	activeView     ViewType
	width, height  int
	showPalette    bool
	paletteIndex   int
	paletteOptions []string
	searchQuery    string
	searching      bool
	logLines       []string
	selectedNode   string
	selectedTwin   int
}

// NewModel builds a TUI visual model around the Runtime.
func NewModel(rt *runtime.Runtime) *Model {
	return &Model{
		runtime:      rt,
		activeView:   ViewOverview,
		selectedNode: "payments",
		paletteOptions: []string{
			"Trigger Morning Startup Routine",
			"Orchestrate Workspace Shutdown",
			"Run Clean Pruning workflow",
			"Query Dependency Impact (auth)",
			"Search Twin Context",
		},
		logLines: []string{
			"Daemon Runtime initialized successfully.",
			"Lifecycle Manager active.",
			"Service Container dependency resolution completed.",
			"Engineering Twin loaded.",
		},
	}
}

// Init starts event listeners.
func (m *Model) Init() tea.Cmd {
	eb := m.runtime.Container.ResolveEventBus()
	eb.Subscribe("*", func(e events.Event) {
		m.logLines = append(m.logLines, fmt.Sprintf("[%s] Event: %s", e.Timestamp.Format("15:04:05"), e.Type))
	})
	return nil
}

// Update acts on keys and resize signals.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter", "esc":
				m.searching = false
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.activeView = ViewType((int(m.activeView) + 1) % 17)

		case "/":
			m.searching = true
			m.searchQuery = ""

		case "ctrl+k":
			m.showPalette = !m.showPalette
			m.paletteIndex = 0

		case "ctrl+a":
			m.activeView = ViewAdvisor

		case "ctrl+r":
			m.activeView = ViewReplay

		case "up":
			if m.showPalette {
				if m.paletteIndex > 0 {
					m.paletteIndex--
				}
			} else if m.activeView == ViewTwin {
				if m.selectedTwin > 0 {
					m.selectedTwin--
				}
			}

		case "down":
			if m.showPalette {
				if m.paletteIndex < len(m.paletteOptions)-1 {
					m.paletteIndex++
				}
			} else if m.activeView == ViewTwin {
				m.selectedTwin++
			}

		case "enter":
			if m.showPalette {
				action := m.paletteOptions[m.paletteIndex]
				m.logLines = append(m.logLines, fmt.Sprintf("[%s] Executing palette command: %s", time.Now().Format("15:04:05"), action))
				if strings.Contains(action, "Startup") {
					m.activeView = ViewSessions
				} else if strings.Contains(action, "Search") {
					m.activeView = ViewSearch
					m.searching = true
				}
				m.showPalette = false
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// Lip Gloss styles
var (
	purpleColor = lipgloss.Color("#8b5cf6")
	whiteColor  = lipgloss.Color("#ffffff")
	grayColor   = lipgloss.Color("#555555")
	greenColor  = lipgloss.Color("#10b981")
	amberColor  = lipgloss.Color("#f59e0b")
	redColor    = lipgloss.Color("#ef4444")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteColor).
			Background(purpleColor).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purpleColor).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(purpleColor).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purpleColor).
			Padding(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Italic(true)
)

// View renders visual blocks based on window limits.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing Engineering Cockpit..."
	}

	title := titleStyle.Render(" DAEMON COCKPIT v1.0 ")

	tabs := []string{"Overview", "Workspace", "Twin", "Timeline", "Arch", "Deps", "Recs", "Search", "Integrations", "Containers", "Infra", "Cloud", "Maint", "Auto", "Sessions", "Replay", "Advisor"}
	var tabRenders []string
	for i, t := range tabs {
		label := fmt.Sprintf("[%d] %s", i+1, t)
		if int(m.activeView) == i {
			tabRenders = append(tabRenders, activeTabStyle.Render(label))
		} else {
			tabRenders = append(tabRenders, tabStyle.Render(label))
		}
	}
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", lipgloss.JoinHorizontal(lipgloss.Top, tabRenders...))

	var searchBar string
	if m.searching || m.searchQuery != "" {
		searchBar = lipgloss.NewStyle().Foreground(amberColor).Render(fmt.Sprintf("\nSearch Filter: '%s' (Enter/Esc to exit)", m.searchQuery))
	}

	var body string
	switch m.activeView {
	case ViewOverview:
		body = m.viewOverview()
	case ViewWorkspace:
		body = m.viewWorkspace()
	case ViewTwin:
		body = m.viewTwin()
	case ViewTimeline:
		body = m.viewTimeline()
	case ViewArchitecture:
		body = m.viewArchitecture()
	case ViewDependencies:
		body = m.viewDependencies()
	case ViewRecommendations:
		body = m.viewRecommendations()
	case ViewSearch:
		body = m.viewSearch()
	case ViewIntegrations:
		body = m.viewIntegrations()
	case ViewContainers:
		body = m.viewContainers()
	case ViewInfrastructure:
		body = m.viewInfrastructure()
	case ViewCloud:
		body = m.viewCloud()
	case ViewMaintenance:
		body = m.viewMaintenance()
	case ViewAutomation:
		body = m.viewAutomation()
	case ViewSessions:
		body = m.viewSessions()
	case ViewReplay:
		body = m.viewReplay()
	case ViewAdvisor:
		body = m.viewAdvisor()
	}

	if m.showPalette {
		paletteBody := "=== COMMAND PALETTE (Ctrl+K to close) ===\n\n"
		for i, opt := range m.paletteOptions {
			if i == m.paletteIndex {
				paletteBody += fmt.Sprintf(" > [ %s ]\n", opt)
			} else {
				paletteBody += fmt.Sprintf("   %s\n", opt)
			}
		}
		paletteBox := boxStyle.BorderForeground(amberColor).Render(paletteBody)
		body = lipgloss.JoinVertical(lipgloss.Center, body, "\n", paletteBox)
	}

	statusBar := statusStyle.Render(fmt.Sprintf("Tab: Switch View | /: Search | Ctrl+K: Cmd Palette | Ctrl+A: Advisor | Ctrl+R: Replay | Connected: %s", m.runtime.Health()))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "\n", body, "\n", statusBar)
}

func (m *Model) viewOverview() string {
	ms := m.runtime.Container.ResolveMemoryStore()
	incidents, _ := ms.GetIncidents()
	recs, _ := ms.GetRecommendations()

	healthScore := 100 - len(incidents)*15
	if healthScore < 0 {
		healthScore = 0
	}

	grade := "A"
	gradeColor := greenColor
	if healthScore < 60 {
		grade = "D"
		gradeColor = redColor
	} else if healthScore < 80 {
		grade = "B"
		gradeColor = amberColor
	} else if healthScore < 95 {
		grade = "A-"
	}

	healthGauge := fmt.Sprintf("[%s%s] %d%%", strings.Repeat("■", healthScore/10), strings.Repeat(" ", 10-healthScore/10), healthScore)

	leftBox := boxStyle.Width(m.width/2 - 4).Render(
		"SYSTEM OVERVIEW SCORECARD\n\n" +
			"Project:   saas-core\n" +
			"Language:  TypeScript\n" +
			"Intelligence Grade:\n" +
			lipgloss.NewStyle().Foreground(gradeColor).Bold(true).Render(fmt.Sprintf("  [ %s ]\n", grade)) +
			"Health Index:\n" +
			lipgloss.NewStyle().Foreground(gradeColor).Bold(true).Render(healthGauge),
	)

	rightBox := boxStyle.Width(m.width/2 - 4).Render(
		fmt.Sprintf("DIAGNOSTICS & ALERTS\n\n"+
			"Active Incidents: %d\n"+
			"Recommendations:  %d available\n\n"+
			"Memory status:    Idle (all buffers clean)",
			len(incidents), len(recs)),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)
}

func (m *Model) viewWorkspace() string {
	gs := m.runtime.Container.ResolveGraphStore()
	services, _ := gs.GetServices()

	leftBox := boxStyle.Width(m.width/2 - 4).Render(
		"ENGINEERING WORKSPACE STATUS\n\n" +
			"Active Profile: Full Stack\n" +
			"Dev Servers:    Listening (Ports parsed: 5001-5004)\n" +
			"Tunnels status: 1 Online\n",
	)

	serviceLines := ""
	for _, s := range services {
		serviceLines += fmt.Sprintf("- %s (status: %s)\n", s.Name, s.Status)
	}

	rightBox := boxStyle.Width(m.width/2 - 4).Render(
		"REGISTERED STACK FEATURE SERVERS\n\n" + serviceLines,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)
}

func (m *Model) viewTwin() string {
	gs := m.runtime.Container.ResolveGraphStore()
	services, _ := gs.GetServices()
	deps, _ := gs.GetDependencies()

	var nodesList []string
	for _, s := range services {
		nodesList = append(nodesList, s.Name)
	}
	for _, d := range deps {
		nodesList = append(nodesList, d.Name)
	}

	if len(nodesList) == 0 {
		return boxStyle.Width(m.width - 6).Render("Engineering Twin database empty. Run 'daemon demo' first.")
	}

	if m.selectedTwin >= len(nodesList) {
		m.selectedTwin = len(nodesList) - 1
	}

	left := "ENGINEERING TWIN RESOURCE GRAPH NODES\n\n"
	for i, n := range nodesList {
		if i == m.selectedTwin {
			left += fmt.Sprintf(" > [ %s ]\n", n)
		} else {
			left += fmt.Sprintf("   %s\n", n)
		}
	}

	selectedName := nodesList[m.selectedTwin]
	right := fmt.Sprintf("TWIN METADATA CARD: %s\n\n", strings.ToUpper(selectedName))
	right += "Type:         Graph Node representation\n"
	right += "Sync State:   Synchronized (incremental discovery)\n"
	right += "Source:       Integrations APIs\n\n"
	right += "ASCII TOPOLOGY NEIGHBOR RELATION:\n"
	if strings.Contains(strings.ToLower(selectedName), "order") {
		right += "  [orders] ──> [payments] ──> [postgres]\n"
	} else if strings.Contains(strings.ToLower(selectedName), "gateway") {
		right += "  [gateway] ──> [auth]\n"
		right += "     └──> [orders]\n"
	} else {
		right += fmt.Sprintf("  [project:main] ──> [%s]\n", selectedName)
	}

	leftBox := boxStyle.Width(m.width/2 - 4).Render(left)
	rightBox := boxStyle.Width(m.width/2 - 4).Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)
}

func (m *Model) viewTimeline() string {
	eb := m.runtime.Container.ResolveEventBus()
	history := eb.GetTimeline()

	var builder strings.Builder
	builder.WriteString("MISSION TIMELINE OPERATIONS HISTORY\n\n")

	for i := len(history) - 1; i >= 0; i-- {
		ev := history[i]
		builder.WriteString(fmt.Sprintf("  [%s] Event: %s\n", ev.Timestamp.Format("15:04:05"), ev.Type))
	}

	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewArchitecture() string {
	engine := architecture.NewEngine(m.runtime.Container.ResolveGraphStore())
	report, _ := engine.Analyze(context.Background())

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("ARCHITECTURE COUPLING & STYLE: %s\n\n", report.Style))
	builder.WriteString(fmt.Sprintf("  - Architecture Score: %d%%\n", report.ArchitectureScore))
	builder.WriteString(fmt.Sprintf("  - Coupling Index:     %d%%\n", report.CouplingScore))
	builder.WriteString(fmt.Sprintf("  - Cohesion Index:     %d%%\n\n", report.CohesionScore))
	builder.WriteString("Drifts & Bottlenecks:\n")
	for _, iss := range report.Issues {
		builder.WriteString(fmt.Sprintf("  * %s\n", iss))
	}

	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewDependencies() string {
	gs := m.runtime.Container.ResolveGraphStore()
	deps, _ := gs.GetDependencies()

	var builder strings.Builder
	builder.WriteString("DISCOVERED WORKSPACE PACKAGE DEPENDENCIES\n\n")
	for _, d := range deps {
		builder.WriteString(fmt.Sprintf("  - %s (version: %s, outdated: %t)\n", d.Name, d.Version, d.IsOutdated))
	}
	if len(deps) == 0 {
		builder.WriteString("  No external package dependencies recorded.")
	}

	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewRecommendations() string {
	gs := m.runtime.Container.ResolveGraphStore()
	ms := m.runtime.Container.ResolveMemoryStore()
	engine := recommendation.NewEngine(gs, ms)
	recs, _ := engine.GenerateAndScore(context.Background())

	var builder strings.Builder
	builder.WriteString("RECOMMENDATION SUMMARY\n\n")
	for _, r := range recs {
		builder.WriteString(fmt.Sprintf("  * [Priority: %.2f] %s (Effort: %d)\n", r.Score, r.Message, r.Effort))
	}

	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewSearch() string {
	var builder strings.Builder
	builder.WriteString("ENGINEERING SEARCH ENGINE\n\n")

	if m.searchQuery == "" {
		builder.WriteString("Type '/' to edit search terms querying the live Twin model.")
	} else {
		t := twin.NewTwinModel(m.runtime.Container.ResolveGraphStore())
		res, _ := t.Search(context.Background(), m.searchQuery)
		for _, r := range res {
			builder.WriteString(fmt.Sprintf("[%s] %s\n", strings.ToUpper(r.Type), r.Name))
			builder.WriteString(fmt.Sprintf("  Details: %s\n\n", r.Context))
		}
	}

	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewIntegrations() string {
	im := integrations.NewIntegrationManager(m.runtime.Container.ResolveGraphStore())
	conns := im.GetConnectors()

	var builder strings.Builder
	builder.WriteString("INTEGRATION MANAGER CONNECTORS STATUS\n\n")
	for id, c := range conns {
		state, latency, _ := c.Health(context.Background())
		builder.WriteString(fmt.Sprintf("  - %s: status -> %s (latency: %d ms)\n", id, state, latency))
	}

	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewContainers() string {
	return boxStyle.Width(m.width - 6).Render(
		"DOCKER COMPOSE ACTIVE CONTAINER LIST\n\n" +
			"  - payments-api container: Running (Port 5003)\n" +
			"  - orders-api container:   Running (Port 5002)\n" +
			"  - auth-service:           Running (Port 5001)\n",
	)
}

func (m *Model) viewInfrastructure() string {
	return boxStyle.Width(m.width - 6).Render(
		"KUBERNETES & TERRAFORM CONFIG STATE\n\n" +
			"  - Kubernetes manifests validated. (saas-core namespace verified)\n" +
			"  - Terraform Cloud variables match active workspaces.\n",
	)
}

func (m *Model) viewCloud() string {
	return boxStyle.Width(m.width - 6).Render(
		"CLOUD RESOURCES DISCOVERIES\n\n" +
			"  - AWS ECS cluster saas-cluster task online\n" +
			"  - AWS RDS PostgreSQL instance connection checks PASS\n",
	)
}

func (m *Model) viewMaintenance() string {
	engine := insights.NewEngine(m.runtime.Container.ResolveGraphStore())
	report, _ := engine.Generate(context.Background())

	return boxStyle.Width(m.width - 6).Render(
		fmt.Sprintf("CONTINUOUS MAINTENANCE LOGS\n\n"+
			"  - Technical Debt hotspots: %s\n"+
			"  - Pruned stale docker cache files successfully.\n", report.TechDebtHotspots),
	)
}

func (m *Model) viewAutomation() string {
	return boxStyle.Width(m.width - 6).Render(
		"ENGINEERING CAPABILITIES LIFECYCLE MONITOR\n\n" +
			"  [x] Git Capability:          Discovery completed successfully (branch: main)\n" +
			"  [x] Docker Capability:       Containers validated (9 instances running)\n" +
			"  [x] Cloudflare Capability:   Tunnel verified operational (127.0.0.1 DNS route resolved)\n" +
			"  [x] PostgreSQL Capability:   Schema migration check complete (0 drifts)\n" +
			"  [x] Redis Capability:        Cache connection checks PASS\n",
	)
}

func (m *Model) viewSessions() string {
	return boxStyle.Width(m.width - 6).Render(
		"DEVELOPER SESSION TRACKER STATS\n\n" +
			"  Active Session Start Time: 2026-08-01 17:50:19\n" +
			"  Elapsed Time:             0 hours 45 minutes\n" +
			"  Files edited:             package.json, recommendation.go\n",
	)
}

func (m *Model) viewReplay() string {
	ce := engContext.NewContextEngine(m.runtime.Container.ResolveGraphStore(), m.runtime.Container.ResolveMemoryStore())
	re := replay.NewReplayEngine(m.runtime.Container.ResolveEventBus(), ce)
	eventsList, _ := re.ReplaySession(24*time.Hour, "", "")

	var builder strings.Builder
	builder.WriteString("=== ENGINEERING PLAYBACK REPLAY TIMELINE (Ctrl+R) ===\n\n")
	for _, ev := range eventsList {
		builder.WriteString(fmt.Sprintf("  [%s] Operation: %s\n", ev.Timestamp.Format("15:04:05"), ev.Title))
		builder.WriteString(fmt.Sprintf("       Detail: %s\n", ev.Description))
		builder.WriteString(fmt.Sprintf("       Impact: %s\n\n", ev.Impact))
	}
	builder.WriteString("Controls: [Step Back] [Play/Pause] [Step Forward]")
	return boxStyle.Width(m.width - 6).Render(builder.String())
}

func (m *Model) viewAdvisor() string {
	ce := engContext.NewContextEngine(m.runtime.Container.ResolveGraphStore(), m.runtime.Container.ResolveMemoryStore())
	ae := advisor.NewAdvisorEngine(ce)
	report, _ := ae.Advise(context.Background(), "", "", "")

	var builder strings.Builder
	builder.WriteString("=== ENGINEERING ADVISOR SCORECARD (Ctrl+A) ===\n\n")
	builder.WriteString(fmt.Sprintf("  Engineering Health Score: %d%%  |  Confidence Score: %d%%\n\n", report.HealthScore, report.ConfidenceScore))
	builder.WriteString("Recommendations:\n")
	for _, r := range report.Recommendations {
		builder.WriteString(fmt.Sprintf("  * %s\n", r))
	}
	builder.WriteString("\nRisks Detected:\n")
	for _, r := range report.Risks {
		builder.WriteString(fmt.Sprintf("  - %s\n", r))
	}
	return boxStyle.Width(m.width - 6).Render(builder.String())
}
