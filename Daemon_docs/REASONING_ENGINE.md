# REASONING_ENGINE.md

**Daemon Reasoning Engine Specification** Version: 1.0 (MVP Foundation)

------------------------------------------------------------------------

# Purpose

The Reasoning Engine is Daemon's decision-making subsystem.

It does not execute commands. It does not discover technologies. It does
not modify the Engineering Graph.

Its responsibility is to determine **what should happen next** based on
developer intent, project context, engineering knowledge, and
confidence.

------------------------------------------------------------------------

# Philosophy

Daemon is built on three principles:

1.  Deterministic before AI.
2.  Explain every decision.
3.  Never act beyond confidence.

The Reasoning Engine transforms developer intent into safe, explainable
engineering decisions.

------------------------------------------------------------------------

# Inputs

The engine consumes:

-   Developer Intent
-   Project Context
-   Engineering Graph
-   User Knowledge
-   Project Knowledge
-   Engineering Health
-   Confidence Policies
-   Execution History

------------------------------------------------------------------------

# Outputs

The engine produces:

-   Decision Graph
-   Execution Plan Request
-   Recovery Recommendation
-   Health Recommendation
-   Confidence Score
-   Human-readable Explanation

------------------------------------------------------------------------

# Core Pipeline

``` text
Developer Command
        │
        ▼
Intent Resolution
        │
        ▼
Context Analysis
        │
        ▼
Graph Traversal
        │
        ▼
Constraint Analysis
        │
        ▼
Goal Decomposition
        │
        ▼
Risk Analysis
        │
        ▼
Confidence Evaluation
        │
        ▼
Decision
        │
        ▼
Planning Engine
```

------------------------------------------------------------------------

# Intent Resolution

The first task is to determine **what the developer actually wants**.

Examples:

`daemon start` → Prepare the local development environment.

`daemon doctor` → Diagnose engineering issues.

`daemon recover` → Repair verified problems.

Future examples:

-   Host a website
-   Add authentication
-   Deploy application
-   Connect a database

Intent becomes the root node for reasoning.

------------------------------------------------------------------------

# Context Analysis

The engine evaluates:

-   Detected technologies
-   Running services
-   Current environment
-   User preferences
-   Previous executions
-   Current health
-   Missing prerequisites

Reasoning always occurs in the current engineering context.

------------------------------------------------------------------------

# Constraint Analysis

Before planning, Daemon identifies constraints.

Examples:

-   Docker unavailable
-   Missing Node.js
-   Database offline
-   Port already in use
-   Unsupported framework
-   Insufficient confidence

Constraints influence every downstream decision.

------------------------------------------------------------------------

# Goal Decomposition

Large goals are broken into smaller tasks.

Example:

Intent: Host Website

↓

Required Tasks

-   Detect framework
-   Install dependencies
-   Build application
-   Start infrastructure
-   Verify health
-   Expose service

Each task can generate additional subtasks.

------------------------------------------------------------------------

# Decision Model

Every decision follows the same lifecycle.

Detect

↓

Understand

↓

Reason

↓

Evaluate

↓

Decide

↓

Explain

↓

Plan

↓

Execute (handled elsewhere)

------------------------------------------------------------------------

# Decision Types

## Deterministic

Uses explicit Engineering Graph rules.

Preferred mode.

## Contextual

Uses project-specific facts and history.

## AI-Assisted

Used only when deterministic reasoning cannot confidently resolve the
problem.

AI suggestions are advisory only.

------------------------------------------------------------------------

# Confidence Integration

Confidence determines behaviour.

95--100 - Execute automatically

80--94 - Ask for confirmation

60--79 - Recommend only

Below 60 - Refuse and explain

Confidence is derived from:

-   graph certainty
-   context completeness
-   historical success
-   rule coverage
-   AI agreement (optional)

------------------------------------------------------------------------

# Explainability

Every decision must answer:

-   Why was this chosen?
-   Which rules were used?
-   Which graph nodes participated?
-   What alternatives were rejected?
-   Why is the confidence score what it is?

Explainability is mandatory.

------------------------------------------------------------------------

# Interaction with Other Engines

Discovery Engine → Provides project facts.

Graph Manager → Supplies engineering knowledge.

Health Engine → Supplies current health.

Confidence Engine → Supplies execution policy.

Planning Engine ← Receives approved decisions.

Execution Engine ← Executes plans.

Recovery Engine ← Executes recovery plans.

AI Bridge ↔ Provides suggestions for uncertain scenarios.

------------------------------------------------------------------------

# Public Interfaces

Examples:

-   resolveIntent()
-   analyzeContext()
-   evaluateConstraints()
-   decomposeGoal()
-   buildDecision()
-   explainDecision()
-   generateRecommendation()

------------------------------------------------------------------------

# Failure Handling

If reasoning fails:

1.  Stop execution.
2.  Explain why.
3.  Recommend safe next steps.
4.  Escalate to AI only if enabled.
5.  Never fabricate certainty.

------------------------------------------------------------------------

# Design Principles

-   Deterministic first
-   Context-aware
-   AI-assisted, never AI-driven
-   Modular
-   Testable
-   Explainable
-   Observable
-   Extensible

------------------------------------------------------------------------

# Future Evolution

Future versions may add:

-   Multi-goal planning
-   Cost-aware reasoning
-   Security-aware reasoning
-   Architecture impact analysis
-   Cross-project learning
-   Organization-wide policy reasoning

The Reasoning Engine remains independent of any specific AI provider and
always treats the Engineering Graph as the primary source of truth.
