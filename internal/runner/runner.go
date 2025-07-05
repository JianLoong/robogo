package runner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/your-org/robogo/internal/actions"
	"github.com/your-org/robogo/internal/parser"
)

// TestRunner runs test cases
type TestRunner struct {
	variables     map[string]interface{} // Store variables for the test case
	secretManager *actions.SecretManager // Secret manager for handling secrets
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		variables:     make(map[string]interface{}),
		secretManager: actions.NewSecretManager(),
	}
}

// RunTestFile runs a test case from a file
func RunTestFile(filename string, silent bool) (*parser.TestResult, error) {
	// Parse the test case
	testCase, err := parser.ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	// Run the test case
	return RunTestCase(testCase, silent)
}

// RunTestCase runs a single test case
func RunTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	runner := NewTestRunner()
	return runner.runTestCase(testCase, silent)
}

// runTestCase runs a single test case
func (tr *TestRunner) runTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	startTime := time.Now()

	// Initialize variables from test case
	tr.initializeVariables(testCase)

	// Create action executor
	executor := actions.NewActionExecutor()

	// Initialize result
	result := &parser.TestResult{
		TestCase:    *testCase,
		Status:      "PASSED",
		TotalSteps:  len(testCase.Steps),
		StepResults: make([]parser.StepResult, 0, len(testCase.Steps)),
	}

	if !silent {
		fmt.Printf("🚀 Running test case: %s\n", testCase.Name)
		if testCase.Description != "" {
			fmt.Printf("📋 Description: %s\n", testCase.Description)
		}
		fmt.Printf("📝 Steps: %d\n\n", len(testCase.Steps))
	}

	// Execute each step
	for i, step := range testCase.Steps {
		stepStart := time.Now()
		stepLabel := step.Name
		if stepLabel == "" {
			stepLabel = step.Action
		}
		if !silent {
			fmt.Printf("Step %d: %s\n", i+1, stepLabel)
		}

		// Handle control flow structures
		if step.If != nil {
			if err := tr.executeIfStatement(step.If, executor, silent); err != nil {
				result.Status = "FAILED"
				result.FailedSteps++
				result.ErrorMessage = err.Error()
				if !silent {
					fmt.Printf("❌ Step %d (if statement) failed: %s\n", i+1, err.Error())
				}
			} else {
				result.PassedSteps++
				if !silent {
					fmt.Printf("✅ Step %d (if statement) completed in %v\n", i+1, time.Since(stepStart))
				}
			}
			continue
		}

		if step.For != nil {
			if err := tr.executeForLoop(step.For, executor, silent); err != nil {
				result.Status = "FAILED"
				result.FailedSteps++
				result.ErrorMessage = err.Error()
				if !silent {
					fmt.Printf("❌ Step %d (for loop) failed: %s\n", i+1, err.Error())
				}
			} else {
				result.PassedSteps++
				if !silent {
					fmt.Printf("✅ Step %d (for loop) completed in %v\n", i+1, time.Since(stepStart))
				}
			}
			continue
		}

		if step.While != nil {
			if err := tr.executeWhileLoop(step.While, executor, silent); err != nil {
				result.Status = "FAILED"
				result.FailedSteps++
				result.ErrorMessage = err.Error()
				if !silent {
					fmt.Printf("❌ Step %d (while loop) failed: %s\n", i+1, err.Error())
				}
			} else {
				result.PassedSteps++
				if !silent {
					fmt.Printf("✅ Step %d (while loop) completed in %v\n", i+1, time.Since(stepStart))
				}
			}
			continue
		}

		// Execute regular step with retry support
		substitutedArgs := tr.substituteVariables(step.Args)
		output, err := tr.executeStepWithRetry(step, substitutedArgs, executor, silent)
		stepDuration := time.Since(stepStart)

		// Get verbosity level for this step
		verbosityLevel := parser.GetVerbosityLevel(&step, testCase)

		// Mask secrets in output for display
		maskedOutput := tr.secretManager.MaskSecretsInString(output)

		// Format verbose output if enabled
		verboseOutput := parser.FormatVerboseOutput(verbosityLevel, step.Action, substitutedArgs, maskedOutput, stepDuration.String())

		stepResult := parser.StepResult{
			Step:      step,
			Status:    "PASSED",
			Duration:  stepDuration,
			Output:    maskedOutput, // Use masked output for display
			Timestamp: time.Now(),
		}

		if err != nil {
			stepResult.Status = "FAILED"
			stepResult.Error = err.Error()
			result.Status = "FAILED"
			result.FailedSteps++
			result.ErrorMessage = err.Error()

			if !silent {
				fmt.Printf("❌ Step %d failed: %s\n", i+1, err.Error())
			}
		} else {
			result.PassedSteps++
			if !silent {
				// Display verbose output if enabled
				if verboseOutput != "" {
					fmt.Print(verboseOutput)
				} else {
					// For log actions, mask the message before displaying
					if step.Action == "log" && len(step.Args) > 0 {
						message := fmt.Sprintf("%v", step.Args[0])
						maskedMessage := tr.secretManager.MaskSecretsInString(message)
						fmt.Printf("📝 %s\n", maskedMessage)
					} else {
						fmt.Printf("✅ Step %d completed in %v\n", i+1, stepDuration)
					}
				}
			}
		}

		// Store result in variable if specified (store actual value, not masked)
		if step.Result != "" {
			tr.variables[step.Result] = output // Store actual output
			if !silent {
				// Display masked version
				maskedValue := tr.secretManager.MaskSecretsInString(fmt.Sprintf("%v", output))
				fmt.Printf("💾 Stored result in variable: %s = %s\n", step.Result, maskedValue)
			}
		}

		result.StepResults = append(result.StepResults, stepResult)
	}

	result.Duration = time.Since(startTime)

	if !silent {
		fmt.Printf("\n🏁 Test completed in %v\n", result.Duration)
		fmt.Printf("\n📊 Test Results:\n")
		fmt.Printf("✅ Status: %s\n", result.Status)
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		fmt.Printf("📝 Steps: %d total, %d passed, %d failed\n", result.TotalSteps, result.PassedSteps, result.FailedSteps)
	}

	return result, nil
}

// initializeVariables initializes variables from the test case
func (tr *TestRunner) initializeVariables(testCase *parser.TestCase) {
	// Initialize regular variables
	if testCase.Variables.Regular != nil {
		for name, value := range testCase.Variables.Regular {
			tr.variables[name] = value
		}
	}

	// Initialize secret variables (inline or file)
	if testCase.Variables.Secrets != nil {
		secretMap := make(map[string]interface{})
		for name, secret := range testCase.Variables.Secrets {
			secretMap[name] = map[string]interface{}{
				"value":       secret.Value,
				"file":        secret.File,
				"mask_output": secret.MaskOutput,
			}
		}

		// Resolve secrets (inline or file)
		if err := tr.secretManager.ResolveSecrets(secretMap); err != nil {
			fmt.Printf("⚠️  Warning: Failed to resolve secrets: %v\n", err)
		}

		// Add resolved secrets to variables
		for name := range testCase.Variables.Secrets {
			if value, exists := tr.secretManager.GetSecret(name); exists {
				tr.variables[name] = value
			}
		}
	}
}

// executeIfStatement executes an if/else block
func (tr *TestRunner) executeIfStatement(ifBlock *parser.ConditionalBlock, executor *actions.ActionExecutor, silent bool) error {
	// Evaluate condition
	condition := tr.substituteString(ifBlock.Condition)
	output, err := executor.Execute("control", []interface{}{"if", condition})
	if err != nil {
		return fmt.Errorf("failed to evaluate if condition: %w", err)
	}

	// Determine which branch to execute
	var stepsToExecute []parser.Step
	if output == "true" {
		stepsToExecute = ifBlock.Then
		if !silent {
			fmt.Printf("🔍 If condition '%s' is true, executing 'then' branch (%d steps)\n", condition, len(stepsToExecute))
		}
	} else {
		stepsToExecute = ifBlock.Else
		if !silent {
			fmt.Printf("🔍 If condition '%s' is false, executing 'else' branch (%d steps)\n", condition, len(stepsToExecute))
		}
	}

	// Execute the selected branch
	return tr.executeSteps(stepsToExecute, executor, silent)
}

// executeForLoop executes a for loop
func (tr *TestRunner) executeForLoop(forBlock *parser.LoopBlock, executor *actions.ActionExecutor, silent bool) error {
	// Evaluate loop condition to get iteration count
	condition := tr.substituteString(forBlock.Condition)
	output, err := executor.Execute("control", []interface{}{"for", condition})
	if err != nil {
		return fmt.Errorf("failed to evaluate for loop condition: %w", err)
	}

	iterations, err := strconv.Atoi(output)
	if err != nil {
		return fmt.Errorf("failed to parse iteration count: %w", err)
	}

	maxIterations := forBlock.MaxIterations
	if maxIterations > 0 && iterations > maxIterations {
		iterations = maxIterations
		if !silent {
			fmt.Printf("⚠️  Limiting iterations to %d (max_iterations)\n", maxIterations)
		}
	}

	if !silent {
		fmt.Printf("🔄 For loop: executing %d iterations\n", iterations)
	}

	// Execute the loop body for each iteration
	for i := 0; i < iterations; i++ {
		if !silent {
			fmt.Printf("  🔄 Iteration %d/%d\n", i+1, iterations)
		}

		// Set iteration variable
		tr.variables["iteration"] = i + 1
		tr.variables["index"] = i

		if err := tr.executeSteps(forBlock.Steps, executor, silent); err != nil {
			return fmt.Errorf("iteration %d failed: %w", i+1, err)
		}
	}

	return nil
}

// executeWhileLoop executes a while loop
func (tr *TestRunner) executeWhileLoop(whileBlock *parser.LoopBlock, executor *actions.ActionExecutor, silent bool) error {
	iteration := 0
	maxIterations := whileBlock.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 1000 // Default max iterations to prevent infinite loops
	}

	for {
		iteration++
		if iteration > maxIterations {
			return fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
		}

		// Evaluate condition
		condition := tr.substituteString(whileBlock.Condition)
		output, err := executor.Execute("control", []interface{}{"while", condition})
		if err != nil {
			return fmt.Errorf("failed to evaluate while condition: %w", err)
		}

		if output != "true" {
			if !silent {
				fmt.Printf("🔄 While loop: condition '%s' is false, exiting after %d iterations\n", condition, iteration-1)
			}
			break
		}

		if !silent {
			fmt.Printf("  🔄 While iteration %d\n", iteration)
		}

		// Set iteration variable
		tr.variables["iteration"] = iteration

		// Execute loop body
		if err := tr.executeSteps(whileBlock.Steps, executor, silent); err != nil {
			return fmt.Errorf("while iteration %d failed: %w", iteration, err)
		}
	}

	return nil
}

// executeSteps executes a slice of steps
func (tr *TestRunner) executeSteps(steps []parser.Step, executor *actions.ActionExecutor, silent bool) error {
	for _, step := range steps {
		// Handle nested control flow
		if step.If != nil {
			if err := tr.executeIfStatement(step.If, executor, silent); err != nil {
				return err
			}
			continue
		}

		if step.For != nil {
			if err := tr.executeForLoop(step.For, executor, silent); err != nil {
				return err
			}
			continue
		}

		if step.While != nil {
			if err := tr.executeWhileLoop(step.While, executor, silent); err != nil {
				return err
			}
			continue
		}

		// Execute regular step with retry support
		substitutedArgs := tr.substituteVariables(step.Args)
		output, err := tr.executeStepWithRetry(step, substitutedArgs, executor, true) // Silent for nested steps
		if err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}

		// Store result in variable if specified
		if step.Result != "" {
			tr.variables[step.Result] = output
		}
	}

	return nil
}

// substituteVariables replaces ${variable} references with actual values
func (tr *TestRunner) substituteVariables(args []interface{}) []interface{} {
	substituted := make([]interface{}, len(args))

	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			substituted[i] = tr.substituteString(v)
		case []interface{}:
			substituted[i] = tr.substituteVariables(v)
		case map[string]interface{}:
			substituted[i] = tr.substituteMap(v)
		default:
			substituted[i] = arg
		}
	}

	return substituted
}

// substituteString replaces ${variable} references in a string
func (tr *TestRunner) substituteString(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value, exists := tr.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		return match // Return original if variable not found
	})
}

// substituteStringForDisplay replaces ${variable} references and masks secrets for display
func (tr *TestRunner) substituteStringForDisplay(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value, exists := tr.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		return match // Return original if variable not found
	})

	// Mask secrets in the result for display only
	return tr.secretManager.MaskSecretsInString(result)
}

// substituteMap substitutes variables in map values
func (tr *TestRunner) substituteMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = tr.substituteString(val)
		case []interface{}:
			result[k] = tr.substituteVariables(val)
		case map[string]interface{}:
			result[k] = tr.substituteMap(val)
		default:
			result[k] = v
		}
	}
	return result
}

// executeStepWithRetry executes a step with retry logic if configured
func (tr *TestRunner) executeStepWithRetry(step parser.Step, args []interface{}, executor *actions.ActionExecutor, silent bool) (string, error) {
	// If no retry configuration, execute normally
	if step.Retry == nil {
		return executor.Execute(step.Action, args)
	}

	// Validate retry configuration
	if err := parser.ValidateRetryConfig(step.Retry); err != nil {
		return "", fmt.Errorf("invalid retry configuration: %w", err)
	}

	// Merge with defaults
	retryConfig := parser.MergeRetryConfig(step.Retry)

	// Initialize retry result
	retryResult := parser.RetryResult{
		RetryLogs: make([]string, 0),
	}

	startTime := time.Now()
	var lastError error
	var lastOutput string

	// Execute with retries
	for attempt := 1; attempt <= retryConfig.Attempts; attempt++ {
		// Execute the action
		output, err := executor.Execute(step.Action, args)
		lastOutput = output

		// If successful, return immediately
		if err == nil {
			retryResult.Success = true
			retryResult.Attempts = attempt
			retryResult.TotalTime = time.Since(startTime)

			if attempt > 1 && !silent {
				fmt.Printf("✅ %s\n", parser.FormatRetrySummary(retryResult, false))
			}
			return output, nil
		}

		lastError = err

		// Check if we should retry based on conditions
		if !parser.ShouldRetry(err, output, retryConfig.Conditions) {
			break
		}

		// If this is the last attempt, don't retry
		if attempt == retryConfig.Attempts {
			break
		}

		// Calculate delay for next attempt
		delay := parser.CalculateDelay(retryConfig.Delay, attempt, retryConfig.Backoff, retryConfig.MaxDelay, retryConfig.Jitter)

		// Log retry attempt
		retryLog := parser.FormatRetryLog(attempt, retryConfig.Attempts, delay, err, false)
		retryResult.RetryLogs = append(retryResult.RetryLogs, retryLog)

		if !silent {
			fmt.Printf("%s\n", retryLog)
		}

		// Wait before next attempt
		time.Sleep(delay)
	}

	// All attempts failed
	retryResult.Success = false
	retryResult.Attempts = retryConfig.Attempts
	retryResult.TotalTime = time.Since(startTime)
	retryResult.LastError = lastError
	retryResult.LastOutput = lastOutput

	if !silent {
		fmt.Printf("❌ %s\n", parser.FormatRetrySummary(retryResult, false))
	}

	return lastOutput, lastError
}
