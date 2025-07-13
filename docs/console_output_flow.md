# Console Output Flow in Robogo Test Execution

## Overview Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           TEST CASE EXECUTION FLOW                          │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Command   │───▶│  Test Execution │───▶│  Step Execution │───▶│   Action Level  │
│                 │    │     Service     │    │     Service     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │                       │
         ▼                       ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Parse Args &   │    │  Initialize     │    │  Execute Step   │    │  Execute Action │
│  Setup Context  │    │  Test Case      │    │  with Retry &   │    │  (http, log,    │
│                 │    │  Variables      │    │  Error Handling │    │   assert, etc.) │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Detailed Console Output Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CONSOLE OUTPUT FLOW                            │
└─────────────────────────────────────────────────────────────────────────────┘

1. CLI LEVEL (cmd/robogo/main.go)
   ┌─────────────────────────────────────────────────────────────────────────┐
   │ go run cmd/robogo/main.go run test.yaml                                │
   │                                                                         │
   │ Output: None (silent startup)                                          │
   └─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
2. TEST EXECUTION SERVICE (internal/runner/test_execution_service.go)
   ┌─────────────────────────────────────────────────────────────────────────┐
   │ ExecuteTestCase()                                                       │
   │                                                                         │
   │ 2.1 Start Output Capture with Real-time Display                        │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ if outputCapture, ok := tes.context.Output().(*OutputCapture) │ │
   │    │     outputCapture.StartWithRealTimeDisplay(!silent)           │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 2.2 Print Test Case Info (if !silent)                                  │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ Running test case: Test Name                                    │ │
   │    │ Description: Test description                                   │ │
   │    │ Steps: 5                                                        │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 2.3 Execute Steps (calls StepExecutionService)                        │
   │                                                                         │
   │ 2.4 Capture Final Output                                              │
   │    result.CapturedOutput = tes.context.Output().StopCapture()         │
   └─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
3. STEP EXECUTION SERVICE (internal/runner/step_execution_service.go)
   ┌─────────────────────────────────────────────────────────────────────────┐
   │ ExecuteSteps()                                                          │
   │                                                                         │
   │ For each step:                                                          │
   │ 3.1 Preprocess Step (variable substitution)                            │
   │                                                                         │
   │ 3.2 Execute Step with Enhanced Error Handling                          │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ executeStepWithEnhancedErrorHandling()                         │ │
   │    │ - Check retry configuration                                    │ │
   │    │ - Check recovery configuration                                 │ │
   │    │ - Execute with retry logic                                     │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 3.3 Handle Step Results                                               │
   │    - Store step result in variables if specified                     │
   │    - Handle continue_on_failure logic                                │
   └─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
4. ACTION EXECUTION (internal/actions/*.go)
   ┌─────────────────────────────────────────────────────────────────────────┐
   │ Actions.Execute()                                                       │
   │                                                                         │
   │ 4.1 HTTP Action (internal/actions/http.go)                             │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ if !silent {                                                    │ │
   │    │     fmt.Printf("%s %s %s %d (%v)\n", method, url, statusIcon,   │ │
   │    │                resp.StatusCode, duration)                       │ │
   │    │     if len(respBody) > 0 {                                      │ │
   │    │         fmt.Printf("Response body: %s\n", string(respBody))     │ │
   │    │     }                                                           │ │
   │    │ }                                                               │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 4.2 Log Action (internal/actions/log.go)                              │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ if !silent {                                                    │ │
   │    │     fmt.Printf("%s\n", message)                                │ │
   │    │ }                                                               │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 4.3 Assert Action (internal/actions/assert.go)                        │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ if !result {                                                    │ │
   │    │     if !silent {                                                │ │
   │    │         fmt.Printf("Failed: %s\n", fullMsg)                    │ │
   │    │     }                                                           │ │
   │    │ } else {                                                        │ │
   │    │     if !silent {                                                │ │
   │    │         fmt.Printf("%s\n", msg)                                │ │
   │    │     }                                                           │ │
   │    │ }                                                               │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 4.4 Retry Messages (internal/runner/retry_manager.go)                  │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ if !silent && attempt > 1 {                                     │ │
   │    │     fmt.Printf("Retry attempt %d/%d for step '%s'\n",          │ │
   │    │                attempt, attempts, step.Name)                   │ │
   │    │ }                                                               │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   └─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
5. OUTPUT CAPTURE SYSTEM (internal/runner/output_capture.go)
   ┌─────────────────────────────────────────────────────────────────────────┐
   │ Real-time Display Mode (when !silent)                                  │
   │                                                                         │
   │ 5.1 StartWithRealTimeDisplay(true)                                     │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ Create temporary file as new stdout                            │ │
   │    │ Start goroutine: tempFile → MultiWriter(pipe, original_stdout) │ │
   │    │ Redirect os.Stdout = tempFile                                   │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 5.2 During Execution                                                   │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ All fmt.Printf() calls → tempFile → MultiWriter →              │ │
   │    │ - pipe (for capture)                                            │ │
   │    │ - original_stdout (for real-time display)                       │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 5.3 StopCapture()                                                      │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ Close tempFile → Wait for goroutine → Read from pipe →         │ │
   │    │ Restore original stdout → Return captured output               │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   └─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
6. FINAL OUTPUT FORMATTING (internal/output/console.go)
   ┌─────────────────────────────────────────────────────────────────────────┐
   │ ConsoleFormatter.FormatTestResults()                                   │
   │                                                                         │
   │ 6.1 Print Captured Output                                              │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ if result.CapturedOutput != "" {                               │ │
   │    │     fmt.Print(result.CapturedOutput)                           │ │
   │    │ }                                                               │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 6.2 Print Test Summary                                                 │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ ## Test Results for: Test Name                                  │ │
   │    │ **✅ Status:** PASSED                                           │ │
   │    │ **Duration:** 1.2s                                              │ │
   │    │ **Steps Summary:**                                              │ │
   │    │ | Total | Passed | Failed | Skipped |                          │ │
   │    │ |-------|--------|--------|---------|                          │ │
   │    │ | 5     | 5      | 0      | 0       |                          │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 6.3 Print Detailed Step Results                                       │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ ## Detailed Step Results                                        │ │
   │    │ ✅ Step 1: HTTP Request                                         │ │
   │    │    Action: http                                                 │ │
   │    │    Status: PASSED                                               │ │
   │    │    Duration: 1.2s                                               │ │
   │    │    Args: ["GET", "https://api.example.com"]                    │ │
   │    │    Output: {"status_code": 200, "body": "..."}                 │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   │                                                                         │
   │ 6.4 Print Compact Table                                               │
   │    ┌─────────────────────────────────────────────────────────────────┐ │
   │    │ Step Results (Compact Table):                                   │ │
   │    │ | # | Name | Action | Status | Dur | Output | Error |          │ │
   │    │ |---|------|--------|--------|-----|--------|-------|          │ │
   │    │ | 1 | HTTP | http   | PASSED | 1.2s| {...}  |       |          │ │
   │    └─────────────────────────────────────────────────────────────────┘ │
   └─────────────────────────────────────────────────────────────────────────┘
```

## Timing Diagram

```
Timeline: 0ms    100ms   200ms   300ms   400ms   500ms   600ms   700ms   800ms
         │        │       │       │       │       │       │       │       │
         ▼        ▼       ▼       ▼       ▼       ▼       ▼       ▼       ▼
CLI:     [Start]  │       │       │       │       │       │       │       │
                  │       │       │       │       │       │       │       │
Test:             [Init]  │       │       │       │       │       │       │
                         │       │       │       │       │       │       │
Output:                   [Start Capture + Real-time]    │       │       │
                                                         │       │       │
Console:                  [Test Case Info]              │       │       │
                         Running test case: API Test    │       │       │
                         Description: Test API calls    │       │       │
                         Steps: 3                       │       │       │
                                                         │       │       │
Step 1:                          [Execute]              │       │       │
                                                         │       │       │
Action:                          [HTTP Request]         │       │       │
                                                         │       │       │
Console:                          [Real-time Output]    │       │       │
                                 GET https://api... ✅ 200 (1.2s)       │
                                 Response body: {...}   │       │       │
                                                         │       │       │
Step 2:                                                  [Execute]       │
                                                                         │
Action:                                                  [Log]           │
                                                                         │
Console:                                                  [Real-time]    │
                                                                 API call successful
                                                                         │
Step 3:                                                          [Execute]
                                                                         
Action:                                                          [Assert]
                                                                         
Console:                                                          [Real-time]
                                                                 Success: Status is 200

Final:                                                                   [Stop Capture]
                                                                         
Console:                                                                 [Final Summary]
                                                                         
                                                                         ## Test Results...
                                                                         **✅ Status:** PASSED
                                                                         **Duration:** 1.5s
                                                                         
                                                                         ## Detailed Step Results...
```

## Key Points

### 1. **Real-time Output Generation**
- **When**: During step execution, as each action runs
- **How**: Through `fmt.Printf()` calls in action implementations
- **Where**: In action files (`internal/actions/*.go`)

### 2. **Output Capture System**
- **When**: Started before test execution, stopped after completion
- **How**: Redirects `os.Stdout` to capture all output
- **Real-time**: Uses tee-like approach to display AND capture simultaneously

### 3. **Silent Mode Control**
- **Real-time mode**: `silent=false` → See output as it happens
- **Capture-only mode**: `silent=true` → Output only in final summary
- **Parallel execution**: Always uses `silent=true` to avoid deadlocks

### 4. **Output Types**
- **Immediate**: HTTP status, log messages, assertion results
- **Retry messages**: Retry attempts and delays
- **Error messages**: Failed assertions, network errors
- **Final summary**: Test results, step details, statistics

### 5. **Flow Control**
- **CLI** → **Test Service** → **Step Service** → **Actions** → **Console**
- Each level can generate output
- Output capture happens at test service level
- Real-time display controlled by `silent` parameter

This architecture ensures that you get immediate feedback during test execution while still capturing all output for comprehensive reporting. 