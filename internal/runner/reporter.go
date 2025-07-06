package runner

import (
	"fmt"
	"strings"

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

// PrintTestSummary prints the test summary (duration, status, steps)
func PrintTestSummary(result *parser.TestResult) {
	fmt.Printf("\n🏁 Test completed in %v\n", result.Duration)
	fmt.Printf("\n📊 Test Results:\n")
	fmt.Printf("✅ Status: %s\n", result.Status)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📝 Steps: %d total, %d passed, %d failed, %d skipped\n", len(result.StepResults), result.PassedSteps, result.FailedSteps, result.SkippedSteps)

	// Print each step result, including skipped
	for i, stepResult := range result.StepResults {
		icon := ""
		switch stepResult.Status {
		case "PASSED":
			icon = "✅"
		case "FAILED":
			icon = "❌"
		case "SKIPPED":
			icon = "⏭️"
		}
		fmt.Printf("   %s Step %d: %s | Status: %s", icon, i+1, stepResult.Step.Name, stepResult.Status)
		if stepResult.Error != "" {
			fmt.Printf(" | Reason: %s", stepResult.Error)
		}
		fmt.Println()
	}
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
