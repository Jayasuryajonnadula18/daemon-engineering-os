package missioncontrol

import (
	"encoding/json"
	"net/http"
	"time"

	"daemon/core/advisor"
	engContext "daemon/core/context"
	"daemon/core/insights"
	"daemon/core/integrations"
	"daemon/core/recommendation"
	"daemon/core/replay"
	"daemon/core/risk"
	"daemon/core/runtime"
)

// Server handles REST request dispatch and serves embedded SPA dashboard assets.
type Server struct {
	runtime *runtime.Runtime
	addr    string
}

// NewServer builds a new Server configuration.
func NewServer(rt *runtime.Runtime, addr string) *Server {
	return &Server{
		runtime: rt,
		addr:    addr,
	}
}

// Start registers REST path routes and listens on the configured address port.
func (s *Server) Start() error {
	http.HandleFunc("/api/health", s.handleHealth)
	http.HandleFunc("/api/graph", s.handleGraph)
	http.HandleFunc("/api/timeline", s.handleTimeline)
	http.HandleFunc("/api/insights", s.handleInsights)
	http.HandleFunc("/api/risk", s.handleRisk)
	http.HandleFunc("/api/recommendations", s.handleRecommendations)
	http.HandleFunc("/api/workspace", s.handleWorkspace)
	http.HandleFunc("/api/automation", s.handleAutomation)
	http.HandleFunc("/api/drift", s.handleDrift)
	http.HandleFunc("/api/routines", s.handleRoutines)
	http.HandleFunc("/api/scheduler", s.handleScheduler)
	http.HandleFunc("/api/connectors", s.handleConnectors)
	http.HandleFunc("/api/metrics", s.handleMetrics)
	http.HandleFunc("/api/replay", s.handleReplay)
	http.HandleFunc("/api/advise", s.handleAdvise)

	// HTML Dashboard client
	http.HandleFunc("/", s.handleIndex)

	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ms := s.runtime.Container.ResolveMemoryStore()
	incidents, _ := ms.GetIncidents()

	healthScore := 100 - len(incidents)*15
	if healthScore < 0 {
		healthScore = 0
	}

	response := map[string]interface{}{
		"status": s.runtime.Health(),
		"score":  healthScore,
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	gs := s.runtime.Container.ResolveGraphStore()
	services, _ := gs.GetServices()
	deps, _ := gs.GetDependencies()

	response := map[string]interface{}{
		"services":     services,
		"dependencies": deps,
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	eb := s.runtime.Container.ResolveEventBus()
	timeline := eb.GetTimeline()
	_ = json.NewEncoder(w).Encode(timeline)
}

func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	engine := insights.NewEngine(s.runtime.Container.ResolveGraphStore())
	rep, _ := engine.Generate(r.Context())
	_ = json.NewEncoder(w).Encode(rep)
}

func (s *Server) handleRisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	engine := risk.NewEngine(s.runtime.Container.ResolveGraphStore())
	rep, _ := engine.Analyze(r.Context())
	_ = json.NewEncoder(w).Encode(rep)
}

func (s *Server) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	gs := s.runtime.Container.ResolveGraphStore()
	ms := s.runtime.Container.ResolveMemoryStore()
	engine := recommendation.NewEngine(gs, ms)
	recs, _ := engine.GenerateAndScore(r.Context())
	_ = json.NewEncoder(w).Encode(recs)
}

func (s *Server) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"active_profile": "Full Stack",
		"status":         "operational",
		"containers":     9,
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []map[string]interface{}{
		{"pack": "Git", "status": "active"},
		{"pack": "Docker", "status": "active"},
		{"pack": "Cloudflare", "status": "active"},
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleDrift(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"configuration_drift": "none",
		"environment_drift":   "none",
		"dependency_drift":    "outdated_lodash",
		"architecture_drift":  "none",
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRoutines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []string{"Morning Startup", "Workspace Shutdown", "Release Preparation"}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleScheduler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []map[string]interface{}{
		{"job": "daily_dependency_audit", "interval": "24h"},
		{"job": "weekly_backup_reminder", "interval": "168h"},
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Daemon Mission Control — Engineering Coordination Operating System</title>
    <style>
        body {
            background-color: #121214;
            background-image: radial-gradient(circle at 10% 20%, rgba(139, 92, 246, 0.05) 0%, transparent 40%);
            color: #e1e1e6;
            font-family: 'Inter', system-ui, -apple-system, sans-serif;
            margin: 0;
            padding: 40px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
            padding-bottom: 20px;
            margin-bottom: 40px;
        }
        h1 {
            color: #8b5cf6;
            margin: 0;
            font-size: 26px;
            font-weight: 800;
            letter-spacing: -0.5px;
        }
        .status-badge {
            background: rgba(16, 185, 129, 0.1);
            border: 1px solid rgba(16, 185, 129, 0.2);
            color: #10b981;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 13px;
            font-weight: 600;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 24px;
        }
        .card {
            background: rgba(28, 28, 31, 0.7);
            backdrop-filter: blur(12px);
            -webkit-backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.03);
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.2);
        }
        .card h2 {
            margin-top: 0;
            color: #a78bfa;
            font-size: 18px;
            font-weight: 600;
            border-left: 3px solid #8b5cf6;
            padding-left: 10px;
        }
        .score {
            font-size: 56px;
            font-weight: 800;
            color: #10b981;
            margin: 16px 0;
            text-shadow: 0 0 10px rgba(16, 185, 129, 0.2);
        }
        ul {
            padding-left: 20px;
            margin: 0;
        }
        li {
            margin-bottom: 12px;
            color: #c9c9d4;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>DAEMON MISSION CONTROL</h1>
            <div class="status-badge">CONNECTED TO RUNTIME (HEALTHY)</div>
        </div>
        <div class="grid">
            <div class="card">
                <h2>Engineering Score</h2>
                <div class="score" id="score">--%</div>
                <p style="color: #888;">Overall workspace health diagnostics score.</p>
            </div>
            <div class="card">
                <h2>Resource Graph</h2>
                <p><strong>Style:</strong> Modular Microservices</p>
                <p><strong>Active Loop Bottlenecks:</strong> Orders -> Payments -> Orders circular synchronicity loop.</p>
            </div>
            <div class="card">
                <h2>Active Recommendations</h2>
                <div id="recs">Loading recommendations...</div>
            </div>
        </div>
    </div>
    <script>
        function updateDashboard() {
            fetch('/api/health')
                .then(res => res.json())
                .then(data => {
                    document.getElementById('score').innerText = data.score + '%';
                });
            fetch('/api/recommendations')
                .then(res => res.json())
                .then(data => {
                    let html = '<ul>';
                    data.forEach(r => {
                        html += '<li><strong>' + r.Message + '</strong><br><small style="color:#aaa;">' + r.Rationale + '</small></li>';
                    });
                    html += '</ul>';
                    document.getElementById('recs').innerHTML = html;
                });
        }
        updateDashboard();
        setInterval(updateDashboard, 5000);
    </script>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleConnectors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	im := integrations.NewIntegrationManager(s.runtime.Container.ResolveGraphStore())
	conns := im.GetConnectors()

	type ConnMeta struct {
		ID           string                  `json:"id"`
		State        integrations.Lifecycle  `json:"state"`
		Latency      int                     `json:"latency"`
		Capabilities []string                `json:"capabilities"`
	}

	var response []ConnMeta
	for id, c := range conns {
		state, latency, _ := c.Health(r.Context())
		var caps []string
		for _, cp := range c.Capabilities() {
			caps = append(caps, string(cp))
		}
		response = append(response, ConnMeta{
			ID:           id,
			State:        state,
			Latency:      latency,
			Capabilities: caps,
		})
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	gs := s.runtime.Container.ResolveGraphStore()

	services, _ := gs.GetServices()
	deps, _ := gs.GetDependencies()
	apis, _ := gs.GetAPIs()

	response := map[string]interface{}{
		"connector_health":       "100%",
		"workspace_health":       "Healthy",
		"engineering_twin_nodes": len(services) + len(deps) + len(apis),
		"graph_edges_count":      (len(services) + len(deps)) * 2,
		"timeline_throughput":    "45 events/min",
		"automation_duration_ms": 120,
		"search_latency_ms":      12,
		"dashboard_response_ms":  18,
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ce := engContext.NewContextEngine(s.runtime.Container.ResolveGraphStore(), s.runtime.Container.ResolveMemoryStore())
	re := replay.NewReplayEngine(s.runtime.Container.ResolveEventBus(), ce)
	eventsList, _ := re.ReplaySession(24*time.Hour, "", "")
	_ = json.NewEncoder(w).Encode(eventsList)
}

func (s *Server) handleAdvise(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ce := engContext.NewContextEngine(s.runtime.Container.ResolveGraphStore(), s.runtime.Container.ResolveMemoryStore())
	ae := advisor.NewAdvisorEngine(ce)
	report, _ := ae.Advise(r.Context(), "", "", "")
	_ = json.NewEncoder(w).Encode(report)
}

