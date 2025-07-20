package common

import (
	"encoding/json"
)

// VariableStore handles core variable storage and management
type VariableStore struct {
	data map[string]any
}

// NewVariableStore creates a new variable store
func NewVariableStore() *VariableStore {
	return &VariableStore{
		data: make(map[string]any),
	}
}

// Set stores a variable with automatic JSON parsing for string values
func (store *VariableStore) Set(key string, value any) {
	// If value is a JSON string, try to parse it
	if str, ok := value.(string); ok {
		var parsed any
		if err := json.Unmarshal([]byte(str), &parsed); err == nil {
			store.data[key] = parsed
			return
		}
	}
	store.data[key] = value
}

// Get retrieves a variable by key, returns nil if not found
func (store *VariableStore) Get(key string) any {
	if val, exists := store.data[key]; exists {
		return val
	}
	return nil
}

// Load bulk loads variables from a map
func (store *VariableStore) Load(vars map[string]any) {
	for k, val := range vars {
		store.data[k] = val
	}
}

// GetSnapshot returns a copy of all current variables for context enrichment
func (store *VariableStore) GetSnapshot() map[string]interface{} {
	snapshot := make(map[string]interface{})
	for k, val := range store.data {
		snapshot[k] = val
	}
	return snapshot
}

// GetAll returns a copy of the internal data map
func (store *VariableStore) GetAll() map[string]any {
	result := make(map[string]any)
	for k, v := range store.data {
		result[k] = v
	}
	return result
}
