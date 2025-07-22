package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
)

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