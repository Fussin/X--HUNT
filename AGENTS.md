# Agent Instructions for SentinelX Repository

This document provides guidance for AI agents working on the SentinelX codebase.

## Project Overview

This is a monorepo for SentinelX, an AI-powered cybersecurity framework. The project is composed of multiple modules written in different languages. Refer to the `README.md` for the high-level project structure.

## Development Environment & Tooling

When working on a specific module, be sure to use the correct language and tools.

### Go (`core-proxy`, `distributed-mesh`, `web-ui/backend-api`, `plugin-sdk`)

-   **Language:** Go (version 1.21 or later).
-   **Dependency Management:** Go Modules (`go.mod`). Run `go mod tidy` after changing dependencies.
-   **Linting:** We use `golangci-lint`. The configuration is in the root `.golangci.yml`. Before submitting code, ensure `golangci-lint run ./...` passes from within the module's directory.
-   **Formatting:** Use standard `gofmt` or `goimports`.

### Python (`ai-hacker`, `vuln-scanner`, `poc-engine`, etc.)

-   **Language:** Python (version 3.8 or later).
-   **Dependency Management:** Use `pyproject.toml` for project metadata. For adding dependencies, use a tool like `poetry` or `pip-tools` if available, otherwise manually edit `pyproject.toml` and resolve dependencies as needed.
-   **Linting:** We use `ruff` for fast, comprehensive linting.
-   **Formatting:** We use `black` for consistent code formatting.
-   **Running checks:** From the root of any Python module, you can run `ruff check .` and `black --check .`.

### JavaScript/TypeScript (`web-ui/frontend`)

-   **Framework:** Next.js 14+ with App Router.
-   **Language:** TypeScript.
-   **Dependency Management:** `npm`. The project is located in `web-ui/frontend`. Run `npm install` from within that directory to install dependencies.
-   **Linting:** ESLint is configured. Run `npm run lint` from `web-ui/frontend`.
-   **Formatting:** Prettier is configured. Run `npm exec prettier -- --check .` from `web-ui/frontend` to check formatting.

## CI/CD

The CI pipeline is defined in `.github/workflows/ci.yml`. It automatically runs linting checks for all three languages on every push and pull request to the `main` branch. Ensure your changes pass these checks.

## Core Interfaces

-   **gRPC Message Bus:** The core communication protocol is defined in `protos/messagebus.proto`. If you need to add new message types for inter-module communication, modify this file and re-generate the Protobuf code for the relevant languages.
-   **Plugin SDK:** The interfaces for plugins are in `plugin-sdk/`. Refer to these when creating new plugins.

Your goal is to maintain the clean, modular architecture of this project. Always add new code to the appropriate module and respect the established coding standards.
