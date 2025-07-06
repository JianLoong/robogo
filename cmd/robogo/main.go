package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/runner"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"

	// CLI flags
	outputFormat    string
	parallelEnabled bool
	maxConcurrency  int
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "robogo",
		Short: "Robogo - A modern, git-driven test automation framework",
		Long: `Robogo is a modern, git-driven test automation framework written in Go.
It provides fast, extensible, and developer-friendly test automation with YAML-based test cases.

Key Features:
- Dynamic variable management with the 'variable' action (set_variable, get_variable, list_variables)
- Secure secret management: supports inline and file-based secrets (single value per file)
- PostgreSQL database actions (connect, query, execute, close)
- Control flow: if, for, while loops
- Secret masking in output for security
- URL encoding for PostgreSQL connection strings
- Comprehensive test coverage
`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
	}

	var runCmd = &cobra.Command{
		Use:   "run [test-files...]",
		Short: "Run one or more test case files",
		Long:  `Run one or more test case files. You can specify multiple files or a directory.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine silent mode
			silent := false
			switch outputFormat {
			case "json", "markdown":
				silent = true
			}

			// Create parallelism configuration
			var parallelConfig *parser.ParallelConfig
			if parallelEnabled {
				parallelConfig = &parser.ParallelConfig{
					Enabled:        true,
					MaxConcurrency: maxConcurrency,
					TestCases:      true,
					Steps:          true,
					HTTPRequests:   true,
				}
			}

			// Run the tests with parallelism configuration
			results, err := runner.RunTestFilesWithConfig(args, silent, parallelConfig)
			if err != nil {
				return fmt.Errorf("failed to run tests: %w", err)
			}

			// Output results in specified format
			switch outputFormat {
			case "json":
				return outputJSON(results)
			case "markdown":
				return outputMarkdown(results)
			case "console", "":
				return outputConsole(results)
			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}
		},
	}

	var runSuiteCmd = &cobra.Command{
		Use:   "run-suite [test-suite-file]...",
		Short: "Run one or more test suite files",
		Long:  `Run one or more test suite files that contain multiple test cases with shared setup and teardown.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			suiteFilePaths := args

			// Parse all test suites
			var testSuites []*parser.TestSuite
			for _, suiteFilePath := range suiteFilePaths {
				testSuite, err := parser.ParseTestSuite(suiteFilePath)
				if err != nil {
					return fmt.Errorf("failed to parse test suite '%s': %w", suiteFilePath, err)
				}
				testSuites = append(testSuites, testSuite)
			}

			// Create test suite runner
			suiteRunner := runner.NewTestSuiteRunner(runner.NewTestRunner())

			// Run all test suites
			var allResults []*parser.TestSuiteResult
			var grandTotal struct {
				TotalCases   int
				PassedCases  int
				FailedCases  int
				SkippedCases int
				Duration     time.Duration
			}

			startTime := time.Now()

			for _, testSuite := range testSuites {
				// Run the test suite
				result, err := suiteRunner.RunTestSuite(testSuite, suiteFilePaths[0]) // Use first path for relative file resolution
				if err != nil {
					return fmt.Errorf("failed to run test suite '%s': %w", testSuite.Name, err)
				}

				allResults = append(allResults, result)

				// Accumulate grand totals
				grandTotal.TotalCases += result.TotalCases
				grandTotal.PassedCases += result.PassedCases
				grandTotal.FailedCases += result.FailedCases
				grandTotal.SkippedCases += result.SkippedCases
			}

			grandTotal.Duration = time.Since(startTime)

			// Output results in specified format
			if len(allResults) == 1 {
				// Single test suite - use original output format
				switch outputFormat {
				case "json":
					return outputSuiteJSON(allResults[0])
				case "markdown":
					return outputSuiteMarkdown(allResults[0])
				case "console", "":
					return outputSuiteConsole(allResults[0])
				default:
					return fmt.Errorf("unsupported output format: %s", outputFormat)
				}
			} else {
				// Multiple test suites - use grand total format
				switch outputFormat {
				case "json":
					return outputMultipleSuitesJSON(allResults, grandTotal)
				case "markdown":
					return outputMultipleSuitesMarkdown(allResults, grandTotal)
				case "console", "":
					return outputMultipleSuitesConsole(allResults, grandTotal)
				default:
					return fmt.Errorf("unsupported output format: %s", outputFormat)
				}
			}
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List available actions",
		Long:  `List all available actions with their descriptions and examples.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := actions.NewActionRegistry()
			jsonData, err := json.MarshalIndent(registry.List(), "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal actions: %w", err)
			}
			fmt.Println(string(jsonData))
			fmt.Printf("\n📋 Available Actions (%d total):\n\n", len(registry.List()))
			for _, action := range registry.List() {
				fmt.Printf("- %s: %s\n  Example: %s\n\n", action.Name, action.Description, action.Example)
			}
			return nil
		},
	}

	var completionsCmd = &cobra.Command{
		Use:   "completions [prefix]",
		Short: "Get action completions for autocomplete",
		Long:  `Get action completions for VS Code extension autocomplete.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := actions.NewActionRegistry()

			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}

			completions := registry.GetCompletions(prefix)

			if outputFormat == "json" {
				jsonData, err := json.Marshal(completions)
				if err != nil {
					return fmt.Errorf("failed to marshal completions: %w", err)
				}
				fmt.Println(string(jsonData))
				return nil
			}

			fmt.Printf("🔍 Completions for '%s':\n", prefix)
			for _, completion := range completions {
				fmt.Printf("  %s\n", completion)
			}
			return nil
		},
	}

	// Add flags
	runCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json, markdown)")
	runCmd.Flags().BoolVarP(&parallelEnabled, "parallel", "p", false, "Enable parallel execution")
	runCmd.Flags().IntVarP(&maxConcurrency, "concurrency", "c", 4, "Maximum concurrency for parallel execution")
	runSuiteCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json, markdown)")
	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json)")
	completionsCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json)")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(runSuiteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(completionsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// outputConsole outputs results in console format
func outputConsole(results []*parser.TestResult) error {
	for _, result := range results {
		// Print captured output (step-by-step execution details)
		if result.CapturedOutput != "" {
			fmt.Print(result.CapturedOutput)
		}

		// Print test summary
		fmt.Printf("\n📊 Test Results for: %s\n", result.TestCase.Name)

		// Choose appropriate status icon
		statusIcon := "✅"
		if result.Status == "FAILED" {
			statusIcon = "❌"
		} else if result.Status == "SKIPPED" {
			statusIcon = "⏭️"
		}

		fmt.Printf("%s Status: %s\n", statusIcon, result.Status)
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		fmt.Printf("📝 Steps: %d total, %d passed, %d failed\n",
			result.TotalSteps, result.PassedSteps, result.FailedSteps)

		// Print step details with consistent duration formatting
		if len(result.StepResults) > 0 {
			fmt.Println("\nStep Results:")
			fmt.Printf("%-4s | %-24s | %-12s | %-6s | %-10s | %-24s | %-24s\n", "#", "Name", "Action", "Status", "Duration", "Output", "Error")
			fmt.Println(strings.Repeat("-", 116))
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
					stepIcon,
					duration,
					output,
					error,
				)
			}
		}
	}

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedSteps > 0 {
			os.Exit(1)
		}
	}
	return nil
}

// outputJSON outputs results in JSON format
func outputJSON(results []*parser.TestResult) error {
	// Add a duration_str field to each step result for human readability
	type StepResultWithStr struct {
		parser.StepResult
		DurationStr string `json:"duration_str"`
	}
	type TestResultWithStr struct {
		*parser.TestResult
		StepResults []StepResultWithStr `json:"step_results"`
	}

	var resultsWithStr []TestResultWithStr
	for _, r := range results {
		var stepsWithStr []StepResultWithStr
		for _, s := range r.StepResults {
			var durationStr string
			if s.Duration < time.Millisecond {
				durationStr = fmt.Sprintf("%dµs", s.Duration.Microseconds())
			} else {
				durationStr = s.Duration.String()
			}
			stepsWithStr = append(stepsWithStr, StepResultWithStr{
				StepResult:  s,
				DurationStr: durationStr,
			})
		}
		resultsWithStr = append(resultsWithStr, TestResultWithStr{
			TestResult:  r,
			StepResults: stepsWithStr,
		})
	}

	jsonBytes, err := json.MarshalIndent(resultsWithStr, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedSteps > 0 {
			os.Exit(1)
		}
	}
	return nil
}

// outputMarkdown outputs results in Markdown format
func outputMarkdown(results []*parser.TestResult) error {
	pad := func(s string, width int) string {
		if len(s) > width {
			return s[:width-3] + "..."
		}
		return s + strings.Repeat(" ", width-len(s))
	}

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
			markdown += "|-----|--------------------------|-------------|-------------------------|\n"
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
		markdown += "|------|--------------------------|-------------|--------|-----------|------------------------|-------------------------|\n"

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
			if stepResult.Duration < time.Millisecond {
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
			os.Exit(1)
		}
	}
	return nil
}

// outputSuiteConsole outputs test suite results in console format
func outputSuiteConsole(result *parser.TestSuiteResult) error {
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

	fmt.Printf("📊 Test Suite Results: %s\n", result.TestSuite.Name)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📋 Total Cases: %d\n", result.TotalCases)
	fmt.Printf("✅ Passed: %d\n", result.PassedCases)
	fmt.Printf("❌ Failed: %d\n", result.FailedCases)
	fmt.Printf("⏭️  Skipped: %d\n", result.SkippedCases)
	fmt.Printf("📝 Steps: %d total, %d passed, %d failed, %d skipped\n", totalSteps, passedSteps, failedSteps, skippedSteps)

	if result.SetupStatus != "" {
		fmt.Printf("🔧 Setup: %s\n", result.SetupStatus)
	}
	if result.TeardownStatus != "" {
		fmt.Printf("🧹 Teardown: %s\n", result.TeardownStatus)
	}

	fmt.Println("\nTest Case Results:")
	for i, caseResult := range result.CaseResults {
		caseIcon := ""
		switch caseResult.Status {
		case "passed":
			caseIcon = "✅"
		case "failed":
			caseIcon = "❌"
		case "skipped":
			caseIcon = "⏭️"
		default:
			caseIcon = "❌" // fallback for unknown status
		}
		dur := caseResult.Duration
		if dur == 0 && caseResult.Result != nil {
			dur = caseResult.Result.Duration
		}
		err := caseResult.Error
		if len(err) > 60 {
			err = err[:57] + "..."
		}
		fmt.Printf("%d. %s | %s | %v | %s\n", i+1, caseResult.TestCase.Name, caseIcon, dur, err)

		// Step-level details
		if caseResult.Result != nil && len(caseResult.Result.StepResults) > 0 {
			fmt.Println("   Steps:")
			fmt.Println("   # | Name | Status | Duration | Output | Error")
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
				fmt.Printf("   %d | %s | %s | %v | %s | %s\n", j+1, nameStr, stepStatus, step.Duration, outputStr, errStr)
			}
		}
	}

	// Exit with non-zero code if any test failed
	if result.FailedCases > 0 {
		os.Exit(1)
	}
	return nil
}

// outputSuiteJSON outputs test suite results in JSON format
func outputSuiteJSON(result *parser.TestSuiteResult) error {
	type StepSummary struct {
		Name     string        `json:"name"`
		Status   string        `json:"status"`
		Error    string        `json:"error,omitempty"`
		Duration time.Duration `json:"duration"`
	}
	type TestCaseSummary struct {
		Name     string        `json:"name"`
		Status   string        `json:"status"`
		Error    string        `json:"error,omitempty"`
		Duration time.Duration `json:"duration"`
		Steps    []StepSummary `json:"steps"`
	}
	type SuiteSummary struct {
		Name           string            `json:"suite_name"`
		Status         string            `json:"status"`
		Duration       time.Duration     `json:"duration"`
		TotalCases     int               `json:"total_cases"`
		PassedCases    int               `json:"passed_cases"`
		FailedCases    int               `json:"failed_cases"`
		SkippedCases   int               `json:"skipped_cases"`
		TotalSteps     int               `json:"total_steps"`
		PassedSteps    int               `json:"passed_steps"`
		FailedSteps    int               `json:"failed_steps"`
		SkippedSteps   int               `json:"skipped_steps"`
		SetupStatus    string            `json:"setup_status,omitempty"`
		TeardownStatus string            `json:"teardown_status,omitempty"`
		ErrorMessage   string            `json:"error_message,omitempty"`
		TestCases      []TestCaseSummary `json:"test_cases"`
	}

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

	suite := SuiteSummary{
		Name:           result.TestSuite.Name,
		Status:         result.Status,
		Duration:       result.Duration,
		TotalCases:     result.TotalCases,
		PassedCases:    result.PassedCases,
		FailedCases:    result.FailedCases,
		SkippedCases:   result.SkippedCases,
		TotalSteps:     totalSteps,
		PassedSteps:    passedSteps,
		FailedSteps:    failedSteps,
		SkippedSteps:   skippedSteps,
		SetupStatus:    result.SetupStatus,
		TeardownStatus: result.TeardownStatus,
		ErrorMessage:   result.ErrorMessage,
	}

	for _, caseResult := range result.CaseResults {
		steps := []StepSummary{}
		if caseResult.Result != nil {
			for _, step := range caseResult.Result.StepResults {
				steps = append(steps, StepSummary{
					Name:     step.Step.Name,
					Status:   step.Status,
					Error:    step.Error,
					Duration: step.Duration,
				})
			}
		}
		dur := caseResult.Duration
		if dur == 0 && caseResult.Result != nil {
			dur = caseResult.Result.Duration
		}
		suite.TestCases = append(suite.TestCases, TestCaseSummary{
			Name:     caseResult.TestCase.Name,
			Status:   caseResult.Status,
			Error:    caseResult.Error,
			Duration: dur,
			Steps:    steps,
		})
	}

	jsonBytes, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test suite results to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))

	// Exit with non-zero code if any test failed
	if result.FailedCases > 0 {
		os.Exit(1)
	}
	return nil
}

// outputSuiteMarkdown outputs test suite results in Markdown format
func outputSuiteMarkdown(result *parser.TestSuiteResult) error {
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
			markdown += "|---|------|--------|----------|--------|-------|\n"
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

	// Exit with non-zero code if any test failed
	if result.FailedCases > 0 {
		os.Exit(1)
	}
	return nil
}

// outputMultipleSuitesConsole outputs multiple test suite results in console format
func outputMultipleSuitesConsole(results []*parser.TestSuiteResult, grandTotal struct {
	TotalCases   int
	PassedCases  int
	FailedCases  int
	SkippedCases int
	Duration     time.Duration
}) error {
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
		os.Exit(1)
	}
	return nil
}

// outputMultipleSuitesJSON outputs multiple test suite results in JSON format
func outputMultipleSuitesJSON(results []*parser.TestSuiteResult, grandTotal struct {
	TotalCases   int
	PassedCases  int
	FailedCases  int
	SkippedCases int
	Duration     time.Duration
}) error {
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test suite results to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedCases > 0 {
			os.Exit(1)
		}
	}
	return nil
}

// outputMultipleSuitesMarkdown outputs multiple test suite results in Markdown format
func outputMultipleSuitesMarkdown(results []*parser.TestSuiteResult, grandTotal struct {
	TotalCases   int
	PassedCases  int
	FailedCases  int
	SkippedCases int
	Duration     time.Duration
}) error {
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
		os.Exit(1)
	}
	return nil
}
