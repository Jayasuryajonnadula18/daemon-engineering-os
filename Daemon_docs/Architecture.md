# Architecture.md

**Daemon CLI -- Software Architecture (MVP)**

## 1. Architecture Philosophy

Daemon is an **Engineering Orchestrator**, not an IDE, package manager,
or container platform.

It coordinates existing developer tools using deterministic engineering
logic.

    Developer Intent
          │
          ▼
     CLI Command
          │
          ▼
    Discovery Engine
          │
          ▼
    Project Context
          │
          ▼
    Planning Engine
          │
          ▼
    Execution Plan
          │
          ▼
    Confidence Engine
          │
          ▼
    Orchestration Engine
          │
          ▼
    Adapters
          │
          ▼
    Git / Docker / Node / PostgreSQL / Kubernetes

------------------------------------------------------------------------

# 2. Core Engines

## Discovery Engine

Purpose: Identify the engineering environment.

Responsibilities

-   Detect framework
-   Detect runtime
-   Detect package manager
-   Detect Docker
-   Detect databases
-   Detect Git
-   Detect services
-   Build ProjectContext

Output

``` ts
ProjectContext
```

------------------------------------------------------------------------

## Planning Engine

Input

ProjectContext

Output

``` ts
ExecutionPlan
```

Responsibilities

-   Startup order
-   Health checks
-   Recovery actions
-   Verification steps

------------------------------------------------------------------------

## Confidence Engine

Purpose

Every action receives a confidence score.

      Score Behaviour
  --------- -----------------------
    95--100 Execute automatically
     80--94 Ask confirmation
     60--79 Recommend
       \<60 Do not execute

------------------------------------------------------------------------

## Health Engine

Produces

-   Engineering Health Score
-   Issues
-   Warnings
-   Recommendations

------------------------------------------------------------------------

## Recovery Engine

Consumes diagnosis.

Produces recovery plan.

Never performs destructive operations automatically.

------------------------------------------------------------------------

## AI Engine (Optional)

Supported Providers

-   Ollama
-   OpenAI
-   Anthropic
-   Gemini

Rules

-   AI never executes commands.
-   AI never edits files directly.
-   AI reasons over ProjectContext.
-   Rule engine remains source of truth.

------------------------------------------------------------------------

# 3. Adapters

Every technology exposes a common interface.

``` ts
interface Adapter {
 detect(): Promise<boolean>;
 verify(): Promise<void>;
 health(): Promise<void>;
 recover(): Promise<void>;
}
```

Examples

-   GitAdapter
-   DockerAdapter
-   NodeAdapter
-   PackageManagerAdapter
-   PostgreSQLAdapter

------------------------------------------------------------------------

# 4. Folder Structure

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
 context.json
 diagnosis.json
 health.json
 execution-plan.json
```

------------------------------------------------------------------------

# 5. Project Context

Stores discovered engineering facts.

``` ts
interface ProjectContext {
 framework?: string;
 runtime?: string;
 packageManager?: string;
 docker: boolean;
 git: boolean;
 database?: string;
 services: string[];
 envFiles: string[];
}
```

------------------------------------------------------------------------

# 6. Execution Plan

Example

``` text
daemon start

↓

Verify Node

↓

Verify Package Manager

↓

Install Dependencies

↓

Start Docker

↓

Start Database

↓

Start Backend

↓

Start Frontend

↓

Run Health Checks
```

------------------------------------------------------------------------

# 7. Command Flow

## daemon doctor

    Discovery

    ↓

    Health Checks

    ↓

    Diagnosis

    ↓

    Recovery Plan

    ↓

    Engineering Health

## daemon recover

    Read Diagnosis

    ↓

    Generate Actions

    ↓

    Ask Confirmation

    ↓

    Execute

    ↓

    Verify

    ↓

    Update Health

------------------------------------------------------------------------

# 8. Design Rules

-   CLI commands contain no business logic.
-   Engines communicate through typed models.
-   Adapters isolate external tools.
-   New technologies are added by implementing adapters.
-   AI is replaceable.
-   Deterministic logic comes first.

------------------------------------------------------------------------

# 9. Future Extensions

-   Engineering Knowledge Graph
-   Plugin System
-   VS Code Extension
-   Cloud Sync
-   Daemon IDE

------------------------------------------------------------------------

# 10. MVP Boundary

Included

-   Discovery
-   Startup orchestration
-   Diagnosis
-   Recovery
-   Verification
-   Health

Excluded

-   Autonomous coding
-   Production deployments
-   Self-modifying AI
-   Multi-user cloud platform
