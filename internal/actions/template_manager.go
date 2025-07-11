package actions

import (
	"sync"
)

// TemplateManager manages template contexts with proper lifecycle
type TemplateManager struct {
	templateContext map[string]string
	mutex           sync.RWMutex
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templateContext: make(map[string]string),
	}
}

// SetTemplateContext sets the available templates for the current execution context
func (tm *TemplateManager) SetTemplateContext(templates map[string]string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	// Create a new map to avoid sharing references
	tm.templateContext = make(map[string]string, len(templates))
	for key, value := range templates {
		tm.templateContext[key] = value
	}
}

// GetTemplateContext returns a copy of the current template context
func (tm *TemplateManager) GetTemplateContext() map[string]string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	result := make(map[string]string, len(tm.templateContext))
	for key, value := range tm.templateContext {
		result[key] = value
	}
	return result
}

// GetTemplate retrieves a specific template by name
func (tm *TemplateManager) GetTemplate(name string) (string, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	template, exists := tm.templateContext[name]
	return template, exists
}

// HasTemplate checks if a template exists
func (tm *TemplateManager) HasTemplate(name string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	_, exists := tm.templateContext[name]
	return exists
}

// ListTemplates returns the names of all available templates
func (tm *TemplateManager) ListTemplates() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	names := make([]string, 0, len(tm.templateContext))
	for name := range tm.templateContext {
		names = append(names, name)
	}
	return names
}

// ClearTemplates removes all templates from the context
func (tm *TemplateManager) ClearTemplates() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.templateContext = make(map[string]string)
}

// AddTemplate adds a single template to the context
func (tm *TemplateManager) AddTemplate(name, content string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.templateContext[name] = content
}

// RemoveTemplate removes a specific template from the context
func (tm *TemplateManager) RemoveTemplate(name string) bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	_, exists := tm.templateContext[name]
	if exists {
		delete(tm.templateContext, name)
	}
	return exists
}