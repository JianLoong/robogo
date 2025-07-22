# Robogo Test Execution Flow

This diagram shows how a test case flows through the simplified Robogo framework architecture.

```mermaid
graph TD
    A[YAML Test File] --> B[CLI Parser]
    B --> C[TestRunner.LoadTest]
    C --> D[Parse YAML Structure]
    D --> E[Create Variables Store]
    E --> F[Load Pre-defined Variables]
    
    F --> G[Create ExecutionStrategyRouter]
    G --> G1[ActionRegistry.NewActionRegistry]
    G1 --> G2[BasicConditionEvaluator]
    G2 --> H[Register Strategies Directly]
    H --> H1[ConditionalExecutionStrategy - Priority 4]
    H --> H2[RetryExecutionStrategy - Priority 3] 
    H --> H3[NestedStepsExecutionStrategy - Priority 2]
    H --> H4[BasicExecutionStrategy - Priority 1]
    
    H1 --> I[Execute Setup Steps]
    H2 --> I
    H3 --> I
    H4 --> I
    
    I --> J[Execute Main Steps]
    J --> K{Step Type?}
    
    K -->|Has If Condition| L[ConditionalExecutionStrategy]
    L --> L1[Evaluate Condition]
    L1 -->|True| L2[Remove If, Execute via Router]
    L1 -->|False| L3[Return SKIPPED Result]
    
    K -->|Has Retry Config| M[RetryExecutionStrategy]
    M --> M1[Attempt 1]
    M1 -->|Failed| M2[Wait Delay]
    M2 --> M3[Attempt 2...N]
    M3 -->|Success or Max Attempts| M4[Return Result]
    
    K -->|Has Nested Steps| N[NestedStepsExecutionStrategy]
    N --> N1[Execute Each Nested Step]
    N1 --> N2[Aggregate Results]
    
    K -->|Simple Action| O[BasicExecutionStrategy]
    O --> O1[Get Action from Registry]
    O1 --> O2[Substitute Variables in Args]
    O2 --> O3[Execute Action Function]
    O3 --> O4[Apply Extraction if Configured]
    O4 --> O5[Store Result Variable if Specified]
    
    L2 --> P[Step Completed]
    M4 --> P
    N2 --> P
    O5 --> P
    L3 --> P
    
    P --> Q{More Steps?}
    Q -->|Yes| J
    Q -->|No| R[Execute Teardown Steps]
    
    R --> S[Generate Test Result]
    S --> T[Print Summary]
    T --> U[Exit with Status Code]

    style A fill:#e1f5fe
    style G fill:#f3e5f5
    style K fill:#fff3e0
    style P fill:#e8f5e8
    style U fill:#ffebee
```

## Architecture Overview

### Core Components

1. **CLI**: Direct command-line interface handling `run`, `list`, `version` commands
2. **TestRunner**: Orchestrates test execution lifecycle with direct strategy router
3. **ExecutionStrategyRouter**: Priority-based routing to appropriate execution strategy
4. **Variables**: Simple map-based variable storage with substitution (no complex abstractions)
5. **Strategy Pattern**: Clean execution routing for different step types

### Execution Strategies (Priority Order)

1. **ConditionalExecutionStrategy** (Priority 4): Handles `if` conditions
2. **RetryExecutionStrategy** (Priority 3): Handles `retry` configuration  
3. **NestedStepsExecutionStrategy** (Priority 2): Handles `steps` arrays
4. **BasicExecutionStrategy** (Priority 1): Default fallback for simple actions

### Variable System

- **Simple Direct Approach**: Single `Variables` struct with map storage
- **Substitution**: `${variable}` and `${ENV:VARIABLE}` syntax
- **No Complex Abstractions**: Removed VariableManager, TemplateSubstitution layers

### Action System

- **ActionRegistry**: Instance-based action storage (no global state)
- **Direct Function Calls**: Actions are simple functions, no interfaces
- **Built-in Actions**: HTTP, Database (PostgreSQL, Spanner), Messaging (Kafka, RabbitMQ), etc.

## Example Flows

### Simple HTTP Test
```
CLI → TestRunner → ExecutionStrategyRouter → BasicExecutionStrategy → HTTP Action → Result
```

### Conditional Test with Retry
```
CLI → TestRunner → ExecutionStrategyRouter → ConditionalExecutionStrategy → 
RetryExecutionStrategy → BasicExecutionStrategy → Action → Result
```

### Nested Steps Test
```
CLI → TestRunner → ExecutionStrategyRouter → NestedStepsExecutionStrategy → 
(Multiple BasicExecutionStrategy calls) → Aggregated Result
```

## Key Simplifications Made

- **Removed**: VariableManager, TemplateSubstitution, ActionExecutor interface layers
- **Removed**: ControlFlowExecutor, StepExecutor, LoopExecutionStrategy dead code
- **Eliminated**: ExecutionPipeline, Dependencies, DependencyInjector, UnifiedExecutor abstraction layers
- **Simplified**: Direct Variables map instead of 4-layer abstraction
- **Direct Construction**: TestRunner directly creates and uses ExecutionStrategyRouter
- **Consolidated**: Strategy pattern handles all control flow (conditions, retry, nesting)
- **Maintained**: All functionality while removing ~1500+ lines of over-engineered code