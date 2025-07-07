package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// getTemplateNames returns a comma-separated list of template names
func getTemplateNames(templates map[string]string) string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// getStepStatusIcon returns the appropriate icon for a step status
func getStepStatusIcon(status string) string {
	switch status {
	case parser.StatusPassed:
		return "✅"
	case parser.StatusFailed:
		return "❌"
	case parser.StatusSkipped:
		return "⏭️"
	default:
		return "❓"
	}
}

// getTestStatusIcon returns the appropriate icon for a test status
func getTestStatusIcon(status string) string {
	switch status {
	case parser.StatusPassed:
		return "✅"
	case parser.StatusFailed:
		return "❌"
	case parser.StatusSkipped:
		return "⏭️"
	default:
		return "❓"
	}
}

// PrintStepResultSimple prints a single step result in simple format
func PrintStepResultSimple(stepNum int, stepResult parser.StepResult, indent string) {
	icon := getStepStatusIcon(stepResult.Status)
	stepName := stepResult.Step.Name
	if stepName == "" {
		stepName = "(unnamed)"
	}

	fmt.Printf("%s%s Step %d: %s | Status: %s", indent, icon, stepNum, stepName, stepResult.Status)
	if stepResult.Error != "" {
		fmt.Printf(" | Error: %s", stepResult.Error)
	}
	fmt.Println()
}

// PrintStepResultsSimple prints step results in simple format
func PrintStepResultsSimple(stepResults []parser.StepResult, title string, indent string) {
	if len(stepResults) == 0 {
		return
	}

	fmt.Printf("%s%s\n", indent, title)
	for i, stepResult := range stepResults {
		PrintStepResultSimple(i+1, stepResult, indent+"   ")
	}
}

// PrintStepResultsTable prints step results in detailed table format
func PrintStepResultsTable(stepResults []parser.StepResult, title string) {
	if len(stepResults) == 0 {
		return
	}

	fmt.Printf("\n%s\n", title)
	fmt.Printf("%-4s | %-24s | %-12s | %-6s | %-10s | %-24s | %-24s\n", "#", "Name", "Action", "Status", "Duration", "Output", "Error")
	fmt.Println(strings.Repeat("-", 116))

	for i, stepResult := range stepResults {
		icon := getStepStatusIcon(stepResult.Status)
		output := stepResult.Output
		if len(output) > 24 {
			output = output[:21] + "..."
		}
		error := stepResult.Error
		if len(error) > 24 {
			error = error[:21] + "..."
		}
		stepName := stepResult.Step.Name
		if stepName == "" {
			stepName = "(unnamed)"
		}
		if len(stepName) > 24 {
			stepName = stepName[:21] + "..."
		}

		// Duration formatting: higher precision for <1ms
		var duration string
		if stepResult.Duration < time.Millisecond {
			duration = fmt.Sprintf("%dµs", stepResult.Duration.Microseconds())
		} else {
			duration = stepResult.Duration.String()
		}
		if len(duration) > 10 {
			duration = duration[:7] + "..."
		}

		fmt.Printf("%-4s | %-24s | %-12s | %-6s | %-10s | %-24s | %-24s\n",
			fmt.Sprintf("#%d", i+1),
			stepName,
			stepResult.Step.Action,
			icon,
			duration,
			output,
			error,
		)
	}
}

// PrintStepResultsMarkdown prints step results in markdown table format
func PrintStepResultsMarkdown(stepResults []parser.StepResult, title string) {
	if len(stepResults) == 0 {
		return
	}

	fmt.Printf("\n%s\n", title)
	fmt.Println("| # | Name | Action | Status | Duration | Output | Error |")
	fmt.Println("|---|------|--------|--------|----------|--------|-------|")

	for i, stepResult := range stepResults {
		icon := getStepStatusIcon(stepResult.Status)
		status := stepResult.Status
		output := stepResult.Output
		if len(output) > 30 {
			output = output[:27] + "..."
		}
		errorMsg := stepResult.Error
		if len(errorMsg) > 30 {
			errorMsg = errorMsg[:27] + "..."
		}

		fmt.Printf("| %d | %s | %s | %s %s | %v | %s | %s |\n",
			i+1,
			stepResult.Step.Name,
			stepResult.Step.Action,
			icon,
			status,
			stepResult.Duration.Truncate(1e6), // ms precision
			output,
			errorMsg,
		)
	}
}

// PrintTestSummary prints the test summary (duration, status, steps)
func PrintTestSummary(result *parser.TestResult) {
	fmt.Printf("\n🏁 Test completed in %v\n", result.Duration)
	fmt.Printf("\n📊 Test Results:\n")

	statusIcon := getTestStatusIcon(result.Status)
	fmt.Printf("%s Status: %s\n", statusIcon, result.Status)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📝 Steps: %d total, %d passed, %d failed, %d skipped\n",
		len(result.StepResults), result.PassedSteps, result.FailedSteps, result.SkippedSteps)

	// Print each step result in simple format
	for i, stepResult := range result.StepResults {
		PrintStepResultSimple(i+1, stepResult, "   ")
	}

	// Print detailed table format
	PrintStepResultsTable(result.StepResults, "Step Results:")
}

// PrintTDMSetup prints the TDM setup message
func PrintTDMSetup() {
	fmt.Printf("🔧 Executing TDM setup...\n")
}

// PrintTDMTeardown prints the TDM teardown message
func PrintTDMTeardown() {
	fmt.Printf("🧹 Executing TDM teardown...\n")
}

// PrintDataValidationFailure prints a data validation failure message
func PrintDataValidationFailure(name, message string) {
	fmt.Printf("❌ Data validation failed: %s - %s\n", name, message)
}

// PrintDataValidationWarning prints a data validation warning message
func PrintDataValidationWarning(name, message string) {
	fmt.Printf("⚠️  Data validation warning: %s - %s\n", name, message)
}

// PrintWarning prints a general warning message
func PrintWarning(format string, args ...interface{}) {
	fmt.Printf("⚠️  Warning: "+format+"\n", args...)
}

// PrintStepStart prints the start of a step
func PrintStepStart(stepNum int, stepLabel string) {
	fmt.Printf("Step %d: %s\n", stepNum, stepLabel)
}

// PrintStepSkipped prints a skipped step
func PrintStepSkipped(stepNum int, errMsg string) {
	fmt.Printf("⏭️  Step %d skipped: %s\n", stepNum, errMsg)
}

// PrintStepFailed prints a failed step
func PrintStepFailed(stepNum int, errMsg string) {
	fmt.Printf("❌ Step %d failed: %s\n", stepNum, errMsg)
}

// PrintStepErrorExpectationPassed prints when error expectation passes
func PrintStepErrorExpectationPassed(stepNum int) {
	fmt.Printf("✅ Error expectation passed\n")
}

// PrintStepVerboseOutput prints verbose output
func PrintStepVerboseOutput(output string) {
	fmt.Print(output)
}

// PrintStepLog prints a log message for a step
func PrintStepLog(message string) {
	fmt.Printf("📝 %s\n", message)
}

// PrintStepContinueOnFailure prints continue on failure warning
func PrintStepContinueOnFailure(stepName string) {
	fmt.Printf("⚠️  Step '%s' failed but continuing due to continue_on_failure\n", stepName)
}

// PrintStepResultStored prints when a step result is stored in a variable
func PrintStepResultStored(varName, value string) {
	fmt.Printf("💾 Stored result in variable: %s = %s\n", varName, value)
}

// PrintParallelStepGroups prints parallel step groups execution
func PrintParallelStepGroups(groupCount int) {
	fmt.Printf("📊 Executing %d step groups (parallel execution enabled)\n", groupCount)
}

// PrintParallelSteps prints parallel steps execution
func PrintParallelSteps(stepCount, groupIdx int) {
	fmt.Printf("⚡ Executing %d steps in parallel (group %d)\n", stepCount, groupIdx+1)
}

// PrintTestCaseStart prints test case start
func PrintTestCaseStart(testCaseName string) {
	fmt.Printf("🚀 Running test case: %s\n", testCaseName)
}

// PrintTestCaseDescription prints test case description
func PrintTestCaseDescription(description string) {
	fmt.Printf("📋 Description: %s\n", description)
}

// PrintTestCaseSteps prints test case steps count
func PrintTestCaseSteps(stepCount int) {
	fmt.Printf("📝 Steps: %d\n\n", stepCount)
}

// PrintTemplatesLoaded prints templates loaded
func PrintTemplatesLoaded(templateCount int, templateNames string) {
	fmt.Printf("📄 Loaded %d templates: %s\n", templateCount, templateNames)
}

// PrintParallelFiles prints parallel file execution
func PrintParallelFiles(fileCount, maxConcurrency int) {
	fmt.Printf("🚀 Running %d test files in parallel (max concurrency: %d)\n", fileCount, maxConcurrency)
}
