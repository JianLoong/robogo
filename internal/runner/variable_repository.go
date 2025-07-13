package runner

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// VariableRepository provides data persistence for variables
// Separates data storage from business logic
type VariableRepository interface {
	// Core CRUD operations
	Store(key string, value interface{}) error
	Retrieve(key string) (interface{}, bool)
	Delete(key string) error
	List() map[string]interface{}
	Clear() error
	
	// Batch operations
	StoreBatch(variables map[string]interface{}) error
	RetrieveBatch(keys []string) map[string]interface{}
	
	// Metadata and introspection
	Exists(key string) bool
	Count() int
	Keys() []string
	
	// Change tracking
	GetVersion(key string) int64
	GetLastModified(key string) time.Time
	
	// Transactions (for complex operations)
	BeginTransaction() VariableTransaction
}

// VariableTransaction provides transactional variable operations
type VariableTransaction interface {
	Store(key string, value interface{}) error
	Delete(key string) error
	Commit() error
	Rollback() error
}

// VariableService provides business logic for variable management
// Uses VariableRepository for data persistence
type VariableService interface {
	// Variable operations
	SetVariable(key string, value interface{}) error
	GetVariable(key string) (interface{}, bool)
	DeleteVariable(key string) error
	ListVariables() map[string]interface{}
	ClearVariables() error
	
	// Template substitution
	SubstituteTemplate(template string) string
	SubstituteArgs(args []interface{}) []interface{}
	
	// Initialization and configuration
	Initialize(variables map[string]interface{}) error
	LoadSecrets(secrets map[string]parser.Secret) error
	
	// Event subscription
	Subscribe(listener VariableChangeListener)
	Unsubscribe(listener VariableChangeListener)
	
	// Debugging and introspection
	EnableDebugging(enabled bool)
	GetSubstitutionHistory() []SubstitutionEvent
	GetVariableMetadata(key string) VariableMetadata
}

// VariableMetadata provides information about a variable
type VariableMetadata struct {
	Key          string                 `json:"key"`
	Type         string                 `json:"type"`
	Source       VariableSource         `json:"source"`
	Created      time.Time              `json:"created"`
	Modified     time.Time              `json:"modified"`
	Version      int64                  `json:"version"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Readonly     bool                   `json:"readonly"`
}

type VariableSource string

const (
	SourceTestCase    VariableSource = "testcase"
	SourceSecret      VariableSource = "secret"
	SourceStepResult  VariableSource = "step_result"
	SourceEnvironment VariableSource = "environment"
	SourceDefault     VariableSource = "default"
)

// DefaultVariableRepository implements VariableRepository with in-memory storage
type DefaultVariableRepository struct {
	mu        sync.RWMutex
	variables map[string]interface{}
	metadata  map[string]VariableMetadata
	versions  map[string]int64
	modified  map[string]time.Time
}

func NewDefaultVariableRepository() VariableRepository {
	return &DefaultVariableRepository{
		variables: make(map[string]interface{}),
		metadata:  make(map[string]VariableMetadata),
		versions:  make(map[string]int64),
		modified:  make(map[string]time.Time),
	}
}

func (r *DefaultVariableRepository) Store(key string, value interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	r.variables[key] = value
	r.versions[key] = r.versions[key] + 1
	r.modified[key] = now
	
	// Update metadata
	if meta, exists := r.metadata[key]; exists {
		meta.Modified = now
		meta.Version = r.versions[key]
		r.metadata[key] = meta
	} else {
		r.metadata[key] = VariableMetadata{
			Key:      key,
			Type:     fmt.Sprintf("%T", value),
			Source:   SourceDefault,
			Created:  now,
			Modified: now,
			Version:  1,
			Tags:     make(map[string]string),
			Readonly: false,
		}
	}
	
	return nil
}

func (r *DefaultVariableRepository) Retrieve(key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	value, exists := r.variables[key]
	return value, exists
}

func (r *DefaultVariableRepository) Delete(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.variables, key)
	delete(r.metadata, key)
	delete(r.versions, key)
	delete(r.modified, key)
	
	return nil
}

func (r *DefaultVariableRepository) List() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range r.variables {
		result[k] = v
	}
	return result
}

func (r *DefaultVariableRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.variables = make(map[string]interface{})
	r.metadata = make(map[string]VariableMetadata)
	r.versions = make(map[string]int64)
	r.modified = make(map[string]time.Time)
	
	return nil
}

func (r *DefaultVariableRepository) StoreBatch(variables map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	for key, value := range variables {
		r.variables[key] = value
		r.versions[key] = r.versions[key] + 1
		r.modified[key] = now
		
		if meta, exists := r.metadata[key]; exists {
			meta.Modified = now
			meta.Version = r.versions[key]
			r.metadata[key] = meta
		} else {
			r.metadata[key] = VariableMetadata{
				Key:      key,
				Type:     fmt.Sprintf("%T", value),
				Source:   SourceDefault,
				Created:  now,
				Modified: now,
				Version:  1,
				Tags:     make(map[string]string),
				Readonly: false,
			}
		}
	}
	
	return nil
}

func (r *DefaultVariableRepository) RetrieveBatch(keys []string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := r.variables[key]; exists {
			result[key] = value
		}
	}
	return result
}

func (r *DefaultVariableRepository) Exists(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.variables[key]
	return exists
}

func (r *DefaultVariableRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.variables)
}

func (r *DefaultVariableRepository) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	keys := make([]string, 0, len(r.variables))
	for key := range r.variables {
		keys = append(keys, key)
	}
	return keys
}

func (r *DefaultVariableRepository) GetVersion(key string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.versions[key]
}

func (r *DefaultVariableRepository) GetLastModified(key string) time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.modified[key]
}

func (r *DefaultVariableRepository) BeginTransaction() VariableTransaction {
	return NewVariableTransaction(r)
}

// DefaultVariableTransaction implements VariableTransaction
type DefaultVariableTransaction struct {
	repo    *DefaultVariableRepository
	changes map[string]interface{}
	deletes map[string]bool
	active  bool
}

func NewVariableTransaction(repo *DefaultVariableRepository) VariableTransaction {
	return &DefaultVariableTransaction{
		repo:    repo,
		changes: make(map[string]interface{}),
		deletes: make(map[string]bool),
		active:  true,
	}
}

func (t *DefaultVariableTransaction) Store(key string, value interface{}) error {
	if !t.active {
		return fmt.Errorf("transaction is not active")
	}
	
	t.changes[key] = value
	delete(t.deletes, key) // Remove from deletes if it was marked for deletion
	return nil
}

func (t *DefaultVariableTransaction) Delete(key string) error {
	if !t.active {
		return fmt.Errorf("transaction is not active")
	}
	
	t.deletes[key] = true
	delete(t.changes, key) // Remove from changes if it was modified
	return nil
}

func (t *DefaultVariableTransaction) Commit() error {
	if !t.active {
		return fmt.Errorf("transaction is not active")
	}
	
	t.repo.mu.Lock()
	defer t.repo.mu.Unlock()
	
	// Apply changes
	for key, value := range t.changes {
		t.repo.variables[key] = value
		t.repo.versions[key] = t.repo.versions[key] + 1
		t.repo.modified[key] = time.Now()
	}
	
	// Apply deletions
	for key := range t.deletes {
		delete(t.repo.variables, key)
		delete(t.repo.metadata, key)
		delete(t.repo.versions, key)
		delete(t.repo.modified, key)
	}
	
	t.active = false
	return nil
}

func (t *DefaultVariableTransaction) Rollback() error {
	if !t.active {
		return fmt.Errorf("transaction is not active")
	}
	
	// Simply discard all changes
	t.changes = make(map[string]interface{})
	t.deletes = make(map[string]bool)
	t.active = false
	
	return nil
}

// DefaultVariableService implements VariableService with event-driven architecture
type DefaultVariableService struct {
	mu                sync.RWMutex
	repository        VariableRepository
	listeners         []VariableChangeListener
	debuggingEnabled  bool
	substitutionHistory []SubstitutionEvent
	secretMasking     bool
}

func NewDefaultVariableService(repository VariableRepository) VariableService {
	return &DefaultVariableService{
		repository:          repository,
		listeners:           make([]VariableChangeListener, 0),
		debuggingEnabled:    false,
		substitutionHistory: make([]SubstitutionEvent, 0),
		secretMasking:       true,
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
	oldValue, existed := s.repository.Retrieve(key)
	if !existed {
		return fmt.Errorf("variable '%s' does not exist", key)
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
		NewValue:  nil,
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
	variables := s.repository.List()
	result := s.performSubstitution(template, variables)
	
	if s.debuggingEnabled {
		s.mu.Lock()
		s.substitutionHistory = append(s.substitutionHistory, SubstitutionEvent{
			Timestamp: time.Now(),
			Original:  template,
			Resolved:  result,
			Variables: variables,
			Context:   "variable_service",
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

func (s *DefaultVariableService) Initialize(variables map[string]interface{}) error {
	return s.repository.StoreBatch(variables)
}

func (s *DefaultVariableService) LoadSecrets(secrets map[string]parser.Secret) error {
	for key, secret := range secrets {
		var value string
		if secret.Value != "" {
			value = secret.Value
		} else if secret.File != "" {
			// Load from file (simplified - in real implementation would use SecretRepository)
			value = fmt.Sprintf("loaded_from_%s", secret.File)
		}
		
		if err := s.SetVariable(key, value); err != nil {
			return fmt.Errorf("failed to load secret %s: %w", key, err)
		}
	}
	return nil
}

func (s *DefaultVariableService) Subscribe(listener VariableChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.listeners = append(s.listeners, listener)
}

func (s *DefaultVariableService) Unsubscribe(listener VariableChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i, l := range s.listeners {
		// Simple pointer comparison - in real implementation might need better equality check
		if &l == &listener {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			break
		}
	}
}

func (s *DefaultVariableService) EnableDebugging(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.debuggingEnabled = enabled
}

func (s *DefaultVariableService) GetSubstitutionHistory() []SubstitutionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]SubstitutionEvent, len(s.substitutionHistory))
	copy(result, s.substitutionHistory)
	return result
}

func (s *DefaultVariableService) GetVariableMetadata(key string) VariableMetadata {
	// For now, return basic metadata - in real implementation would get from repository
	version := s.repository.GetVersion(key)
	modified := s.repository.GetLastModified(key)
	
	return VariableMetadata{
		Key:      key,
		Version:  version,
		Modified: modified,
		Source:   SourceDefault,
		Readonly: false,
		Tags:     make(map[string]string),
	}
}

func (s *DefaultVariableService) notifyListeners(event VariableChangeEvent) {
	s.mu.RLock()
	listeners := make([]VariableChangeListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.RUnlock()
	
	// Notify listeners asynchronously
	for _, listener := range listeners {
		go listener.OnVariableChanged(event)
	}
}

func (s *DefaultVariableService) performSubstitution(template string, variables map[string]interface{}) string {
	result := template
	
	// Enhanced variable substitution with dot notation support: ${variable_name} or ${object.property}
	for strings.Contains(result, "${") {
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
		replacement := ""
		
		// Handle dot notation for nested properties
		if strings.Contains(varName, ".") {
			if value := s.resolveDotNotation(varName, variables); value != nil {
				replacement = fmt.Sprintf("%v", value)
			}
		} else {
			// Simple variable lookup
			if value, exists := variables[varName]; exists {
				replacement = fmt.Sprintf("%v", value)
			}
		}
		
		result = result[:start] + replacement + result[end+1:]
	}
	
	return result
}

// resolveDotNotation resolves nested properties like "user.name" or "http_response.status_code"
func (s *DefaultVariableService) resolveDotNotation(varName string, variables map[string]interface{}) interface{} {
	parts := strings.Split(varName, ".")
	if len(parts) < 2 {
		return nil
	}

	// Get the root variable
	rootVar, exists := variables[parts[0]]
	if !exists {
		return nil
	}

	// Try to parse JSON if the root variable is a string or byte array
	current := rootVar
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
				return nil
			}
		case map[interface{}]interface{}:
			if next, ok := v[parts[i]]; ok {
				current = next
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// VariableServiceFactory creates variable services with proper dependencies
type VariableServiceFactory struct {
	defaultRepository VariableRepository
}

func NewVariableServiceFactory() *VariableServiceFactory {
	return &VariableServiceFactory{
		defaultRepository: NewDefaultVariableRepository(),
	}
}

func (f *VariableServiceFactory) CreateVariableService() VariableService {
	return NewDefaultVariableService(f.defaultRepository)
}

func (f *VariableServiceFactory) CreateVariableServiceWithRepository(repo VariableRepository) VariableService {
	return NewDefaultVariableService(repo)
}

func (f *VariableServiceFactory) CreateVariableRepository() VariableRepository {
	return NewDefaultVariableRepository()
}