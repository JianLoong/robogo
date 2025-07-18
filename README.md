# Robogo

A simple, modern test automation framework written in Go. Robogo provides a clean YAML-based DSL for writing test cases with support for HTTP APIs, databases, messaging systems, and more.

## Features

- **Simple YAML Tests**: Write tests in clean, readable YAML format
- **HTTP Testing**: Full HTTP client with response validation and variable substitution
- **Database Support**: PostgreSQL and Google Cloud Spanner with immediate connections
- **Messaging**: Kafka and RabbitMQ operations with auto-commit support
- **Variable Substitution**: Dynamic variables with expression evaluation using `${variable}` syntax
- **Enhanced Assertions**: Support for multiple comparison operators and string matching
- **Clean CLI Tool**: Immediate connection handling - no hanging processes
- **Formatted Output**: Clean table summaries with execution details

## Quick Start

### Installation

```bash
# Build for your platform
go build -o robogo ./cmd/robogo
```

### Basic Usage

```bash
# Run a single test
./robogo run my-test.yaml

# List available actions
./robogo list

# Show version
./robogo version
```

### Your First Test

Create `hello-world.yaml`:

```yaml
testcase: "Hello World Test"
description: "A simple API test"

steps:
  - name: "Make HTTP request"
    action: http
    args: ["GET", "https://httpbin.org/json"]
    result: response
    
  - name: "Verify response"
    action: assert
    args: ["${response.status_code}", "==", "200"]
    
  - name: "Log success"
    action: log
    args: ["Test passed!"]
```

Run it:
```bash
./robogo run hello-world.yaml
```

## Available Actions

### Core Actions
- **`log`** - Print messages with immediate output
- **`assert`** - Verify conditions with multiple operators
- **`variable`** - Set and manage variables
- **`uuid`** - Generate UUID values
- **`time`** - Generate timestamps (RFC3339, Unix, custom formats)

### HTTP Actions  
- **`http`** - HTTP requests (GET, POST, PUT, DELETE, etc.)
  - Status code: `${response.status_code}`
  - Response body: `${response.body}`
  - Headers: `${response.headers}`

### Database Actions
- **`postgres`** - PostgreSQL operations (query, execute)
- **`spanner`** - Google Cloud Spanner operations (query, execute)

### Messaging Actions
- **`kafka`** - Kafka publish/consume with auto-commit support
- **`rabbitmq`** - RabbitMQ publish/consume operations

## Test Structure

### Basic Test Case

```yaml
testcase: "User API Test"
description: "Test user registration"

variables:
  vars:
    api_url: "https://api.example.com"
    user_email: "test@example.com"

steps:
  - name: "Create user"
    action: http
    args: ["POST", "${api_url}/users", '{"email": "${user_email}"}']
    result: response
    
  - name: "Verify creation"
    action: assert
    args: ["${response.status_code}", "==", "201"]
    
  - name: "Check response contains email"
    action: assert
    args: ["${response.body}", "contains", "${user_email}"]
```

### Advanced Assertions

Robogo supports multiple assertion operators:

```yaml
steps:
  # Equality
  - name: "Test equality"
    action: assert
    args: ["${value}", "==", "expected"]
    
  # Numeric comparisons
  - name: "Test greater than"
    action: assert
    args: ["${count}", ">", "0"]
    
  # String matching
  - name: "Test contains"
    action: assert
    args: ["${response.body}", "contains", "success"]
    
  # Boolean assertions
  - name: "Test boolean"
    action: assert
    args: ["${is_valid}"]  # Single boolean argument
```

**Supported operators**: `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`

### Variable Substitution

Variables support expression evaluation with dot notation:

```yaml
variables:
  vars:
    api_url: "https://api.example.com"
    timeout: 30

steps:
  - name: "Use nested data"
    action: log
    args: ["User ID: ${response.json.user.id}"]
    
  - name: "Use array access"
    action: log
    args: ["First item: ${response.json.items[0].name}"]
```

### Action Options

Many actions support options for additional configuration:

```yaml
steps:
  # HTTP with custom headers and timeout
  - name: "HTTP with options"
    action: http
    args: ["POST", "${api_url}/users", '{"name": "test"}']
    options:
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer ${token}"
      timeout: "10s"
    result: response
    
  # Kafka with auto-commit
  - name: "Consume with auto-commit"
    action: kafka
    args: ["consume", "localhost:9092", "test-topic"]
    options:
      auto_commit: true
      count: 5
      timeout: "30s"
    result: messages
```

## Development Environment

### Prerequisites
- Go 1.24+
- Docker & Docker Compose (for services)

### Development Services

```bash
# Start all services
docker-compose up -d

# Services available:
# - PostgreSQL: localhost:5432
# - Kafka: localhost:9092  
# - Spanner Emulator: localhost:9010
# - HTTPBin: localhost:8000
```

### Database Setup

**PostgreSQL** - Ready to use:
```yaml
- action: postgres
  args: ["query", "postgres://robogo_testuser:robogo_testpass@localhost:5432/robogo_testdb?sslmode=disable", "SELECT 1"]
```

**Spanner** - Run setup first:
```bash
# Linux/Mac
SPANNER_EMULATOR_HOST=localhost:9010 ./setup-spanner.sh

# Windows PowerShell  
.\setup-spanner.ps1
```

### Kafka Setup

```bash
# Create topic
docker exec kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic test-topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

## Project Structure

```
robogo/
├── cmd/robogo/                  # CLI entry point
├── internal/
│   ├── actions/                 # Action implementations
│   │   ├── assert.go           # Enhanced assertion logic
│   │   ├── http.go             # HTTP operations
│   │   ├── kafka.go            # Kafka with auto-commit
│   │   ├── postgres.go         # PostgreSQL operations
│   │   ├── spanner.go          # Spanner operations
│   │   ├── rabbitmq.go         # RabbitMQ operations
│   │   ├── log.go              # Logging action
│   │   ├── variable.go         # Variable management
│   │   ├── uuid.go             # UUID generation
│   │   ├── time.go             # Time/timestamp generation
│   │   └── registry.go         # Action registry
│   ├── common/                  # Shared utilities
│   │   └── variables.go        # Variable substitution with expr
│   ├── types/                   # Data structures
│   │   ├── action_result.go    # Action result types
│   │   ├── step.go             # Step definitions
│   │   └── testcase.go         # Test case structures
│   ├── cli.go                   # CLI interface
│   ├── runner.go                # Test execution
│   ├── parser.go                # YAML parsing
│   └── control_flow.go          # Control flow execution
├── examples/                    # Example tests
├── setup-spanner.sh            # Spanner setup script
├── setup-spanner.ps1           # Spanner setup (Windows)
└── docker-compose.yml         # Development services
```

## Architecture Principles

- **Simple & Direct**: No over-engineering, interfaces, or dependency injection
- **Immediate Connections**: Database and messaging connections open/close per operation
- **CLI Tool Design**: Clean exit, no hanging processes
- **Minimal Dependencies**: Only essential libraries
- **KISS Principle**: Keep it simple and straightforward

## Example Tests

The `examples/` directory contains comprehensive test examples organized by complexity:

### Basic Examples
- **`01-http-get.yaml`** - HTTP GET requests with assertions
- **`02-http-post.yaml`** - HTTP POST with string validation
- **`03-postgres-basic.yaml`** - PostgreSQL operations
- **`04-postgres-advanced.yaml`** - Advanced database verification
- **`05-kafka-basic.yaml`** - Kafka messaging with auto-commit

### Advanced Examples
- **`06-spanner-basic.yaml`** - Spanner operations
- **`07-spanner-advanced.yaml`** - Advanced Spanner verification
- **`08-control-flow.yaml`** - Control flow examples (if, for, while)
- **`09-e2e-integration.yaml`** - End-to-end integration test

### Running Examples

```bash
# HTTP testing (no services required)
./robogo run examples/01-http-get.yaml
./robogo run examples/02-http-post.yaml

# Database testing (requires docker-compose up -d)
./robogo run examples/03-postgres-basic.yaml
./robogo run examples/04-postgres-advanced.yaml

# Kafka testing (requires docker-compose up -d)
./robogo run examples/05-kafka-basic.yaml

# Spanner testing (requires docker-compose up -d + setup script)
./robogo run examples/06-spanner-basic.yaml
./robogo run examples/07-spanner-advanced.yaml

# Advanced features
./robogo run examples/08-control-flow.yaml
./robogo run examples/09-e2e-integration.yaml
```

## Advanced Features

### Kafka Auto-commit

```yaml
- name: "Consume with auto-commit"
  action: kafka
  args: ["consume", "localhost:9092", "test-topic"]
  options:
    auto_commit: true  # Prevents re-reading same messages
    count: 1
    timeout: "5s"
  result: messages
```

### Count Validation

```yaml
- name: "Test count validation"
  action: kafka
  args: ["consume", "localhost:9092", "test-topic"]
  options:
    count: 0  # Returns empty result immediately
    # count: -1  # Would fail with clear error message
  result: empty_result
```

### Timeout Configuration

```yaml
- name: "Quick timeout test"
  action: kafka
  args: ["consume", "localhost:9092", "non-existent-topic"]
  options:
    timeout: "3s"  # Fail fast instead of default 30s
  result: result
```

## Troubleshooting

### Common Issues

1. **Kafka timeout errors**: Make sure topics exist and Kafka is running
2. **Database connection errors**: Verify Docker services are running
3. **Spanner errors**: Run the setup script after starting the emulator
4. **Variable resolution errors**: Check `${variable}` syntax and variable names

### Debug Tips

- Use `log` actions to inspect variable values
- Check Docker service logs: `docker-compose logs <service>`
- Verify service connectivity before running tests
- Use shorter timeouts for faster feedback during development

## License

MIT License - see LICENSE file for details.