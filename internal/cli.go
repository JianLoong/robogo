package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/types"
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
		os.Exit(0)
	}()

	// No cleanup needed - connections close automatically

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: run command requires a test file")
			printUsage()
			os.Exit(1)
		}
		runTest(os.Args[2])

	case "list":
		listActions()

	case "version":
		fmt.Println("Robogo Simple v1.0.0")

	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runTest(filename string) {
	runner := NewTestRunner()
	result, err := runner.RunTest(filename)

	if err != nil {
		fmt.Printf("\nERROR: Test execution failed: %s\n", err.Error())
		os.Exit(2)
	}

	printTestSummary(result)

	if result.Status == "FAILED" || result.Status == "failed" || result.Status == "error" || result.Status == "ERROR" {
		os.Exit(1)
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
		if len(stepName) > 32 {
			stepName = stepName[:29] + "..."
		}
		output := step.Result.Output
		if len(output) > 29 {
			output = output[:26] + "..."
		}
		errorMsg := step.Result.Error
		if len(errorMsg) > 29 {
			errorMsg = errorMsg[:26] + "..."
		}
		reason := step.Result.Reason
		if len(reason) > 29 {
			reason = reason[:26] + "..."
		}
		fmt.Printf("| %3d | %-32s | %-8s | %-10s | %-29s | %-29s | %-29s |\n",
			i+1, stepName, step.Result.Status, step.Duration.String(), output, errorMsg, reason)
	}
}
