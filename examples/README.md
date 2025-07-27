# Test Examples

This directory contains comprehensive test case examples demonstrating all features and capabilities of the Robogo test automation framework. Examples are organized by category for easy navigation and learning.

## 📁 Directory Structure

| Category | Directory | Description | Examples Count |
|----------|-----------|-------------|----------------|
| **Basics** | [`01-basics/`](01-basics/) | Fundamental operations and utilities | 1 |
| **HTTP** | [`02-http/`](02-http/) | HTTP requests, REST APIs, TLS handling | 5 |
| **Database** | [`03-database/`](03-database/) | PostgreSQL, MongoDB, Spanner, data extraction | 7 |
| **Messaging** | [`04-messaging/`](04-messaging/) | Kafka, SWIFT, message processing | 4 |
| **Files** | [`05-files/`](05-files/) | File operations, SCP transfers | 3 |
| **Data Processing** | [`06-data-processing/`](06-data-processing/) | JSON, XML, CSV parsing and extraction | 7 |
| **Strings & Encoding** | [`07-strings-encoding/`](07-strings-encoding/) | String manipulation, encoding operations | 6 |
| **Utilities** | [`08-utilities/`](08-utilities/) | Sleep, timing, logging utilities | 4 |
| **Advanced** | [`09-advanced/`](09-advanced/) | Control flow, retry logic, nested operations, summary filtering | 11 |
| **Security** | [`10-security/`](10-security/) | Environment variables, data masking | 4 |
| **Network** | [`11-network/`](11-network/) | Network testing, SSL certificates, TCP connectivity | 3 |
| **Integration** | [`12-integration/`](12-integration/) | End-to-end integration tests | 1 |

**Total Examples: 56**

## 🚀 Quick Start Guide

### Prerequisites
Start the development services:
```bash
# Start all services
docker-compose up -d

# Services available:
# - PostgreSQL: localhost:5432
# - Kafka: localhost:9092
# - Spanner Emulator: localhost:9010  
# - HTTPBin: localhost:8000
# - SSH Server: localhost:2222 (user: testuser, pass: testpass)
```

### Recommended Learning Path

#### 1. Start with Basics (No services required)
```bash
# Basic utilities and operations
./robogo run examples/01-basics/00-util.yaml

# Simple HTTP requests
./robogo run examples/02-http/01-http-get.yaml
./robogo run examples/02-http/02-http-post.yaml
```

#### 2. Learn Data Processing
```bash
# JSON and XML operations
./robogo run examples/06-data-processing/json-build-comparison.yaml
./robogo run examples/06-data-processing/17-xml-operations.yaml

# CSV parsing
./robogo run examples/06-data-processing/35-csv-parsing.yaml
```

#### 3. Explore Advanced Features
```bash
# Control flow and conditionals
./robogo run examples/09-advanced/08-control-flow.yaml

# Retry logic
./robogo run examples/09-advanced/13-retry-demo.yaml

# Nested operations
./robogo run examples/09-advanced/21-simple-nested-test.yaml
```

#### 4. Security and Production Features
```bash
# Environment variables
./robogo run examples/10-security/17-env-var-test.yaml

# Data masking
./robogo run examples/10-security/19-no-log-security.yaml
```

#### 5. Service Integration (Requires docker-compose)
```bash
# Database operations
./robogo run examples/03-database/03-postgres-basic.yaml

# Messaging systems
./robogo run examples/04-messaging/05-kafka-basic.yaml

# File transfers
./robogo run examples/05-files/23-scp-simple-test.yaml
```

## 📋 Category Details

### 01-basics/ - Fundamental Operations
Essential building blocks for all tests.

| File | Description | Complexity |
|------|-------------|------------|
| `00-util.yaml` | UUID generation, variables, basic logging | Beginner |

### 02-http/ - HTTP Testing
HTTP requests, REST APIs, and TLS handling.

| File | Description | Complexity |
|------|-------------|------------|
| `01-http-get.yaml` | Simple HTTP GET with response validation | Beginner |
| `02-http-post.yaml` | HTTP POST with JSON data | Beginner |
| `02-http-post-with-json-build.yaml` | HTTP POST using json_build action | Intermediate |
| `36-http-skip-tls.yaml` | HTTP with TLS verification disabled | Intermediate |
| `37-http-tls-validation.yaml` | HTTP with strict TLS validation | Intermediate |

### 03-database/ - Database Operations
PostgreSQL, Google Cloud Spanner, and data extraction.

| File | Description | Complexity |
|------|-------------|------------|
| `03-postgres-basic.yaml` | Basic PostgreSQL queries | Beginner |
| `03-postgres-secure.yaml` | PostgreSQL with environment variables | Intermediate |
| `04-postgres-advanced.yaml` | Advanced database operations | Advanced |
| `06-spanner-basic.yaml` | Google Cloud Spanner queries | Intermediate |
| `07-spanner-advanced.yaml` | Advanced Spanner operations | Advanced |
| `29-database-extraction.yaml` | Database result extraction patterns | Advanced |
| `40-mongodb-basic.yaml` | Basic MongoDB operations (insert, find, update, delete) | Intermediate |
| `41-mongodb-advanced.yaml` | Advanced MongoDB queries, aggregations, complex operations | Advanced |

### 04-messaging/ - Messaging Systems
Kafka, SWIFT, and message processing.

| File | Description | Complexity |
|------|-------------|------------|
| `05-kafka-basic.yaml` | Kafka producer/consumer operations | Intermediate |
| `10-swift-mt103.yaml` | SWIFT financial messaging (MT103) | Advanced |
| `31-kafka-extraction.yaml` | Kafka message data extraction | Advanced |
| `32-kafka-list-topics.yaml` | Kafka topic management | Intermediate |
| `33-swift-dynamic-date.yaml` | SWIFT with dynamic date generation | Advanced |

### 05-files/ - File Operations
File reading, SCP transfers, and file validation.

| File | Description | Complexity |
|------|-------------|------------|
| `13-file-read-basic.yaml` | Basic file reading operations | Beginner |
| `14-file-read-practical.yaml` | Practical file processing examples | Intermediate |
| `23-scp-simple-test.yaml` | Simple SCP file transfer test | Intermediate |
| `24-scp-validation.yaml` | SCP parameter validation and error handling | Advanced |
| `25-scp-download-test.yaml` | SCP upload/download round-trip test | Advanced |

### 06-data-processing/ - Data Processing
JSON, XML, CSV parsing and data extraction.

| File | Description | Complexity |
|------|-------------|------------|
| `17-xml-operations.yaml` | XML parsing and manipulation | Intermediate |
| `json-build-comparison.yaml` | JSON construction examples | Intermediate |
| `35-csv-parsing.yaml` | Comprehensive CSV processing (35 steps) | Expert |
| `26-fixed-extraction.yaml` | Data extraction from responses | Advanced |
| `27-retry-extraction-fixed.yaml` | Extraction with retry logic | Advanced |
| `28-plain-text-extraction.yaml` | Plain text data extraction | Intermediate |
| `test-data.csv` | Sample CSV data for testing | - |

### 07-strings-encoding/ - String Operations
String manipulation, encoding, and random generation.

| File | Description | Complexity |
|------|-------------|------------|
| `11-encoding-basic.yaml` | Base64, URL encoding operations | Beginner |
| `12-encoding-practical.yaml` | Practical encoding examples | Intermediate |
| `15-string-simple.yaml` | Basic string operations | Beginner |
| `15-string-random-basic.yaml` | Random string generation | Beginner |
| `16-string-practical.yaml` | Practical string manipulation | Intermediate |
| `16-string-practical-simple.yaml` | Simplified string examples | Beginner |

### 08-utilities/ - Utility Operations
Sleep, timing, and logging utilities.

| File | Description | Complexity |
|------|-------------|------------|
| `09-sleep-timing.yaml` | Sleep and timing operations | Beginner |
| `10-sleep-practical.yaml` | Practical timing examples | Intermediate |
| `11-sleep-errors.yaml` | Sleep with error scenarios | Intermediate |
| `20-log-formatting.yaml` | Secure log formatting | Intermediate |

### 09-advanced/ - Advanced Features
Control flow, retry logic, nested operations, and complex scenarios.

| File | Description | Complexity |
|------|-------------|------------|
| `08-control-flow.yaml` | Conditional execution with if statements | Advanced |
| `12-retry-scenarios.yaml` | Various retry configurations | Advanced |
| `13-retry-demo.yaml` | Retry demonstration with backoff | Advanced |
| `14-retry-with-failures.yaml` | Retry with different failure types | Expert |
| `15-retry-success-demo.yaml` | Successful retry examples | Advanced |
| `16-setup-teardown-demo.yaml` | Lifecycle management | Advanced |
| `19-continue-on-error.yaml` | Error handling with continue flags | Advanced |
| `20-nested-while-loop.yaml` | Nested step collections | Expert |
| `21-simple-nested-test.yaml` | Simple nested operations | Advanced |
| `22-debug-while-nested.yaml` | Debugging nested operations | Expert |
| `30-retry-on-errors.yaml` | Retry on specific error types | Advanced |
| `39-summary-filtering-test.yaml` | Summary filtering with `summary: false` option | Advanced |

### 10-security/ - Security Features
Environment variables, data masking, and secure operations.

| File | Description | Complexity |
|------|-------------|------------|
| `17-env-var-test.yaml` | Environment variable usage | Intermediate |
| `18-test-env-missing.yaml` | Missing environment variable handling | Intermediate |
| `19-no-log-security.yaml` | No-log security for sensitive operations | Advanced |
| `20-step-level-masking.yaml` | Step-level sensitive data masking | Advanced |

### 11-network/ - Network Testing
Network connectivity, SSL certificates, and network validation.

| File | Description | Complexity |
|------|-------------|------------|
| `26-ping-network-test.yaml` | ICMP ping connectivity testing | Intermediate |
| `34-ssl-cert-check.yaml` | SSL certificate validation and security | Advanced |
| `38-tcp-connect-test.yaml` | TCP connectivity testing with timeout handling | Intermediate |

### 12-integration/ - Integration Testing
End-to-end integration tests combining multiple systems.

| File | Description | Complexity |
|------|-------------|------------|
| `09-e2e-integration.yaml` | Full integration test workflow | Expert |

## 🎯 Examples by Complexity Level

### Beginner (Simple Actions)
Perfect for getting started with Robogo:
- `01-basics/00-util.yaml` - Basic utility actions
- `02-http/01-http-get.yaml` - Single HTTP request
- `05-files/13-file-read-basic.yaml` - File reading
- `07-strings-encoding/11-encoding-basic.yaml` - Basic encoding
- `08-utilities/09-sleep-timing.yaml` - Sleep operations

### Intermediate (Multiple Steps)
Building more complex workflows:
- `02-http/02-http-post.yaml` - HTTP with JSON data
- `03-database/03-postgres-basic.yaml` - Database queries
- `04-messaging/05-kafka-basic.yaml` - Kafka operations
- `06-data-processing/17-xml-operations.yaml` - XML processing
- `10-security/17-env-var-test.yaml` - Environment variables

### Advanced (Complex Flows)
Sophisticated test scenarios:
- `09-advanced/08-control-flow.yaml` - Conditional logic
- `09-advanced/13-retry-demo.yaml` - Retry mechanisms
- `09-advanced/16-setup-teardown-demo.yaml` - Lifecycle management
- `10-security/19-no-log-security.yaml` - Security-aware testing
- `11-network/34-ssl-cert-check.yaml` - SSL validation

### Expert (Production-Ready)
Complex, production-ready test patterns:
- `06-data-processing/35-csv-parsing.yaml` - Comprehensive CSV processing (35 steps)
- `09-advanced/14-retry-with-failures.yaml` - Complex retry scenarios
- `09-advanced/20-nested-while-loop.yaml` - Nested operations
- `09-advanced/22-debug-while-nested.yaml` - Advanced debugging
- `12-integration/09-e2e-integration.yaml` - Full integration test

## 🔧 Running Examples

### Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your values
# DB_HOST=localhost
# DB_USER=robogo_testuser
# DB_PASSWORD=robogo_testpass
# etc.
```

### Run Individual Examples
```bash
# Basic HTTP test (no services required)
./robogo run examples/02-http/01-http-get.yaml

# Database test (requires docker-compose up -d)
./robogo run examples/03-database/03-postgres-basic.yaml

# Security test with environment variables
export TEST_ENV_VAR="test_value"
./robogo run examples/10-security/17-env-var-test.yaml
```

### Run by Category
```bash
# Run all HTTP examples
./robogo run examples/02-http/*.yaml

# Run all basic examples
./robogo run examples/01-basics/*.yaml

# Run all security examples
./robogo run examples/10-security/*.yaml
```

## 📚 Test Structure Reference

### Basic Structure
```yaml
testcase: "Test Name"
description: "Optional description of what this test does"

variables:
  vars:
    api_url: "https://api.example.com"
    user_id: "12345"

steps:
  - name: "Step description"
    action: action_name
    args: [arg1, arg2, arg3]
    options:
      option1: value1
      option2: value2
    result: result_variable
```

### Advanced Structure with All Features
```yaml
testcase: "Advanced Test Example"
description: "Demonstrates all test features"

variables:
  vars:
    base_url: "${ENV:API_BASE_URL}"
    token: "${ENV:API_TOKEN}"

setup:
  - name: "Initialize test data"
    action: variable
    args: ["test_id", "TEST-${uuid}"]

steps:
  # Conditional execution
  - name: "Admin-only operation"
    if: "${user_role} == 'admin'"
    action: log
    args: ["Running admin operation"]
    
  # Retry with backoff
  - name: "HTTP request with retry"
    action: http
    args: ["GET", "${base_url}/data"]
    options:
      headers:
        Authorization: "Bearer ${token}"
    retry:
      attempts: 3
      delay: "2s"
      backoff: "exponential"
      retry_on: ["http_error", "timeout"]
    result: api_response
    
  # Nested steps
  - name: "Multi-step operation"
    steps:
      - name: "Step 1"
        action: log
        args: ["Executing step 1"]
      - name: "Step 2"
        action: log
        args: ["Executing step 2"]
        continue: true
        
  # Data extraction
  - name: "Extract user data"
    action: jq
    args: ["${api_response}", ".users[0].name"]
    result: user_name
    
  # Security features
  - name: "Sensitive operation"
    action: http
    args: ["POST", "${base_url}/secure"]
    options:
      json:
        password: "${ENV:USER_PASSWORD}"
    sensitive_fields: ["password"]
    no_log: true
    result: secure_result

teardown:
  - name: "Cleanup test data"
    action: log
    args: ["Test completed: ${test_id}"]
```

## 🛡️ Best Practices Demonstrated

### Variable Management
- Use environment variables for secrets (`${ENV:SECRET}`)
- Set descriptive variable names
- Use variables for reusable values

### Error Handling
- Use `continue: true` for non-critical steps
- Implement retry logic for unstable operations
- Handle missing environment variables gracefully

### Security
- Use `no_log: true` for sensitive operations
- Mark sensitive fields with `sensitive_fields`
- Keep secrets in environment variables, not YAML files

### Test Organization
- Use descriptive test and step names
- Group related operations in nested steps
- Include cleanup in teardown sections

### Data Extraction
- Use `jq` for JSON data processing
- Use `xpath` for XML data processing
- Use `csv` extract type for CSV data processing
- Store extracted data in descriptive variable names

## 🚨 Common Issues and Solutions

### Service Connection Errors
- Ensure Docker services are running: `docker-compose ps`
- Check service health: `docker-compose logs <service>`
- Verify ports are not in use by other applications

### Environment Variable Issues
- Check .env file exists and has correct values
- Verify environment variables are exported
- Use `${ENV:VAR}` syntax, not `$VAR`

### SCP/SSH Connection Issues
- Ensure SSH server is running: `docker-compose ps ssh-server`
- Check SSH server logs: `docker logs ssh-server`
- Verify SSH credentials: user=testuser, password=testpass, port=2222

### Data Extraction Failures
- Validate JSON structure with `jq` command line tool
- Check XML structure with `xmllint`
- Use proper jq/xpath syntax for complex queries

## 🤝 Contributing Examples

When adding new examples:

1. **Follow naming convention**: Place in appropriate category directory
2. **Include description**: Add clear testcase description
3. **Document prerequisites**: Note any required services or setup
4. **Test thoroughly**: Ensure example works with standard setup
5. **Update this README**: Add to appropriate category table
6. **Security check**: No hardcoded secrets or sensitive data
7. **Complexity level**: Mark as Beginner/Intermediate/Advanced/Expert

## 📖 Additional Resources

- **[Main README](../README.md)** - Project overview and installation
- **[Architecture Documentation](../internal/README.md)** - Core architecture principles
- **[Action Reference](../internal/actions/README.md)** - Complete action documentation
- **[Execution Flow](../docs/execution-flow-diagram.md)** - Visual architecture diagram
- **[Error Handling](../docs/error-failure-states-diagram.md)** - Error handling flow