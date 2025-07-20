# Robogo

A simple, modern test automation framework written in Go. Robogo provides a clean YAML-based DSL for writing test cases with support for HTTP APIs, databases, messaging systems, and more.

**Shift-Left Testing**: Robogo enables developers to run comprehensive end-to-end tests early in the development cycle with clear, readable test definitions that improve collaboration between development and testing teams.

## Features

- **Developer-Friendly**: Clear, readable YAML tests that developers can easily understand and maintain
- **Shift-Left Ready**: Run full end-to-end tests in development environments
- **Simple YAML Tests**: Write tests in clean, readable YAML format
- **HTTP Testing**: Full HTTP client with response validation and variable substitution
- **Database Support**: PostgreSQL and Google Cloud Spanner with immediate connections
- **Messaging**: Kafka and RabbitMQ operations with auto-commit support
- **Financial Messaging**: SWIFT message generation for banking and financial testing
- **JSON Construction**: Build complex JSON structures directly from YAML
- **Variable Substitution**: Dynamic variables with expression evaluation using `${variable}` syntax
- **Enhanced Assertions**: Support for multiple comparison operators and string matching
- **Retry Logic**: Built-in retry capabilities for handling eventual consistency and processing delays
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
    
  - name: "Extract status code"
    action: jq
    args: ["${response}", ".status_code"]
    result: status_code
    
  - name: "Verify response"
    action: assert
    args: ["${status_code}", "==", "200"]
    
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
  - Extract data with jq: `.status_code`, `.body`, `.headers`

### Database Actions
- **`postgres`** - PostgreSQL operations (query, execute)
- **`spanner`** - Google Cloud Spanner operations (query, execute)

### Messaging Actions
- **`kafka`** - Kafka publish/consume with auto-commit support
- **`rabbitmq`** - RabbitMQ publish/consume operations

### Financial Actions
- **`swift_message`** - Generate SWIFT financial messages from templates

### Data Processing Actions
- **`jq`** - Query and transform JSON/structured data using jq syntax
- **`xpath`** - Query XML documents using XPath expressions

### JSON/XML Actions
- **`json_build`** - Create JSON objects and arrays from nested YAML structures
- **`json_parse`** - Parse JSON strings into structured data
- **`xml_build`** - Create XML documents from structured data
- **`xml_parse`** - Parse XML strings into structured data

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
    
  - name: "Extract status code"
    action: jq
    args: ["${response}", ".status_code"]
    result: status_code
    
  - name: "Extract response body"
    action: jq
    args: ["${response}", ".body"]
    result: response_body
    
  - name: "Verify creation"
    action: assert
    args: ["${status_code}", "==", "201"]
    
  - name: "Check response contains email"
    action: assert
    args: ["${response_body}", "contains", "${user_email}"]
```

### JSON Construction

Create complex JSON structures directly from YAML:

```yaml
steps:
  - name: "Create user object"
    action: json_build
    args:
      - id: "${user_id}"
        name: "${user_name}"
        email: "${user_email}"
        active: true
        profile:
          age: 30
          preferences: ["email", "sms"]
    result: user_json
    
  - name: "Send JSON to API"
    action: http
    args: ["POST", "${api_url}/users", "${user_json}"]
    options:
      headers:
        Content-Type: "application/json"
    result: response
```

The `json_build` action automatically handles:
- Variable substitution in nested structures
- Proper JSON marshaling for HTTP requests
- Property access via dot notation (e.g., `${user_json.name}`)

### SWIFT Message Generation

Generate SWIFT financial messages using templates:

```yaml
steps:
  - name: "Generate MT103 message"
    action: swift_message
    args: ["mt103"]
    options:
      data:
        SenderBIC: "BANKGB2L"
        ReceiverBIC: "BANKUS33"
        TransactionRef: "TXN-${transaction_id}"
        BankOperationCode: "CRED"
        ValueDate: "240719"
        Currency: "USD"
        InterbankAmount: "1000.00"
        OrderingCustomer: "John Doe\\nMain Street 123"
        BeneficiaryCustomer: "Jane Smith\\nOak Avenue 456"
        DetailsOfCharges: "OUR"
    result: swift_msg
    
  - name: "Send to processing system"
    action: http
    args: ["POST", "${swift_endpoint}", "${swift_msg}"]
    options:
      headers:
        Content-Type: "text/plain"
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

### Retry Logic

Handle eventual consistency and processing delays with built-in retry capabilities:

```yaml
steps:
  # Basic retry with fixed delay
  - name: "Poll for completion"
    action: http
    args: ["GET", "${api_url}/status/${job_id}"]
    retry:
      attempts: 5
      delay: "2s"
      backoff: "fixed"
    result: job_status
    
  # Exponential backoff for quick failures
  - name: "Wait for database consistency"
    action: postgres
    args: ["query", "${db_url}", "SELECT status FROM orders WHERE id = ${order_id}"]
    retry:
      attempts: 6
      delay: "500ms"
      backoff: "exponential"  # 500ms, 1s, 2s, 4s, 8s, 16s
      retry_on: ["connection_error", "timeout"]
    result: order_status
    
  # Assertion retry for eventual consistency
  - name: "Verify processing completed"
    action: assert
    args: ["${job_status.json.state}", "==", "completed"]
    retry:
      attempts: 10
      delay: "3s"
      backoff: "linear"  # 3s, 6s, 9s, 12s...
      retry_on: ["assertion_failed"]
      stop_on_success: true
    result: completion_check
```

**Retry Configuration:**
- `attempts`: Number of total attempts (including first try)
- `delay`: Base delay between attempts (e.g., "1s", "500ms")
- `backoff`: Strategy - "fixed", "linear", or "exponential"
- `retry_on`: Specific error types - ["assertion_failed", "http_error", "timeout", "connection_error", "all"]
- `stop_on_success`: Stop retrying immediately on success (default: true)

### Variable Substitution & Data Extraction

Variables use simple `${variable}` substitution. For complex data extraction, use dedicated actions:

```yaml
variables:
  vars:
    api_url: "https://api.example.com"

steps:
  # Simple variable substitution
  - name: "Make request"
    action: http
    args: ["GET", "${api_url}/users"]
    result: response
    
  # Extract data with jq for JSON/structured data
  - name: "Extract user ID"
    action: jq
    args: ["${response}", ".body | fromjson | .user.id"]
    result: user_id
    
  # Extract data with xpath for XML
  - name: "Extract XML value"
    action: xpath
    args: ["${xml_response}", "//user[@id='1']/name/text()"]
    result: user_name
    
  - name: "Use extracted values"
    action: log
    args: ["User ID: ${user_id}, Name: ${user_name}"]
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
│   │   ├── swift.go            # SWIFT message generation
│   │   ├── json.go             # JSON construction utilities
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
├── templates/                   # Message templates
│   └── swift/                  # SWIFT message templates
│       └── mt103.txt           # MT103 credit transfer template
├── setup-spanner.sh            # Spanner setup script
├── setup-spanner.ps1           # Spanner setup (Windows)
└── docker-compose.yml         # Development services
```

## Architecture Principles

- **Developer-Centric Design**: Tests are written as clear, readable YAML that developers can easily understand and modify
- **Shift-Left Enablement**: Full end-to-end testing capabilities that run efficiently in development environments
- **Simple & Direct**: No over-engineering, interfaces, or dependency injection
- **Immediate Connections**: Database and messaging connections open/close per operation
- **CLI Tool Design**: Clean exit, no hanging processes
- **Minimal Dependencies**: Only essential libraries
- **KISS Principle**: Keep it simple and straightforward
- **Test Clarity**: Every test step is self-documenting with clear names and expected outcomes

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
- **`10-swift-mt103.yaml`** - SWIFT MT103 message generation
- **`11-json-build.yaml`** - Complex JSON construction and HTTP integration
- **`12-retry-scenarios.yaml`** - Retry functionality demonstrations
- **`13-retry-demo.yaml`** - Simple retry examples
- **`14-retry-with-failures.yaml`** - Retry with failure scenarios
- **`15-retry-success-demo.yaml`** - Retry timing and backoff strategies

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

# SWIFT and JSON features (no services required)
./robogo run examples/10-swift-mt103.yaml
./robogo run examples/11-json-build.yaml

# Retry functionality examples (no services required)
./robogo run examples/13-retry-demo.yaml
./robogo run examples/15-retry-success-demo.yaml
```

## Shift-Left Testing Benefits

Robogo enables **true shift-left testing** by allowing developers to:

### For Developers
- **Run Full E2E Tests Locally**: Complete integration tests with databases, messaging, and external APIs
- **Clear Test Intent**: YAML format makes test logic immediately visible and understandable
- **Fast Feedback**: Quick test execution with immediate connection handling
- **Easy Setup**: Simple environment setup for comprehensive testing

### For Teams
- **Improved Collaboration**: QA and developers can read and modify the same test definitions
- **Living Documentation**: Tests serve as executable specifications of system behavior
- **Early Bug Detection**: Catch integration issues before they reach staging environments
- **Reduced Testing Debt**: E2E tests written during development, not as an afterthought

### Example: Developer Workflow
```bash
# 1. Run relevant tests during development
./robogo run tests/user-registration-flow.yaml
./robogo run tests/payment-processing.yaml

# 2. Validate changes before commit
./robogo run tests/critical-paths.yaml
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