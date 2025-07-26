# Advanced Examples

Control flow, retry logic, nested operations, and complex test scenarios.

## Examples

### 08-control-flow.yaml - Conditional Execution
**Complexity:** Advanced  
**Prerequisites:** None  
**Description:** Demonstrates conditional execution using `if` statements and logical operators.

**What you'll learn:**
- Conditional step execution with `if`
- Logical operators (`==`, `!=`, `>`, `<`, etc.)
- Variable-based conditional logic
- Branching test flows

**Run it:**
```bash
./robogo run examples/09-advanced/08-control-flow.yaml
```

### 13-retry-demo.yaml - Retry with Backoff
**Complexity:** Advanced  
**Prerequisites:** None  
**Description:** Comprehensive retry logic demonstration with exponential backoff.

**What you'll learn:**
- Retry configuration options
- Exponential backoff strategies
- Retry condition specification
- Error handling with retries

**Run it:**
```bash
./robogo run examples/09-advanced/13-retry-demo.yaml
```

### 21-simple-nested-test.yaml - Nested Operations
**Complexity:** Advanced  
**Prerequisites:** None  
**Description:** Simple nested step collections for grouping related operations.

**What you'll learn:**
- Nested step collections
- Grouped operation execution
- Continue-on-error patterns
- Hierarchical test organization

**Run it:**
```bash
./robogo run examples/09-advanced/21-simple-nested-test.yaml
```

### 16-setup-teardown-demo.yaml - Lifecycle Management
**Complexity:** Advanced  
**Prerequisites:** None  
**Description:** Test lifecycle management with setup and teardown phases.

**What you'll learn:**
- Setup phase execution
- Teardown phase execution
- Resource initialization and cleanup
- Test lifecycle patterns

**Run it:**
```bash
./robogo run examples/09-advanced/16-setup-teardown-demo.yaml
```

### 14-retry-with-failures.yaml - Complex Retry Scenarios
**Complexity:** Expert  
**Prerequisites:** None  
**Description:** Advanced retry scenarios with different failure types and recovery strategies.

**What you'll learn:**
- Multiple retry strategies
- Failure type classification
- Recovery patterns
- Complex error handling

**Run it:**
```bash
./robogo run examples/09-advanced/14-retry-with-failures.yaml
```

### 20-nested-while-loop.yaml - Complex Nested Operations
**Complexity:** Expert  
**Prerequisites:** None  
**Description:** Complex nested step collections with advanced control flow.

**What you'll learn:**
- Deep nesting patterns
- Complex control flow
- Advanced step organization
- Debugging nested operations

**Run it:**
```bash
./robogo run examples/09-advanced/20-nested-while-loop.yaml
```

## Key Concepts

### Conditional Execution
```yaml
steps:
  - name: "Conditional step"
    if: "${user_role} == 'admin'"
    action: log
    args: ["Admin operation executed"]
    
  - name: "Numeric comparison"
    if: "${response_code} >= 200 && ${response_code} < 300"
    action: log
    args: ["Success response received"]
```

### Retry Configuration
```yaml
steps:
  - name: "HTTP request with retry"
    action: http
    args: ["GET", "https://api.example.com/data"]
    retry:
      attempts: 3
      delay: "2s"
      backoff: "exponential"  # or "linear", "fixed"
      retry_on: ["http_error", "timeout", "connection_error"]
    result: api_response
```

### Nested Steps
```yaml
steps:
  - name: "User management workflow"
    steps:
      - name: "Create user"
        action: http
        args: ["POST", "/users", '{"name": "test"}']
        result: create_response
        
      - name: "Verify user created"
        action: assert
        args: ["${create_response.status}", "==", "201"]
        continue: true  # Continue even if this fails
        
      - name: "Update user"
        action: http
        args: ["PUT", "/users/1", '{"name": "updated"}']
```

### Setup and Teardown
```yaml
setup:
  - name: "Initialize test data"
    action: variable
    args: ["test_id", "TEST-${uuid}"]
    
  - name: "Create test resources"
    action: http
    args: ["POST", "/test-resources", '{"id": "${test_id}"}']

steps:
  # Main test steps here
  
teardown:
  - name: "Cleanup test resources"
    action: http
    args: ["DELETE", "/test-resources/${test_id}"]
    continue: true  # Always try cleanup, even if test failed
```

## Advanced Patterns

### Error Recovery
```yaml
- name: "Operation with fallback"
  action: http
  args: ["GET", "${primary_url}"]
  retry:
    attempts: 2
    delay: "1s"
  result: primary_response
  continue: true

- name: "Fallback operation"
  if: "${primary_response.status} != 200"
  action: http
  args: ["GET", "${fallback_url}"]
  result: fallback_response
```

### Complex Conditionals
```yaml
- name: "Multi-condition check"
  if: "${env} == 'production' && ${user_type} == 'premium' && ${feature_enabled} == true"
  action: log
  args: ["Premium production feature accessed"]
```

### Dynamic Retry Conditions
```yaml
- name: "Smart retry"
  action: http
  args: ["POST", "/api/data"]
  retry:
    attempts: 5
    delay: "1s"
    backoff: "exponential"
    retry_on: ["http_5xx", "timeout"]  # Only retry on server errors and timeouts
    max_delay: "30s"
```

## Best Practices

1. **Use descriptive names** for nested step collections
2. **Include continue flags** for non-critical operations
3. **Set appropriate retry limits** to avoid infinite loops
4. **Use setup/teardown** for resource management
5. **Test both success and failure paths**
6. **Keep nesting levels reasonable** for maintainability
7. **Document complex conditional logic** in step descriptions