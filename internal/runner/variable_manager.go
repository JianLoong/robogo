package runner

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"github.com/JianLoong/robogo/internal/parser"
)

// VariableManager handles variable storage, substitution, and scoping
type VariableManager struct {
	variables map[string]interface{}
}

// NewVariableManager creates a new variable manager
func NewVariableManager() *VariableManager {
	return &VariableManager{
		variables: make(map[string]interface{}),
	}
}

// InitializeVariables initializes variables from test case configuration
func (vm *VariableManager) InitializeVariables(testCase *parser.TestCase) {
	// Initialize secret variables FIRST
	if testCase.Variables.Secrets != nil {
		for key, secret := range testCase.Variables.Secrets {
			// Handle secret values (inline or file-based)
			if secret.Value != "" {
				// Substitute variables in secret values
				substitutedValue := vm.substituteString(secret.Value)
				vm.variables[key] = substitutedValue
			} else if secret.File != "" {
				// Read the file and set the variable to its contents
				data, err := ioutil.ReadFile(secret.File)
				if err != nil {
					panic(fmt.Sprintf("Failed to read secret file '%s': %v", secret.File, err))
				}
				fileContent := strings.TrimSpace(string(data))
				// Substitute variables in file content
				substitutedContent := vm.substituteString(fileContent)
				vm.variables[key] = substitutedContent
			}
		}
	}

	// Initialize regular variables with support for dynamic construction
	if testCase.Variables.Regular != nil {
		// First pass: set all variables as-is
		for key, value := range testCase.Variables.Regular {
			vm.variables[key] = value
		}

		// Multiple passes: substitute variables until no more changes
		maxPasses := 10 // Prevent infinite loops
		for pass := 0; pass < maxPasses; pass++ {
			changed := false
			for key, value := range vm.variables {
				substitutedValue := vm.substituteValue(value)
				// Safe comparison that handles map types
				if !vm.valuesEqual(substitutedValue, value) {
					vm.variables[key] = substitutedValue
					changed = true
				}
			}
			// If no changes in this pass, we're done
			if !changed {
				break
			}
		}
	}
}

// SetVariable sets a variable value
func (vm *VariableManager) SetVariable(name string, value interface{}) {
	vm.variables[name] = value
}

// GetVariable retrieves a variable value
func (vm *VariableManager) GetVariable(name string) (interface{}, bool) {
	value, exists := vm.variables[name]
	return value, exists
}

// ListVariables returns all variable names
func (vm *VariableManager) ListVariables() []string {
	var names []string
	for name := range vm.variables {
		names = append(names, name)
	}
	return names
}

// SubstituteVariables substitutes variables in arguments
func (vm *VariableManager) SubstituteVariables(args []interface{}) []interface{} {
	substituted := make([]interface{}, len(args))
	for i, arg := range args {
		substituted[i] = vm.substituteValue(arg)
	}
	return substituted
}

// substituteValue recursively substitutes variables in a value
func (vm *VariableManager) substituteValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return vm.substituteString(v)
	case map[string]interface{}:
		return vm.substituteMap(v)
	case []interface{}:
		substituted := make([]interface{}, len(v))
		for i, item := range v {
			substituted[i] = vm.substituteValue(item)
		}
		return substituted
	default:
		return value
	}
}

// substituteString substitutes variables in a string using ${variable} syntax
func (vm *VariableManager) substituteString(s string) string {
	// Match ${variable} or ${variable.property} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${...}
		varName := match[2 : len(match)-1]

		// Handle dot notation for nested properties
		if strings.Contains(varName, ".") {
			value, _ := vm.resolveDotNotation(varName)
			if value != nil {
				return fmt.Sprintf("%v", value)
			}
		}

		// Simple variable lookup
		if value, exists := vm.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}

		// Return original if variable not found
		return match
	})
}

// substituteStringForDisplay substitutes variables for display purposes (without changing the original)
func (vm *VariableManager) substituteStringForDisplay(s string) string {
	// This is the same as substituteString but for display purposes
	return vm.substituteString(s)
}

// substituteMap substitutes variables in a map
func (vm *VariableManager) substituteMap(m map[string]interface{}) map[string]interface{} {
	substituted := make(map[string]interface{})
	for key, value := range m {
		// Substitute in both key and value
		substitutedKey := vm.substituteString(key)
		substituted[substitutedKey] = vm.substituteValue(value)
	}
	return substituted
}

// resolveDotNotation resolves nested properties like "user.name"
func (vm *VariableManager) resolveDotNotation(varName string) (interface{}, bool) {
	parts := strings.Split(varName, ".")
	if len(parts) < 2 {
		return nil, false
	}

	// Get the root variable
	rootVar, exists := vm.variables[parts[0]]
	if !exists {
		return nil, false
	}

	// Navigate through the nested structure
	current := rootVar
	for i := 1; i < len(parts); i++ {
		switch v := current.(type) {
		case map[string]interface{}:
			if next, ok := v[parts[i]]; ok {
				current = next
			} else {
				return nil, false
			}
		case map[interface{}]interface{}:
			if next, ok := v[parts[i]]; ok {
				current = next
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// GetVariableCount returns the number of variables
func (vm *VariableManager) GetVariableCount() int {
	return len(vm.variables)
}

// ClearVariables clears all variables
func (vm *VariableManager) ClearVariables() {
	vm.variables = make(map[string]interface{})
}

// GetVariableNames returns all variable names for debugging
func (vm *VariableManager) GetVariableNames() []string {
	var names []string
	for name := range vm.variables {
		names = append(names, name)
	}
	return names
}

// valuesEqual safely compares two values, handling map types that are not directly comparable
func (vm *VariableManager) valuesEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	if reflect.TypeOf(a) != nil && reflect.TypeOf(a).Kind() == reflect.Map {
		return false
	}
	if reflect.TypeOf(b) != nil && reflect.TypeOf(b).Kind() == reflect.Map {
		return false
	}
	return reflect.DeepEqual(a, b)
}
