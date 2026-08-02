# DISCOVERY_ENGINE.md

**Daemon Discovery Engine Specification** Version: 1.0 (MVP Foundation)

------------------------------------------------------------------------

# Purpose

The Discovery Engine is Daemon's perception system.

Its responsibility is to understand a project before any planning or
execution occurs.

It answers questions such as:

-   What technologies are being used?
-   How is the project structured?
-   Which services are required?
-   What is currently running?
-   What is missing?
-   How confident are these findings?

The Discovery Engine never executes changes. It only observes, detects,
classifies and produces a structured Project Context.

------------------------------------------------------------------------

# Core Responsibilities

-   Repository scanning
-   Technology detection
-   Framework identification
-   Runtime detection
-   Package manager detection
-   Infrastructure discovery
-   Environment discovery
-   Service discovery
-   Dependency inference
-   Confidence scoring
-   Project Context generation

------------------------------------------------------------------------

# Inputs

-   Current working directory
-   Repository contents
-   Configuration files
-   Lock files
-   Running processes (optional)
-   Docker daemon state
-   Kubernetes context
-   User preferences
-   Engineering Graph

------------------------------------------------------------------------

# Outputs

## ProjectContext

``` yaml
project:
  name:
  root:
frameworks: []
languages: []
runtimes: []
packageManagers: []
databases: []
containers: []
cloud: []
services: []
environments: []
healthHints: []
confidence: 0-100
```

------------------------------------------------------------------------

# Discovery Pipeline

``` text
Working Directory
      │
Repository Scan
      │
File Fingerprinting
      │
Technology Detection
      │
Dependency Resolution
      │
Infrastructure Detection
      │
Environment Detection
      │
Confidence Scoring
      │
ProjectContext
      │
Graph Manager
```

------------------------------------------------------------------------

# Phase 1 -- Repository Scan

Identify:

-   Git repository
-   Monorepo or single project
-   Workspace managers
-   Root folders
-   Hidden configuration
-   Existing .daemon directory

Supported indicators:

-   .git
-   package.json
-   pnpm-workspace.yaml
-   turbo.json
-   nx.json
-   lerna.json

------------------------------------------------------------------------

# Phase 2 -- File Fingerprinting

Detect technologies using known fingerprints.

Examples

Next.js - next.config.js - next.config.ts - package.json dependency:
next

React - react dependency - src/App.tsx

Express - express dependency

Docker - Dockerfile - docker-compose.yml - compose.yaml

Kubernetes - deployment.yaml - service.yaml - ingress.yaml -
kustomization.yaml - helm charts

Prisma - prisma/schema.prisma

------------------------------------------------------------------------

# Phase 3 -- Runtime Detection

Supported runtimes

-   Node.js
-   Bun
-   Deno
-   Python (future)
-   Java (future)
-   .NET (future)

Verify:

-   Installed version
-   Supported version
-   Executable path

------------------------------------------------------------------------

# Phase 4 -- Package Manager Detection

Supported

-   npm
-   pnpm
-   yarn
-   bun

Priority

1.  Lock file
2.  packageManager field
3.  Historical project knowledge
4.  User preference

------------------------------------------------------------------------

# Phase 5 -- Infrastructure Discovery

Discover

-   Docker
-   Docker Compose
-   Kubernetes
-   Local databases
-   Redis
-   Message brokers

Capture:

-   Running state
-   Version
-   Reachability
-   Configuration

------------------------------------------------------------------------

# Phase 6 -- Environment Discovery

Detect

-   .env
-   .env.local
-   .env.development
-   .env.production

Validate:

-   Required variables
-   Missing variables
-   Duplicate variables
-   Invalid values

Never expose secret values.

------------------------------------------------------------------------

# Phase 7 -- Dependency Inference

Infer implicit technologies.

Example

Next.js → Node.js

Prisma → Database

Docker Compose → Multiple services

Inference rules come from the Engineering Graph.

------------------------------------------------------------------------

# Unknown Technology Handling

Unknown tools are classified.

Workflow

Unknown ↓ Collect evidence ↓ AI classification (optional) ↓ Developer
review ↓ Graph Manager

No automatic promotion into Global Knowledge.

------------------------------------------------------------------------

# Confidence Model

Confidence is based on:

-   fingerprint matches
-   multiple evidence sources
-   graph relationships
-   runtime validation
-   successful verification

Policy

95--100 : Certain

80--94 : Likely

60--79 : Partial

Below 60 : Unknown

------------------------------------------------------------------------

# Integration

Discovery Engine sends ProjectContext to:

-   Graph Manager
-   Reasoning Engine
-   Planning Engine
-   Health Engine

It never communicates directly with the Execution Engine.

------------------------------------------------------------------------

# Public APIs

-   discoverProject()
-   scanRepository()
-   detectTechnologies()
-   detectRuntime()
-   detectInfrastructure()
-   detectEnvironment()
-   inferDependencies()
-   calculateConfidence()
-   buildProjectContext()

------------------------------------------------------------------------

# Developer Commands

``` bash
daemon dev detect
daemon dev fingerprints
daemon dev scan
daemon dev context
```

Used to debug and extend discovery rules.

------------------------------------------------------------------------

# Design Principles

-   Read-only
-   Deterministic
-   Explainable
-   Fast
-   Modular adapters
-   Confidence driven
-   Extensible via Engineering Graph

------------------------------------------------------------------------

# Future Roadmap

-   Live file watching
-   IDE integration
-   Remote repository discovery
-   Cloud environment discovery
-   CI/CD pipeline discovery
-   Organization-wide project analysis

The Discovery Engine is the first stage of Daemon's intelligence. Every
downstream decision depends on the quality and confidence of the Project
Context it produces.
