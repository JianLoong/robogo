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

### Core Components

- **CLI (`internal/cli.go`)**: Direct CLI implementation with no abstractions, handles `run`, `list`, and `version` commands
- **TestRunner (`internal/runner.go`)**: Executes test cases, manages variables and control flow
- **Actions (`internal/actions/`)**: Action implementations registered in `registry.go`
- **Variables (`internal/common/variables.go`)**: Variable substitution using `${variable}` syntax with expr evaluation
- **Types (`internal/types/`)**: Core data structures for tests, steps, and results

### Action System

Actions are registered in `internal/actions/registry.go` and include:
- **Core**: `assert`, `log`, `variable`
- **HTTP**: `http` (GET, POST, PUT, DELETE, etc.)
- **Database**: `postgres`, `spanner`
- **Messaging**: `kafka`, `rabbitmq`
- **Utilities**: `uuid`, `time`

### Variable System

- Uses `${variable}` syntax for substitution
- Supports nested object access (e.g., `${response.json.field}`)
- Powered by `expr-lang/expr` for expression evaluation
- Unresolved variables show warnings and use `__UNRESOLVED__` marker

### Test Structure

Tests are defined in YAML with:
- `testcase`: Test name
- `description`: Optional description
- `variables`: Pre-defined variables
- `steps`: Array of test steps with `name`, `action`, `args`, and optional `result`

### Connection Management

The framework follows a "immediate connection" pattern:
- Database and messaging connections open/close per operation
- No persistent connections or connection pooling
- Clean exit with no hanging processes

### Architecture Principles

- **Simple & Direct**: No over-engineering, interfaces, or dependency injection
- **CLI Tool Design**: Clean exit, no hanging processes  
- **Minimal Dependencies**: Only essential libraries
- **KISS Principle**: Keep it simple and straightforward

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