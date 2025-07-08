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

// getTestStatusIcon returns the appropriate icon for a test status
func getTestStatusIcon(status string) string {
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

// padOrTruncate pads or truncates a string to a fixed width
func padOrTruncate(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
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

// PrintTDMSetup prints the TDM setup message
func PrintTDMSetup() {
	fmt.Printf("Executing TDM setup...\n")
}

// PrintTDMTeardown prints the TDM teardown message
func PrintTDMTeardown() {
	fmt.Printf("Executing TDM teardown...\n")
}

// PrintDataValidationFailure prints a data validation failure message
func PrintDataValidationFailure(name, message string) {
	fmt.Printf("Data validation failed: %s - %s\n", name, message)
}

// PrintDataValidationWarning prints a data validation warning message
func PrintDataValidationWarning(name, message string) {
	fmt.Printf("Data validation warning: %s - %s\n", name, message)
}

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
