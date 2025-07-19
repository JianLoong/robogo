# Simple Error Functions - No More Boilerplate

## Problem
Functions are super long with fluent interface chains:
```go
return types.NewErrorBuilder(types.ErrorCategoryValidation, "KAFKA_MISSING_ARGS").
    WithTemplate("kafka action requires at least 2 arguments: operation, broker").
    WithContext("broker", broker).
    WithContext("topic", topic).
    WithSuggestion("Check arguments").
    Build()
```

## Solution: Simple Helper Functions

### Create Simple Error Functions
```go
// internal/types/simple_errors.go

// Validation errors
func MissingArgsError(action string, required, provided int) ActionResult {
    return NewErrorBuilder(ErrorCategoryValidation, "MISSING_ARGS").
        WithTemplate("%s action requires %d arguments, got %d").
        Build(action, required, provided)
}

func InvalidArgError(action, arg string, expected any) ActionResult {
    return NewErrorBuilder(ErrorCategoryValidation, "INVALID_ARG").
        WithTemplate("%s action: invalid %s, expected %v").
        Build(action, arg, expected)
}

// Network errors  
func ConnectionError(service, details string) ActionResult {
    return NewErrorBuilder(ErrorCategoryNetwork, "CONNECTION_FAILED").
        WithTemplate("Failed to connect to %s: %s").
        Build(service, details)
}

func RequestError(method, url, details string) ActionResult {
    return NewErrorBuilder(ErrorCategoryNetwork, "REQUEST_FAILED").
        WithTemplate("%s %s failed: %s").
        Build(method, url, details)
}

// Database errors
func DatabaseError(operation, details string) ActionResult {
    return NewErrorBuilder(ErrorCategoryDatabase, "DB_ERROR").
        WithTemplate("Database %s failed: %s").
        Build(operation, details)
}

// Assertion failures (using simple approach)
func AssertionFailure(expected, actual any, operator string) ActionResult {
    return NewErrorBuilder(ErrorCategoryAssertion, "ASSERTION_FAILED").
        WithTemplate("Expected %v %s %v, got %v").
        BuildFailure(expected, operator, expected, actual)  // Returns FAILED status
}
```

## Usage - Much Cleaner!

### Before (verbose)
```go
if len(args) < 2 {
    return types.NewErrorBuilder(types.ErrorCategoryValidation, "KAFKA_MISSING_ARGS").
        WithTemplate("kafka action requires at least 2 arguments: operation, broker").
        Build()
}
```

### After (clean)
```go
if len(args) < 2 {
    return types.MissingArgsError("kafka", 2, len(args))
}
```

### Before (verbose)
```go
if err != nil {
    return types.NewErrorBuilder(types.ErrorCategoryNetwork, "KAFKA_PUBLISH_FAILED").
        WithTemplate("failed to publish message: %v").
        WithContext("broker", broker).
        WithContext("topic", topic).
        Build(err)
}
```

### After (clean)
```go
if err != nil {
    return types.RequestError("kafka publish", fmt.Sprintf("%s/%s", broker, topic), err.Error())
}
```

### Before (verbose assertion)
```go
if !result {
    return types.NewErrorBuilder(types.ErrorCategoryAssertion, "ASSERT_COMPARISON_FAILED").
        WithTemplate("assertion failed: expected %v %s %v but got %v").
        WithExpected(expected).
        WithActual(actual).
        Build(expected, operator, expected, actual)
}
```

### After (clean assertion)
```go
if !result {
    return types.AssertionFailure(expected, actual, fmt.Sprintf("%v", operator))
}
```

## Even Simpler - Action-Specific Helpers

```go
// internal/actions/errors.go

func httpMissingArgs() ActionResult {
    return types.MissingArgsError("http", 2, 0)
}

func httpRequestFailed(method, url string, err error) ActionResult {
    return types.RequestError(method, url, err.Error())
}

func kafkaMissingArgs(provided int) ActionResult {
    return types.MissingArgsError("kafka", 2, provided)
}

func kafkaPublishFailed(broker, topic string, err error) ActionResult {
    return types.RequestError("kafka publish", fmt.Sprintf("%s/%s", broker, topic), err.Error())
}
```

## Result: Super Clean Action Code

```go
func kafkaAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
    if len(args) < 2 {
        return kafkaMissingArgs(len(args))  // One line!
    }

    // ... logic ...

    err := w.WriteMessages(ctx, kafka.Message{Value: []byte(message)})
    if err != nil {
        return kafkaPublishFailed(broker, topic, err)  // One line!
    }

    return types.ActionResult{Status: constants.ActionStatusPassed, Data: map[string]any{"status": "published"}}
}
```

## Benefits

✅ **Functions are short again**
✅ **No repetitive boilerplate** 
✅ **Consistent error messages**
✅ **Easy to maintain**
✅ **Still flexible when needed**

## Implementation Strategy

1. **Create simple error functions** for common patterns
2. **Update actions one by one** to use simple functions
3. **Keep ErrorBuilder** for complex cases
4. **Result**: Clean, readable code

Much better, right? Functions go from 10+ lines back to 1-2 lines for errors!