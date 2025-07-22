package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/templates"
	"github.com/JianLoong/robogo/internal/types"
)

// BasicExecutionStrategy handles simple action execution without any control flow
type BasicExecutionStrategy struct {
	variables      *common.Variables
	actionRegistry *actions.ActionRegistry
}

// NewBasicExecutionStrategy creates a new basic execution strategy
func NewBasicExecutionStrategy(variables *common.Variables, actionRegistry *actions.ActionRegistry) *BasicExecutionStrategy {
	return &BasicExecutionStrategy{
		variables:      variables,
		actionRegistry: actionRegistry,
	}
}

// Execute performs basic action execution directly
func (s *BasicExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) *types.StepResult {
	start := time.Now()

	result := &types.StepResult{
		Name:   step.Name,
		Action: step.Action,
		Result: types.ActionResult{Status: constants.ActionStatusError},
	}

	// Get action from registry
	action, exists := s.actionRegistry.Get(step.Action)
	if !exists {
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "UNKNOWN_ACTION").
			WithTemplate(templates.GetTemplateConstant(constants.TemplateUnknownAction)).
			WithContext("action", step.Action).
			WithContext("step", step.Name).
			Build(step.Action)
		
		result.Result = errorResult
		result.Duration = time.Since(start)
		return result
	}

	// Substitute variables in arguments
	args := s.variables.SubstituteArgs(step.Args)

	// Substitute variables in options
	options := make(map[string]any)
	for k, v := range step.Options {
		if str, ok := v.(string); ok {
			options[k] = s.variables.Substitute(str)
		} else {
			options[k] = v
		}
	}
	
	// Pass security information to actions for security-aware behavior
	if step.NoLog {
		options["__no_log"] = true
	}
	if len(step.SensitiveFields) > 0 {
		// Convert []string to []any for options interface
		sensitiveFieldsAny := make([]any, len(step.SensitiveFields))
		for i, field := range step.SensitiveFields {
			sensitiveFieldsAny[i] = field
		}
		options["sensitive_fields"] = sensitiveFieldsAny
	}

	// Print step execution details (unless no_log is enabled)
	if !step.NoLog {
		// Apply masking using step-level sensitive fields
		maskedArgs := s.getMaskedArgsForPrinting(step.Action, args, step.SensitiveFields)
		s.printStepExecution(step, stepNum, maskedArgs, options)
	} else {
		// For no_log steps, print minimal info without sensitive details
		fmt.Printf("Step %d: %s [no_log enabled]\n", stepNum, step.Name)
		fmt.Printf("  Action: %s\n", step.Action)
		fmt.Println("  Executing... ")
	}

	// Execute action directly
	output := action(args, options, s.variables)
	result.Duration = time.Since(start)
	result.Result = output

	// Print execution result (unless no_log is enabled)
	if !step.NoLog {
		s.printStepResult(output, result.Duration)
	} else {
		// For no_log steps, print only status and duration, no sensitive data
		s.printSecureStepResult(output, result.Duration)
	}

	// Apply extraction if specified and action was successful
	var finalData any = output.Data
	if step.Extract != nil && output.Status == constants.ActionStatusPassed {
		extractedData, err := s.applyExtraction(output.Data, step.Extract)
		if err != nil {
			errorResult := types.NewErrorBuilder(types.ErrorCategoryExecution, "EXTRACTION_FAILED").
				WithTemplate("Failed to extract data: %s").
				WithContext("extraction_type", step.Extract.Type).
				WithContext("extraction_path", step.Extract.Path).
				WithContext("error", err.Error()).
				Build(err)
			result.Result = errorResult
			return result
		}
		finalData = extractedData
		result.Result.Data = finalData
	}

	// Store result variable if specified and action was successful
	if step.Result != "" && (output.Status == constants.ActionStatusPassed || finalData != nil) {
		s.variables.Set(step.Result, finalData)
	}

	return result
}

// CanHandle returns true for steps that have an action and no control flow
func (s *BasicExecutionStrategy) CanHandle(step types.Step) bool {
	return step.Action != "" && 
		step.Retry == nil && 
		step.If == "" && 
		step.For == "" && 
		step.While == "" &&
		len(step.Steps) == 0
}

// Priority returns low priority as this is the fallback strategy
func (s *BasicExecutionStrategy) Priority() int {
	return 1
}

// Helper methods for extraction

// applyExtraction applies the specified extraction to the data
func (s *BasicExecutionStrategy) applyExtraction(data any, config *types.ExtractConfig) (any, error) {
	if data == nil {
		return nil, types.NewNilDataError()
	}

	switch config.Type {
	case "jq":
		return s.applyJQExtraction(data, config.Path)
	case "xpath":
		return s.applyXPathExtraction(data, config.Path)
	case "regex":
		return s.applyRegexExtraction(data, config.Path, config.Group)
	default:
		return nil, types.NewUnsupportedExtractionTypeError(config.Type)
	}
}

// applyJQExtraction applies JQ extraction to data
func (s *BasicExecutionStrategy) applyJQExtraction(data any, path string) (any, error) {
	jqAction, exists := s.actionRegistry.Get("jq")
	if !exists {
		return nil, types.NewExtractionError("jq action not available")
	}
	
	result := jqAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetMessage())
	}
	
	return result.Data, nil
}

// applyXPathExtraction applies XPath extraction to data  
func (s *BasicExecutionStrategy) applyXPathExtraction(data any, path string) (any, error) {
	xpathAction, exists := s.actionRegistry.Get("xpath")
	if !exists {
		return nil, types.NewExtractionError("xpath action not available")
	}
	
	result := xpathAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetMessage())
	}
	
	return result.Data, nil
}

// applyRegexExtraction applies regex extraction to data
func (s *BasicExecutionStrategy) applyRegexExtraction(data any, pattern string, group int) (any, error) {
	// Convert data to string
	var text string
	switch v := data.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		text = fmt.Sprintf("%v", v)
	}
	
	// Apply regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, types.NewInvalidRegexPatternError(pattern, err.Error())
	}
	
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return nil, types.NewNoRegexMatchError(pattern)
	}
	
	// Default to group 1, or use specified group
	if group == 0 {
		group = 1
	}
	
	if group >= len(matches) {
		return nil, types.NewInvalidCaptureGroupError(group, len(matches)-1)
	}
	
	return matches[group], nil
}

// printStepExecution prints step execution details to console
func (s *BasicExecutionStrategy) printStepExecution(
	step types.Step,
	stepNum int,
	args []any,
	options map[string]any,
) {
	fmt.Printf("Step %d: %s\n", stepNum, step.Name)
	fmt.Printf("  Action: %s\n", step.Action)

	if len(args) > 0 {
		// Args are already masked at this point
		fmt.Printf("  Args: %v\n", args)
	}

	if len(options) > 0 {
		fmt.Printf("  Options: %v\n", options)
	}

	// Show conditions if present
	if step.If != "" {
		condition := s.variables.Substitute(step.If)
		fmt.Printf("  If: %s\n", condition)
	}

	if step.For != "" {
		forValue := s.variables.Substitute(step.For)
		fmt.Printf("  For: %s\n", forValue)
	}

	if step.While != "" {
		whileValue := s.variables.Substitute(step.While)
		fmt.Printf("  While: %s\n", whileValue)
	}

	if step.Result != "" {
		fmt.Printf("  Result Variable: %s\n", step.Result)
	}

	fmt.Println("  Executing... ")
}

// printStepResult prints the result of step execution
func (s *BasicExecutionStrategy) printStepResult(result types.ActionResult, duration time.Duration) {
	// Print status with color-like indicators
	switch result.Status {
	case constants.ActionStatusPassed:
		fmt.Printf("✓ PASSED (%s)\n", duration)
	case constants.ActionStatusFailed:
		fmt.Printf("✗ FAILED (%s)\n", duration)
		if errorMsg := result.GetMessage(); errorMsg != "" {
			fmt.Printf("    Error: %s\n", errorMsg)
		}
	case constants.ActionStatusSkipped:
		fmt.Printf("- SKIPPED (%s)\n", duration)
		if skipReason := result.GetSkipReason(); skipReason != "" {
			fmt.Printf("    Reason: %s\n", skipReason)
		}
	case constants.ActionStatusError:
		fmt.Printf("! ERROR (%s)\n", duration)
		if errorMsg := result.GetMessage(); errorMsg != "" {
			fmt.Printf("    Error: %s\n", errorMsg)
		}
	default:
		fmt.Printf("? %s (%s)\n", result.Status, duration)
	}

	// Show result data if present and not too large
	if result.Data != nil {
		dataStr := fmt.Sprintf("%v", result.Data)
		if len(dataStr) <= 100 { // Only show small data to avoid cluttering output
			fmt.Printf("    Data: %s\n", dataStr)
		} else {
			fmt.Printf("    Data: [%d characters]\n", len(dataStr))
		}
	}

	fmt.Println() // Add blank line for readability
}

// maskSensitiveArgs masks sensitive information in step arguments based on action type
func (s *BasicExecutionStrategy) maskSensitiveArgs(action string, args []any) []any {
	maskedArgs := make([]any, len(args))
	copy(maskedArgs, args)
	
	switch action {
	case "postgres", "spanner":
		// Database actions: mask connection strings (usually first argument)
		if len(args) > 0 {
			if connStr, ok := args[0].(string); ok {
				maskedArgs[0] = common.MaskConnectionString(connStr)
			}
		}
		
	case "http":
		// HTTP actions: mask request bodies that might contain sensitive data
		if len(args) > 2 { // method, url, body
			if bodyStr, ok := args[2].(string); ok {
				maskedArgs[2] = s.maskHTTPBody(bodyStr)
			}
		}
		
	case "kafka", "rabbitmq":
		// Messaging actions: mask connection strings/brokers (usually second argument)
		if len(args) > 1 {
			if connStr, ok := args[1].(string); ok {
				maskedArgs[1] = common.MaskConnectionString(connStr)
			}
		}
		
	case "assert":
		// Assertion actions: be careful with sensitive comparison values
		for i, arg := range args {
			if str, ok := arg.(string); ok {
				maskedArgs[i] = s.maskSensitiveStringArg(str)
			}
		}
		
	case "log":
		// Log actions: mask any sensitive data in log messages
		for i, arg := range args {
			if str, ok := arg.(string); ok {
				maskedArgs[i] = s.maskSensitiveStringArg(str)
			}
		}
		
	default:
		// For all other actions, scan string arguments for sensitive patterns
		for i, arg := range args {
			if str, ok := arg.(string); ok {
				maskedArgs[i] = s.maskSensitiveStringArg(str)
			}
		}
	}
	
	return maskedArgs
}

// maskHTTPBody masks sensitive data in HTTP request bodies
func (s *BasicExecutionStrategy) maskHTTPBody(body string) string {
	// Use the same sophisticated JSON-aware masking as the HTTP action
	return s.maskSensitiveHTTPData(body)
}

// maskSensitiveStringArg masks sensitive data in string arguments
func (s *BasicExecutionStrategy) maskSensitiveStringArg(str string) string {
	// Use common security utilities for general string masking
	return common.MaskSensitiveData(str, common.DefaultSensitiveKeys)
}

// getMaskedArgsForPrinting returns masked arguments for printing, considering step-level sensitive_fields
func (s *BasicExecutionStrategy) getMaskedArgsForPrinting(action string, args []any, sensitiveFields []string) []any {
	// Start with the standard masking
	maskedArgs := s.maskSensitiveArgs(action, args)
	
	// Apply additional masking with step-level custom sensitive fields
	if len(sensitiveFields) > 0 {
		// Apply additional masking with custom keys
		for i, arg := range maskedArgs {
			if str, ok := arg.(string); ok {
				// For HTTP actions, use sophisticated JSON-aware masking for body arguments
				if action == "http" && i == 2 { // HTTP body is the 3rd argument
					maskedArgs[i] = s.maskSensitiveHTTPDataWithCustom(str, sensitiveFields)
				} else {
					// For other arguments and actions, use general string masking
					maskedArgs[i] = common.MaskSensitiveData(str, sensitiveFields)
				}
			}
		}
	}
	
	return maskedArgs
}

// printSecureStepResult prints the result of step execution for no_log steps
// Only shows status and duration, no sensitive data
func (s *BasicExecutionStrategy) printSecureStepResult(result types.ActionResult, duration time.Duration) {
	// Print status with color-like indicators, but no sensitive data
	switch result.Status {
	case constants.ActionStatusPassed:
		fmt.Printf("✓ PASSED (%s) [no sensitive data logged]\n", duration)
	case constants.ActionStatusFailed:
		fmt.Printf("✗ FAILED (%s) [no sensitive data logged]\n", duration)
		// Don't show error message as it might contain sensitive information
		fmt.Printf("    Error details suppressed for security\n")
	case constants.ActionStatusSkipped:
		fmt.Printf("- SKIPPED (%s) [no sensitive data logged]\n", duration)
		fmt.Printf("    Reason details suppressed for security\n")
	case constants.ActionStatusError:
		fmt.Printf("! ERROR (%s) [no sensitive data logged]\n", duration)
		fmt.Printf("    Error details suppressed for security\n")
	default:
		fmt.Printf("? %s (%s) [no sensitive data logged]\n", result.Status, duration)
	}

	// Never show result data for no_log steps
	fmt.Println() // Add blank line for readability
}

// maskSensitiveHTTPData masks sensitive information in HTTP request bodies
// This mirrors the implementation from the HTTP action for consistency
func (s *BasicExecutionStrategy) maskSensitiveHTTPData(data string) string {
	// Try to parse as JSON first for more intelligent masking
	var jsonData map[string]any
	if json.Unmarshal([]byte(data), &jsonData) == nil {
		// For JSON data, use field-based masking
		return s.maskJSONSensitiveFields(data)
	}
	
	// Fallback to regex-based masking for non-JSON data
	sensitiveKeys := []string{
		"password", "pass", "passwd", "pwd",
		"secret", "token", "key", "apikey", "api_key",
		"authorization", "auth", "bearer",
		"credential", "cred", "access_token", "refresh_token",
		"session", "cookie", "jwt",
	}
	
	result := data
	for _, key := range sensitiveKeys {
		// Match various patterns: "key":"value", key=value, key: value
		patterns := []string{
			fmt.Sprintf(`(?i)"%s"\s*:\s*"[^"]*"`, key),
			fmt.Sprintf(`(?i)"%s"\s*:\s*'[^']*'`, key),
			fmt.Sprintf(`(?i)%s\s*=\s*"[^"]*"`, key),
			fmt.Sprintf(`(?i)%s\s*=\s*'[^']*'`, key),
			fmt.Sprintf(`(?i)%s\s*=\s*[^\s&;]+`, key),
		}
		
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			result = re.ReplaceAllStringFunc(result, func(match string) string {
				// Keep the key but mask the value
				if strings.Contains(match, ":") {
					if strings.Contains(match, `"`) {
						return fmt.Sprintf(`"%s": "***"`, key)
					} else {
						return fmt.Sprintf(`"%s": '***'`, key)
					}
				} else {
					return fmt.Sprintf(`%s=***`, key)
				}
			})
		}
	}
	
	return result
}

// maskJSONSensitiveFields masks sensitive fields in JSON strings
func (s *BasicExecutionStrategy) maskJSONSensitiveFields(jsonStr string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr // Return original if not valid JSON
	}
	
	s.maskSensitiveJSONValues(data)
	
	maskedBytes, err := json.Marshal(data)
	if err != nil {
		return jsonStr // Return original if can't re-marshal
	}
	
	return string(maskedBytes)
}

// maskSensitiveJSONValues recursively masks sensitive values in JSON objects
func (s *BasicExecutionStrategy) maskSensitiveJSONValues(obj map[string]any) {
	sensitiveKeys := []string{
		"password", "pass", "passwd", "pwd",
		"secret", "token", "key", "apikey", "api_key", "access_token",
		"authorization", "auth", "bearer", "credential", "cred",
		"session", "cookie", "jwt", "refresh_token",
	}
	
	for key, value := range obj {
		lowerKey := strings.ToLower(key)
		
		// Check if this key should be masked
		shouldMask := false
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				shouldMask = true
				break
			}
		}
		
		if shouldMask {
			obj[key] = "***"
		} else if nested, ok := value.(map[string]any); ok {
			// Recursively process nested objects
			s.maskSensitiveJSONValues(nested)
		}
	}
}

// maskSensitiveHTTPDataWithCustom masks sensitive information in HTTP request bodies with custom fields
func (s *BasicExecutionStrategy) maskSensitiveHTTPDataWithCustom(data string, customKeys []string) string {
	// Combine default sensitive keys with custom fields
	sensitiveKeys := []string{
		"password", "pass", "passwd", "pwd",
		"secret", "token", "key", "apikey", "api_key",
		"authorization", "auth", "bearer",
		"credential", "cred", "access_token", "refresh_token",
		"session", "cookie", "jwt",
	}
	sensitiveKeys = append(sensitiveKeys, customKeys...)
	
	// Try to parse as JSON first for more intelligent masking
	var jsonData map[string]any
	if json.Unmarshal([]byte(data), &jsonData) == nil {
		// For JSON data, use field-based masking with custom fields
		return s.maskJSONSensitiveFieldsWithCustom(data, sensitiveKeys)
	}
	
	// Fallback to regex-based masking for non-JSON data
	result := data
	for _, key := range sensitiveKeys {
		// Match various patterns: "key":"value", key=value, key: value
		patterns := []string{
			fmt.Sprintf(`(?i)"%s"\s*:\s*"[^"]*"`, key),
			fmt.Sprintf(`(?i)"%s"\s*:\s*'[^']*'`, key),
			fmt.Sprintf(`(?i)%s\s*=\s*"[^"]*"`, key),
			fmt.Sprintf(`(?i)%s\s*=\s*'[^']*'`, key),
			fmt.Sprintf(`(?i)%s\s*=\s*[^\s&;]+`, key),
		}
		
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			result = re.ReplaceAllStringFunc(result, func(match string) string {
				// Keep the key but mask the value
				if strings.Contains(match, ":") {
					if strings.Contains(match, `"`) {
						return fmt.Sprintf(`"%s": "***"`, key)
					} else {
						return fmt.Sprintf(`"%s": '***'`, key)
					}
				} else {
					return fmt.Sprintf(`%s=***`, key)
				}
			})
		}
	}
	
	return result
}

// maskJSONSensitiveFieldsWithCustom masks sensitive fields in JSON strings with custom keys
func (s *BasicExecutionStrategy) maskJSONSensitiveFieldsWithCustom(jsonStr string, sensitiveKeys []string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr // Return original if not valid JSON
	}
	
	s.maskSensitiveJSONValuesWithCustom(data, sensitiveKeys)
	
	maskedBytes, err := json.Marshal(data)
	if err != nil {
		return jsonStr // Return original if can't re-marshal
	}
	
	return string(maskedBytes)
}

// maskSensitiveJSONValuesWithCustom recursively masks sensitive values in JSON objects with custom keys
func (s *BasicExecutionStrategy) maskSensitiveJSONValuesWithCustom(obj map[string]any, sensitiveKeys []string) {
	for key, value := range obj {
		lowerKey := strings.ToLower(key)
		
		// Check if this key should be masked
		shouldMask := false
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				shouldMask = true
				break
			}
		}
		
		if shouldMask {
			obj[key] = "***"
		} else if nested, ok := value.(map[string]any); ok {
			// Recursively process nested objects
			s.maskSensitiveJSONValuesWithCustom(nested, sensitiveKeys)
		}
	}
}