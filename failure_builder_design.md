# FailureBuilder Design Plan

## Overview

The `FailureBuilder` extends the existing error handling system to properly distinguish between technical errors and logical test failures. This addresses the current limitation where all failures return ERROR status instead of the appropriate FAIL status.

## Key Distinctions

### ErrorBuilder (Technical Errors)
- **Purpose**: System/infrastructure problems that prevent execution
- **Status**: `ActionStatusError`
- **Examples**: Network timeouts, missing arguments, database connection failures
- **User Action**: Fix configuration, check system resources, verify setup

### FailureBuilder (Logical Failures)  
- **Purpose**: Test expectations not met, business logic violations
- **Status**: `ActionStatusFailed`
- **Examples**: Assertion failures, unexpected response values, data mismatches
- **User Action**: Review test logic, check expected values, analyze business requirements

## FailureBuilder Interface

```go
// FailureInfo contains structured information about a logical failure
type FailureInfo struct {
    Category    FailureCategory `json:"category"`
    Code        string          `json:"code"`
    Message     string          `json:"message"`
    Context     map[string]any  `json:"context,omitempty"`
    Suggestions []string        `json:"suggestions,omitempty"`
    Timestamp   time.Time       `json:"timestamp"`
    
    // Failure-specific fields
    Expected    any             `json:"expected,omitempty"`
    Actual      any             `json:"actual,omitempty"`
    Comparison  string          `json:"comparison,omitempty"`
}

// FailureCategory represents different categories of logical failures
type FailureCategory string

const (
    FailureCategoryAssertion   FailureCategory = "assertion"
    FailureCategoryValidation  FailureCategory = "validation"
    FailureCategoryBusiness    FailureCategory = "business_rule"
    FailureCategoryData        FailureCategory = "data_mismatch"
    FailureCategoryResponse    FailureCategory = "response_validation"
)

// FailureBuilder provides a fluent interface for building structured failures
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

// Constructor
func NewFailureBuilder(category FailureCategory, code string) *FailureBuilder

// Fluent interface methods
func (fb *FailureBuilder) WithTemplate(template string) *FailureBuilder
func (fb *FailureBuilder) WithContext(key string, value any) *FailureBuilder
func (fb *FailureBuilder) WithSuggestion(suggestion string) *FailureBuilder
func (fb *FailureBuilder) WithExpected(expected any) *FailureBuilder
func (fb *FailureBuilder) WithActual(actual any) *FailureBuilder
func (fb *FailureBuilder) WithComparison(comparison string) *FailureBuilder
func (fb *FailureBuilder) Build(args ...any) ActionResult
```

## Enhanced ActionResult

```go
type ActionResult struct {
    Status      ActionStatus `json:"status"`
    ErrorInfo   *ErrorInfo   `json:"error_info,omitempty"`   // For technical errors
    FailureInfo *FailureInfo `json:"failure_info,omitempty"` // For logical failures
    Data        any          `json:"data,omitempty"`
    Meta        any          `json:"meta,omitempty"`
}

// Helper methods
func (ar *ActionResult) GetMessage() string
func (ar *ActionResult) IsError() bool    // Technical error
func (ar *ActionResult) IsFailed() bool   // Logical failure
func (ar *ActionResult) IsSuccess() bool
func (ar *ActionResult) HasIssue() bool   // Either error or failure
```

## Usage Examples

### Assertion Failures
```go
// Current (incorrect)
return types.NewErrorBuilder(types.ErrorCategoryAssertion, "ASSERT_FAILED").
    WithTemplate("assertion failed: %v != %v").
    Build(actual, expected)

// New (correct)
return types.NewFailureBuilder(types.FailureCategoryAssertion, "ASSERT_COMPARISON_FAILED").
    WithTemplate("assertion failed: expected %v but got %v").
    WithExpected(expected).
    WithActual(actual).
    WithComparison("equality").
    WithSuggestion("Check if the expected value is correct").
    Build()
```

### HTTP Response Validation
```go
// When HTTP request succeeds but returns unexpected status
if resp.StatusCode != expectedStatus {
    return types.NewFailureBuilder(types.FailureCategoryResponse, "HTTP_UNEXPECTED_STATUS").
        WithTemplate("HTTP request returned status %d, expected %d").
        WithExpected(expectedStatus).
        WithActual(resp.StatusCode).
        WithContext("url", url).
        WithContext("method", method).
        WithSuggestion("Verify the expected status code for this endpoint").
        Build(resp.StatusCode, expectedStatus)
}
```

### Database Data Validation
```go
// When query succeeds but returns unexpected data
if len(rows) != expectedCount {
    return types.NewFailureBuilder(types.FailureCategoryData, "DB_ROW_COUNT_MISMATCH").
        WithTemplate("Query returned %d rows, expected %d").
        WithExpected(expectedCount).
        WithActual(len(rows)).
        WithContext("query", query).
        WithSuggestion("Check if the query conditions are correct").
        Build(len(rows), expectedCount)
}
```

## Implementation Strategy

### Phase 1: Core Infrastructure
1. Create `FailureInfo` struct and `FailureCategory` enum
2. Implement `FailureBuilder` with fluent interface
3. Add `FailureInfo` field to `ActionResult`
4. Create helper functions: `NewFailedResult()`, etc.

### Phase 2: Template System
1. Add failure-specific templates to the template system
2. Extend `SafeFormatter` to handle failure templates
3. Create predefined failure templates for common scenarios

### Phase 3: Action Updates
1. Update assertion action to use `FailureBuilder` for assertion failures
2. Update HTTP action to use `FailureBuilder` for response validation failures
3. Update database actions to use `FailureBuilder` for data validation failures

### Phase 4: Enhanced Reporting
1. Update CLI output to distinguish between errors and failures
2. Add failure-specific formatting and colors
3. Update test result aggregation to handle both errors and failures

## Migration Considerations

### Backward Compatibility
- Existing `IsError()` method should return true for both errors and failures
- Add new methods: `IsError()` (technical only), `IsFailed()` (logical only), `HasIssue()` (either)
- Maintain existing error message retrieval methods

### Template Migration
- Assertion templates move from error templates to failure templates
- Response validation templates become failure templates
- Keep error templates for technical issues

### Status Code Handling
- Update runner to handle both ERROR and FAILED statuses
- Update control flow to distinguish between recoverable failures and blocking errors
- Update test result aggregation logic

## Benefits

1. **Clearer Semantics**: Distinguish between "can't run" vs "ran but failed"
2. **Better Debugging**: Failure context includes expected vs actual values
3. **Improved Tooling**: Tools can handle errors and failures differently
4. **Enhanced Reporting**: Different visual treatment for errors vs failures
5. **Better User Experience**: More actionable error messages

## Testing Strategy

### Unit Tests
- Test `FailureBuilder` with all categories and templates
- Test `ActionResult` helper methods for failures
- Test failure template formatting and validation

### Integration Tests
- Test assertion failures return FAILED status
- Test HTTP response validation failures
- Test database data validation failures
- Test mixed error/failure scenarios

### Migration Tests
- Ensure backward compatibility of existing error handling
- Test that existing tests still work with new failure handling
- Validate that CLI output is improved but not broken