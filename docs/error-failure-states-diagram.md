# Robogo Error & Failure States Diagram

This diagram explains Robogo's execution flow and the distinction between errors, failures, and various step states.

## Execution Strategy Flow

```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'primaryColor': '#2563eb',
    'primaryTextColor': '#1f2937',
    'primaryBorderColor': '#3b82f6',
    'lineColor': '#6b7280',
    'secondaryColor': '#f3f4f6',
    'tertiaryColor': '#e5e7eb',
    'background': '#ffffff',
    'mainBkg': '#f9fafb',
    'secondBkg': '#f3f4f6',
    'tertiaryBkg': '#e5e7eb'
  }
}}%%
flowchart TD
    Start([Step Execution]) --> Router{Strategy Router}
    
    Router --> |Priority 4| Conditional[Conditional Strategy]
    Router --> |Priority 3| Retry[Retry Strategy]  
    Router --> |Priority 2| Nested[Nested Strategy]
    Router --> |Priority 1| Basic[Basic Strategy]
    
    Conditional --> CondCheck{Condition Check}
    CondCheck --> |true| Execute[Execute Action]
    CondCheck --> |false| Skipped[SKIPPED]
    
    Basic --> Execute
    Retry --> Execute
    Nested --> Execute
    
    Execute --> Result{Action Result}
    
    classDef successStyle fill:#dcfce7,stroke:#16a34a,stroke-width:2px,color:#15803d
    classDef skipStyle fill:#fef3c7,stroke:#ca8a04,stroke-width:2px,color:#a16207
    classDef processStyle fill:#dbeafe,stroke:#2563eb,stroke-width:2px,color:#1d4ed8
    
    class Execute,Result successStyle
    class Skipped skipStyle
    class Router,Conditional,Retry,Nested,Basic,CondCheck processStyle
```

## Error vs Failure Classification

```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'primaryColor': '#2563eb',
    'primaryTextColor': '#1f2937',
    'primaryBorderColor': '#3b82f6',
    'lineColor': '#6b7280',
    'secondaryColor': '#f3f4f6',
    'tertiaryColor': '#e5e7eb',
    'background': '#ffffff',
    'mainBkg': '#f9fafb',
    'secondBkg': '#f3f4f6',
    'tertiaryBkg': '#e5e7eb'
  }
}}%%
flowchart TD
    ActionResult{Action Execution Result}
    
    ActionResult --> |Success| Pass[PASS Status]
    ActionResult --> |Technical Problems| ErrorInfo[ErrorInfo]
    ActionResult --> |Logical Problems| FailureInfo[FailureInfo]
    
    ErrorInfo --> NetworkErr[Network Issues]
    ErrorInfo --> DatabaseErr[Database Issues]
    ErrorInfo --> ParseErr[Parse Errors]
    ErrorInfo --> SystemErr[System Issues]
    ErrorInfo --> ConfigErr[Config Issues]
    
    FailureInfo --> AssertErr[Assertion Failures]
    FailureInfo --> ValidErr[Validation Failures]  
    FailureInfo --> BusinessErr[Business Logic Issues]
    
    NetworkErr --> ErrorStatus[ERROR Status]
    DatabaseErr --> ErrorStatus
    ParseErr --> ErrorStatus
    SystemErr --> ErrorStatus
    ConfigErr --> ErrorStatus
    
    AssertErr --> FailStatus[FAIL Status]
    ValidErr --> FailStatus
    BusinessErr --> FailStatus
    
    Pass --> Report[Test Report]
    ErrorStatus --> Report
    FailStatus --> Report
    
    classDef successStyle fill:#dcfce7,stroke:#16a34a,stroke-width:2px,color:#15803d
    classDef errorStyle fill:#fecaca,stroke:#dc2626,stroke-width:2px,color:#b91c1c
    classDef failureStyle fill:#fed7aa,stroke:#ea580c,stroke-width:2px,color:#c2410c
    classDef processStyle fill:#dbeafe,stroke:#2563eb,stroke-width:2px,color:#1d4ed8
    
    class Pass,Report successStyle
    class ErrorInfo,NetworkErr,DatabaseErr,ParseErr,SystemErr,ConfigErr,ErrorStatus errorStyle
    class FailureInfo,AssertErr,ValidErr,BusinessErr,FailStatus failureStyle
    class ActionResult processStyle
```

## Step Status Outcomes

```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'primaryColor': '#2563eb',
    'primaryTextColor': '#1f2937',
    'primaryBorderColor': '#3b82f6',
    'lineColor': '#6b7280',
    'secondaryColor': '#f3f4f6',
    'tertiaryColor': '#e5e7eb',
    'background': '#ffffff',
    'mainBkg': '#f9fafb',
    'secondBkg': '#f3f4f6',
    'tertiaryBkg': '#e5e7eb',
    'stateBkg': '#f1f5f9',
    'stateLabelColor': '#374151',
    'cScale0': '#dcfce7',
    'cScale1': '#fef3c7',
    'cScale2': '#fecaca',
    'cScale3': '#fed7aa'
  }
}}%%
stateDiagram-v2
    [*] --> Executing
    
    Executing --> PASS : Action succeeds
    Executing --> SKIPPED : Condition false
    Executing --> ERROR : Technical error
    Executing --> FAIL : Logical failure
    
    PASS --> [*]
    SKIPPED --> [*] 
    ERROR --> [*]
    FAIL --> [*]
    
    note right of PASS
        ✅ Action completed successfully  
        All assertions passed
    end note
    
    note right of SKIPPED
        ⏭️ Step bypassed
        Conditional logic (if: false)
    end note
    
    note right of ERROR
        ❌ Technical problem
        ErrorInfo - system/infrastructure issue
    end note
    
    note right of FAIL
        ❌ Logical failure
        FailureInfo - test expectation not met
    end note
```

## Key Concepts Explained

### 1. **Execution Strategy Priority System**
- **Priority 4**: Conditional logic (`if` statements) - highest priority for control flow
- **Priority 3**: Retry logic - handles retry configurations with backoff
- **Priority 2**: Nested steps - manages step collections and sub-workflows  
- **Priority 1**: Basic execution - fallback for standard action execution

### 2. **Dual Error Classification System**

#### **ErrorInfo (Technical Problems)**
These prevent proper execution and indicate infrastructure or configuration issues:
- **Network Issues**: Connection timeouts, DNS failures, unreachable services
- **Database Issues**: Invalid credentials, server downtime, malformed connection strings
- **Parse/Serialization**: Malformed JSON/XML, invalid data formats
- **System Issues**: File permissions, missing resources, OS-level problems
- **Configuration**: Missing parameters, invalid URLs, bad authentication

#### **FailureInfo (Logical Problems)**
These indicate the system worked but produced unexpected results:
- **Assertion Failures**: Expected vs actual value mismatches
- **Validation Failures**: Response format issues, missing fields, schema violations
- **Business Logic**: User conflicts, permission issues, data integrity problems

### 3. **Step Status Outcomes**

| Status | Meaning | When It Occurs |
|--------|---------|----------------|
| **PASS** | ✅ Success | Action completed successfully, all assertions passed |
| **SKIPPED** | ⏭️ Bypassed | Conditional logic (`if: false`) caused step to be skipped |
| **ERROR** | ❌ Technical Problem | Infrastructure/system issues (ErrorInfo) occurred |
| **FAIL** | ❌ Logical Problem | Test expectations not met (FailureInfo) occurred |

### 4. **Unified Result Access**
Despite the internal distinction between `ErrorInfo` and `FailureInfo`, both are accessible through:
- `result.GetMessage()` - Returns human-readable error or failure message
- `result.HasIssue()` - Returns true for either errors or failures
- Final test reports treat both as failures for execution flow control

### 5. **Design Benefits**
- **Clear Problem Classification**: Separates "system broken" from "test failed"
- **Targeted Debugging**: Technical errors suggest infrastructure fixes, failures suggest code issues
- **Consistent Interface**: Unified access pattern despite internal complexity
- **Comprehensive Coverage**: All execution paths lead to meaningful status reporting

This dual-classification system enables Robogo to provide precise feedback while maintaining a simple interface for test authors.