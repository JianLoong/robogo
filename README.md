# Robogo - Modern Test Automation Framework

A modern, git-driven test automation framework written in Go, designed for comprehensive API testing, SWIFT message generation, and automated validation.

## ✨ Key Features

- **🔧 Template-based SWIFT Message Generation** - Create and test SWIFT messages with dynamic variable substitution
- **🌐 HTTP API Testing** - Full HTTP support with mTLS, custom headers, and comprehensive response validation
- **💾 Database Integration** - PostgreSQL operations with connection pooling and secure credential management
- **🎲 Enhanced Random Generation** - Support for both integer and decimal random values
- **🔄 Control Flow** - If statements, for loops, while loops with conditional logic
- **🔐 Secret Management** - Secure handling of API keys, certificates, and sensitive data
- **📊 Multiple Output Formats** - Console, JSON, and Markdown reporting
- **⚡ Performance Testing** - Built-in timing and load testing capabilities

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

# Run decimal random testing
./robogo.exe run tests/test-random-decimals.robogo
```

## 📋 Available Actions

### Basic Operations
- **`log`** - Output messages to console
- **`sleep`** - Pause execution for specified duration
- **`assert`** - Verify conditions with comparison operators

### Time and Random
- **`get_time`** - Get current timestamp (iso, datetime, date, time, unix, unix_ms)
- **`get_random`** - Generate random numbers (integers and decimals)

### String Operations
- **`concat`** - Concatenate multiple strings
- **`length`** - Get length of strings or arrays

### HTTP Operations
- **`http`** - Generic HTTP requests with mTLS support
- **`http_get`** - Simplified GET requests
- **`http_post`** - Simplified POST requests

### Database Operations
- **`postgres`** - PostgreSQL operations (query, execute, connect, close)

### Control Flow
- **`if`** - Conditional execution
- **`for`** - Loop execution (ranges, arrays, counts)
- **`while`** - Conditional loops

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

Support for both integer and decimal random values:

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
description: "Test with variables, secrets, and control flow"

variables:
  vars:
    api_url: "https://api.example.com"
    timeout: 30
  secrets:
    api_key:
      file: "secret.txt"
      mask_output: true

steps:
  # Control flow with loops
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

- **`examples/sample.robogo`** - Basic functionality demonstration
- **`tests/test-swift-working.robogo`** - SWIFT message generation and testing
- **`tests/test-random-decimals.robogo`** - Enhanced random number generation
- **`tests/test-control-flow.robogo`** - Control flow features
- **`tests/test-http.robogo`** - HTTP API testing examples
- **`tests/test-postgres.robogo`** - Database operations

## 🏗️ Project Structure

```
robogo/
├── cmd/robogo/          # CLI entry point
├── internal/           # Core framework code
│   ├── actions/        # Built-in actions
│   ├── parser/         # YAML parsing and test execution
│   └── runner/         # Test orchestration
├── tests/              # Comprehensive test examples
├── examples/           # Basic examples
├── docs/              # Documentation
└── prd/               # Product requirements
```

## 🔧 Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run specific test
./robogo.exe run tests/test-swift-working.robogo
```

### Build

```bash
go build -o robogo.exe ./cmd/robogo
```

### List Available Actions

```bash
./robogo.exe list
```

## 📊 Output Formats

Robogo supports multiple output formats:

```bash
# Console output (default)
./robogo.exe run test.robogo

# JSON output
./robogo.exe run test.robogo --output json

# Markdown output
./robogo.exe run test.robogo --output markdown
```

## 📚 Documentation

Comprehensive documentation available in the [docs/](docs/) directory:

- [Actions Reference](docs/actions.md) - Complete list of available actions
- [Quick Start Guide](docs/quickstart.md) - Get started quickly
- [Test Cases Guide](docs/test-cases.md) - Writing effective test cases
- [CLI Reference](docs/cli-reference.md) - Command-line interface documentation
- [Contributing Guide](docs/CONTRIBUTING.md) - How to contribute

## 🎯 Use Cases

### Financial Services
- **SWIFT Message Testing** - Generate and validate SWIFT messages
- **Payment API Testing** - Test payment processing systems
- **Banking Integration** - Validate banking APIs and workflows

### API Testing
- **REST API Validation** - Comprehensive HTTP API testing
- **mTLS Security Testing** - Test secure API endpoints
- **Performance Testing** - Load and performance validation

### Database Testing
- **PostgreSQL Operations** - Database query and transaction testing
- **Data Validation** - Verify database state and results
- **Integration Testing** - End-to-end database workflows

## 🚀 Roadmap

- [ ] **Plugin System** - Custom action development
- [ ] **Parallel Execution** - Concurrent test execution
- [ ] **Web Interface** - Browser-based test management
- [ ] **Advanced Reporting** - Detailed analytics and dashboards
- [ ] **Cloud Integration** - AWS, Azure, GCP support
- [ ] **CI/CD Integration** - Jenkins, GitHub Actions, GitLab CI

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## 📄 License

[Add your license here]

---

**Robogo** - Modern test automation for the Go ecosystem, with powerful SWIFT message generation and comprehensive API testing capabilities. 