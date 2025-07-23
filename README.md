# Robogo

A simple, modern test automation framework written in Go. Robogo provides a clean YAML-based DSL for writing test cases with support for HTTP APIs, databases, messaging systems, file operations, and more.

**Shift-Left Testing**: Robogo enables developers to run comprehensive end-to-end tests early in the development cycle with clear, readable test definitions that improve collaboration between development and testing teams.

## Features

### Core Capabilities
- **Developer-Friendly**: Clear, readable YAML tests that developers can easily understand and maintain
- **Shift-Left Ready**: Run full end-to-end tests in development environments
- **Simple YAML Tests**: Write tests in clean, readable YAML format with powerful features
- **KISS Architecture**: Keep It Simple and Straightforward - no over-engineering or complex abstractions

### Actions & Integrations
- **HTTP Testing**: Full HTTP client with all methods, authentication, and response validation
- **Database Support**: PostgreSQL and Google Cloud Spanner with secure credential management
- **Messaging Systems**: Kafka and RabbitMQ operations with producer/consumer support
- **File Operations**: Local file reading and secure SCP file transfers via SSH/SFTP
- **Financial Messaging**: SWIFT message generation for banking and financial testing
- **Data Processing**: JSON/XML parsing, construction, and extraction with jq/xpath support
- **String Operations**: Random generation, formatting, encoding/decoding, and manipulation
- **Utility Actions**: UUID generation, time operations, sleep/timing, assertions, and logging

### Advanced Features
- **Variable Substitution**: Dynamic variables with `${variable}` and `${ENV:VARIABLE}` syntax
- **Security-First**: Automatic sensitive data masking, no-log mode, and environment variable support
- **Control Flow**: Conditional execution (`if`), retry logic with backoff, and nested step collections
- **Data Extraction**: Extract data from responses using jq, xpath, or regex patterns
- **Error Handling**: Comprehensive error categorization with user-friendly messages
- **Clean CLI Tool**: Immediate connection handling - no hanging processes

### Recent Improvements (2024)
- **Architecture Simplification**: Removed 6+ abstraction layers, eliminated dependency injection
- **SCP File Transfer**: Secure SSH/SFTP support with password and key authentication
- **Enhanced Security**: Step-level security controls, comprehensive data masking
- **File Organization**: Split large files into focused, maintainable modules
- **Comprehensive Documentation**: README files throughout codebase for better navigation

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

The quickest way to get started is with our showcase HTTP example:

```bash
# Run the main showcase test (no setup needed)
./robogo run examples/01-http-get.yaml
```

**Main Showcase Example:**
```yaml
testcase: "TC-HTTP-001"
description: "Test to understand HTTP response format"

variables:
  vars:
    base_url: "http://localhost:8000/base64/SFRUUEJJTiBpcyBhd2Vzb21l"
    expected_value: "HTTPBIN is awesome"

steps:
  - name: "Make HTTP request"
    action: http
    args: ["GET", "${base_url}"]
    result: "http_response"

  - name: "Extract status code"
    action: jq
    args: ["${http_response}", ".status_code"]
    result: "status_code"

  - name: "Verify status code"
    action: assert
    args: ["${status_code}", "==", "200"]

  - name: "Extract response body"
    action: jq
    args: ["${http_response}", ".body"]
    result: "response_body"

  - name: "Verify expected content"
    action: assert
    args: ["${response_body}", "==", "${expected_value}"]
```

**Expected output:**
```
Running test case: TC-HTTP-001
Description: Test to understand HTTP response format
Setup: 0, Steps: 7, Teardown: 0

Step 1: Make HTTP request
  Action: http
  Args: [GET http://localhost:8000/base64/SFRUUEJJTiBpcyBhd2Vzb21l]
  Options: map[timeout:5s]
  Result Variable: http_response
  Executing... 
✓ PASSED (10.959687ms)

Step 2: Extract status code
  Action: jq
  Args: [.status_code]
  Result Variable: status_code
  Executing... 
✓ PASSED (973.434µs)
    Data: 200

Step 3: Verify status code
  Action: assert
  Args: [200 == 200]
  Executing... 
✓ PASSED (306.971µs)

Step 4: Extract response body
  Action: jq
  Args: [.body]
  Result Variable: response_body
  Executing... 
✓ PASSED (151.688µs)
    Data: HTTPBIN is awesome

Step 5: Verify expected content
  Action: assert
  Args: [HTTPBIN is awesome == HTTPBIN is awesome]
  Executing... 
✓ PASSED (475.234µs)

Test Summary:
  Name: TC-HTTP-001
  Status: PASS
  Duration: 13.589684ms

|   # | Step Name                                | Status   | Duration     |
|-----|------------------------------------------|----------|--------------|
|   1 | Make HTTP request                        | PASS     | 10.959687ms  |
|   2 | Extract status code                      | PASS     | 973.434µs    |
|   3 | Verify status code                       | PASS     | 306.971µs    |
|   4 | Extract response body                    | PASS     | 151.688µs    |
|   5 | Verify expected content                  | PASS     | 475.234µs    |
```

**What this demonstrates:**
- HTTP requests with response capture (`result: "http_response"`)
- JSON data extraction using `jq` action (`.status_code`, `.body`)
- Variable substitution and storage (`${http_response}`, `${status_code}`)
- Assertions for validation (`assert` action)
- Clean test structure with clear step-by-step execution
- Performance timing for each step

## Documentation

### 📚 **Getting Started**
- **[examples/README.md](examples/README.md)** - Comprehensive test examples from beginner to expert
- **[docs/execution-flow-diagram.md](docs/execution-flow-diagram.md)** - Visual architecture flow diagram
- **[docs/error-failure-states-diagram.md](docs/error-failure-states-diagram.md)** - Error handling and state management flow

### 🏗️ **Architecture**
- **[internal/README.md](internal/README.md)** - Core architecture principles and KISS design
- **[internal/execution/README.md](internal/execution/README.md)** - Execution strategy pattern system
- **[internal/actions/README.md](internal/actions/README.md)** - Complete action system documentation

### 📖 **Reference**
- **[docs/README.md](docs/README.md)** - Documentation overview and navigation guide
- **[CLAUDE.md](CLAUDE.md)** - Development instructions and project context

## Action Categories

### Core Actions
- **`assert`** - Test assertions and validations
- **`log`** - Logging and output messages  
- **`variable`** - Variable manipulation and setting

### HTTP & API Testing
- **`http`** - HTTP requests (GET, POST, PUT, DELETE, etc.) with full header and authentication support

### Database Operations
- **`postgres`** - PostgreSQL database queries and operations
- **`spanner`** - Google Cloud Spanner distributed database support

### File Operations
- **`file_read`** - Local file reading with format detection
- **`scp`** - Secure file transfer via SSH/SFTP (upload/download)

### Messaging Systems
- **`kafka`** - Apache Kafka producer/consumer operations
- **`rabbitmq`** - RabbitMQ message operations
- **`swift_message`** - SWIFT financial messaging (MT103)

### Data Processing
- **`jq`** - JSON data processing and extraction
- **`xpath`** - XML data processing and queries
- **`json_parse`/`json_build`** - JSON parsing and construction
- **`xml_parse`/`xml_build`** - XML parsing and construction

### String & Encoding
- **`string_random`** - Random string generation
- **`string_replace`/`string_format`** - String manipulation
- **`base64_encode`/`base64_decode`** - Base64 operations
- **`url_encode`/`url_decode`** - URL encoding
- **`hash`** - Cryptographic hashing (MD5, SHA1, SHA256)

### Utilities
- **`uuid`** - UUID v4 generation
- **`time`** - Time operations and formatting
- **`sleep`** - Delays and timing control

## Test Structure

### Example Tests

Robogo includes 50+ comprehensive test examples. Here are the key categories:

| Category | Example | Features Demonstrated | Command |
|----------|---------|----------------------|---------|
| **HTTP Basics** | [01-http-get.yaml](examples/01-http-get.yaml) | GET requests, jq extraction, assertions | `./robogo run examples/01-http-get.yaml` |
| **HTTP POST** | [02-http-post.yaml](examples/02-http-post.yaml) | POST with JSON, nested data extraction | `./robogo run examples/02-http-post.yaml` |
| **Utilities** | [00-util.yaml](examples/00-util.yaml) | UUID generation, basic logging | `./robogo run examples/00-util.yaml` |
| **Environment Variables** | [17-env-var-test.yaml](examples/17-env-var-test.yaml) | ${ENV:VAR} syntax, credential management | `./robogo run examples/17-env-var-test.yaml` |
| **Security** | [19-no-log-security.yaml](examples/19-no-log-security.yaml) | no_log, sensitive_fields, data masking | `./robogo run examples/19-no-log-security.yaml` |
| **Database** | [03-postgres-basic.yaml](examples/03-postgres-basic.yaml) | PostgreSQL operations, queries | `./robogo run examples/03-postgres-basic.yaml` |
| **Conditional Logic** | [08-control-flow.yaml](examples/08-control-flow.yaml) | Conditional execution with if statements | `./robogo run examples/08-control-flow.yaml` |
| **Retry Logic** | [13-retry-demo.yaml](examples/13-retry-demo.yaml) | Retry with backoff, error handling | `./robogo run examples/13-retry-demo.yaml` |
| **Nested Steps** | [21-simple-nested-test.yaml](examples/21-simple-nested-test.yaml) | Grouped operations, continue-on-error | `./robogo run examples/21-simple-nested-test.yaml` |
| **File Transfer** | [23-scp-simple-test.yaml](examples/23-scp-simple-test.yaml) | SSH/SFTP file operations | `./robogo run examples/23-scp-simple-test.yaml` |

**📁 Browse all examples:** See **[examples/README.md](examples/README.md)** for the complete catalog with beginner to expert examples.

### Key Test Structure Elements
All Robogo tests follow this pattern:
```yaml
testcase: "Test Name"
description: "What this test does"

variables:
  vars:
    variable_name: "value"
    api_url: "${ENV:API_URL}"  # Environment variables

steps:
  - name: "Step description"
    action: action_name
    args: [arg1, arg2, arg3]
    result: result_variable
    
  - name: "Verify result"
    action: assert
    args: ["${result_variable.some_field}", "==", "expected_value"]
```

**Important:** Use `jq` action to extract data from HTTP responses - simple `${response.field}` syntax doesn't work for complex objects.

**Advanced Features:** The examples table above includes advanced patterns like retry logic, control flow, nested steps, and security features. For the complete catalog with complexity levels, see **[examples/README.md](examples/README.md)**.

## Security Features

### Security Examples

Robogo provides comprehensive security features for sensitive data handling:

| Security Feature | Example | What It Demonstrates | Command |
|-----------------|---------|---------------------|---------|
| **Environment Variables** | [17-env-var-test.yaml](examples/17-env-var-test.yaml) | `${ENV:VAR}` syntax, credential management | `export TEST_ENV_VAR="test_value" && ./robogo run examples/17-env-var-test.yaml` |
| **Database Security** | [03-postgres-secure.yaml](examples/03-postgres-secure.yaml) | .env file usage, secure DB connections | `./robogo run examples/03-postgres-secure.yaml` |
| **No-Log Mode** | [19-no-log-security.yaml](examples/19-no-log-security.yaml) | Complete logging suppression, sensitive fields | `./robogo run examples/19-no-log-security.yaml` |
| **Step-Level Masking** | [20-step-level-masking.yaml](examples/20-step-level-masking.yaml) | Custom field masking, fine-grained controls | `./robogo run examples/20-step-level-masking.yaml` |

**Key Security Features:**
- **Automatic masking**: Password, token, key fields automatically hidden
- **Custom masking**: Use `sensitive_fields: ["field_name"]` for custom fields
- **No-log mode**: Use `no_log: true` to suppress all step logging
- **Environment variables**: Use `${ENV:VARIABLE}` for secure credential access

## Development Environment

### Prerequisites
- Go 1.24+
- Docker & Docker Compose (for services)

### Development Services

```bash
# Start all services
docker-compose up -d

# Services available:
# - PostgreSQL: localhost:5432 (user: robogo_testuser, pass: robogo_testpass, db: robogo_testdb)
# - Kafka: localhost:9092  
# - Spanner Emulator: localhost:9010
# - HTTPBin: localhost:8000
# - SSH Server: localhost:2222 (user: testuser, pass: testpass)
```

### Environment Configuration

Create `.env` file for secure credential management:
```bash
# Database credentials
DB_HOST=localhost
DB_PORT=5432
DB_USER=robogo_testuser
DB_PASSWORD=robogo_testpass
DB_NAME=robogo_testdb

# API credentials
API_BASE_URL=https://api.example.com
API_TOKEN=your_secret_token_here

# SSH credentials for SCP testing
SSH_PASSWORD=testpass
```

### Database Setup

**PostgreSQL** - Use environment variables for credentials:
```yaml
variables:
  vars:
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/${ENV:DB_NAME}?sslmode=disable"

steps:
  - name: "Test database connection"
    action: postgres
    args: ["query", "${db_url}", "SELECT version()"]
    result: db_version
```

**Google Cloud Spanner** - Set up emulator:
```bash
# After starting docker-compose
# Linux/Mac:
SPANNER_EMULATOR_HOST=localhost:9010 ./setup-spanner.sh
# Windows:
.\setup-spanner.ps1
```

## Example Tests

The **[examples/](examples/)** directory contains 50+ comprehensive test examples organized by complexity and feature:

### Quick Examples
```bash
# HTTP testing (no services required)
./robogo run examples/01-http-get.yaml

# Database testing (requires docker-compose up -d)
./robogo run examples/03-postgres-basic.yaml

# SCP file transfer testing
./robogo run examples/23-scp-simple-test.yaml

# Messaging systems
./robogo run examples/05-kafka-basic.yaml

# Security features
./robogo run examples/19-no-log-security.yaml
```

### Example Categories
- **Beginner**: Basic HTTP, database, and file operations
- **Intermediate**: Multi-step workflows, environment variables, data extraction
- **Advanced**: Complex control flow, retry logic, nested operations
- **Expert**: Security-aware testing, production-ready patterns

## Architecture

### KISS Principles
Robogo follows **Keep It Simple and Straightforward** architecture:

- **No Dependency Injection**: Direct object construction throughout
- **No Over-abstraction**: Simple, direct implementations
- **Minimal Interfaces**: Only where absolutely necessary
- **Strategy Pattern**: Clean execution routing for different step types

### Design Philosophy: Explicit Tests Over Loops

Robogo intentionally **does not support `for` and `while` loops** in test definitions. This design decision prioritizes **test clarity and maintainability** over code brevity.

**Why no loops?**
- **Test purpose matters**: Behavioral tests should be explicit about what they're testing
- **Debugging clarity**: `test_user_creation_with_missing_email()` is clearer than "step 7 failed in user creation loop"
- **Living documentation**: Tests serve as executable specifications - loops obscure intent
- **Industry alignment**: Most YAML-based testing frameworks avoid complex control flow

**When to use explicit tests vs loops:**
- ✅ **Explicit tests for**: Business logic validation, user workflows, API contract testing
- ❌ **Avoid loops for**: Individual test scenarios, specific edge cases, acceptance criteria
- ⚠️ **Loops might be appropriate for**: Framework testing, property-based testing, infrastructure validation

**Example of preferred explicit approach:**
```yaml
# ✅ Clear and maintainable
steps:
  - name: "User registration accepts valid email"
    action: http
    args: ["POST", "/users", '{"email": "user@example.com"}']
    
  - name: "User registration rejects email without @"
    action: http  
    args: ["POST", "/users", '{"email": "invalid-email"}']
```

Instead of:
```yaml  
# ❌ Obscures test intent
steps:
  - name: "Test email validation"
    for: "[user@example.com, invalid-email]"
    action: http
    args: ["POST", "/users", '{"email": "${item}"}']
```

This philosophy aligns with industry best practices where test automation frameworks either avoid loops entirely (GitHub Actions) or use specialized syntax (Robot Framework) rather than general-purpose loops.

### Execution Flow
1. **CLI** receives command and parses YAML test file
2. **TestRunner** creates execution environment with variables and strategy router
3. **ExecutionStrategyRouter** routes steps based on priority:
   - **ConditionalExecutionStrategy** (Priority 4): Handles `if` conditions
   - **RetryExecutionStrategy** (Priority 3): Handles `retry` configuration
   - **NestedStepsExecutionStrategy** (Priority 2): Handles `steps` arrays
   - **BasicExecutionStrategy** (Priority 1): Handles simple actions
4. **Actions** perform actual operations and return structured results
5. **Results** are processed, masked for security, and displayed

For detailed architecture documentation, see **[internal/README.md](internal/README.md)** and **[docs/execution-flow-diagram.md](docs/execution-flow-diagram.md)**.

## Error Handling

### Dual Error System
- **ErrorInfo**: Technical problems (network failures, syntax errors, etc.)
- **FailureInfo**: Logical test failures (assertion failures, unexpected responses)

### Visual Error Flow Diagram
For a comprehensive visual explanation of Robogo's error handling, execution flow, and state management, see:
**[docs/error-failure-states-diagram.md](docs/error-failure-states-diagram.md)** - Complete mermaid diagram showing execution strategies, error classification, and result processing.

### Structured Error Messages
```yaml
# Technical error example
steps:
  - name: "Invalid database query"
    action: postgres
    args: ["query", "invalid://connection", "SELECT 1"]
    # Results in ErrorInfo with connection details and suggestions

# Logical failure example  
  - name: "Assertion failure"
    action: assert
    args: ["${status_code}", "==", "200"]  # After extracting with jq
    # Results in FailureInfo showing expected vs actual values
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

# 3. Run in CI/CD pipeline
./robogo run tests/smoke-tests.yaml
```

## Parallel Execution

Robogo is designed for **test-level parallelism** (multiple test files) rather than **step-level parallelism** (steps within a test).

### ⚠️ Output Mixing Issue

When running tests in parallel **without output redirection**, the stdout will be mixed and confusing:

```bash
# ❌ This will create mixed, unreadable output
./robogo run test1.yaml & \
./robogo run test2.yaml & \
./robogo run test3.yaml & \
wait
```

### ✅ Recommended Parallel Approaches

**Option 1: Redirect to separate files**
```bash
# Each test outputs to its own file
./robogo run test1.yaml > test1.log 2>&1 & \
./robogo run test2.yaml > test2.log 2>&1 & \
./robogo run test3.yaml > test3.log 2>&1 & \
wait

# View results separately
cat test1.log test2.log test3.log
```

**Option 2: Run sequentially with timing**
```bash
# Fast sequential execution, readable output
time ./robogo run test1.yaml
time ./robogo run test2.yaml  
time ./robogo run test3.yaml
```

**Option 3: Use CI/CD parallelism**
```yaml
# GitHub Actions example
jobs:
  test:
    strategy:
      matrix:
        test: [test1.yaml, test2.yaml, test3.yaml]
    runs-on: ubuntu-latest
    steps:
      - run: ./robogo run ${{ matrix.test }}
```

### Why Sequential Steps?
Steps within a test are intentionally sequential because they represent a logical flow where later steps depend on earlier results:

```yaml
steps:
  - name: "Create user"
    action: http
    args: ["POST", "/users", "..."]
    result: response
    
  - name: "Verify user created"  # This DEPENDS on the above step
    action: assert
    args: ["${status_code}", "==", "201"]  # After extracting with jq
```

## Troubleshooting

### Common Issues

1. **Service connection errors**: Ensure Docker services are running (`docker-compose ps`)
2. **Environment variable issues**: Check `.env` file exists and variables are properly formatted
3. **SCP/SSH connection issues**: Verify SSH server is running and credentials are correct
4. **Variable resolution errors**: Check `${variable}` syntax and variable names
5. **Database connection errors**: Verify Docker services are healthy

### Debug Tips

- Use `log` actions to inspect variable values
- Check Docker service logs: `docker-compose logs <service>`
- Use shorter timeouts for faster feedback during development
- Enable verbose logging for debugging complex variable substitution

### Getting Help

- **Documentation**: Start with **[examples/README.md](examples/README.md)** for practical examples
- **Architecture**: See **[internal/README.md](internal/README.md)** for understanding the codebase
- **Issues**: Report bugs or feature requests on the project repository

## Contributing

1. **Follow KISS principles**: Avoid over-engineering and complex abstractions
2. **Add examples**: Every new feature should include working test examples
3. **Update documentation**: Keep README files current with code changes
4. **Security-first**: Ensure sensitive data is properly masked
5. **Test thoroughly**: Verify examples work with standard Docker setup

## License

MIT License - see LICENSE file for details.