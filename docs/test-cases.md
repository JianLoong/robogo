# Test Case Writing Guide

Learn how to write effective, maintainable test cases using Robogo's YAML-based syntax.

## Test Case Structure

A basic Robogo test case consists of:

```yaml
testcase: "Test Case Name"
description: "Optional description of what this test does"
timeout: 30s  # Optional timeout for the entire test case
steps:
  - action: log
    args: ["Starting test"]
  - action: http_get
    args: ["https://api.example.com/health"]
    result: response
  - action: assert
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
steps:
  - action: log
    args: ["First step"]
  - action: sleep
    args: [1]

# Nested objects
steps:
  - action: http
    args:
      - "POST"
      - "https://api.example.com"
      - 
        Content-Type: "application/json"
        Authorization: "${API_TOKEN}"
```

### Variable Substitution

Use `${variable_name}` syntax to reference variables:

```yaml
testcase: "Variable Test"
steps:
  - action: get_time
    args: ["iso"]
    result: timestamp
  
  - action: log
    args: ["Current time: ${timestamp}"]
  
  - action: http_get
    args: ["${API_BASE_URL}/health"]
    result: response
```

## Step Types

### 1. Logging Steps

```yaml
steps:
  - action: log
    args: ["Simple message"]
  
  - action: log
    args: ["Formatted message with value: ${VALUE}"]
```

### 2. HTTP Request Steps

```yaml
steps:
  # Simple GET request
  - action: http_get
    args: ["https://api.example.com/users"]
    result: response
  
  # Simple POST request
  - action: http_post
    args: 
      - "https://api.example.com/users"
      - '{"name": "John Doe", "email": "john@example.com"}'
    result: response
  
  # Advanced HTTP request with mTLS
  - action: http
    args:
      - "POST"
      - "https://secure.example.com/api/users"
      - 
        cert: "${CLIENT_CERT_PATH}"
        key: "${CLIENT_KEY_PATH}"
        ca: "${CA_CERT_PATH}"
        Content-Type: "application/json"
        Authorization: "Bearer ${API_TOKEN}"
    result: response
```

### 3. Assertion Steps

```yaml
steps:
  # Simple boolean assertion
  - action: assert
    args: [true, true, "This should always be true"]
  
  # Value comparison
  - action: assert
    args: ["${response.status_code}", "==", 200]
  
  # String comparison
  - action: assert
    args: ["hello", "hello", "String comparison should pass"]
  
  # Number comparison
  - action: assert
    args: [42, 42, "Number comparison should pass"]
```

### 4. Time and Random Operations

```yaml
steps:
  # Get current time
  - action: get_time
    args: ["iso"]  # iso, datetime, date, time, timestamp, unix, unix_ms
    result: timestamp
  
  # Get random number
  - action: get_random
    args: [100]  # Generate random number 0-99
    result: random_number
  
  # Sleep for duration
  - action: sleep
    args: [1]  # Sleep for 1 second
```

### 5. String Operations

```yaml
steps:
  # Concatenate strings
  - action: concat
    args: ["Hello", " ", "World", "!"]
    result: greeting
  
  # Get string length
  - action: length
    args: ["${greeting}"]
    result: greeting_length
```

## Advanced Patterns

### 1. Variable Management

Store and reuse values across steps:

```yaml
testcase: "Variable Management Test"
steps:
  - action: get_time
    args: ["iso"]
    result: start_time
  
  - action: log
    args: ["Test started at: ${start_time}"]
  
  - action: http_get
    args: ["https://api.example.com/health"]
    result: health_response
  
  - action: assert
    args: ["${health_response.status_code}", "==", 200]
  
  - action: get_time
    args: ["iso"]
    result: end_time
  
  - action: log
    args: ["Test completed. Start: ${start_time}, End: ${end_time}"]
```

### 2. Conditional Logic

Use assertions to create conditional behavior:

```yaml
testcase: "Conditional Test"
steps:
  - action: http_get
    args: ["https://api.example.com/status"]
    result: status_response
  
  - action: assert
    args: ["${status_response.status_code}", "==", 200, "API should be healthy"]
  
  - action: log
    args: ["API is healthy, proceeding with test"]
  
  # Continue with more steps only if API is healthy
  - action: http_post
    args: 
      - "https://api.example.com/data"
      - '{"test": "data"}'
    result: post_response
```

### 3. Error Handling

Robogo automatically handles errors and stops execution:

```yaml
testcase: "Error Handling Test"
steps:
  - action: log
    args: ["Starting test with error handling"]
  
  - action: http_get
    args: ["https://invalid-url-that-will-fail.com"]
    result: response
  
  # This step will not execute if the previous step fails
  - action: log
    args: ["This message will not appear if the HTTP request fails"]
```

### 4. Complex HTTP Requests

```yaml
testcase: "Complex HTTP Test"
steps:
  - action: log
    args: ["Testing complex HTTP operations"]
  
  # GET request with custom headers
  - action: http
    args:
      - "GET"
      - "https://api.example.com/users"
      - 
        Authorization: "Bearer ${API_TOKEN}"
        Accept: "application/json"
        User-Agent: "Robogo-Test/1.0"
    result: users_response
  
  - action: assert
    args: ["${users_response.status_code}", "==", 200]
  
  # POST request with JSON body
  - action: http
    args:
      - "POST"
      - "https://api.example.com/users"
      - 
        Content-Type: "application/json"
        Authorization: "Bearer ${API_TOKEN}"
      - |
        {
          "name": "Test User",
          "email": "test@example.com",
          "role": "tester"
        }
    result: create_response
  
  - action: assert
    args: ["${create_response.status_code}", "==", 201]
```

## Best Practices

### 1. Naming Conventions

```yaml
# Good: Descriptive test case names
testcase: "API User Creation Test"
testcase: "Database Connection Validation"
testcase: "Authentication Flow Test"

# Good: Clear step descriptions
steps:
  - action: log
    args: ["Starting user creation test"]
  - action: http_post
    args: ["${API_URL}/users", '{"name": "John"}']
    result: user_response
```

### 2. Variable Organization

```yaml
testcase: "Organized Variable Test"
steps:
  # Group related operations
  - action: get_time
    args: ["iso"]
    result: test_start_time
  
  - action: log
    args: ["Test started at: ${test_start_time}"]
  
  # Use descriptive variable names
  - action: http_get
    args: ["${API_BASE_URL}/health"]
    result: health_check_response
  
  - action: assert
    args: ["${health_check_response.status_code}", "==", 200]
```

### 3. Error Messages

```yaml
steps:
  # Good: Descriptive error messages
  - action: assert
    args: ["${response.status_code}", "==", 200, "API should return 200 OK"]
  
  - action: assert
    args: ["${user_count}", ">", 0, "User count should be positive"]
  
  # Good: Context in log messages
  - action: log
    args: ["API health check passed with status: ${response.status_code}"]
```

### 4. Test Structure

```yaml
testcase: "Well-Structured Test"
description: "Test API user management functionality"
steps:
  # Setup phase
  - action: log
    args: ["Setting up test environment"]
  
  - action: http_get
    args: ["${API_URL}/health"]
    result: health_response
  
  - action: assert
    args: ["${health_response.status_code}", "==", 200, "API should be healthy"]
  
  # Test phase
  - action: log
    args: ["Creating test user"]
  
  - action: http_post
    args: 
      - "${API_URL}/users"
      - '{"name": "Test User", "email": "test@example.com"}'
    result: create_response
  
  - action: assert
    args: ["${create_response.status_code}", "==", 201, "User should be created successfully"]
  
  # Cleanup phase
  - action: log
    args: ["Test completed successfully"]
```

## Common Patterns

### 1. API Testing

```yaml
testcase: "API End-to-End Test"
steps:
  - action: log
    args: ["Starting API end-to-end test"]
  
  # Health check
  - action: http_get
    args: ["${API_URL}/health"]
    result: health_response
  
  - action: assert
    args: ["${health_response.status_code}", "==", 200]
  
  # Create resource
  - action: http_post
    args: 
      - "${API_URL}/resources"
      - '{"name": "Test Resource"}'
    result: create_response
  
  - action: assert
    args: ["${create_response.status_code}", "==", 201]
  
  # Retrieve resource
  - action: http_get
    args: ["${API_URL}/resources/1"]
    result: get_response
  
  - action: assert
    args: ["${get_response.status_code}", "==", 200]
```

### 2. Data Validation

```yaml
testcase: "Data Validation Test"
steps:
  - action: http_get
    args: ["${API_URL}/users"]
    result: users_response
  
  - action: assert
    args: ["${users_response.status_code}", "==", 200]
  
  # Validate response structure (basic)
  - action: assert
    args: ["${users_response.body}", "contains", "users"]
  
  - action: log
    args: ["Data validation completed"]
```

### 3. Performance Testing

```yaml
testcase: "Performance Test"
steps:
  - action: get_time
    args: ["unix_ms"]
    result: start_time
  
  - action: http_get
    args: ["${API_URL}/heavy-operation"]
    result: response
  
  - action: get_time
    args: ["unix_ms"]
    result: end_time
  
  - action: assert
    args: ["${response.status_code}", "==", 200]
  
  # Calculate duration (basic)
  - action: log
    args: ["Operation completed in ${end_time} - ${start_time} ms"]
```

## Troubleshooting

### Common Issues

**"unknown action" error**
- Check action name spelling
- Use `./robogo list` to see available actions
- Ensure correct YAML syntax

**Variable not found**
- Check variable name spelling
- Ensure variable is set before use
- Use `${variable_name}` syntax

**HTTP request failures**
- Verify URL is accessible
- Check network connectivity
- For mTLS, ensure certificate paths are correct

**Assertion failures**
- Check expected vs actual values
- Verify data types match
- Use descriptive error messages

For more help, see the [Troubleshooting Guide](troubleshooting.md). 