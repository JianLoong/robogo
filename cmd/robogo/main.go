package main

import (
	"encoding/json"
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// outputConsole outputs results in console format
func outputConsole(results []*parser.TestResult) error {
	for _, result := range results {
		fmt.Print(result.CapturedOutput)
		fmt.Printf("\n Test Results for: %s\n", result.TestCase.Name)
		fmt.Printf("✅ Status: %s\n", result.Status)
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)

		fmt.Printf("📝 Steps: %d total, %d passed, %d failed\n",
			len(result.TestCase.Steps), result.PassedSteps, result.FailedSteps)
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
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
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
			markdown += "| # | Name | Action | Error |\n"
			markdown += "|---|------|--------|-------|\n"
			for i, stepResult := range result.StepResults {
				if stepResult.Status == "FAILED" {
					stepName := stepResult.Step.Name
					if stepName == "" {
						stepName = "(unnamed)"
					}
					error := stepResult.Error
					if len(error) > 60 {
						error = error[:57] + "..."
					}
					markdown += fmt.Sprintf("| %d | %s | %s | %s |\n",
						i+1,
						stepName,
						stepResult.Step.Action,
						error,
					)
				}
			}
		}

		markdown += "\n## Step Results\n"
		// Add markdown table header
		markdown += "| Step | Name | Action | Status | Duration | Output | Error |\n"
		markdown += "|------|------|--------|--------|----------|--------|-------|\n"

		// Add step details as table rows
		for i, stepResult := range result.StepResults {
			stepIcon := "✅"
			if stepResult.Status == "FAILED" {
				stepIcon = "❌"
			}
			output := stepResult.Output
			if len(output) > 40 {
				output = output[:37] + "..."
			}
			error := stepResult.Error
			if len(error) > 40 {
				error = error[:37] + "..."
			}
			stepName := stepResult.Step.Name
			if stepName == "" {
				stepName = "(unnamed)"
			}
			markdown += fmt.Sprintf("| %d | %s | %s | %s | %v | %s | %s |\n",
				i+1,
				stepName,
				stepResult.Step.Action,
				stepIcon,
				stepResult.Duration,
				output,
				error,
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
