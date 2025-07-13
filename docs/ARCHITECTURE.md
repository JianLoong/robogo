# Robogo Architecture

Robogo is a modern, git-driven test automation framework built with a clean, service-oriented architecture. This document outlines the key architectural patterns, components, and design decisions.

## Architecture Overview

Robogo follows modern Go architectural patterns with a focus on:
- **Interface-driven design** for maximum decoupling
- **Dependency injection** for clean separation of concerns
- **Service-oriented architecture** with single responsibilities
- **Composition over inheritance** for flexibility
- **Event-driven patterns** for loose coupling

## Core Architecture Principles

### 1. Interface-First Design
All major components are interface-driven, enabling easy testing, mocking, and extensibility:

```go
type TestExecutor interface {
    ExecuteTestCase(ctx context.Context, testCase *parser.TestCase, silent bool) (*parser.TestResult, error)
    ExecuteTestSuite(ctx context.Context, testSuite *parser.TestSuite, filePath string, silent bool) (*parser.TestSuiteResult, error)
}

type StepExecutor interface {
    ExecuteStep(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error)
    ExecuteSteps(ctx context.Context, steps []parser.Step, silent bool) ([]parser.StepResult, error)
}
```

### 2. Composition-Based Context
The execution context is composed of focused, single-responsibility interfaces:

```go
type TestExecutionContext interface {
    Variables() VariableContext
    Secrets() SecretContext
    Output() OutputContext
    Actions() ActionContext
    Lifecycle() LifecycleContext
}
```

### 3. Service Factory Pattern
Centralized service creation with proper dependency wiring:

```go
type ServiceFactory interface {
    CreateTestExecutor(executor *actions.ActionExecutor) TestExecutor
    CreateStepExecutor(context TestExecutionContext) StepExecutor
    CreateTestSuiteExecutor(runner TestExecutor) TestSuiteExecutor
}
```

## Component Architecture

### Core Services

#### TestExecutionService
The primary service for test case execution with clean separation of concerns:

**Location**: `internal/runner/test_execution_service.go`

**Responsibilities**:
- Test case lifecycle management
- Variable initialization and cleanup
- Result aggregation and reporting
- Error handling and context management

**Key Features**:
- Interface-driven design (`TestExecutor`)
- Dependency injection of `TestExecutionContext`
- Comprehensive error handling with context preservation

#### StepExecutionService
Handles individual step execution with advanced control flow:

**Location**: `internal/runner/step_execution_service.go`

**Responsibilities**:
- Single step execution with retry logic
- Control flow (`if`, `for`, `while` loops)
- Parallel step execution with dependency analysis
- Variable substitution and preprocessing

**Key Features**:
- Enhanced error handling with retry/recovery patterns
- Parallel execution with concurrency control
- Variable substitution with dot notation support

#### TestSuiteRunner
Orchestrates test suite execution with setup/teardown support:

**Location**: `internal/runner/testsuite_runner.go`

**Responsibilities**:
- Test suite orchestration
- Setup and teardown execution
- Suite-level variable management
- Parallel test case execution

### Context Architecture

The execution context is built using composition of focused interfaces:

#### VariableContext
**Location**: `internal/runner/context.go`

**Responsibilities**:
- Variable storage and retrieval
- Template substitution with dot notation
- Variable debugging and introspection
- Event-driven variable change notifications

**Key Features**:
- Dot notation support: `${http_response.status_code}`
- JSON string parsing for nested property access
- Variable debugging with substitution history
- Repository pattern with event notifications

#### SecretContext
**Responsibilities**:
- Secure credential management
- File-based and inline secret loading
- Output masking for sensitive data
- Secret metadata and source tracking

#### ActionContext
**Responsibilities**:
- Action execution and delegation
- Action metadata and discovery
- Action validation and completions

#### OutputContext
**Responsibilities**:
- Output capture and management
- Real-time output display
- Output filtering and processing

#### LifecycleContext
**Responsibilities**:
- Resource lifecycle management
- Cleanup handler registration
- Health monitoring and resource usage

### Variable Management Architecture

#### Repository Pattern
**Location**: `internal/runner/variable_repository.go`

Uses the repository pattern for data persistence with event-driven notifications:

```go
type VariableRepository interface {
    Store(key string, value interface{}) error
    Retrieve(key string) (interface{}, bool)
    Delete(key string) error
    List() map[string]interface{}
    Subscribe(listener VariableChangeListener)
}
```

#### Variable Service
Provides business logic layer over the repository:

```go
type VariableService interface {
    SetVariable(key string, value interface{}) error
    GetVariable(key string) (interface{}, bool)
    SubstituteTemplate(template string) string
    SubstituteArgs(args []interface{}) []interface{}
}
```

**Key Features**:
- Enhanced dot notation with JSON parsing
- Event-driven change notifications
- Comprehensive substitution with context preservation
- Debug support with substitution history

### Validation Architecture

#### Rule-Based Validation Engine
**Location**: `internal/runner/enhanced_validation_engine.go`

Implements a pluggable, rule-based validation system:

```go
type ValidationEngine interface {
    ValidateTestCase(testCase *parser.TestCase) ValidationReport
    ValidateTestSuite(testSuite *parser.TestSuite) ValidationReport
    RegisterRule(rule ValidationRule)
}

type ValidationRule interface {
    Name() string
    Validate(context ValidationContext) []ValidationError
    ShouldApply(context ValidationContext) bool
}
```

**Built-in Rules**:
- **RequiredFieldRule**: Validates mandatory fields
- **ActionValidationRule**: Validates action existence and parameters
- **DependencyValidationRule**: Prevents circular dependencies
- **SecurityValidationRule**: Detects potential security issues
- **PerformanceValidationRule**: Identifies performance concerns
- **BestPracticeValidationRule**: Enforces testing best practices

### CLI Integration

#### FileProcessor
**Location**: `internal/cli/fileops.go`

Handles file discovery and processing with unified architecture:

```go
// Both test cases and test suites use the modern service architecture
func (fp *FileProcessor) processCaseFile(ctx context.Context, filePath string) (*RunResults, error) {
    testCase, err := parser.ParseTestFile(filePath)
    if err != nil {
        return nil, err
    }

    testExecutor := runner.NewTestExecutionService(fp.executor)
    if fp.options.VariableDebug {
        testExecutor.GetContext().EnableVariableDebugging(true)
    }

    result, err := testExecutor.ExecuteTestCase(ctx, testCase, fp.options.Silent)
    // ...
}
```

## Data Flow

### Test Case Execution Flow

1. **CLI** → `FileProcessor.processCaseFile()`
2. **Parser** → `parser.ParseTestFile()` → `TestCase`
3. **Service Factory** → `NewTestExecutionService()` → `TestExecutor`
4. **Test Executor** → `ExecuteTestCase()`:
   - Initialize context and variables
   - Create step executor
   - Execute steps with dependency management
   - Aggregate results and cleanup
5. **Step Executor** → `ExecuteSteps()`:
   - Preprocess steps (variable substitution)
   - Execute individual steps with control flow
   - Handle parallel execution and dependencies
6. **Action Execution** → `ActionContext.Execute()`
7. **Result Aggregation** → `TestResult`

### Test Suite Execution Flow

1. **CLI** → `FileProcessor.processSuiteFile()`
2. **Parser** → `parser.ParseTestSuite()` → `TestSuite`
3. **Service Factory** → `NewTestSuiteRunner()` → `TestSuiteExecutor`
4. **Suite Runner** → `RunTestSuite()`:
   - Execute setup steps
   - Run test cases (sequential/parallel)
   - Execute teardown steps
   - Aggregate suite results
5. **Test Case Execution** → Delegates to `TestExecutor`

### Variable Substitution Flow

1. **Template Input** → `${variable.property}`
2. **Variable Service** → `SubstituteTemplate()`
3. **Dot Notation Parser** → Extract base variable and property path
4. **Repository Lookup** → Retrieve base variable
5. **JSON Parsing** → Parse JSON strings if needed
6. **Property Navigation** → Navigate nested structure
7. **Substituted Output** → Resolved value

## Key Design Patterns

### 1. Dependency Injection
All services receive their dependencies through constructors:

```go
func NewTestExecutionService(executor *actions.ActionExecutor) TestExecutor {
    execContext := NewTestExecutionContext(executor)
    stepService := NewStepExecutionService(execContext)
    
    return &TestExecutionService{
        context:     execContext,
        stepService: stepService,
    }
}
```

### 2. Repository Pattern
Variable storage is abstracted behind repository interface:

```go
type VariableRepository interface {
    Store(key string, value interface{}) error
    Retrieve(key string) (interface{}, bool)
    // ...
}
```

### 3. Event-Driven Architecture
Variable changes trigger events for loose coupling:

```go
type VariableChangeListener interface {
    OnVariableChanged(event VariableChangeEvent)
}
```

### 4. Strategy Pattern
Validation rules implement pluggable validation strategies:

```go
type ValidationRule interface {
    Validate(context ValidationContext) []ValidationError
}
```

### 5. Template Method Pattern
Step execution follows a consistent template with hooks:

```go
func (ses *StepExecutionService) ExecuteStep(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
    // 1. Preprocess (variable substitution)
    // 2. Execute with retry/recovery
    // 3. Store results
    // 4. Handle errors
}
```

## Error Handling Strategy

### 1. Structured Error Context
Errors include rich context for debugging:

```go
type RobogoError struct {
    Type         string
    Message      string
    Step         string
    Action       string
    Cause        error
    Context      map[string]interface{}
    Breadcrumbs  []string
}
```

### 2. Retry and Recovery
Steps can be configured with retry and recovery strategies:

```go
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    BackoffFactor float64
}

type RecoveryConfig struct {
    FallbackAction string
    MaxRecoveries  int
}
```

### 3. Graceful Degradation
Services handle partial failures gracefully and continue execution where possible.

## Performance Characteristics

### 1. Parallel Execution
- **Test-level parallelism**: Multiple test cases in parallel
- **Step-level parallelism**: Independent steps within a test
- **Dependency analysis**: Automatic detection of step dependencies
- **Concurrency control**: Configurable limits to prevent resource exhaustion

### 2. Memory Management
- **Context cleanup**: Automatic resource cleanup after test execution
- **Variable scoping**: Proper variable lifecycle management
- **Output streaming**: Efficient output handling without memory accumulation

### 3. Efficient Variable Substitution
- **Lazy evaluation**: Variables resolved only when needed
- **Caching**: Substitution results cached for repeated access
- **Minimal parsing**: Efficient dot notation parsing

## Security Architecture

### 1. Secret Management
- **File-based secrets**: Secrets loaded from external files
- **Output masking**: Automatic masking of sensitive data in logs
- **Secure storage**: Secrets stored separately from regular variables

### 2. Validation Security
- **Input validation**: Comprehensive validation of test definitions
- **Security rules**: Automatic detection of potential security issues
- **Safe execution**: Sandboxed test execution environment

## Extensibility Points

### 1. Custom Actions
New actions can be added through the action registry:

```go
registry.RegisterAction("custom_action", NewCustomAction())
```

### 2. Custom Validation Rules
Additional validation rules can be registered:

```go
engine.RegisterRule(NewCustomValidationRule())
```

### 3. Custom Variable Sources
New variable sources can be implemented through the repository interface.

### 4. Custom Output Filters
Output processing can be extended with custom filters:

```go
outputContext.AddFilter(NewCustomOutputFilter())
```

## Testing Strategy

### 1. Interface Mocking
All major components are interface-driven, enabling comprehensive unit testing with mocks.

### 2. Service Testing
Individual services can be tested in isolation with dependency injection.

### 3. Integration Testing
Full end-to-end testing through the CLI interface.

### 4. Contract Testing
Interface contracts are validated through comprehensive test suites.

## Future Architecture Considerations

### 1. Microservice Extraction
The service-oriented architecture enables future extraction of components into microservices if needed.

### 2. Plugin Architecture
The interface-driven design supports future plugin systems for extending functionality.

### 3. Distributed Execution
The context abstraction enables future distributed test execution across multiple nodes.

### 4. Enhanced Observability
The event-driven architecture provides hooks for comprehensive monitoring and observability.

## File Structure

```
internal/runner/
├── context.go                     # Execution context interfaces and implementations
├── test_execution_service.go      # Primary test execution service
├── step_execution_service.go      # Step execution with control flow
├── testsuite_runner.go           # Test suite orchestration
├── variable_repository.go        # Variable management with repository pattern
├── enhanced_validation_engine.go # Rule-based validation system
├── validation_framework.go       # Validation interfaces and rules
├── service_factory.go           # Service creation and dependency injection
├── interfaces.go                # Core service interfaces
├── output_capture.go           # Output capture and management
├── retry_manager.go            # Retry and recovery mechanisms
└── skip_logic.go               # Skip condition evaluation
```

This architecture provides a solid foundation for maintainable, testable, and extensible test automation that can evolve with changing requirements while maintaining clean separation of concerns.