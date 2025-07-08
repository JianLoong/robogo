package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func isTestSuiteFile(filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	content := string(buf[:n])
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "testsuite:") {
			return true, nil
		}
		if strings.HasPrefix(line, "testcase:") {
			return false, nil
		}
		// fallback: if it looks like a suite
		if strings.HasPrefix(line, "testcases:") {
			return true, nil
		}
		break
	}
	return false, nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "robogo",
		Short: "Robogo - A modern, git-driven test automation framework",
		Long: `Robogo is a modern, git-driven test automation framework written in Go.
It provides fast, extensible, and developer-friendly test automation with YAML-based test cases or test suites.

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
		Short: "Run one or more test case or test suite files",
		Long:  `Run one or more test case or test suite files. You can specify multiple files or a directory.`,
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

			var suiteResults []*parser.TestSuiteResult
			var caseResults []*parser.TestResult

			for _, path := range args {
				info, err := os.Stat(path)
				if err != nil {
					return fmt.Errorf("failed to stat %s: %w", path, err)
				}
				if info.IsDir() {
					// Recursively find all .robogo files
					err := filepath.Walk(path, func(fp string, fi os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if fi.IsDir() || !strings.HasSuffix(fp, ".robogo") {
							return nil
						}
						isSuite, err := isTestSuiteFile(fp)
						if err != nil {
							return err
						}
						if isSuite {
							ts, err := parser.ParseTestSuite(fp)
							if err != nil {
								return err
							}
							suiteRunner := runner.NewTestSuiteRunner(runner.NewTestRunner())
							result, err := suiteRunner.RunTestSuite(ts, fp, false)
							if err != nil {
								return err
							}
							suiteResults = append(suiteResults, result)
						} else {
							results, err := runner.RunTestFilesWithConfig([]string{fp}, silent, parallelConfig)
							if err != nil {
								return err
							}
							caseResults = append(caseResults, results...)
						}
						return nil
					})
					if err != nil {
						return err
					}
				} else {
					isSuite, err := isTestSuiteFile(path)
					if err != nil {
						return err
					}
					if isSuite {
						ts, err := parser.ParseTestSuite(path)
						if err != nil {
							return err
						}
						suiteRunner := runner.NewTestSuiteRunner(runner.NewTestRunner())
						result, err := suiteRunner.RunTestSuite(ts, path, false)
						if err != nil {
							return err
						}
						suiteResults = append(suiteResults, result)
					} else {
						results, err := runner.RunTestFilesWithConfig([]string{path}, silent, parallelConfig)
						if err != nil {
							return err
						}
						caseResults = append(caseResults, results...)
					}
				}
			}

			// Output results in specified format
			if len(suiteResults) > 0 && len(caseResults) == 0 {
				// Only suites
				if len(suiteResults) == 1 {
					switch outputFormat {
					case "json":
						return outputSuiteJSON(suiteResults[0])
					case "markdown":
						return outputSuiteMarkdown(suiteResults[0])
					case "console", "":
						return outputSuiteConsole(suiteResults[0])
					default:
						return fmt.Errorf("unsupported output format: %s", outputFormat)
					}
				} else {
					// Multiple suites
					// TODO: aggregate and output grand total
					return outputMultipleSuitesConsole(suiteResults, struct {
						TotalCases   int
						PassedCases  int
						FailedCases  int
						SkippedCases int
						Duration     time.Duration
					}{})
				}
			} else if len(caseResults) > 0 && len(suiteResults) == 0 {
				// Only test cases
				switch outputFormat {
				case "json":
					return outputJSON(caseResults)
				case "markdown":
					return outputMarkdown(caseResults)
				case "console", "":
					return outputConsole(caseResults)
				default:
					return fmt.Errorf("unsupported output format: %s", outputFormat)
				}
			} else {
				// Mixed (rare)
				fmt.Println("Warning: Mixed test suites and test cases in one run. Outputting all results.")
				for _, sr := range suiteResults {
					_ = outputSuiteConsole(sr)
				}
				_ = outputConsole(caseResults)
				return nil
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
	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json)")
	completionsCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json)")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(completionsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
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

		// Print test summary in markdown format
		fmt.Printf("\n## 📊 Test Results for: %s\n\n", result.TestCase.Name)

		// Choose appropriate status icon
		statusIcon := "✅"
		if result.Status == "FAILED" {
			statusIcon = "❌"
		} else if result.Status == "SKIPPED" {
			statusIcon = "⏭️"
		}

		fmt.Printf("**%s Status:** %s\n\n", statusIcon, result.Status)
		fmt.Printf("**⏱️ Duration:** %v\n\n", result.Duration)
		fmt.Printf("**📝 Steps Summary:**\n\n")
		fmt.Printf("| %-6s | %-7s | %-6s | %-7s |\n", "Total", "Passed", "Failed", "Skipped")
		fmt.Printf("|--------|---------|--------|---------|\n")
		fmt.Printf("| %-6d | %-7d | %-6d | %-7d |\n\n",
			result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps)

		// Print step details as a markdown table
		if len(result.StepResults) > 0 {
			fmt.Println("\nStep Results (Markdown Table):")
			runner.PrintStepResultsMarkdown(result.StepResults, "Step Results:")
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
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 Test Suite Results: %s\n", result.TestSuite.Name)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("\n## Test Case Summary\n")
	fmt.Printf("| %-4s | %-24s | %-8s | %-10s | %-24s |\n", "#", "Name", "Status", "Duration", "Error")
	fmt.Printf("|------|--------------------------|----------|------------|--------------------------|\n")
	for i, caseResult := range result.CaseResults {
		icon := ""
		status := ""
		switch caseResult.Status {
		case "passed":
			icon = "✅"
			status = "PASSED"
		case "failed":
			icon = "❌"
			status = "FAILED"
		case "skipped":
			icon = "⏭️"
			status = "SKIPPED"
		}
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
		fmt.Printf("| %-4d | %-24s | %-8s | %-10s | %-24s |\n", i+1, name, icon+" "+status, duration, error)
	}

	// Print step results for each test case
	for _, caseResult := range result.CaseResults {
		if caseResult.Result != nil && len(caseResult.Result.StepResults) > 0 {
			title := "### Step Results for " + caseResult.TestCase.Name
			runner.PrintStepResultsMarkdown(caseResult.Result.StepResults, title)
		}
	}

	// Print step summary table
	fmt.Printf("\n## Step Summary\n")
	fmt.Printf("| %-8s | %-8s | %-8s | %-8s |\n", "Total", "Passed", "Failed", "Skipped")
	fmt.Printf("|----------|----------|----------|----------|\n")
	fmt.Printf("| %-8d | %-8d | %-8d | %-8d |\n", result.TotalSteps, result.PassedSteps, result.FailedSteps, result.SkippedSteps)

	if result.SetupStatus != "" {
		fmt.Printf("\n🔧 Setup: %s\n", result.SetupStatus)
	}
	if result.TeardownStatus != "" {
		fmt.Printf("\n🧹 Teardown: %s\n", result.TeardownStatus)
	}

	// switch result.Status {
	// case "passed":
	// 	fmt.Printf("\n🎉 Test suite completed successfully!\n")
	// case "failed":
	// 	fmt.Printf("\n💥 Test suite failed!\n")
	// case "skipped":
	// 	fmt.Printf("\n⏭️  Test suite skipped due to setup failure!\n")
	// }
	fmt.Printf(strings.Repeat("=", 60) + "\n")

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
