package runner

import (
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
	// Core operations
	SetVariable(key string, value interface{}) error
	GetVariable(key string) (interface{}, bool)
	DeleteVariable(key string) error
	ListVariables() map[string]interface{}
	ClearVariables() error
	
	// Template operations
	SubstituteTemplate(template string) string
	SubstituteArgs(args []interface{}) []interface{}
	
	// Secrets management
	LoadSecrets(secrets map[string]parser.Secret) error
	GetSecret(name string) (string, bool)
	
	// Advanced operations
	Initialize(variables map[string]interface{}) error
	GetMetadata(key string) (VariableMetadata, bool)
	Subscribe(listener VariableChangeListener)
	
	// Debug operations
	EnableDebugging(enabled bool)
	GetSubstitutionHistory() []SubstitutionEvent
}

// VariableServiceFactory creates variable services with specific configurations
type VariableServiceFactory interface {
	CreateInMemoryService() VariableService
	CreatePersistentService(path string) VariableService
	CreateEventDrivenService(listeners []VariableChangeListener) VariableService
}

// VariableMetadata contains metadata about a variable
type VariableMetadata struct {
	Key         string            `json:"key"`
	Type        string            `json:"type"`
	Source      VariableSource    `json:"source"`
	Created     time.Time         `json:"created"`
	Modified    time.Time         `json:"modified"`
	Version     int64             `json:"version"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Readonly    bool              `json:"readonly"`
}

// VariableSource indicates where a variable originated
type VariableSource string

const (
	SourceDefault    VariableSource = "default"
	SourceTestCase   VariableSource = "testcase"
	SourceStep       VariableSource = "step"
	SourceSecret     VariableSource = "secret"
	SourceGlobal     VariableSource = "global"
	SourceEnvironment VariableSource = "environment"
)

// Note: VariableChangeEvent, VariableChangeType, VariableChangeListener, and SubstitutionEvent
// are already defined in context.go