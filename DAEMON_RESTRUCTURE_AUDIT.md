# DAEMON RESTRUCTURE AUDIT

This document presents a comprehensive repository-wide architectural and structural audit of Daemon v1.2. The analysis covers existing sub-systems, duplicate components, dead code, incorrectly coupled modules, and concrete barriers preventing language-agnostic execution.

---

## A. What Actually Exists

Daemon is organized as a Go application comprising a Cobra CLI, REST server, and a modular core system divided into 35 sub-packages under `core/`.

1. **CLI Engine (`cli/`)**:
   * Cobra router (`cli/commands/root.go`) registering 42 CLI command files.
   * BubbleTea dashboard TUI (`cli/tui/tui.go`).
2. **Deterministic & AI Reasoning (`core/reasoning/`)**:
   * `llm_client.go` with Ollama local runner and `OfflineClient` structured prompt fallback.
   * Reasoner, context builder, router, and evaluator.
3. **Safety Gateway (`core/policies/` & `core/resource/`)**:
   * Policy Engine enforcing ceiling limits (production writes, force pushes, credentials).
   * Resource Governor monitoring real-time host hardware usage (CPU/RAM metrics, constraints).
4. **Engineering Reality & Graph Store (`core/storage/` & `core/graph/` & `core/twin/`)**:
   * SQLite store (`core/graph/sqlite.go`) mapping project topology to the Engineering Twin.
   * Memory store tracking historical incidents, resolutions, and recommendations.
5. **Universal Instrument Platform (`core/instruments/` & `core/debug/`)**:
   * Capability registry and adapter discovery.
   * Debugger running a progressive triage and experiment loop.
   * Hardcoded Go build checks, test runners, and unclosed body leak AST analyzer.
6. **Workspace Maintenance (`core/maintenance/`)**:
   * Scanners for `.env` drifts, lockfile/virtual-env drifts, dangling docker images, broken symlinks, and uncommitted conflict markers.

---

## B. What Is Duplicated

We identified major duplications where multiple components solve identical requirements with isolated logic:

1. **SQL Database Connection Management**:
   * **`core/graph/sqlite.go`** (`SQLiteStore` opens database `daemon.db`).
   * **`core/debug/persistence.go`** (`DebugStore` opens database `daemon.db` separately).
   * **`core/intelligence/state.go`** (`IntelligenceStateStore` opens database `daemon.db` separately).
   * *Impact*: Creates redundant file locks, multiple connection pools, and query concurrency issues on the single SQLite file.
2. **Capability and Registry Abstractions**:
   * **`sdk/plugin/plugin.go`**: Defines `Capability` string type (`CapDiscovery`, `CapDiagnosis`, etc.) and a plugin registration framework.
   * **`core/capabilities/registry.go`**: Defines a struct-based `Capability` (Name, Risk, Inputs, Preconditions) and local `Registry`.
   * **`core/instruments/capability.go`**: Defines another distinct string `Capability` (`CapDebug`, `CapUnitTesting`, etc.) and availability mapping.
   * *Impact*: Confuses "plugin type" with "instrument measurement capability" and "action capability". No single registry maps them coherently.
3. **Tool Availability & Discovery Checks**:
   * **`cli/commands/doctor.go`** (`CheckBinaryInstalled` checks PATH for `go`/`node`/`docker` via `exec.LookPath`).
   * **`core/instruments/discovery.go`** (`IsBinaryInstalled` duplicates the exact same `exec.LookPath` check).
4. **Environment / Tech Profiling**:
   * **`core/discovery/discovery.go`** (inspects `go.mod`, `Cargo.toml`, `requirements.txt` to profile project languages).
   * **`core/instruments/discovery.go`** (`EnvironmentDetector` duplicates files walking and extension checks for language matching).

---

## C. What Is Incomplete

Several modules are declared in file structures but contain zero logic or generic stubs:

1. **Mock Investigators (`core/debug/investigators/`)**:
   * 13 files (`triage.go`, `build_failure.go`, `concurrency.go`, `memory.go`, etc.) return static, fake evidence objects.
   * *Zero Imports*: The entire package is dead code and is never imported or called by the main debugging engine.
2. **Empty Package Stubs**:
   * `core/search/` contains no Go source files.
   * `core/agent/testing/testing.go` is an empty shell.
3. **Automation & Workflows (`core/automation/` & `core/workflow/`)**:
   * YAML recipes are parsed by string checks (`strings.Contains(recipeYAML, "Reset")`) rather than a real parser.
   * Vertices in the Orchestration DAG execute mock actions.

---

## D. What Is Over-Engineered

1. **Unused CLI Commands**:
   * 42 separate command files in `cli/commands/` partition functionality excessively (e.g., `daemon search`, `daemon why`, `daemon diagnose`, `daemon cockpit`, `daemon onboard`, `daemon daily`, `daemon simulate`).
   * `diagnose.go` duplicates 90% of `analyze.go` pipeline setup.
2. **Disconnected Reasoning Subsystems**:
   * A heavy suite of packages (`core/advisor/`, `core/insights/`, `core/risk/`, `core/recommendation/`) generate shallow text findings using simple string switches, but are mapped to separate interfaces (`Engine`, `Advisor`) rather than contributing directly to the unified Knowledge Graph / Engineering Twin.

---

## E. What Is Incorrectly Coupled

1. **CLI Commands Bypassing Runtime Container**:
   * Rather than referencing dependencies resolved in `rt.Container`, multiple Cobra commands manually instantiate their own helper engines, databases, and adapters.
   * `cli/commands/debug.go` spins up its own `DebugStore` database connection and `NewDebugger` wrapper directly, bypassing the Container's lifecycle.
   * `cli/commands/analyze.go` passes `nil, nil` to the analyzer pipeline, neutralizing resource monitoring.

---

## F. What Should Be Deleted

1. **Unused Investigators**:
   * `core/debug/investigators/` (all 13 file stubs) since their behavior is replaced by native instrument adapters.
2. **Duplicate/Redundant CLI Commands**:
   * `diagnose.go` (merge into `analyze.go`).
   * `why.go` (merge into `ask.go` or `debug.go`).
   * `daily.go` (unused, static stub).
   * `demo.go` (used for walkthrough demos, redundant in production platform).
   * `search.go` (redundant, `inspect` and `graph` provide identical utility).

---

## G. What Should Be Merged

1. **Database Persistence**:
   * Merge `DebugStore` and `IntelligenceStateStore` database connections and helper tables into a single database provider inside `core/storage` or share the `SQLiteStore` instance.
2. **Unified Capability Registry**:
   * Merge `sdk/plugin.Capability`, `capabilities.Capability`, and `instruments.Capability` into a single, clean capability registry managed inside the Universal Instrument Platform.

---

## H. What Should Become a Reusable Platform Primitive

1. **`SafeExecutor` / Safety Gateway**:
   * Wrap executing raw commands into a single platform-level primitive that automatically handles dry-run flags, sanitizes arguments, checks credentials, queries the Policy Engine, and monitors the Resource Governor.
2. **Chronological Event Bus**:
   * Make `events.EventBus` the standard audit trail for all tools, CLI executions, and model predictions.

---

## I. What Should Remain Product-Specific

1. **BubbleTea TUI Cockpit (`cli/tui/`)**:
   * Interactive developer rendering.
2. **Cobra CLI Shell (`cli/commands/`)**:
   * Keeps CLI integration decoupled from core business libraries.

---

## J. What Capabilities Are Missing

1. **Generic Test Discovery & Run Rules**:
   * Testing relies on hardcoded `go test` executes rather than dynamically detecting testing toolchains via the Tech Profile.
2. **Non-Go AST / Static Scanners**:
   * No toolchain adapters exist to run generic tools like Semgrep, Ruff, Pylint, or Jest.

---

## K. What Assumptions Currently Make Daemon Go-Specific

1. **Hardcoded Checks in `RunInvestigation` (`core/debug/debugger.go`)**:
   * Direct execution of `go build ./...` and `go test ./...` on detecting `go.mod`.
   * Unclosed HTTP body scanner parses only Go source ASTs (`checkBodyCloseLeak`).
2. **Go-Only Adapters**:
   * Only `adapters/build/go` and `adapters/testing/go` compile and verify results.

---

## L. What Prevents Daemon from Working Across Languages and Ecosystems

1. **Lack of Language-Agnostic Profile Discovery**:
   * The debugger is unaware of how to run build compile checks or test regressions for a Node, Python, Java, or Rust repository because compiler/test operations are hardcoded to check for `go.mod`.
2. **Absence of Standard Instrument Schemas**:
   * Normalization is hand-rolled inside each hardcoded compiler check. A generic platform requires that any instrument outputting standard diagnostic formats (e.g., compile errors, test failure traces, leak states) is mapped via technology profiles into unified evidence structures.

---

## Proposed Restructuring Plan (Phase 1-3)

To address these weaknesses, we will perform the following targeted refactors:

1. **Consolidate SQLite Connections**:
   * Update `NewDebugger` and CLI commands to share the main `SQLiteStore` database connection from `rt.Container`, avoiding multi-connection locks on `daemon.db`.
2. **Language-Agnostic Environment Profiling**:
   * Migrate the hardcoded compiler, test, and leak checks in `debugger.go` into generic capability requests (`CapBuild`, `CapUnitTesting`, `CapStaticAnalysis`).
   * Discover and register Go, Node, and Python capability adapters in the registry.
   * Query the profile inside `RunInvestigation` to select the correct instrument dynamically.
