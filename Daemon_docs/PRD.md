# DAEMON PRD (Product Requirements Document)

**Version:** 0.1 MVP\
**Status:** Draft\
**Owner:** Founding Team

------------------------------------------------------------------------

# 1. Vision

Daemon is an Engineering Operating System that allows developers to
express engineering intent instead of remembering commands.

Instead of:

-   Start Docker
-   Check PostgreSQL
-   Install dependencies
-   Verify Node
-   Start frontend
-   Start backend

The developer simply writes:

``` bash
daemon start
```

Daemon discovers the project, builds an execution plan, orchestrates the
required tools, verifies health, and reports results.

------------------------------------------------------------------------

# 2. Mission

Reduce the time developers spend configuring, debugging, and
orchestrating tools so they can spend more time building software.

------------------------------------------------------------------------

# 3. Problem Statement

Modern software development is fragmented.

Developers constantly switch between:

-   Terminal
-   Git
-   Docker
-   Kubernetes
-   npm / pnpm
-   Databases
-   Cloud dashboards
-   Documentation

Most engineering effort is spent coordinating tools rather than writing
business logic.

Daemon aims to reduce that coordination overhead.

------------------------------------------------------------------------

# 4. Target Users

## Primary

-   Full-stack developers
-   Startup engineers
-   Indie hackers
-   DevOps engineers

## Secondary

-   Students
-   Open-source contributors
-   Small engineering teams

------------------------------------------------------------------------

# 5. Product Principles

1.  Intent over commands.
2.  Deterministic by default.
3.  AI is optional.
4.  Explain every action.
5.  Never perform destructive actions without confirmation.
6.  Build trust before automation.

------------------------------------------------------------------------

# 6. MVP Scope

Commands:

``` text
daemon init
daemon start
daemon doctor
daemon recover
daemon verify
daemon health
```

Out of scope:

-   IDE
-   Cloud sync
-   Autonomous coding
-   Production deployment orchestration

------------------------------------------------------------------------

# 7. User Stories

### As a developer

I want to type `daemon start` so my local environment is prepared
automatically.

### As a developer

I want `daemon doctor` to tell me what is wrong before I waste time
debugging.

### As a developer

I want `daemon recover` to safely fix common engineering issues.

------------------------------------------------------------------------

# 8. Functional Requirements

## Discovery

Detect:

-   Framework
-   Runtime
-   Package manager
-   Docker
-   Git
-   Database
-   Environment variables

## Planning

Generate:

-   Startup plan
-   Recovery plan
-   Verification plan

## Health

Produce:

-   Engineering Health Score
-   Warnings
-   Recommendations

------------------------------------------------------------------------

# 9. Non-Functional Requirements

-   Cross-platform (Windows, macOS, Linux)
-   TypeScript
-   Node.js
-   Fast startup
-   Modular architecture
-   Extensible adapters
-   Offline-first

------------------------------------------------------------------------

# 10. Success Metrics

-   Reduce local setup time.
-   Reduce environment-related failures.
-   One-command startup for supported projects.
-   Accurate engineering diagnosis.

------------------------------------------------------------------------

# 11. Roadmap

## Phase 1

CLI MVP

## Phase 2

VS Code Extension

## Phase 3

Engineering Knowledge Platform

## Phase 4

Daemon IDE

------------------------------------------------------------------------

# 12. Technical Direction

Core Engines:

-   Discovery Engine
-   Planning Engine
-   Orchestration Engine
-   Health Engine
-   Recovery Engine
-   Confidence Engine
-   Optional AI Engine

AI never executes commands. It reasons over engineering context.

------------------------------------------------------------------------

# 13. Risks

-   Supporting too many technologies too early.
-   Overusing AI where deterministic logic is sufficient.
-   Complex plugin compatibility.

Mitigation:

-   Focus on the most common web-development stack first.
-   Add technologies incrementally.

------------------------------------------------------------------------

# 14. Definition of MVP Success

A developer should be able to:

1.  Install Daemon.
2.  Run `daemon doctor`.
3.  Understand project health.
4.  Run `daemon start`.
5.  Begin coding with minimal manual setup.
6.  Run `daemon recover` to resolve common local issues.

If Daemon reliably accomplishes these goals, the MVP is successful.
