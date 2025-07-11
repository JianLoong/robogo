package output

import (
	"fmt"
	"os"
	"strings"
	"regexp"

	"github.com/JianLoong/robogo/internal/parser"
)

// ConsoleFormatter handles console output formatting
type ConsoleFormatter struct{}

// maskSecretsInArgs applies secret masking to arguments for display
func maskSecretsInArgs(args []interface{}) []interface{} {
	if len(args) == 0 {
		return args
	}
	
	maskedArgs := make([]interface{}, len(args))
	copy(maskedArgs, args)
	
	// Common secret patterns to mask
	secretPatterns := []*regexp.Regexp{
		// Database URLs with passwords
		regexp.MustCompile(`(postgresql://[^:]+:)([^@]+)(@.*)`),
		regexp.MustCompile(`(postgres://[^:]+:)([^@]+)(@.*)`),
		regexp.MustCompile(`(mysql://[^:]+:)([^@]+)(@.*)`),
		// Generic protocol://user:password@host patterns
		regexp.MustCompile(`(://[^:]+:)([^@]+)(@.*)`),
		// Bearer tokens
		regexp.MustCompile(`(Bearer\s+)([A-Za-z0-9\-._~+/]+=*)`),
		// API keys and tokens (common patterns)
		regexp.MustCompile(`([Aa]pi[_-]?[Kk]ey[=\s:]+)([A-Za-z0-9\-._~+/]{8,})`),
		regexp.MustCompile(`([Tt]oken[=\s:]+)([A-Za-z0-9\-._~+/]{8,})`),
		// Password fields
		regexp.MustCompile(`([Pp]assword[=\s:]+)([^\s&]+)`),
		regexp.MustCompile(`([Pp]ass[=\s:]+)([^\s&]+)`),
	}
	
	for i, arg := range maskedArgs {
		if argStr, ok := arg.(string); ok {
			masked := argStr
			for _, pattern := range secretPatterns {
				masked = pattern.ReplaceAllString(masked, "${1}[MASKED]${3}")
			}
			maskedArgs[i] = masked
		}
	}
	
	return maskedArgs
}

// Utility functions for console output formatting

// GetTemplateNames returns a comma-separated list of template names
func GetTemplateNames(templates map[string]string) string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// getStepStatusIcon returns the appropriate icon for a step status
func getStepStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "passed":
		return "PASSED"
	case "failed":
		return "FAILED"
	case "skipped":
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

// padOrTruncate pads or truncates a string to a fixed width
func padOrTruncate(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PrintStepResultsSimple prints step results in simple format
func PrintStepResultsSimple(stepResults []parser.StepResult, title string, indent string) {
	if len(stepResults) == 0 {
		return
	}

	fmt.Printf("%s%s\n", indent, title)
	for i, stepResult := range stepResults {
		icon := getStepStatusIcon(stepResult.Status)
		stepName := stepResult.DisplayName
		if stepName == "" {
			stepName = stepResult.Step.Name
		}
		if stepName == "" {
			stepName = "(unnamed)"
		}

		fmt.Printf("%s%s Step %d: %s | Status: %s", indent+"   ", icon, i+1, stepName, stepResult.Status)
		if stepResult.Error != "" {
			fmt.Printf(" | Error: %s", stepResult.Error)
		}
		fmt.Println()
	}
}

// PrintStepResultsDetailed prints detailed step results for debugging
func PrintStepResultsDetailed(stepResults []parser.StepResult, title string) {
	if len(stepResults) == 0 {
		return
	}

	fmt.Printf("\n%s\n", title)
	fmt.Printf("%s\n", strings.Repeat("=", len(title)))

	for i, stepResult := range stepResults {
		status := stepResult.Status
		icon := "✅"
		if status == "FAILED" {
			icon = "❌"
		} else if status == "SKIPPED" {
			icon = "⏭️"
		}

		fmt.Printf("\n%s Step %d: %s\n", icon, i+1, stepResult.Step.Name)
		fmt.Printf("   Action: %s\n", stepResult.Step.Action)
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Duration: %v\n", stepResult.Duration)
		
		if stepResult.Step.Args != nil && len(stepResult.Step.Args) > 0 {
			maskedArgs := maskSecretsInArgs(stepResult.Step.Args)
			fmt.Printf("   Args: %v\n", maskedArgs)
		}
		
		if stepResult.Output != "" {
			fmt.Printf("   Output: %s\n", stepResult.Output)
		}
		
		if stepResult.Error != "" {
			fmt.Printf("   Error: %s\n", stepResult.Error)
		}
	}
	fmt.Printf("\n")
}

// PrintStepResultsMarkdown prints step results in markdown table format
func PrintStepResultsMarkdown(stepResults []parser.StepResult, title string) {
	if len(stepResults) == 0 {
		return
	}

	fmt.Printf("\n%s\n", title)
	header := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |",
		padOrTruncate("#", 4),
		padOrTruncate("Name", 24),
		padOrTruncate("Action", 12),
		padOrTruncate("Status", 12),
		padOrTruncate("Dur", 6),
		padOrTruncate("Output", 24),
		padOrTruncate("Error", 24))
	fmt.Println(header)

	separator := fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|",
		strings.Repeat("-", 6),  // 4 for content + 2 for padding
		strings.Repeat("-", 26), // 24 for content + 2 for padding
		strings.Repeat("-", 14), // 12 for content + 2 for padding (action)
		strings.Repeat("-", 14), // 12 for content + 2 for padding (status)
		strings.Repeat("-", 8),  // 6 for content + 2 for padding
		strings.Repeat("-", 26), // 24 for content + 2 for padding (output)
		strings.Repeat("-", 26)) // 24 for content + 2 for padding (error)
	fmt.Println(separator)

	for i, stepResult := range stepResults {
		name := stepResult.DisplayName
		if name == "" {
			name = stepResult.Step.Name
		}
		if name == "" {
			name = "(unnamed)"
		}
		stepNum := padOrTruncate(fmt.Sprintf("%d", i+1), 4)
		name = padOrTruncate(name, 24)
		action := padOrTruncate(stepResult.Step.Action, 12)
		// Create status text to fit in 12 characters with icon (icon + space + up to 10 chars)
		shortStatus := stepResult.Status
		if len(shortStatus) > 10 {
			shortStatus = shortStatus[:10]
		}
		statusWithIcon := padOrTruncate(shortStatus, 12)
		duration := padOrTruncate(stepResult.Duration.Truncate(1e6).String(), 6)
		output := stepResult.Output
		if len(output) > 24 {
			output = output[:21] + "..."
		}
		output = padOrTruncate(output, 24)
		errorMsg := stepResult.Error
		if len(errorMsg) > 24 {
			errorMsg = errorMsg[:21] + "..."
		}
		errorMsg = padOrTruncate(errorMsg, 24)

		fmt.Printf("| %s | %s | %s | %s | %s | %s | %s |\n",
			stepNum,
			name,
			action,
			statusWithIcon,
			duration,
			output,
			errorMsg,
		)
	}
}

// Console output utility functions


// PrintWarning prints a general warning message
func PrintWarning(format string, args ...interface{}) {
	fmt.Printf("Warning: "+format+"\n", args...)
}

// PrintParallelStepGroups prints parallel step groups execution
func PrintParallelStepGroups(groupCount int) {
	fmt.Printf("Executing %d step groups (parallel execution enabled)\n", groupCount)
}

// PrintParallelSteps prints parallel steps execution
func PrintParallelSteps(stepCount, groupIdx int) {
	fmt.Printf("⚡ Executing %d steps in parallel (group %d)\n", stepCount, groupIdx+1)
}

// PrintTestCaseStart prints test case start
func PrintTestCaseStart(testCaseName string) {
	fmt.Printf("Running test case: %s\n", testCaseName)
}

// PrintTestCaseDescription prints test case description
func PrintTestCaseDescription(description string) {
	fmt.Printf("Description: %s\n", description)
}

// PrintTestCaseSteps prints test case steps count
func PrintTestCaseSteps(stepCount int) {
	fmt.Printf("Steps: %d\n\n", stepCount)
}

// PrintTemplatesLoaded prints templates loaded
func PrintTemplatesLoaded(templateCount int, templateNames string) {
	fmt.Printf("Loaded %d templates: %s\n", templateCount, templateNames)
}

// PrintParallelFiles prints parallel file execution
func PrintParallelFiles(fileCount, maxConcurrency int) {
	fmt.Printf("Running %d test files in parallel (max concurrency: %d)\n", fileCount, maxConcurrency)
}

// FormatTestResults outputs test results in console format
func (f *ConsoleFormatter) FormatTestResults(results []*parser.TestResult) error {
	for _, result := range results {
		// Print captured output (step-by-step execution details)
		if result.CapturedOutput != "" {
			fmt.Print(result.CapturedOutput)
		}

		// Print test summary in markdown format
		fmt.Printf("\n## Test Results for: %s\n\n", result.TestCase.Name)

		// Choose appropriate status icon
		statusIcon := "✅"
		if result.Status == "FAILED" {
			statusIcon = "❌"
		} else if result.Status == "SKIPPED" {
			statusIcon = "⏭️"
		}

		fmt.Printf("**%s Status:** %s\n\n", statusIcon, result.Status)
		fmt.Printf("**Duration:** %v\n\n", result.Duration)
		fmt.Printf("**Steps Summary:**\n\n")
		fmt.Printf("| %-6s | %-7s | %-6s | %-7s |\n", "Total", "Passed", "Failed", "Skipped")
		fmt.Printf("|--------|---------|--------|---------|\n")
		fmt.Printf("| %-6d | %-7d | %-6d | %-7d |\n\n",
			result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps)

		// Print step details as a markdown table
		if len(result.StepResults) > 0 {
			// Show detailed step breakdown for debugging
			PrintStepResultsDetailed(result.StepResults, "## Detailed Step Results")
			
			// Also show compact table
			fmt.Println("\nStep Results (Compact Table):")
			PrintStepResultsMarkdown(result.StepResults, "Step Results:")
		}
	}

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedSteps > 0 {
			return fmt.Errorf("test suite failed")
		}
	}
	return nil
}

// FormatSuiteResult outputs test suite results in console format
func (f *ConsoleFormatter) FormatSuiteResult(result *parser.TestSuiteResult) error {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("Test Suite Results: %s\n", result.TestSuite.Name)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("\n## Test Case Summary\n")
	fmt.Printf("| %-4s | %-24s | %-8s | %-10s | %-24s |\n", "#", "Name", "Status", "Duration", "Error")
	fmt.Printf("|------|--------------------------|----------|------------|--------------------------|\n")
	for i, caseResult := range result.CaseResults {
		status := strings.ToUpper(caseResult.Status)
		duration := ""
		if caseResult.Duration > 0 {
			duration = fmt.Sprintf("%.4gs", caseResult.Duration.Seconds())
		}
		error := caseResult.Error
		if len(error) > 24 {
			error = error[:21] + "..."
		}
		name := caseResult.TestCase.Name
		if len(name) > 24 {
			name = name[:21] + "..."
		}
		fmt.Printf("| %-4d | %-24s | %-8s | %-10s | %-24s |\n", i+1, name, status, duration, error)
	}

	// Print step results for each test case
	for _, caseResult := range result.CaseResults {
		if caseResult.Result != nil && len(caseResult.Result.StepResults) > 0 {
			title := "### Step Results for " + caseResult.TestCase.Name
			PrintStepResultsMarkdown(caseResult.Result.StepResults, title)
		}
	}

	// Print step summary table
	fmt.Printf("\n## Step Summary\n")
	fmt.Printf("| %-8s | %-8s | %-8s | %-8s |\n", "Total", "Passed", "Failed", "Skipped")
	fmt.Printf("|----------|----------|----------|----------|\n")
	fmt.Printf("| %-8d | %-8d | %-8d | %-8d |\n", result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps)

	if result.SetupStatus != "" {
		fmt.Printf("\nSetup: %s\n", result.SetupStatus)
	}
	if result.TeardownStatus != "" {
		fmt.Printf("\nTeardown: %s\n", result.TeardownStatus)
	}

	fmt.Printf(strings.Repeat("=", 60) + "\n")

	// Ensure all output is flushed before exit
	os.Stdout.Sync()

	// Exit with non-zero code if test suite failed
	if result.Status == "failed" {
		return fmt.Errorf("test suite failed")
	}
	return nil
}

// FormatMultipleSuites outputs multiple test suite results in console format
func (f *ConsoleFormatter) FormatMultipleSuites(results []*parser.TestSuiteResult, grandTotal GrandTotal) error {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("🎯 GRAND TOTAL SUMMARY\n")
	fmt.Printf("📊 Test Suite Results: Grand Total\n")
	fmt.Printf("⏱️  Duration: %v\n", grandTotal.Duration)
	fmt.Printf("📋 Total Cases: %d\n", grandTotal.TotalCases)
	fmt.Printf("✅ Passed: %d\n", grandTotal.PassedCases)
	fmt.Printf("❌ Failed: %d\n", grandTotal.FailedCases)
	fmt.Printf("⏭️  Skipped: %d\n", grandTotal.SkippedCases)
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	// Exit with non-zero code if any test failed
	if grandTotal.FailedCases > 0 {
		return fmt.Errorf("test suite failed")
	}
	return nil
}