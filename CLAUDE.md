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
- **Actions (`internal/actions/`)**: 26 action implementations with consistent signature pattern
- **Variables (`internal/common/variables.go`)**: Variable substitution using `${variable}` and `${ENV:VAR}` syntax
- **Types (`internal/types/`)**: Core data structures for tests, steps, and results
- **Constants (`internal/constants/`)**: Consolidated execution and configuration constants

### Action System

Actions follow a consistent function signature: `func(args []any, options map[string]any, vars *Variables) ActionResult`

Registered in `internal/actions/action_registry.go` with 26 actions across categories:
- **Core**: `assert`, `log`, `variable` (3)
- **HTTP**: `http` (supports GET, POST, PUT, DELETE, PATCH, HEAD) (1)
- **Database**: `postgres`, `spanner` (2)
- **File Operations**: `file_read`, `scp` (2)
- **Messaging**: `kafka`, `rabbitmq`, `swift_message` (3)
- **Data Processing**: `jq`, `xpath` (2)
- **JSON/XML/CSV**: `json_parse`, `json_build`, `xml_parse`, `xml_build`, `csv_parse` (5)
- **String Operations**: `string_random`, `string_replace`, `string_format`, `string` (4)
- **Encoding**: `base64_encode`, `base64_decode`, `url_encode`, `url_decode`, `hash` (5)
- **Utilities**: `uuid`, `time`, `sleep`, `ping` (4)
- **Security/Validation**: `ssl_cert_check` (1)

### Execution Strategy System

Priority-based routing with 4 strategies:
1. **ConditionalExecutionStrategy** (Priority 4) - Handles `if` conditions
2. **RetryExecutionStrategy** (Priority 3) - Retry logic with configurable attempts
3. **NestedStepsExecutionStrategy** (Priority 2) - Nested step execution
4. **BasicExecutionStrategy** (Priority 1) - Fallback for standard actions

### Error Handling System

Robogo distinguishes between **Errors** and **Failures** for clear problem classification:

#### **Errors** (`ErrorInfo`) - Technical Problems
Technical issues that prevent proper execution:
- **Network connectivity problems** (timeouts, connection refused)
- **Database connection failures** (invalid credentials, server down)
- **Parse/serialization errors** (malformed JSON, invalid XML)
- **System resource issues** (file not found, permission denied)
- **Invalid configuration** (missing required parameters, bad URLs)

#### **Failures** (`FailureInfo`) - Logical Test Problems  
Expected execution that produces unexpected results:
- **Assertion failures** (expected 200, got 404)
- **Validation failures** (expected "success", got "error")
- **Business logic violations** (user already exists)
- **Data integrity issues** (missing required fields)

#### **Status Distinction**
Robogo provides four distinct step statuses:
- **PASS** ✅: Action completed successfully
- **SKIPPED** ⏭️: Step bypassed due to conditional logic (`if: false`)
- **ERROR** ❌: Technical problems (ErrorInfo) - infrastructure/system issues
- **FAIL** ❌: Logical problems (FailureInfo) - test expectations not met

#### **Unified Error Access**
Both error types are accessible through:
```yaml
# Both ErrorInfo and FailureInfo accessible via GetMessage()
result.GetMessage()  # Returns error or failure message
result.HasIssue()    # True for either errors or failures
```

**Runner Integration**: The TestRunner preserves the status distinction:
- Technical errors (ErrorInfo) result in ERROR status with system-focused messages
- Logical failures (FailureInfo) result in FAIL status with test-focused messages
- Both provide structured context for debugging and suggestions for resolution

### Variable System

- Uses `${variable}` syntax for simple variable substitution
- Uses `${ENV:VARIABLE_NAME}` syntax for environment variable access
- For complex data extraction, use `jq` action for JSON/structured data, `xpath` action for XML, or `csv` extract type for CSV data
- Simple substitution engine replaces `${variable_name}` patterns
- Unresolved variables show warnings with hints to use `jq` for complex access or `csv` extract for CSV data

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
./robogo run examples/03-database/03-postgres-secure.yaml

# Or specify custom .env file
./robogo --env my-custom.env run examples/03-database/03-postgres-secure.yaml
```

**Option 2: Export environment variables**
```bash
export DB_USER=robogo_testuser
export DB_PASSWORD=robogo_testpass
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=robogo_testdb
./robogo run examples/03-database/03-postgres-secure.yaml
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
- `no_log`: Optional step-level flag to suppress sensitive data logging

### Security Features

#### **`no_log` Sensitive Data Protection**

Robogo provides comprehensive sensitive data protection similar to Ansible's `no_log` directive:

**Step-Level Protection:**
```yaml
steps:
  - name: "Authenticate with API"
    action: http
    args: ["POST", "${auth_url}", '{"password": "${secret}"}']
    no_log: true  # 🔒 Suppress all logging for this step
    result: auth_response
```

**Custom Field Masking:**
```yaml
steps:
  - name: "Process user data"
    action: http
    args: ["POST", "/users", "${user_data}"]
    sensitive_fields: ["ssn", "credit_card", "phone"]  # Step-level custom fields to mask
    result: user_response
```

**Built-in Security Masking:**
- **Database connections**: Automatically masks `password=`, `pwd=`, etc. in connection strings
- **HTTP requests**: Masks sensitive fields in JSON bodies and headers
- **Message queues**: Masks credentials in broker connection strings
- **Assertions**: Protects sensitive comparison values
- **Log statements**: Masks sensitive data in log messages

**Security Benefits:**
- **Compliance Ready**: Meets SOC2, GDPR, PCI-DSS logging requirements
- **Developer Safe**: Prevents credential exposure in CI/CD logs
- **Enterprise Grade**: Granular control from complete suppression to field-level masking
- **Zero Config**: Sensible defaults with opt-in enhanced security

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

### Design Philosophy: Explicit Tests Over Loops

Robogo intentionally **does not support `for` and `while` loops** in test definitions. This design decision prioritizes **test clarity and maintainability** over code brevity.

**Rationale:**
- **Test purpose matters**: Behavioral tests should be explicit about what they're testing
- **Debugging clarity**: Named test steps are clearer than "step N failed in loop iteration M"
- **Living documentation**: Tests serve as executable specifications - loops obscure intent
- **Industry alignment**: Most YAML-based testing frameworks avoid complex control flow

**Supported control flow:**
- ✅ **Conditional execution**: `if` statements for branching logic
- ✅ **Retry logic**: Built-in retry mechanisms with configurable backoff
- ✅ **Nested steps**: Grouping related operations for organization
- ❌ **Loops**: Removed in favor of explicit, named test scenarios

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
- ✅ **Four-status system** - PASS, SKIPPED, ERROR (technical), FAIL (logical) with proper distinction
- ✅ **Structured error types** - ErrorInfo vs FailureInfo with rich context and suggestions
- ✅ **Variable resolution validation** - Added `validateArgsResolved()` helper for critical actions (assert, http, postgres)
- ✅ **Visual documentation** - Added [docs/error-failure-states-diagram.md](docs/error-failure-states-diagram.md) with mermaid diagrams

## Development Services

When working with tests that require external services:

- **PostgreSQL**: `localhost:5432` (user: `robogo_testuser`, pass: `robogo_testpass`, db: `robogo_testdb`)
- **Kafka**: `localhost:9092`
- **Spanner Emulator**: `localhost:9010`
- **HTTPBin**: `localhost:8000`

## Testing

The project uses YAML-based integration tests in the `examples/` directory with **51 comprehensive test examples**. There are no traditional Go unit tests - the framework is designed for end-to-end testing of external services.

### Quick Test Examples

**No setup required (HTTP-based):**
```bash
./robogo run examples/02-http/01-http-get.yaml         # Basic HTTP GET with jq extraction
./robogo run examples/02-http/02-http-post.yaml        # HTTP POST with JSON data
./robogo run examples/01-basics/00-util.yaml           # Utility actions (UUID, time, variables)
./robogo run examples/09-advanced/08-control-flow.yaml # Conditional execution (if statements)
./robogo run examples/11-network/26-ping-network-test.yaml # Network connectivity testing
./robogo run examples/11-network/34-ssl-cert-check.yaml    # SSL certificate validation
./robogo run examples/06-data-processing/35-csv-parsing.yaml # CSV data processing and extraction
```

**Requires docker-compose up -d:**
```bash
./robogo run examples/03-database/03-postgres-basic.yaml   # PostgreSQL database operations
./robogo run examples/04-messaging/05-kafka-basic.yaml     # Kafka producer/consumer
./robogo run examples/03-database/06-spanner-basic.yaml    # Google Cloud Spanner
./robogo run examples/05-files/23-scp-simple-test.yaml     # SSH/SCP file transfer
```

### Test Categories

- **HTTP Testing**: GET/POST requests, authentication, response validation (examples 01-02)
- **Database Operations**: PostgreSQL and Spanner queries, secure connections (examples 03-07)
- **Messaging Systems**: Kafka, RabbitMQ, SWIFT financial messaging (examples 05, 09-10)
- **File Operations**: Local file reading, secure SCP transfers (examples 23-25)
- **Security Features**: Environment variables, no-log mode, data masking (examples 17-20)
- **Control Flow**: Conditional logic, retry mechanisms, nested steps (examples 08, 13, 21)
- **Data Processing**: JSON/XML/CSV parsing, jq queries, string operations (examples 11-12, 14-16, 35)
- **Network Testing**: ICMP ping connectivity, SSL certificate validation (examples 26, 34)

### Development Testing

The examples serve as both **documentation** and **integration tests** for the framework itself:
- Each feature is demonstrated with working examples
- Examples progress from beginner to expert complexity
- Real-world scenarios with actual external services
- Security-conscious patterns with credential management

For the complete catalog with complexity levels and detailed descriptions, see **[examples/README.md](examples/README.md)**.