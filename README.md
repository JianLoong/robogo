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

Create `hello-world.yaml`:

```yaml
testcase: "Hello World Test"
description: "Basic HTTP test demonstrating Robogo"

steps:
  - name: "Test HTTPBin service"
    action: http
    args: ["GET", "https://httpbin.org/json"]
    result: response
    
  - name: "Verify response"
    action: assert
    args: ["${response.status}", "==", "200"]
    
  - name: "Log success"
    action: log
    args: ["Test completed successfully!"]
```

Run it:
```bash
./robogo run hello-world.yaml
```

## Documentation

### 📚 **Getting Started**
- **[examples/README.md](examples/README.md)** - Comprehensive test examples from beginner to expert
- **[docs/execution-flow-diagram.md](docs/execution-flow-diagram.md)** - Visual architecture flow diagram

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

### Basic Test
```yaml
testcase: "User Registration Test"
description: "Test user registration API endpoint"

variables:
  vars:
    api_url: "https://api.example.com"
    test_user: "testuser@example.com"

steps:
  - name: "Register new user"
    action: http
    args: ["POST", "${api_url}/users"]
    options:
      json:
        email: "${test_user}"
        password: "SecurePass123!"
    result: registration_response
    
  - name: "Verify registration success"
    action: assert
    args: ["${registration_response.status}", "==", "201"]
```

### Advanced Test with All Features
```yaml
testcase: "Advanced Integration Test"
description: "Demonstrates advanced Robogo features"

variables:
  vars:
    api_base: "${ENV:API_BASE_URL}"
    auth_token: "${ENV:API_TOKEN}"

setup:
  - name: "Initialize test data"
    action: variable
    args: ["test_id", "${uuid}"]

steps:
  # Conditional execution
  - name: "Admin-only setup"
    if: "${user_role} == 'admin'"
    action: log
    args: ["Running admin setup"]
    
  # HTTP with retry logic
  - name: "Create user with retry"
    action: http
    args: ["POST", "${api_base}/users"]
    options:
      headers:
        Authorization: "Bearer ${auth_token}"
      json:
        id: "${test_id}"
        email: "test-${test_id}@example.com"
    retry:
      attempts: 3
      delay: "2s"
      backoff: "exponential"
      retry_on: ["http_error", "timeout"]
    result: user_response
    
  # Data extraction
  - name: "Extract user ID"
    action: jq
    args: ["${user_response.data}", ".user.id"]
    result: user_id
    
  # Database verification
  - name: "Verify user in database"
    action: postgres
    args: ["query", "${ENV:DB_URL}", "SELECT * FROM users WHERE id = $1", "${user_id}"]
    result: db_user
    
  # File operations
  - name: "Upload user avatar"
    action: scp
    args: ["upload", "user@server:22", "./avatar.png", "/uploads/${user_id}/avatar.png"]
    options:
      password: "${ENV:SSH_PASSWORD}"
    sensitive_fields: ["password"]
    result: upload_result
    
  # Nested operations
  - name: "Notification workflow"
    steps:
      - name: "Send welcome email"
        action: http
        args: ["POST", "${api_base}/emails/welcome"]
        options:
          json:
            user_id: "${user_id}"
        continue: true
        
      - name: "Log to analytics"
        action: kafka
        args: ["produce", "localhost:9092", "user-events"]
        options:
          message:
            event: "user_registered"
            user_id: "${user_id}"
            timestamp: "${time}"

teardown:
  - name: "Cleanup test user"
    action: http
    args: ["DELETE", "${api_base}/users/${user_id}"]
    options:
      headers:
        Authorization: "Bearer ${auth_token}"
```

## Security Features

### Environment Variables
```yaml
variables:
  vars:
    # Secure credential management
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/${ENV:DB_NAME}"
    api_token: "${ENV:API_TOKEN}"
```

### Sensitive Data Masking
```yaml
steps:
  # Automatic masking of password fields
  - name: "Login request"
    action: http
    args: ["POST", "/auth/login"]
    options:
      json:
        username: "testuser"
        password: "${ENV:USER_PASSWORD}"  # Automatically masked in logs
    result: auth_response
    
  # Custom field masking
  - name: "API call with secrets"
    action: http
    args: ["GET", "/secure-data"]
    options:
      headers:
        X-API-Key: "${ENV:SECRET_API_KEY}"
        X-Session-Token: "${session_token}"
    sensitive_fields: ["X-Session-Token"]  # Custom masking
    result: secure_data
    
  # Complete logging suppression
  - name: "Highly sensitive operation"
    action: http
    args: ["POST", "/admin/reset-passwords"]
    no_log: true  # No step details logged at all
    result: reset_result
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
    args: ["${response.status}", "==", "200"]
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

Robogo is designed for **test-level parallelism** (multiple test files) rather than **step-level parallelism** (steps within a test):

```bash
# Run multiple tests in parallel
./robogo run test1.yaml & \
./robogo run test2.yaml & \
./robogo run test3.yaml & \
wait
```

**Why Sequential Steps?** Steps within a test are intentionally sequential because they represent a logical flow where later steps depend on earlier results:

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