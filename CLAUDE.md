# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Robogo is a simple, modern test automation framework written in Go, designed for API testing, database operations, and messaging system verification. It provides a clean YAML-based DSL for writing test cases and test suites.

## Key Features & Architecture

### Core Components
- **Test Engine**: Located in `internal/` with simplified, direct architecture:
  - `actions/`: Built-in actions for HTTP, database, messaging, core operations
  - `common/`: Shared types and variable management
  - `cli.go`: Simple CLI interface with direct command handling
  - `runner.go`: Direct test execution without abstraction layers
  - `parser.go`: YAML parsing and test case validation
  - `types.go`: Core data structures

### Architecture Highlights
- **Simple & Direct**: No interfaces, dependency injection, or over-engineering
- **Immediate Connections**: Database and messaging connections open/close per operation
- **CLI Tool Design**: Clean exit, no hanging processes or persistent connections
- **KISS Principle**: Keep it simple and straightforward
- **Minimal Dependencies**: Only essential libraries

### Supported Test Types
- **Test Cases**: Individual `.yaml` files with YAML-based test definitions
- **Test Suites**: Collections of test cases with shared setup/teardown

### Built-in Actions
- **HTTP Operations**: `http` with automatic JSON parsing (GET, POST, PUT, DELETE, etc.)
- **Database**: `postgres` and `spanner` operations with immediate connection management
- **Messaging**: `kafka` and `rabbitmq` publish/consume operations
- **Core Utilities**: `assert`, `log`, `variable` for basic test operations

## Development Commands

### Building and Running
```bash
# Build the main binary
go build -o robogo ./cmd/robogo

# Run a single test file
./robogo run test.yaml

# Run test suite
./robogo run-suite suite.yaml

# List available actions
./robogo list

# Show version
./robogo version
```

### Development Environment
```bash
# Start development services (Docker required)
docker-compose up -d

# This starts:
# - PostgreSQL (localhost:5432) - ready to use
# - Kafka (localhost:9092) - needs topic creation
# - Spanner Emulator (localhost:9010) - needs setup
# - HTTPBin (localhost:8000) - ready to use
```

### Service Setup
```bash
# Spanner setup
SPANNER_EMULATOR_HOST=localhost:9010 ./setup-spanner.sh

# Kafka topic creation
docker exec kafka kafka-topics.sh --create --topic test-topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

## Project Structure

```
robogo/
├── cmd/robogo/                    # CLI entry point and main.go
├── internal/
│   ├── actions/                   # Action implementations
│   │   ├── core.go               # assert, log, variable actions
│   │   ├── http.go               # HTTP operations with JSON parsing
│   │   ├── database.go           # PostgreSQL and Spanner operations
│   │   ├── messaging.go          # Kafka and RabbitMQ operations
│   │   └── registry.go           # Action registration and lookup
│   ├── common/                   # Shared types
│   │   └── variables.go          # Variable management with dot notation
│   ├── cli.go                    # Simple CLI interface
│   ├── runner.go                 # Direct test execution
│   ├── parser.go                 # YAML parsing and validation
│   └── types.go                  # Core data structures
├── examples/                     # Example test files
├── setup-spanner.sh             # Spanner emulator setup (Linux/Mac)
├── setup-spanner.ps1            # Spanner emulator setup (Windows)
└── docker-compose.yml           # Development environment services
```

## Test File Format

Robogo uses YAML-based test definitions with `.yaml` extension:

### Test Case Structure
```yaml
testcase: "Test Name"
description: "Test description"

variables:
  vars:
    api_url: "https://api.example.com"
    timeout: 30

steps:
  - name: "HTTP GET request"
    action: http
    args: ["GET", "${api_url}/users"]
    result: response
    
  - name: "Assert response"
    action: assert
    args: ["${response.status_code}", "==", "200"]
    
  - name: "Check JSON data"
    action: assert
    args: ["${response.json.data[0].id}", "==", "1"]
```

### Test Suite Structure
```yaml
testsuite: "Suite Name"

testcases:
  - examples/test-http-get.yaml
  - examples/test-database-basic.yaml
```

## Key Implementation Details

### Simple Architecture
- **No Interfaces**: Direct function calls instead of interface abstractions
- **No Dependency Injection**: Services created directly when needed
- **Direct Execution**: TestRunner struct with simple methods
- **Immediate Connections**: All database/messaging connections open/close per operation

### Variable Management
- **Simple Variables struct**: `map[string]interface{}` storage
- **Dot notation support**: Access nested properties like `${response.json.data[0].id}`
- **Variable substitution**: Template-style replacement with `${variable_name}`

### Connection Management (Key Principle)
- **PostgreSQL**: Opens connection, executes query, closes immediately
- **Spanner**: Creates client, executes operation, closes client immediately  
- **Kafka**: Creates writer/reader, sends/receives message, closes immediately
- **RabbitMQ**: Dials connection, creates channel, operates, closes immediately
- **No persistent pools**: Prevents hanging processes for CLI tool design

### Action Registry
- **Simple map**: `map[string]ActionFunc` for action lookup
- **ActionFunc signature**: `func(args []interface{}, options map[string]interface{}, vars *Variables) (interface{}, error)`
- **Direct registration**: Actions registered in `registry.go`

## Development Guidelines

### Architecture Principles
- **Keep It Simple**: Avoid over-engineering, interfaces, or complex patterns
- **Direct Approach**: Use direct function calls and simple structs
- **CLI Tool Design**: Ensure clean exit and no hanging processes
- **Immediate Resources**: Open/use/close resources immediately

### Adding New Actions
1. Implement `ActionFunc` signature in appropriate file (`core.go`, `http.go`, etc.)
2. Register action in `registry.go`
3. Follow immediate connection pattern for external resources
4. Add example test to `examples/` directory

### Connection Management Pattern
```go
// CORRECT: Immediate connection pattern
func myDatabaseAction(args []interface{}, options map[string]interface{}, vars *Variables) (interface{}, error) {
    // Open connection
    db, err := sql.Open("driver", connectionString)
    if err != nil {
        return nil, err
    }
    defer db.Close() // Always close immediately
    
    // Use connection
    result, err := db.Query("SELECT 1")
    if err != nil {
        return nil, err
    }
    defer result.Close()
    
    // Return result
    return data, nil
}
```

### Testing Best Practices
- Use descriptive test and step names
- Leverage variable substitution for reusability
- Use `${response.json}` for JSON response data access
- Test connection management with timeout to ensure clean exit

### Output Format
- **Step output**: Simple PASSED/FAILED per step
- **Summary**: Clean markdown table with test results
- **No emojis**: Plain text output for universal compatibility

## Service Credentials

### PostgreSQL (Ready to use)
- **Connection**: `postgres://robogo_testuser:robogo_testpass@localhost:5432/robogo_testdb?sslmode=disable`
- **Usage**: Direct in test files

### Spanner (Requires setup)
- **Setup**: Run `./setup-spanner.sh` or `.\setup-spanner.ps1`
- **Connection**: `projects/test-project/instances/test-instance/databases/test-database`
- **Environment**: Set `SPANNER_EMULATOR_HOST=localhost:9010`

### Kafka (Requires topic creation)
- **Broker**: `localhost:9092`
- **Setup**: Create topics using docker exec commands
- **Usage**: Direct in test files

## Common Patterns

### HTTP Testing
```yaml
- name: "API call"
  action: http
  args: ["POST", "https://api.example.com/users", '{"name": "test"}']
  result: response

- name: "Check status"
  action: assert
  args: ["${response.status_code}", "==", "201"]

- name: "Check JSON response"
  action: assert
  args: ["${response.json.user.name}", "==", "test"]
```

### Database Testing
```yaml
- name: "Query database"
  action: postgres
  args: ["query", "postgres://user:pass@localhost:5432/db", "SELECT * FROM users"]
  result: db_result

- name: "Check result count"
  action: assert
  args: ["${db_result.rows|length}", ">", "0"]
```

### Messaging Testing
```yaml
- name: "Publish message"
  action: kafka
  args: ["publish", "localhost:9092", "test-topic", "Hello World"]
  result: publish_result

- name: "Consume message"
  action: kafka
  args: ["consume", "localhost:9092", "test-topic"]
  result: message

- name: "Verify message"
  action: assert
  args: ["${message.message}", "==", "Hello World"]
```

## Important Reminders

- **No emojis**: Never add emojis to code or output
- **No legacy code**: Remove old patterns, don't keep deprecated functions
- **No quick fixes**: Write proper, clean solutions
- **Resource management**: Always address potential resource leaks
- **Binary naming**: Use only `robogo` as binary name
- **Simple approach**: Follow Linux philosophy - do one thing and do it well