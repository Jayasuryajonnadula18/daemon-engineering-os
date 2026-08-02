# DAEMON CLI MASTER ARCHITECTURE SPECIFICATION (MVP)

## Vision

Daemon is an Engineering Operating System that discovers projects,
orchestrates developer tools, diagnoses issues, recovers safely,
verifies changes, and maintains Engineering Health.

## MVP Commands

-   daemon init
-   daemon start
-   daemon doctor
-   daemon recover
-   daemon verify
-   daemon health

## Core Engines

-   Discovery Engine
-   Planning Engine
-   Orchestration Engine
-   Health Engine
-   Recovery Engine
-   Confidence Engine
-   Optional AI Engine

## Project Structure

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

## Discovery Engine

Detect framework, runtime, package manager, Docker, Git, database,
environment and services.

## Planning Engine

Generate startup, diagnosis, recovery and verification plans.

## Orchestration Engine

Execute plans using adapters instead of directly calling shell commands.

## Confidence Rules

95-100% execute automatically. 80-94% ask for confirmation. 60-79%
recommend only. Below 60% do not execute.

## AI

Optional. Supports Ollama/OpenAI/Anthropic/Gemini. AI explains and
reasons but never executes commands.

## Development Principles

-   Strong typing
-   SOLID
-   Modular architecture
-   Dependency injection
-   Testable services
-   No business logic inside CLI commands

## Roadmap

Phase 1: init, start, doctor, recover, verify, health. Phase 2: plugins,
impact, ship, deploy. Phase 3: knowledge graph, AI planning, IDE.
