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
    api_url: "https://httpbin.org"

steps:
  - name: "Check API health"
    action: http
    args: ["GET", "${api_url}/status/200"]
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

setup:
  - name: "Initialize test environment"
    action: log
    args: ["Starting test suite"]

teardown:
  - name: "Cleanup test environment"
    action: log
    args: ["Test suite completed"]

testcases:
  - tests/auth-test.robogo
  - tests/user-test.robogo
  - tests/product-test.robogo

parallel:
  enabled: true
  max_concurrency: 3
```

## Available Actions

### HTTP Operations

```yaml
# GET request
- name: "Fetch user data"
  action: http
  args: ["GET", "https://api.example.com/users/123"]
  result: user_data

# POST with body and headers
- name: "Create user"
  action: http
  args: ["POST", "https://api.example.com/users", '{"name": "John"}']
  options:
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer ${token}"
  result: create_response
```

### Database Operations

```yaml
# PostgreSQL
- name: "Connect to database"
  action: postgres
  args: ["connect", "postgres://user:pass@localhost/db", "main"]

- name: "Query users"
  action: postgres
  args: ["query", "SELECT * FROM users WHERE id = $1", "123", "main"]
  result: users

# Google Cloud Spanner
- name: "Connect to Spanner"
  action: spanner
  args: ["connect", "projects/my-project/instances/my-instance/databases/my-db", "main"]

- name: "Query Spanner"
  action: spanner
  args: ["query", "SELECT * FROM Users WHERE UserId = @userId", "123", "main"]
  result: spanner_users
```

### Messaging

```yaml
# Kafka
- name: "Publish to Kafka"
  action: kafka
  args: ["publish", "localhost:9092", "test-topic", "Hello Kafka"]

- name: "Consume from Kafka"
  action: kafka
  args: ["consume", "localhost:9092", "test-topic", "30s"]
  result: kafka_message

# RabbitMQ
- name: "Connect to RabbitMQ"
  action: rabbitmq
  args: ["connect", "amqp://guest:guest@localhost:5672/", "main"]

- name: "Publish message"
  action: rabbitmq
  args: ["publish", "main", "test-exchange", "test.route", "Hello RabbitMQ"]
```

### Control Flow

```yaml
# Conditional execution
- name: "Check if user exists"
  action: if
  args: ["${response.status_code}", "==", "200"]
  then:
    - name: "User exists"
      action: log
      args: ["User found"]
  else:
    - name: "User not found"
      action: log
      args: ["User does not exist"]

# Loop over collection
- name: "Process users"
  action: for
  args: ["user", "${users}"]
  steps:
    - name: "Process user"
      action: log
      args: ["Processing user: ${user.name}"]

# While loop
- name: "Retry until success"
  action: while
  args: ["${attempts}", "<", "3"]
  steps:
    - name: "Attempt request"
      action: http
      args: ["GET", "https://api.example.com/data"]
      result: response
    - name: "Increment attempts"
      action: variable
      args: ["set", "attempts", "${attempts + 1}"]
```

### Templates

```yaml
# Generate SWIFT MT103 message
- name: "Generate payment message"
  action: template
  args: ["templates/mt103.tmpl", {
    "TransactionID": "TXN123456",
    "Amount": "1000.00",
    "Currency": "USD",
    "Sender": {
      "BIC": "BANKUSXX",
      "Name": "Test Bank"
    },
    "Beneficiary": {
      "Account": "987654321",
      "Name": "John Doe"
    }
  }]
  result: swift_message
```

### Validation & Utilities

```yaml
# Assertions
- name: "Validate response"
  action: assert
  args: ["${response.status_code}", "==", "200", "Expected successful response"]

- name: "Check response contains data"
  action: assert
  args: ["${response.body}", "contains", "user_id"]

# Logging
- name: "Log debug info"
  action: log
  args: ["Processing user ID: ${user_id}", "debug"]

# Variable management
- name: "Set dynamic variable"
  action: variable
  args: ["set", "current_time", "${get_time}"]
```

## Advanced Features

### Secret Management

```yaml
variables:
  secrets:
    database_url:
      file: "secrets/db_url.txt"
      mask_output: true
    api_token:
      value: "secret-token-here"
      mask_output: true

steps:
  - name: "Connect with secret"
    action: postgres
    args: ["connect", "${SECRETS.database_url}", "main"]
```

### Parallel Execution

```yaml
# Enable parallel execution
parallel:
  enabled: true
  max_concurrency: 4
  test_cases: true  # Run test cases in parallel
  steps: true       # Run steps in parallel where possible

# Individual step can override
- name: "Parallel HTTP requests"
  action: for
  args: ["url", "${urls}"]
  parallel: true
  steps:
    - name: "Fetch ${url}"
      action: http
      args: ["GET", "${url}"]
```

### Retry Logic

```yaml
- name: "Retry on failure"
  action: http
  args: ["GET", "https://api.example.com/data"]
  retry:
    attempts: 3
    delay: "1s"
    exponential_backoff: true
```

### Conditional Steps

```yaml
- name: "Only run in production"
  action: log
  args: ["Production deployment detected"]
  if: "${environment} == 'production'"
```

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

### Building and Testing

```bash
# Build the binary
go build -o robogo.exe ./cmd/robogo

# Run Go unit tests
go test ./...

# Run integration tests
./robogo.exe run tests/integration/*.robogo

# Run with debug output
./robogo.exe run test.robogo --debug-vars
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


## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

MIT License - see LICENSE file for details.