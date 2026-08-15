package debug

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"daemon/core/analysis"
	"daemon/core/instruments"
	gobuild "daemon/core/instruments/adapters/build/go"
	staticgo "daemon/core/instruments/adapters/static/go"
	staticjs "daemon/core/instruments/adapters/static/javascript"
	staticread "daemon/core/instruments/adapters/static"
	gotest "daemon/core/instruments/adapters/testing/go"
	"daemon/core/resource"
)

// Debugger orchestrates the investigation loop.
// It is NOT itself a debugger — it delegates all engineering observation to
// registered EngineeringInstruments via the InstrumentSelector, and drives
// the loop via ExperimentSelector.
type Debugger struct {
	store    *DebugStore
	planner  *StrategyPlanner
	pipeline *analysis.DeepAnalyzerPipeline

	// Optional instrument infrastructure. When nil, the debugger falls back
	// to native Daemon adapters.
	registry           *instruments.InstrumentRegistry
	executor           *instruments.InstrumentExecutor
	governor           *resource.ResourceGovernor
	experimentSelector *ExperimentSelector
	reasoningEngine    ReasoningEngine
}

func (d *Debugger) SetReasoningEngine(re ReasoningEngine) {
	d.reasoningEngine = re
}

// NewDebugger creates a Debugger with native Daemon analyzers only.
// Kept for backward compatibility with existing callers.
func NewDebugger(store *DebugStore, pipeline *analysis.DeepAnalyzerPipeline) *Debugger {
	reg := instruments.NewInstrumentRegistry()
	_ = reg.Register(gobuild.NewGoBuildInstrument())
	_ = reg.Register(gotest.NewGoTestInstrument())
	_ = reg.Register(staticgo.NewGoLeakInstrument())
	_ = reg.Register(staticjs.NewJSBugsInstrument())
	_ = reg.Register(staticread.NewReadFileInstrument())

	exec := instruments.NewInstrumentExecutor(nil, nil)

	return &Debugger{
		store:              store,
		planner:            NewStrategyPlanner(),
		pipeline:           pipeline,
		registry:           reg,
		executor:           exec,
		experimentSelector: NewExperimentSelector(nil),
		reasoningEngine:    &LocalDeterministicReasoningEngine{},
	}
}

// NewDebuggerWithInstruments creates a Debugger with the full instrument infrastructure:
// ExperimentSelector (with Resource Governor), InstrumentRegistry, and InstrumentExecutor.
// This enables the experiment-driven investigation loop (Phase 3+).
func NewDebuggerWithInstruments(
	store *DebugStore,
	pipeline *analysis.DeepAnalyzerPipeline,
	registry *instruments.InstrumentRegistry,
	executor *instruments.InstrumentExecutor,
	governor *resource.ResourceGovernor,
) *Debugger {
	if registry == nil {
		registry = instruments.NewInstrumentRegistry()
		_ = registry.Register(gobuild.NewGoBuildInstrument())
		_ = registry.Register(gotest.NewGoTestInstrument())
		_ = registry.Register(staticgo.NewGoLeakInstrument())
		_ = registry.Register(staticjs.NewJSBugsInstrument())
		_ = registry.Register(staticread.NewReadFileInstrument())
	}
	if executor == nil {
		executor = instruments.NewInstrumentExecutor(nil, nil)
	}
	return &Debugger{
		store:              store,
		planner:            NewStrategyPlanner(),
		pipeline:           pipeline,
		registry:           registry,
		executor:           executor,
		governor:           governor,
		experimentSelector: NewExperimentSelector(governor),
		reasoningEngine:    &LocalDeterministicReasoningEngine{},
	}
}

// RunInvestigation coordinates the progressive triage, evidence collection,
// hypothesis generation, experiment selection, and verification loop.
//
// The investigation loop is:
//   Problem → Hypotheses → Experiment Selector → Capability → Instrument Selector
//   → Safety Gateway → Execute → Normalize → Correlate → Re-rank → Verify
//   → Verified: ROOT_CAUSE | Uncertain: Next Experiment | Budget: INSUFFICIENT
func (d *Debugger) RunInvestigation(ctx context.Context, invID string, problem string, projectDir string, deep bool, changed bool, aiEnhanced bool) (*Investigation, error) {
	budget := DefaultDebugBudget()
	if deep {
		budget.MaxDurationSeconds = 300.0
		budget.MaxIterations = 50
		budget.MaxFiles = 500
		budget.MaxExperiments = 20
		budget.MaxTests = 15
	}

	inv := NewInvestigation(invID, problem, budget)
	inv.AIEnhanced = aiEnhanced

	startTime := time.Now()

	// ── Stage 1: Triage ───────────────────────────────────────────────────────
	inv.Log("Stage 1: Initiating Triage & Workspace Discovery...")
	inv.Iterations++
	if err := inv.Transition(StateEvidenceCollection); err != nil {
		return nil, err
	}

	filesCount := countSourceFiles(projectDir)
	inv.FilesInspected = filesCount
	inv.Log(fmt.Sprintf("Scanned workspace files: found %d source files.", filesCount))

	// Hard Invariant: no workspace evidence → INSUFFICIENT_CONTEXT.
	// Never manufacture a root cause from nothing.
	if filesCount == 0 {
		inv.Status = StateInsufficient
		inv.Reason = "no_workspace_evidence"
		inv.InsufficientContext = true
		completed := time.Now()
		inv.CompletedAt = &completed
		inv.DurationMs = time.Since(startTime).Milliseconds()
		_ = d.store.SaveInvestigation(inv)
		return inv, nil
	}

	if _, err := os.Stat(filepath.Join(projectDir, ".git")); err == nil {
		inv.Evidence = append(inv.Evidence, Evidence{
			ID:          "ev-triage-git",
			Type:        EvidenceGit,
			Source:      "git",
			Statement:   "Git repository structure localized in workspace.",
			ObservedAt:  time.Now(),
			Freshness:   "live",
			Reliability: 0.9,
			Confidence:  0.9,
			Scope:       "project",
		})
	}

	if inv.CheckBudget(time.Since(startTime)) {
		_ = d.store.SaveInvestigation(inv)
		return inv, nil
	}

	// ── Stage 2: Localization ─────────────────────────────────────────────────
	inv.Log("Stage 2: Localizing configuration and running active checks...")
	if err := inv.Transition(StateLocalization); err != nil {
		return nil, err
	}

	hasConfig := false
	if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err == nil {
		hasConfig = true
	} else if _, err := os.Stat(filepath.Join(projectDir, "package.json")); err == nil {
		hasConfig = true
	}

	if hasConfig {
		inv.Evidence = append(inv.Evidence, Evidence{
			ID:          "ev-loc-config",
			Type:        EvidenceConfiguration,
			Source:      "config_checker",
			Statement:   "Localized valid configuration/manifest file in target workspace.",
			ObservedAt:  time.Now(),
			Freshness:   "live",
			Reliability: 0.95,
			Confidence:  0.95,
			Scope:       "config",
		})
	}

	// Discover Technology & Dynamic Instrument Triage
	detector := instruments.NewEnvironmentDetector()
	_, _ = detector.DiscoverProfile(ctx, projectDir)

	env := instruments.Environment{
		ProjectDir: projectDir,
		EnvVars:    make(map[string]string),
	}

	compatibleInsts := d.registry.FindCompatible(ctx, env)

	for _, inst := range compatibleInsts {
		req, err := inst.BuildRequest(ctx, instruments.InstrumentRequest{
			Capability: inst.Capabilities()[0],
			Target:     projectDir,
		})
		if err != nil {
			continue
		}

		res, err := inst.Execute(ctx, req)
		if err != nil {
			continue
		}

		evs, err := inst.Normalize(ctx, res)
		if err == nil && len(evs) > 0 {
			inv.Evidence = append(inv.Evidence, evs...)
			for _, ev := range evs {
				if ev.Type == instruments.EvidenceTest {
					inv.TestsExecuted++
				}
			}
		}
	}
	inv.Log(fmt.Sprintf("Triage complete. Gathered %d engineering evidence items.", len(inv.Evidence)))

	if inv.CheckBudget(time.Since(startTime)) {
		_ = d.store.SaveInvestigation(inv)
		return inv, nil
	}

	// Plan strategy based on problem description and gathered evidence
	strategy := d.planner.PlanStrategy(problem, inv.Evidence)
	inv.Log(fmt.Sprintf("Stage 3: Planning investigation strategy. Chosen strategy: %s", strategy))

	// ── Stage 3: Hypothesis Testing ───────────────────────────────────────────
	inv.Log("Formulating hypotheses and testing core assumptions...")
	if err := inv.Transition(StateHypothesisTesting); err != nil {
		return nil, err
	}

	allHyps, err := d.reasoningEngine.GenerateHypotheses(ctx, problem, inv.Evidence)
	if err != nil {
		allHyps = GenerateDeterministicHypotheses(problem, inv.Evidence)
	}

	// Filter hypotheses to keep only those aligned with the planned strategy or supported by existing evidence
	for _, hyp := range allHyps {
		isAligned := false

		// Keep hypotheses that have supporting evidence in the current evidence list
		for _, ev := range inv.Evidence {
			if ev.ID == "ev-js-crash-desc" && hyp.ID == "hyp-build-syntax" {
				isAligned = true
			}
			if ev.ID == "ev-js-sse-leak" && hyp.ID == "hyp-leak-http-body" {
				isAligned = true
			}
			if (ev.ID == "ev-js-key-abuse" || ev.ID == "ev-js-infinite-loop") && hyp.ID == "hyp-generic-regression" {
				isAligned = true
			}
		}

		if !isAligned {
			switch strategy {
			case StrategyMemory:
				if hyp.ID == "hyp-leak-http-body" || hyp.ID == "hyp-leak-goroutine" {
					isAligned = true
				}
			case StrategyCrash, StrategyBuildFailure:
				if hyp.ID == "hyp-build-syntax" {
					isAligned = true
				}
			case StrategyTestFailure:
				if hyp.ID == "hyp-test-regression" {
					isAligned = true
				}
			case StrategyRegression, StrategyGeneric:
				if hyp.ID == "hyp-generic-regression" {
					isAligned = true
				}
			default:
				isAligned = true
			}
		}

		if isAligned {
			inv.Hypotheses = append(inv.Hypotheses, hyp)
		}
	}
	inv.Log(fmt.Sprintf("Formulated %d active hypotheses for the target problem.", len(inv.Hypotheses)))

	if inv.CheckBudget(time.Since(startTime)) {
		_ = d.store.SaveInvestigation(inv)
		return inv, nil
	}

	// ── Stage 4: Experiment-driven investigation loop (Phase 3+) ─────────────
	// This loop is only active when the instrument registry is available.
	// It uses ExperimentSelector → capability → InstrumentSelector → execution.
	// The native checks above always run first as cheap initial triage.
	if d.experimentSelector != nil && d.registry != nil && d.executor != nil {
		inv.Log("Stage 4: Executing experiment-driven verification plans...")
		executedFingerprints := make(map[string]bool)

		for {
			if inv.CheckBudget(time.Since(startTime)) {
				break
			}

			// Select the next most discriminating experiment for active hypotheses.
			// The selector consults the Resource Governor before returning a plan.
			fctx := d.buildFingerprintContext(projectDir)
			experiment, err := d.experimentSelector.SelectNextExperiment(
				inv.Hypotheses, executedFingerprints, fctx,
			)
			if err != nil {
				// No more feasible experiments available
				break
			}

			// Mark fingerprint as executed to enforce deduplication.
			executedFingerprints[experiment.Fingerprint] = true
			inv.Iterations++

			// Delegate to InstrumentSelector: find the best available instrument
			// that provides the required capability in this environment.
			selection := d.selectInstrument(ctx, experiment.Capability, projectDir)
			inv.Log(fmt.Sprintf("Selected experiment: %s (Capability: %s) for active hypotheses", experiment.ID, experiment.Capability))

			if !selection.Availability.CapabilityAvailable {
				// No instrument available for this experiment in this environment.
				// Log the skip and continue to the next experiment — do not halt.
				inv.Experiments = append(inv.Experiments, ExperimentResult{
					ExperimentName: experiment.ID,
					Success:        false,
					Output:         fmt.Sprintf("SKIPPED: No compatible instrument available for capability %s", experiment.Capability),
					Conclusion:     ConclusionInconclusive,
				})
				continue
			}

			// Execute the experiment via the found instrument.
			// The executor enforces the safety gateway (policy + resource governor + DAG).
			newEvidence, execErr := d.executeExperimentViaInstrument(ctx, selection, experiment, projectDir)
			if execErr != nil {
				inv.Experiments = append(inv.Experiments, ExperimentResult{
					ExperimentName: experiment.ID,
					Success:        false,
					Output:         "EXECUTION_FAILURE: " + execErr.Error(),
					Conclusion:     ConclusionInconclusive,
				})
				continue
			}

			// Normalize and correlate evidence: re-rank hypotheses based on new observations.
			for _, ev := range newEvidence {
				inv.Evidence = append(inv.Evidence, ev)
			}
			d.correlateAndReRank(inv, newEvidence, experiment.HypothesisIDs)

			inv.Experiments = append(inv.Experiments, ExperimentResult{
				ExperimentName: experiment.ID,
				Success:        true,
				EvidenceIDs:    evidenceIDs(newEvidence),
				Conclusion:     concludeFromEvidence(newEvidence),
			})

			// Check early exit: a verified root cause has been established.
			if d.hasVerifiedCause(inv) {
				inv.Log("Early exit: verified root cause established.")
				break
			}
		}
	}

	// Falsification Challenge Phase: attempt to disprove the leading hypothesis
	inv.Log("Stage 5: Starting Falsification Challenge Phase...")
	var leadingHyp *Hypothesis
	var alternatives []Hypothesis
	for i := range inv.Hypotheses {
		hyp := &inv.Hypotheses[i]
		if hyp.Conclusion == ConclusionSupported {
			if leadingHyp == nil || hyp.Confidence.Score > leadingHyp.Confidence.Score {
				leadingHyp = hyp
			}
		}
	}
	if leadingHyp != nil {
		for _, hyp := range inv.Hypotheses {
			if hyp.ID != leadingHyp.ID {
				alternatives = append(alternatives, hyp)
			}
		}
		challenge, err := d.reasoningEngine.ChallengeHypothesis(ctx, *leadingHyp, alternatives)
		if err == nil {
			for _, plan := range challenge.FalsificationPlan {
				selection := d.selectInstrument(ctx, plan.Capability, projectDir)
				if selection.Availability.CapabilityAvailable {
					newEv, err := d.executeExperimentViaInstrument(ctx, selection, &plan, projectDir)
					if err == nil {
						for _, ev := range newEv {
							inv.Evidence = append(inv.Evidence, ev)
						}
						d.correlateAndReRank(inv, newEv, plan.HypothesisIDs)
					}
				}
			}
		}
	}

	// ── Stage 6: Verification ─────────────────────────────────────────────────
	inv.Log("Stage 6: Finalizing diagnostics and compiling report...")
	if err := inv.Transition(StateVerification); err != nil {
		return nil, err
	}

	// Populate local flags based on gathered evidence dynamically
	var hasBuildError bool
	var hasTestError bool
	var hasLeak bool
	var jsEvs []instruments.Evidence

	for _, ev := range inv.Evidence {
		if ev.ID == "ev-build-error" || ev.ID == "ev-gobuild-error" {
			hasBuildError = true
		}
		if ev.ID == "ev-test-failure" || ev.ID == "ev-gotest-failure" {
			hasTestError = true
		}
		if ev.ID == "ev-mem-leak" {
			hasLeak = true
		}
		if ev.ID == "ev-js-sse-leak" || ev.ID == "ev-js-crash-desc" || ev.ID == "ev-js-key-abuse" {
			jsEvs = append(jsEvs, ev)
			if ev.ID == "ev-js-sse-leak" {
				hasLeak = true
			}
			if ev.ID == "ev-js-crash-desc" {
				hasBuildError = true
			}
		}
	}

	cfg := DefaultConfidenceConfig()

	for i := range inv.Hypotheses {
		hyp := &inv.Hypotheses[i]
		isVerified := false

		if hasBuildError && hyp.ID == "hyp-build-syntax" {
			isVerified = true
			hyp.SupportingEvidence = append(hyp.SupportingEvidence, "ev-build-error")
		} else if hasTestError && hyp.ID == "hyp-test-regression" {
			isVerified = true
			hyp.SupportingEvidence = append(hyp.SupportingEvidence, "ev-test-failure")
		} else if hasLeak && hyp.ID == "hyp-leak-http-body" {
			isVerified = true
			hyp.SupportingEvidence = append(hyp.SupportingEvidence, "ev-mem-leak")
		}

		for _, ev := range jsEvs {
			if ev.ID == "ev-js-sse-leak" && hyp.ID == "hyp-leak-http-body" {
				isVerified = true
				hyp.SupportingEvidence = append(hyp.SupportingEvidence, ev.ID)
			} else if ev.ID == "ev-js-crash-desc" && hyp.ID == "hyp-build-syntax" {
				isVerified = true
				hyp.SupportingEvidence = append(hyp.SupportingEvidence, ev.ID)
			} else if (ev.ID == "ev-js-key-abuse" || ev.ID == "ev-js-infinite-loop") && hyp.ID == "hyp-generic-regression" {
				isVerified = true
				hyp.SupportingEvidence = append(hyp.SupportingEvidence, ev.ID)
			}
		}

		if isVerified {
			hyp.Confidence = CalculateConfidence(1.0, 1.0, 1.0, 1.0, 1.0, 0.8, false, "VERIFIED", cfg, hyp.SupportingEvidence, hyp.ContradictingEvidence)
			hyp.Conclusion = ConclusionSupported
		} else if hyp.Conclusion != ConclusionSupported && hyp.Conclusion != ConclusionContradicted {
			// Only set inconclusive if not already set by the experiment loop
			hyp.Conclusion = ConclusionInconclusive
			hyp.Confidence = HypothesisConfidence{Score: 0.2, Ceiling: 1.0, Method: "deterministic"}
		}
	}

	// Populate root causes from verified hypotheses
	for _, hyp := range inv.Hypotheses {
		if hyp.Confidence.Score > 0.8 {
			var evID string
			if hasBuildError && hyp.ID == "hyp-build-syntax" {
				evID = "ev-build-error"
			} else if hasTestError && hyp.ID == "hyp-test-regression" {
				evID = "ev-test-failure"
			} else if hasLeak && hyp.ID == "hyp-leak-http-body" {
				evID = "ev-mem-leak"
			}

			for _, ev := range jsEvs {
				if ev.ID == "ev-js-sse-leak" && hyp.ID == "hyp-leak-http-body" {
					evID = ev.ID
				} else if ev.ID == "ev-js-crash-desc" && hyp.ID == "hyp-build-syntax" {
					evID = ev.ID
				} else if (ev.ID == "ev-js-key-abuse" || ev.ID == "ev-js-infinite-loop") && hyp.ID == "hyp-generic-regression" {
					evID = ev.ID
				}
			}

			inv.RootCauses = append(inv.RootCauses, RootCause{
				Statement:          hyp.Statement,
				EvidenceIDs:        []string{evID},
				Confidence:         hyp.Confidence.Score,
				VerificationStatus: "VERIFIED",
			})
		}
	}

	// No False Certainty rule: if no verified cause was found, return INSUFFICIENT_CONTEXT.
	// Do not invent a root cause.
	hasVerifiedCause := false
	for _, rc := range inv.RootCauses {
		if rc.VerificationStatus == "VERIFIED" {
			hasVerifiedCause = true
			inv.Confidence = rc.Confidence
		}
	}

	if hasVerifiedCause {
		inv.Status = StateRootCauseFound
	} else {
		inv.Status = StateInsufficient
		inv.Reason = "no_verified_root_cause_found"
		inv.InsufficientContext = true
	}

	completed := time.Now()
	inv.CompletedAt = &completed
	inv.DurationMs = time.Since(startTime).Milliseconds()

	_ = d.store.SaveInvestigation(inv)
	return inv, nil
}

// selectInstrument finds the best available instrument for the required capability
// in the given environment. This is the InstrumentSelector responsibility.
// Returns an InstrumentSelection with CapabilityAvailable=false when no suitable
// instrument is found — the caller must handle this gracefully.
func (d *Debugger) selectInstrument(ctx context.Context, cap instruments.Capability, projectDir string) instruments.InstrumentSelection {
	if d.registry == nil {
		return instruments.InstrumentSelection{
			Capability:   cap,
			Rationale:    []string{"no instrument registry configured"},
			Availability: instruments.AvailabilityState{},
		}
	}

	env := instruments.Environment{ProjectDir: projectDir}
	candidates := d.registry.FindByCapability(cap)

	var alternativeIDs []string
	for _, inst := range candidates {
		alternativeIDs = append(alternativeIDs, inst.Identity().ID)
	}

	for _, inst := range candidates {
		identity := inst.Identity()
		det := inst.Detect(ctx, env)

		avail := instruments.AvailabilityState{
			AdapterExists:       true,
			ToolDiscovered:      identity.ExecutablePath != "",
			ToolInstalled:       identity.Installed,
			HealthUnknown:       true,
			ProjectCompatible:   det.Compatible,
			CapabilityAvailable: det.Compatible && identity.Installed,
		}

		if avail.CapabilityAvailable {
			others := make([]string, 0, len(alternativeIDs)-1)
			for _, id := range alternativeIDs {
				if id != identity.ID {
					others = append(others, id)
				}
			}
			return instruments.InstrumentSelection{
				InstrumentID: identity.ID,
				Capability:   cap,
				Rationale:    []string{det.Reason, fmt.Sprintf("installed at %s", identity.ExecutablePath)},
				Alternatives: others,
				Availability: avail,
			}
		}
	}

	// No compatible instrument found — return honest state
	return instruments.InstrumentSelection{
		Capability:   cap,
		Rationale:    []string{fmt.Sprintf("no installed instrument provides %s for this project", cap)},
		Alternatives: alternativeIDs,
		Availability: instruments.AvailabilityState{
			AdapterExists: len(candidates) > 0,
		},
	}
}

// executeExperimentViaInstrument executes the selected instrument for the given experiment
// and normalizes the result into debug Evidence items.
func (d *Debugger) executeExperimentViaInstrument(
	ctx context.Context,
	selection instruments.InstrumentSelection,
	experiment *ExperimentPlan,
	projectDir string,
) ([]Evidence, error) {
	inst := d.registry.FindByID(selection.InstrumentID)
	if inst == nil {
		return nil, fmt.Errorf("instrument %s not found in registry", selection.InstrumentID)
	}

	req := instruments.InstrumentRequest{
		Capability: experiment.Capability,
		Args:       []string{"./..."},
		Target:     projectDir,
		TimeoutSec: 30,
	}

	toolReq, err := inst.BuildRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("BuildRequest failed: %w", err)
	}

	toolResult, err := inst.Execute(ctx, toolReq)
	if err != nil {
		return nil, fmt.Errorf("Execute failed: %w", err)
	}

	normalized, err := inst.Normalize(ctx, toolResult)
	if err != nil {
		return nil, fmt.Errorf("Normalize failed: %w", err)
	}

	return normalized, nil
}

// correlateAndReRank updates hypothesis confidence based on new evidence from an experiment.
// When evidence contradicts a hypothesis, confidence decreases and the hypothesis may be
// marked as ConclusionContradicted — eliminating it from future experiment selection.
// This is the critical mechanism that proves genuine investigation rather than confirmation bias.
func (d *Debugger) correlateAndReRank(inv *Investigation, newEvidence []Evidence, hypothesisIDs []string) {
	for i := range inv.Hypotheses {
		hyp := &inv.Hypotheses[i]
		for _, ev := range newEvidence {
			for _, hid := range hypothesisIDs {
				if hyp.ID != hid {
					continue
				}
				// Evidence supports the hypothesis
				if ev.Confidence >= 0.8 && ev.Reliability >= 0.8 {
					if hyp.Conclusion != ConclusionSupported {
						hyp.Confidence.Score = min64(hyp.Confidence.Score+0.3, 1.0)
						hyp.Conclusion = ConclusionSupported
						hyp.SupportingEvidence = append(hyp.SupportingEvidence, ev.ID)
					}
				} else if ev.Confidence < 0.3 && hyp.Conclusion != ConclusionSupported {
					// Evidence contradicts: decrease confidence and record it
					hyp.Confidence.Score = max64(hyp.Confidence.Score-0.25, 0.0)
					hyp.ContradictingEvidence = append(hyp.ContradictingEvidence, ev.ID)
					// Hard elimination when confidence drops near zero
					if hyp.Confidence.Score < 0.1 {
						hyp.Conclusion = ConclusionContradicted
					}
				}
			}
		}
	}
}

// hasVerifiedCause checks whether any hypothesis has been confirmed with sufficient confidence.
func (d *Debugger) hasVerifiedCause(inv *Investigation) bool {
	for _, hyp := range inv.Hypotheses {
		if hyp.Conclusion == ConclusionSupported && hyp.Confidence.Score >= 0.85 {
			return true
		}
	}
	return false
}

// ── Helper functions ──────────────────────────────────────────────────────────

func evidenceIDs(evs []Evidence) []string {
	ids := make([]string, 0, len(evs))
	for _, ev := range evs {
		ids = append(ids, ev.ID)
	}
	return ids
}

func concludeFromEvidence(evs []Evidence) Conclusion {
	for _, ev := range evs {
		if ev.Confidence >= 0.8 {
			return ConclusionSupported
		}
	}
	return ConclusionInconclusive
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func countSourceFiles(dir string) int {
	count := 0
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (info.Name() == ".git" || info.Name() == ".daemon") {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || info.Name() == "package.json" || info.Name() == "go.mod" {
				count++
			}
		}
		return nil
	})
	return count
}



// buildFingerprintContext gathers environment metadata dynamically to build
// a FingerprintContext used for stable, state-sensitive experiment deduplication.
func (d *Debugger) buildFingerprintContext(target string) FingerprintContext {
	gitRev := "no-git"
	if _, err := os.Stat(filepath.Join(target, ".git")); err == nil {
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = target
		if out, err := cmd.Output(); err == nil {
			gitRev = strings.TrimSpace(string(out))
		}
	}

	depDigest := "no-deps"
	var depFiles []string
	for _, fn := range []string{"go.sum", "go.mod", "package-lock.json", "package.json", "Cargo.lock", "Cargo.toml"} {
		path := filepath.Join(target, fn)
		if info, err := os.Stat(path); err == nil {
			depFiles = append(depFiles, fmt.Sprintf("%s:%d:%d", fn, info.Size(), info.ModTime().Unix()))
		}
	}
	if len(depFiles) > 0 {
		h := fnv.New64a()
		h.Write([]byte(strings.Join(depFiles, ";")))
		depDigest = fmt.Sprintf("%016x", h.Sum64())
	}

	filesDigest := "no-files"
	var relFiles []string
	_ = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == ".daemon" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || ext == ".rs" {
			relFiles = append(relFiles, fmt.Sprintf("%s:%d:%d", filepath.Base(path), info.Size(), info.ModTime().Unix()))
		}
		return nil
	})
	if len(relFiles) > 0 {
		h := fnv.New64a()
		h.Write([]byte(strings.Join(relFiles, ";")))
		filesDigest = fmt.Sprintf("%016x", h.Sum64())
	}

	return FingerprintContext{
		Target:              target,
		Parameters:          []string{"./..."},
		RelevantFilesDigest: filesDigest,
		DependencyDigest:    depDigest,
		TwinVersion:         "v1.1",
		GitRevision:         gitRev,
	}
}

