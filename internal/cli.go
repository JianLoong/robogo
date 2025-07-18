package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/types"
)

// Exit codes for CLI
const (
	ExitSuccess     = 0 // Normal successful exit
	ExitUsageError  = 1 // Usage or argument error
	ExitTestFailure = 2 // Test execution failed
)

// Table formatting and truncation widths for printTestSummary
const (
	colStepNumWidth  = 3  // Width for step number column
	colStepNameWidth = 32 // Width for step name column
	colStatusWidth   = 8  // Width for status column
	colDurationWidth = 10 // Width for duration column
	colOutputWidth   = 29 // Width for output column
	colErrorWidth    = 29 // Width for error column
	colReasonWidth   = 29 // Width for reason column

	truncStepName = 29 // Truncate step name to this length before adding '...'
	truncOutput   = 26 // Truncate output to this length before adding '...'
	truncError    = 26 // Truncate error to this length before adding '...'
	truncReason   = 26 // Truncate reason to this length before adding '...'
)

// SimpleCLI - direct, no-abstraction CLI
func RunCLI() {
	// Setup signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down gracefully...")
		// No cleanup needed - connections close automatically
		os.Exit(ExitSuccess)
	}()

	// No cleanup needed - connections close automatically

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(ExitUsageError)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: run command requires a test file")
			printUsage()
			os.Exit(ExitUsageError)
		}
		runTest(os.Args[2])

	case "list":
		listActions()

	case "version":
		fmt.Println("Robogo Simple v1.0.0")

	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(ExitUsageError)
	}
}

func runTest(filename string) {
	runner := NewTestRunner()
	result, err := runner.RunTest(filename)

	if err != nil {
		fmt.Printf("\nERROR: Test execution failed: %s\n", err.Error())
		os.Exit(ExitTestFailure)
	}

	printTestSummary(result)

	if result.Status == "FAILED" || result.Status == "failed" || result.Status == "error" || result.Status == "ERROR" {
		os.Exit(ExitTestFailure)
	}
}

func listActions() {
	fmt.Println("Available actions:")
	for _, action := range actions.ListActions() {
		fmt.Printf("  - %s\n", action)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  robogo run <test-file>        Run a single test")
	fmt.Println("  robogo list                   List available actions")
	fmt.Println("  robogo version                Show version")
}

func printTestSummary(result *types.TestResult) {
	fmt.Println("\nTest Summary:")
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  Duration: %s\n", result.Duration)
	if result.Error != "" {
		fmt.Printf("  Error: %s\n", result.Error)
	}
	fmt.Println("\n|  #  | Step Name                        | Status   | Duration   | Output                        | Error                         | Reason                        |")
	fmt.Println("|-----|----------------------------------|----------|------------|-------------------------------|-------------------------------|-------------------------------|")
	for i, step := range result.Steps {
		stepName := step.Name
		if len(stepName) > colStepNameWidth {
			stepName = stepName[:truncStepName] + "..."
		}
		output := step.Result.Output
		if len(output) > colOutputWidth {
			output = output[:truncOutput] + "..."
		}
		errorMsg := step.Result.Error
		if len(errorMsg) > colErrorWidth {
			errorMsg = errorMsg[:truncError] + "..."
		}
		reason := step.Result.Reason
		if len(reason) > colReasonWidth {
			reason = reason[:truncReason] + "..."
		}
		fmt.Printf("| %"+fmt.Sprintf("%dd", colStepNumWidth)+" | %-"+fmt.Sprintf("%ds", colStepNameWidth)+" | %-"+fmt.Sprintf("%ds", colStatusWidth)+" | %-"+fmt.Sprintf("%ds", colDurationWidth)+" | %-"+fmt.Sprintf("%ds", colOutputWidth)+" | %-"+fmt.Sprintf("%ds", colErrorWidth)+" | %-"+fmt.Sprintf("%ds", colReasonWidth)+" |\n",
			i+1, stepName, step.Result.Status, step.Duration.String(), output, errorMsg, reason)
	}
}
