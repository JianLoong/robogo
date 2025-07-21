package internal

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// Exit codes for CLI
const (
	ExitSuccess     = 0 // Normal successful exit
	ExitUsageError  = 1 // Usage or argument error
	ExitTestFailure = 2 // Test execution failed
)

// ParsedArgs holds parsed command line arguments
type ParsedArgs struct {
	envFile    string   // --env flag value
	positional []string // non-flag arguments
}

// Table formatting and truncation widths for printTestSummary
const (
	colStepNumWidth  = 3  // Width for step number column
	colStepNameWidth = 40 // Width for step name column
	colStatusWidth   = 8  // Width for status column
	colDurationWidth = 12 // Width for duration column
	colMessageWidth  = 50 // Width for message column (error/failure message)
	colCategoryWidth = 13 // Width for category column

	truncStepName = 37 // Truncate step name to this length before adding '...'
	truncMessage  = 47 // Truncate message to this length before adding '...'
	truncCategory = 9  // Truncate category to this length before adding '...'
)

// parseArgs parses command line arguments, handling flags and positional arguments
func parseArgs() ParsedArgs {
	args := ParsedArgs{
		envFile:    "",
		positional: []string{},
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		
		if strings.HasPrefix(arg, "--env=") {
			args.envFile = arg[6:] // Remove "--env=" prefix
		} else if arg == "--env" && i+1 < len(os.Args) {
			i++ // Move to next argument
			args.envFile = os.Args[i]
		} else if !strings.HasPrefix(arg, "-") {
			args.positional = append(args.positional, arg)
		} else {
			fmt.Printf("Error: unknown flag '%s'\n", arg)
			printUsage()
			os.Exit(ExitUsageError)
		}
	}

	return args
}

// SimpleCLI - direct, no-abstraction CLI
func RunCLI() {
	// Parse command line arguments first to check for --env flag
	args := parseArgs()

	// Load .env file - use custom file if specified, otherwise try default
	if args.envFile != "" {
		if err := common.LoadDotEnv(args.envFile); err != nil {
			fmt.Printf("[ERROR] Failed to load specified .env file '%s': %v\n", args.envFile, err)
			os.Exit(ExitUsageError)
		}
	} else {
		// Try to load default .env file (no error if it doesn't exist)
		if err := common.LoadDotEnvWithDefault(); err != nil {
			fmt.Printf("[WARN] Failed to load .env file: %v\n", err)
		}
	}

	// Setup signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down gracefully...")
		// No cleanup needed - connections close automatically
		os.Exit(ExitSuccess)
	}()

	if len(args.positional) < 1 {
		printUsage()
		os.Exit(ExitUsageError)
	}

	command := args.positional[0]

	switch command {
	case "run":
		if len(args.positional) < 2 {
			fmt.Println("Error: run command requires a test file")
			printUsage()
			os.Exit(ExitUsageError)
		}
		runTest(args.positional[1])

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

	if result.Status == "FAIL" || result.Status == "FAILED" || result.Status == "failed" || result.Status == "error" || result.Status == "ERROR" {
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
	fmt.Println("  robogo [flags] <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  run <test-file>               Run a single test")
	fmt.Println("  list                          List available actions")
	fmt.Println("  version                       Show version")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --env <file>                  Load environment variables from specified file")
	fmt.Println("                                (default: .env in current directory)")
}

// getCategory returns the category from ErrorInfo or FailureInfo
func getCategory(result types.ActionResult) string {
	if result.ErrorInfo != nil {
		return string(result.ErrorInfo.Category)
	}
	if result.FailureInfo != nil {
		return string(result.FailureInfo.Category)
	}
	return ""
}

func printTestSummary(result *types.TestResult) {
	fmt.Println("\nTest Summary:")
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  Duration: %s\n", result.Duration)
	if errorMsg := result.GetErrorMessage(); errorMsg != "" {
		fmt.Printf("  Error: %s\n", errorMsg)
	}
	fmt.Println()

	// Print table header
	headerFormat := "| %*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n"
	separatorFormat := "|%s|%s|%s|%s|%s|%s|\n"

	fmt.Printf(headerFormat,
		colStepNumWidth, "#",
		colStepNameWidth, "Step Name",
		colStatusWidth, "Status",
		colDurationWidth, "Duration",
		colMessageWidth, "Message",
		colCategoryWidth, "Category")

	fmt.Printf(separatorFormat,
		strings.Repeat("-", colStepNumWidth+2),
		strings.Repeat("-", colStepNameWidth+2),
		strings.Repeat("-", colStatusWidth+2),
		strings.Repeat("-", colDurationWidth+2),
		strings.Repeat("-", colMessageWidth+2),
		strings.Repeat("-", colCategoryWidth+2))

	stepNum := 1

	// Print setup steps
	for _, step := range result.SetupSteps {
		printStepRow(stepNum, step, "[SETUP] ")
		stepNum++
	}

	// Print main test steps
	for _, step := range result.Steps {
		printStepRow(stepNum, step, "")
		stepNum++
	}

	// Print teardown steps
	for _, step := range result.TeardownSteps {
		printStepRow(stepNum, step, "[TEARDOWN] ")
		stepNum++
	}
}

// printStepRow prints a single step row in the summary table
func printStepRow(stepNum int, step types.StepResult, prefix string) {
	stepName := prefix + step.Name
	if len(stepName) > colStepNameWidth {
		stepName = stepName[:truncStepName] + "..."
	}

	// Get message (error or failure message)
	message := step.Result.GetMessage()
	if len(message) > colMessageWidth {
		message = message[:truncMessage] + "..."
	}

	// Get category from ErrorInfo or FailureInfo
	category := getCategory(step.Result)
	if len(category) > colCategoryWidth {
		category = category[:truncCategory] + "..."
	}

	// Print table row
	rowFormat := "| %*d | %-*s | %-*s | %-*s | %-*s | %-*s |\n"
	fmt.Printf(rowFormat,
		colStepNumWidth, stepNum,
		colStepNameWidth, stepName,
		colStatusWidth, step.Result.Status,
		colDurationWidth, step.Duration.String(),
		colMessageWidth, message,
		colCategoryWidth, category)
}
