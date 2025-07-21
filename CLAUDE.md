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
- **Messaging**: `kafka`, `rabbitmq`, `swift_message`
- **Data Processing**: `jq`, `xpath`
- **JSON/XML**: `json_parse`, `json_build`, `xml_parse`, `xml_build`
- **Utilities**: `uuid`, `time`

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