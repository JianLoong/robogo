package common

import "strings"

// Variables - DEPRECATED: Use VariableManager instead
// Maintained for backward compatibility during refactoring
type Variables struct {
	manager *VariableManager
}

// NewVariables creates a Variables instance with the new variable manager
func NewVariables() *Variables {
	return &Variables{
		manager: NewVariableManager(),
	}
}

// Set stores a variable
func (v *Variables) Set(key string, value any) {
	v.manager.Set(key, value)
}

// Get retrieves a variable
func (v *Variables) Get(key string) any {
	return v.manager.Get(key)
}

// Load bulk loads variables with environment variable substitution
func (v *Variables) Load(vars map[string]any) {
	v.manager.LoadFromMapWithEnvSubstitution(vars)
}

// GetSnapshot returns a copy of all current variables for context enrichment
func (v *Variables) GetSnapshot() map[string]interface{} {
	return v.manager.GetSnapshot()
}

// Substitute performs variable substitution using ${variable} syntax
func (v *Variables) Substitute(template string) string {
	return v.manager.SubstituteSimple(template)
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
				if v.manager.Has(varName) {
					result[i] = v.manager.Get(varName)
					continue
				}
			}
			// For complex templates or non-variable strings, do normal substitution
			result[i] = v.manager.SubstituteSimple(str)
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
	newVars.manager = v.manager.Clone()
	return newVars
}
