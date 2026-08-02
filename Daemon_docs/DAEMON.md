# DAEMON.md

> Master Context File for AI Coding Assistants
>
> Read this document before generating, modifying, or reviewing any code
> in the Daemon repository.

------------------------------------------------------------------------

# Identity

Daemon is an **Engineering Operating System**.

It is **not**: - an IDE - a package manager - a container runtime - an
AI coding assistant - a deployment platform

Daemon orchestrates engineering tools and understands developer intent.

------------------------------------------------------------------------

# Product Vision

A developer should express **intent**, not remember workflows.

Instead of:

``` text
docker compose up
npm install
npm run dev
```

The developer writes:

``` bash
daemon start
```

Daemon discovers the project, plans execution, orchestrates tools,
verifies health, and explains every action.

------------------------------------------------------------------------

# Engineering Principles

1.  Deterministic before AI.
2.  AI is optional.
3.  Never guess when confidence is low.
4.  Never execute destructive actions without confirmation.
5.  Every action must be explainable.
6.  Orchestrate existing tools instead of replacing them.

------------------------------------------------------------------------

# Core Commands

-   daemon init
-   daemon start
-   daemon doctor
-   daemon recover
-   daemon verify
-   daemon health

Future: - daemon ask - daemon impact - daemon deploy - daemon plugins

------------------------------------------------------------------------

# Architecture

``` text
CLI
 ↓
Discovery Engine
 ↓
Planning Engine
 ↓
Confidence Engine
 ↓
Orchestration Engine
 ↓
Adapters
 ↓
Git • Docker • Kubernetes • Node • Databases
```

------------------------------------------------------------------------

# Coding Rules

-   Use TypeScript.
-   Use Commander.js.
-   Business logic never belongs inside CLI command files.
-   Commands delegate to engines.
-   Engines delegate to adapters.
-   Adapters are the only layer allowed to interact with external tools.
-   Prefer composition over inheritance.
-   Every public interface must be strongly typed.

------------------------------------------------------------------------

# Project Structure

``` text
src/
  commands/
  core/
  adapters/
  services/
  models/
  storage/
  utils/

.daemon/
```

------------------------------------------------------------------------

# AI Rules

AI must never: - execute shell commands - modify files directly - invent
project state

AI may: - explain issues - recommend actions - analyze unknown
technologies - summarize engineering context

The Rule Engine is always the source of truth.

------------------------------------------------------------------------

# Supported Technologies (MVP)

Frameworks: - React - Next.js - Express

Infrastructure: - Docker - Docker Compose - Kubernetes (detection only)

Languages: - TypeScript - JavaScript

Package Managers: - npm - pnpm

Version Control: - Git

Databases: - PostgreSQL - MongoDB - Redis

------------------------------------------------------------------------

# Engineering Health

Health is calculated from:

-   Runtime
-   Dependencies
-   Docker
-   Environment
-   Services
-   Database
-   Git
-   Configuration

------------------------------------------------------------------------

# Long-Term Direction

Phase 1: CLI

Phase 2: VS Code Extension

Phase 3: Engineering Knowledge Graph

Phase 4: Daemon IDE

------------------------------------------------------------------------

# Definition of Done

Every feature must:

-   solve a real developer problem
-   reduce context switching
-   be testable
-   be modular
-   preserve deterministic behavior
-   improve Engineering Health

If a proposed feature does not reduce developer friction or strengthen
orchestration, it does not belong in Daemon.

# Knowledge Layers

Daemon operates with four independent knowledge layers.

## 1. Global Knowledge

Bundled with Daemon.

Contains verified engineering knowledge.

Examples

- Next.js
- Docker
- PostgreSQL
- Kubernetes
- Redis
- Prisma

Global Knowledge is versioned and updated with Daemon releases.

Users cannot modify it directly.

---

## 2. Project Knowledge

Created automatically.

Stored inside

.daemon/

Contains

- detected technologies
- execution history
- health history
- preferred startup order
- previous recoveries

This knowledge belongs only to the current project.

---

## 3. User Knowledge

Each developer works differently.

Daemon learns personal preferences.

Examples

- preferred package manager
- preferred terminal
- preferred formatter
- preferred ports
- favourite database
- preferred startup sequence

Stored globally.

Never shared.

---

## 4. Developer Knowledge

Reserved for Daemon contributors.

Allows extending Daemon itself.

Developer Knowledge becomes part of future Daemon releases after validation.
