# Daemon CLI

Daemon CLI is a developer-focused engineering assistant for JavaScript/TypeScript projects. It discovers project context, evaluates health, generates recovery guidance, and executes startup workflows with memory-aware decisions.

## Install

1. Clone the repository:
   ```bash
   git clone <repo-url>
   cd "Daemon CLI"
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Build the CLI:
   ```bash
   npm run build
   ```
4. Install globally for easy access:
   ```bash
   npm install -g .
   ```

## Developer commands

- `daemon doctor`
  - Discover project environment
  - Evaluate Git, Docker, Node, env vars, and system state
  - Render an engineering health report

- `daemon status`
  - Show current project context and health

- `daemon verify`
  - Run a targeted verification check
  - Report PASS/PARTIAL/FAIL based on health score

- `daemon recover`
  - Generate recovery actions for missing dependencies, env vars, Git, and Docker when required

- `daemon start`
  - Build and execute a startup plan
  - Reuse prior startup state to skip redundant installation when safe
  - Avoid unnecessary Docker validation unless the project actually requires it

## Current capabilities

- adaptive project discovery from `package.json`, lockfiles, scripts, env files, and dependency patterns
- environment-aware health rules instead of fixed procedural checks
- recovery planning that surfaces explicit actions rather than vague guidance
- stateful startup planning that remembers prior successful runs and can skip redundant work

## Notes

- The CLI is designed to reduce developer friction by replacing manual inspection with a single command workflow.
- It is still rule-driven: recognized frameworks, lockfiles, dependency hints, and service requirements are encoded in the discovery and health engines.

## Local execution

If you do not want to install globally, run:
```bash
npm start -- start
```

or use `npx` during development:
```bash
npx ts-node src/index.ts start
```
