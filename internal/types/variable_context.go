package types

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// VariableResolutionStatus represents the outcome of variable resolution
type VariableResolutionStatus string

const (
	VariableStatusResolved    VariableResolutionStatus = "resolved"
	VariableStatusUnresolved  VariableResolutionStatus = "unresolved"
	VariableStatusPartial     VariableResolutionStatus = "partial"
	VariableStatusExprError   VariableResolutionStatus = "expr_error"
	VariableStatusNested      VariableResolutionStatus = "nested"
	VariableStatusCyclic      VariableResolutionStatus = "cyclic"
)

// VariableFailureReason provides specific reasons for resolution failures
type VariableFailureReason string

const (
	FailureReasonNotFound       VariableFailureReason = "variable_not_found"
	FailureReasonInvalidSyntax  VariableFailureReason = "invalid_syntax"
	FailureReasonExpressionError VariableFailureReason = "expression_error"
	FailureReasonCircularRef    VariableFailureReason = "circular_reference"
	FailureReasonTypeError      VariableFailureReason = "type_error"
	FailureReasonAccessError    VariableFailureReason = "access_error"
	FailureReasonNullValue      VariableFailureReason = "null_value"
)

// VariableAccessAttempt represents a single attempt to access a variable
type VariableAccessAttempt struct {
	Expression      string                   `json:"expression"`
	Status          VariableResolutionStatus `json:"status"`
	FailureReason   VariableFailureReason    `json:"failure_reason,omitempty"`
	ResolvedValue   any                      `json:"resolved_value,omitempty"`
	ErrorMessage    string                   `json:"error_message,omitempty"`
	Timestamp       time.Time                `json:"timestamp"`
	AccessPath      []string                 `json:"access_path,omitempty"`     // e.g., ["user", "profile", "name"]
	AvailableKeys   []string                 `json:"available_keys,omitempty"`  // Available keys at failure point
	Suggestions     []string                 `json:"suggestions,omitempty"`     // Suggested fixes
}

// VariableContext provides comprehensive tracking of variable resolution
type VariableContext struct {
	OriginalTemplate   string                   `json:"original_template"`
	ProcessedTemplate  string                   `json:"processed_template"`
	Status             VariableResolutionStatus `json:"status"`
	AttemptedVariables []VariableAccessAttempt  `json:"attempted_variables"`
	AvailableVariables map[string]any           `json:"available_variables"`
	UnresolvedCount    int                      `json:"unresolved_count"`
	ResolvedCount      int                      `json:"resolved_count"`
	ProcessingTime     time.Duration            `json:"processing_time"`
	Timestamp          time.Time                `json:"timestamp"`
}

// NewVariableContext creates a new VariableContext for tracking resolution
func NewVariableContext(template string, availableVars map[string]any) *VariableContext {
	return &VariableContext{
		OriginalTemplate:   template,
		ProcessedTemplate:  template,
		Status:             VariableStatusUnresolved,
		AttemptedVariables: make([]VariableAccessAttempt, 0),
		AvailableVariables: copyMap(availableVars),
		UnresolvedCount:    0,
		ResolvedCount:      0,
		Timestamp:          time.Now(),
	}
}

// AddAttempt records a variable access attempt
func (vc *VariableContext) AddAttempt(attempt VariableAccessAttempt) {
	attempt.Timestamp = time.Now()
	vc.AttemptedVariables = append(vc.AttemptedVariables, attempt)
	
	if attempt.Status == VariableStatusResolved {
		vc.ResolvedCount++
	} else {
		vc.UnresolvedCount++
	}
	
	// Update overall status
	vc.updateOverallStatus()
}

// updateOverallStatus determines the overall resolution status
func (vc *VariableContext) updateOverallStatus() {
	if vc.UnresolvedCount == 0 && vc.ResolvedCount > 0 {
		vc.Status = VariableStatusResolved
	} else if vc.ResolvedCount > 0 && vc.UnresolvedCount > 0 {
		vc.Status = VariableStatusPartial
	} else if vc.UnresolvedCount > 0 {
		vc.Status = VariableStatusUnresolved
	}
}

// SetProcessedTemplate updates the final processed template
func (vc *VariableContext) SetProcessedTemplate(processed string) {
	vc.ProcessedTemplate = processed
	vc.ProcessingTime = time.Since(vc.Timestamp)
}

// GetUnresolvedVariables returns all variables that failed to resolve
func (vc *VariableContext) GetUnresolvedVariables() []VariableAccessAttempt {
	var unresolved []VariableAccessAttempt
	for _, attempt := range vc.AttemptedVariables {
		if attempt.Status != VariableStatusResolved {
			unresolved = append(unresolved, attempt)
		}
	}
	return unresolved
}

// GetFailureAnalysis provides detailed analysis of resolution failures
func (vc *VariableContext) GetFailureAnalysis() *VariableFailureAnalysis {
	analysis := &VariableFailureAnalysis{
		TotalAttempts:     len(vc.AttemptedVariables),
		SuccessfulCount:   vc.ResolvedCount,
		FailedCount:       vc.UnresolvedCount,
		FailuresByReason:  make(map[VariableFailureReason]int),
		CommonPatterns:    make([]string, 0),
		Recommendations:   make([]string, 0),
	}
	
	// Analyze failure patterns
	for _, attempt := range vc.AttemptedVariables {
		if attempt.Status != VariableStatusResolved {
			analysis.FailuresByReason[attempt.FailureReason]++
		}
	}
	
	// Generate recommendations based on failure patterns
	analysis.generateRecommendations(vc)
	
	return analysis
}

// VariableFailureAnalysis provides insights into variable resolution issues
type VariableFailureAnalysis struct {
	TotalAttempts     int                              `json:"total_attempts"`
	SuccessfulCount   int                              `json:"successful_count"`
	FailedCount       int                              `json:"failed_count"`
	FailuresByReason  map[VariableFailureReason]int    `json:"failures_by_reason"`
	CommonPatterns    []string                         `json:"common_patterns"`
	Recommendations   []string                         `json:"recommendations"`
}

// generateRecommendations creates helpful suggestions based on failure patterns
func (analysis *VariableFailureAnalysis) generateRecommendations(vc *VariableContext) {
	// Analyze not found errors
	if analysis.FailuresByReason[FailureReasonNotFound] > 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Check variable names for typos or case sensitivity issues")
		
		// Suggest similar variable names
		for _, attempt := range vc.AttemptedVariables {
			if attempt.FailureReason == FailureReasonNotFound {
				similar := findSimilarVariables(attempt.Expression, vc.AvailableVariables)
				if len(similar) > 0 {
					analysis.Recommendations = append(analysis.Recommendations,
						fmt.Sprintf("Did you mean one of: %s?", strings.Join(similar, ", ")))
				}
			}
		}
	}
	
	// Analyze expression errors
	if analysis.FailuresByReason[FailureReasonExpressionError] > 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Review expression syntax - check for proper operators and parentheses")
	}
	
	// Analyze access errors
	if analysis.FailuresByReason[FailureReasonAccessError] > 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Verify nested property access paths - check if intermediate objects exist")
	}
	
	// Analyze type errors
	if analysis.FailuresByReason[FailureReasonTypeError] > 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Check data types - ensure operations are compatible with variable types")
	}
	
	// General recommendations
	if analysis.FailedCount > 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Use variable action to debug: 'action: variable, args: [variable_name]'")
		analysis.Recommendations = append(analysis.Recommendations,
			"Check test setup - ensure all required variables are defined")
	}
}

// GetDetailedErrorMessage returns a comprehensive error message
func (vc *VariableContext) GetDetailedErrorMessage() string {
	if vc.Status == VariableStatusResolved {
		return ""
	}
	
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Variable resolution failed in template: '%s'", vc.OriginalTemplate))
	
	if vc.UnresolvedCount > 0 {
		parts = append(parts, fmt.Sprintf("Failed to resolve %d variable(s)", vc.UnresolvedCount))
		
		for _, attempt := range vc.GetUnresolvedVariables() {
			reason := getHumanReadableReason(attempt.FailureReason)
			parts = append(parts, fmt.Sprintf("  - '${%s}': %s", attempt.Expression, reason))
			
			if attempt.ErrorMessage != "" {
				parts = append(parts, fmt.Sprintf("    Error: %s", attempt.ErrorMessage))
			}
			
			if len(attempt.Suggestions) > 0 {
				parts = append(parts, fmt.Sprintf("    Suggestions: %s", strings.Join(attempt.Suggestions, "; ")))
			}
		}
	}
	
	// Show available variables for context
	if len(vc.AvailableVariables) > 0 {
		var availableKeys []string
		for key := range vc.AvailableVariables {
			availableKeys = append(availableKeys, key)
		}
		parts = append(parts, fmt.Sprintf("Available variables: %s", strings.Join(availableKeys, ", ")))
	}
	
	return strings.Join(parts, "\n")
}

// ParseVariableExpression extracts variable expressions from a template
func ParseVariableExpression(template string) []string {
	var expressions []string
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(template, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			expressions = append(expressions, match[1])
		}
	}
	
	return expressions
}

// AnalyzeExpression provides detailed analysis of a variable expression
func AnalyzeExpression(expression string, availableVars map[string]any) VariableAccessAttempt {
	attempt := VariableAccessAttempt{
		Expression: expression,
		Timestamp:  time.Now(),
	}
	
	// Parse access path (e.g., "user.profile.name" -> ["user", "profile", "name"])
	parts := strings.Split(expression, ".")
	attempt.AccessPath = parts
	
	if len(parts) == 0 {
		attempt.Status = VariableStatusUnresolved
		attempt.FailureReason = FailureReasonInvalidSyntax
		attempt.ErrorMessage = "Empty expression"
		return attempt
	}
	
	// Check if root variable exists
	rootVar := parts[0]
	rootValue, exists := availableVars[rootVar]
	
	if !exists {
		attempt.Status = VariableStatusUnresolved
		attempt.FailureReason = FailureReasonNotFound
		attempt.ErrorMessage = fmt.Sprintf("Variable '%s' not found", rootVar)
		attempt.AvailableKeys = getMapKeys(availableVars)
		attempt.Suggestions = findSimilarVariables(rootVar, availableVars)
		return attempt
	}
	
	// If it's a simple variable (no nested access), we're done
	if len(parts) == 1 {
		attempt.Status = VariableStatusResolved
		attempt.ResolvedValue = rootValue
		return attempt
	}
	
	// Handle nested access
	currentValue := rootValue
	currentPath := []string{rootVar}
	
	for i := 1; i < len(parts); i++ {
		key := parts[i]
		currentPath = append(currentPath, key)
		
		// Try to access the nested property
		nextValue, err := accessNestedProperty(currentValue, key)
		if err != nil {
			attempt.Status = VariableStatusUnresolved
			attempt.FailureReason = FailureReasonAccessError
			attempt.ErrorMessage = fmt.Sprintf("Cannot access '%s' in path '%s': %s", 
				key, strings.Join(currentPath, "."), err.Error())
			
			// Suggest available keys at this level
			if availableKeys := getAvailableKeysForValue(currentValue); len(availableKeys) > 0 {
				attempt.AvailableKeys = availableKeys
				attempt.Suggestions = append(attempt.Suggestions, 
					fmt.Sprintf("Available keys at '%s': %s", 
						strings.Join(currentPath[:len(currentPath)-1], "."), 
						strings.Join(availableKeys, ", ")))
			}
			return attempt
		}
		
		currentValue = nextValue
	}
	
	attempt.Status = VariableStatusResolved
	attempt.ResolvedValue = currentValue
	return attempt
}

// Helper functions

func copyMap(original map[string]any) map[string]any {
	copy := make(map[string]any)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func findSimilarVariables(target string, available map[string]any) []string {
	var similar []string
	target = strings.ToLower(target)
	
	for key := range available {
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
		
		// Levenshtein distance (simplified)
		if calculateSimpleDistance(target, keyLower) <= 2 {
			similar = append(similar, key)
		}
	}
	
	return similar
}

func calculateSimpleDistance(s1, s2 string) int {
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

func accessNestedProperty(value any, key string) (any, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot access property '%s' on nil value", key)
	}
	
	// Handle map access
	if m, ok := value.(map[string]any); ok {
		if val, exists := m[key]; exists {
			return val, nil
		}
		return nil, fmt.Errorf("property '%s' does not exist", key)
	}
	
	// Handle slice/array index access
	if s, ok := value.([]any); ok {
		if key == "length" {
			return len(s), nil
		}
		
		// Try to parse as index
		if index, err := strconv.Atoi(key); err == nil {
			if index < 0 || index >= len(s) {
				return nil, fmt.Errorf("index %d out of range [0, %d]", index, len(s)-1)
			}
			return s[index], nil
		}
		return nil, fmt.Errorf("invalid array index '%s'", key)
	}
	
	// Handle struct field access using reflection
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() == reflect.Struct {
		field := v.FieldByName(key)
		if !field.IsValid() {
			return nil, fmt.Errorf("field '%s' does not exist", key)
		}
		if !field.CanInterface() {
			return nil, fmt.Errorf("field '%s' is not accessible", key)
		}
		return field.Interface(), nil
	}
	
	return nil, fmt.Errorf("cannot access property '%s' on type %T", key, value)
}

func getAvailableKeysForValue(value any) []string {
	if value == nil {
		return nil
	}
	
	// Handle map
	if m, ok := value.(map[string]any); ok {
		return getMapKeys(m)
	}
	
	// Handle slice (show available indices)
	if s, ok := value.([]any); ok {
		keys := []string{"length"}
		for i := 0; i < len(s) && i < 10; i++ { // Show first 10 indices
			keys = append(keys, strconv.Itoa(i))
		}
		if len(s) > 10 {
			keys = append(keys, "...")
		}
		return keys
	}
	
	// Handle struct using reflection
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() == reflect.Struct {
		var keys []string
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				keys = append(keys, field.Name)
			}
		}
		return keys
	}
	
	return nil
}

func getHumanReadableReason(reason VariableFailureReason) string {
	switch reason {
	case FailureReasonNotFound:
		return "variable not found"
	case FailureReasonInvalidSyntax:
		return "invalid syntax"
	case FailureReasonExpressionError:
		return "expression evaluation error"
	case FailureReasonCircularRef:
		return "circular reference detected"
	case FailureReasonTypeError:
		return "type incompatibility"
	case FailureReasonAccessError:
		return "property access failed"
	case FailureReasonNullValue:
		return "null or undefined value"
	default:
		return "unknown error"
	}
}