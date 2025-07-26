# Robogo Error & Failure States Diagram

This diagram explains Robogo's execution flow and the distinction between errors, failures, and various step states.

## Execution Strategy Flow

```mermaid
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
    

```

## Error vs Failure Classification

```mermaid
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
    

```

## Step Status Outcomes

```mermaid
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