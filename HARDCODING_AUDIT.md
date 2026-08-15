# Daemon V1 — Comprehensive Anti-Hardcoding Audit

This document records the full audit pass across the Daemon codebase to eliminate hardcoded assumptions, paths, magic numbers, and fixed stack expectations.

---

## Audit Pass Categories & Corrections

### 1. User-Facing Messages & Strings
- **Finding**: Terminal status outputs previously assumed specific package manager names (`npm`) or language runtimes without checking detected context.
- **Correction**: All CLI messages in `daemon/cli/commands/doctor.go`, `maintain.go`, `inspect.go`, and output renderers now interpolate runtime-detected values (`info.PackageManager`, `info.Language`, `info.Framework`). Tool names are only referenced when verified present on the host system PATH.

### 2. Debug Logs & Leftover Development Artifacts
- **Finding**: Raw `fmt.Println` debug statements were prone to leaking into stdout during structured output operations.
- **Correction**: Centralized structured output rendering in `daemon/cli/output/renderer.go`. When `--json` is specified, all decorative terminal banners, progress spinners, and ANSI color codes are completely suppressed, emitting strictly valid JSON payload to stdout.

### 3. Paths & File Locations
- **Finding**: Fixed path references risked assuming specific user profile directories or single-subfolder project structures.
- **Correction**: All path operations utilize `filepath.Join`, `os.UserHomeDir()`, `os.Getwd()`, and `os.UserConfigDir()`. Canonical project databases are scoped exclusively to `project/.daemon/daemon.db`, while global user profiles reside in `~/.daemon/`. Zero hardcoded `C:\Users\` or `/home/` absolute paths remain in executable source.

### 4. Configurable Thresholds, Limits & Magic Numbers
- **Documented Default Configuration Constants**:
  - `HealthScoreWarningThreshold`: `70` (Config key: `health.warning_threshold`)
  - `HealthScoreCriticalThreshold`: `40` (Config key: `health.critical_threshold`)
  - `MaxContextTokenBudget`: `8000` tokens (Config key: `context.max_token_budget`)
  - `MaxContextEntitiesCap`: `50` entities (Config key: `context.max_entities`)
  - `DependencyDriftContentHashAlg`: `SHA256` (Config key: `maintenance.hash_algorithm`)
  - `DockerExitedContainerAgeLimit`: `24h` (Config key: `docker.exited_age_limit`)
  - `StaleBranchAgeDays`: `30d` (Config key: `git.stale_branch_days`)

### 5. Multi-Stack Language & Tool Discovery
- **Finding**: Projects containing multiple stacks (e.g. Go API + React Frontend + Python Worker) risked defaulting to a single language runtime.
- **Correction**: Enhanced `daemon/core/discovery/discovery.go` to scan for multi-stack markers (`go.mod`, `package.json`, `Cargo.toml`, `pom.xml`, `requirements.txt`, `pnpm-workspace.yaml`). Unrecognized stacks degrade safely to `"Unknown / Custom"` instead of guessing with high confidence.

### 6. Credentials, Tokens & Identifiers
- **Finding**: Authentication require zero plaintext secrets or hardcoded fallback organization names.
- **Correction**: Master secret tokens are generated securely and stored in the OS Keyring (`Service: DaemonEngineeringOS`). Access checks accept `$env:DAEMON_PASSWORD` or `--password` flags. Placeholder values in generated templates are explicitly formatted as `<your-org>/<your-repo>`.

### 7. Locale, Timezone & Formatting
- **Correction**: All event timestamps, Twin updates, and memory logs use standard RFC3339 format (`2006-01-02T15:04:05Z07:00`) tied to system clock state.

### 8. AI & Model-Layer Dynamism
- **Correction**: Reasoning prompts in `daemon/core/reasoning/` dynamically inject context compiled by the `ContextEngine` (`info.Language`, `info.Framework`, `info.Services`), ensuring local LLM prompts adapt to whatever project Daemon is currently analyzing.

---

## Multi-Project Empirical Runs & Verification Results

1. **Test Run 1: Python Project Fixture (`tmp-fixtures/python-project`)**
   - Result: Detected stack as Python without Node/npm requirements.
2. **Test Run 2: Go Module Project Fixture (`tmp-fixtures/go-project`)**
   - Result: Detected stack as Go module service (`go.mod`) without hardcoded path errors.
3. **Test Run 3: Daemon Standalone Executable (`daemon.exe doctor --json`)**
   - Result: Executed clean schema-conforming JSON output with zero debug logs.
