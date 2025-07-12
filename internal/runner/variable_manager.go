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
	"github.com/JianLoong/robogo/internal/util"
)

// VariableManager handles variable storage, substitution, and scoping
// Implements VariableManagerInterface with both regular variables and secrets
type VariableManager struct {
	variables   map[string]interface{} // Regular variables namespace
	secrets     map[string]SecretValue // Secrets with metadata
	maskConfig  map[string]bool        // Masking configuration per secret
	mutex       sync.RWMutex
}

// SecretValue holds secret data with metadata
type SecretValue struct {
	Value      interface{}
	MaskOutput bool
	Source     string // "file" or "inline"
}

// NewVariableManager creates a new variable manager
func NewVariableManager() VariableManagerInterface {
	return &VariableManager{
		variables:  make(map[string]interface{}),
		secrets:    make(map[string]SecretValue),
		maskConfig: make(map[string]bool),
	}
}

// InitializeVariables initializes variables from test case configuration
func (vm *VariableManager) InitializeVariables(testCase *parser.TestCase) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Initialize secrets FIRST
	if testCase.Variables.Secrets != nil {
		for key, secret := range testCase.Variables.Secrets {
			var secretValue interface{}
			var source string

			// Handle secret values (inline or file-based)
			if secret.Value != "" {
				secretValue = secret.Value
				source = "inline"
			} else if secret.File != "" {
				// Read the file and set the secret to its contents
				data, err := ioutil.ReadFile(secret.File)
				if err != nil {
					panic(fmt.Sprintf("Failed to read secret file '%s': %v", secret.File, err))
				}
				secretValue = strings.TrimSpace(string(data))
				source = "file"
			} else {
				panic(fmt.Sprintf("Secret '%s' must have either 'value' or 'file' specified", key))
			}

			// Store secret with metadata
			vm.secrets[key] = SecretValue{
				Value:      secretValue,
				MaskOutput: secret.MaskOutput,
				Source:     source,
			}

			// Store masking config
			vm.maskConfig[key] = secret.MaskOutput
		}
	}

	// Initialize regular variables
	if testCase.Variables.Regular != nil {
		for key, value := range testCase.Variables.Regular {
			vm.variables[key] = value
		}
	}

	// Multiple passes: substitute variables until no more changes
	// This includes both secrets and regular variables for cross-substitution
	maxPasses := 10 // Prevent infinite loops
	for pass := 0; pass < maxPasses; pass++ {
		changed := false
		
		// Substitute in secrets
		for key, secretVal := range vm.secrets {
			substitutedValue := vm.substituteValue(secretVal.Value)
			if !vm.valuesEqual(substitutedValue, secretVal.Value) {
				vm.secrets[key] = SecretValue{
					Value:      substitutedValue,
					MaskOutput: secretVal.MaskOutput,
					Source:     secretVal.Source,
				}
				vm.variables[key] = substitutedValue
				changed = true
			}
		}
		
		// Substitute in regular variables
		for key, value := range vm.variables {
			// Skip secrets (already handled above)
			if _, isSecret := vm.secrets[key]; isSecret {
				continue
			}
			substitutedValue := vm.substituteValue(value)
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

// SetVariable sets a variable value
func (vm *VariableManager) SetVariable(name string, value interface{}) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	vm.variables[name] = value
}

// Delete removes a variable
func (vm *VariableManager) Delete(name string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	delete(vm.variables, name)
	return nil
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

// SubstituteStringWithDebug substitutes variables with debugging support
func (vm *VariableManager) SubstituteStringWithDebug(s string, debugger *util.VariableResolutionDebugger) string {
	original := s
	result := vm.substituteString(s)
	
	if debugger != nil {
		// Get all available variables for debugging
		availableVars := make(map[string]interface{})
		vm.mutex.RLock()
		for k, v := range vm.variables {
			availableVars[k] = v
		}
		vm.mutex.RUnlock()
		
		// Log the substitution if there are variables or unresolved variables
		if strings.Contains(original, "${") {
			debugger.LogVariableSubstitution(original, result, availableVars)
		}
	}
	
	return result
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

		// Handle SECRETS namespace (SECRETS.var_name)
		if strings.HasPrefix(varName, "SECRETS.") {
			secretName := varName[8:] // Remove "SECRETS." prefix
			if secret, exists := vm.secrets[secretName]; exists {
				return fmt.Sprintf("%v", secret.Value)
			}
			return match // Secret not found, return original
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
			// For complex objects, serialize to JSON instead of Go representation
			// This provides a more stable and parseable string format
			switch v := value.(type) {
			case map[string]interface{}, []interface{}:
				if jsonBytes, err := json.Marshal(v); err == nil {
					return string(jsonBytes)
				}
				// Fallback to original behavior if JSON marshal fails
				return fmt.Sprintf("%v", v)
			default:
				return fmt.Sprintf("%v", v)
			}
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


// MaskSensitiveOutput masks sensitive values in output strings
func (vm *VariableManager) MaskSensitiveOutput(output string) string {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	maskedOutput := output
	for _, secretVal := range vm.secrets {
		if secretVal.MaskOutput {
			if secretStr, ok := secretVal.Value.(string); ok && secretStr != "" {
				maskedOutput = strings.ReplaceAll(maskedOutput, secretStr, "[MASKED]")
			}
		}
	}
	return maskedOutput
}

// IsSecretMasked checks if a secret should be masked in output
func (vm *VariableManager) IsSecretMasked(secretName string) bool {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	if secretVal, exists := vm.secrets[secretName]; exists {
		return secretVal.MaskOutput
	}
	return false
}

// GetSecretInfo returns information about a secret without exposing the value
func (vm *VariableManager) GetSecretInfo(secretName string) (source string, masked bool, exists bool) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	if secretVal, ok := vm.secrets[secretName]; ok {
		return secretVal.Source, secretVal.MaskOutput, true
	}
	return "", false, false
}

// ListSecrets returns a list of secret names without their values
func (vm *VariableManager) ListSecrets() []string {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	secrets := make([]string, 0, len(vm.secrets))
	for key := range vm.secrets {
		secrets = append(secrets, key)
	}
	return secrets
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
