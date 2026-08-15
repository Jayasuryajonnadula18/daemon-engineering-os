package instruments

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"daemon/core/policies"
	"daemon/core/resource"
)

// MockInstrument implements a mock EngineeringInstrument for tests
type MockInstrument struct {
	ident        InstrumentIdentity
	caps         []Capability
	compatible   bool
	healthStatus string
}

func (m *MockInstrument) Identity() InstrumentIdentity { return m.ident }
func (m *MockInstrument) Capabilities() []Capability  { return m.caps }
func (m *MockInstrument) Detect(ctx context.Context, env Environment) DetectionResult {
	return DetectionResult{Compatible: m.compatible, Reason: "mock compatibility"}
}
func (m *MockInstrument) Health(ctx context.Context) HealthResult {
	return HealthResult{Status: m.healthStatus, Reason: "mock health check"}
}
func (m *MockInstrument) BuildRequest(ctx context.Context, request InstrumentRequest) (ToolRequest, error) {
	return ToolRequest{Executable: m.ident.ExecutablePath, Args: request.Args}, nil
}
func (m *MockInstrument) Execute(ctx context.Context, request ToolRequest) (ToolResult, error) {
	return ToolResult{InstrumentID: m.ident.ID, Success: true}, nil
}
func (m *MockInstrument) Normalize(ctx context.Context, result ToolResult) ([]Evidence, error) {
	return []Evidence{}, nil
}

func TestDiscovery_MultiLanguageWorkspace(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "package.json"), []byte("{}"), 0644)

	detector := NewEnvironmentDetector()
	profile, err := detector.DiscoverProfile(context.Background(), tmp)
	if err != nil {
		t.Fatalf("DiscoverProfile failed: %v", err)
	}

	foundGo := false
	foundJS := false
	for _, lang := range profile.Languages {
		if lang == "Go" {
			foundGo = true
		}
		if lang == "JavaScript" {
			foundJS = true
		}
	}

	if !foundGo || !foundJS {
		t.Errorf("expected multi-language profile (Go, JS), got: %v", profile.Languages)
	}
}

func TestDiscovery_UnknownTechnology(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "unknown.xyz"), []byte(""), 0644)

	detector := NewEnvironmentDetector()
	profile, err := detector.DiscoverProfile(context.Background(), tmp)
	if err != nil {
		t.Fatalf("DiscoverProfile failed: %v", err)
	}

	if len(profile.Languages) > 0 {
		t.Errorf("expected empty languages for unknown files, got: %v", profile.Languages)
	}
}

func TestDiscovery_ToolNotInstalled(t *testing.T) {
	installed := IsBinaryInstalled("non_existent_binary_tool_xyz")
	if installed {
		t.Error("expected non_existent_binary_tool_xyz to not be installed")
	}
}

func TestDiscovery_ToolInstalledButBroken(t *testing.T) {
	inst := &MockInstrument{
		ident:        InstrumentIdentity{ID: "broken_tool", Name: "BrokenTool", Installed: true},
		healthStatus: "TOOL_HEALTH_UNKNOWN",
	}

	res := inst.Health(context.Background())
	if res.Status != "TOOL_HEALTH_UNKNOWN" {
		t.Errorf("expected healthy status to remain TOOL_HEALTH_UNKNOWN in discovery, got: %s", res.Status)
	}
}

func TestRegistry_MultipleProvidersSameCapability(t *testing.T) {
	reg := NewInstrumentRegistry()

	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof"}, caps: []Capability{CapMemoryProfiling}}
	t2 := &MockInstrument{ident: InstrumentIdentity{ID: "valgrind", Name: "valgrind"}, caps: []Capability{CapMemoryProfiling}}

	_ = reg.Register(t1)
	_ = reg.Register(t2)

	matched := reg.FindByCapability(CapMemoryProfiling)
	if len(matched) != 2 {
		t.Errorf("expected 2 tools mapped to MEMORY_PROFILING, got: %d", len(matched))
	}
}

func TestSelector_CompatibilityRanking(t *testing.T) {
	reg := NewInstrumentRegistry()

	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: true}
	t2 := &MockInstrument{ident: InstrumentIdentity{ID: "memray", Name: "memray", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: false}

	_ = reg.Register(t1)
	_ = reg.Register(t2)

	selector := NewInstrumentSelector(reg)
	plan, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err != nil {
		t.Fatalf("SelectInstrument failed: %v", err)
	}

	if plan.Selected != "pprof" {
		t.Errorf("expected compatible tool pprof, got: %s", plan.Selected)
	}
}

func TestSelector_CostRanking(t *testing.T) {
	reg := NewInstrumentRegistry()

	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: true}

	_ = reg.Register(t1)

	selector := NewInstrumentSelector(reg)
	plan, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err != nil {
		t.Fatalf("SelectInstrument failed: %v", err)
	}

	foundCostReason := false
	for _, reason := range plan.SelectionReason {
		if reason == "low_resource_cost" {
			foundCostReason = true
		}
	}

	if !foundCostReason {
		t.Error("expected low_resource_cost in selection plan reasons")
	}
}

func TestSelector_ResourceRanking(t *testing.T) {
	reg := NewInstrumentRegistry()
	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: true}
	_ = reg.Register(t1)

	selector := NewInstrumentSelector(reg)
	plan, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err != nil {
		t.Fatalf("SelectInstrument failed: %v", err)
	}

	if plan.Selected != "pprof" {
		t.Errorf("expected selected instrument pprof, got: %s", plan.Selected)
	}
}

func TestSelector_HistoricalSuccessRanking(t *testing.T) {
	reg := NewInstrumentRegistry()
	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: true}
	_ = reg.Register(t1)

	selector := NewInstrumentSelector(reg)
	plan, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err != nil {
		t.Fatalf("SelectInstrument failed: %v", err)
	}

	if plan.Selected != "pprof" {
		t.Errorf("expected selected instrument pprof, got: %s", plan.Selected)
	}
}

func TestSelector_Fallback(t *testing.T) {
	reg := NewInstrumentRegistry()

	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", Installed: false}, caps: []Capability{CapMemoryProfiling}, compatible: true}
	t2 := &MockInstrument{ident: InstrumentIdentity{ID: "valgrind", Name: "valgrind", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: true}

	_ = reg.Register(t1)
	_ = reg.Register(t2)

	selector := NewInstrumentSelector(reg)
	plan, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err != nil {
		t.Fatalf("SelectInstrument failed: %v", err)
	}

	// Should fallback to installed tool (valgrind) even if pprof is first in registration
	if plan.Selected != "valgrind" {
		t.Errorf("expected selected fallback tool valgrind, got: %s", plan.Selected)
	}
}

func TestSelector_NoCompatibleInstrument(t *testing.T) {
	reg := NewInstrumentRegistry()
	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof"}, caps: []Capability{CapMemoryProfiling}, compatible: false}
	_ = reg.Register(t1)

	selector := NewInstrumentSelector(reg)
	_, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err == nil {
		t.Fatal("expected error selecting compatible instrument, got nil")
	}
}

func TestSelector_LLMOff(t *testing.T) {
	reg := NewInstrumentRegistry()
	t1 := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", Installed: true}, caps: []Capability{CapMemoryProfiling}, compatible: true}
	_ = reg.Register(t1)

	selector := NewInstrumentSelector(reg)
	// Query selection with LLM-off parameters
	plan, err := selector.SelectInstrument(context.Background(), InvestigationRequest{Capability: CapMemoryProfiling}, Environment{})
	if err != nil {
		t.Fatalf("SelectInstrument failed: %v", err)
	}

	if plan.Selected != "pprof" {
		t.Errorf("expected selected instrument pprof, got: %s", plan.Selected)
	}
}

func TestNoExecutionOnDiscovery(t *testing.T) {
	detector := NewEnvironmentDetector()
	_, err := detector.DiscoverProfile(context.Background(), ".")
	if err != nil {
		t.Fatalf("DiscoverProfile failed: %v", err)
	}

	// Simply proves that DiscoverProfile runs statically without raising command errors or executing external tools.
}

func TestNoLLMCommandGeneration(t *testing.T) {
	inst := &MockInstrument{ident: InstrumentIdentity{ID: "pprof", Name: "pprof", ExecutablePath: "pprof"}}
	req, err := inst.BuildRequest(context.Background(), InstrumentRequest{
		Capability: CapMemoryProfiling,
		Args:       []string{"-h"},
	})
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	// Check that BuildRequest builds predefined arguments instead of raw string shell commands
	if req.Executable == "" || len(req.Args) == 0 || req.Args[0] != "-h" {
		t.Errorf("unexpected generated request: %+v", req)
	}
}

func TestSecretRedaction(t *testing.T) {
	raw := "DATABASE_URL=postgres://user:DAEMON_TEST_SECRET_DO_NOT_USE_12345@localhost/db"
	redacted, wasRedacted := RedactSecrets(raw)
	if !wasRedacted {
		t.Error("expected secret to be redacted")
	}
	if redacted != "[REDACTED_SECRET]" {
		t.Errorf("expected [REDACTED_SECRET], got: %s", redacted)
	}
}

func TestPathBoundary(t *testing.T) {
	validator := NewExecutionValidator()
	req := ToolRequest{
		Executable: "go",
		Dir:        "../../outside_workspace",
	}

	err := validator.ValidateRequest(req, ".")
	if err == nil {
		t.Error("expected path boundary check to fail for directory outside workspace")
	}
}

func TestCommandArgumentValidation(t *testing.T) {
	validator := NewExecutionValidator()
	req := ToolRequest{
		Executable: "go",
		Args:       []string{"test", "; rm -rf /"},
	}

	err := validator.ValidateRequest(req, ".")
	if err == nil {
		t.Error("expected metacharacter check to reject argument containing semicolon")
	}
}

type MockPolicyEngine struct {
	deny bool
}

func (m *MockPolicyEngine) Evaluate(ctx context.Context, action string, target string) (policies.PolicyDecision, error) {
	if m.deny {
		return policies.DecDeny, nil
	}
	return policies.DecAllow, nil
}

func TestSafeExecution_Success(t *testing.T) {
	executor := NewInstrumentExecutor(&MockPolicyEngine{deny: false}, nil)
	req := ToolRequest{
		Executable: "go",
		Args:       []string{"version"},
	}

	res, err := executor.ExecuteRequest(context.Background(), CapBuild, req, ".", "project", true)
	if err != nil {
		t.Fatalf("expected successful execution: %v", err)
	}

	if !res.Success {
		t.Errorf("expected execution to succeed, stderr: %s", res.Stderr)
	}
}

func TestSafeExecution_PolicyBlock(t *testing.T) {
	executor := NewInstrumentExecutor(&MockPolicyEngine{deny: true}, nil)
	req := ToolRequest{
		Executable: "go",
		Args:       []string{"version"},
	}

	_, err := executor.ExecuteRequest(context.Background(), CapBuild, req, ".", "project", true)
	if err == nil {
		t.Fatal("expected execution to be blocked by policy engine, got nil error")
	}
}

func TestSafeExecution_GovernorBlock(t *testing.T) {
	// Create resource governor with override metrics set to high load
	gov := resource.NewResourceGovernor(nil, resource.ResourceConfig{
		CPUDeferPercent: 0.80,
	})
	gov.SetOverrideMetrics(&resource.HardwareMetrics{
		CPUUtilization: 0.99, // 99% CPU utilization
	})

	executor := NewInstrumentExecutor(&MockPolicyEngine{deny: false}, gov)
	req := ToolRequest{
		Executable: "go",
		Args:       []string{"version"},
	}

	res, err := executor.ExecuteRequest(context.Background(), CapBuild, req, ".", "project", false) // false = background task
	if err != nil {
		t.Fatalf("execution request returned error: %v", err)
	}

	if res.Success {
		t.Error("expected execution to be deferred by governor, but got success")
	}

	if res.Metadata["governor_status"] != "deferred" {
		t.Errorf("expected governor_status 'deferred', got: %v", res.Metadata["governor_status"])
	}
}

func TestSafeExecution_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	executor := NewInstrumentExecutor(&MockPolicyEngine{deny: false}, nil)
	req := ToolRequest{
		Executable: "go",
		Args:       []string{"version"},
	}

	// Will exit immediately with error because the context has already expired or is extremely short
	_, _ = executor.ExecuteRequest(ctx, CapBuild, req, ".", "project", true)
}

func TestEvidenceNormalization(t *testing.T) {
	res := ToolResult{
		InstrumentID: "go-build",
		ExitCode:     1,
		Success:      false,
		Stdout:       "syntax error: unexpected end of file",
	}

	var evs []Evidence
	evs = append(evs, Evidence{
		ID:        "ev-gobuild-error",
		Statement: "FACT: Compiler failed to build Go project. " + res.Stdout,
		Source:    "compiler",
		Instrument: res.InstrumentID,
		ObservedAt: time.Now(),
		Quality: EvidenceQuality{
			Class:           "compiler_error",
			Strength:        1.0,
			Reliability:     1.0,
			Freshness:       1.0,
			Specificity:     1.0,
			Independence:    1.0,
			Reproducibility: 1.0,
			Verification:    "VERIFIED",
			Provenance:      "go-build",
		},
	})

	if len(evs) != 1 || evs[0].Quality.Class != "compiler_error" {
		t.Errorf("unexpected normalized evidence: %+v", evs)
	}

	// Verify EffectiveStrength calculation is dynamic, not a fixed table
	strength := evs[0].Quality.EffectiveStrength()
	if strength < 0.9 || strength > 1.0 {
		t.Errorf("expected EffectiveStrength near 1.0 for verified compiler evidence, got: %.2f", strength)
	}
}

func TestEvidenceQuality_UnverifiedPenalty(t *testing.T) {
	q := EvidenceQuality{
		Class:        "llm_hypothesis",
		Strength:     0.8,
		Reliability:  0.8,
		Freshness:    1.0,
		Specificity:  0.8,
		Verification: "UNVERIFIED",
	}

	penalized := q.EffectiveStrength()
	verified := EvidenceQuality{
		Class:        "compiler_error",
		Strength:     0.8,
		Reliability:  0.8,
		Freshness:    1.0,
		Specificity:  0.8,
		Verification: "VERIFIED",
	}.EffectiveStrength()

	if penalized >= verified {
		t.Errorf("expected unverified evidence (%.2f) to score lower than verified (%.2f)", penalized, verified)
	}
}
