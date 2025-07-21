package common

import (
	"encoding/json"
	"fmt"
	"os"
)

// VariableManager provides a clean interface for variable operations
type VariableManager struct {
	storage      VariableStorage
	substitution *TemplateSubstitution
}

// NewVariableManager creates a new variable manager
func NewVariableManager() *VariableManager {
	storage := NewSimpleVariableStorage()
	substitution := NewTemplateSubstitution(storage)
	
	return &VariableManager{
		storage:      storage,
		substitution: substitution,
	}
}

// Set stores a variable value without automatic parsing
func (vm *VariableManager) Set(key string, value any) {
	vm.storage.Set(key, value)
}

// SetFromJSON explicitly parses a JSON string and stores the parsed value
func (vm *VariableManager) SetFromJSON(key, jsonString string) error {
	var parsed any
	if err := json.Unmarshal([]byte(jsonString), &parsed); err != nil {
		return fmt.Errorf("failed to parse JSON for variable '%s': %w", key, err)
	}
	vm.storage.Set(key, parsed)
	return nil
}

// Get retrieves a variable value
func (vm *VariableManager) Get(key string) any {
	return vm.storage.Get(key)
}

// Has checks if a variable exists
func (vm *VariableManager) Has(key string) bool {
	return vm.storage.Has(key)
}

// Delete removes a variable
func (vm *VariableManager) Delete(key string) {
	vm.storage.Delete(key)
}

// Substitute performs template substitution and returns detailed results
func (vm *VariableManager) Substitute(template string) SubstitutionResult {
	result := vm.substitution.Substitute(template)
	// Clean unresolved markers for final display
	result.Result = vm.substitution.CleanUnresolvedMarkers(result.Result)
	return result
}

// SubstituteSimple performs template substitution and returns just the result string
func (vm *VariableManager) SubstituteSimple(template string) string {
	return vm.Substitute(template).Result
}

// LoadFromMap bulk loads variables from a map
func (vm *VariableManager) LoadFromMap(variables map[string]any) {
	for key, value := range variables {
		vm.Set(key, value)
	}
}

// LoadFromMapWithEnvSubstitution loads variables with environment variable substitution
func (vm *VariableManager) LoadFromMapWithEnvSubstitution(variables map[string]any) {
	for key, value := range variables {
		if strValue, ok := value.(string); ok {
			// Substitute environment variables in string values
			substituted := vm.SubstituteSimple(strValue)
			vm.Set(key, substituted)
		} else {
			vm.Set(key, value)
		}
	}
}

// GetSnapshot returns a copy of all variables
func (vm *VariableManager) GetSnapshot() map[string]any {
	return vm.storage.GetSnapshot()
}

// Clone creates a copy of the variable manager
func (vm *VariableManager) Clone() *VariableManager {
	newManager := NewVariableManager()
	snapshot := vm.GetSnapshot()
	newManager.LoadFromMap(snapshot)
	return newManager
}

// Clear removes all variables
func (vm *VariableManager) Clear() {
	vm.storage.Clear()
}

// LoadFromEnvironment loads specific environment variables into storage
func (vm *VariableManager) LoadFromEnvironment(envVars []string) {
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			vm.Set("ENV_"+envVar, value)
		}
	}
}

// GetSubstitutionStats returns statistics about unresolved variables
func (vm *VariableManager) GetSubstitutionStats(template string) SubstitutionResult {
	return vm.substitution.Substitute(template)
}