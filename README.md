# Robogo - Modern Test Automation Framework

A modern, git-driven test automation framework written in Go, designed for comprehensive API testing, SWIFT message generation, database operations, and Test Data Management (TDM).

## ✨ Key Features

- **🔧 Template-based SWIFT Message Generation** - Create and test SWIFT messages with dynamic variable substitution
- **🌐 HTTP API Testing** - Full HTTP support with mTLS, custom headers, and comprehensive response validation
- **💾 Database Integration** - PostgreSQL operations with connection pooling and secure credential management
- **📊 Test Data Management (TDM)** - Structured data sets, environment management, and data lifecycle
- **🎲 Enhanced Random Generation** - Support for both integer and decimal random values with precision control
- **🔄 Advanced Control Flow** - If statements, for loops, while loops with conditional logic and retry mechanisms
- **🔐 Secret Management** - Secure handling of API keys, certificates, and sensitive data with masking
- **📊 Multiple Output Formats** - Console, JSON, and Markdown reporting with detailed analytics
- **⚡ Performance Testing** - Built-in timing, load testing, and retry capabilities
- **🔍 Comprehensive Validation** - Data validation, format checking, and assertion framework
- **🛠️ VS Code Integration** - Syntax highlighting, autocomplete, and code snippets

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or later

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/robogo.git
cd robogo

# Install dependencies
go mod download

# Build the binary
go build -o robogo.exe ./cmd/robogo
```

### Run Your First Test

```bash
# Run a basic test
./robogo.exe run examples/sample.robogo

# Run SWIFT message testing
./robogo.exe run tests/test-swift-working.robogo

# Run Test Data Management
./robogo.exe run tests/test-tdm-simple.robogo

# Run decimal random testing
./robogo.exe run tests/test-random-decimals.robogo
```

## 📋 Available Actions

### Basic Operations
- **`log`** - Output messages to console with verbosity control
- **`sleep`** - Pause execution for specified duration
- **`assert`** - Verify conditions with comparison operators (==, !=, >, <, >=, <=, contains, starts_with, ends_with)

### Time and Random
- **`get_time`** - Get current timestamp (iso, datetime, date, time, unix, unix_ms, custom formats)
- **`get_random`** - Generate random numbers (integers and decimals with precision control)

### String Operations
- **`concat`** - Concatenate multiple strings
- **`length`** - Get length of strings or arrays

### HTTP Operations
- **`http`** - Generic HTTP requests with mTLS support and custom options
- **`http_get`** - Simplified GET requests
- **`http_post`** - Simplified POST requests

### Database Operations
- **`postgres`** - PostgreSQL operations (query, execute, connect, close)

### Control Flow
- **`control`** - Conditional execution and loop control
- **`if`** - Conditional execution with then/else blocks
- **`for`** - Loop execution (ranges, arrays, counts)
- **`while`** - Conditional loops with max iteration limits

### Test Data Management
- **`tdm`** - Test Data Management operations (generate, validate, load_dataset, set_environment)
- **`variable`** - Variable management operations (set_variable, get_variable, list_variables)

## 📊 Test Data Management (TDM)

Robogo includes a comprehensive Test Data Management system for structured data handling:

```yaml
testcase: "TDM Example"
description: "Demonstrate Test Data Management features"

# Environment configuration
environments:
  - name: "development"
    description: "Development environment"
    variables:
      api_base_url: "https://dev-api.example.com"
      timeout: 30
    overrides:
      debug_mode: true

# Test Data Management configuration
data_management:
  environment: "development"
  isolation: true
  cleanup: true
  
  # Structured data sets
  data_sets:
    - name: "test_users"
      description: "Test user data"
      version: "1.0"
      data:
        user1:
          name: "John Doe"
          email: "john@example.com"
          age: 30
        user2:
          name: "Jane Smith"
          email: "jane@example.com"
          age: 25
      schema:
        name: "string"
        email: "email"
        age: "number"
      required: ["name", "email"]
      unique: ["email"]

  # Data validation
  validation:
    - name: "email_validation"
      type: "format"
      field: "test_users.user1.email"
      rule: "email"
      message: "User email must be valid"
      severity: "error"

  # Setup and teardown
  setup:
    - name: "TDM Setup"
      action: log
      args: ["Setting up test environment"]

  teardown:
    - name: "TDM Cleanup"
      action: log
      args: ["Cleaning up test environment"]

steps:
  # Use TDM data
  - name: "Log user data"
    action: log
    args: ["User: ${test_users.user1.name} (${test_users.user1.email})"]
  
  # Database operations with TDM data
  - name: "Insert user"
    action: postgres
    args: ["execute", "postgres://user:pass@localhost/db", "INSERT INTO users (name, email) VALUES ($1, $2)", ["${test_users.user1.name}", "${test_users.user1.email}"]]
```

## 🏦 SWIFT Message Testing

Robogo excels at SWIFT message generation and testing:

```yaml
testcase: "SWIFT Message Test"
description: "Generate and test SWIFT messages"

variables:
  vars:
    bank_bic: "DEUTDEFF"
    currency: "EUR"
    test_amount: "1000.00"
  secrets:
    swift_api_key:
      file: "secret.txt"
      mask_output: true

steps:
  # Generate unique transaction ID
  - action: get_time
    args: ["unix_ms"]
    result: timestamp_ms
  
  - action: concat
    args: ["TXN", "${timestamp_ms}"]
    result: transaction_id

  # Generate SWIFT MT103 message
  - action: concat
    args: [
      "{1:F01", "${bank_bic}", "XXXX", "U", "3003", "1234567890", "}",
      "{2:I103", "${bank_bic}", "XXXX", "U}",
      "{3:{113:SEPA}",
      "{108:${transaction_id}}",
      "{111:001}",
      "{121:${timestamp_ms}}}",
      "{4:",
      ":20:${transaction_id}",
      ":23B:CRED",
      ":32A:${current_date}${currency}${test_amount}",
      ":33B:${currency}${test_amount}",
      ":50K:/1234567890",
      "1/Account Name",
      ":59:/0987654321",
      "1/Beneficiary Name",
      ":70:INV-2024-001",
      ":71A:SHA",
      "-}",
      "{5:{CHK:1234567890ABCD}{TNG:}}{S:{COP:S}}"
    ]
    result: swift_message

  # Test via HTTP API
  - action: http_post
    args: 
      - "https://api.swift.com/v1/messages"
      - '{"message": "${swift_message}", "type": "MT103"}'
    result: api_response

  # Validate response
  - action: assert
    args: ["${api_response.status_code}", "==", "200", "API should return 200"]
```

## 🎲 Enhanced Random Generation

Support for both integer and decimal random values with precision control:

```yaml
# Integer random (backward compatible)
- action: get_random
  args: [100]
  result: int_random

# Decimal random (new feature)
- action: get_random
  args: [100.5]
  result: decimal_random

# SWIFT amount generation
- action: get_random
  args: [50000.00]
  result: swift_amount

# Multiple random values in loop
- for:
    condition: "1..5"
    steps:
      - action: get_random
        args: [1000.25]
        result: iteration_amount
      
      - action: log
        args: ["Amount ${iteration}: ${iteration_amount}"]
```

## 🌐 HTTP API Testing

Comprehensive HTTP testing with mTLS support:

```yaml
# Simple GET request
- action: http_get
  args: ["https://api.example.com/users"]
  result: response

# POST with JSON body
- action: http_post
  args: 
    - "https://api.example.com/users"
    - '{"name": "John", "email": "john@example.com"}'
  result: create_response

# mTLS request with certificates
- action: http
  args: 
    - "POST"
    - "https://secure.example.com/api"
    - '{"secure": true}'
    - 
      Content-Type: "application/json"
      Authorization: "Bearer ${API_TOKEN}"
    - 
      cert: "${CLIENT_CERT_PATH}"
      key: "${CLIENT_KEY_PATH}"
      ca: "${CA_CERT_PATH}"
  result: secure_response
```

## 💾 Database Operations

PostgreSQL integration with secure credential management:

```yaml
variables:
  secrets:
    db_password:
      file: "db-secret.txt"
      mask_output: true

steps:
  # Build connection string
  - action: concat
    args: ["postgres://user:", "${db_password}", "@localhost/db"]
    result: db_connection

  # Execute query
  - action: postgres
    args: ["query", "${db_connection}", "SELECT * FROM users"]
    result: query_result

  # Validate results
  - action: assert
    args: ["${query_result.rows_affected}", ">", "0", "Should return results"]
```

## 📝 Test Case Format

Test cases are written in YAML format with comprehensive features:

```yaml
testcase: "Comprehensive Test"
description: "Test with variables, secrets, control flow, and TDM"

variables:
  vars:
    api_url: "https://api.example.com"
    timeout: 30
  secrets:
    api_key:
      file: "secret.txt"
      mask_output: true

steps:
  # Control flow with loops and retry
  - for:
      condition: "1..3"
      steps:
        - action: get_random
          args: [1000.50]
          result: amount
        
        - action: http_post
          args: 
            - "${api_url}/transactions"
            - '{"amount": "${amount}"}'
          result: response
          retry:
            attempts: 3
            delay: "1s"
            backoff: "exponential"
        
        - action: assert
          args: ["${response.status_code}", "==", "200"]

  # Conditional execution
  - if:
      condition: "${response.status_code} == 200"
      then:
        - action: log
          args: ["Transaction successful"]
      else:
        - action: log
          args: ["Transaction failed"]
```

## 🧪 Example Test Cases

### Core Functionality
- **`examples/sample.robogo`** - Basic functionality demonstration
- **`tests/test-syntax.robogo`** - Syntax and basic operations
- **`tests/test-variables.robogo`** - Variable management and substitution
- **`tests/test-assert.robogo`** - Assertion and validation examples

### Advanced Features
- **`tests/test-tdm-simple.robogo`** - Simple Test Data Management
- **`tests/test-tdm.robogo`** - Comprehensive TDM with PostgreSQL integration
- **`tests/test-control-flow.robogo`** - Control flow features (if, for, while)
- **`tests/test-retry.robogo`** - Retry mechanisms and error handling

### SWIFT and Financial
- **`tests/test-swift-working.robogo`** - SWIFT message generation and testing
- **`tests/test-swift-messages.robogo`** - Advanced SWIFT message examples
- **`tests/test-swift-advanced.robogo`** - Complex SWIFT workflows

### API and Database Testing
- **`tests/test-http.robogo`** - HTTP API testing examples
- **`tests/test-postgres.robogo`** - Database operations and queries
- **`tests/test-secrets.robogo`** - Secret management and security

### Random Generation and Utilities
- **`tests/test-random-decimals.robogo`** - Enhanced random number generation
- **`tests/test-random-ranges.robogo`** - Random value ranges and validation
- **`tests/test-random-edge-cases.robogo`** - Edge cases and boundary testing
- **`tests/test-time-formats.robogo`** - Time formatting and manipulation

### Error Handling and Validation
- **`tests/test-fail-in-loop.robogo`** - Error handling in loops
- **`tests/test-continue-on-failure.robogo`** - Continue on failure scenarios
- **`tests/test-verbosity.robogo`** - Verbosity levels and logging

## 🏗️ Project Structure

```
robogo/
├── cmd/robogo/          # CLI entry point
├── internal/           # Core framework code
│   ├── actions/        # Built-in actions (HTTP, DB, TDM, etc.)
│   ├── parser/         # YAML parsing and test execution
│   └── runner/         # Test orchestration and TDM integration
├── tests/              # Comprehensive test examples (25+ test cases)
├── examples/           # Basic examples and tutorials
├── docs/              # Documentation and guides
│   ├── tdm-implementation.md    # TDM system documentation
│   ├── tdm-evaluation-summary.md # TDM evaluation and analysis
│   ├── actions.md      # Complete action reference
│   ├── cli-reference.md # CLI documentation
│   └── quickstart.md   # Getting started guide
├── prd/               # Product requirements and specifications
└── .vscode/           # VS Code extension and configuration
```

## 🔧 Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run specific test
./robogo.exe run tests/test-swift-working.robogo

# Run TDM test
./robogo.exe run tests/test-tdm-simple.robogo

# Run with specific output format
./robogo.exe run test.robogo --output json
```

### Build

```bash
go build -o robogo.exe ./cmd/robogo
```

### List Available Actions

```bash
./robogo.exe list
```

### Get Action Completions

```bash
./robogo.exe completions get_random
```

## 📊 Output Formats

Robogo supports multiple output formats with detailed analytics:

```bash
# Console output (default) - Human-readable with colors and formatting
./robogo.exe run test.robogo

# JSON output - Machine-readable for CI/CD integration
./robogo.exe run test.robogo --output json

# Markdown output - Documentation-friendly format
./robogo.exe run test.robogo --output markdown
```

## 📚 Documentation

Comprehensive documentation available in the [docs/](docs/) directory:

- **[TDM Implementation Guide](docs/tdm-implementation.md)** - Complete Test Data Management system documentation
- **[TDM Evaluation Summary](docs/tdm-evaluation-summary.md)** - TDM system analysis and evaluation
- **[Actions Reference](docs/actions.md)** - Complete list of available actions with examples
- **[Quick Start Guide](docs/quickstart.md)** - Get started quickly with Robogo
- **[Test Cases Guide](docs/test-cases.md)** - Writing effective test cases
- **[CLI Reference](docs/cli-reference.md)** - Command-line interface documentation
- **[Contributing Guide](docs/CONTRIBUTING.md)** - How to contribute to Robogo

## 🎯 Use Cases

### Financial Services
- **SWIFT Message Testing** - Generate and validate SWIFT messages (MT103, MT202, etc.)
- **Payment API Testing** - Test payment processing systems and workflows
- **Banking Integration** - Validate banking APIs, compliance, and regulatory requirements
- **Test Data Management** - Structured data sets for financial testing scenarios

### API Testing
- **REST API Validation** - Comprehensive HTTP API testing with authentication
- **mTLS Security Testing** - Test secure API endpoints with certificate validation
- **Performance Testing** - Load testing, retry mechanisms, and performance validation
- **Data-Driven Testing** - TDM-powered test scenarios with multiple data sets

### Database Testing
- **PostgreSQL Operations** - Database query, transaction, and integration testing
- **Data Validation** - Verify database state, results, and data integrity
- **Integration Testing** - End-to-end database workflows with TDM data sets
- **Data Lifecycle Management** - Setup, teardown, and cleanup operations

### Test Automation
- **CI/CD Integration** - Automated testing in continuous integration pipelines
- **Regression Testing** - Comprehensive test suites with TDM data management
- **Load and Performance** - Scalable testing with retry mechanisms and timing
- **Cross-Platform Testing** - Consistent testing across different environments

## 🚀 Roadmap

### Completed Features ✅
- [x] **Test Data Management (TDM)** - Structured data sets and lifecycle management
- [x] **Enhanced Random Generation** - Decimal support with precision control
- [x] **Comprehensive HTTP Testing** - mTLS, headers, and response validation
- [x] **PostgreSQL Integration** - Database operations with connection pooling
- [x] **Advanced Control Flow** - If, for, while loops with retry mechanisms
- [x] **Secret Management** - Secure credential handling with masking
- [x] **VS Code Integration** - Syntax highlighting and autocomplete

### Planned Features 🚧
- [ ] **Plugin System** - Custom action development and extensibility
- [ ] **Parallel Execution** - Concurrent test execution for performance
- [ ] **Web Interface** - Browser-based test management and monitoring
- [ ] **Advanced Reporting** - Detailed analytics, dashboards, and metrics
- [ ] **Cloud Integration** - AWS, Azure, GCP support and cloud-native testing
- [ ] **CI/CD Integration** - Jenkins, GitHub Actions, GitLab CI templates
- [ ] **Multi-Database Support** - MySQL, SQLite, MongoDB integration
- [ ] **GraphQL Testing** - Native GraphQL query and mutation testing

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details on:

- Code style and standards
- Testing requirements
- Pull request process
- Issue reporting
- Feature requests

## 📄 License

[Add your license here]

---

**Robogo** - Modern test automation for the Go ecosystem, with powerful SWIFT message generation, comprehensive API testing, and advanced Test Data Management capabilities. Built for financial services, API testing, and enterprise automation needs. 