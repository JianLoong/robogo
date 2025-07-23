# Execution System

This directory contains the execution strategy system that handles different types of step execution patterns in Robogo tests.

## Architecture Overview

The execution system uses a **Strategy Pattern** with **Priority-based Routing** to handle different step types:

1. **ExecutionStrategyRouter** - Routes steps to appropriate strategies
2. **Execution Strategies** - Handle specific execution patterns
3. **Action Registry** - Provides access to all available actions
4. **Variable System** - Manages test variables and substitution

## Strategy Priority System

Strategies are evaluated in **descending priority order** (highest first):

| Priority | Strategy | Handles | Description |
|----------|----------|---------|-------------|
| 4 | ConditionalExecutionStrategy | `step.If != ""` | Conditional step execution |
| 3 | RetryExecutionStrategy | `step.Retry != nil` | Retry logic with backoff |
| 2 | NestedStepsExecutionStrategy | `len(step.Steps) > 0` | Nested step collections |
| 1 | BasicExecutionStrategy | Simple actions | Default fallback strategy |

## Strategy Implementations

### ConditionalExecutionStrategy
**File**: `control_flow_strategies.go`

**Purpose**: Handles conditional step execution based on variable expressions

**Logic**:
1. Evaluate condition using BasicConditionEvaluator
2. If `true`: Remove `if` property and route back to router
3. If `false`: Return SKIPPED result

**Example**:
```yaml
- name: "Conditional step"
  if: "${user_type} == 'admin'"
  action: log
  args: ["Admin user detected"]
```

### RetryExecutionStrategy  
**File**: `retry_strategy.go`

**Purpose**: Implements retry logic with configurable attempts, delays, and backoff strategies

**Features**:
- **Retry Attempts**: Configurable number of retry attempts
- **Delay Strategies**: Fixed, exponential, linear backoff
- **Selective Retry**: `retry_on` filters for specific error types
- **Status Variables**: Sets `error_occurred`, `error_message`, `step_status`

**⚠️ Limitation**: Does not handle `step.Result` variable storage

**Example**:
```yaml
- name: "HTTP with retry"
  action: http
  args: ["GET", "https://api.example.com/data"]  
  retry:
    attempts: 3
    delay: "2s"
    backoff: "exponential"
    retry_on: ["http_error", "timeout"]
```

### NestedStepsExecutionStrategy
**File**: `nested_steps_strategy.go`

**Purpose**: Executes collections of sub-steps and aggregates results

**Features**:
- **Sub-step Execution**: Each step executed via router recursively
- **Continue on Error**: `continue: true` allows continuing after failures
- **Result Aggregation**: Combines all sub-step results
- **Early Termination**: Stops on first error unless `continue` is set

**⚠️ Limitation**: Does not handle aggregate `step.Result` variable storage

**Example**:
```yaml
- name: "User registration flow"
  steps:
    - name: "Create user"
      action: http
      args: ["POST", "/users"]
      continue: true
    - name: "Send welcome email"  
      action: http
      args: ["POST", "/emails/welcome"]
```

### BasicExecutionStrategy
**File**: `basic_strategy.go`

**Purpose**: Handles simple action execution with full result processing

**Features**:
- **Action Execution**: Direct action registry calls
- **Variable Substitution**: Full `${variable}` and `${ENV:VAR}` support
- **Security Handling**: `no_log` and `sensitive_fields` processing
- **Data Extraction**: `jq`, `xpath`, and `regex` extraction support
- **Result Storage**: ✅ Properly handles `step.Result` variable storage

**Process Flow**:
1. Get action from registry
2. Apply variable substitution to arguments
3. Check security settings (`no_log`, `sensitive_fields`)
4. Execute action function
5. Apply data extraction if configured
6. Store result in variable if `step.Result` specified

## Supporting Components

### ExecutionStrategyRouter
**File**: `strategy_router.go`

**Purpose**: Routes steps to appropriate execution strategies

**Logic**:
1. Iterate through strategies in priority order
2. Call `CanHandle(step)` on each strategy
3. Execute with first strategy that returns `true`
4. Return error if no strategy can handle the step

### BasicConditionEvaluator
**File**: `condition_evaluator.go`

**Purpose**: Evaluates conditional expressions for `if` statements

**Supported Operators**:
- **Comparison**: `==`, `!=`, `>`, `<`, `>=`, `<=`
- **Boolean**: `&&`, `||`, `!`
- **Containment**: `contains`, `starts_with`, `ends_with`
- **Existence**: `exists`, `empty`

### Step Processing Modules

**File**: `step_extraction.go`
- Data extraction functions (`jq`, `xpath`, `regex`)
- Handles complex data access from action results

**File**: `step_masking.go` 
- Sensitive data masking functions
- JSON-aware masking with custom field support
- Automatic detection of passwords, tokens, API keys

**File**: `step_output.go`
- Printing and output functions
- Security-aware logging with masking
- Result formatting and display

## Architectural Decisions

### Priority-Based Strategy Selection
- **Why**: Ensures most specific strategies are tried first
- **Example**: A step with both `if` and `retry` will be handled by ConditionalExecutionStrategy first
- **Benefit**: Clean separation of concerns, predictable behavior

### Strategy Routing Pattern
- **Why**: Allows strategies to delegate back to router for complex scenarios
- **Example**: ConditionalExecutionStrategy removes `if` and routes back
- **Benefit**: Enables composition of multiple execution patterns

### No Strategy Inheritance
- **Why**: KISS principle - avoid complex inheritance hierarchies
- **Implementation**: Each strategy is independent and self-contained
- **Benefit**: Easy to understand, modify, and extend

## Known Limitations

### Result Storage Inconsistency
- ✅ **BasicExecutionStrategy**: Properly handles `step.Result`
- ⚠️ **RetryExecutionStrategy**: Calls actions directly, doesn't store results
- ⚠️ **NestedStepsExecutionStrategy**: Doesn't handle aggregate result storage
- ⚠️ **ConditionalExecutionStrategy**: Inherits behavior from delegated strategy

### Solution Approaches
1. **Move result storage to router level** (after strategy execution)
2. **Standardize result handling** across all strategies
3. **Create base strategy interface** with common result handling

## Error Handling

### Strategy-Level Errors
- **NO_STRATEGY_FOUND**: No strategy can handle the step
- **CONDITION_EVALUATION_FAILED**: Invalid condition syntax
- **EXTRACTION_FAILED**: Data extraction error

### Propagated Errors
- Action-level errors bubble up through strategies
- Security masking applied at strategy level
- Context information preserved through error chain

## Testing Strategies

Each strategy should be tested with:

1. **Happy Path**: Normal successful execution
2. **Error Cases**: Various failure scenarios  
3. **Edge Cases**: Boundary conditions and invalid inputs
4. **Integration**: With real actions and variable systems
5. **Security**: Sensitive data masking verification

## File Structure

```
execution/
├── strategy_router.go           # Strategy routing and coordination
├── execution_strategy.go        # Strategy interface definition
├── basic_strategy.go           # Basic action execution (141 lines)
├── control_flow_strategies.go  # Conditional execution logic
├── retry_strategy.go           # Retry logic with backoff
├── nested_steps_strategy.go    # Nested step collections
├── condition_evaluator.go      # Condition evaluation logic
├── step_extraction.go          # Data extraction functions (92 lines)
├── step_masking.go            # Sensitive data masking (291 lines)
└── step_output.go             # Output and printing (113 lines)
```

## Performance Considerations

- **Strategy Selection**: O(n) where n is number of strategies (typically 4)
- **Recursive Execution**: Nested and conditional strategies can recurse
- **Memory Usage**: Each strategy maintains minimal state
- **Connection Handling**: No persistent connections, clean exit guaranteed

## Future Enhancements

1. **Result Storage Standardization**: Fix inconsistent result handling
2. **Strategy Composition**: Allow multiple strategies per step
3. **Custom Strategies**: Plugin system for user-defined strategies
4. **Performance Optimization**: Strategy caching and pre-selection
5. **Enhanced Conditions**: More complex conditional expressions