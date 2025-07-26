# Test Examples

This directory contains comprehensive test case examples demonstrating all features and capabilities of the Robogo test automation framework.

## Quick Start Examples

### Basic HTTP Testing
- **`01-http-get.yaml`** - Simple HTTP GET request
- **`02-http-post.yaml`** - HTTP POST with JSON data  
- **`02-http-post-with-json-build.yaml`** - HTTP POST using json_build action

### Database Testing
- **`03-postgres-basic.yaml`** - Basic PostgreSQL queries
- **`03-postgres-secure.yaml`** - PostgreSQL with environment variables
- **`04-postgres-advanced.yaml`** - Advanced database operations
- **`06-spanner-basic.yaml`** - Google Cloud Spanner queries
- **`07-spanner-advanced.yaml`** - Advanced Spanner operations

### Messaging Systems
- **`05-kafka-basic.yaml`** - Kafka producer/consumer operations
- **`10-swift-mt103.yaml`** - SWIFT financial messaging

## Feature Categories

### Core Functionality
- **`00-util.yaml`** - Utility actions (UUID, time, variables)
- **`08-control-flow.yaml`** - Conditional logic and control flow
- **`19-continue-on-error.yaml`** - Error handling with continue flags

### File Operations
- **`13-file-read-basic.yaml`** - Basic file reading operations
- **`14-file-read-practical.yaml`** - Practical file processing examples
- **`23-scp-simple-test.yaml`** - Simple SCP file transfer test
- **`24-scp-validation.yaml`** - SCP parameter validation and error handling
- **`25-scp-download-test.yaml`** - SCP upload/download round-trip test

### Data Processing
- **`17-xml-operations.yaml`** - XML parsing and manipulation
- **`json-build-comparison.yaml`** - JSON construction examples
- **`35-csv-parsing.yaml`** - CSV parsing and extraction with comprehensive examples

### String Operations
- **`15-string-simple.yaml`** - Basic string operations
- **`15-string-random-basic.yaml`** - Random string generation
- **`16-string-practical.yaml`** - Practical string manipulation
- **`16-string-practical-simple.yaml`** - Simplified string examples

### Encoding and Utilities
- **`11-encoding-basic.yaml`** - Base64, URL encoding operations
- **`12-encoding-practical.yaml`** - Practical encoding examples
- **`09-sleep-timing.yaml`** - Sleep and timing operations
- **`10-sleep-practical.yaml`** - Practical timing examples
- **`11-sleep-errors.yaml`** - Sleep with error scenarios

### Advanced Features

#### Retry Logic
- **`12-retry-scenarios.yaml`** - Various retry configurations
- **`13-retry-demo.yaml`** - Retry demonstration
- **`14-retry-with-failures.yaml`** - Retry with different failure types
- **`15-retry-success-demo.yaml`** - Successful retry examples
- **`30-retry-on-errors.yaml`** - Retry on specific error types

#### Nested Steps and Control Flow
- **`20-nested-while-loop.yaml`** - Nested step collections
- **`21-simple-nested-test.yaml`** - Simple nested operations
- **`22-debug-while-nested.yaml`** - Debugging nested operations

#### Setup and Teardown
- **`16-setup-teardown-demo.yaml`** - Lifecycle management

#### Data Extraction
- **`26-fixed-extraction.yaml`** - Data extraction from responses
- **`27-retry-extraction-fixed.yaml`** - Extraction with retry logic
- **`28-plain-text-extraction.yaml`** - Plain text data extraction
- **`29-database-extraction.yaml`** - Database result extraction
- **`31-kafka-extraction.yaml`** - Kafka message extraction

### Security Features
- **`19-no-log-security.yaml`** - No-log security for sensitive operations
- **`20-step-level-masking.yaml`** - Step-level sensitive data masking
- **`20-log-formatting.yaml`** - Secure log formatting

### Network Testing
- **`26-ping-network-test.yaml`** - ICMP ping connectivity testing
- **`34-ssl-cert-check.yaml`** - SSL certificate validation and security testing

### Environment and Configuration
- **`17-env-var-test.yaml`** - Environment variable usage
- **`18-test-env-missing.yaml`** - Missing environment variable handling

### Integration Testing
- **`09-e2e-integration.yaml`** - End-to-end integration test

## Test File Structure

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
    args: ["${api_response.data}", ".users[0].name"]
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

## Running Examples

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

### Run Individual Tests
```bash
# Basic HTTP test
./robogo run examples/01-http-get.yaml

# Database test
./robogo run examples/03-postgres-basic.yaml

# SCP file transfer test
./robogo run examples/23-scp-simple-test.yaml
```

### Run with Environment Variables
```bash
# Using .env file (recommended)
cp .env.example .env
# Edit .env with your values
./robogo run examples/03-postgres-secure.yaml

# Using explicit environment variables
export DB_HOST=localhost
export DB_USER=robogo_testuser  
export DB_PASSWORD=robogo_testpass
./robogo run examples/03-postgres-secure.yaml
```

## Example Categories by Complexity

### Beginner (Simple Actions)
- `01-http-get.yaml` - Single HTTP request
- `00-util.yaml` - Basic utility actions
- `13-file-read-basic.yaml` - File reading

### Intermediate (Multiple Steps)
- `02-http-post.yaml` - HTTP with JSON data
- `03-postgres-basic.yaml` - Database queries
- `16-setup-teardown-demo.yaml` - Lifecycle management
- `26-ping-network-test.yaml` - Network connectivity testing

### Advanced (Complex Flows)
- `09-e2e-integration.yaml` - Full integration test
- `14-retry-with-failures.yaml` - Complex retry scenarios
- `22-debug-while-nested.yaml` - Nested operations with debugging
- `35-csv-parsing.yaml` - Comprehensive CSV processing with 35 test steps

### Expert (Security & Production)
- `19-no-log-security.yaml` - Security-aware testing
- `25-scp-download-test.yaml` - Secure file operations
- `31-kafka-extraction.yaml` - Message processing with extraction
- `34-ssl-cert-check.yaml` - SSL certificate validation and security analysis

## Best Practices Demonstrated

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

## Contributing Examples

When adding new examples:

1. **Follow naming convention**: `##-descriptive-name.yaml`
2. **Include description**: Add clear testcase description
3. **Document prerequisites**: Note any required services or setup
4. **Test thoroughly**: Ensure example works with standard setup
5. **Update this README**: Add to appropriate category
6. **Security check**: No hardcoded secrets or sensitive data

## Common Issues and Solutions

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