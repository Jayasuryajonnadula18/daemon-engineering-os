# MVP.md

**Project:** Daemon CLI\
**Version:** 0.1

------------------------------------------------------------------------

# MVP Goal

Build a developer CLI that understands a project, orchestrates the local
development environment, diagnoses engineering issues, safely recovers
common problems, and establishes the foundation for the future Daemon
platform.

The MVP is **not** an AI coding assistant or an IDE.

------------------------------------------------------------------------

# Success Criteria

A developer should be able to:

1.  Install Daemon.
2.  Run `daemon init`.
3.  Run `daemon doctor`.
4.  Run `daemon start`.
5.  Begin coding.
6.  Run `daemon recover` if something breaks.
7.  View engineering health.

------------------------------------------------------------------------

# Core Features

## 1. Project Discovery

Command:

``` bash
daemon doctor
```

Detect:

-   Framework
-   Runtime
-   Package manager
-   Docker
-   Docker Compose
-   Git
-   Database
-   Environment files
-   Ports
-   Installed dependencies

Output: - ProjectContext - Detected technologies - Confidence score

Acceptance Criteria

-   Detects supported technologies with high confidence.
-   Produces a structured context object.

------------------------------------------------------------------------

## 2. Startup Orchestration

Command:

``` bash
daemon start
```

Responsibilities

-   Read ProjectContext
-   Build execution plan
-   Verify prerequisites
-   Start required services
-   Verify engineering health

Acceptance Criteria

-   Starts supported projects using deterministic plans.
-   Reports each execution step.

------------------------------------------------------------------------

## 3. Engineering Doctor

Command

``` bash
daemon doctor
```

Checks

-   Missing dependencies
-   Docker availability
-   Git status
-   Node installation
-   Package manager
-   Database connectivity
-   Environment variables
-   Common configuration issues

Output

-   Engineering Health
-   Diagnosis
-   Recovery suggestions

------------------------------------------------------------------------

## 4. Recovery

Command

``` bash
daemon recover
```

Recoverable actions

-   Install dependencies
-   Start Docker
-   Restart containers
-   Generate missing .env from template
-   Free occupied ports (with confirmation)
-   Restart local services

Acceptance Criteria

-   Never performs destructive actions without confirmation.
-   Automatically verifies recovery.

------------------------------------------------------------------------

## 5. Verification

Command

``` bash
daemon verify
```

Checks

-   Runtime
-   Dependencies
-   Docker
-   Database
-   Services
-   Environment

Outputs pass/fail results.

------------------------------------------------------------------------

## 6. Engineering Health

Command

``` bash
daemon health
```

Display

-   Health score
-   Healthy services
-   Warnings
-   Recommendations
-   Last scan timestamp

------------------------------------------------------------------------

# Technology Support (MVP)

Frameworks

-   React
-   Next.js
-   Express

Languages

-   JavaScript
-   TypeScript

Infrastructure

-   Docker
-   Docker Compose
-   Kubernetes (detection only)

Package Managers

-   npm
-   pnpm

Databases

-   PostgreSQL
-   MongoDB
-   Redis (basic detection)

------------------------------------------------------------------------

# Developer Mode

Reserved for Daemon contributors.

Commands

``` bash
daemon dev tech add
daemon dev detect
daemon dev graph
daemon dev simulate
daemon dev knowledge
```

Purpose

-   Extend Daemon
-   Test adapters
-   Debug discovery
-   Build engineering knowledge

------------------------------------------------------------------------

# User Teaching Layer

Commands

``` bash
daemon teach preferences
daemon teach workflow
daemon teach remember
daemon teach forget
```

Stores user-specific workflows without changing global knowledge.

------------------------------------------------------------------------

# AI (Optional)

Supported Providers

-   Ollama
-   OpenAI
-   Anthropic
-   Gemini

AI Responsibilities

-   Explain issues
-   Suggest recovery
-   Analyze unknown technologies

AI Never

-   Executes commands
-   Modifies files directly
-   Overrides deterministic logic

------------------------------------------------------------------------

# Non-Goals

The MVP does not include

-   Autonomous coding
-   Cloud synchronization
-   Production deployment
-   IDE
-   Engineering Knowledge Graph runtime
-   Plugin marketplace

------------------------------------------------------------------------

# Milestones

## Sprint 1

-   CLI bootstrap
-   Command framework
-   Folder structure

## Sprint 2

-   Discovery Engine
-   ProjectContext
-   Adapters

## Sprint 3

-   Planning Engine
-   Orchestration Engine
-   daemon start

## Sprint 4

-   doctor
-   verify
-   health

## Sprint 5

-   recover
-   developer mode
-   user teaching layer

## Sprint 6

-   AI integration (optional)
-   polish
-   documentation
-   first public release

------------------------------------------------------------------------

# Definition of Done

A feature is complete only if:

-   Strongly typed
-   Unit tested
-   Integrated with engines
-   Uses adapters
-   Updates engineering health
-   Logs execution
-   Produces explainable output
-   Has documentation

------------------------------------------------------------------------

# Release Target

Version 0.1 should solve a real developer problem:

> "I cloned a project. I typed one command. My environment was analyzed,
> started, verified, and ready for development."
