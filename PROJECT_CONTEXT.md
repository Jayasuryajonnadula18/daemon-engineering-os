# DAEMON ENGINEERING OS — PROJECT CONTEXT & STATE SUMMARY

> **Version:** 1.0.0 (Alpha)  
> **Status:** Alpha — Pre-Validation  
> **Active Pillars:** 24/24 (Scope Hardening In Progress)  
> **Master Authentication:** OS Keyring (`zalando/go-keyring`) / `daemon token`  
> **Last Updated:** August 2, 2026  

---

## 🎯 Executive Overview

**Daemon** is an **Engineering Operating System** for developer platforms. Unlike reactive coding assistants (e.g., Copilot, Claude Code) that sit idle waiting for prompts, Daemon operates as a stateful, proactive platform layer. 

It maintains a living model of your workspace (**Engineering Twin**), tracks dependencies inside a **SQLite Knowledge Graph**, proactively cares for environment health (**Maintenance Engine**), and orchestrates multi-step workflows using a **provider-independent Model Router** (Ollama local LLM + Cloud fallbacks).

---

## 🏗️ Architectural Foundations & The 24 Pillars

Daemon implements **24 active engineering pillars** inside a single Go binary:

1. **Engineering Context Engine** — Reads and unifies workspace signals.
2. **Engineering Twin** — Live SQLite in-memory model of local & remote resources.
3. **Knowledge Graph** — Node & edge relational graph of services, APIs, DBs, containers.
4. **Automation Engine** — Background cron routines & triggers.
5. **Integration Manager** — Real connectors for GitHub, Docker, Kubernetes, Cloudflare.
6. **Workspace Manager** — Local port, process, and environment variable control.
7. **Cockpit TUI** — Keyboard-first terminal interface.
8. **Mission Control** — Web dashboard running at `http://127.0.0.1:8081`.
9. **SDK / Capability Registry** — Plugin ecosystem for external capabilities.
10. **Engineering Search** — Semantic and dependency search across live models.
11. **Engineering Memory** — Long-term context storage across development sessions.
12. **Timeline Engine** — Historical recording of developer sessions and changes.
13. **Recommendation Engine** — Context-aware advice generation.
14. **Architecture Intelligence** — Dependency impact analysis & graph reasoning.
15. **Risk Engine** — Pre-flight deployment & security risk assessment.
16. **Context Engine** — Multi-source context aggregator.
17. **Workflow Engine** — Deterministic DAG execution generator.
18. **Advisor Engine** — AI-powered proactive guidance.
19. **Replay Engine** — Historical execution & decision replay (`daemon replay`).
20. **Orchestrator Engine** — Task graph compiler and execution controller.
21. **Policy Engine** — Guardrails and security rules for automated actions.
22. **Fix Engine** — Automated repair workflow generator & rollback controller.
23. **Deploy Engine** — Multi-strategy deployment pipeline coordinator.
24. **Maintenance Engine** — Proactive workspace care and controlled self-healing (`daemon maintain`).

---

## 🚀 Key Accomplishments & Features Built

### 1. Pillar 24: Engineering Maintenance Engine (`core/maintenance/`)
- Implemented proactive observation, evidence gathering, and self-healing.
- Enforces strict **Core Four** scope: `.env` key presence, dependency lockfile drift, dangling Docker state (>24h exited containers / >10MB dangling images), broken symlinks & dead references.
- **The Trust Vision & Manifesto:**
  1. *Invisible when things are fine, precise when they're not.*
  2. *Knows the difference between "checked and clean" and "not checked" (Silence Contract).*
  3. *Never asks for trust — shows empirical evidence (exact file paths, line numbers, key names, sizes).*
  4. *Mechanical, deterministic, and boring — no guessing, no unrequested value edits.*
  5. *Read-only until explicitly instructed otherwise (`--apply` is opt-in).*
  6. *Serves as the trust onboarding for the entire Engineering Operating System.*
- Exposed via `daemon maintain`, with backward-compatible aliases `daemon care` and `daemon health`.

### 2. Redesigned Git-Style CLI Hierarchy (`cli/commands/`)
- Re-architected CLI around a memorable, product-focused hierarchy:
  - `daemon init` — Initialize Daemon in project repository.
  - `daemon workspace` — Manage local infrastructure and ports.
  - `daemon doctor` — Run pre-flight health diagnostics.
  - `daemon advise` — Surfacing AI advice (`workspace`, `security`, `deploy`, `daily`).
  - `daemon plan` — Generate deterministic dry-run DAG execution plans.
  - `daemon fix` — Execute approved repair workflows with rollback support.
  - `daemon deploy` — Coordinate multi-strategy deployment pipelines (`standard`, `canary`, `blue-green`).
  - `daemon maintain` — Run proactive maintenance routines.
  - `daemon graph` — Explore the Engineering Twin Knowledge Graph.
  - `daemon cockpit` — Launch the TUI interface.
  - `daemon mission` — Launch Mission Control web dashboard (alias: `dashboard`).
  - `daemon automate` — Manage cron automation routines.
  - `daemon plugin` — Manage SDK capability plugins.
  - `daemon config` — View and edit system configuration.
  - `daemon version` — Display build metadata & active pillars.
  - `daemon sync` — Synchronize integrations & run local workspace discovery.

### 3. Invincible Engine Upgrades
- **Real GitHub Integration (`GitHubConnector`):** When `GITHUB_PAT` is configured, calls `https://api.github.com/user/repos` to discover real repository names, languages, visibility, and open issue counts.
- **Real Docker Integration (`DockerConnector`):** Interfaces with live Docker socket (`docker ps`) to discover container states and ports.
- **Local LLM Client (`reasoning/llm_client.go`):** Auto-detects local Ollama instance (`http://localhost:11434`, defaulting to `qwen2.5-coder:7b`). Enriches `daemon advise` output with real local AI recommendations, gracefully falling back to structured offline reasoning if Ollama is offline.
- **Persistent Workspace Discovery:** `daemon sync` automatically walks the local directory tree, scans `package.json`, `Dockerfile`, `docker-compose.yml`, and `go.mod`, populating the SQLite Knowledge Graph with real service nodes and dependency edges.

### 4. Global Distribution & Tooling
- **Global PowerShell Installer (`install.ps1`):** Copies `daemon.exe` directly to `%LOCALAPPDATA%\Microsoft\WindowsApps\daemon.exe`, setting `DAEMON_PATH` so `daemon` is available globally in any terminal without path prefixes.
- **VS Code Extension (`vscode-extension/`):** Upgraded `daemon-vscode` with 12 contributed commands, live status bar item (`⚡ Daemon: Active`), interactive deployment strategy pickers, and persistent output channels.

---

## 🛠️ Environment Setup & Configuration

| Environment Variable | Recommended Value | Purpose |
|---|---|---|
| `DAEMON_PASSWORD` | Value from `daemon token` | Master secret token stored in OS Keyring |
| `OLLAMA_HOST` | `http://localhost:11434` | Endpoint for local Ollama LLM service |
| `OLLAMA_MODEL` | `qwen2.5-coder:7b` | Local LLM model for AI reasoning |
| `GITHUB_PAT` | `ghp_...` | GitHub Personal Access Token (Read-Only) |
| `DAEMON_PATH` | `%LOCALAPPDATA%\Microsoft\WindowsApps\daemon.exe` | Global executable path |

---

## 🧪 Verification & Health Check

All unit tests and builds have been verified:
```powershell
# Build binary
cd daemon
go build -o daemon.exe

# Run full test suite
go test ./...

# View/set your OS Keyring master secret token
$env:DAEMON_PASSWORD = (daemon token | Select-String "Token:").Line.Split()[-1]

# Verify global access & active pillars
daemon version

# Run live sync and local discovery
daemon sync

# Run proactive maintenance care
daemon maintain
```

---

## 📋 Next Steps & Roadmap

1. **Start Ollama Locally:** Run `ollama pull qwen2.5-coder:7b` to activate offline AI reasoning.
2. **Configure GitHub PAT:** Set `$env:GITHUB_PAT` to pull live GitHub repos into your Engineering Twin.
3. **Expand Unit Test Coverage:** Add test suites for remaining packages (`core/discovery`, `core/graph`, `cli/commands`).
4. **Cloud Router Expansion:** Optionally add `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` for multi-tier model routing (Claude Sonnet 4.5 for complex incident analysis).
