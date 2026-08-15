# Daemon V1 — Master Architecture & Project Context
Last updated: 2026-08-08
Consolidated Architecture: Daemon V1 (5 Frozen Layers, Pure Go Standalone Binary `daemon.exe`)

---

## 🏛️ Daemon V1 Frozen Layer Hierarchy

### Layer 5 — Experience
- **Components**: CLI (`daemon/cli/commands`), VS Code Extension (`vscode-extension/`), Engineering Cockpit, Mission Control (`daemon/mission-control`).
- **Core Rule**: Clients contain zero business logic; all clients consume identical Core APIs from Layer 1–4.

### Layer 4 — Orchestration
- **Components**: Engineering Orchestrator (`daemon/core/reasoning`), Impact Intelligence (`daemon/core/insights`), Cross-Team Coordination.

### Layer 3 — Intelligence
- **Components**: Model Router (`daemon/core/reasoning`), Workflow Intelligence, Advisor (`daemon/core/advisor`), Replay Intelligence (`daemon/core/replay`).

### Layer 2 — Engineering Knowledge
- **Components**: Engineering Context Engine (`daemon/core/context` — First-Class), Engineering Twin (`daemon/core/twin`), Knowledge Graph (`daemon/core/graph`), Event Bus (`daemon/core/events`), Memory / Fix Ledger (`daemon/core/storage`), Knowledge Ranking (`Personal > Project > Organization > Generic`).
- **Canonical Storage Model**:
  - Project DB (Source of Truth): `project/.daemon/daemon.db` (SQLite)
  - Project Exports / Inspection: `project/.daemon/context.json`, `project/.daemon/graph.json`
  - Global / User Profile DB: `~/.daemon/personal/`, `~/.daemon/config/`, `~/.daemon/credentials/`

### Layer 1 — Trust + Execution
- **Components**: Policy Engine (`daemon/core/policies`), Capability Registry (`daemon/core/capabilities`), Automation Engine (`daemon/core/automation`), Fix Engine (`daemon/cli/commands/fix.go`), Maintenance Engine (`daemon/core/maintenance`), Integrations (`daemon/core/integrations`), Verification Engine, Rollback Engine.
- **Strict Rule**: AI reasoning NEVER executes shell commands directly (`AI -> Intent -> Plan -> DAG -> Policy -> Capability -> Automation -> Integration -> Verification -> Rollback`).

---

## 📊 Current Layer Graduation Status

| Layer | Name | Status | Verified Evidence |
|---|---|---|---|
| **Layer 1** | Trust + Execution | **MET (PASS)** | Content-hash dependency drift, datamine classifier, 43+ spec test suite passing, Policy Engine hard ceilings, 3-fix apply/verify/rollback, zero AI shell bypasses. |
| **Layer 2** | Engineering Knowledge | **MET (PASS)** | Context Engine, SQLite Twin, Knowledge Graph traversal, Event Bus, `--json` output across CLI, multi-stack Go/Node/Python/Rust/Java discovery. |
| **Layer 3** | Intelligence | **MET (PASS)** | ModelRouter (ModelCapability, task-aware), EngineeringReasoner (ReasoningResult, evidence refs, hard confidence caps), IntelligenceStateStore (SQLite), Workflow Intelligence, PredictionEngine, Deterministic Benchmark. |
| **Layer 4** | Orchestration | **MET (PASS)** | DAGCompiler (PlanFreshness, plan hash), WaveScheduler (ResourceLock matrix), ImpactEngine (blast radius), CheckpointStore (SQLite idempotency, --resume), Orchestrator state machine, --dry-run mode, cooperative cancellation. |
| **Layer 5** | Experience | **MET (PASS)** | Single Go CLI binary `daemon.exe`, Versioned `/api/v1/` REST API, `APIResponse[T]` envelope, SSE real-time stream (`Last-Event-ID`), Authenticated IDE event ingestion, Embedded Engineering Cockpit SPA, `daemon cockpit` CLI commands. |

---

## 🔒 Non-Negotiable V1 Architectural Rules

1. **Architecture Frozen**: 5 layers only. No unclassified top-level engines.
2. **Context Engine is First-Class**: `User Intent -> Context Engine -> Twin/Graph/History -> Bounded Context -> AI`.
3. **State vs History Separated**: Twin (Now), Memory (Learned), Replay (Happened), Graph (Relations), Workflow History (Work patterns).
4. **SQLite is Source of Truth**: SQLite DB is canonical; JSON files (`context.json`) are exports/inspectors only.
5. **Knowledge Provenance**: Every observation records `source`, `timestamp`, `freshness`, `confidence`, `scope`.
6. **AI / Execution Boundary**: Never `LLM -> shell`. LLM generates plan/intent; Automation Engine executes registered Capability under Policy Engine control.
7. **First-Class Verification & Rollback**: Every Capability defines explicit `Verify` and `Rollback` execution checks.
8. **Knowledge Ranking**: Fixed priority: `Personal > Project > Organization > Generic`.
9. **Stabilized CLI Command Set**: `init`, `inspect`, `doctor`, `search`, `ask`, `plan`, `fix`, `maintain`, `workspace`, `deploy`, `replay`, `advise`, `integrations`, `sync`, `config`.

