# Constants Package

This package defines all constants used throughout the Robogo framework. It provides a centralized location for framework-wide enumerations, operation codes, and configuration values, following the KISS principle with clear, organized constant definitions.

## Components

### ⚙️ **Execution Constants** (`execution.go`)

Core execution constants that define action states, operations, and comparison operators.

#### **Action Status Enumeration**
Defines the four possible outcomes for any step execution:

```go
type ActionStatus string

const (
    ActionStatusPassed  ActionStatus = "PASS"    // ✅ Action completed successfully
    ActionStatusFailed  ActionStatus = "FAIL"    // ❌ Logical test failure (FailureInfo)
    ActionStatusError   ActionStatus = "ERROR"   // ❌ Technical problem (ErrorInfo)  
    ActionStatusSkipped ActionStatus = "SKIPPED" // ⏭️ Step bypassed (conditional logic)
)
```

**Usage in Framework:**
```go
// Action implementations return these statuses
return types.ActionResult{
    Status: constants.ActionStatusPassed,
    Data:   resultData,
}

// Runner uses statuses for flow control
if result.Status == constants.ActionStatusError {
    // Handle technical errors
}
```

#### **Comparison Operators**
Operators used by the `assert` action for value comparisons:

```go
const (
    OperatorEqual              = "=="       // Exact equality
    OperatorNotEqual           = "!="       // Inequality  
    OperatorGreaterThan        = ">"        // Numeric greater than
    OperatorLessThan           = "<"        // Numeric less than
    OperatorGreaterThanOrEqual = ">="       // Greater than or equal
    OperatorLessThanOrEqual    = "<="       // Less than or equal
    OperatorContains           = "contains" // String/array contains
    OperatorStartsWith         = "starts_with" // String prefix match
    OperatorEndsWith           = "ends_with"   // String suffix match
)
```

**Usage in Tests:**
```yaml
steps:
  - name: "Verify status code"
    action: assert
    args: ["${status_code}", "==", "200"]
    
  - name: "Check response time"  
    action: assert
    args: ["${response_time}", "<", "1000"]
    
  - name: "Validate message content"
    action: assert
    args: ["${message}", "contains", "success"]
```

#### **HTTP Method Constants**
Standard HTTP methods supported by the `http` action:

```go
const (
    HTTPGet    = "GET"
    HTTPPost   = "POST" 
    HTTPPut    = "PUT"
    HTTPPatch  = "PATCH"
    HTTPDelete = "DELETE"
    HTTPHead   = "HEAD"
)
```

**Usage Example:**
```yaml
steps:
  - name: "Create user"
    action: http
    args: ["POST", "${api_url}/users", "${user_data}"]
    
  - name: "Update user"
    action: http  
    args: ["PUT", "${api_url}/users/${user_id}", "${updated_data}"]
```

#### **Database Operation Constants**
Operations supported by database actions (`postgres`, `spanner`):

```go
const (
    OperationQuery   = "query"   // SELECT statements and read operations
    OperationSelect  = "select"  // Alias for query (legacy support)
    OperationExecute = "execute" // DDL/DML operations (CREATE, INSERT, UPDATE, DELETE)
    OperationInsert  = "insert"  // INSERT operations (specific)
    OperationUpdate  = "update"  // UPDATE operations (specific) 
    OperationDelete  = "delete"  // DELETE operations (specific)
)
```

**Usage in Database Tests:**
```yaml
steps:
  - name: "Create test table"
    action: postgres
    args: ["execute", "${db_url}", "CREATE TABLE test_users (id SERIAL PRIMARY KEY, name VARCHAR(100))"]
    
  - name: "Insert test data"
    action: postgres
    args: ["insert", "${db_url}", "INSERT INTO test_users (name) VALUES ('Alice'), ('Bob')"]
    
  - name: "Query users"
    action: postgres
    args: ["query", "${db_url}", "SELECT * FROM test_users"]
    result: users
```

#### **Messaging Operation Constants**
Operations for messaging systems (`kafka`, `rabbitmq`):

```go
const (
    OperationPublish    = "publish"     // Send messages to topic/queue
    OperationConsume    = "consume"     // Receive messages from topic/queue
    OperationListTopics = "list_topics" // Discover available topics (Kafka)
)
```

**Usage in Messaging Tests:**
```yaml
steps:
  - name: "List available topics"
    action: kafka
    args: ["list_topics", "${kafka_broker}"]
    result: topics
    
  - name: "Publish test message"
    action: kafka
    args: ["publish", "${kafka_broker}", "test-topic", "Hello World"]
    
  - name: "Consume message"
    action: kafka
    args: ["consume", "${kafka_broker}", "test-topic"]
    result: messages
```

#### **Variable Operation Constants** 
Operations for the `variable` action:

```go
const (
    VariableOperationSet    = "set"    // Store a variable value
    VariableOperationGet    = "get"    // Retrieve a variable value
    VariableOperationList   = "list"   // List all variables
    VariableOperationDelete = "delete" // Remove a variable
    VariableOperationDebug  = "debug"  // Debug variable resolution
)
```

**Usage in Variable Management:**
```yaml
steps:
  - name: "Set test variable"
    action: variable
    args: ["set", "test_mode", "true"]
    
  - name: "Debug variables"
    action: variable
    args: ["debug"]
```

### 🔧 **Configuration Constants** (`config.go`)

Framework configuration values and defaults.

**File Organization:**
- **Timeout Values**: Default timeouts for various operations
- **Connection Limits**: Default connection pool sizes and limits
- **Buffer Sizes**: Default buffer sizes for I/O operations
- **Retry Configuration**: Default retry counts and backoff values

## Design Principles

### 📋 **Centralized Definition**
- **Single Source of Truth**: All constants defined in one package
- **Import Consistency**: `constants.ActionStatusPassed` everywhere
- **No Magic Strings**: Replace hardcoded strings with named constants

### 🎯 **Clear Naming**
- **Descriptive Names**: `OperationListTopics` instead of `LIST_TOPICS`
- **Consistent Prefixes**: All HTTP methods start with `HTTP`, all operators with `Operator`
- **Action Context**: Operation constants grouped by action type

### 🔄 **Extensibility**
- **Easy Addition**: New constants added without breaking existing code
- **Grouped Organization**: Related constants grouped in logical sections
- **Version Compatibility**: Backwards-compatible constant additions

## Usage Patterns

### In Action Implementations
```go
func httpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
    method := strings.ToUpper(fmt.Sprintf("%v", args[0]))
    
    switch method {
    case constants.HTTPGet:
        // Handle GET request
    case constants.HTTPPost:
        // Handle POST request
    default:
        return types.InvalidArgError("http", "method", "valid HTTP method")
    }
    
    return types.ActionResult{
        Status: constants.ActionStatusPassed,
        Data:   response,
    }
}
```

### In Test Runner
```go
func (r *TestRunner) executeStep(step types.Step) *types.StepResult {
    result := r.strategyRouter.Execute(step, stepNumber, nil)
    
    switch result.Status {
    case constants.ActionStatusPassed:
        // Continue execution
    case constants.ActionStatusSkipped:
        // Log skip reason and continue
    case constants.ActionStatusError, constants.ActionStatusFailed:
        // Handle error/failure based on continue flag
    }
    
    return result
}
```

### In Execution Strategies
```go
func (s *RetryExecutionStrategy) Execute(step types.Step) *types.StepResult {
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        result := s.basicStrategy.Execute(step, stepNumber, nil)
        
        if result.Status == constants.ActionStatusPassed {
            return result // Success, no retry needed
        }
        
        if attempt < maxAttempts {
            // Wait and retry
            time.Sleep(backoffDelay)
        }
    }
}
```

## Error Handling

### Constant Validation
Constants are validated at compile-time, preventing runtime errors from typos or invalid values.

### Type Safety  
Using typed constants (`ActionStatus`) instead of raw strings provides compile-time checking:

```go
// ✅ Type-safe - will catch errors at compile time
status := constants.ActionStatusPassed

// ❌ Error-prone - typos only caught at runtime  
status := "PASS"
```

### Extensibility Safety
New constants can be added without breaking existing code, as long as:
1. Existing constant values are not changed
2. New constants follow established naming patterns
3. Default cases handle unknown values gracefully

## Integration Points

### With Actions Package
- All actions import and use operation constants
- Status constants returned in `ActionResult` structures
- HTTP method constants used in `http` action

### With Execution Package
- Status constants used for flow control decisions
- Operation constants used for action routing
- Comparison operators used in assertion logic

### With Types Package  
- `ActionStatus` type defined here, used in `ActionResult`
- Constants provide valid values for type fields
- Integration maintains type safety across packages

## Contributing

### Adding New Constants
1. **Choose Appropriate File**: Add to `execution.go` for framework constants, `config.go` for configuration
2. **Follow Naming Patterns**: Use consistent prefixes and descriptive names
3. **Group Logically**: Add to appropriate constant group with related values
4. **Document Usage**: Include examples of how/where the constant is used

### Deprecating Constants
1. **Mark as Deprecated**: Add deprecation comments
2. **Provide Alternatives**: Document replacement constants
3. **Maintain Backwards Compatibility**: Keep deprecated constants for several versions
4. **Update Documentation**: Remove deprecated constants from examples

### Best Practices
- **Use Constants Everywhere**: Replace magic strings/numbers with named constants
- **Import Consistently**: Always use `constants.ConstantName` pattern
- **Type-Safe Where Possible**: Use typed constants instead of raw primitives
- **Document Purpose**: Include comments explaining when/how constants are used

This constants package provides the foundation for consistent, type-safe values throughout the Robogo framework, supporting maintainability and reducing runtime errors.