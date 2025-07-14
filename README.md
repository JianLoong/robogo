# Robogo

A simple, modern test automation framework written in Go. Robogo provides a clean YAML-based DSL for writing test cases with support for HTTP APIs, databases, messaging systems, and more.

## Features

- **Simple YAML Tests**: Write tests in clean, readable YAML format
- **HTTP Testing**: Full HTTP client with JSON parsing and response validation
- **Database Support**: PostgreSQL and Google Cloud Spanner with immediate connections
- **Messaging**: Kafka and RabbitMQ operations with proper connection management
- **Variable Substitution**: Dynamic variables with dot notation support
- **Clean CLI Tool**: Immediate connection handling - no hanging processes
- **Formatted Output**: Clean markdown table summaries

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

# Run a test suite
./robogo run-suite my-suite.yaml

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
- `log` - Print messages
- `assert` - Verify conditions 
- `variable` - Set variables

### HTTP Actions  
- `http` - HTTP requests (GET, POST, PUT, DELETE, etc.)
  - Automatic JSON parsing available as `${response.json}`
  - Status code: `${response.status_code}`
  - Response body: `${response.body}`

### Database Actions
- `postgres` - PostgreSQL operations (query, execute)
- `spanner` - Google Cloud Spanner operations (query, execute)

### Messaging Actions
- `kafka` - Kafka publish/consume
- `rabbitmq` - RabbitMQ publish/consume

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
    
  - name: "Check user email"
    action: assert
    args: ["${response.json.email}", "==", "${user_email}"]
```

### Test Suite

```yaml
testsuite: "API Tests"

testcases:
  - examples/test-http-get.yaml
  - examples/test-database-basic.yaml
```

## Development Environment

### Prerequisites
- Go 1.21+
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
docker exec kafka kafka-topics.sh --create --topic test-topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

## Project Structure

```
robogo/
├── cmd/robogo/                  # CLI entry point
├── internal/
│   ├── actions/                 # Action implementations
│   │   ├── core.go             # assert, log, variable
│   │   ├── http.go             # HTTP operations
│   │   ├── database.go         # postgres, spanner
│   │   ├── messaging.go        # kafka, rabbitmq
│   │   └── registry.go         # action registry
│   ├── common/                  # shared types
│   │   └── variables.go        # variable management
│   ├── cli.go                   # CLI interface
│   ├── runner.go                # test execution
│   ├── parser.go                # YAML parsing
│   └── types.go                 # data structures
├── examples/                    # example tests
├── setup-spanner.sh            # Spanner setup script
├── setup-spanner.ps1           # Spanner setup (Windows)
└── docker-compose.yml         # development services
```

## Architecture Principles

- **Simple & Direct**: No over-engineering, interfaces, or dependency injection
- **Immediate Connections**: Database and messaging connections open/close per operation
- **CLI Tool Design**: Clean exit, no hanging processes
- **Minimal Dependencies**: Only essential libraries
- **KISS Principle**: Keep it simple and straightforward

## Example Tests

See `examples/` directory for:
- `test-http-get.yaml` - HTTP GET requests
- `test-http-post.yaml` - HTTP POST with JSON
- `test-database-basic.yaml` - PostgreSQL operations
- `test-spanner.yaml` - Spanner operations
- `test-kafka.yaml` - Kafka messaging

## License

MIT License - see LICENSE file for details.