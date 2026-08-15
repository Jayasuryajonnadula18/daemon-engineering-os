# DAEMON HYBRID ENGINEERING INTELLIGENCE AUDIT

This document presents a comprehensive repository-wide architectural and security audit of Daemon's intelligence and reasoning structures, mapping the division of labor between LLM reasoning and deterministic execution.

---

## 1. Duplicated Reasoning Logic

We identified three independent reasoning controllers running different loops:

1. **`core/agent/agent.go` (`AgentRuntime.RunLoop`)**:
   * Runs an Observe → Understand → Hypothesize → Plan → Execute TUI-oriented agent loop.
   * Simulates tool proposals internally based on hardcoded step numbers (`ar.iterations == 1`, etc.).
2. **`core/debug/debugger.go` (`Debugger.RunInvestigation`)**:
   * Runs an initial native triage followed by a deterministic/experiment loop.
   * Manually drives hypothesis generation and verification.
3. **`core/reasoning/reasoner.go` (`EngineeringReasoner.Reason`)**:
   * Processes queries statically and outputs a `ReasoningResult` structure containing facts and inferences.

*Impact*: Logic for hypothesis generation, confidence calculation, and verification is split across three packages, making it impossible to govern AI activity under a single interface.

---

## 2. Duplicated Evidence and Observation Models

1. **`core/agent/observation.go`**:
   * Defines `Observation` (ID, Type, Statement, Confidence, Scope) with observation types: `ObsFile`, `ObsCode`, `ObsAST`, `ObsGit`, `ObsBuild`, `ObsTest`, `ObsProcess`, `ObsContainer`, `ObsKubernetes`, `ObsNetwork`, `ObsDatabase`, `ObsLog`, `ObsMetric`, `ObsEvent`, `ObsHuman`, `ObsModel`.
2. **`core/debug/evidence.go`**:
   * Defines `Evidence` (ID, Type, Statement, Confidence, Scope) with types: `EvidenceGit`, `EvidenceCode`, `EvidenceAST`, `EvidenceBuild`, `EvidenceTest`, `EvidenceLog`, `EvidenceMetric`, `EvidenceHistory`, `EvidenceDatamine`, `EvidenceModel`.
3. **`core/instruments/evidence.go`**:
   * Defines `Evidence` (ID, Statement, Source, Instrument, ObservedAt, Quality) where `Quality` is a rich `EvidenceQuality` struct.

*Impact*: These structures represent the same fundamental concept—an engineering observation from an instrument—but map to completely different Go structs, preventing adapters from normalizing findings into a single shared primitive.

---

## 3. Duplicated Tool and Capability Registries

1. **`core/agent/tools/tools.go`**:
   * Defines a `Tool` interface and a `globalRegistry` (`ToolRegistry`) targeting agent workflows.
2. **`core/instruments/registry.go`**:
   * Defines `EngineeringInstrument` interface and `InstrumentRegistry` mapping tool capabilities.
3. **`sdk/plugin/plugin.go`**:
   * Defines a `Plugin` interface and `CapabilityRegistry` tracking third-party extensions.

*Impact*: The debugger and the agent use separate, incompatible registry systems to resolve available tools and capabilities.

---

## 4. Duplicated Execution Gateways

1. **`core/agent/agent.go`**:
   * Evaluates safety policies and resource governor checks inline inside the agent's execution block.
2. **`core/instruments/execution.go` (`InstrumentExecutor`)**:
   * Evaluates the same safety policies and resource governor checks inside a dedicated sub-process runner.

*Impact*: Security and resource limits are bypassed or duplicated, creating unsafe authority boundaries where agent tools do not inherit the same sandbox safety as debugger tools.

---

## 5. Unsafe Authority Boundaries

1. **LLM Command Execution (Risk of Instruction injection)**:
   * The reasoning client (`core/reasoning/llm_client.go`) returns structured suggestions, but there is no strict validation step ensuring the model's output cannot be directly passed to the shell.
   * If the LLM proposes an arbitrary command string, the agent execution layer accepts and executes it, presenting a major security risk.
2. **No Falsification checks**:
   * Currently, once the agent or debugger finds a hypothesis with >0.8 confidence, it declares it the root cause. It does not challenge itself or search for contradicting evidence before making a final determination.

---

## 6. Go-Specific Assumptions and Ecosystem Bias

1. **Hardcoded build/test commands**:
   * `debugger.go` directly executes `go build ./...` and `go test ./...` upon detecting `go.mod`.
   * Unclosed body checks are restricted to parsing Go AST files (`checkBodyCloseLeak`).
2. **Ecosystem adapters**:
   * Registry contains only `adapters/build/go` and `adapters/testing/go`, neglecting other technology profiles.

---

## 7. Merge and Restructure Strategy

We will apply the following package merges and abstractions:

### A. Shared Primitives (Merge)
* **Unified Evidence**: Merge `core/agent/observation.go` and `core/debug/evidence.go` into `core/instruments/evidence.go` (the rich `Evidence` and `EvidenceQuality` primitives).
* **Unified Registry**: Consolidate `core/agent/tools/` registry into `core/instruments/registry.go`. Make `AgentRuntime` resolve instruments directly from the dynamic registry.
* **Unified Execution**: Run all subprocesses exclusively via the `InstrumentExecutor` in `core/instruments/execution.go` to enforce uniform safety checks.

### B. Shared Reasoning Primitives (New Interface)
* Create `ReasoningEngine` in `core/reasoning/reasoner.go` exposing:
  * `GenerateHypotheses(ctx, query)`
  * `ProposeExperiments(ctx, hypotheses, evidence)`
  * `ChallengeHypothesis(ctx, leading, alternatives)`
  * `ExplainConclusion(ctx, conclusion)`
* Implement three variants:
  * `DeterministicReasoningEngine` (rule-based)
  * `LLMReasoningEngine` (Ollama/remote LLM backend)
  * `HybridReasoningEngine` (coordinating both)

### C. Safety Gateways
* Enforce that the LLM/reasoner can only output structured `ExperimentProposal` and `HypothesisChallenge` objects. It is completely isolated from command executing and file writing APIs.
