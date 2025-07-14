package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JianLoong/robogo/internal/actions"
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

	case "run-suite":
		if len(os.Args) < 3 {
			fmt.Println("Error: run-suite command requires a suite file")
			printUsage()
			os.Exit(1)
		}
		runSuite(os.Args[2])

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
		fmt.Printf("\nFAILED: Test execution failed: %s\n", err.Error())
		os.Exit(1)
	}

	printTestSummary(result)

	if result.Status == "FAILED" {
		os.Exit(1)
	}
}

func runSuite(filename string) {
	runner := NewTestRunner()
	err := runner.RunSuite(filename)

	if err != nil {
		fmt.Printf("\nFAILED: Suite execution failed: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("\nSUCCESS: Suite execution completed successfully")
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
	fmt.Println("  robogo run-suite <suite-file> Run a test suite")
	fmt.Println("  robogo list                   List available actions")
	fmt.Println("  robogo version                Show version")
}

func printTestSummary(result *TestResult) {
	passed := 0
	failed := 0
	for _, step := range result.Steps {
		if step.Status == "PASSED" {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\n")
	fmt.Printf("| %-11s | %-20s |\n", "Field", "Value")
	fmt.Printf("|-------------|----------------------|\n")
	fmt.Printf("| %-11s | %-20s |\n", "Test", result.Name)
	fmt.Printf("| %-11s | %-20s |\n", "Status", result.Status)
	fmt.Printf("| %-11s | %-20s |\n", "Duration", result.Duration.String())
	fmt.Printf("| %-11s | %-20d |\n", "Total Steps", len(result.Steps))
	fmt.Printf("| %-11s | %-20d |\n", "Passed", passed)
	fmt.Printf("| %-11s | %-20d |\n", "Failed", failed)

	if result.Status == "FAILED" && result.Error != "" {
		fmt.Printf("| %-11s | %-20s |\n", "Error", result.Error)
	}
}