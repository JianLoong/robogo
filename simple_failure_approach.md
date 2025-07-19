# Simple Failure Approach - Minimal Boilerplate

## Problem with Current Design
- Duplicate builders (`ErrorBuilder` + `FailureBuilder`)
- Duplicate info structs (`ErrorInfo` + `FailureInfo`) 
- Duplicate helper methods and templates
- Complex ActionResult with multiple optional fields

## Simple Solution: Status-Aware ErrorBuilder

### 1. Just Add Status Parameter
```go
// Current
func (eb *ErrorBuilder) Build(args ...any) ActionResult {
    return ActionResult{
        Status:    ActionStatusError,  // Always ERROR
        ErrorInfo: errorInfo,
    }
}

// New - just add status parameter
func (eb *ErrorBuilder) BuildWithStatus(status ActionStatus, args ...any) ActionResult {
    return ActionResult{
        Status:    status,  // Can be ERROR or FAILED
        ErrorInfo: errorInfo,
    }
}

// Convenience methods
func (eb *ErrorBuilder) BuildError(args ...any) ActionResult {
    return eb.BuildWithStatus(ActionStatusError, args...)
}

func (eb *ErrorBuilder) BuildFailure(args ...any) ActionResult {
    return eb.BuildWithStatus(ActionStatusFailed, args...)
}
```

### 2. Add Expected/Actual to Existing ErrorInfo
```go
// Just extend existing ErrorInfo - no new struct needed
type ErrorInfo struct {
    Category    ErrorCategory  `json:"category"`
    Code        string         `json:"code"`
    Message     string         `json:"message"`
    Context     map[string]any `json:"context,omitempty"`
    Suggestions []string       `json:"suggestions,omitempty"`
    Timestamp   time.Time      `json:"timestamp"`
    
    // NEW: Optional fields for failures
    Expected   any    `json:"expected,omitempty"`
    Actual     any    `json:"actual,omitempty"`
    Comparison string `json:"comparison,omitempty"`
}
```

### 3. Add Convenience Methods to ErrorBuilder
```go
// Just add these methods to existing ErrorBuilder
func (eb *ErrorBuilder) WithExpected(expected any) *ErrorBuilder {
    eb.context["expected"] = expected
    return eb
}

func (eb *ErrorBuilder) WithActual(actual any) *ErrorBuilder {
    eb.context["actual"] = actual
    return eb
}

func (eb *ErrorBuilder) WithComparison(comparison string) *ErrorBuilder {
    eb.context["comparison"] = comparison
    return eb
}
```

## Usage Examples

### Technical Errors (unchanged)
```go
return types.NewErrorBuilder(types.ErrorCategoryValidation, "LOG_MISSING_ARGS").
    WithTemplate("log action requires at least 1 argument").
    BuildError()  // Still returns ERROR status
```

### Assertion Failures (minimal change)
```go
return types.NewErrorBuilder(types.ErrorCategoryAssertion, "ASSERT_FAILED").
    WithTemplate("assertion failed: expected %v but got %v").
    WithExpected(expected).
    WithActual(actual).
    WithSuggestion("Check the expected value").
    BuildFailure(expected, actual)  // Returns FAILED status
```

## Benefits of Simple Approach

✅ **Minimal Code Changes**: Just add a few methods to existing ErrorBuilder
✅ **No Duplicate Systems**: Reuse all existing infrastructure  
✅ **Backward Compatible**: All existing code works unchanged
✅ **Same Flexibility**: Can still distinguish errors from failures
✅ **Less Boilerplate**: No new builders, structs, or parallel systems

## Implementation (30 minutes instead of 5-8 days!)

### Step 1: Add methods to ErrorBuilder
```go
func (eb *ErrorBuilder) BuildWithStatus(status ActionStatus, args ...any) ActionResult
func (eb *ErrorBuilder) BuildError(args ...any) ActionResult  
func (eb *ErrorBuilder) BuildFailure(args ...any) ActionResult
func (eb *ErrorBuilder) WithExpected(expected any) *ErrorBuilder
func (eb *ErrorBuilder) WithActual(actual any) *ErrorBuilder
func (eb *ErrorBuilder) WithComparison(comparison string) *ErrorBuilder
```

### Step 2: Update assert action
```go
// Change this line:
return builder.Build(args...)

// To this:
return builder.BuildFailure(args...)
```

### Step 3: Done! 🎉

## Even Simpler Alternative

If we want to be even more minimal, we could just:

1. **Add a single method**: `NewFailureBuilder()` that's identical to `NewErrorBuilder()` but sets a flag
2. **Override Build()**: Check the flag and return appropriate status
3. **That's it**: 5 lines of code total

```go
func NewFailureBuilder(category ErrorCategory, code string) *ErrorBuilder {
    eb := NewErrorBuilder(category, code)
    eb.isFailure = true  // Add this field
    return eb
}

func (eb *ErrorBuilder) Build(args ...any) ActionResult {
    status := ActionStatusError
    if eb.isFailure {
        status = ActionStatusFailed
    }
    
    return ActionResult{
        Status:    status,
        ErrorInfo: errorInfo,
    }
}
```

## Recommendation

I'd go with the **Status-Aware ErrorBuilder** approach:
- Minimal boilerplate
- Maximum reuse of existing code
- Easy to implement and maintain
- Achieves the same end goal

What do you think? Much cleaner, right?