# Robogo Test Execution Flow

This diagram shows how a test case flows through the Robogo framework, from YAML parsing to final results.

```mermaid
graph TD
    A[YAML Test File] --> B[CLI Parser]
    B --> C[TestRunner.LoadTest]
    C --> D[Parse YAML Structure]
    D --> E[Create Variables Store]
    E --> F[Load Pre-defined Variables]
    
    F --> G[Start Test Execution]
    G --> G1[Execute Setup Phase]
    G1 --> G2{Setup Steps Exist?}
    G2 -->|Yes| G3[Execute Setup Steps]
    G2 -->|No| H[Start Main Test Steps]
    G3 --> H
    H --> H1[For Each Main Step]
    
    H1 --> I[ControlFlowExecutor.ExecuteStepWithControlFlow]
    
    I --> J{Has For Loop?}
    J -->|Yes| K[LoopExecutor.ExecuteStepForLoop]
    K --> L[LoopParser.ParseIterations]
    L --> M[Execute Step for Each Iteration]
    
    J -->|No| N{Has While Loop?}
    N -->|Yes| O[LoopExecutor.ExecuteStepWhileLoop]
    O --> P[Evaluate While Condition]
    P --> Q[Execute Step Until Condition False]
    
    N -->|No| R{Has If Condition?}
    R -->|Yes| S[ConditionEvaluator.Evaluate]
    S --> T{Condition True?}
    T -->|No| U[Skip Step - Create Skipped Result]
    T -->|Yes| V[Execute Single Step]
    
    R -->|No| V
    M --> V
    Q --> V
    
    V --> W[StepExecutor.ExecuteSingleStep]
    W --> X[Variable Substitution]
    
    X --> Y[SubstitutionEngine.SubstituteArgs]
    Y --> Z[ExpressionEvaluator.Evaluate]
    Z --> AA{Variable Resolved?}
    AA -->|No| AB[ErrorAnalyzer.GenerateExprErrorSuggestions]
    AB --> AC[Mark as __UNRESOLVED__]
    AA -->|Yes| AD[Replace with Actual Value]
    
    AC --> AE[Get Action from Registry]
    AD --> AE
    
    AE --> AF{Action Exists?}
    AF -->|No| AG[Return Unknown Action Error]
    AF -->|Yes| AH[Execute Action]
    
    AH --> AI{Action Type}
    AI -->|HTTP| AJ[HTTP Action - Make Request]
    AI -->|Database| AK[Database Action - Execute Query]
    AI -->|Assert| AL[Assert Action - Check Values]
    AI -->|Variable| AM[Variable Action - Store Value]
    AI -->|Log| AN[Log Action - Print Message]
    AI -->|jq| AO1[jq Action - Extract JSON Data]
    AI -->|xpath| AO2[xpath Action - Extract XML Data]
    AI -->|Other| AO[Custom Action Handler]
    
    AJ --> AP[Process HTTP Response]
    AK --> AQ[Process Database Result]
    AL --> AR{Assert Check}
    AR -->|Pass| AS[Return Success Result]
    AR -->|Fail| AT[Return Failure Result]
    AM --> AU[Store in Variable Store]
    AN --> AV[Print to Console]
    AO1 --> AW1[Extract JSON Data]
    AO2 --> AW2[Extract XML Data]
    AO --> AW[Custom Processing]
    
    AP --> AX[Create ActionResult]
    AQ --> AX
    AS --> AX
    AT --> AX
    AU --> AX
    AV --> AX
    AW1 --> AX
    AW2 --> AX
    AW --> AX
    
    AX --> AY{Has Retry Config?}
    AY -->|Yes| AZ[RetryExecutor.ExecuteStepWithRetry]
    AZ --> BA{Should Retry?}
    BA -->|Yes| BB[Calculate Delay]
    BB --> BC[Wait and Retry]
    BC --> W
    BA -->|No| BD[Return Final Result]
    
    AY -->|No| BD
    
    BD --> BE[Print Step Result]
    BE --> BF{Store Result Variable?}
    BF -->|Yes| BG[Store in Variables]
    BF -->|No| BH[Continue to Next Step]
    BG --> BH
    
    BH --> BI{More Steps?}
    BI -->|Yes| H1
    BI -->|No| BJ[Execute Teardown Phase]
    BJ --> BJ1{Teardown Steps Exist?}
    BJ1 -->|Yes| BJ2[Execute Teardown Steps]
    BJ1 -->|No| BK[Calculate Test Duration]
    BJ2 --> BK
    
    BK --> BL[Print Test Summary]
    BL --> BM[Print Results Table]
    BM --> BN{Test Passed?}
    BN -->|Yes| BO[Exit Code 0]
    BN -->|No| BP[Exit Code 2]
    
    AG --> BQ[Print Error]
    BQ --> BP
    U --> BH

    style A fill:#e1f5fe
    style AX fill:#f3e5f5
    style BD fill:#e8f5e8
    style BO fill:#c8e6c8
    style BP fill:#ffcdd2
    style AL fill:#fff3e0
    style AR fill:#fff3e0
```

## Key Components

### **Entry Points**
- **CLI Parser**: Handles command-line arguments and routes to appropriate handlers
- **TestRunner**: Main orchestrator that loads and executes tests

### **Control Flow Management**
- **ControlFlowExecutor**: Main coordinator for step execution
- **LoopExecutor**: Handles for/while loop execution
- **ConditionEvaluator**: Evaluates if/while conditions

### **Variable System**
- **VariableStore**: Core storage for variables
- **SubstitutionEngine**: Processes simple ${variable} templates
- **Data Extraction**: Uses `jq` for JSON/structured data, `xpath` for XML
- **ErrorAnalyzer**: Provides hints to use `jq`/`xpath` for complex access

### **Action Execution**
- **Action Registry**: Maps action names to implementations
- **StepExecutor**: Core step execution without retry
- **RetryExecutor**: Handles retry logic with backoff strategies

### **Result Handling**
- **ActionResult**: Standardized result format
- **StepResult**: Complete step execution result
- **TestResult**: Final test outcome with summary

## Execution Patterns

### **Test with Setup/Teardown**
```yaml
testcase: "User Management Test"
setup:
  - name: "Create test database"
    action: postgres
    args: ["execute", "${db_url}", "CREATE TABLE test_users..."]

steps:
  - name: "Test user creation"
    action: log
    args: ["Running main test..."]

teardown:
  - name: "Clean up test data"
    action: postgres
    args: ["execute", "${db_url}", "DROP TABLE test_users"]
```
Flow: CLI → TestRunner → Setup Phase → Main Steps → Teardown Phase → Results

### **Simple Step**
```yaml
- name: "Log message"
  action: log
  args: ["Hello World"]
```
Flow: CLI → TestRunner → ControlFlow → StepExecutor → Log Action → Result

### **Step with Variables & Data Extraction**
```yaml
- name: "HTTP Request"
  action: http
  args: ["GET", "${base_url}/api/users"]
  result: users_response

- name: "Extract user count"
  action: jq
  args: ["${users_response}", ".body | fromjson | length"]
  result: user_count
```
Flow: CLI → TestRunner → ControlFlow → Variable Substitution → HTTP Action → Store Result → jq Action → Extract Data

### **Step with Loops**
```yaml
- name: "Process each user"
  action: log
  args: ["Processing user ${item}"]
  for: "[alice,bob,charlie]"
```
Flow: CLI → TestRunner → ControlFlow → LoopExecutor → (3x Step Execution) → Results

### **Step with Retry**
```yaml
- name: "Flaky API call"
  action: http
  args: ["GET", "${api_url}"]
  retry:
    attempts: 3
    delay: "1s"
    backoff: "exponential"
```
Flow: CLI → TestRunner → ControlFlow → RetryExecutor → (Up to 3x HTTP Action) → Final Result
