# 🏔️ Daemon Maintain — Full-Scale Architecture Spec (North Star Blueprint)

> **Status:** Long-Term North Star Specification  
> **MVP Foundation:** Core Four (Strict Mechanical Precision: `.env` key parity, lockfile timestamps, dangling Docker >24h/>10MB, broken symlinks).  
> **Philosophy:** *The four-check MVP earns the right to build toward this catalog one piece at a time under strict evidence, zero false positives, fault isolation, and silence contract principles.*

---

## 1. Detection Surface — Full Catalog

### 1.1 Environment & Configuration
- **`.env` vs `.env.example` Key Parity** (Existing MVP Scope).
- **Multi-Environment Reconciliation:** `.env.local`, `.env.staging`, `.env.production` cross-checked for structural consistency (same key structure expected across targets).
- **Secret Structural Sanity Validation:** Detect keys expected to be integers/URLs/tokens that contain malformed patterns (e.g. `PORT=abc`, `REDIS_URL=localhost`).
- **Environment Variable Declaration Mismatch:** Variables referenced in source code (`process.env.X`, `os.Getenv("X")`) missing from `.env*` files, and vice versa.
- **Config File Schema Drift:** `tsconfig.json`, `next.config.js`, `vite.config.ts` checked against target framework version schemas.
- **Feature Flag / Config Toggle Staleness:** Toggles referenced in code no longer existing in config, or vice versa.

### 1.2 Dependencies & Package Management
- **Content-Hash Lockfile Drift:** Content hash verification across npm, yarn, pnpm, pip, poetry, cargo, go modules.
- **Duplicate Package Manager Detection:** Flags conflicting lockfiles (e.g. `package-lock.json` AND `yarn.lock` present simultaneously).
- **Peer Dependency Resolution Conflicts.**
- **Transitive Dependency Version Conflicts:** Monorepo package resolution collisions.
- **License Compliance Drift:** Transitive dependencies introducing incompatible licenses (e.g. GPL-3 in MIT project).
- **Deprecated Package Detection:** Registry deprecation notices.
- **Unused Declared Dependencies:** Manifest packages never imported in code.
- **Used-But-Undeclared Dependencies:** Imported packages missing from manifest (hoisting artifacts).

### 1.3 Containers & Local Infrastructure
- **Project-Scoped Container State:** Docker containers, images, volumes, networks filtered strictly by compose/Dockerfile labels (never whole-host).
- **Compose vs Running-State Reconciliation:** Verification that running container state matches `docker-compose.yml` declarations.
- **Port Collision Detection:** Host port collisions across declared compose services and host bindings.
- **Resource Ceiling Warnings:** Containers approaching CPU/memory limit ceilings.
- **Multi-Service Health Check Aggregation:** Real-time health probe status across compose services.
- **Volume Mount Path Drift:** Mount paths declared in compose missing on host.

### 1.4 Filesystem Integrity
- **Broken Symlinks & Path Aliases** (Existing MVP Scope).
- **Orphaned Generated Files:** Build artifacts referencing deleted source files.
- **Case-Sensitivity Landmines:** Case mismatches (e.g. `import Component from './component'`) causing silent Linux CI breaks.
- **Accidentally Committed Large Binary Files.**
- **File Permission Anomalies:** Missing `+x` on scripts or world-writable config files.

### 1.5 Version Control State
- **Stale Local Branches:** Local branches merged upstream but never deleted.
- **Divergence From Remote:** Local branch drifting far behind remote origin.
- **Uncommitted Changes Aging Threshold:** Stale uncommitted changes older than threshold.
- **Detached HEAD State Warnings.**
- **Accidental Merge Conflict Markers:** Committed conflict markers (`<<<<<<< HEAD`).
- **Submodule Pointer Drift.**

### 1.6 Database & Data Layer
- **Pending Migrations:** Unapplied ORM migrations relative to local database.
- **Schema Drift:** ORM model definitions vs live database schema.
- **Seed Data Staleness.**
- **Orphaned Test Databases.**
- **Connection Pool Exhaustion Patterns.**

### 1.7 Testing & Quality Signals
- **Orphaned Test Files:** Test files with no corresponding source implementation.
- **Structural Test Coverage Gaps:** Source components with zero corresponding test files.
- **Flaky Test Pattern Detection:** Historical intermittent failure tracking.
- **Stale Test Snapshots.**
- **Disabled / Skipped Test Accumulation.**

### 1.8 Security & Secrets Hygiene
- **Secret-Shaped Strings:** Hardcoded API keys and JWT tokens in tracked files.
- **Dependency CVE Vulnerability Exposure.**
- **Overly Permissive File Permissions on Sensitive Config.**
- **Expired Certificates / Tokens.**
- **`.gitignore` Coverage Gaps:** Missing stack patterns in `.gitignore`.

### 1.9 Documentation & Contract Sync
- **README Example Code Drift:** Outdated CLI flags or function names in READMEs.
- **OpenAPI / GraphQL Schema vs Implementation Drift.**
- **Changelog Staleness.**
- **Broken Internal Documentation Links.**

### 1.10 Build & Artifact Hygiene
- **Stale Build Cache Growth.**
- **Unbounded Log File Growth.**
- **Orphaned Process Temp Files.**
- **Conflicting Build Output Directories.**

### 1.11 Cross-Project / Workspace-Level (Engineering Twin)
- **Port Allocation Conflicts Across Open Projects.**
- **Shared Local Service Dependency Awareness.**
- **Global Toolchain Version Consistency:** Active shell runtime vs project expected runtime.

---

## 2. Operating Modes

1. **On-Demand Mode (`daemon maintain`)**: Full synchronous scan, complete empirical evidence report, non-zero exit code on critical drift.
2. **Watch / Background Mode (`daemon maintain --watch`)**: Event-driven mtime diffing via Engineering Twin. Notifies only on genuine state changes.
3. **CI Mode (`daemon maintain --ci`)**: Machine-readable output (`--json`), non-zero exit code gating CI pull request merges.
4. **Pre-Commit Hook Mode**: Fast subset scan (`.env`, lockfiles, secrets) running <200ms before git commit.
5. **Scheduled Deep Scan**: Nightly/weekly catalog run catching external updates (CVE databases, cert expiry).
6. **Team-Shared Mode**: Project-wide workspace health state synced across team members.

---

## 3. Intelligence Layer (Opt-In AI Assistance)

- **Root-Cause Correlation:** Correlates multi-check findings back to shared events (e.g. *"Dependency drift + broken symlinks both trace back to incomplete npm install after branch switch"*).
- **Natural-Language Catchup Summaries:** Summarizes workspace activity after time away.
- **Historical Trend Narration:** Identifies recurring drift patterns over time.
- **Priority Ranking:** Ranks multi-finding lists by severity and blast radius.

---

## 4. Fix & Remediation Layer

- **Exact Command & Diff Mapping:** Every finding maps to a previewable fix.
- **Tiered Safety Execution:** Notify ──> Suggest with diff ──> Apply with snapshot ──> Full auto.
- **Batch Mode:** Review and approve multiple proposed fixes in a single interactive pass.
- **Rollback Ledger:** Persistent history of applied repairs via `daemon maintain --rollback`.
- **Dry-Run Mode:** Full preview before filesystem mutation (`--dry-run`).

---

## 5. Reporting & Interface Surface

- **Human-Readable CLI Dashboard** (Existing MVP).
- **Machine-Readable `--json` / `--yaml`** (Existing MVP).
- **TUI Interactive Dashboard** (`daemon cockpit`).
- **Mission Control Web Dashboard** (`daemon mission` at `http://127.0.0.1:8081`).
- **VS Code Status Bar & Inline Gutter Surfacing.**
- **Webhook & Team Channel Notifications.**
- **Exportable Health Reports** (JSON / Markdown).

---

## 6. State & Data Model

- **Persistent SQLite State Store (`graph.db`):** Last scan times, baselines, finding history, fix ledgers.
- **Project Configuration (`.daemon/maintain.yaml`):** Check toggles, thresholds, and path exclusions.
- **Baseline Snapshotting:** Explicit "known-good" baseline marking.

---

## 7. Performance & Reliability Envelope

- **Warm Cache Scan Speed:** `<1 second` for typical single-service repositories.
- **Monorepo Scan Speed:** `<5 seconds` for monorepos with hundreds of packages.
- **Event-Driven Watch Overhead:** Zero CPU polling loops.
- **Fault Isolation:** 1 failing check never aborts the remaining checks.
- **100% Offline Core Operations:** All mechanical checks run locally without external network dependencies.
