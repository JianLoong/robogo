package runner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// DefaultVariableService implements VariableService with event-driven architecture
type DefaultVariableService struct {
	mu                  sync.RWMutex
	repository          VariableRepository
	listeners           []VariableChangeListener
	debuggingEnabled    bool
	substitutionHistory []SubstitutionEvent
	secretMasking       bool
	secrets             map[string]string
}

// NewDefaultVariableService creates a new variable service with repository
func NewDefaultVariableService(repository VariableRepository) VariableService {
	return &DefaultVariableService{
		repository:          repository,
		listeners:           make([]VariableChangeListener, 0),
		debuggingEnabled:    false,
		substitutionHistory: make([]SubstitutionEvent, 0),
		secretMasking:       true,
		secrets:             make(map[string]string),
	}
}

func (s *DefaultVariableService) SetVariable(key string, value interface{}) error {
	oldValue, _ := s.repository.Retrieve(key)
	
	err := s.repository.Store(key, value)
	if err != nil {
		return err
	}
	
	// Notify listeners
	event := VariableChangeEvent{
		Type:      VariableSet,
		Variable:  key,
		OldValue:  oldValue,
		NewValue:  value,
		Timestamp: time.Now(),
		Context:   "variable_service",
	}
	
	s.notifyListeners(event)
	
	return nil
}

func (s *DefaultVariableService) GetVariable(key string) (interface{}, bool) {
	return s.repository.Retrieve(key)
}

func (s *DefaultVariableService) DeleteVariable(key string) error {
	oldValue, exists := s.repository.Retrieve(key)
	if !exists {
		return nil // Already deleted
	}
	
	err := s.repository.Delete(key)
	if err != nil {
		return err
	}
	
	// Notify listeners
	event := VariableChangeEvent{
		Type:      VariableDeleted,
		Variable:  key,
		OldValue:  oldValue,
		Timestamp: time.Now(),
		Context:   "variable_service",
	}
	
	s.notifyListeners(event)
	
	return nil
}

func (s *DefaultVariableService) ListVariables() map[string]interface{} {
	return s.repository.List()
}

func (s *DefaultVariableService) ClearVariables() error {
	return s.repository.Clear()
}

func (s *DefaultVariableService) SubstituteTemplate(template string) string {
	originalTemplate := template
	variables := s.repository.List()
	
	// Process ${variable} patterns
	result := s.processTemplate(template, variables)
	
	// Log substitution if debugging is enabled
	if s.debuggingEnabled {
		s.mu.Lock()
		s.substitutionHistory = append(s.substitutionHistory, SubstitutionEvent{
			Timestamp: time.Now(),
			Original:  originalTemplate,
			Resolved:  result,
			Variables: variables,
			Context:   "template_substitution",
		})
		s.mu.Unlock()
	}
	
	return result
}

func (s *DefaultVariableService) SubstituteArgs(args []interface{}) []interface{} {
	result := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			result[i] = s.SubstituteTemplate(str)
		} else {
			result[i] = arg
		}
	}
	return result
}

func (s *DefaultVariableService) LoadSecrets(secrets map[string]parser.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for name, secret := range secrets {
		var value string
		var err error
		
		if secret.Value != "" {
			// Use inline value
			value = secret.Value
		} else if secret.File != "" {
			// Load from file
			value, err = s.loadSecretFromFile(secret.File)
			if err != nil {
				return fmt.Errorf("failed to load secret %s from file %s: %w", name, secret.File, err)
			}
		} else {
			return fmt.Errorf("secret %s has neither value nor file specified", name)
		}
		
		s.secrets[name] = value
		
		// Store as regular variable for substitution
		if err := s.repository.Store(name, value); err != nil {
			return fmt.Errorf("failed to store secret %s: %w", name, err)
		}
	}
	
	return nil
}

func (s *DefaultVariableService) GetSecret(name string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	value, exists := s.secrets[name]
	return value, exists
}

func (s *DefaultVariableService) Initialize(variables map[string]interface{}) error {
	return s.repository.StoreBatch(variables)
}

func (s *DefaultVariableService) GetMetadata(key string) (VariableMetadata, bool) {
	// TODO: Implement metadata retrieval from repository
	return VariableMetadata{}, false
}

func (s *DefaultVariableService) Subscribe(listener VariableChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.listeners = append(s.listeners, listener)
}

func (s *DefaultVariableService) EnableDebugging(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.debuggingEnabled = enabled
}

func (s *DefaultVariableService) GetSubstitutionHistory() []SubstitutionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to prevent external modification
	history := make([]SubstitutionEvent, len(s.substitutionHistory))
	copy(history, s.substitutionHistory)
	return history
}

// processTemplate performs the actual variable substitution
func (s *DefaultVariableService) processTemplate(template string, variables map[string]interface{}) string {
	result := template
	
	// Simple implementation - replace ${variable} patterns
	for {
		startIdx := strings.Index(result, "${")
		if startIdx == -1 {
			break
		}
		
		endIdx := strings.Index(result[startIdx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx
		
		varName := result[startIdx+2 : endIdx]
		if varName == "" {
			result = result[:startIdx] + result[endIdx+1:]
			continue
		}
		
		// Handle dot notation for nested properties
		value := s.resolveDotNotation(varName, variables)
		
		replacement := fmt.Sprintf("%v", value)
		result = result[:startIdx] + replacement + result[endIdx+1:]
	}
	
	return result
}

// resolveDotNotation handles variable.property access patterns
func (s *DefaultVariableService) resolveDotNotation(varName string, variables map[string]interface{}) interface{} {
	if !strings.Contains(varName, ".") {
		// Simple variable lookup
		if value, exists := variables[varName]; exists {
			return value
		}
		return fmt.Sprintf("${%s}", varName) // Return unresolved if not found
	}
	
	// Handle dot notation
	parts := strings.Split(varName, ".")
	baseVar := parts[0]
	
	baseValue, exists := variables[baseVar]
	if !exists {
		return fmt.Sprintf("${%s}", varName) // Return unresolved if base variable not found
	}
	
	// Navigate through the dot notation path
	current := baseValue
	for _, part := range parts[1:] {
		// Try to access as map
		if m, ok := current.(map[string]interface{}); ok {
			if value, exists := m[part]; exists {
				current = value
				continue
			}
		}
		
		// Try to parse as JSON and access
		if str, ok := current.(string); ok {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(str), &jsonData); err == nil {
				if m, ok := jsonData.(map[string]interface{}); ok {
					if value, exists := m[part]; exists {
						current = value
						continue
					}
				}
			}
		}
		
		// Property not found
		return fmt.Sprintf("${%s}", varName)
	}
	
	return current
}

// loadSecretFromFile loads a secret value from a file
func (s *DefaultVariableService) loadSecretFromFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// notifyListeners sends events to all registered listeners
func (s *DefaultVariableService) notifyListeners(event VariableChangeEvent) {
	s.mu.RLock()
	listeners := make([]VariableChangeListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.RUnlock()
	
	for _, listener := range listeners {
		go listener.OnVariableChanged(event)
	}
}

// DefaultVariableServiceFactory creates variable services
type DefaultVariableServiceFactory struct{}

func NewVariableServiceFactory() VariableServiceFactory {
	return &DefaultVariableServiceFactory{}
}

func (f *DefaultVariableServiceFactory) CreateInMemoryService() VariableService {
	repository := NewDefaultVariableRepository()
	return NewDefaultVariableService(repository)
}

func (f *DefaultVariableServiceFactory) CreatePersistentService(path string) VariableService {
	// TODO: Implement persistent repository
	repository := NewDefaultVariableRepository()
	return NewDefaultVariableService(repository)
}

func (f *DefaultVariableServiceFactory) CreateEventDrivenService(listeners []VariableChangeListener) VariableService {
	repository := NewDefaultVariableRepository()
	service := NewDefaultVariableService(repository).(*DefaultVariableService)
	
	for _, listener := range listeners {
		service.Subscribe(listener)
	}
	
	return service
}