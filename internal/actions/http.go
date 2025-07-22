package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// httpAction performs an HTTP request. It always returns status code, headers, and raw body.
func httpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {

	if len(args) < 2 {
		return types.MissingArgsError("http", 2, len(args))
	}

	// Check for unresolved variables in critical arguments (method and URL)
	if errorResult := validateArgsResolved("http", args[:2]); errorResult != nil {
		return *errorResult
	}

	method := fmt.Sprintf("%v", args[0])
	url := fmt.Sprintf("%v", args[1])

	// Extract request headers for context first (needed for body processing)
	var requestHeaders map[string]string
	if headers, ok := options["headers"].(map[string]any); ok {
		requestHeaders = make(map[string]string)
		for key, value := range headers {
			requestHeaders[key] = fmt.Sprintf("%v", value)
		}
	}

	var bodyReader io.Reader
	if len(args) > 2 {
		// Get the body argument
		bodyArg := args[2]

		// Always convert the body to a string first
		var bodyStr string

		// Check content type for special handling
		contentType := ""
		if headers, ok := options["headers"].(map[string]any); ok {
			for k, v := range headers {
				if strings.ToLower(k) == "content-type" {
					contentType = strings.ToLower(fmt.Sprintf("%v", v))
					break
				}
			}
		}

		// Special handling for JSON content type with map/slice data
		if (contentType == "application/json" || strings.HasPrefix(contentType, "application/json")) &&
			(isMap(bodyArg) || isSlice(bodyArg)) {
			// For JSON content type with structured data, serialize it
			jsonData, err := json.Marshal(bodyArg)
			if err != nil {
				return types.RequestError("JSON marshaling", err.Error())
			}
			bodyStr = string(jsonData)
		} else {
			// For all other cases, just convert to string
			bodyStr = fmt.Sprintf("%v", bodyArg)
		}

		// Create the body reader from the string
		bodyReader = strings.NewReader(bodyStr)

		// Log the request body for debugging if debug option is set
		if debugOpt, ok := options["debug"].(bool); ok && debugOpt {
			// Check if this is a no_log step
			if noLogOpt, ok := options["__no_log"].(bool); ok && noLogOpt {
				fmt.Printf("HTTP Request Body: [body suppressed - no_log enabled]\n")
			} else {
				// Mask sensitive data in the body before logging
				var maskedBody string
				if sensitiveFields, ok := options["sensitive_fields"]; ok {
					if fieldsSlice, ok := sensitiveFields.([]any); ok {
						// Convert []any to []string for custom sensitive fields
						customKeys := make([]string, len(fieldsSlice))
						for i, field := range fieldsSlice {
							customKeys[i] = fmt.Sprintf("%v", field)
						}
						maskedBody = maskSensitiveHTTPData(bodyStr, customKeys)
					} else {
						maskedBody = maskSensitiveHTTPData(bodyStr)
					}
				} else {
					maskedBody = maskSensitiveHTTPData(bodyStr)
				}
				fmt.Printf("HTTP Request Body: %s\n", maskedBody)
			}
		}
	}

	// Extract timeout for context
	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return types.RequestError(fmt.Sprintf("HTTP %s %s", method, url), err.Error())
	}

	if headers, ok := options["headers"].(map[string]any); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(req)

	if err != nil {
		return types.RequestError(fmt.Sprintf("HTTP %s %s", method, url), err.Error())
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.RequestError(fmt.Sprintf("HTTP %s %s response read", method, url), err.Error())
	}

	respBodyStr := string(responseBody)
	result := map[string]any{
		"status_code": resp.StatusCode,
		"body":        respBodyStr,
		"headers":     resp.Header,
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   result,
	}
}

// Helper functions to check types
func isMap(v any) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	kind := t.Kind()
	return kind == reflect.Map
}

func isSlice(v any) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	kind := t.Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

// maskSensitiveHTTPData masks sensitive information in HTTP request bodies
// Now accepts custom sensitive fields from options for enhanced masking
func maskSensitiveHTTPData(data string, customFields ...[]string) string {
	// Combine default sensitive keys with custom fields
	sensitiveKeys := []string{
		"password", "pass", "passwd", "pwd",
		"secret", "token", "key", "apikey", "api_key",
		"authorization", "auth", "bearer",
		"credential", "cred", "access_token", "refresh_token",
		"session", "cookie", "jwt",
	}
	
	// Add custom sensitive fields if provided
	for _, customKeySet := range customFields {
		sensitiveKeys = append(sensitiveKeys, customKeySet...)
	}
	
	// Try to parse as JSON first for more intelligent masking
	var jsonData map[string]any
	if json.Unmarshal([]byte(data), &jsonData) == nil {
		// For JSON data, use field-based masking with custom fields
		return maskJSONSensitiveFieldsWithCustom(data, sensitiveKeys)
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

// maskJSONSensitiveFields masks sensitive fields in JSON strings using default keys
func maskJSONSensitiveFields(jsonStr string) string {
	defaultKeys := []string{
		"password", "pass", "passwd", "pwd",
		"secret", "token", "key", "apikey", "api_key", "access_token",
		"authorization", "auth", "bearer", "credential", "cred",
		"session", "cookie", "jwt", "refresh_token",
	}
	return maskJSONSensitiveFieldsWithCustom(jsonStr, defaultKeys)
}

// maskJSONSensitiveFieldsWithCustom masks sensitive fields in JSON strings with custom keys
func maskJSONSensitiveFieldsWithCustom(jsonStr string, sensitiveKeys []string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr // Return original if not valid JSON
	}
	
	maskSensitiveJSONValuesWithCustom(data, sensitiveKeys)
	
	maskedBytes, err := json.Marshal(data)
	if err != nil {
		return jsonStr // Return original if can't re-marshal
	}
	
	return string(maskedBytes)
}

// maskSensitiveJSONValues recursively masks sensitive values in JSON objects using default keys
func maskSensitiveJSONValues(obj map[string]any) {
	defaultKeys := []string{
		"password", "pass", "passwd", "pwd",
		"secret", "token", "key", "apikey", "api_key", "access_token",
		"authorization", "auth", "bearer", "credential", "cred",
		"session", "cookie", "jwt", "refresh_token",
	}
	maskSensitiveJSONValuesWithCustom(obj, defaultKeys)
}

// maskSensitiveJSONValuesWithCustom recursively masks sensitive values in JSON objects with custom keys
func maskSensitiveJSONValuesWithCustom(obj map[string]any, sensitiveKeys []string) {
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
			maskSensitiveJSONValuesWithCustom(nested, sensitiveKeys)
		}
	}
}
