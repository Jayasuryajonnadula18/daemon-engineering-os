# Security & Data Isolation Policy — Daemon Engineering OS

> **Current Project Status:** Alpha — Pre-Validation  
> **Security Audit Level:** Phase 0 Trust Foundations  

---

## 🔒 1. Credential Security & OS Keyring Storage

Daemon uses **OS-native credential storage** (`zalando/go-keyring`) to manage authentication secrets. 
- Secrets are stored directly in **Windows Credential Manager**, **macOS Keychain**, or **Linux Secret Service (libsecret)** under service name `DaemonEngineeringOS`.
- On first execution, Daemon generates a 256-bit cryptographically random token (`crypto/rand`).
- **No hardcoded secrets or default passwords exist anywhere in the binary or codebase.**
- To view your active master token, run `daemon token` from your authenticated terminal session.

---

## 🛡️ 2. System Access Levels & Permitted Scope

### A. GitHub API Scope
- **Required Scope:** Read-Only (`contents:read`, `metadata:read`, `issues:read`).
- **Private Repositories:** Requires `repo` (Classic) or Fine-Grained Read-Only PAT.
- **Strict Boundary:** Daemon **NEVER** performs `git push`, `git push --force`, pull request creation, branch deletion, or repository setting modifications without explicit interactive prompt.

### B. Docker Socket Access
- **Required Access:** Inspect & Status (`docker ps`, `docker inspect`).
- **Strict Boundary:** Daemon will **NEVER** delete host volumes, modify host network interfaces, or prune running production containers without explicit user confirmation.

### C. Local Filesystem Write Boundaries
- **Scoped Directories:** Workspace project root (for approved repair fixes) and `~/.daemon/` (for SQLite Knowledge Graph, configuration, and audit logs).
- **Strict Boundary:** Daemon **NEVER** modifies files outside your active project directory or `~/.daemon/`.

### D. Network Calls & Telemetry
- **Local-First Policy:** Daemon sends **ZERO** cloud telemetry, analytics, or tracking data.
- **Permitted Outbound Destinations:**
  1. `http://localhost:11434` (or user-configured local `OLLAMA_HOST`)
  2. `https://api.github.com` (only if `GITHUB_PAT` is provided by user)
  3. `http://127.0.0.1:8081` (Mission Control local web dashboard)

---

## ⛔ 3. Non-Negotiable Interactive Confirmation Rules

Daemon enforces a strict Policy Engine barrier. The following actions **CANNOT** be executed automatically or unattended:

| Action | Interactive Confirmation Required |
|---|---|
| Force-pushing Git commits (`git push -f`) | 🔴 MANDATORY (No flag bypass) |
| Production / Staging Deployment (`daemon deploy prod`) | 🔴 MANDATORY (No flag bypass) |
| Secret or Token Rotation | 🔴 MANDATORY |
| Deleting uncommitted git working tree changes | 🔴 MANDATORY |
| Pruning container volumes or host images | 🔴 MANDATORY |

---

## ☣️ 4. Threat Model & Blast Radius Mitigation

### Fix Engine Blast Radius
- **Risk:** An AI-generated code repair or configuration update introduces a syntax or logic error.
- **Mitigation:**
  1. **Diff Preview:** Every fix produces a line-by-line before/after diff before application.
  2. **Atomic Backup:** Prior to applying any fix, Daemon creates an in-memory or patch-file snapshot.
  3. **Verified Rollback Path:** Every fix supports `daemon fix --rollback`, which instantly restores the exact pre-fix state.

### Deploy Engine Blast Radius
- **Risk:** Unintentional deployment of unverified code to staging or production environments.
- **Mitigation:**
  1. Any deployment targeting environments matching `prod`, `production`, or `staging` halts execution.
  2. Requires explicit `stdin` user input typing `yes` or confirming the interactive prompt.
  3. Flag overrides (such as `--yes` or `--force`) are **explicitly ignored and rejected** for production/staging target environments.

---

## 🌐 Reporting Vulnerabilities

If you discover a security vulnerability or credential leak risk within Daemon, please do not open a public issue. Email security reports directly to the maintainers or submit a private security advisory on GitHub.
