# Test Case Writing Guide

Learn how to write effective, maintainable test cases using Gobot's YAML-based syntax.

## Test Case Structure

A basic Gobot test case consists of:

```yaml
testcase: "Test Case Name"
description: "Optional description of what this test does"
tags: ["api", "smoke", "critical"]  # Optional tags for categorization
timeout: 30s  # Optional timeout for the entire test case
steps:
  - keyword: log
    args: ["Starting test"]
  - keyword: http_request
    args:
      url: "https://api.example.com/health"
      method: "GET"
  - keyword: assert
    args: ["${response.status_code}", "==", 200]
```

## YAML Syntax

### Basic Elements

```yaml
# String values
testcase: "Simple Test"

# Multi-line strings
description: |
  This is a multi-line description
  that can span multiple lines
  for better readability

# Lists
tags:
  - api
  - smoke
  - critical

# Nested objects
steps:
  - keyword: http_request
    args:
      url: "https://api.example.com"
      method: "POST"
      headers:
        Content-Type: "application/json"
        Authorization: "${API_TOKEN}"
```

### Environment Variables

Use `${VARIABLE_NAME}` syntax to reference environment variables:

```yaml
testcase: "API Test with Environment Variables"
steps:
  - keyword: log
    args: ["Using API URL: ${API_BASE_URL}"]
  - keyword: http_request
    args:
      url: "${API_BASE_URL}/health"
      headers:
        Authorization: "Bearer ${API_TOKEN}"
```

### Default Values

Provide default values for environment variables:

```yaml
steps:
  - keyword: log
    args: ["User: ${USER:-anonymous}"]
  - keyword: http_request
    args:
      url: "${API_URL:-https://api.example.com}/health"
```

## Step Types

### 1. Logging Steps

```yaml
steps:
  - keyword: log
    args: ["Simple message"]
  
  - keyword: log
    args: ["Formatted message with value: ${VALUE}"]
  
  - keyword: log
    args: ["Debug message"]
    level: "debug"  # debug, info, warn, error
```

### 2. HTTP Request Steps

```yaml
steps:
  - keyword: http_request
    args:
      url: "https://api.example.com/users"
      method: "POST"
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer ${TOKEN}"
      body: |
        {
          "name": "John Doe",
          "email": "john@example.com"
        }
      timeout: 30
      mtls:
        client_cert: "${CLIENT_CERT_PATH}"
        client_key: "${CLIENT_KEY_PATH}"
        ca_cert: "${CA_CERT_PATH}"
```

### 3. Assertion Steps

```yaml
steps:
  # Simple boolean assertion
  - keyword: assert
    args: [true, "This should always be true"]
  
  # Value comparison
  - keyword: assert
    args: ["${response.status_code}", "==", 200]
  
  # String contains
  - keyword: assert
    args: ["${response.body}", "contains", "success"]
  
  # Numeric comparison
  - keyword: assert
    args: ["${response.time}", "<", 1000]
  
  # Custom assertion with message
  - keyword: assert
    args: ["${user_count}", ">", 0, "User count should be positive"]
```

### 4. File Operations

```yaml
steps:
  # Read file
  - keyword: file_read
    args: ["config.json"]
  
  # Write file
  - keyword: file_write
    args: ["output.txt", "Test results"]
  
  # Check file exists
  - keyword: file_exists
    args: ["important-file.txt"]
  
  # Delete file
  - keyword: file_delete
    args: ["temp-file.txt"]
```

### 5. Sleep/Delay

```yaml
steps:
  - keyword: sleep
    args: [5]  # Sleep for 5 seconds
  
  - keyword: sleep
    args: [1000]  # Sleep for 1000 milliseconds
```

## Advanced Patterns

### 1. Test Setup and Teardown

```yaml
testcase: "User Management Test"
setup:
  - keyword: log
    args: ["Setting up test environment"]
  - keyword: http_request
    args:
      url: "${API_URL}/setup"
      method: "POST"

steps:
  - keyword: http_request
    args:
      url: "${API_URL}/users"
      method: "GET"

teardown:
  - keyword: log
    args: ["Cleaning up test environment"]
  - keyword: http_request
    args:
      url: "${API_URL}/cleanup"
      method: "POST"
```

### 2. Conditional Steps

```yaml
testcase: "Conditional Test"
steps:
  - keyword: http_request
    args:
      url: "${API_URL}/status"
      method: "GET"
  
  - keyword: if
    condition: "${response.status_code} == 200"
    then:
      - keyword: log
        args: ["Service is healthy"]
      - keyword: assert
        args: ["${response.body}", "contains", "ok"]
    else:
      - keyword: log
        args: ["Service is unhealthy"]
      - keyword: assert
        args: [false, "Service should be healthy"]
```

### 3. Loops

```yaml
testcase: "Loop Test"
steps:
  - keyword: foreach
    items: ["user1", "user2", "user3"]
    as: "username"
    do:
      - keyword: log
        args: ["Processing user: ${username}"]
      - keyword: http_request
        args:
          url: "${API_URL}/users/${username}"
          method: "GET"
      - keyword: assert
        args: ["${response.status_code}", "==", 200]
```

### 4. Variable Assignment

```yaml
testcase: "Variable Test"
steps:
  - keyword: set
    args: ["base_url", "https://api.example.com"]
  
  - keyword: set
    args: ["user_id", "12345"]
  
  - keyword: http_request
    args:
      url: "${base_url}/users/${user_id}"
      method: "GET"
```

## Best Practices

### 1. Naming Conventions

```yaml
# Good: Descriptive names
testcase: "User API - Create User Successfully"
testcase: "Database - Verify User Persistence"
testcase: "Authentication - Login with Valid Credentials"

# Bad: Vague names
testcase: "Test 1"
testcase: "API Test"
testcase: "Check Something"
```

### 2. Organization

```yaml
# Group related test cases
testcases:
  - name: "User Management - Create"
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/users"
          method: "POST"
          body: '{"name": "John", "email": "john@example.com"}'
  
  - name: "User Management - Read"
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/users/1"
          method: "GET"
  
  - name: "User Management - Update"
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/users/1"
          method: "PUT"
          body: '{"name": "John Updated"}'
```

### 3. Error Handling

```yaml
testcase: "Robust API Test"
steps:
  - keyword: http_request
    args:
      url: "${API_URL}/health"
      method: "GET"
      timeout: 30
      retry_attempts: 3
      retry_delay: 5
  
  - keyword: assert
    args: ["${response.status_code}", "in", [200, 201, 202]]
    message: "Expected successful status code"
```

### 4. Documentation

```yaml
testcase: "Complex Business Logic Test"
description: |
  This test verifies the complete user registration workflow:
  1. Creates a new user account
  2. Sends verification email
  3. Verifies email confirmation
  4. Checks user permissions
  5. Validates profile data

tags:
  - critical
  - user-management
  - e2e

author: "team@example.com"
created: "2024-01-15"
last_updated: "2024-01-15"
```

### 5. Reusability

```yaml
# Create reusable test components
components:
  login:
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/login"
          method: "POST"
          body: '{"username": "${USERNAME}", "password": "${PASSWORD}"}'
      - keyword: assert
        args: ["${response.status_code}", "==", 200]

testcase: "User Dashboard Access"
steps:
  - keyword: include
    args: ["login"]
  - keyword: http_request
    args:
      url: "${API_URL}/dashboard"
      method: "GET"
      headers:
        Authorization: "Bearer ${response.body.token}"
```

## Validation

### YAML Schema Validation

Gobot validates test cases against a schema. Common validation errors:

```yaml
# ❌ Invalid: Missing required field
steps:
  - keyword: log  # Missing args

# ✅ Valid
steps:
  - keyword: log
    args: ["Message"]

# ❌ Invalid: Wrong data type
steps:
  - keyword: sleep
    args: "5"  # Should be number, not string

# ✅ Valid
steps:
  - keyword: sleep
    args: [5]
```

### Custom Validation

```yaml
testcase: "Validated Test"
validation:
  required_env_vars:
    - API_URL
    - API_TOKEN
  required_files:
    - config.json
  min_timeout: 10s
  max_timeout: 300s

steps:
  - keyword: log
    args: ["Test with validation"]
```

## Examples

### Complete API Test Suite

```yaml
# api-tests.yaml
testcases:
  - name: "API Health Check"
    description: "Verify API is responding"
    tags: ["smoke", "health"]
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/health"
          method: "GET"
      - keyword: assert
        args: ["${response.status_code}", "==", 200]
      - keyword: assert
        args: ["${response.body.status}", "==", "healthy"]

  - name: "User Authentication"
    description: "Test user login functionality"
    tags: ["auth", "critical"]
    setup:
      - keyword: set
        args: ["test_user", "test@example.com"]
      - keyword: set
        args: ["test_password", "password123"]
    
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/auth/login"
          method: "POST"
          body: |
            {
              "email": "${test_user}",
              "password": "${test_password}"
            }
      - keyword: assert
        args: ["${response.status_code}", "==", 200]
      - keyword: assert
        args: ["${response.body.token}", "!=", ""]
    
    teardown:
      - keyword: log
        args: ["Cleaning up test data"]

  - name: "Data Validation"
    description: "Test input validation"
    tags: ["validation"]
    steps:
      - keyword: http_request
        args:
          url: "${API_URL}/users"
          method: "POST"
          body: '{"email": "invalid-email"}'
      - keyword: assert
        args: ["${response.status_code}", "==", 400]
      - keyword: assert
        args: ["${response.body.error}", "contains", "Invalid email"]
```

## Next Steps

- Read the [Built-in Keywords](keywords.md) reference
- Learn about [Git Integration](git-integration.md) for team workflows
- Explore [Parallel Execution](parallel.md) for performance
- Check out [Plugin Development](plugins.md) for custom keywords 