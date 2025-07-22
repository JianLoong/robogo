# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Robogo is a simple, modern test automation framework written in Go. It provides a clean YAML-based DSL for writing test cases with support for HTTP APIs, databases (PostgreSQL, Spanner), messaging systems (Kafka, RabbitMQ), and more.

## Commands

### Build
```bash
go build -o robogo ./cmd/robogo
```

### Run Tests
```bash
# Run a single test
./robogo run <test-file.yaml>

# Run test with custom .env file
./robogo --env production.env run <test-file.yaml>

# List available actions
./robogo list

# Show version
./robogo version
```

### Development Environment Setup
```bash
# Start all services (PostgreSQL, Kafka, Spanner, HTTPBin)
docker-compose up -d

# Setup Spanner (run after docker-compose up)
# Linux/Mac:
SPANNER_EMULATOR_HOST=localhost:9010 ./setup-spanner.sh
# Windows:
.\setup-spanner.ps1
```

### Go Commands
```bash
# Standard Go commands work
go run ./cmd/robogo <command>
go test ./...
go mod tidy
```

## Architecture

### Current Architecture (Post-Simplification)

The codebase follows a clean, layered architecture with excellent principle adherence:

```
CLI → TestRunner → ExecutionStrategyRouter → Strategies → Actions
```

**Architecture Quality: 8.7/10** - Clean layers, strong KISS principle adherence, no dependency injection

### Core Components

- **CLI (`internal/cli.go`)**: Direct CLI implementation with no abstractions, handles `run`, `list`, and `version` commands
- **TestRunner (`internal/runner.go`)**: Core orchestrator, creates ExecutionStrategyRouter directly
- **ExecutionStrategyRouter (`internal/execution/strategy_router.go`)**: Priority-based strategy routing system
- **Execution Strategies (`internal/execution/`)**: Strategy pattern for different step types (conditional, retry, nested, basic)
- **Actions (`internal/actions/`)**: 19 action implementations with consistent signature pattern
- **Variables (`internal/common/variables.go`)**: Variable substitution using `${variable}` and `${ENV:VAR}` syntax
- **Types (`internal/types/`)**: Core data structures for tests, steps, and results
- **Constants (`internal/constants/`)**: Consolidated execution and configuration constants

### Action System

Actions follow a consistent function signature: `func(args []any, options map[string]any, vars *Variables) ActionResult`

Registered in `internal/actions/registry.go` with 19 actions across categories:
- **Core**: `assert`, `log`, `variable` (3)
- **HTTP**: `http` (supports GET, POST, PUT, DELETE, PATCH, HEAD) (1)
- **Database**: `postgres`, `spanner` (2)
- **Messaging**: `kafka`, `rabbitmq`, `swift_message` (3)
- **Data Processing**: `jq`, `xpath` (2)
- **JSON/XML**: `json_parse`, `json_build`, `xml_parse`, `xml_build` (4)
- **Encoding**: `base64_encode`, `base64_decode`, `url_encode`, `url_decode`, `hash` (5)
- **Utilities**: `uuid`, `time`, `sleep` (3)

### Execution Strategy System

Priority-based routing with 4 strategies:
1. **ConditionalExecutionStrategy** (Priority 4) - Handles `if` conditions
2. **RetryExecutionStrategy** (Priority 3) - Retry logic with configurable attempts
3. **NestedStepsExecutionStrategy** (Priority 2) - Nested step execution
4. **BasicExecutionStrategy** (Priority 1) - Fallback for standard actions

### Variable System

- Uses `${variable}` syntax for simple variable substitution
- Uses `${ENV:VARIABLE_NAME}` syntax for environment variable access
- For complex data extraction, use `jq` action for JSON/structured data or `xpath` action for XML
- Simple substitution engine replaces `${variable_name}` patterns
- Unresolved variables show warnings with hints to use `jq` for complex access

#### Environment Variables

Environment variables provide secure credential management:
```yaml
variables:
  vars:
    # Secure database connection using environment variables
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/${ENV:DB_NAME}?sslmode=disable"
    
    # API authentication
    api_token: "${ENV:API_TOKEN}"
    api_base_url: "${ENV:API_BASE_URL}"
```

Required environment variables can be set in multiple ways:

**Option 1: Using .env file (recommended)**
```bash
# Copy example file and edit with your values
cp .env.example .env

# Run test (automatically loads .env)
./robogo run examples/03-postgres-secure.yaml

# Or specify custom .env file
./robogo --env my-custom.env run examples/03-postgres-secure.yaml
```

**Option 2: Export environment variables**
```bash
export DB_USER=robogo_testuser
export DB_PASSWORD=robogo_testpass
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=robogo_testdb
./robogo run examples/03-postgres-secure.yaml
```

**Note:** Explicitly set environment variables take precedence over .env file values.

### Test Structure

Tests are defined in YAML with:
- `testcase`: Test name
- `description`: Optional description
- `variables`: Pre-defined variables
- `setup`: Optional setup steps (run before main steps)
- `steps`: Array of test steps with `name`, `action`, `args`, and optional `result`
- `teardown`: Optional teardown steps (always run, even if test fails)

### Connection Management

The framework follows a "immediate connection" pattern:
- Database and messaging connections open/close per operation
- No persistent connections or connection pooling
- Clean exit with no hanging processes

### Architecture Principles

- **Simple & Direct**: No over-engineering, abstractions, or dependency injection
- **CLI Tool Design**: Clean exit, no hanging processes, immediate connections
- **Minimal Dependencies**: Only essential libraries (no frameworks)
- **KISS Principle**: Keep it simple and straightforward - direct construction over abstractions
- **Strategy Pattern**: Priority-based execution routing for extensibility without complexity
- **Consistent Patterns**: All actions follow identical signature, error handling, and result patterns

### Architecture Quality Assessment

**Overall Score: 9.1/10**

| Area | Score | Status |
|------|--------|---------|
| Layer Organization | 9/10 | ✅ Clean separation, clear boundaries |
| Principle Adherence | 9/10 | ✅ Strong KISS, no DI, direct construction |
| Code Organization | 9/10 | ✅ Domain-driven packages, logical grouping |
| Action System | 9/10 | ✅ Consistent, extensible pattern |
| Variable System | 8/10 | ✅ Good but could enhance path resolution |
| Execution System | 9/10 | ✅ Flexible strategy pattern |
| Error Handling | 9/10 | ✅ Standardized patterns, consistent message access |

### Recent Architectural Improvements (2024)

**Phase 1: Architecture Simplification**
- ✅ **Eliminated dependency injection system** - Removed ExecutionPipeline, Dependencies, DependencyInjector
- ✅ **Simplified execution architecture** - Reduced from 6 layers to 2 clean layers  
- ✅ **Split large action files** - Improved maintainability (string.go 266→38 lines, xml.go 239→25 lines)
- ✅ **Consolidated constants** - Organized 6 files into 2 logical groups (execution.go, config.go)
- ✅ **Strategy priority normalization** - Clean 1,2,3,4 priority sequence

**Phase 2: Error Handling Standardization**
- ✅ **Unified execution strategy returns** - Single `*StepResult` return pattern, eliminated dual `(result, error)` 
- ✅ **Consistent error message access** - Removed `GetErrorMessage()` alias, standardized on `GetMessage()`
- ✅ **Complete error extraction** - Runner checks both `ErrorInfo` and `FailureInfo` with automatic conversion
- ✅ **Variable resolution validation** - Added `validateArgsResolved()` helper for critical actions (assert, http, postgres)
- ✅ **Predictable error boundaries** - Actions use `ActionResult`, orchestration uses Go `error`

## Development Services

When working with tests that require external services:

- **PostgreSQL**: `localhost:5432` (user: `robogo_testuser`, pass: `robogo_testpass`, db: `robogo_testdb`)
- **Kafka**: `localhost:9092`
- **Spanner Emulator**: `localhost:9010`
- **HTTPBin**: `localhost:8000`

## Testing

The project uses YAML-based integration tests in the `examples/` directory. There are no traditional Go unit tests - the framework is designed for end-to-end testing of external services.

Run example tests:
```bash
./robogo run examples/01-http-get.yaml
./robogo run examples/03-postgres-basic.yaml
```