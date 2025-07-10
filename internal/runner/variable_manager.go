package runner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/JianLoong/robogo/internal/parser"
)

// VariableManager handles variable storage, substitution, and scoping
// Implements VariableManagerInterface
type VariableManager struct {
	variables map[string]interface{}
	mutex     sync.RWMutex
}

// NewVariableManager creates a new variable manager
func NewVariableManager() VariableManagerInterface {
	return &VariableManager{
		variables: make(map[string]interface{}),
	}
}

// InitializeVariables initializes variables from test case configuration
func (vm *VariableManager) InitializeVariables(testCase *parser.TestCase) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

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
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	vm.variables[name] = value
}

// GetVariable retrieves a variable value
func (vm *VariableManager) GetVariable(name string) (interface{}, bool) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	value, exists := vm.variables[name]
	return value, exists
}


// SubstituteVariables substitutes variables in arguments
func (vm *VariableManager) SubstituteVariables(args []interface{}) []interface{} {
	substituted := make([]interface{}, len(args))
	for i, arg := range args {
		substituted[i] = vm.substituteValue(arg)
	}
	return substituted
}

// SubstituteString substitutes variables in a string - public version of substituteString
func (vm *VariableManager) SubstituteString(s string) string {
	return vm.substituteString(s)
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
	// Match ${variable}, ${variable.property}, or ${variable[index].property} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${...}
		varName := match[2 : len(match)-1]

		// Handle array/slice access: e.g., __robogo_steps[0].error
		if strings.HasPrefix(varName, "__robogo_steps[") {
			// Parse index and property
			leftBracket := strings.Index(varName, "[")
			rightBracket := strings.Index(varName, "]")
			if leftBracket > 0 && rightBracket > leftBracket {
				indexStr := varName[leftBracket+1 : rightBracket]
				property := ""
				if dotIdx := strings.Index(varName[rightBracket:], "."); dotIdx >= 0 {
					property = varName[rightBracket+dotIdx+1:]
				}
				// Parse index
				var idx int
				_, err := fmt.Sscanf(indexStr, "%d", &idx)
				if err == nil {
					stepsVar, exists := vm.variables["__robogo_steps"]
					if exists {
						if stepsSlice, ok := stepsVar.([]interface{}); ok && idx >= 0 && idx < len(stepsSlice) {
							stepMap, ok := stepsSlice[idx].(map[string]interface{})
							if ok {
								if property == "" {
									return fmt.Sprintf("%v", stepMap)
								}
								if val, ok := stepMap[property]; ok {
									return fmt.Sprintf("%v", val)
								}
							}
						}
					}
				}
			}
			// If parsing fails, return original
			return match
		}

		// Handle dot notation for nested properties
		if strings.Contains(varName, ".") {
			value, _ := vm.resolveDotNotationInternal(varName)
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

// resolveDotNotation resolves variables with dot notation (exposed for interface)
func (vm *VariableManager) resolveDotNotation(varName string) (interface{}, bool) {
	return vm.resolveDotNotationInternal(varName)
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

// resolveDotNotationInternal resolves nested properties like "user.name"
func (vm *VariableManager) resolveDotNotationInternal(varName string) (interface{}, bool) {
	parts := strings.Split(varName, ".")
	if len(parts) < 2 {
		return nil, false
	}

	// Get the root variable
	rootVar, exists := vm.variables[parts[0]]
	if !exists {
		return nil, false
	}

	// Try to parse JSON if the root variable is a byte array or string
	var current interface{} = rootVar
	switch v := rootVar.(type) {
	case []byte:
		// Try to parse as JSON
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err == nil {
			current = parsed
		}
	case string:
		// Try to parse as JSON string
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			current = parsed
		}
	}

	// Navigate through the nested structure
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
