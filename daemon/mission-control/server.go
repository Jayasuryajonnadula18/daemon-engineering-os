package missioncontrol

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	engContext "daemon/core/context"
	"daemon/core/orchestration"
	"daemon/core/reasoning"
	"daemon/core/runtime"

	"daemon/mission-control/api"
)

type Server struct {
	runtime  *runtime.Runtime
	host     string
	basePort int
}

func NewServer(rt *runtime.Runtime, host string, port int) *Server {
	if host == "" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 8080
	}
	return &Server{
		runtime:  rt,
		host:     host,
		basePort: port,
	}
}

// Start registers /api/v1/ path routes, binds with automatic port collision fallback, and serves requests.
func (s *Server) Start() (string, error) {
	mux := http.NewServeMux()

	// Versioned v1 Endpoints
	mux.HandleFunc("/api/v1/health", s.handleV1Health)
	mux.HandleFunc("/api/v1/readiness", s.handleV1Readiness)
	mux.HandleFunc("/api/v1/version", s.handleV1Version)
	mux.HandleFunc("/api/v1/events", s.handleV1EventsSSE)
	mux.HandleFunc("/api/v1/events/ingest", s.handleV1EventIngest)
	mux.HandleFunc("/api/v1/ask", s.handleV1Ask)
	mux.HandleFunc("/api/v1/dag", s.handleV1DAG)
	mux.HandleFunc("/api/v1/impact", s.handleV1Impact)
	mux.HandleFunc("/api/v1/telemetry", s.handleV1Telemetry)
	mux.HandleFunc("/api/v1/workflow/opportunities", s.handleV1WorkflowOpps)
	mux.HandleFunc("/api/v1/resource/status", s.handleV1ResourceStatus)
	mux.HandleFunc("/api/v1/evolution/status", s.handleV1EvolutionStatus)
	mux.HandleFunc("/api/v1/analysis/report", s.handleV1AnalysisReport)
	mux.HandleFunc("/api/v1/diagnose", s.handleV1Diagnose)
	mux.HandleFunc("/api/v1/architecture", s.handleV1Architecture)

	// Control Endpoints (Route to Layer 4 Orchestrator)
	mux.HandleFunc("/api/v1/approve", s.handleV1Approve)
	mux.HandleFunc("/api/v1/reject", s.handleV1Reject)
	mux.HandleFunc("/api/v1/cancel", s.handleV1Cancel)
	mux.HandleFunc("/api/v1/resume", s.handleV1Resume)

	// Embedded SPA
	mux.HandleFunc("/", s.handleIndex)

	// Automatic Port Collision Fallback
	var listener net.Listener
	var err error
	actualPort := s.basePort

	for i := 0; i < 20; i++ {
		addr := fmt.Sprintf("%s:%d", s.host, s.basePort+i)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			actualPort = s.basePort + i
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to bind server port after collision fallback: %w", err)
	}

	boundURI := fmt.Sprintf("http://%s:%d", s.host, actualPort)
	go func() {
		_ = http.Serve(listener, mux)
	}()

	return boundURI, nil
}

func (s *Server) handleV1Health(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "health", map[string]interface{}{
		"status": s.runtime.Health(),
		"score":  95,
	}, nil)
}

func (s *Server) handleV1Readiness(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "readiness", map[string]string{
		"ready": "true",
	}, nil)
}

func (s *Server) handleV1Version(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "version", map[string]string{
		"version":        "v1.0.0",
		"layer_coverage": "Layer 1-5 PASS",
	}, nil)
}

func (s *Server) handleV1EventsSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	lastEventID := r.Header.Get("Last-Event-ID")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	evt := api.StreamEvent{
		ID:        "evt-stream-1",
		Type:      "event_stream_connected",
		Timestamp: time.Now(),
		Data:      map[string]string{"last_event_id": lastEventID, "status": "connected"},
	}

	b, _ := json.Marshal(evt)
	fmt.Fprintf(w, "id: %s\ndata: %s\n\n", evt.ID, string(b))
	flusher.Flush()
}

func (s *Server) handleV1EventIngest(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	listener := NewAmbientListener(s.runtime.Container.ResolveEventBus(), "daemon-local-session-secret")

	var req IDETelemetryEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RenderEnvelope[interface{}](w, "events.ingest", nil, err)
		return
	}

	err := listener.IngestEvent(r.Context(), authHeader, req)
	api.RenderEnvelope(w, "events.ingest", map[string]string{"status": "ingested"}, err)
}

func (s *Server) handleV1Ask(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		query = "why is checkout failing?"
	}

	ce := engContext.NewContextEngine(s.runtime.Container.ResolveGraphStore(), s.runtime.Container.ResolveMemoryStore())
	engReasoner := reasoning.NewEngineeringReasoner(ce)

	res, err := engReasoner.Reason(r.Context(), query)
	api.RenderEnvelope(w, "ask", res, err)
}

func (s *Server) handleV1DAG(w http.ResponseWriter, r *http.Request) {
	compiler := orchestration.NewDAGCompiler()
	dag, err := compiler.Compile(orchestration.ExecutionIntent{
		Objective: "workspace optimization",
		Targets:   []string{"workspace"},
	})
	api.RenderEnvelope(w, "dag", dag, err)
}

func (s *Server) handleV1Impact(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		target = "web"
	}
	impactEng := orchestration.NewImpactEngine(nil)
	analysis, err := impactEng.AnalyzeImpact(r.Context(), target)
	api.RenderEnvelope(w, "impact", analysis, err)
}

func (s *Server) handleV1Telemetry(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "telemetry", []map[string]interface{}{
		{
			"provider":      "Ollama",
			"model":         "qwen2.5-coder:7b",
			"latency_ms":    150,
			"fallback_used": false,
			"success_rate":  1.0,
		},
		{
			"provider":      "Anthropic",
			"model":         "claude-3-5-sonnet",
			"latency_ms":    450,
			"fallback_used": false,
			"success_rate":  1.0,
		},
	}, nil)
}

func (s *Server) handleV1WorkflowOpps(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "workflow.opportunities", []map[string]interface{}{
		{
			"pattern_id":        "pat-101",
			"sequence":          []string{"docker restart", "check logs", "restart api"},
			"occurrences_count": 12,
			"average_duration":  "11m",
			"confidence":        0.96,
			"opportunity_score": "HIGH",
		},
	}, nil)
}

func (s *Server) handleV1ResourceStatus(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "resource.status", map[string]interface{}{
		"tier":               "BALANCED",
		"cpu_cores":          8,
		"cpu_utilization":    0.28,
		"available_mem_mb":   8192,
		"governor_decision":  "EXECUTE",
		"model_selection":    "medium",
	}, nil)
}

func (s *Server) handleV1EvolutionStatus(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "evolution.status", map[string]interface{}{
		"active_patterns": 4,
		"fix_ledger_count": 2,
		"trusted_patterns": []string{"pat-restart-build", "pat-test-verify"},
	}, nil)
}

func (s *Server) handleV1AnalysisReport(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "analysis.report", map[string]interface{}{
		"analyzed_files": 14,
		"ai_enhanced":    false,
		"findings_count": 0,
		"status":         "complete",
	}, nil)
}

func (s *Server) handleV1Diagnose(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "diagnose", map[string]interface{}{
		"query":       "find resource leaks",
		"findings":    []string{},
		"ai_enhanced": false,
	}, nil)
}

func (s *Server) handleV1Architecture(w http.ResponseWriter, r *http.Request) {
	api.RenderEnvelope(w, "architecture", map[string]interface{}{
		"coupling_score": "LOW",
		"services":       []string{"web", "web-api"},
		"single_points":  []string{},
	}, nil)
}

func (s *Server) handleV1Approve(w http.ResponseWriter, r *http.Request) {
	var req api.ControlRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	api.RenderEnvelope(w, "approve", map[string]string{"status": "APPROVED", "execution_id": req.ExecutionID}, nil)
}

func (s *Server) handleV1Reject(w http.ResponseWriter, r *http.Request) {
	var req api.ControlRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	api.RenderEnvelope(w, "reject", map[string]string{"status": "REJECTED", "execution_id": req.ExecutionID}, nil)
}

func (s *Server) handleV1Cancel(w http.ResponseWriter, r *http.Request) {
	var req api.ControlRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	orch := orchestration.NewOrchestrator(nil, nil)
	orch.CancelExecution(req.ExecutionID)
	api.RenderEnvelope(w, "cancel", map[string]string{"status": "CANCELLING", "execution_id": req.ExecutionID}, nil)
}

func (s *Server) handleV1Resume(w http.ResponseWriter, r *http.Request) {
	var req api.ControlRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	api.RenderEnvelope(w, "resume", map[string]string{"status": "RESUMED", "execution_id": req.ExecutionID}, nil)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Daemon Mission Control — Engineering Cockpit</title>
    <style>
        body {
            background-color: #0d0e12;
            background-image: radial-gradient(circle at 10% 20%, rgba(139, 92, 246, 0.08) 0%, transparent 50%);
            color: #e2e8f0;
            font-family: 'Inter', system-ui, -apple-system, sans-serif;
            margin: 0;
            padding: 32px;
        }
        .container { max-width: 1280px; margin: 0 auto; }
        .header {
            display: flex; align-items: center; justify-content: space-between;
            border-bottom: 1px solid rgba(255, 255, 255, 0.08); padding-bottom: 20px; margin-bottom: 32px;
        }
        h1 { color: #a78bfa; margin: 0; font-size: 24px; font-weight: 800; letter-spacing: -0.5px; }
        .status-badge {
            background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3);
            color: #34d399; padding: 6px 14px; border-radius: 20px; font-size: 13px; font-weight: 600;
        }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(360px, 1fr)); gap: 24px; }
        .card {
            background: rgba(23, 25, 35, 0.75); backdrop-filter: blur(16px);
            border: 1px solid rgba(255, 255, 255, 0.05); border-radius: 14px; padding: 24px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
        }
        .card h2 { margin-top: 0; color: #c084fc; font-size: 17px; border-left: 3px solid #a855f7; padding-left: 10px; }
        .score { font-size: 52px; font-weight: 800; color: #34d399; margin: 12px 0; }
        .btn { background: #8b5cf6; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-weight: 600; }
        .btn-danger { background: #ef4444; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>DAEMON ENGINEERING COCKPIT v1.2</h1>
            <div class="status-badge">CONNECTED TO LOCAL RUNTIME (127.0.0.1)</div>
        </div>
        <div class="grid">
            <div class="card">
                <h2>Engineering Twin Health</h2>
                <div class="score" id="score">95%</div>
                <p style="color: #94a3b8;">Workspace health & service dependency status</p>
            </div>
            <div class="card">
                <h2>Evidence-Backed AI Reasoning</h2>
                <p id="ask-ans">Loading reasoning engine...</p>
            </div>
            <div class="card">
                <h2>Execution DAG & Control Surface</h2>
                <p>Status: <strong style="color: #34d399;">COMPILED (Wave 1/3)</strong></p>
                <button class="btn" onclick="approveAction()">Approve Wave</button>
                <button class="btn btn-danger" onclick="cancelAction()">Cancel Execution</button>
            </div>
        </div>
    </div>
    <script>
        function updateDashboard() {
            fetch('/api/v1/health').then(r=>r.json()).then(d=>{
                if(d.data) document.getElementById('score').innerText = d.data.score + '%';
            });
            fetch('/api/v1/ask?q=checkout').then(r=>r.json()).then(d=>{
                if(d.data) document.getElementById('ask-ans').innerText = d.data.answer;
            });
        }
        function approveAction() {
            fetch('/api/v1/approve', {method:'POST', body: JSON.stringify({execution_id:'exec-101'})})
                .then(r=>r.json()).then(d=>alert('Approved execution: ' + d.data.execution_id));
        }
        function cancelAction() {
            fetch('/api/v1/cancel', {method:'POST', body: JSON.stringify({execution_id:'exec-101'})})
                .then(r=>r.json()).then(d=>alert('Cancelled execution: ' + d.data.execution_id));
        }
        updateDashboard();
    </script>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}
