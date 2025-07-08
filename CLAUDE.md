# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Robogo is a modern, git-driven test automation framework written in Go, designed for comprehensive API testing, SWIFT message generation, database operations, and Test Data Management (TDM). It provides a developer-friendly YAML-based DSL for writing test cases and test suites.

## Key Features & Architecture

### Core Components
- **Test Engine**: Located in `internal/` with three main packages:
  - `parser/`: YAML parsing, test case validation, and test suite support
  - `runner/`: Test orchestration, execution engine, and parallel execution
  - `actions/`: Built-in actions for HTTP, database, control flow, templating, etc.
- **CLI Interface**: `cmd/robogo/main.go` - Cobra-based CLI with multiple output formats
- **VS Code Extension**: Complete extension in `.vscode/extensions/robogo/` with syntax highlighting, autocomplete, and validation

### Supported Test Types
- **Test Cases**: Individual `.robogo` files with YAML-based test definitions
- **Test Suites**: Collections of test cases with shared setup/teardown
- **Parallel Execution**: Both test-level and step-level parallelism with dependency analysis

### Built-in Actions
- **HTTP Operations**: `http`, `http_get`, `http_post` with mTLS support
- **Database**: `postgres` operations and `spanner` (Google Cloud Spanner)
- **Messaging**: `kafka` publish/consume and `rabbitmq` operations
- **Control Flow**: `if`, `for`, `while` loops with conditional logic
- **Data Management**: `tdm` (Test Data Management) with structured data sets
- **Templating**: `template` action for SWIFT messages and SEPA XML generation
- **Utilities**: `assert`, `log`, `get_time`, `get_random`, `concat`, `length`
- **Variables**: `variable` action for dynamic variable management
- **Secrets**: Secure credential handling with file-based secrets and output masking

## Development Commands

### Building and Running
```bash
# Build the main binary
go build -o robogo.exe ./cmd/robogo

# Run Go tests
go test ./...

# Run a single test file
./robogo.exe run tests/core/test-assert.robogo

# Run test suite
./robogo.exe run-suite examples/test-suite.robogo

# Run with parallel execution
./robogo.exe run tests/*.robogo --parallel --max-concurrency 4

# Run with different output formats
./robogo.exe run test.robogo --output json
./robogo.exe run test.robogo --output markdown
```

### Development Environment
```bash
# Start development services (Docker required)
docker-compose up -d

# This starts:
# - Kafka (localhost:9092)
# - PostgreSQL (localhost:5432)
# - RabbitMQ (localhost:5672, management: localhost:15672)
# - Google Cloud Spanner Emulator (localhost:9010, REST: localhost:9020)
```

### VS Code Extension Development
```bash
# Quick launch (Windows PowerShell)
./run-extension.ps1

# Manual build and launch
cd .vscode/extensions/robogo
npm install
npm run compile
code --new-window --extensionDevelopmentPath="$(pwd)/.vscode/extensions/robogo"
```

### Action Development
```bash
# List available actions
./robogo.exe list

# Get action completions for autocomplete
./robogo.exe completions get_random
```

## Project Structure

```
robogo/
├── cmd/robogo/                    # CLI entry point and main.go
├── internal/
│   ├── actions/                   # Built-in actions (HTTP, DB, control flow, etc.)
│   │   ├── http.go               # HTTP operations with mTLS support
│   │   ├── postgres.go           # PostgreSQL database operations
│   │   ├── spanner.go            # Google Cloud Spanner operations
│   │   ├── kafka.go              # Kafka publish/consume operations
│   │   ├── control.go            # Control flow (if, for, while)
│   │   ├── template.go           # Template rendering for SWIFT/SEPA
│   │   ├── tdm.go                # Test Data Management operations
│   │   └── ...                   # Other action implementations
│   ├── parser/                   # YAML parsing and validation
│   │   ├── parser.go             # Test case and suite parsing
│   │   ├── types.go              # Data structures and types
│   │   ├── parallel.go           # Parallel execution configuration
│   │   └── testsuite.go          # Test suite support
│   ├── runner/                   # Test execution engine
│   │   ├── runner.go             # Main test runner
│   │   ├── execution_engine.go   # Step execution logic
│   │   ├── testsuite_runner.go   # Test suite orchestration
│   │   └── variable_manager.go   # Variable and secret management
│   └── util/                     # Utility functions
├── tests/                        # Comprehensive test examples
│   ├── core/                     # Core functionality tests
│   ├── integration/              # Integration tests
│   ├── templates/                # Template-based tests (SWIFT, SEPA)
│   └── tdm/                      # Test Data Management tests
├── examples/                     # Basic examples and tutorials
├── templates/                    # SWIFT and SEPA message templates
│   ├── mt103.tmpl               # SWIFT MT103 Customer Transfer
│   ├── mt202.tmpl               # SWIFT MT202 Institution Transfer
│   ├── sepa-credit-transfer.xml.tmpl  # SEPA Credit Transfer XML
│   └── ...
├── .vscode/extensions/robogo/    # Complete VS Code extension
└── docker-compose.yml           # Development environment services
```

## Test File Format

Robogo uses YAML-based test definitions with `.robogo` extension:

### Test Case Structure
```yaml
testcase: "Test Name"
description: "Test description"

variables:
  vars:
    api_url: "https://api.example.com"
    timeout: 30
  secrets:
    api_key:
      file: "secret.txt"
      mask_output: true

steps:
  - name: "HTTP GET request"
    action: http_get
    args: ["${api_url}/users"]
    result: response
    
  - name: "Assert response"
    action: assert
    args: ["${response.status_code}", "==", "200"]
```

### Test Suite Structure
```yaml
testsuite: "Suite Name"
description: "Suite description"

setup:
  - name: "Setup step"
    action: log
    args: ["Setting up..."]

teardown:
  - name: "Cleanup step"
    action: log
    args: ["Cleaning up..."]

testcases:
  - tests/test1.robogo
  - tests/test2.robogo
```

## Key Implementation Details

### Variable Management
- Variables are managed through `internal/runner/variable_manager.go`
- Support for both regular variables and secrets
- Secrets can be loaded from files with automatic output masking
- Reserved `__robogo_steps` variable contains step execution history

### Parallel Execution
- Configured via `ParallelConfig` in `internal/parser/parallel.go`
- Supports test-level and step-level parallelism
- Automatic dependency analysis for step execution order
- Configurable concurrency limits

### Template System
- Go template engine with file-based and inline templates
- Pre-built templates for SWIFT messages (MT103, MT202, MT900, MT910)
- SEPA XML templates for payment processing
- Variable substitution and complex data structures

### Database Integration
- PostgreSQL support with connection pooling
- Google Cloud Spanner with emulator support
- Secure credential management with URL encoding
- Parameterized queries for security

### Action Registry
- Extensible action system in `internal/actions/registry.go`
- Each action implements the `Action` interface
- Support for autocomplete and documentation
- Built-in validation and error handling

## Development Guidelines

### Adding New Actions
1. Implement the `Action` interface in `internal/actions/`
2. Register the action in `registry.go`
3. Add comprehensive tests in `tests/`
4. Update VS Code extension completions

### Test Writing Best Practices
- Use descriptive test names and step names
- Leverage TDM for structured test data
- Implement proper error handling with assertions
- Use secrets for sensitive data
- Consider parallel execution for performance

### Common Patterns
- **SWIFT Message Testing**: Use template action with pre-built templates
- **API Testing**: Combine HTTP actions with assertions
- **Database Testing**: Use TDM with PostgreSQL actions
- **Control Flow**: Implement retry logic with while loops
- **Data Validation**: Use assert action with various operators

## Docker Services Configuration

The project includes a complete Docker Compose setup for development:
- **Kafka**: Message streaming (port 9092)
- **PostgreSQL**: Database testing (port 5432)
- **RabbitMQ**: Message queuing (port 5672, management: 15672)
- **Spanner Emulator**: Cloud Spanner testing (port 9010)

### Service Credentials
- **PostgreSQL**: `robogo_testuser` / `robogo_testpass` / `robogo_testdb`
- **RabbitMQ**: `robogo_user` / `robogo_pass`
- **Kafka**: No authentication (development only)
- **Spanner**: Uses emulator (no authentication required)

## VS Code Extension Features

- **Syntax Highlighting**: Custom grammar for .robogo files
- **Autocomplete**: Intelligent action and parameter suggestions
- **Validation**: Real-time syntax and semantic validation
- **Documentation**: Hover tooltips with action documentation
- **Execution**: One-click test execution with integrated output
- **Debugging**: Step-by-step execution with detailed results

## Output Formats

- **Console**: Human-readable with colors and formatting
- **JSON**: Machine-readable for CI/CD integration
- **Markdown**: Documentation-friendly with collapsible sections
- **Step-level reporting**: Detailed execution metrics and timing