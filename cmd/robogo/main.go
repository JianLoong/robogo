package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/robogo/internal/keywords"
	"github.com/your-org/robogo/internal/parser"
	"github.com/your-org/robogo/internal/runner"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
	
	// CLI flags
	outputFormat string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "robogo",
		Short: "Robogo - A modern, git-driven test automation framework",
		Long: `Robogo is a modern, git-driven test automation framework written in Go.
It provides fast, extensible, and developer-friendly test automation with YAML-based test cases.`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
	}

	var runCmd = &cobra.Command{
		Use:   "run [test-file]",
		Short: "Run a test case file",
		Long:  `Run a test case from a YAML or .robogo file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			testFile := args[0]
			
			// Determine silent mode
			silent := false
			switch outputFormat {
			case "json", "markdown":
				silent = true
			}
			// Run the test
			result, err := runner.RunTestFile(testFile, silent)
			if err != nil {
				return fmt.Errorf("failed to run test: %w", err)
			}

			// Output results in specified format
			switch outputFormat {
			case "json":
				return outputJSON(result)
			case "markdown":
				return outputMarkdown(result)
			case "console", "":
				return outputConsole(result)
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
			registry := keywords.NewActionRegistry()
			actions := registry.List()
			jsonData, err := json.MarshalIndent(actions, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal actions: %w", err)
			}
			fmt.Println(string(jsonData))
			fmt.Printf("\n📋 Available Actions (%d total):\n\n", len(actions))
			for _, action := range actions {
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
			registry := keywords.NewActionRegistry()
			
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
func outputConsole(result *parser.TestResult) error {
	fmt.Printf("\n📊 Test Results:\n")
	fmt.Printf("✅ Status: %s\n", result.Status)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("📝 Steps: %d total, %d passed, %d failed\n", 
		result.TotalSteps, result.PassedSteps, result.FailedSteps)

	if result.FailedSteps > 0 {
		os.Exit(1)
	}
	return nil
}

// outputJSON outputs results in JSON format
func outputJSON(result *parser.TestResult) error {
	jsonOutput := fmt.Sprintf(`{
  "testcase": "%s",
  "status": "%s",
  "duration": "%v",
  "total_steps": %d,
  "passed_steps": %d,
  "failed_steps": %d,
  "error_message": "%s"
}`, 
		result.TestCase.Name,
		result.Status,
		result.Duration,
		result.TotalSteps,
		result.PassedSteps,
		result.FailedSteps,
		result.ErrorMessage)

	fmt.Println(jsonOutput)
	
	if result.FailedSteps > 0 {
		os.Exit(1)
	}
	return nil
}

// outputMarkdown outputs results in Markdown format
func outputMarkdown(result *parser.TestResult) error {
	statusIcon := "✅"
	if result.Status == "FAILED" {
		statusIcon = "❌"
	}

	markdown := fmt.Sprintf(`# Test Results: %s

## Summary
%s **Status:** %s  
⏱️ **Duration:** %v  
📝 **Steps:** %d total, %d passed, %d failed

## Test Case Details
- **Name:** %s
- **Description:** %s

## Step Results
`, 
		result.TestCase.Name,
		statusIcon,
		result.Status,
		result.Duration,
		result.TotalSteps,
		result.PassedSteps,
		result.FailedSteps,
		result.TestCase.Name,
		result.TestCase.Description)

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
	
	if result.FailedSteps > 0 {
		os.Exit(1)
	}
	return nil
} 