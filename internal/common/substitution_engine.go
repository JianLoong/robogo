package common

import (
	"fmt"
	"strings"
)

// SubstitutionEngine handles simple variable substitution without complex expressions
type SubstitutionEngine struct {
	store *VariableStore
}

// NewSubstitutionEngine creates a new simple substitution engine
func NewSubstitutionEngine(store *VariableStore) *SubstitutionEngine {
	return &SubstitutionEngine{
		store: store,
	}
}

// Substitute performs simple variable substitution using ${variable} syntax
// Only supports basic variable names, no complex expressions
func (engine *SubstitutionEngine) Substitute(template string) string {
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

		varName := result[start+2 : end]
		
		// Only allow simple variable names (alphanumeric + underscore)
		if !isSimpleVariableName(varName) {
			fmt.Printf("[WARN] Complex variable expression not supported: %q\n", varName)
			fmt.Printf("[HINT] Use jq action for complex data access instead of ${%s}\n", varName)
			result = result[:start] + "__UNRESOLVED__" + result[end+1:]
			continue
		}

		// Get variable value
		value := engine.store.Get(varName)
		var replacement string
		if value == nil {
			fmt.Printf("[WARN] Undefined variable: %q\n", varName)
			replacement = "__UNRESOLVED__"
		} else {
			replacement = fmt.Sprintf("%v", value)
		}
		
		result = result[:start] + replacement + result[end+1:]
	}
	
	// Warn if unresolved marker remains
	if strings.Contains(result, "__UNRESOLVED__") {
		fmt.Printf("[WARN] Unresolved variables in template: %q\n", result)
	}
	
	return result
}

// SubstituteArgs performs variable substitution on arguments
func (engine *SubstitutionEngine) SubstituteArgs(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			// Check if this is a simple variable reference like "${variable_name}"
			if isSimpleVariableReference(str) {
				// For simple variable references, return the actual object instead of string representation
				varName := extractVariableName(str)
				if value := engine.store.Get(varName); value != nil {
					result[i] = value
				} else {
					result[i] = engine.Substitute(str)
				}
			} else {
				result[i] = engine.Substitute(str)
			}
		} else {
			result[i] = arg
		}
	}
	return result
}

// isSimpleVariableName checks if a string is a simple variable name (alphanumeric + underscore)
func isSimpleVariableName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

// isSimpleVariableReference checks if a string is a simple variable reference like "${variable_name}"
func isSimpleVariableReference(str string) bool {
	if !strings.HasPrefix(str, "${") || !strings.HasSuffix(str, "}") {
		return false
	}

	varName := strings.TrimSuffix(strings.TrimPrefix(str, "${"), "}")
	return isSimpleVariableName(varName)
}

// extractVariableName extracts the variable name from a simple variable reference
func extractVariableName(str string) string {
	if !isSimpleVariableReference(str) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(str, "${"), "}")
}