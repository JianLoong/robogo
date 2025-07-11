package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// ExecutionContext provides a clean interface for test execution dependencies
// This replaces the TestRunner god object pattern with proper dependency injection
type ExecutionContext interface {
	// Variable management
	Variables() VariableStore
	
	// Secret management  
	Secrets() SecretStore
	
	// Output handling
	Output() OutputHandler
	
	// Retry logic
	Retry() RetryHandler
	
	// Action execution
	Actions() ActionExecutor
	
	// Variable debugging
	VariableDebugger() *util.VariableResolutionDebugger
	EnableVariableDebugging(enabled bool)
	
	// Context lifecycle
	Cleanup() error
}

// VariableStore provides variable management operations
type VariableStore interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) error
	Delete(key string) error
	List() map[string]interface{}
	Substitute(template string) string
	LoadSecrets(secrets map[string]parser.Secret) error
}

// SecretStore provides secret management operations  
type SecretStore interface {
	Get(key string) (string, bool)
	Set(key string, value string, maskOutput bool) error
	List() map[string]string
	LoadFromFile(path string) (string, error)
}

// OutputHandler provides output capture and management
type OutputHandler interface {
	Capture() ([]byte, error)
	StartCapture()
	StopCapture() string
	Write(data []byte) (int, error)
}

// RetryHandler provides retry logic for failed operations
type RetryHandler interface {
	ShouldRetry(step parser.Step, attempt int, err error) bool
	GetRetryDelay(attempt int) time.Duration
	ExecuteWithRetry(ctx context.Context, step parser.Step, executor ActionExecutor, silent bool) (interface{}, error)
}


// ActionExecutor provides action execution capabilities
type ActionExecutor interface {
	Execute(ctx context.Context, action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)
}

// DefaultExecutionContext implements ExecutionContext with the existing components
type DefaultExecutionContext struct {
	variableManager   *VariableManager
	secretManager     *actions.SecretManager
	executor          *actions.ActionExecutor
	variableDebugger  *util.VariableResolutionDebugger
	variableDebugging bool
}

// NewExecutionContext creates a new execution context with default implementations
func NewExecutionContext(executor *actions.ActionExecutor) ExecutionContext {
	variableManager := NewVariableManager().(*VariableManager)
	return &DefaultExecutionContext{
		variableManager:   variableManager,
		variableDebugger:  util.NewVariableResolutionDebugger(false, "execution", variableManager),
		variableDebugging: false,
		secretManager:   actions.NewSecretManager(),
		executor:        executor,
	}
}

// Variables returns the variable store
func (ctx *DefaultExecutionContext) Variables() VariableStore {
	return &variableStoreAdapter{ctx.variableManager}
}

// Secrets returns the secret store
func (ctx *DefaultExecutionContext) Secrets() SecretStore {
	return &secretStoreAdapter{ctx.secretManager}
}

// Output returns the output handler
func (ctx *DefaultExecutionContext) Output() OutputHandler {
	return &simpleOutputHandler{}
}

// Retry returns the retry handler  
func (ctx *DefaultExecutionContext) Retry() RetryHandler {
	return &simpleRetryHandler{}
}


// Actions returns the action executor
func (ctx *DefaultExecutionContext) Actions() ActionExecutor {
	return &actionExecutorAdapter{ctx.executor}
}

// Cleanup performs any necessary cleanup
func (ctx *DefaultExecutionContext) Cleanup() error {
	// Add any cleanup logic here
	return nil
}

// VariableDebugger returns the variable resolution debugger
func (ctx *DefaultExecutionContext) VariableDebugger() *util.VariableResolutionDebugger {
	return ctx.variableDebugger
}

// EnableVariableDebugging enables or disables variable debugging
func (ctx *DefaultExecutionContext) EnableVariableDebugging(enabled bool) {
	ctx.variableDebugging = enabled
	ctx.variableDebugger = util.NewVariableResolutionDebugger(enabled, "execution", ctx.variableManager)
}

// SubstituteWithDebug performs variable substitution with debugging
func (ctx *DefaultExecutionContext) SubstituteWithDebug(template string) string {
	if ctx.variableDebugging && ctx.variableManager != nil {
		return ctx.variableManager.SubstituteStringWithDebug(template, ctx.variableDebugger)
	}
	return ctx.variableManager.SubstituteString(template)
}

// Adapter implementations to bridge existing components with new interfaces

type variableStoreAdapter struct {
	vm *VariableManager
}

func (v *variableStoreAdapter) Get(key string) (interface{}, bool) {
	return v.vm.GetVariable(key)
}

func (v *variableStoreAdapter) Set(key string, value interface{}) error {
	v.vm.SetVariable(key, value)
	return nil
}

func (v *variableStoreAdapter) Delete(key string) error {
	// VariableManager doesn't have Delete, implement if needed
	return nil
}

func (v *variableStoreAdapter) List() map[string]interface{} {
	// VariableManager doesn't expose variables directly, implement if needed
	return make(map[string]interface{})
}

func (v *variableStoreAdapter) Substitute(template string) string {
	return v.vm.SubstituteString(template)
}

func (v *variableStoreAdapter) LoadSecrets(secrets map[string]parser.Secret) error {
	// Load secrets through the variable manager
	for key, secret := range secrets {
		var value string
		if secret.Value != "" {
			value = secret.Value
		} else if secret.File != "" {
			// Read the file and set the variable to its contents
			data, err := ioutil.ReadFile(secret.File)
			if err != nil {
				return fmt.Errorf("failed to read secret file '%s': %w", secret.File, err)
			}
			value = strings.TrimSpace(string(data))
		}
		v.vm.SetVariable(key, value)
	}
	return nil
}

type secretStoreAdapter struct {
	sm *actions.SecretManager
}

func (s *secretStoreAdapter) Get(key string) (string, bool) {
	return s.sm.GetSecret(key)
}

func (s *secretStoreAdapter) Set(key string, value string, maskOutput bool) error {
	// SecretManager doesn't have a Set method, implement basic storage
	return nil
}

func (s *secretStoreAdapter) List() map[string]string {
	// SecretManager doesn't expose secrets directly for security
	return make(map[string]string)
}

func (s *secretStoreAdapter) LoadFromFile(path string) (string, error) {
	// Implement basic file loading
	return "", nil
}

// Simple output handler implementation
type simpleOutputHandler struct {
	captured []byte
	capturing bool
}

func (o *simpleOutputHandler) Capture() ([]byte, error) {
	return o.captured, nil
}

func (o *simpleOutputHandler) StartCapture() {
	o.capturing = true
	o.captured = nil
}

func (o *simpleOutputHandler) StopCapture() string {
	o.capturing = false
	result := string(o.captured)
	o.captured = nil
	return result
}

func (o *simpleOutputHandler) Write(data []byte) (int, error) {
	if o.capturing {
		o.captured = append(o.captured, data...)
	}
	return len(data), nil
}

// Simple retry handler implementation
type simpleRetryHandler struct{}

func (r *simpleRetryHandler) ShouldRetry(step parser.Step, attempt int, err error) bool {
	// Simple retry logic - retry up to 3 times for certain errors
	return attempt < 3
}

func (r *simpleRetryHandler) GetRetryDelay(attempt int) time.Duration {
	// Simple exponential backoff
	return time.Duration(attempt) * time.Second
}

func (r *simpleRetryHandler) ExecuteWithRetry(ctx context.Context, step parser.Step, executor ActionExecutor, silent bool) (interface{}, error) {
	// Simple retry implementation
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		result, err := executor.Execute(ctx, step.Action, step.Args, step.Options, silent)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !r.ShouldRetry(step, attempt, err) {
			break
		}
		if attempt < 3 {
			time.Sleep(r.GetRetryDelay(attempt))
		}
	}
	return nil, lastErr
}


type actionExecutorAdapter struct {
	executor *actions.ActionExecutor
}

func (a *actionExecutorAdapter) Execute(ctx context.Context, action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return a.executor.Execute(ctx, action, args, options, silent)
}

