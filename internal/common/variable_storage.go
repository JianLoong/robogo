package common

// VariableStorage defines the interface for variable storage operations
type VariableStorage interface {
	// Set stores a variable value
	Set(key string, value any)
	
	// Get retrieves a variable value
	Get(key string) any
	
	// Has checks if a variable exists
	Has(key string) bool
	
	// Delete removes a variable
	Delete(key string)
	
	// GetSnapshot returns a copy of all variables
	GetSnapshot() map[string]any
	
	// Clear removes all variables
	Clear()
}

// SimpleVariableStorage provides basic in-memory variable storage
type SimpleVariableStorage struct {
	data map[string]any
}

// NewSimpleVariableStorage creates a new simple variable storage
func NewSimpleVariableStorage() *SimpleVariableStorage {
	return &SimpleVariableStorage{
		data: make(map[string]any),
	}
}

// Set stores a variable value without any automatic parsing
func (s *SimpleVariableStorage) Set(key string, value any) {
	s.data[key] = value
}

// Get retrieves a variable value
func (s *SimpleVariableStorage) Get(key string) any {
	return s.data[key]
}

// Has checks if a variable exists
func (s *SimpleVariableStorage) Has(key string) bool {
	_, exists := s.data[key]
	return exists
}

// Delete removes a variable
func (s *SimpleVariableStorage) Delete(key string) {
	delete(s.data, key)
}

// GetSnapshot returns a copy of all variables
func (s *SimpleVariableStorage) GetSnapshot() map[string]any {
	snapshot := make(map[string]any, len(s.data))
	for key, value := range s.data {
		snapshot[key] = value
	}
	return snapshot
}

// Clear removes all variables
func (s *SimpleVariableStorage) Clear() {
	s.data = make(map[string]any)
}