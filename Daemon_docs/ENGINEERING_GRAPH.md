# ENGINEERING_GRAPH.md

**Daemon Engineering Graph Specification (MVP)**

## Purpose

The Engineering Graph is Daemon's internal knowledge model describing
how technologies relate to one another. It powers discovery, planning,
startup orchestration, health analysis, recovery and recommendations.

Unlike an LLM, the graph is deterministic and versioned.

------------------------------------------------------------------------

# Core Concept

Each technology is represented as a **Technology Node**.

Nodes are connected through typed relationships.

    Next.js
       │
    requires
       ▼
    Node.js
       │
    uses
       ▼
    npm / pnpm
       │
    starts
       ▼
    Application

    Docker ──hosts──► PostgreSQL
             │
             └────hosts──► Redis

    Prisma ──depends_on──► PostgreSQL

------------------------------------------------------------------------

# Node Schema

``` yaml
id: nextjs
name: Next.js
category: framework
detect:
  files:
    - package.json
  package:
    - next
startup:
  command: pnpm dev
health:
  checks:
    - node
    - dependencies
relationships:
  - requires: node
  - uses: pnpm
confidence: 100
```

------------------------------------------------------------------------

# Relationship Types

-   requires
-   depends_on
-   hosts
-   communicates_with
-   generates
-   builds
-   deploys_to
-   starts_after
-   optional_with
-   incompatible_with

------------------------------------------------------------------------

# Discovery Pipeline

1.  Scan repository.
2.  Detect files.
3.  Detect packages.
4.  Detect containers.
5.  Detect services.
6.  Build graph.
7.  Score confidence.
8.  Produce ProjectContext.

------------------------------------------------------------------------

# Planning

The Planning Engine traverses the graph.

Example:

Project contains:

-   Next.js
-   Prisma
-   PostgreSQL
-   Docker

Generated startup order:

1.  Verify Node
2.  Verify pnpm
3.  Start Docker
4.  Wait for PostgreSQL
5.  Apply Prisma
6.  Start Next.js
7.  Run health checks

------------------------------------------------------------------------

# Health Graph

Every node exposes health signals.

Example

Node.js - installed - version - executable

Docker - daemon running - compose available

PostgreSQL - accepting connections - port open

Health Score is aggregated across all active nodes.

------------------------------------------------------------------------

# Recovery Graph

Each node defines recoverable actions.

Example

Docker

-   start service
-   restart containers

Node

-   install runtime
-   reinstall dependencies

Prisma

-   generate client
-   migrate database

Recovery never bypasses confidence rules.

------------------------------------------------------------------------

# Confidence

Every edge and action has a confidence score.

100: Known deterministic workflow.

80: Likely workflow. Ask user.

Below 60: Recommend only.

------------------------------------------------------------------------

# AI Integration

AI receives:

-   ProjectContext
-   Engineering Graph
-   Diagnosis

AI returns:

-   explanations
-   unknown technology mapping
-   recovery suggestions

AI never edits the graph directly.

------------------------------------------------------------------------

# Developer Mode

    daemon dev graph

Displays the graph.

    daemon dev tech add

Creates a new node.

    daemon dev tech edit

Updates relationships.

    daemon dev validate

Runs graph validation.

------------------------------------------------------------------------

# User Teaching

    daemon teach workflow

Creates project-specific preference edges without changing the global
graph.

Example:

Docker starts_before Backend

Backend starts_before Frontend

------------------------------------------------------------------------

# Storage

    knowledge/
      technologies/
        nextjs.yaml
        docker.yaml
        postgres.yaml

    graphs/
      engineering.graph.json

------------------------------------------------------------------------

# Design Principles

-   Deterministic first
-   Explainable relationships
-   Version-controlled knowledge
-   Extensible node schema
-   Plugin-friendly
-   AI augments but never owns the graph

------------------------------------------------------------------------

# Future Evolution

The Engineering Graph will evolve into the Daemon Knowledge Graph by
adding:

-   cloud providers
-   CI/CD
-   monitoring
-   infrastructure
-   security
-   architecture patterns
-   deployment targets
-   organizational best practices

The graph remains the primary reasoning substrate for Daemon, while AI
reasons over it rather than replacing it.

---

# Engineering Knowledge Graph Extension

## Philosophy

The Engineering Graph is not merely a dependency graph.

It is Daemon's internal representation of engineering knowledge.

Every project, technology, developer intent, engineering problem and solution is represented as a connected graph.

Daemon reasons over this graph instead of relying solely on an LLM.

The LLM augments the graph; it never replaces it.

---

# Knowledge Domains

The graph consists of multiple interconnected domains.

```
Engineering Graph

├── Intent Layer
├── Task Layer
├── Technology Layer
├── Environment Layer
├── Infrastructure Layer
├── Problem Layer
├── Solution Layer
├── Knowledge Layer
├── User Layer
└── Project Layer
```

Each domain owns a specific type of knowledge.

---

# Intent Nodes

Intent Nodes represent what the developer wants to achieve.

Examples

```
Start Development

Host Website

Deploy Application

Create REST API

Create Authentication

Connect Database

Run Tests

Publish Package

Scale Service

Monitor Application
```

Intent is always the root of planning.

Example

```
Intent

↓

Host Website

↓

Daemon builds execution plan

↓

Framework

↓

Container

↓

Database

↓

Reverse Proxy

↓

Cloud

↓

Deployment
```

---

# Task Nodes

Intent is converted into executable Tasks.

Example

```
Host Website

↓

Create Build

↓

Create Docker Image

↓

Start Container

↓

Verify Health

↓

Expose Port

↓

Verify Domain

↓

Deployment Complete
```

Task Nodes are deterministic.

Each task exposes:

- prerequisites
- outputs
- health checks
- recovery actions
- confidence score

---

# Technology Nodes

Technology Nodes represent engineering tools.

Categories

Framework

Language

Runtime

Package Manager

Database

Cache

Container

Cloud

Authentication

Storage

Monitoring

CI/CD

Messaging

Search

AI Provider

Examples

```
Next.js

Node.js

Docker

PostgreSQL

Redis

Prisma

Kubernetes

Nginx

AWS

Supabase

Firebase

RabbitMQ
```

Technology Nodes expose

```
Detection

Startup

Verification

Health

Recovery

Relationships
```

---

# Environment Nodes

Environment Nodes describe where execution occurs.

Examples

```
Local

Development

Testing

Staging

Production

CI

CD

Container

VM

Cloud
```

Each Environment changes execution plans.

Example

Local

↓

Docker Desktop

Production

↓

Kubernetes

---

# Infrastructure Nodes

Infrastructure represents external resources.

Examples

```
VM

Container

DNS

Load Balancer

Storage

Object Storage

Secrets

Ingress

Certificate

Firewall
```

Infrastructure is linked to deployment planning.

---

# Problem Nodes

Problems become first-class citizens.

Example

```
Port Already Used

Missing Dependency

Docker Not Running

Database Offline

Node Version Mismatch

Package Conflict

Permission Denied

Disk Full

Invalid Environment Variables

Container CrashLoop

Image Pull Failure
```

Problem Nodes expose

```
Symptoms

Detection Rules

Severity

Recovery Plans

Confidence

References
```

---

# Solution Nodes

Each problem maps to one or more solutions.

Example

```
Problem

↓

Port 3000 Occupied

↓

Solutions

Kill Process

Use Different Port

Stop Previous Server

Ask User
```

Solution Nodes never execute directly.

They generate Recovery Plans.

---

# User Nodes

User Nodes personalize Daemon.

Examples

Preferred Package Manager

Preferred Database

Preferred Startup Order

Favourite Cloud

Preferred IDE

Preferred Formatter

Preferred Shell

Preferred Ports

Stored inside

```
~/.daemon/user/
```

User Nodes never modify Global Knowledge.

---

# Project Nodes

Every project receives its own graph.

```
Project

↓

Detected Technologies

↓

Detected Problems

↓

Execution History

↓

Health History

↓

Recovery History

↓

Custom Workflow

↓

Project Memory
```

Stored inside

```
.daemon/
```

---

# Relationship Types

Relationships are typed.

```
requires

depends_on

uses

communicates_with

runs_inside

hosts

starts_after

starts_before

reads

writes

creates

generates

deploys_to

verifies

recovers

detects

configures

extends

inherits

optional_with

incompatible_with

replaces
```

Relationships are directional.

Example

```
Next.js

requires

Node.js

Node.js

uses

pnpm

pnpm

installs

Dependencies
```

---

# Engineering Context

Daemon never reasons using isolated nodes.

Instead it creates an Engineering Context.

Example

```
Project

↓

Framework

↓

Runtime

↓

Infrastructure

↓

Database

↓

Cloud

↓

Current Problems

↓

Developer Intent

↓

Execution Plan
```

Everything is evaluated together.

---

# Knowledge Layers

The graph is divided into four knowledge layers.

Global Knowledge

Official engineering knowledge shipped with Daemon.

Project Knowledge

Automatically generated for each project.

User Knowledge

Personal workflows and preferences.

Developer Knowledge

Technology definitions created by Daemon contributors.

Only Developer Knowledge can be promoted into Global Knowledge.

---

# Graph Traversal

Traversal always starts with Intent.

```
Intent

↓

Tasks

↓

Required Technologies

↓

Required Infrastructure

↓

Verification

↓

Health

↓

Recovery

↓

Completion
```

Traversal never begins with AI.

---

# Confidence Engine Integration

Every traversal returns a confidence score.

```
95-100

Execute

80-94

Ask Confirmation

60-79

Recommend

Below 60

Stop

Explain Why
```

No command bypasses confidence evaluation.

---

# AI Integration

The LLM receives

Engineering Context

Graph Snapshot

Diagnosis

Execution History

User Preferences

Project Memory

The LLM returns only

Explanations

Recommendations

Unknown Technology Classification

Recovery Suggestions

Learning Candidates

The LLM never mutates the graph.

All mutations pass through the Graph Manager.

---

# Graph Manager

The Graph Manager is the only component allowed to change the Engineering Graph.

Responsibilities

- validate nodes
- validate relationships
- version graph
- remove duplicates
- resolve conflicts
- approve updates
- maintain integrity

---

# Plugin Model

Every plugin contributes additional nodes.

Example

```
Prisma Plugin

↓

Technology Nodes

↓

Task Nodes

↓

Problem Nodes

↓

Recovery Nodes

↓

Relationships
```

No plugin modifies existing core nodes directly.

---

# Teaching & Learning Unit (TLU)

The TLU is responsible for incorporating new knowledge safely.

Sources

- User teaching (`daemon teach`)
- Developer teaching (`daemon dev teach`)
- Approved AI suggestions
- Plugin definitions
- Official documentation imports

Workflow

```
New Knowledge
      ↓
Validation
      ↓
Conflict Detection
      ↓
Confidence Evaluation
      ↓
Graph Manager Approval
      ↓
Knowledge Layer Update
```

---

# Long-Term Vision

The Engineering Graph should evolve into an Engineering Knowledge Platform capable of answering questions such as:

- "What do I need to host this application?"
- "Why is my database failing?"
- "Which technologies are missing?"
- "Can I replace Docker with Podman?"
- "If I migrate from Express to FastAPI, what changes?"
- "What is the safest deployment path?"
- "Which tools should I initialize for this project?"
- "What are the consequences of removing Redis?"

At this stage, Daemon is no longer executing commands—it is reasoning about engineering systems.

# GRAPH_MANAGER.md

# Graph Manager Specification

Version 1.0

---

# Purpose

The Graph Manager is the central authority responsible for maintaining, validating, querying and evolving Daemon's Engineering Knowledge Graph.

No component is allowed to modify the Engineering Graph directly.

Every modification passes through the Graph Manager.

The Graph Manager guarantees:

- consistency
- integrity
- versioning
- explainability
- confidence
- determinism

---

# Philosophy

Daemon should never "guess".

Daemon reasons.

Reasoning is impossible without structured knowledge.

The Graph Manager transforms thousands of isolated engineering facts into one coherent engineering model.

It becomes the operating kernel of Daemon.

---

# Responsibilities

The Graph Manager owns:

✓ Node Management

✓ Relationship Management

✓ Graph Validation

✓ Graph Traversal

✓ Knowledge Layers

✓ Version Control

✓ Conflict Resolution

✓ Confidence Evaluation

✓ Query Engine

✓ Graph Persistence

✓ Plugin Integration

✓ AI Bridge

✓ Teaching & Learning Unit

---

# High-Level Architecture

                Developer
                     │
                     ▼
                CLI Commands
                     │
                     ▼
             Discovery Engine
                     │
                     ▼
              Planning Engine
                     │
                     ▼
              Graph Manager
     ┌───────────────┼────────────────┐
     │               │                │
     ▼               ▼                ▼
 Graph Store   Traversal Engine   Validation Engine
     │               │                │
     ▼               ▼                ▼
 Knowledge      Confidence      Query Engine
 Manager          Engine
     │
     ▼
 AI Bridge

---

# Internal Components

The Graph Manager contains several independent subsystems.

## 1. Node Manager

Responsible for:

- creating nodes
- deleting nodes
- updating nodes
- validating node schema

Supported Nodes

Technology

Task

Intent

Problem

Solution

Environment

Infrastructure

Plugin

User

Project

Knowledge

---

## 2. Relationship Manager

Responsible for all graph edges.

Examples

requires

depends_on

uses

hosts

communicates_with

reads

writes

generates

creates

starts_after

starts_before

deploys_to

verifies

detects

recovers

Every relationship is directional.

Every relationship has confidence.

Every relationship is versioned.

---

## 3. Traversal Engine

The traversal engine answers engineering questions.

Example

Developer

"I want to host my website."

↓

Intent Node

↓

Task Nodes

↓

Technology Nodes

↓

Environment Nodes

↓

Infrastructure Nodes

↓

Execution Plan

Traversal never starts from technology.

Traversal always starts from intent.

---

Supported Traversals

Intent Traversal

Dependency Traversal

Recovery Traversal

Health Traversal

Impact Traversal

Deployment Traversal

Plugin Traversal

Knowledge Traversal

---

## 4. Validation Engine

Every graph update is validated.

Checks

Duplicate Nodes

Circular Dependencies

Broken References

Invalid Relationships

Schema Violations

Confidence Rules

Relationship Rules

If validation fails

↓

Reject Update

---

## 5. Confidence Engine

Every node

Every relationship

Every traversal

Every recovery

Every recommendation

has confidence.

Scale

95-100

Automatic

80-94

Confirmation

60-79

Recommendation

Below 60

Reject

---

## 6. Knowledge Manager

Responsible for maintaining knowledge layers.

Global

Project

User

Developer

Plugin

Knowledge updates never bypass this component.

---

Knowledge Flow

Developer

↓

Validation

↓

Review

↓

Version

↓

Knowledge Layer

↓

Graph

---

## 7. Query Engine

Allows engines to ask questions.

Examples

Find technologies required for Next.js.

Find startup order.

Find all databases.

Find recovery plans.

Find health checks.

Find deployment targets.

Find missing dependencies.

Find plugins supporting Kubernetes.

The Query Engine never mutates data.

---

## 8. Graph Store

Stores the graph.

Example

knowledge/

technologies/

docker.yaml

nextjs.yaml

postgres.yaml

tasks/

deploy.yaml

hosting.yaml

problems/

port-conflict.yaml

missing-node.yaml

solutions/

kill-process.yaml

relationships/

engineering.graph.json

---

# Node Schema

Every node follows a common schema.

```yaml
id:

type:

name:

version:

description:

confidence:

metadata:

relationships:

health:

recovery:

tags:
```

---

# Edge Schema

```yaml
source:

target:

relationship:

confidence:

conditions:

metadata:
```

---

# Graph Lifecycle

Create

↓

Validate

↓

Version

↓

Publish

↓

Traverse

↓

Update

↓

Archive

---

# Plugin Integration

Plugins never edit core knowledge.

Plugins contribute isolated subgraphs.

Example

Prisma Plugin

↓

Technology Nodes

↓

Task Nodes

↓

Problem Nodes

↓

Solution Nodes

↓

Relationships

↓

Merge

↓

Validation

↓

Approval

↓

Graph

---

# AI Bridge

The AI communicates only through the Graph Manager.

AI receives

Project Context

Engineering Graph

Health

History

Intent

AI returns

Explanation

Suggestions

Unknown Technologies

Possible Relationships

Confidence

The Graph Manager decides whether those suggestions become knowledge.

---

# Teaching & Learning Unit

The Graph Manager owns the Teaching & Learning Unit.

Developer

daemon dev teach

↓

Graph Manager

↓

Validation

↓

Conflict Detection

↓

Knowledge Layer

↓

Engineering Graph

User

daemon teach

↓

User Knowledge

↓

Project Knowledge

↓

Graph Manager

Global Knowledge never changes from user teaching.

---

# Graph APIs

Every engine uses the same API.

Examples

getNode()

getRelationship()

findDependencies()

findTasks()

findRecovery()

findHealthChecks()

findEnvironment()

calculateConfidence()

validateGraph()

traverse()

search()

updateNode()

updateRelationship()

promoteKnowledge()

---

# Events

Every graph change emits events.

Examples

NodeCreated

NodeUpdated

RelationshipAdded

TraversalStarted

TraversalCompleted

KnowledgePromoted

PluginInstalled

ConfidenceChanged

RecoveryGenerated

These events power future telemetry and extensions.

---

# Future Components

Impact Analyzer

Predictive Planner

Architecture Analyzer

Security Analyzer

Deployment Planner

Migration Planner

Cost Optimizer

Technology Recommendation Engine

These modules consume the same Engineering Graph through the Graph Manager.

---

# Design Principles

The Graph Manager is the single source of truth.

No component bypasses it.

No AI bypasses it.

No plugin bypasses it.

No user bypasses it.

Every decision is explainable.

Every change is versioned.

Every recommendation is traceable.

Every action is deterministic first, intelligent second.