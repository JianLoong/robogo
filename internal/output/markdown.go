package output

import (
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/parser"
)

// MarkdownFormatter handles Markdown output formatting
type MarkdownFormatter struct{}

// pad pads or truncates a string to a specific width
func pad(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}

// FormatTestResults outputs test results in Markdown format
func (f *MarkdownFormatter) FormatTestResults(results []*parser.TestResult) error {
	for _, result := range results {
		statusIcon := "✅"
		if result.Status == "FAILED" {
			statusIcon = "❌"
		}

		markdown := fmt.Sprintf("# Test Results: %s\n\n## Summary\n%s **Status:** %s  \n⏱️ **Duration:** %v  \n📝 **Steps:** %d total, %d passed, %d failed\n\n## Test Case Details\n- **Name:** %s\n- **Description:** %s\n",
			result.TestCase.Name,
			statusIcon,
			result.Status,
			result.Duration,
			result.TotalSteps,
			result.PassedSteps,
			result.FailedSteps,
			result.TestCase.Name,
			result.TestCase.Description)

		// Add Failed Steps section if any failed
		if result.FailedSteps > 0 {
			markdown += "\n## Failed Steps\n"
			markdown += "| #   | Name                     | Action       | Error                   |\n"
			markdown += "|-----|--------------------------|--------------|-------------------------|\n"
			for i, stepResult := range result.StepResults {
				if stepResult.Status == "FAILED" {
					stepName := stepResult.Step.Name
					if stepName == "" {
						stepName = "(unnamed)"
					}
					error := stepResult.Error
					if len(error) > 24 {
						error = error[:21] + "..."
					}
					markdown += fmt.Sprintf("| %s | %s | %s | %s |\n",
						pad(fmt.Sprintf("%d", i+1), 4),
						pad(stepName, 24),
						pad(stepResult.Step.Action, 12),
						pad(error, 24),
					)
				}
			}
		}

		markdown += "\n## Step Results\n"
		// Add markdown table header
		markdown += "| Step | Name                     | Action       | Status | Duration   | Output                  | Error                   |\n"
		markdown += "|------|--------------------------|-------------|--------|-----------|------------------------|--------------------------|\n"

		// Add step details as table rows
		for i, stepResult := range result.StepResults {
			stepIcon := "✅"
			if stepResult.Status == "FAILED" {
				stepIcon = "❌"
			}
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
			// Duration formatting: higher precision for <1ms
			var duration string
			if stepResult.Duration < 1000000 { // 1ms in nanoseconds
				duration = fmt.Sprintf("%dµs", stepResult.Duration.Microseconds())
			} else {
				duration = stepResult.Duration.String()
			}
			if len(duration) > 10 {
				duration = duration[:7] + "..."
			}
			markdown += fmt.Sprintf("| %s | %s | %s | %s   | %s | %s | %s |\n",
				pad(fmt.Sprintf("%d", i+1), 4),
				pad(stepName, 24),
				pad(stepResult.Step.Action, 12),
				pad(stepIcon, 6),
				pad(duration, 10),
				pad(output, 24),
				pad(error, 24),
			)
		}

		// Add error message if test failed
		if result.ErrorMessage != "" {
			markdown += fmt.Sprintf("\n## Error\n❌ %s\n", result.ErrorMessage)
		}

		fmt.Print(markdown)
	}

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedSteps > 0 {
			return fmt.Errorf("test suite failed")
		}
	}
	return nil
}

// FormatSuiteResult outputs test suite results in Markdown format
func (f *MarkdownFormatter) FormatSuiteResult(result *parser.TestSuiteResult) error {
	// Step summary calculation
	totalSteps, passedSteps, failedSteps, skippedSteps := 0, 0, 0, 0
	for _, caseResult := range result.CaseResults {
		if caseResult.Result != nil {
			for _, step := range caseResult.Result.StepResults {
				totalSteps++
				switch step.Status {
				case "PASSED", "passed":
					passedSteps++
				case "FAILED", "failed":
					failedSteps++
				case "SKIPPED", "skipped":
					skippedSteps++
				}
			}
		}
	}

	statusIcon := "✅"
	if result.Status == "failed" {
		statusIcon = "❌"
	}

	markdown := fmt.Sprintf("# Test Suite Results: %s\n\n## Summary\n%s **Status:** %s  \n⏱️ **Duration:** %v  \n📋 **Cases:** %d total, %d passed, %d failed, %d skipped  \n📝 **Steps:** %d total, %d passed, %d failed, %d skipped\n\n",
		result.TestSuite.Name,
		statusIcon,
		result.Status,
		result.Duration,
		result.TotalCases,
		result.PassedCases,
		result.FailedCases,
		result.SkippedCases,
		totalSteps,
		passedSteps,
		failedSteps,
		skippedSteps)

	if result.SetupStatus != "" {
		markdown += fmt.Sprintf("🔧 **Setup:** %s  \n", result.SetupStatus)
	}
	if result.TeardownStatus != "" {
		markdown += fmt.Sprintf("🧹 **Teardown:** %s  \n", result.TeardownStatus)
	}

	// Add test case results
	markdown += "\n## Test Case Results\n"
	markdown += "| # | Name | Status | Duration | Error |\n"
	markdown += "|---|------|--------|----------|-------|\n"

	for i, caseResult := range result.CaseResults {
		caseIcon := "✅"
		if caseResult.Status != "passed" {
			caseIcon = "❌"
		}
		dur := caseResult.Duration
		if dur == 0 && caseResult.Result != nil {
			dur = caseResult.Result.Duration
		}
		err := caseResult.Error
		if len(err) > 60 {
			err = err[:57] + "..."
		}
		markdown += fmt.Sprintf("| %d | %s | %s | %v | %s |\n",
			i+1,
			caseResult.TestCase.Name,
			caseIcon,
			dur,
			err,
		)

		// Step-level details
		if caseResult.Result != nil && len(caseResult.Result.StepResults) > 0 {
			markdown += "\n<details><summary>Steps</summary>\n\n"
			markdown += "| # | Name | Status | Duration | Output | Error |\n"
			markdown += "|---|------|--------|----------|-----------|-------|\n"
			for j, step := range caseResult.Result.StepResults {
				stepStatus := step.Status
				if stepStatus == "" && step.Error != "" {
					stepStatus = "FAILED"
				}
				errStr := step.Error
				if len(errStr) > 60 {
					errStr = errStr[:57] + "..."
				}
				outputStr := step.Output
				if len(outputStr) > 40 {
					outputStr = outputStr[:37] + "..."
				}
				nameStr := step.Step.Name
				if len(nameStr) > 24 {
					nameStr = nameStr[:21] + "..."
				}
				markdown += fmt.Sprintf("| %d | %s | %s | %v | %s | %s |\n",
					j+1,
					nameStr,
					stepStatus,
					step.Duration,
					outputStr,
					errStr,
				)
			}
			markdown += "\n</details>\n"
		}
	}

	// Add error message if suite failed
	if result.ErrorMessage != "" {
		markdown += fmt.Sprintf("\n## Error\n❌ %s\n", result.ErrorMessage)
	}

	fmt.Print(markdown)

	// Exit with non-zero code if test suite failed
	if result.Status == "failed" {
		return fmt.Errorf("test suite failed")
	}
	return nil
}

// FormatMultipleSuites outputs multiple test suite results in Markdown format
func (f *MarkdownFormatter) FormatMultipleSuites(results []*parser.TestSuiteResult, grandTotal GrandTotal) error {
	markdown := fmt.Sprintf("# 🎯 GRAND TOTAL SUMMARY\n\n## Test Suite Results: Grand Total\n\n### Summary\n⏱️ **Duration:** %v  \n📋 **Cases:** %d total, %d passed, %d failed, %d skipped\n\n",
		grandTotal.Duration,
		grandTotal.TotalCases,
		grandTotal.PassedCases,
		grandTotal.FailedCases,
		grandTotal.SkippedCases)

	// Add test suite results
	markdown += "\n## Test Suite Results\n"
	markdown += "| # | Name | Status | Duration | Error |\n"
	markdown += "|---|------|--------|----------|-------|\n"

	for i, result := range results {
		caseIcon := "✅"
		if result.Status != "passed" {
			caseIcon = "❌"
		}
		error := result.ErrorMessage
		if len(error) > 60 {
			error = error[:57] + "..."
		}
		markdown += fmt.Sprintf("| %d | %s | %s | %v | %s |\n",
			i+1,
			result.TestSuite.Name,
			caseIcon,
			result.Duration,
			error,
		)
	}

	// Add error message if any suite failed
	if grandTotal.FailedCases > 0 {
		markdown += fmt.Sprintf("\n## Error\n❌ %d test suites failed\n", grandTotal.FailedCases)
	}

	fmt.Print(markdown)

	// Exit with non-zero code if any test failed
	if grandTotal.FailedCases > 0 {
		return fmt.Errorf("test suite failed")
	}
	return nil
}