# FailureBuilder Implementation Plan

## Phase 1: Core Infrastructure (1-2 days)

### Task 1.1: Create FailureInfo and FailureCategory
```go
// Add to internal/types/error_handling.go

type FailureCategory string

const (
    FailureCategoryAssertion   FailureCategory = "assertion"
    FailureCategoryValidation  FailureCategory = "validation"
    FailureCategoryBusiness    FailureCategory = "business_rule"
    FailureCategoryData        FailureCategory = "data_mismatch"
    FailureCategoryResponse    FailureCategory = "response_validation"
)

type FailureInfo struct {
    Category    FailureCategory `json:"category"`
    Code        string          `json:"code"`
    Message     string          `json:"message"`
    Context     map[string]any  `json:"context,omitempty"`
    Suggestions []string        `json:"suggestions,omitempty"`
    Timestamp   time.Time       `json:"timestamp"`
    Expected    any             `json:"expected,omitempty"`
    Actual      any             `json:"actual,omitempty"`
    Comparison  string          `json:"comparison,omitempty"`
}
```

### Task 1.2: Implement FailureBuilder
```go
type FailureBuilder struct {
    category    FailureCategory
    code        string
    template    string
    context     map[string]any
    suggestions []string
    expected    any
    actual      any
    comparison  string
    formatter   *SafeFormatter
}

func NewFailureBuilder(category FailureCategory, code string) *FailureBuilder {
    return &FailureBuilder{
        category:  category,
        code:      code,
        context:   make(map[string]any),
        formatter: GetDefaultSafeFormatter(),
    }
}

// Implement all fluent interface methods...
```

### Task 1.3: Update ActionResult
```go
// Modify internal/types/action_result.go

type ActionResult struct {
    Status      ActionStatus `json:"status"`
    ErrorInfo   *ErrorInfo   `json:"error_info,omitempty"`
    FailureInfo *FailureInfo `json:"failure_info,omitempty"` // NEW
    Data        any          `json:"data,omitempty"`
    Meta        any          `json:"meta,omitempty"`
}

// Add helper functions
func NewFailedResult(msg string) ActionResult
func NewFailedResultWithInfo(info *FailureInfo) ActionResult

// Update existing methods
func (ar *ActionResult) GetMessage() string {
    if ar.ErrorInfo != nil {
        return ar.ErrorInfo.Message
    }
    if ar.FailureInfo != nil {
        return ar.FailureInfo.Message
    }
    return ""
}

func (ar *ActionResult) IsFailed() bool {
    return ar.Status == ActionStatusFailed
}

func (ar *ActionResult) HasIssue() bool {
    return ar.IsError() || ar.IsFailed()
}
```

## Phase 2: Template System Enhancement (1 day)

### Task 2.1: Add Failure Templates
```go
// Add to internal/templates/error_templates.go

var FailureTemplates = map[string]string{
    // Assertion failures
    "assertion.boolean_failed":     "Assertion failed: expected true but got %v (%s)",
    "assertion.comparison_failed":  "Assertion failed: expected %v %s %v but got %v",
    "assertion.type_mismatch":      "Assertion failed: type mismatch - expected %s but got %s",
    
    // Response validation failures
    "response.status_mismatch":     "HTTP response status mismatch: expected %d but got %d",
    "response.body_mismatch":       "HTTP response body mismatch: expected %v but got %v",
    "response.header_missing":      "HTTP response missing expected header: %s",
    
    // Data validation failures
    "data.count_mismatch":          "Data count mismatch: expected %d items but got %d",
    "data.value_mismatch":          "Data value mismatch: expected %v but got %v",
    "data.format_invalid":          "Data format invalid: expected %s format but got %v",
    
    // Business rule failures
    "business.rule_violated":       "Business rule violated: %s",
    "business.constraint_failed":   "Business constraint failed: %s",
}
```

### Task 2.2: Update SafeFormatter
```go
// Extend SafeFormatter to handle failure templates
func initializeDefaultTemplates(formatter *SafeFormatter) {
    // Existing error templates
    errorTemplates := templates.InitializeErrorTemplates()
    for name, template := range errorTemplates {
        formatter.RegisterTemplate(name, template)
    }
    
    // NEW: Failure templates
    failureTemplates := templates.InitializeFailureTemplates()
    for name, template := range failureTemplates {
        formatter.RegisterTemplate(name, template)
    }
}
```

## Phase 3: Action Updates (2-3 days)

### Task 3.1: Update Assert Action
```go
// Modify internal/actions/assert.go

func assertAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
    // ... validation logic stays the same but uses ErrorBuilder ...
    
    // For assertion failures, use FailureBuilder instead of ErrorBuilder
    if len(args) == 1 {
        if b, ok := args[0].(bool); ok && b {
            return types.ActionResult{Status: constants.ActionStatusPassed}
        }
        
        // CHANGED: Use FailureBuilder for assertion failure
        return types.NewFailureBuilder(types.FailureCategoryAssertion, "ASSERT_BOOLEAN_FAILED").
            WithTemplate("assertion failed: expected true but got %v (%s)").
            WithExpected(true).
            WithActual(args[0]).
            WithComparison("boolean equality").
            WithSuggestion("Check if the condition should evaluate to true").
            Build(args[0], fmt.Sprintf("%T", args[0]))
    }
    
    // For comparison assertions
    if len(args) >= 3 {
        actual := args[0]
        operator := args[1]
        expected := args[2]
        
        // ... comparison logic ...
        
        if !result {
            return types.NewFailureBuilder(types.FailureCategoryAssertion, "ASSERT_COMPARISON_FAILED").
                WithTemplate("assertion failed: expected %v %s %v but got %v").
                WithExpected(expected).
                WithActual(actual).
                WithComparison(fmt.Sprintf("%v", operator)).
                WithSuggestion("Verify the expected value and comparison operator").
                Build(expected, operator, expected, actual)
        }
    }
}
```

### Task 3.2: Update HTTP Action for Response Validation
```go
// Add to internal/actions/http.go

// After successful HTTP request, add optional response validation
func validateHTTPResponse(resp *http.Response, options map[string]any) types.ActionResult {
    // Check expected status code
    if expectedStatus, ok := options["expected_status"]; ok {
        if expected, ok := expectedStatus.(int); ok {
            if resp.StatusCode != expected {
                return types.NewFailureBuilder(types.FailureCategoryResponse, "HTTP_STATUS_MISMATCH").
                    WithTemplate("HTTP response status mismatch: expected %d but got %d").
                    WithExpected(expected).
                    WithActual(resp.StatusCode).
                    WithContext("url", resp.Request.URL.String()).
                    WithContext("method", resp.Request.Method).
                    WithSuggestion("Verify the expected status code for this endpoint").
                    Build(expected, resp.StatusCode)
            }
        }
    }
    
    // Check expected headers
    if expectedHeaders, ok := options["expected_headers"]; ok {
        if headers, ok := expectedHeaders.(map[string]any); ok {
            for key, expectedValue := range headers {
                actualValue := resp.Header.Get(key)
                if actualValue != fmt.Sprintf("%v", expectedValue) {
                    return types.NewFailureBuilder(types.FailureCategoryResponse, "HTTP_HEADER_MISMATCH").
                        WithTemplate("HTTP header mismatch: expected %s=%v but got %v").
                        WithExpected(expectedValue).
                        WithActual(actualValue).
                        WithContext("header_name", key).
                        WithSuggestion("Check if the server returns the expected header").
                        Build(key, expectedValue, actualValue)
                }
            }
        }
    }
    
    return types.ActionResult{Status: constants.ActionStatusPassed}
}
```

### Task 3.3: Update Database Actions for Data Validation
```go
// Add to internal/actions/postgres.go and spanner.go

func validateQueryResults(results [][]any, options map[string]any) types.ActionResult {
    // Check expected row count
    if expectedCount, ok := options["expected_count"]; ok {
        if expected, ok := expectedCount.(int); ok {
            if len(results) != expected {
                return types.NewFailureBuilder(types.FailureCategoryData, "DB_ROW_COUNT_MISMATCH").
                    WithTemplate("Query returned %d rows, expected %d").
                    WithExpected(expected).
                    WithActual(len(results)).
                    WithSuggestion("Check if the query conditions are correct").
                    Build(len(results), expected)
            }
        }
    }
    
    // Check expected values in specific columns
    if expectedValues, ok := options["expected_values"]; ok {
        // Implementation for value validation...
    }
    
    return types.ActionResult{Status: constants.ActionStatusPassed}
}
```

## Phase 4: Enhanced Reporting (1-2 days)

### Task 4.1: Update CLI Output
```go
// Modify internal/control_flow.go

func printStepResult(step types.Step, result types.ActionResult, duration time.Duration) {
    switch result.Status {
    case constants.ActionStatusPassed:
        fmt.Printf("✓ PASSED (%s)\n", duration)
    case constants.ActionStatusFailed:  // NEW
        fmt.Printf("✗ FAILED (%s)\n", duration)
        if msg := result.GetMessage(); msg != "" {
            fmt.Printf("    Failure: %s\n", msg)
        }
        // Show expected vs actual for failures
        if result.FailureInfo != nil {
            if result.FailureInfo.Expected != nil && result.FailureInfo.Actual != nil {
                fmt.Printf("    Expected: %v\n", result.FailureInfo.Expected)
                fmt.Printf("    Actual:   %v\n", result.FailureInfo.Actual)
            }
        }
    case constants.ActionStatusError:
        fmt.Printf("! ERROR (%s)\n", duration)
        if msg := result.GetMessage(); msg != "" {
            fmt.Printf("    Error: %s\n", msg)
        }
    case constants.ActionStatusSkipped:
        fmt.Printf("- SKIPPED (%s)\n", duration)
        if reason := result.GetSkipReason(); reason != "" {
            fmt.Printf("    Reason: %s\n", reason)
        }
    }
}
```

### Task 4.2: Update Test Result Aggregation
```go
// Modify internal/runner.go

func determineOverallStatus(stepResults []types.StepResult) string {
    hasErrors := false
    hasFailures := false
    
    for _, sr := range stepResults {
        switch sr.Result.Status {
        case types.ActionStatusError:
            hasErrors = true
        case types.ActionStatusFailed:
            hasFailures = true
        }
    }
    
    // Priority: Error > Failed > Passed
    if hasErrors {
        return string(types.ActionStatusError)
    }
    if hasFailures {
        return string(types.ActionStatusFailed)
    }
    return string(types.ActionStatusPassed)
}
```

## Phase 5: Testing and Validation (1-2 days)

### Task 5.1: Unit Tests
- Test `FailureBuilder` with all categories
- Test `ActionResult` helper methods
- Test failure template formatting
- Test backward compatibility

### Task 5.2: Integration Tests
- Test assertion failures return FAILED status
- Test HTTP response validation
- Test database data validation
- Test mixed error/failure scenarios

### Task 5.3: Migration Validation
- Ensure existing tests still work
- Validate CLI output improvements
- Test that tools can distinguish errors from failures

## Benefits After Implementation

1. **Clear Status Distinction**: 
   - ERROR = "System couldn't execute the test"
   - FAILED = "Test executed but expectations weren't met"

2. **Better Debugging**:
   - Failures show expected vs actual values
   - Rich context for understanding what went wrong

3. **Improved Tooling**:
   - CI/CD systems can handle errors vs failures differently
   - Test reports can categorize issues appropriately

4. **Enhanced User Experience**:
   - More actionable error messages
   - Visual distinction between errors and failures

## Estimated Timeline: 5-8 days total

This plan provides a comprehensive approach to implementing the FailureBuilder while maintaining backward compatibility and enhancing the overall error handling experience.