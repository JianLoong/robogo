package common

import (
	"fmt"
	"os"
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
		envVar := result[start+6:end] // Skip "${ENV:"
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
		varName := result[start+2:end] // Skip "${"
		
		// Skip if this is an ENV: variable (already handled above)
		if strings.HasPrefix(varName, "ENV:") {
			// Find next occurrence
			nextStart := strings.Index(result[end+1:], "${")
			if nextStart == -1 {
				break
			}
			continue
		}
		
		// Replace with stored variable value
		if value, exists := v.data[varName]; exists {
			strValue := ""
			if value != nil {
				strValue = strings.TrimSpace(strings.Trim(strings.Trim(strings.Trim(fmt.Sprintf("%v", value), "\""), "'"), "`"))
			}
			result = result[:start] + strValue + result[end+1:]
		} else {
			// Mark as unresolved but continue processing
			result = result[:start] + "__UNRESOLVED_"+varName+"__" + result[end+1:]
		}
	}
	
	return result
}

// SubstituteArgs performs variable substitution on arguments
func (v *Variables) SubstituteArgs(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			// Check if this is a simple variable reference like "${var_name}"
			if v.isSimpleVariableReference(str) {
				// For simple variable references, return the actual value, not string conversion
				varName := str[2 : len(str)-1] // Remove ${ and }
				if v.Has(varName) {
					result[i] = v.Get(varName)
					continue
				}
			}
			// For complex templates or non-variable strings, do normal substitution
			result[i] = v.Substitute(str)
		} else {
			result[i] = arg
		}
	}
	return result
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

// Clone creates a copy of the Variables with the same data
func (v *Variables) Clone() *Variables {
	newVars := NewVariables()
	for key, value := range v.data {
		newVars.data[key] = value
	}
	return newVars
}