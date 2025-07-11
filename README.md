# Robogo

A modern, git-driven test automation framework written in Go. Robogo provides a YAML-based DSL for writing comprehensive test cases with support for HTTP APIs, databases, messaging systems, and more.

## Features

- **YAML-based Test Definition**: Write tests in simple, readable YAML format
- **HTTP Testing**: Full HTTP client with mTLS, custom headers, and response validation
- **Database Support**: PostgreSQL and Google Cloud Spanner integration
- **Messaging**: Kafka and RabbitMQ publish/consume operations
- **Control Flow**: Conditional execution, loops, and branching logic
- **Template System**: SWIFT message generation and custom templating
- **Parallel Execution**: Run tests and steps concurrently for better performance
- **Secret Management**: Secure handling of sensitive data with output masking
- **Clear Output**: Readable console output with colors and formatting

## Quick Start

### Installation

```bash
# Build it for your OS
go build -o robogo.exe ./cmd/robogo
```

### Basic Usage

```bash
# Run a single test file
./robogo.exe run my-test.robogo

# Run multiple tests in parallel
./robogo.exe run tests/*.robogo --parallel --max-concurrency 4

# Run a test file
./robogo.exe run test.robogo

# Run a test suite
./robogo.exe run-suite my-suite.robogo

# List all available actions
./robogo.exe list
```

### Your First Test

Create a file called `hello-world.robogo`:

```yaml
testcase: "Hello World API Test"
description: "A simple test to verify API health"

variables:
  vars:
    api_url: "https://jsonplaceholder.typicode.com"

steps:
  - name: "Check API health"
    action: http
    args: ["GET", "${api_url}"]
    result: response
    
  - name: "Verify response code"
    action: assert
    args: ["${response.status_code}", "==", "200"]
    
  - name: "Log success"
    action: log
    args: ["API is healthy!"]
```

Run it:
```bash
./robogo.exe run hello-world.robogo
```

## Test Structure

### Test Cases

Test cases are individual `.robogo` files that define a single test scenario:

```yaml
testcase: "User Registration Test"
description: "Test user registration flow"

variables:
  vars:
    base_url: "https://api.example.com"
    user_email: "test@example.com"
  secrets:
    api_key:
      file: "secrets/api_key.txt"
      mask_output: true

steps:
  - name: "Register new user"
    action: http
    args: ["POST", "${base_url}/users", '{"email": "${user_email}"}']
    options:
      headers:
        Authorization: "Bearer ${SECRETS.api_key}"
        Content-Type: "application/json"
    result: registration_response
    
  - name: "Verify registration success"
    action: assert
    args: ["${registration_response.status_code}", "==", "201"]
    
  - name: "Verify user data"
    action: assert
    args: ["${registration_response.body.email}", "==", "${user_email}"]
```

### Test Suites

Test suites run multiple test cases with shared setup and teardown:

```yaml
testsuite: "API Integration Tests"
description: "Complete API test suite"
parallel:
options:
  max_concurrency: 3

setup:
  - name: "Initialize test environment"
    action: log
    args: ["Starting test suite"]

testcases:

  - file: tests/core/test-assert.robogo
  - file: tests/core/test-control-flow.robogo
 
teardown:
  - name: "Cleanup test environment"
    action: log
    args: ["Test suite completed"]

```

## Examples

Examples can be found under tests and integration directories


## Development Environment

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for development services)

### Setting up Development Services

```bash
# Start all development services
docker-compose up -d

# This starts:
# - PostgreSQL (localhost:5432)
# - Kafka (localhost:9092)  
# - RabbitMQ (localhost:5672, management UI: localhost:15672)
# - Google Cloud Spanner Emulator (localhost:9010)
```

## Project Structure

```
robogo/
├── cmd/robogo/              # CLI application entry point
├── internal/
│   ├── actions/             # Built-in action implementations
│   │   ├── http.go          # HTTP operations
│   │   ├── postgres.go      # PostgreSQL database
│   │   ├── spanner.go       # Google Cloud Spanner
│   │   ├── kafka.go         # Kafka messaging
│   │   ├── rabbitmq.go      # RabbitMQ messaging
│   │   ├── control.go       # Control flow (if/for/while)
│   │   ├── template.go      # Template rendering
│   │   └── registry.go      # Action registry
│   ├── parser/              # YAML parsing and validation
│   ├── runner/              # Test execution engine
│   └── util/                # Utility functions
├── tests/                   # Test examples and integration tests
├── examples/                # Example test files
├── templates/               # SWIFT and SEPA message templates
└── docker-compose.yml      # Development services
```

## License

MIT License - see LICENSE file for details.