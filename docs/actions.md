# Built-in Actions Reference

This document provides a comprehensive reference for all built-in actions available in Robogo.

## Overview

Robogo provides a set of built-in actions that you can use in your test cases. Each action has a specific purpose and accepts arguments to customize its behavior.

## Action Categories

### Basic Operations
- [log](#log) - Output messages to console
- [sleep](#sleep) - Pause execution for a specified duration
- [assert](#assert) - Verify conditions and values

### Time and Random
- [get_time](#get_time) - Get current timestamp with various formats
- [get_random](#get_random) - Generate random numbers

### String Operations
- [concat](#concat) - Concatenate strings
- [length](#length) - Get length of strings or arrays

### HTTP Operations
- [http](#http) - Generic HTTP requests with mTLS support
- [http_get](#http_get) - Simplified GET requests
- [http_post](#http_post) - Simplified POST requests

## Basic Operations

### log

Outputs a message to the console.

**Syntax:**
```yaml
- action: log
  args: ["message"]
```

**Arguments:**
- `message` (string): The message to output

**Example:**
```yaml
steps:
  - action: log
    args: ["Starting API test"]
  - action: log
    args: ["Test completed successfully"]
```

**Output:**
```
📝 Starting API test
📝 Test completed successfully
```

### sleep

Pauses execution for a specified duration.

**Syntax:**
```yaml
- action: sleep
  args: [duration]
```

**Arguments:**
- `duration` (number): Duration in seconds

**Example:**
```yaml
steps:
  - action: log
    args: ["Waiting for system to stabilize"]
  - action: sleep
    args: [2]
  - action: log
    args: ["Resuming test"]
```

**Output:**
```
📝 Waiting for system to stabilize
😴 Sleeping for 2s
📝 Resuming test
```

### assert

Verifies that two values are equal.

**Syntax:**
```yaml
- action: assert
  args: [expected, actual, "optional message"]
```

**Arguments:**
- `expected` (any): Expected value
- `actual` (any): Actual value
- `message` (string, optional): Custom error message

**Example:**
```yaml
steps:
  - action: assert
    args: [true, true, "Boolean comparison should pass"]
  - action: assert
    args: ["hello", "hello", "String comparison should pass"]
  - action: assert
    args: [42, 42, "Number comparison should pass"]
```

**Output:**
```
✅ Boolean comparison should pass
✅ String comparison should pass
✅ Number comparison should pass
```

## Time and Random Operations

### get_time

Gets the current timestamp with various format options.

**Syntax:**
```yaml
- action: get_time
  args: [format]
  result: variable_name
```

**Arguments:**
- `format` (string): Time format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)

**Formats:**
- `iso`: ISO 8601 format (2024-01-15T10:30:00Z)
- `datetime`: Human-readable datetime
- `date`: Date only (2024-01-15)
- `time`: Time only (10:30:00)
- `timestamp`: RFC3339 timestamp
- `unix`: Unix timestamp (seconds)
- `unix_ms`: Unix timestamp (milliseconds)

**Example:**
```yaml
steps:
  - action: get_time
    args: ["iso"]
    result: current_time
  - action: log
    args: ["Current time: ${current_time}"]
  
  - action: get_time
    args: ["unix_ms"]
    result: timestamp_ms
  - action: log
    args: ["Timestamp in milliseconds: ${timestamp_ms}"]
```

### get_random

Generates a random number.

**Syntax:**
```yaml
- action: get_random
  args: [max_value]
  result: variable_name
```

**Arguments:**
- `max_value` (number): Maximum value (exclusive). Random number will be 0 to max_value-1

**Example:**
```yaml
steps:
  - action: get_random
    args: [100]
    result: random_number
  - action: log
    args: ["Random number (0-99): ${random_number}"]
  
  - action: get_random
    args: [1000]
    result: large_random
  - action: log
    args: ["Random number (0-999): ${large_random}"]
```

## String Operations

### concat

Concatenates multiple strings together.

**Syntax:**
```yaml
- action: concat
  args: ["string1", "string2", "string3", ...]
  result: variable_name
```

**Arguments:**
- `string1, string2, ...` (strings): Strings to concatenate

**Example:**
```yaml
steps:
  - action: concat
    args: ["Hello", " ", "World", "!"]
    result: greeting
  - action: log
    args: ["Greeting: ${greeting}"]
  
  - action: concat
    args: ["User", "_", "123", "_", "profile"]
    result: filename
  - action: log
    args: ["Filename: ${filename}"]
```

**Output:**
```
📝 Greeting: Hello World!
📝 Filename: User_123_profile
```

### length

Gets the length of a string or array.

**Syntax:**
```yaml
- action: length
  args: ["string_or_array"]
  result: variable_name
```

**Arguments:**
- `string_or_array` (string or array): Input to measure

**Example:**
```yaml
steps:
  - action: length
    args: ["Hello World"]
    result: str_length
  - action: log
    args: ["String length: ${str_length}"]
  
  - action: concat
    args: ["Test", " ", "Message"]
    result: test_message
  - action: length
    args: ["${test_message}"]
    result: message_length
  - action: log
    args: ["Message length: ${message_length}"]
```

## HTTP Operations

### http

Performs generic HTTP requests with full control over method, headers, body, and mTLS configuration.

**Syntax:**
```yaml
- action: http
  args: 
    - method
    - url
    - headers_and_options
    - body (optional)
  result: variable_name
```

**Arguments:**
- `method` (string): HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
- `url` (string): Request URL
- `headers_and_options` (object): Headers and optional mTLS configuration
- `body` (string, optional): Request body for POST/PUT/PATCH requests

**mTLS Options:**
- `cert`: Client certificate file path or PEM content
- `key`: Client private key file path or PEM content
- `ca`: CA certificate file path or PEM content

**Example:**
```yaml
steps:
  # Simple GET request
  - action: http
    args:
      - "GET"
      - "https://api.example.com/users"
      - 
        Authorization: "Bearer ${API_TOKEN}"
        Accept: "application/json"
    result: users_response
  
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
          "name": "John Doe",
          "email": "john@example.com"
        }
    result: create_response
  
  # mTLS request
  - action: http
    args:
      - "GET"
      - "https://secure.example.com/api/health"
      - 
        cert: "${CLIENT_CERT_PATH}"
        key: "${CLIENT_KEY_PATH}"
        ca: "${CA_CERT_PATH}"
        Authorization: "Bearer ${API_TOKEN}"
    result: secure_response
```

**Response Format:**
The response is stored as a JSON object with the following structure:
```json
{
  "status_code": 200,
  "headers": {
    "content-type": "application/json",
    "server": "nginx/1.18.0"
  },
  "body": "{\"message\": \"success\"}",
  "duration": "123.45ms"
}
```

### http_get

Performs simplified HTTP GET requests.

**Syntax:**
```yaml
- action: http_get
  args: [url]
  result: variable_name
```

**Arguments:**
- `url` (string): Request URL

**Example:**
```yaml
steps:
  - action: http_get
    args: ["https://httpbin.org/status/200"]
    result: response
  - action: log
    args: ["Status code: ${response.status_code}"]
  - action: assert
    args: ["${response.status_code}", "==", 200]
```

### http_post

Performs simplified HTTP POST requests.

**Syntax:**
```yaml
- action: http_post
  args: [url, body]
  result: variable_name
```

**Arguments:**
- `url` (string): Request URL
- `body` (string): Request body

**Example:**
```yaml
steps:
  - action: http_post
    args: 
      - "https://httpbin.org/post"
      - '{"name": "John", "email": "john@example.com"}'
    result: response
  - action: log
    args: ["Status code: ${response.status_code}"]
  - action: assert
    args: ["${response.status_code}", "==", 200]
```

## Advanced Usage Patterns

### Variable Management

Store and reuse HTTP response data:

```yaml
testcase: "API Test with Variables"
steps:
  - action: http_get
    args: ["https://api.example.com/users"]
    result: users_response
  
  - action: log
    args: ["Found ${users_response.status_code} users"]
  
  - action: http_post
    args: 
      - "https://api.example.com/users"
      - '{"name": "New User"}'
    result: create_response
  
  - action: assert
    args: ["${create_response.status_code}", "==", 201]
```

### Time-based Testing

Measure operation duration:

```yaml
testcase: "Performance Test"
steps:
  - action: get_time
    args: ["unix_ms"]
    result: start_time
  
  - action: http_get
    args: ["https://api.example.com/heavy-operation"]
    result: response
  
  - action: get_time
    args: ["unix_ms"]
    result: end_time
  
  - action: assert
    args: ["${response.status_code}", "==", 200]
  
  - action: log
    args: ["Operation took ${end_time} - ${start_time} milliseconds"]
```

### String Manipulation

Build dynamic URLs and messages:

```yaml
testcase: "Dynamic API Test"
steps:
  - action: concat
    args: ["https://api.example.com/users/", "123", "/profile"]
    result: user_url
  
  - action: http_get
    args: ["${user_url}"]
    result: profile_response
  
  - action: concat
    args: ["User profile status: ", "${profile_response.status_code}"]
    result: status_message
  
  - action: log
    args: ["${status_message}"]
```

## Error Handling

All actions will stop test execution if they fail:

```yaml
testcase: "Error Handling Example"
steps:
  - action: log
    args: ["Starting test"]
  
  - action: http_get
    args: ["https://invalid-url-that-will-fail.com"]
    result: response
  
  # This step will not execute if the HTTP request fails
  - action: log
    args: ["This message will not appear"]
```

## Best Practices

### 1. Use Descriptive Variable Names

```yaml
# Good
- action: http_get
  args: ["https://api.example.com/users"]
  result: users_list_response

# Avoid
- action: http_get
  args: ["https://api.example.com/users"]
  result: r
```

### 2. Provide Meaningful Log Messages

```yaml
# Good
- action: log
  args: ["Starting user creation test with API token"]

# Avoid
- action: log
  args: ["Starting test"]
```

### 3. Use Assertions with Clear Messages

```yaml
# Good
- action: assert
  args: ["${response.status_code}", "==", 200, "API should return 200 OK"]

# Avoid
- action: assert
  args: ["${response.status_code}", "==", 200]
```

### 4. Group Related Operations

```yaml
testcase: "Well-Organized Test"
steps:
  # Setup
  - action: log
    args: ["Setting up test environment"]
  - action: get_time
    args: ["iso"]
    result: test_start_time
  
  # Test operations
  - action: http_get
    args: ["${API_URL}/health"]
    result: health_check
  - action: assert
    args: ["${health_check.status_code}", "==", 200]
  
  # Cleanup
  - action: log
    args: ["Test completed at ${test_start_time}"]
```

## Troubleshooting

### Common Issues

**"unknown action" error**
- Check action name spelling
- Use `./robogo list` to see available actions
- Ensure correct YAML syntax

**HTTP request failures**
- Verify URL is accessible
- Check network connectivity
- For mTLS, ensure certificate paths are correct
- Verify headers are properly formatted

**Variable substitution issues**
- Use `${variable_name}` syntax
- Ensure variables are set before use
- Check for typos in variable names

**Assertion failures**
- Check expected vs actual values
- Verify data types match
- Use descriptive error messages

For more help, see the [Troubleshooting Guide](troubleshooting.md). 