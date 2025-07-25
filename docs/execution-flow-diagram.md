# Robogo Test Execution Flow

This diagram shows how a test case flows through the current Robogo framework architecture.

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
graph TD
    A[YAML Test File] --> B[CLI Parser]
    B --> C[TestRunner.LoadTest]
    C --> D[Parse YAML Structure]
    D --> E[NewTestRunner Creation]
    E --> E1[Create Variables Store]
    E1 --> E2[Create ActionRegistry]
    E2 --> E3[Create BasicConditionEvaluator]
    E3 --> E4[Create ExecutionStrategyRouter]
    E4 --> F[Register All Strategies Together]
    F --> F1[ConditionalExecutionStrategy - Priority 4]
    F --> F2[RetryExecutionStrategy - Priority 3] 
    F --> F3[NestedStepsExecutionStrategy - Priority 2]
    F --> F4[BasicExecutionStrategy - Priority 1]
    
    F1 --> G[Load Pre-defined Variables]
    F2 --> G
    F3 --> G
    F4 --> G
    
    G --> I[Execute Setup Steps]
    
    I --> J[Execute Main Steps]
    J --> K{Router Selects Strategy<br/>Based on CanHandle + Priority}
    
    K -->|"step.If != ''"| L[ConditionalExecutionStrategy]
    L --> L1[Evaluate Condition with Variables]
    L1 -->|True| L2[Remove If Property, Route Back to Router]
    L1 -->|False| L3[Return SKIPPED Result]
    
    K -->|"step.Retry != nil"| M[RetryExecutionStrategy]
    M --> M1[Attempt Execution via BasicExecutionStrategy]
    M1 -->|Failed & Retries Left| M2[Wait Delay Period]
    M2 --> M3[Next Attempt]
    M3 -->|Success or Max Attempts| M4[Return Final Result<br/>✅ Uses BasicExecutionStrategy internally for proper result storage]
    
    K -->|"len(step.Steps) > 0"| N[NestedStepsExecutionStrategy]
    N --> N1[Execute Each Sub-Step via Router]
    N1 --> N2[Collect All Results]
    N2 --> N3[Return Aggregated Result<br/>📝 Individual nested steps handle their own result storage]
    
    K -->|"step.Action != '' & no other conditions"| O[BasicExecutionStrategy]
    O --> O1[Get Action from Registry]
    O1 --> O2[Apply Variable Substitution]
    O2 --> O3[Check Security Settings]
    O3 -->|no_log = true| O4[Suppress Step Logging]
    O3 -->|sensitive_fields| O5[Mask Specified Fields]
    O3 --> O6[Execute Action Function]
    O6 --> O7[Process Action Result]
    O7 --> O8[Apply Data Extraction if step.Extract]
    O8 --> O9[Store Result Variable if step.Result specified<br/>⚠️ Only BasicExecutionStrategy handles this properly]
    
    L2 --> K
    L3 --> P[Step Completed]
    M4 --> P
    N3 --> P
    O9 --> P
    
    P --> Q{More Steps?}
    Q -->|Yes| J
    Q -->|No| R[Execute Teardown Steps]
    
    R --> S[Generate Test Summary]
    S --> T[Apply Output Masking]
    T --> U[Print Results Table]
    U --> V[Exit with Status Code]

    %% Action Types
    O6 --> A1[HTTP Actions]
    O6 --> A2[Database Actions<br/>PostgreSQL, Spanner]
    O6 --> A3[Messaging Actions<br/>Kafka, RabbitMQ]
    O6 --> A4[File Actions<br/>file_read, scp]
    O6 --> A5[Utility Actions<br/>uuid, time, sleep, ping]
    O6 --> A6[Data Processing<br/>jq, xpath, json_parse]
    O6 --> A7[String Actions<br/>string_random, format]
    O6 --> A8[Core Actions<br/>log, assert, variable]

    %% Dark-mode friendly styling
    classDef startNode fill:#dbeafe,stroke:#3b82f6,stroke-width:2px,color:#1e40af
    classDef processNode fill:#ecfdf5,stroke:#10b981,stroke-width:2px,color:#047857
    classDef decisionNode fill:#fef3c7,stroke:#f59e0b,stroke-width:2px,color:#92400e
    classDef completeNode fill:#f0f9ff,stroke:#0ea5e9,stroke-width:2px,color:#0c4a6e
    classDef securityNode fill:#fef2f2,stroke:#ef4444,stroke-width:2px,color:#b91c1c
    classDef actionNode fill:#f5f3ff,stroke:#8b5cf6,stroke-width:2px,color:#6d28d9

    class A,E,F startNode
    class B,C,D,E1,E2,E3,E4,G,I,J,L1,L2,M1,M2,M3,N1,N2,O1,O2,O7,O8,R,S,T,U processNode
    class K,Q decisionNode
    class P,V completeNode
    class O3,O4,O5 securityNode
    class A1,A2,A3,A4,A5,A6,A7,A8 actionNode
```

## Current Architecture Highlights

**Strategy Pattern (Priority-Based Routing):**
- **Conditional** → **Retry** → **NestedSteps** → **Basic**
- Each strategy handles its specific concern and delegates to the next

**Security-Aware Execution:**
- `no_log` suppresses step logging for sensitive operations
- `sensitive_fields` masks specific data in logs and output
- Automatic masking of passwords, tokens, API keys

**Action System:**
- **27+ Built-in Actions** across 8 categories
- **SCP Action** for secure file transfers via SSH/SFTP
- **Direct Function Calls** - no interfaces or abstractions

**Variable System:**
- Simple `${variable}` and `${ENV:VARIABLE}` substitution  
- No complex templating engines or dependency injection
- Variables resolved before action execution

**Result Storage Notes:**
- **BasicExecutionStrategy & RetryExecutionStrategy**: Properly handle `step.Result` variable storage
- **ConditionalExecutionStrategy**: Inherits result storage behavior from delegated strategies  
- **NestedStepsExecutionStrategy**: Individual nested steps handle their own result storage (parent step aggregation by design)

**Key Simplifications:**
- Removed VariableManager, TemplateSubstitution layers
- Eliminated dependency injection system
- Direct construction throughout
- Single Variables struct instead of 4-layer abstraction

This reflects the current clean, KISS-principle architecture with comprehensive SCP support and security features.