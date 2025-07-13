package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// TestExecutionContext provides a composition-based interface for test execution
type TestExecutionContext interface {
	Variables() VariableContext
	Secrets() SecretContext
	Output() OutputContext
	Actions() ActionContext
	Lifecycle() LifecycleContext
	EnableVariableDebugging(enabled bool)
}

// VariableContext handles all variable-related operations
type VariableContext interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) error
	Delete(key string) error
	List() map[string]interface{}
	Substitute(template string) string
	SubstituteArgs(args []interface{}) []interface{}
	Debug() VariableDebugger
	EnableDebugging(enabled bool)
	GetSubstitutionHistory() []SubstitutionEvent
	Initialize(variables map[string]interface{}) error
	LoadSecrets(secrets map[string]parser.Secret) error
	Clear() error
}

// SecretContext handles secure credential management
type SecretContext interface {
	GetSecret(name string) (string, bool)
	LoadSecret(name string, config parser.Secret) error
	LoadFromFile(path string) (string, error)
	MaskSensitiveOutput(text string) string
	IsSecretMasked(secretName string) bool
	ListSecrets() []string
	GetSecretInfo(secretName string) (source string, masked bool, exists bool)
}

// OutputContext handles output capture and management
type OutputContext interface {
	StartCapture()
	StopCapture() string
	Write(data []byte) (int, error)
	Capture() ([]byte, error)
	StartWithRealTimeDisplay(realTime bool)
	SetRealTimeEnabled(enabled bool)
	AddFilter(filter OutputFilter)
	RemoveFilter(name string)
}

// ActionContext handles action execution and metadata
type ActionContext interface {
	Execute(ctx context.Context, action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)
	GetMetadata(action string) (actions.ActionMetadata, bool)
	ListActions() []string
	ListActionsByCategory(category string) []string
	SearchActions(prefix string) []string
	ValidateAction(action string, args []interface{}) error
	GetActionCompletions(partial string) []string
}

// LifecycleContext handles context lifecycle and resource management
type LifecycleContext interface {
	Initialize() error
	Cleanup() error
	Health() error
	RegisterCleanupHandler(handler CleanupHandler)
	GetResourceUsage() ResourceUsage
}

// Supporting types

type VariableDebugger interface {
	EnableDebugging(enabled bool)
	LogSubstitution(original, resolved string, variables map[string]interface{})
	GetDebugHistory() []SubstitutionEvent
	ClearHistory()
}

type SubstitutionEvent struct {
	Timestamp time.Time
	Original  string
	Resolved  string
	Variables map[string]interface{}
	Context   string
}

type OutputFilter interface {
	Name() string
	Filter(data []byte) []byte
	ShouldApply(context string) bool
}

type CleanupHandler interface {
	Name() string
	Cleanup() error
	Priority() int // Higher priority cleanup happens first
}

type ResourceUsage struct {
	MemoryUsage   int64
	OpenFiles     int
	ActiveOutputs int
	VariableCount int
	SecretCount   int
}

type VariableChangeEvent struct {
	Type      VariableChangeType
	Variable  string
	OldValue  interface{}
	NewValue  interface{}
	Timestamp time.Time
	Context   string
}

type VariableChangeType int

const (
	VariableSet VariableChangeType = iota
	VariableDeleted
	VariableSubstituted
)

type VariableChangeListener interface {
	OnVariableChanged(event VariableChangeEvent)
}

// Implementation

// DefaultTestExecutionContext implements TestExecutionContext using composition
type DefaultTestExecutionContext struct {
	variables VariableContext
	secrets   SecretContext
	output    OutputContext
	actions   ActionContext
	lifecycle LifecycleContext
}

func NewTestExecutionContext(executor *actions.ActionExecutor) TestExecutionContext {
	return &DefaultTestExecutionContext{
		variables: NewDefaultVariableContext(),
		secrets:   NewDefaultSecretContext(),
		output:    NewDefaultOutputContext(),
		actions:   NewDefaultActionContext(executor),
		lifecycle: NewDefaultLifecycleContext(),
	}
}

func (ctx *DefaultTestExecutionContext) Variables() VariableContext {
	return ctx.variables
}

func (ctx *DefaultTestExecutionContext) Secrets() SecretContext {
	return ctx.secrets
}

func (ctx *DefaultTestExecutionContext) Output() OutputContext {
	return ctx.output
}

func (ctx *DefaultTestExecutionContext) Actions() ActionContext {
	return ctx.actions
}

func (ctx *DefaultTestExecutionContext) Lifecycle() LifecycleContext {
	return ctx.lifecycle
}

func (ctx *DefaultTestExecutionContext) EnableVariableDebugging(enabled bool) {
	ctx.variables.EnableDebugging(enabled)
}

// DefaultVariableContext implements VariableContext using VariableService
type DefaultVariableContext struct {
	service  VariableService
	debugger VariableDebugger
}

func NewDefaultVariableContext() VariableContext {
	factory := NewVariableServiceFactory()
	service := factory.CreateVariableService()

	return &DefaultVariableContext{
		service:  service,
		debugger: NewDefaultVariableDebugger(),
	}
}

func (vc *DefaultVariableContext) Get(key string) (interface{}, bool) {
	return vc.service.GetVariable(key)
}

func (vc *DefaultVariableContext) Set(key string, value interface{}) error {
	return vc.service.SetVariable(key, value)
}

func (vc *DefaultVariableContext) Delete(key string) error {
	return vc.service.DeleteVariable(key)
}

func (vc *DefaultVariableContext) List() map[string]interface{} {
	return vc.service.ListVariables()
}

func (vc *DefaultVariableContext) Substitute(template string) string {
	return vc.service.SubstituteTemplate(template)
}

func (vc *DefaultVariableContext) SubstituteArgs(args []interface{}) []interface{} {
	return vc.service.SubstituteArgs(args)
}

func (vc *DefaultVariableContext) Debug() VariableDebugger {
	return vc.debugger
}

func (vc *DefaultVariableContext) EnableDebugging(enabled bool) {
	vc.service.EnableDebugging(enabled)
	vc.debugger.EnableDebugging(enabled)
}

func (vc *DefaultVariableContext) GetSubstitutionHistory() []SubstitutionEvent {
	return vc.service.GetSubstitutionHistory()
}

func (vc *DefaultVariableContext) Initialize(variables map[string]interface{}) error {
	return vc.service.Initialize(variables)
}

func (vc *DefaultVariableContext) LoadSecrets(secrets map[string]parser.Secret) error {
	return vc.service.LoadSecrets(secrets)
}

func (vc *DefaultVariableContext) Clear() error {
	return vc.service.ClearVariables()
}

// DefaultSecretContext implements SecretContext
type DefaultSecretContext struct {
	mu            sync.RWMutex
	secrets       map[string]string
	maskedSecrets map[string]bool
	secretSources map[string]string
}

func NewDefaultSecretContext() SecretContext {
	return &DefaultSecretContext{
		secrets:       make(map[string]string),
		maskedSecrets: make(map[string]bool),
		secretSources: make(map[string]string),
	}
}

func (sc *DefaultSecretContext) GetSecret(name string) (string, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	value, exists := sc.secrets[name]
	return value, exists
}

func (sc *DefaultSecretContext) LoadSecret(name string, config parser.Secret) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var value string
	var source string

	if config.Value != "" {
		value = config.Value
		source = "inline"
	} else if config.File != "" {
		data, err := ioutil.ReadFile(config.File)
		if err != nil {
			return fmt.Errorf("failed to read secret file '%s': %w", config.File, err)
		}
		value = strings.TrimSpace(string(data))
		source = config.File
	} else {
		return fmt.Errorf("secret '%s' has no value or file specified", name)
	}

	sc.secrets[name] = value
	sc.maskedSecrets[name] = config.MaskOutput
	sc.secretSources[name] = source

	return nil
}

func (sc *DefaultSecretContext) LoadFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func (sc *DefaultSecretContext) MaskSensitiveOutput(text string) string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := text
	for secretName, secretValue := range sc.secrets {
		if masked, exists := sc.maskedSecrets[secretName]; exists && masked {
			if secretValue != "" {
				result = strings.ReplaceAll(result, secretValue, "***MASKED***")
			}
		}
	}
	return result
}

func (sc *DefaultSecretContext) IsSecretMasked(secretName string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	masked, exists := sc.maskedSecrets[secretName]
	return exists && masked
}

func (sc *DefaultSecretContext) ListSecrets() []string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var secrets []string
	for name := range sc.secrets {
		secrets = append(secrets, name)
	}
	sort.Strings(secrets)
	return secrets
}

func (sc *DefaultSecretContext) GetSecretInfo(secretName string) (source string, masked bool, exists bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	_, exists = sc.secrets[secretName]
	if !exists {
		return "", false, false
	}

	source = sc.secretSources[secretName]
	masked = sc.maskedSecrets[secretName]
	return source, masked, true
}

// DefaultOutputContext implements OutputContext
type DefaultOutputContext struct {
	mu              sync.RWMutex
	captured        []byte
	capturing       bool
	realTimeEnabled bool
	filters         map[string]OutputFilter
}

func NewDefaultOutputContext() OutputContext {
	return &DefaultOutputContext{
		captured:        make([]byte, 0),
		capturing:       false,
		realTimeEnabled: false,
		filters:         make(map[string]OutputFilter),
	}
}

func (oc *DefaultOutputContext) StartCapture() {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.capturing = true
	oc.captured = make([]byte, 0)
}

func (oc *DefaultOutputContext) StopCapture() string {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.capturing = false
	result := string(oc.captured)
	oc.captured = make([]byte, 0)
	return result
}

func (oc *DefaultOutputContext) Write(data []byte) (int, error) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if oc.capturing {
		// Apply filters
		filteredData := data
		for _, filter := range oc.filters {
			if filter.ShouldApply("capture") {
				filteredData = filter.Filter(filteredData)
			}
		}

		oc.captured = append(oc.captured, filteredData...)
	}

	return len(data), nil
}

func (oc *DefaultOutputContext) Capture() ([]byte, error) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	result := make([]byte, len(oc.captured))
	copy(result, oc.captured)
	return result, nil
}

func (oc *DefaultOutputContext) StartWithRealTimeDisplay(realTime bool) {
	oc.StartCapture()
	oc.SetRealTimeEnabled(realTime)
}

func (oc *DefaultOutputContext) SetRealTimeEnabled(enabled bool) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.realTimeEnabled = enabled
}

func (oc *DefaultOutputContext) AddFilter(filter OutputFilter) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.filters[filter.Name()] = filter
}

func (oc *DefaultOutputContext) RemoveFilter(name string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	delete(oc.filters, name)
}

// DefaultActionContext implements ActionContext
type DefaultActionContext struct {
	executor *actions.ActionExecutor
}

func NewDefaultActionContext(executor *actions.ActionExecutor) ActionContext {
	return &DefaultActionContext{
		executor: executor,
	}
}

func (ac *DefaultActionContext) Execute(ctx context.Context, action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return ac.executor.Execute(ctx, action, args, options, silent)
}

func (ac *DefaultActionContext) GetMetadata(action string) (actions.ActionMetadata, bool) {
	// TODO: Implement metadata retrieval from action registry
	return actions.ActionMetadata{}, false
}

func (ac *DefaultActionContext) ListActions() []string {
	// TODO: Implement action listing
	return []string{}
}

func (ac *DefaultActionContext) ListActionsByCategory(category string) []string {
	// TODO: Implement category-based action listing
	return []string{}
}

func (ac *DefaultActionContext) SearchActions(prefix string) []string {
	// TODO: Implement action search
	return []string{}
}

func (ac *DefaultActionContext) ValidateAction(action string, args []interface{}) error {
	// TODO: Implement action validation
	return nil
}

func (ac *DefaultActionContext) GetActionCompletions(partial string) []string {
	// TODO: Implement action completions
	return []string{}
}

// DefaultLifecycleContext implements LifecycleContext
type DefaultLifecycleContext struct {
	mu              sync.RWMutex
	initialized     bool
	cleanupHandlers []CleanupHandler
	resourceUsage   ResourceUsage
}

func NewDefaultLifecycleContext() LifecycleContext {
	return &DefaultLifecycleContext{
		initialized:     false,
		cleanupHandlers: make([]CleanupHandler, 0),
		resourceUsage:   ResourceUsage{},
	}
}

func (lc *DefaultLifecycleContext) Initialize() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.initialized = true
	return nil
}

func (lc *DefaultLifecycleContext) Cleanup() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Sort cleanup handlers by priority (higher priority first)
	sort.Slice(lc.cleanupHandlers, func(i, j int) bool {
		return lc.cleanupHandlers[i].Priority() > lc.cleanupHandlers[j].Priority()
	})

	var errors []string
	for _, handler := range lc.cleanupHandlers {
		if err := handler.Cleanup(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", handler.Name(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (lc *DefaultLifecycleContext) Health() error {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if !lc.initialized {
		return fmt.Errorf("context not initialized")
	}

	return nil
}

func (lc *DefaultLifecycleContext) RegisterCleanupHandler(handler CleanupHandler) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.cleanupHandlers = append(lc.cleanupHandlers, handler)
}

func (lc *DefaultLifecycleContext) GetResourceUsage() ResourceUsage {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	return lc.resourceUsage
}

// DefaultVariableDebugger implements VariableDebugger
type DefaultVariableDebugger struct {
	mu      sync.RWMutex
	enabled bool
	history []SubstitutionEvent
}

func NewDefaultVariableDebugger() VariableDebugger {
	return &DefaultVariableDebugger{
		enabled: false,
		history: make([]SubstitutionEvent, 0),
	}
}

func (vd *DefaultVariableDebugger) EnableDebugging(enabled bool) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	vd.enabled = enabled
}

func (vd *DefaultVariableDebugger) LogSubstitution(original, resolved string, variables map[string]interface{}) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	if !vd.enabled {
		return
	}

	event := SubstitutionEvent{
		Timestamp: time.Now(),
		Original:  original,
		Resolved:  resolved,
		Variables: variables,
		Context:   "debugger",
	}

	vd.history = append(vd.history, event)
}

func (vd *DefaultVariableDebugger) GetDebugHistory() []SubstitutionEvent {
	vd.mu.RLock()
	defer vd.mu.RUnlock()

	result := make([]SubstitutionEvent, len(vd.history))
	copy(result, vd.history)
	return result
}

func (vd *DefaultVariableDebugger) ClearHistory() {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	vd.history = make([]SubstitutionEvent, 0)
}