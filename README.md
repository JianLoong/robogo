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

# Run test with custom .env file
./robogo --env production.env run my-test.yaml

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

## Environment Variables & Secret Management

### Using Environment Variables

Robogo supports environment variables for secure credential management using `${ENV:VARIABLE_NAME}` syntax:

```yaml
variables:
  vars:
    # Secure database connection using environment variables
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/${ENV:DB_NAME}?sslmode=disable"
    
    # API authentication
    api_token: "${ENV:API_TOKEN}"
    api_base_url: "${ENV:API_BASE_URL}"
```

### .env File Support

**Option 1: Default .env file (recommended)**
```bash
# Copy example and edit with your values
cp .env.example .env

# Run test (automatically loads .env)
./robogo run examples/03-postgres-secure.yaml
```

**Option 2: Custom .env file**
```bash
# Specify custom .env file
./robogo --env production.env run my-test.yaml
```

**Option 3: Export environment variables**
```bash
export DB_USER=myuser
export DB_PASSWORD=mypassword
./robogo run my-test.yaml
```

**Note:** Explicitly set environment variables take precedence over .env file values.

### Secret Management Philosophy

**Robogo consumes secrets but does not manage them.** This design principle keeps the framework focused on test automation while allowing seamless integration with any secret management approach:

**✅ Robogo's Responsibility:**
- Execute tests efficiently
- Provide clean variable substitution
- Support standard patterns (env vars, .env files)

**✅ Your Responsibility (choose your approach):**
- **Development**: Use `.env` files for local testing
- **CI/CD**: Inject secrets as environment variables
- **Production**: Integrate with your secret management system

**Integration Examples:**
```bash
# HashiCorp Vault
eval $(vault kv get -format=json secret/robogo | jq -r '.data.data | to_entries[] | "export \(.key)=\(.value)"')
./robogo run test.yaml

# Kubernetes secrets
export DB_PASSWORD=$(kubectl get secret db-creds -o jsonpath='{.data.password}' | base64 -d)
./robogo run test.yaml

# AWS Secrets Manager
export API_KEY=$(aws secretsmanager get-secret-value --secret-id prod/api-key --query SecretString --output text)
./robogo run test.yaml

# Development with .env
./robogo --env .env.local run test.yaml
```

## Available Actions

### Core Actions
- **`log`** - Print messages with immediate output
- **`assert`** - Verify conditions with multiple operators
- **`variable`** - Set and manage variables
- **`uuid`** - Generate UUID values
- **`time`** - Generate timestamps (RFC3339, Unix, custom formats)
- **`sleep`** - Pause execution for specified duration (supports ns, μs, ms, s, m, h)

### Encoding Actions
- **`base64_encode`** - Encode data to base64 format
- **`base64_decode`** - Decode base64 data to original format
- **`url_encode`** - URL encode data for query parameters
- **`url_decode`** - URL decode data back to original format
- **`hash`** - Generate hash using MD5, SHA1, SHA256, or SHA512

### File Actions
- **`file_read`** - Read files with automatic format detection (JSON, YAML, CSV, text)

### String Actions
- **`string_random`** - Generate random strings with various charsets (numeric, alphabetic, alphanumeric, hex, custom)
- **`string_replace`** - Replace substrings with support for occurrence limits
- **`string_format`** - Format strings with placeholder substitution

### HTTP Actions  
- **`http`** - HTTP requests (GET, POST, PUT, DELETE, etc.)
  - Extract data with jq: `.status_code`, `.body`, `.headers`
  - Automatically serializes maps/objects to JSON when Content-Type is application/json
  - Example:
  ```yaml
  - name: "Make HTTP POST request"
    action: http
    args: ["POST", "https://api.example.com/data", "${json_data}"]
    options:
      headers:
        Content-Type: "application/json"
      debug: true  # Optional: logs the request body for debugging
    result: http_response
  ```

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
- Creates a structured data object that can be used with other actions

By default, `json_build` returns a structured data object (map or array), which is ideal for:
- Passing to HTTP requests with Content-Type: application/json (automatic serialization)
- Further manipulation with other actions
- Accessing specific fields using variable references

You can also request a JSON string instead of a structured object when needed:

```yaml
- name: "Create user object as JSON string"
  action: json_build
  args:
    - id: "${user_id}"
      name: "${user_name}"
  options:
    format: "string"  # Returns a JSON string instead of structured data
  result: user_json_string
```

**Best Practice**: For HTTP requests with JSON data, use the default structured output from `json_build` and set the Content-Type header to "application/json". The HTTP action will automatically serialize the data.
```yaml
# Recommended approach for HTTP with JSON
- name: "Create user data"
  action: json_build
  args:
    - name: "John Doe"
      email: "john@example.com"
  result: user_data

- name: "Send user data"
  action: http
  args: ["POST", "https://api.example.com/users", "${user_data}"]
  options:
    headers:
      Content-Type: "application/json"
  result: response
```
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
- `retry_on`: Specific error types to retry on - ["assertion_failed", "http_error", "timeout", "connection_error", "all"]
- `stop_on_success`: Stop retrying immediately on success (default: true)
- `retry_if`: Custom condition to determine if retry should continue

**Advanced Retry with Conditions:**

```yaml
steps:
  # Retry based on extracted value
  - name: "Poll until author changes"
    action: http
    args: ["GET", "${api_url}/content"]
    extract:
      type: "jq"
      path: ".body | fromjson | .author"
    result: author
    retry:
      attempts: 5
      delay: "3s"
      retry_if: "${author} == 'Pending'"  # Retry as long as author is 'Pending'
      
  # Retry based on error condition
  - name: "Retry on specific errors"
    action: http
    args: ["GET", "${api_url}/status"]
    retry:
      attempts: 3
      delay: "1s"
      retry_if: "${error_occurred} == true && '${error_message}' contains 'timeout'"
```

### Sleep & Timing Control

Control test execution timing for async operations, polling, and delays:

```yaml
steps:
  # Basic sleep with different duration formats
  - name: "Short delay"
    action: sleep
    args: ["500ms"]  # Milliseconds
    
  - name: "Medium delay"
    action: sleep
    args: ["2s"]     # Seconds
    
  - name: "Long delay"
    action: sleep
    args: ["1m30s"]  # Minutes and seconds
    
  # Variable-based delays
  - name: "Configurable delay"
    action: sleep
    args: ["${retry_delay}"]  # From variable
    result: sleep_info
    
  # Polling simulation
  - name: "Check status"
    action: http
    args: ["GET", "${status_url}"]
    result: status_response
    
  - name: "Wait before next poll"
    action: sleep
    args: ["${polling_interval}"]
```

**Supported duration formats**: `ns`, `us`/`μs`, `ms`, `s`, `m`, `h` (e.g., "100ms", "2.5s", "1m30s")

### Encoding & Security

Handle authentication, data integrity, and URL formatting:

```yaml
steps:
  # Basic Authentication
  - name: "Create Basic Auth header"
    action: base64_encode
    args: ["${username}:${password}"]
    result: auth_token
    
  - name: "Make authenticated request"
    action: http
    args: ["GET", "${api_url}"]
    options:
      headers:
        Authorization: "Basic ${auth_token}"
    
  # URL encoding for query parameters
  - name: "Encode search query"
    action: url_encode
    args: ["user name with spaces & symbols"]
    result: encoded_query
    
  - name: "Build search URL"
    action: variable
    args: ["search_url", "${base_url}/search?q=${encoded_query}"]
    
  # Data integrity with hashing
  - name: "Generate payload hash"
    action: hash
    args: ["${json_payload}", "sha256"]
    result: payload_hash
    
  - name: "Extract hash value"
    action: jq
    args: ["${payload_hash}", ".hash"]
    result: hash_value
```

**Supported hash algorithms**: `md5`, `sha1`, `sha256`, `sha512`

### File Operations & Data-Driven Testing

Load external data files for comprehensive test scenarios:

```yaml
steps:
  # Load JSON test data
  - name: "Load user data"
    action: file_read
    args: ["testdata/users.json"]
    result: users
    
  - name: "Extract first user"
    action: jq
    args: ["${users}", ".content[0].name"]
    result: first_user_name
    
  # Load YAML configuration
  - name: "Load config"
    action: file_read
    args: ["config/api.yaml"]
    result: config
    
  - name: "Get API URL from config"
    action: jq
    args: ["${config}", ".content.api.base_url"]
    result: api_url
    
  # Load CSV test cases for data-driven testing
  - name: "Load test cases"
    action: file_read
    args: ["testdata/test_cases.csv"]
    result: test_cases
    
  - name: "Run first test case"
    action: jq
    args: ["${test_cases}", ".content[0]"]
    result: first_test
    
  # Load plain text templates
  - name: "Load request template"
    action: file_read
    args: ["templates/soap_request.xml"]
    result: template
```

**Supported formats**: JSON (parsed), YAML (parsed), CSV (array of objects), Text (raw string)
**Security**: Path traversal protection, working directory restrictions

### String Operations & Unique Data Generation

Generate unique test data and manipulate strings for comprehensive testing:

```yaml
steps:
  # Generate unique identifiers
  - name: "Generate unique user ID"
    action: string_random
    args: [8, "alphanumeric"]
    result: user_id_data
    
  - name: "Extract user ID"
    action: jq
    args: ["${user_id_data}", ".value"]
    result: user_id
    
  # Generate different types of random data
  - name: "Generate numeric ID"
    action: string_random
    args: [6, "numeric"]
    result: numeric_id
    
  - name: "Generate API key"
    action: string_random
    args: [32, "hex"]
    result: api_key
    
  # Format strings with generated data
  - name: "Create unique email"
    action: string_format
    args: ["test-{}@example.com", "${user_id}"]
    result: email_data
    
  # String replacement for templates
  - name: "Personalize message"
    action: string_replace
    args: ["Hello {{USER}}, welcome!", "{{USER}}", "${user_id}"]
    result: personalized_msg
```

**Supported charsets**: `numeric`, `lowercase`, `uppercase`, `alphabetic`, `alphanumeric`, `hex`, `special`, `all`, `custom`
**Use cases**: Unique user data, API keys, database names, email addresses, test isolation

### Variable Substitution & Data Extraction

Variables use simple `${variable}` substitution and `${ENV:VARIABLE_NAME}` for environment variables. For complex data extraction, use dedicated actions:

```yaml
variables:
  vars:
    api_url: "${ENV:API_BASE_URL}"  # From environment variable
    auth_token: "${ENV:API_TOKEN}"  # From environment variable

steps:
  # Variable substitution with environment variables
  - name: "Make authenticated request"
    action: http
    args: ["GET", "${api_url}/users"]
    options:
      headers:
        Authorization: "Bearer ${auth_token}"
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

**PostgreSQL** - Use environment variables for credentials:
```yaml
# Secure approach using environment variables
variables:
  vars:
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/${ENV:DB_NAME}?sslmode=disable"

steps:
  - action: postgres
    args: ["query", "${db_url}", "SELECT 1"]
```

**For development** - Set up .env file:
```bash
# Create .env file
echo "DB_USER=robogo_testuser" >> .env
echo "DB_PASSWORD=robogo_testpass" >> .env  
echo "DB_HOST=localhost" >> .env
echo "DB_PORT=5432" >> .env
echo "DB_NAME=robogo_testdb" >> .env
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
│   │   ├── sleep.go            # Sleep/delay action
│   │   ├── string.go           # String operations and random generation
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
- **CLI Tool Design**: Single-threaded, single-test execution following Unix philosophy
- **Sequential Test Logic**: Steps execute in defined order for predictable, debuggable behavior
- **Process-Level Parallelism**: Parallelization achieved through multiple CLI invocations, not internal threading
- **Simple & Direct**: No over-engineering, interfaces, or dependency injection
- **Immediate Connections**: Database and messaging connections open/close per operation
- **Clean Process Management**: Clean exit, no hanging processes or background threads
- **Minimal Dependencies**: Only essential libraries
- **KISS Principle**: Keep it simple and straightforward
- **Test Clarity**: Every test step is self-documenting with clear names and expected outcomes

### Error Handling Philosophy

Robogo distinguishes between **Errors** and **Failures** for clear problem classification:

- **Errors** - Technical problems (network issues, parse errors, configuration problems)
- **Failures** - Logical test problems (assertion failures, unexpected response values)

Both types are handled uniformly by the test runner, providing consistent error reporting while maintaining semantic distinction for debugging and analysis.

## Parallelism and Performance

### CLI Tool Design Philosophy

Robogo follows the **Unix philosophy** for CLI tools: do one thing well and compose with other tools. Each `robogo run` command executes a single test file sequentially, which provides:

- **Predictable Behavior**: Steps execute in the exact order defined
- **Easy Debugging**: Clear execution flow without race conditions
- **Reliable Results**: No concurrency bugs or timing issues
- **Simple Architecture**: Single-threaded execution is easier to maintain

### Achieving Parallelism

For parallel test execution, use **process-level parallelism** rather than internal threading:

#### Shell-Level Parallelism
```bash
# Run multiple tests in parallel using shell
./robogo run test1.yaml &
./robogo run test2.yaml &
./robogo run test3.yaml &
wait  # Wait for all to complete
```

#### CI/CD Parallelism
```yaml
# GitHub Actions example
strategy:
  matrix:
    test: [test1.yaml, test2.yaml, test3.yaml]
steps:
  - run: ./robogo run ${{ matrix.test }}
```

#### Makefile Parallelism
```makefile
# Run tests in parallel with make
test-parallel:
	./robogo run test1.yaml & \
	./robogo run test2.yaml & \
	./robogo run test3.yaml & \
	wait
```

### Why Not Internal Parallelism?

Internal step parallelism would break the fundamental testing logic:

```yaml
steps:
  - name: "Create user"
    action: http
    args: ["POST", "/users", "..."]
    result: response
    
  - name: "Verify user created"  # This DEPENDS on the above step
    action: assert
    args: ["${response.status}", "==", "201"]
```

Steps within a test are **intentionally sequential** because they represent a logical flow where later steps depend on earlier results.

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