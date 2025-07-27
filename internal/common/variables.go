package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Variables provides simple variable storage and substitution
type Variables struct {
	data map[string]any
}

// NewVariables creates a new Variables instance
func NewVariables() *Variables {
	return &Variables{
		data: make(map[string]any),
	}
}

// Set stores a variable
func (v *Variables) Set(key string, value any) {
	v.data[key] = value
}

// Get retrieves a variable
func (v *Variables) Get(key string) any {
	return v.data[key]
}

// Has checks if a variable exists
func (v *Variables) Has(key string) bool {
	_, exists := v.data[key]
	return exists
}

// Load bulk loads variables with environment variable substitution
func (v *Variables) Load(vars map[string]any) {
	for key, value := range vars {
		if strValue, ok := value.(string); ok {
			// Substitute environment variables in string values
			substituted := v.Substitute(strValue)
			v.Set(key, substituted)
		} else {
			v.Set(key, value)
		}
	}
}

// GetSnapshot returns a copy of all current variables
func (v *Variables) GetSnapshot() map[string]interface{} {
	snapshot := make(map[string]interface{}, len(v.data))
	for key, value := range v.data {
		snapshot[key] = value
	}
	return snapshot
}

// Substitute performs variable substitution using ${variable} syntax
func (v *Variables) Substitute(template string) string {
	result := template

	// Handle ${ENV:VARIABLE_NAME} syntax for environment variables
	for {
		start := strings.Index(result, "${ENV:")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		// Extract environment variable name
		envVar := result[start+6 : end] // Skip "${ENV:"
		envValue := os.Getenv(envVar)

		// Replace with environment value
		result = result[:start] + envValue + result[end+1:]
	}

	// Handle ${variable} syntax for stored variables
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

		// Extract variable name
		varName := result[start+2 : end] // Skip "${"

		// Skip if this is an ENV: variable (already handled above)
		if strings.HasPrefix(varName, "ENV:") {
			// Find next occurrence
			nextStart := strings.Index(result[end+1:], "${")
			if nextStart == -1 {
				break
			}
			continue
		}

		// Check if this is dot notation (e.g., "response.status_code")
		if strings.Contains(varName, ".") {
			resolvedValue := v.resolveDotNotation(varName)
			result = result[:start] + resolvedValue + result[end+1:]
		} else {
			// Replace with stored variable value
			if value, exists := v.data[varName]; exists {
				strValue := ""
				if value != nil {
					strValue = strings.TrimSpace(strings.Trim(strings.Trim(strings.Trim(fmt.Sprintf("%v", value), "\""), "'"), "`"))
				}
				result = result[:start] + strValue + result[end+1:]
			} else {
				// Mark as unresolved but continue processing
				result = result[:start] + "__UNRESOLVED_" + varName + "__" + result[end+1:]
			}
		}
	}

	return result
}

// SubstituteArgs performs variable substitution on arguments
func (v *Variables) SubstituteArgs(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		result[i] = v.substituteInData(arg)
	}
	return result
}

// substituteInData recursively substitutes variables in nested data structures
func (v *Variables) substituteInData(data any) any {
	switch val := data.(type) {
	case string:
		// Check if this is a simple variable reference like "${var_name}"
		if v.isSimpleVariableReference(val) {
			// For simple variable references, return the actual value, not string conversion
			varName := val[2 : len(val)-1] // Remove ${ and }
			if v.Has(varName) {
				return v.Get(varName)
			}
		}
		// For complex templates or non-variable strings, do normal substitution
		return v.Substitute(val)
	case map[string]any:
		result := make(map[string]any)
		for key, value := range val {
			// Substitute variables in both keys and values
			substitutedKey := v.Substitute(key)
			result[substitutedKey] = v.substituteInData(value)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, value := range val {
			result[i] = v.substituteInData(value)
		}
		return result
	case map[any]any:
		result := make(map[any]any)
		for key, value := range val {
			// Handle keys that might be strings needing substitution
			var substitutedKey any = key
			if keyStr, ok := key.(string); ok {
				substitutedKey = v.Substitute(keyStr)
			}
			result[substitutedKey] = v.substituteInData(value)
		}
		return result
	default:
		// For other types (numbers, booleans, etc.), return as-is
		return data
	}
}

// isSimpleVariableReference checks if a string is exactly "${variable_name}" with no other content
func (v *Variables) isSimpleVariableReference(str string) bool {
	if !strings.HasPrefix(str, "${") || !strings.HasSuffix(str, "}") {
		return false
	}

	// Check if there's only one variable and nothing else
	content := str[2 : len(str)-1] // Remove ${ and }

	// Simple variable name should not contain spaces or special characters except ENV: prefix
	if strings.Contains(content, " ") || strings.Contains(content, "${") {
		return false
	}

	return true
}

// resolveDotNotation resolves dot notation like "response.status_code" or "user.profile.name"
func (v *Variables) resolveDotNotation(dotPath string) string {
	parts := strings.Split(dotPath, ".")
	if len(parts) < 2 {
		return "__UNRESOLVED_" + dotPath + "__"
	}

	// Get the root variable
	rootVar := parts[0]
	value, exists := v.data[rootVar]
	if !exists {
		return "__UNRESOLVED_" + dotPath + "__"
	}

	// Navigate through the dot path
	current := value
	for i, field := range parts[1:] {
		current = v.getFieldValue(current, field)
		if current == nil {
			// Build the path up to the failed field for better error reporting
			failedPath := strings.Join(parts[:i+2], ".")
			return "__UNRESOLVED_" + failedPath + "__"
		}
	}

	// Convert final value to string
	if current == nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(strings.Trim(strings.Trim(fmt.Sprintf("%v", current), "\""), "'"), "`"))
}

// getFieldValue extracts a field value from various data types
func (v *Variables) getFieldValue(data any, field string) any {
	if data == nil {
		return nil
	}

	switch val := data.(type) {
	case map[string]any:
		return val[field]
	case map[any]any:
		// Try the field as-is first
		if value, exists := val[field]; exists {
			return value
		}
		// Try to find by string conversion of keys
		for key, value := range val {
			if fmt.Sprintf("%v", key) == field {
				return value
			}
		}
		return nil
	case []any:
		// Handle array indexing (e.g., "items.0" for first element)
		if index, err := strconv.Atoi(field); err == nil && index >= 0 && index < len(val) {
			return val[index]
		}
		return nil
	default:
		// For other types, we can't navigate further
		return nil
	}
}

// Clone creates a copy of the Variables with the same data
func (v *Variables) Clone() *Variables {
	newVars := NewVariables()
	for key, value := range v.data {
		newVars.data[key] = value
	}
	return newVars
}
