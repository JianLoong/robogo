package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/types"
	"github.com/expr-lang/expr"
)

// Variables - variable storage and substitution with optional enhanced tracking
type Variables struct {
	data            map[string]any
	trackingEnabled bool
}

func NewVariables() *Variables {
	return &Variables{
		data:            make(map[string]any),
		trackingEnabled: false, // Default to basic mode for performance
	}
}

// NewVariablesWithTracking creates a Variables instance with optional enhanced tracking
func NewVariablesWithTracking(enableTracking bool) *Variables {
	return &Variables{
		data:            make(map[string]any),
		trackingEnabled: enableTracking,
	}
}

// EnableTracking enables enhanced variable resolution tracking
func (v *Variables) EnableTracking() {
	v.trackingEnabled = true
}

// DisableTracking disables enhanced variable resolution tracking for better performance
func (v *Variables) DisableTracking() {
	v.trackingEnabled = false
}

func (v *Variables) Set(key string, value any) {
	// If value is a JSON string, try to parse it
	if str, ok := value.(string); ok {
		var parsed any
		if err := json.Unmarshal([]byte(str), &parsed); err == nil {
			v.data[key] = parsed
			return
		}
	}
	v.data[key] = value
}

func (v *Variables) Get(key string) any {
	if val, exists := v.data[key]; exists {
		return val
	}
	return nil
}

func (v *Variables) Load(vars map[string]any) {
	for k, val := range vars {
		v.data[k] = val
	}
}

// Substitute performs variable substitution using ${variable} syntax
func (v *Variables) Substitute(template string) string {
	if v.trackingEnabled {
		result, _ := v.SubstituteWithContext(template)
		return result
	}

	// Fast path for basic substitution without tracking
	return v.basicSubstitute(template)
}

// SubstituteWithContext performs variable substitution with detailed tracking and enhanced error information
func (v *Variables) SubstituteWithContext(template string) (string, *types.VariableContext) {
	// Create context for tracking
	context := types.NewVariableContext(template, v.data)

	if template == "" {
		context.SetProcessedTemplate(template)
		return template, context
	}

	result := template
	processedExpressions := make(map[string]bool) // Track processed expressions to detect cycles
	maxIterations := 100                          // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			// Malformed template - record detailed syntax error
			remainingTemplate := result[start:]
			attempt := types.VariableAccessAttempt{
				Expression:    remainingTemplate,
				Status:        types.VariableStatusUnresolved,
				FailureReason: types.FailureReasonInvalidSyntax,
				ErrorMessage:  fmt.Sprintf("Unclosed variable expression starting at position %d", start),
				Suggestions:   v.generateSyntaxSuggestions(remainingTemplate),
			}
			context.AddAttempt(attempt)
			break
		}
		end += start

		exprStr := result[start+2 : end]
		if exprStr == "" {
			// Empty expression - skip and continue
			result = result[:start] + result[end+1:]
			continue
		}

		// Check for circular references
		if processedExpressions[exprStr] {
			attempt := types.VariableAccessAttempt{
				Expression:    exprStr,
				Status:        types.VariableStatusCyclic,
				FailureReason: types.FailureReasonCircularRef,
				ErrorMessage:  "Circular reference detected in variable substitution",
				Suggestions:   []string{"Check for variables that reference themselves", "Review variable dependencies"},
			}
			context.AddAttempt(attempt)
			replacement := fmt.Sprintf("__CIRCULAR_REF_%s__", exprStr)
			result = result[:start] + replacement + result[end+1:]
			continue
		}
		processedExpressions[exprStr] = true

		// Analyze the expression first for detailed tracking
		attempt := v.analyzeExpressionWithContext(exprStr, start, end-start+1)

		// Evaluate using expr with enhanced error handling
		output, err := v.evaluateExpressionWithDetailedError(exprStr)
		var replacement string

		if err != nil {
			// Enhanced error analysis
			v.enhanceAttemptWithExprError(&attempt, err, exprStr)

			if v.trackingEnabled {
				fmt.Printf("[WARN] Variable resolution failed for '%s': %s\n", exprStr, attempt.ErrorMessage)
			}
			replacement = fmt.Sprintf("__UNRESOLVED_%s__", exprStr)
		} else if output == nil {
			attempt.Status = types.VariableStatusUnresolved
			attempt.FailureReason = types.FailureReasonNullValue
			attempt.ErrorMessage = "Expression evaluated to null/undefined"
			attempt.Suggestions = []string{
				"Check if the variable is properly initialized",
				"Verify the variable exists in the current scope",
				"Consider using a default value with the || operator",
			}

			if v.trackingEnabled {
				fmt.Printf("[WARN] Variable '%s' evaluated to null\n", exprStr)
			}
			replacement = fmt.Sprintf("__NULL_%s__", exprStr)
		} else {
			// Success - update attempt with resolution details
			attempt.Status = types.VariableStatusResolved
			attempt.ResolvedValue = output

			// Add type information to the attempt
			v.addTypeInformationToAttempt(&attempt, output)

			replacement = fmt.Sprintf("%v", output)
		}

		// Record the attempt with enhanced context
		context.AddAttempt(attempt)

		result = result[:start] + replacement + result[end+1:]
	}

	// Check for infinite loop detection
	if iteration >= maxIterations {
		attempt := types.VariableAccessAttempt{
			Expression:    "infinite_loop_detected",
			Status:        types.VariableStatusUnresolved,
			FailureReason: types.FailureReasonCircularRef,
			ErrorMessage:  fmt.Sprintf("Maximum substitution iterations (%d) exceeded - possible infinite loop", maxIterations),
			Suggestions:   []string{"Check for circular variable references", "Simplify variable dependencies"},
		}
		context.AddAttempt(attempt)
		if v.trackingEnabled {
			fmt.Printf("[ERROR] Variable substitution exceeded maximum iterations (%d) - possible infinite loop\n", maxIterations)
		}
	}

	// Enhanced final validation and warnings
	if v.trackingEnabled {
		v.validateFinalResult(result, context)
	}

	context.SetProcessedTemplate(result)
	return result, context
}

// basicSubstitute provides fast variable substitution without tracking (for performance)
func (v *Variables) basicSubstitute(template string) string {
	if template == "" {
		return template
	}

	result := template
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		exprStr := result[start+2 : end]
		if exprStr == "" {
			result = result[:start] + result[end+1:]
			continue
		}

		// Evaluate using expr
		output, err := expr.Eval(exprStr, v.data)
		var replacement string
		if err != nil || output == nil {
			fmt.Printf("[WARN] Unresolved variable or expr error in template: %q, error: %v\n", exprStr, err)
			replacement = "__UNRESOLVED__"
		} else {
			replacement = fmt.Sprintf("%v", output)
		}
		result = result[:start] + replacement + result[end+1:]
	}
	// Warn if unresolved marker remains
	if strings.Contains(result, "__UNRESOLVED__") {
		fmt.Printf("[WARN] Unresolved variable in template: %q\n", result)
	}
	return result
}

// Substitute variables in arguments
func (v *Variables) SubstituteArgs(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			result[i] = v.Substitute(str)
		} else {
			result[i] = arg
		}
	}
	return result
}

// SubstituteArgsWithContext performs variable substitution on arguments with tracking
func (v *Variables) SubstituteArgsWithContext(args []any) ([]any, []*types.VariableContext) {
	if !v.trackingEnabled {
		// Fall back to original behavior
		result := v.SubstituteArgs(args)
		return result, nil
	}

	result := make([]any, len(args))
	contexts := make([]*types.VariableContext, len(args))

	for i, arg := range args {
		if str, ok := arg.(string); ok {
			substituted, context := v.SubstituteWithContext(str)
			result[i] = substituted
			contexts[i] = context
		} else {
			result[i] = arg
			contexts[i] = nil // No context needed for non-string args
		}
	}

	return result, contexts
}

// ValidateTemplate checks if all variables in a template can be resolved
func (v *Variables) ValidateTemplate(template string) *types.VariableContext {
	_, context := v.SubstituteWithContext(template)
	return context
}

// GetMissingVariables returns a list of variable names that are referenced but not defined
func (v *Variables) GetMissingVariables(template string) []string {
	context := v.ValidateTemplate(template)
	var missing []string

	for _, attempt := range context.GetUnresolvedVariables() {
		if attempt.FailureReason == types.FailureReasonNotFound {
			// Extract root variable name from expression
			parts := strings.Split(attempt.Expression, ".")
			if len(parts) > 0 {
				missing = append(missing, parts[0])
			}
		}
	}

	return removeDuplicates(missing)
}

// CreateVariableErrorResult creates an ActionResult for variable resolution failures
func (v *Variables) CreateVariableErrorResult(context *types.VariableContext) types.ActionResult {
	if context.Status == types.VariableStatusResolved {
		return types.ActionResult{Status: "passed"}
	}

	// Choose appropriate error template based on failure analysis
	analysis := context.GetFailureAnalysis()

	builder := types.NewErrorBuilder(types.ErrorCategoryVariable, "VARIABLE_RESOLUTION_FAILED")

	// Add detailed context information
	contextMap := map[string]any{
		"original_template":  context.OriginalTemplate,
		"processed_template": context.ProcessedTemplate,
		"status":             string(context.Status),
		"unresolved_count":   context.UnresolvedCount,
		"resolved_count":     context.ResolvedCount,
		"processing_time_ms": context.ProcessingTime.Milliseconds(),
		"total_attempts":     analysis.TotalAttempts,
		"failures_by_reason": analysis.FailuresByReason,
	}

	for key, value := range contextMap {
		builder.WithContext(key, value)
	}

	// Add suggestions from analysis
	for _, recommendation := range analysis.Recommendations {
		builder.WithSuggestion(recommendation)
	}

	// Add individual failed attempts
	for i, attempt := range context.GetUnresolvedVariables() {
		attemptKey := fmt.Sprintf("failed_attempt_%d", i+1)
		attemptData := map[string]any{
			"expression":     attempt.Expression,
			"failure_reason": string(attempt.FailureReason),
			"error_message":  attempt.ErrorMessage,
		}
		if len(attempt.Suggestions) > 0 {
			attemptData["suggestions"] = attempt.Suggestions
		}
		builder.WithContext(attemptKey, attemptData)
	}

	// Use detailed failure template with comprehensive message
	return builder.
		WithTemplate(getTemplateForContext(context)).
		Build(getTemplateArgsForContext(context)...)
}

// Helper methods for enhanced functionality

// analyzeExpressionWithContext performs detailed expression analysis with position context
func (v *Variables) analyzeExpressionWithContext(expression string, position, length int) types.VariableAccessAttempt {
	attempt := types.AnalyzeExpression(expression, v.data)

	// Add position information
	attempt.ErrorMessage = fmt.Sprintf("%s (at position %d-%d)", attempt.ErrorMessage, position, position+length)

	return attempt
}

// evaluateExpressionWithDetailedError evaluates an expression with enhanced error details
func (v *Variables) evaluateExpressionWithDetailedError(expression string) (any, error) {
	return expr.Eval(expression, v.data)
}

// enhanceAttemptWithExprError enhances an attempt with detailed expression error information
func (v *Variables) enhanceAttemptWithExprError(attempt *types.VariableAccessAttempt, err error, expression string) {
	attempt.Status = types.VariableStatusExprError
	attempt.FailureReason = types.FailureReasonExpressionError
	attempt.ErrorMessage = v.parseExprError(err.Error(), expression)
	attempt.Suggestions = v.generateExprErrorSuggestions(err.Error(), expression)
}

// parseExprError parses expr-lang error messages to provide clearer explanations
func (v *Variables) parseExprError(errorMsg, expression string) string {
	// Common expr-lang error patterns and their explanations
	errorPatterns := map[string]string{
		"undefined":                "Variable not found",
		"cannot fetch property":    "Property access failed",
		"unexpected token":         "Syntax error",
		"invalid operation":        "Operation not supported for this type",
		"cannot use operator":      "Invalid operator for these types",
		"array index out of range": "Array index exceeds bounds",
	}

	errorMsgLower := strings.ToLower(errorMsg)
	for pattern, explanation := range errorPatterns {
		if strings.Contains(errorMsgLower, pattern) {
			return fmt.Sprintf("%s: %s", explanation, errorMsg)
		}
	}

	return fmt.Sprintf("Expression evaluation error: %s", errorMsg)
}

// generateExprErrorSuggestions generates helpful suggestions based on expression errors
func (v *Variables) generateExprErrorSuggestions(errorMsg, expression string) []string {
	suggestions := []string{}
	errorMsgLower := strings.ToLower(errorMsg)

	if strings.Contains(errorMsgLower, "undefined") {
		suggestions = append(suggestions, "Check if the variable is defined in the test variables section")
		suggestions = append(suggestions, "Verify variable name spelling and case sensitivity")

		// Suggest similar variables
		parts := strings.Split(expression, ".")
		if len(parts) > 0 {
			similar := v.findSimilarVariableNames(parts[0])
			if len(similar) > 0 {
				suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s?", strings.Join(similar, ", ")))
			}
		}
	}

	if strings.Contains(errorMsgLower, "cannot fetch property") {
		suggestions = append(suggestions, "Check if the parent object exists and has the requested property")
		suggestions = append(suggestions, "Use variable action to inspect the object structure")
		suggestions = append(suggestions, "Verify the property name is spelled correctly")
	}

	if strings.Contains(errorMsgLower, "unexpected token") || strings.Contains(errorMsgLower, "syntax") {
		suggestions = append(suggestions, "Check expression syntax for missing operators or parentheses")
		suggestions = append(suggestions, "Verify all quotes and brackets are properly closed")
		suggestions = append(suggestions, "Review expr-lang documentation for supported syntax")
	}

	if strings.Contains(errorMsgLower, "invalid operation") || strings.Contains(errorMsgLower, "cannot use operator") {
		suggestions = append(suggestions, "Check that the operation is supported for the variable types")
		suggestions = append(suggestions, "Consider type conversion (e.g., string to number)")
		suggestions = append(suggestions, "Use appropriate operators for the data types involved")
	}

	if strings.Contains(errorMsgLower, "array") && strings.Contains(errorMsgLower, "index") {
		suggestions = append(suggestions, "Check array bounds - indices start at 0")
		suggestions = append(suggestions, "Use .length to get array size")
		suggestions = append(suggestions, "Verify the variable is actually an array")
	}

	// Generic suggestions
	suggestions = append(suggestions, "Use variable action to debug: action: variable, args: [\""+expression+"\"]")

	return suggestions
}

// generateSyntaxSuggestions generates suggestions for syntax errors
func (v *Variables) generateSyntaxSuggestions(template string) []string {
	suggestions := []string{}

	if strings.Contains(template, "${") && !strings.Contains(template, "}") {
		suggestions = append(suggestions, "Add missing closing brace '}'")
		suggestions = append(suggestions, "Check for unmatched variable expressions")
	}

	if strings.Count(template, "${") > strings.Count(template, "}") {
		suggestions = append(suggestions, "You have more opening braces than closing braces")
	}

	if strings.Contains(template, "$") && !strings.Contains(template, "{") {
		suggestions = append(suggestions, "Use ${variable} syntax for variable substitution")
		suggestions = append(suggestions, "Check if you meant to use a variable expression")
	}

	return suggestions
}

// addTypeInformationToAttempt adds detailed type information to a successful attempt
func (v *Variables) addTypeInformationToAttempt(attempt *types.VariableAccessAttempt, value any) {
	if value == nil {
		return
	}

	// Add type information to suggestions for debugging
	typeInfo := fmt.Sprintf("Resolved to %T with value: %v", value, value)
	attempt.Suggestions = append(attempt.Suggestions, typeInfo)

	// Add size information for collections
	switch v := value.(type) {
	case []any:
		attempt.Suggestions = append(attempt.Suggestions, fmt.Sprintf("Array with %d elements", len(v)))
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		attempt.Suggestions = append(attempt.Suggestions, fmt.Sprintf("Object with keys: %s", strings.Join(keys, ", ")))
	case string:
		attempt.Suggestions = append(attempt.Suggestions, fmt.Sprintf("String with length %d", len(v)))
	}
}

// validateFinalResult performs final validation and generates warnings
func (v *Variables) validateFinalResult(result string, context *types.VariableContext) {
	// Check for unresolved markers
	unresolvedMarkers := []string{"__UNRESOLVED_", "__NULL_", "__CIRCULAR_REF_"}
	for _, marker := range unresolvedMarkers {
		if strings.Contains(result, marker) {
			fmt.Printf("[WARN] Unresolved variable markers found in final result: %q\n", result)
			break
		}
	}

	// Performance warning for too many variables
	if len(context.AttemptedVariables) > 20 {
		fmt.Printf("[WARN] Template contains many variables (%d) - consider simplifying for better performance\n", len(context.AttemptedVariables))
	}

	// Warning for high failure rate
	if context.UnresolvedCount > 0 && context.ResolvedCount > 0 {
		failureRate := float64(context.UnresolvedCount) / float64(context.UnresolvedCount+context.ResolvedCount)
		if failureRate > 0.5 {
			fmt.Printf("[WARN] High variable resolution failure rate: %.1f%% (%d/%d failed)\n",
				failureRate*100, context.UnresolvedCount, context.UnresolvedCount+context.ResolvedCount)
		}
	}
}

// findSimilarVariableNames finds variables with similar names for suggestions
func (v *Variables) findSimilarVariableNames(target string) []string {
	var similar []string
	target = strings.ToLower(target)

	for key := range v.data {
		keyLower := strings.ToLower(key)

		// Exact match (case insensitive)
		if keyLower == target {
			similar = append(similar, key)
			continue
		}

		// Contains match
		if strings.Contains(keyLower, target) || strings.Contains(target, keyLower) {
			similar = append(similar, key)
			continue
		}

		// Simple distance check (allow 1-2 character differences)
		if v.calculateEditDistance(target, keyLower) <= 2 {
			similar = append(similar, key)
		}
	}

	return similar
}

// calculateEditDistance calculates a simple edit distance between two strings
func (v *Variables) calculateEditDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}

	differences := 0
	for i := 0; i < len(s1); i++ {
		if i < len(s2) && s1[i] != s2[i] {
			differences++
		}
	}
	differences += len(s2) - len(s1)

	return differences
}

// Helper functions

func getTemplateForContext(context *types.VariableContext) string {
	// Use the templates map directly
	templates := map[string]string{
		"variable.detailed_failure": "%s",
	}

	if template, exists := templates["variable.detailed_failure"]; exists {
		return template
	}

	return "Variable resolution failed: %s"
}

func getTemplateArgsForContext(context *types.VariableContext) []any {
	return []any{context.GetDetailedErrorMessage()}
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetSnapshot returns a copy of all current variables for context enrichment
func (v *Variables) GetSnapshot() map[string]interface{} {
	snapshot := make(map[string]interface{})
	for k, val := range v.data {
		snapshot[k] = val
	}
	return snapshot
}
