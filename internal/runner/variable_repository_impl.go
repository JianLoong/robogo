package runner

import (
	"fmt"
	"sync"
	"time"
)

// DefaultVariableRepository provides an in-memory implementation of VariableRepository
type DefaultVariableRepository struct {
	mu        sync.RWMutex
	variables map[string]interface{}
	metadata  map[string]VariableMetadata
	versions  map[string]int64
	modified  map[string]time.Time
}

// NewDefaultVariableRepository creates a new in-memory variable repository
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
	for k := range r.variables {
		keys = append(keys, k)
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
	return &DefaultVariableTransaction{
		repository: r,
		changes:    make(map[string]interface{}),
		deletions:  make([]string, 0),
	}
}

// DefaultVariableTransaction provides transactional operations
type DefaultVariableTransaction struct {
	repository *DefaultVariableRepository
	changes    map[string]interface{}
	deletions  []string
	committed  bool
}

func (t *DefaultVariableTransaction) Store(key string, value interface{}) error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	t.changes[key] = value
	return nil
}

func (t *DefaultVariableTransaction) Delete(key string) error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	t.deletions = append(t.deletions, key)
	return nil
}

func (t *DefaultVariableTransaction) Commit() error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	
	// Apply changes
	for key, value := range t.changes {
		if err := t.repository.Store(key, value); err != nil {
			return err
		}
	}
	
	// Apply deletions
	for _, key := range t.deletions {
		if err := t.repository.Delete(key); err != nil {
			return err
		}
	}
	
	t.committed = true
	return nil
}

func (t *DefaultVariableTransaction) Rollback() error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	
	// Clear pending changes
	t.changes = make(map[string]interface{})
	t.deletions = make([]string, 0)
	t.committed = true
	
	return nil
}