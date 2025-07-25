# Types Package

This package defines all core data structures used throughout the Robogo framework. It provides the type definitions that represent test cases, execution results, error handling, and other fundamental concepts, following the KISS principle with clean, straightforward struct definitions.

## Components

### 📋 **Test Definition Types** (`testcase.go`, `step.go`)

Core structures that define test cases and individual test steps.

#### **TestCase Structure**
Represents a complete test case loaded from YAML:

```go
type TestCase struct {
    Name        string        `yaml:"testcase"`           // Test case name/identifier
    Description string        `yaml:"description,omitempty"` // Optional test description
    Setup       []Step        `yaml:"setup,omitempty"`    // Setup steps (run before main steps)
    Steps       []Step        `yaml:"steps"`              // Main test steps
    Teardown    []Step        `yaml:"teardown,omitempty"` // Teardown steps (always run)
    Variables   TestVariables `yaml:"variables,omitempty"` // Pre-defined variables
}

type TestVariables struct {
    Vars map[string]any `yaml:"vars,omitempty"` // Variable name-value pairs
}
```

**YAML Mapping:**
```yaml
testcase: "User Registration Test"
description: "Test user registration flow with validation"

variables:
  vars:
    api_url: "https://api.example.com"
    test_user: "testuser@example.com"

setup:
  - name: "Create test database"
    action: postgres
    args: ["execute", "${db_url}", "CREATE DATABASE testdb"]

steps:
  - name: "Register new user"
    action: http
    args: ["POST", "${api_url}/users", '{"email": "${test_user}"}']
    result: registration_response

teardown:
  - name: "Cleanup test database"  
    action: postgres
    args: ["execute", "${db_url}", "DROP DATABASE testdb"]
```

#### **Step Structure**
Represents individual test steps with all possible configuration options:

```go
type Step struct {
    Name            string            `yaml:"name"`                    // Step name (required)
    Action          string            `yaml:"action,omitempty"`        // Action to execute
    Args            []any             `yaml:"args,omitempty"`          // Action arguments
    Options         map[string]any    `yaml:"options,omitempty"`       // Action options
    Result          string            `yaml:"result,omitempty"`        // Variable to store result
    If              string            `yaml:"if,omitempty"`            // Conditional execution
    Continue        bool              `yaml:"continue,omitempty"`      // Continue on failure
    NoLog           bool              `yaml:"no_log,omitempty"`        // Suppress logging
    SensitiveFields []string          `yaml:"sensitive_fields,omitempty"` // Custom field masking
    
    // Advanced execution options
    Retry           *RetryConfig      `yaml:"retry,omitempty"`         // Retry configuration
    Steps           []Step            `yaml:"steps,omitempty"`         // Nested steps
}
```

**Step Examples:**
```yaml
# Basic step
- name: "Make HTTP request"
  action: http
  args: ["GET", "${api_url}/users"]
  result: users_response

# Conditional step  
- name: "Create user if not exists"
  action: http
  args: ["POST", "${api_url}/users", "${user_data}"]
  if: "${user_exists} == false"
  result: create_response

# Security-aware step
- name: "Authenticate user"
  action: http
  args: ["POST", "/auth", '{"password": "${password}"}']
  no_log: true
  sensitive_fields: ["password", "token"]
  result: auth_response

# Retry with backoff
- name: "Poll for completion"
  action: http
  args: ["GET", "${api_url}/status/${job_id}"]
  retry:
    max_attempts: 5
    delay: "2s"
    backoff: "exponential"
  result: job_status
```

### 🎯 **Execution Result Types** (`action_result.go`, `testresult.go`)

Structures that represent the results of action execution and complete test runs.

#### **ActionResult Structure**
Returned by every action execution:

```go
type ActionResult struct {
    Status ActionStatus `json:"status"`           // PASS, FAIL, ERROR, SKIPPED
    Data   any          `json:"data,omitempty"`   // Action-specific result data
    
    // Error information (technical problems)
    ErrorInfo *ErrorInfo `json:"error_info,omitempty"`
    
    // Failure information (logical test problems)  
    FailureInfo *FailureInfo `json:"failure_info,omitempty"`
}
```

**Usage in Actions:**
```go
func httpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
    // ... execute HTTP request ...
    
    if err != nil {
        return types.ActionResult{
            Status: constants.ActionStatusError,
            ErrorInfo: &types.ErrorInfo{
                Category: types.ErrorCategoryNetwork,
                Code:     "HTTP_REQUEST_FAILED",
                Message:  "Failed to execute HTTP request",
                // ... additional context ...
            },
        }
    }
    
    return types.ActionResult{
        Status: constants.ActionStatusPassed,
        Data:   responseData,
    }
}
```

#### **StepResult Structure**  
Represents the result of executing a complete step (including any retries, conditions, etc.):

```go
type StepResult struct {
    Name     string       `json:"name"`               // Step name
    Status   ActionStatus `json:"status"`             // Final step status
    Duration time.Duration `json:"duration"`          // Execution time
    Message  string       `json:"message,omitempty"`  // Error/failure message
    Category string       `json:"category,omitempty"` // Error category
    Data     any          `json:"data,omitempty"`     // Step result data
}
```

#### **TestResult Structure**
Complete test execution result:

```go
type TestResult struct {
    Name          string       `json:"name"`                    // Test case name
    Status        ActionStatus `json:"status"`                  // Overall test status
    Duration      time.Duration `json:"duration"`               // Total execution time
    SetupSteps    []*StepResult `json:"setup_steps,omitempty"`  // Setup step results
    Steps         []*StepResult `json:"steps"`                   // Main step results  
    TeardownSteps []*StepResult `json:"teardown_steps,omitempty"` // Teardown step results
    ErrorInfo     *ErrorInfo   `json:"error_info,omitempty"`    // First error encountered
    FailureInfo   *FailureInfo `json:"failure_info,omitempty"`  // First failure encountered
}
```

### ❌ **Error Handling Types** (`error_handling.go`, `failure_handling.go`)

Comprehensive error and failure representation with rich context and suggestions.

#### **ErrorInfo Structure** 
Technical problems that prevent proper execution:

```go
type ErrorInfo struct {
    Category    ErrorCategory     `json:"category"`              // Error classification
    Code        string           `json:"code"`                  // Specific error code  
    Message     string           `json:"message"`               // Human-readable message
    Template    string           `json:"template,omitempty"`    // Message template
    Context     map[string]any   `json:"context,omitempty"`     // Additional context
    Suggestions []string         `json:"suggestions,omitempty"` // Resolution suggestions
    Timestamp   time.Time        `json:"timestamp"`             // When error occurred
}

type ErrorCategory string
const (
    ErrorCategoryNetwork    ErrorCategory = "network"    // Connection, timeout issues
    ErrorCategoryDatabase   ErrorCategory = "database"   // DB connection, query issues  
    ErrorCategorySystem     ErrorCategory = "system"     // File, permission, OS issues
    ErrorCategoryValidation ErrorCategory = "validation" // Invalid arguments, formats
    ErrorCategoryExecution  ErrorCategory = "execution"  // Action execution problems
)
```

**Error Builder Pattern:**
```go
return types.NewErrorBuilder(types.ErrorCategoryNetwork, "CONNECTION_TIMEOUT").
    WithTemplate("Failed to connect to ${service} within ${timeout}").
    WithContext("service", "PostgreSQL").
    WithContext("timeout", "30s").
    WithContext("host", "localhost:5432").
    WithSuggestion("Check if PostgreSQL is running").
    WithSuggestion("Verify network connectivity").
    WithSuggestion("Increase timeout value if needed").
    Build("database connection timeout after 30s")
```

#### **FailureInfo Structure**
Logical test problems where execution succeeded but results were unexpected:

```go
type FailureInfo struct {
    Category    FailureCategory   `json:"category"`              // Failure classification
    Code        string           `json:"code"`                  // Specific failure code
    Message     string           `json:"message"`               // Human-readable message
    Template    string           `json:"template,omitempty"`    // Message template
    Context     map[string]any   `json:"context,omitempty"`     // Additional context
    Suggestions []string         `json:"suggestions,omitempty"` // Resolution suggestions
    Timestamp   time.Time        `json:"timestamp"`             // When failure occurred
}

type FailureCategory string
const (
    FailureCategoryAssertion FailureCategory = "assertion" // Assert action failures
    FailureCategoryResponse  FailureCategory = "response"  // Unexpected response data
    FailureCategoryValidation FailureCategory = "validation" // Data validation failures
    FailureCategoryBusiness  FailureCategory = "business"  // Business logic violations
)
```

**Failure Builder Example:**
```go
return types.NewFailureBuilder(types.FailureCategoryAssertion, "VALUE_MISMATCH").
    WithTemplate("Expected ${expected} but got ${actual}").
    WithContext("expected", 200).
    WithContext("actual", 404).  
    WithContext("field", "status_code").
    WithSuggestion("Check if the API endpoint exists").
    WithSuggestion("Verify request parameters are correct").
    Build("assertion failed: expected 200 but got 404")
```

### 🔄 **Advanced Execution Types** (`loop_context.go`)

Support structures for complex execution patterns.

#### **RetryConfig Structure**
Configuration for retry logic:

```go
type RetryConfig struct {
    MaxAttempts int           `yaml:"max_attempts"`         // Maximum retry attempts
    Delay       string        `yaml:"delay"`               // Initial delay (e.g., "1s", "500ms")
    Backoff     BackoffType   `yaml:"backoff,omitempty"`   // Backoff strategy
    RetryIf     string        `yaml:"retry_if,omitempty"`  // Condition for retrying
}

type BackoffType string
const (
    BackoffFixed       BackoffType = "fixed"       // Same delay every time
    BackoffLinear      BackoffType = "linear"      // Linearly increasing delay
    BackoffExponential BackoffType = "exponential" // Exponentially increasing delay
)
```

**Retry Examples:**
```yaml
# Fixed delay retry
retry:
  max_attempts: 3
  delay: "2s"
  backoff: "fixed"

# Exponential backoff
retry:  
  max_attempts: 5
  delay: "1s"
  backoff: "exponential"  # 1s, 2s, 4s, 8s, 16s

# Conditional retry
retry:
  max_attempts: 10
  delay: "500ms"
  retry_if: "${response_status} >= 500"  # Only retry on server errors
```

### 🛠️ **Utility Types** (`simple_errors.go`)

Convenience functions for creating common error types:

```go
// Quick error creation functions
func MissingArgsError(action string, expected, actual int) ActionResult
func InvalidArgError(action, arg, expected string) ActionResult  
func UnknownOperationError(action, operation string) ActionResult
func RequestError(operation, message string) ActionResult
func TimeoutError(message string) ActionResult
```

**Usage Examples:**
```go
// In action implementations
if len(args) < 2 {
    return types.MissingArgsError("http", 2, len(args))
}

if operation != "GET" && operation != "POST" {
    return types.UnknownOperationError("http", operation)
}

if err != nil {
    return types.RequestError("HTTP GET", err.Error())
}
```

## Design Principles

### 🎯 **Single Responsibility**
- **Clear Purpose**: Each type has one clear responsibility
- **Focused Structure**: No unnecessary fields or methods  
- **Domain Alignment**: Types match conceptual domain models

### 🔄 **Composition over Inheritance**  
- **Struct Embedding**: Used sparingly and only when beneficial
- **Interface Satisfaction**: Types satisfy interfaces through methods, not inheritance
- **Flexible Design**: Easy to extend without breaking existing code

### 📝 **YAML-First Design**
- **Tag-Driven**: YAML tags define the external API
- **Optional Fields**: Extensive use of `omitempty` for clean YAML
- **Intuitive Mapping**: YAML structure matches Go struct layout

### 🛡️ **Error Safety**
- **Explicit Error Types**: Clear distinction between errors and failures  
- **Rich Context**: Detailed information for debugging and resolution
- **Consistent Patterns**: Same error handling approach across all types

## Usage Patterns

### In Test Parsing
```go
func LoadTestCase(filename string) (*types.TestCase, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var testCase types.TestCase
    if err := yaml.Unmarshal(data, &testCase); err != nil {
        return nil, err
    }
    
    return &testCase, nil
}
```

### In Action Implementation
```go
func myAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
    // Validate arguments
    if len(args) < 1 {
        return types.MissingArgsError("my_action", 1, len(args))
    }
    
    // Execute operation
    result, err := performOperation(args[0])
    if err != nil {
        return types.ActionResult{
            Status: constants.ActionStatusError,
            ErrorInfo: &types.ErrorInfo{
                Category: types.ErrorCategorySystem,
                Code:     "OPERATION_FAILED",
                Message:  "Operation failed to complete",
                Context:  map[string]any{"error": err.Error()},
            },
        }
    }
    
    return types.ActionResult{
        Status: constants.ActionStatusPassed,
        Data:   result,
    }
}
```

### In Test Execution
```go
func (r *TestRunner) executeStep(step types.Step) *types.StepResult {
    start := time.Now()
    actionResult := r.strategyRouter.Execute(step, stepNumber, nil)
    
    stepResult := &types.StepResult{
        Name:     step.Name,
        Status:   actionResult.Status,
        Duration: time.Since(start),
        Data:     actionResult.Data,
    }
    
    if actionResult.ErrorInfo != nil {
        stepResult.Message = actionResult.ErrorInfo.Message
        stepResult.Category = string(actionResult.ErrorInfo.Category)
    }
    
    return stepResult
}
```

## Integration Points

### With YAML Parsing
- All types include appropriate YAML tags for external API
- Optional fields use `omitempty` to keep YAML clean
- Complex types support nested structure parsing

### With Common Package
- Variables system uses generic `any` type for flexibility
- Security system operates on type fields via reflection
- Environment integration works with string field values

### With Constants Package
- `ActionStatus` type defined here, values in constants
- Error and failure categories as typed constants
- Integration maintains type safety across packages

### With Actions Package
- All actions return `ActionResult` with consistent structure
- Error builders provide fluent API for rich error creation
- Utility functions reduce boilerplate in action implementations

## Contributing

### Adding New Types
1. **Single Responsibility**: Each type should have one clear purpose  
2. **YAML Integration**: Include appropriate tags for external API
3. **Documentation**: Include field comments explaining purpose and usage
4. **Builder Patterns**: Consider builder pattern for complex types

### Extending Existing Types
1. **Backwards Compatibility**: New fields should be optional with `omitempty`
2. **Default Values**: Use appropriate zero values or add constructors
3. **Validation**: Add validation methods for complex types
4. **Migration**: Consider migration path for existing YAML files

### Error Type Guidelines
1. **Rich Context**: Include all relevant information for debugging
2. **Actionable Suggestions**: Provide concrete steps for resolution  
3. **Consistent Categories**: Use existing categories where possible
4. **Template Messages**: Use templates for consistent error formatting

This types package provides the foundation for type-safe, well-structured data throughout the Robogo framework, supporting both internal consistency and external YAML API clarity.