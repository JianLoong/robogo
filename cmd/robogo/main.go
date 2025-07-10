package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/output"
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

// cleanup performs graceful shutdown of all resources
func cleanup() {
	// Connection cleanup is now handled by ActionContext
	// Close all RabbitMQ connections
	if err := actions.CloseAllRabbitMQConnections(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error closing RabbitMQ connections: %v\n", err)
	}
}

// setupGracefulShutdown sets up signal handlers for graceful shutdown
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Fprintf(os.Stderr, "\nReceived shutdown signal, cleaning up resources...\n")
		cleanup()
		fmt.Fprintf(os.Stderr, "Resource cleanup completed\n")
		os.Exit(0)
	}()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n", r)
		}
	}()
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func realMain() error {
	// Setup graceful shutdown before any other operations
	setupGracefulShutdown()

	// Ensure cleanup happens even if we exit normally
	defer cleanup()

	// Create the action registry and executor once
	registry := actions.NewActionRegistry()
	executor := actions.NewActionExecutor(registry)

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
			// Create context and action context
			ctx := context.Background()
			actionCtx := actions.NewActionContext()
			defer actionCtx.Cleanup()
			ctx = actions.WithActionContext(ctx, actionCtx)
			
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
							testExecutor := runner.NewTestExecutionService(executor)
							suiteRunner := runner.NewTestSuiteRunner(testExecutor)
							result, err := suiteRunner.RunTestSuite(ctx, ts, fp, false)
							if err != nil {
								return err
							}
							suiteResults = append(suiteResults, result)
						} else {
							results, err := runner.RunTestFilesWithConfig(ctx, []string{fp}, silent, parallelConfig, executor)
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
						testExecutor := runner.NewTestExecutionService(executor)
						suiteRunner := runner.NewTestSuiteRunner(testExecutor)
						result, err := suiteRunner.RunTestSuite(ctx, ts, path, false)
						if err != nil {
							return err
						}
						suiteResults = append(suiteResults, result)
					} else {
						results, err := runner.RunTestFilesWithConfig(ctx, []string{path}, silent, parallelConfig, executor)
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
					formatter := output.NewFormatter(output.Format(outputFormat))
					return formatter.FormatSuiteResult(suiteResults[0])
				} else {
					// Multiple suites
					var grandTotal struct {
						TotalCases   int
						PassedCases  int
						FailedCases  int
						SkippedCases int
						Duration     time.Duration
					}
					for _, sr := range suiteResults {
						grandTotal.TotalCases += sr.TotalCases
						grandTotal.PassedCases += sr.PassedCases
						grandTotal.FailedCases += sr.FailedCases
						grandTotal.SkippedCases += sr.SkippedCases
						grandTotal.Duration += sr.Duration
					}
					formatter := output.NewFormatter(output.Format(outputFormat))
					return formatter.FormatMultipleSuites(suiteResults, output.GrandTotal{
						TotalCases:   grandTotal.TotalCases,
						PassedCases:  grandTotal.PassedCases,
						FailedCases:  grandTotal.FailedCases,
						SkippedCases: grandTotal.SkippedCases,
						Duration:     grandTotal.Duration,
					})
				}
			} else if len(caseResults) > 0 && len(suiteResults) == 0 {
				// Only test cases
				formatter := output.NewFormatter(output.Format(outputFormat))
				return formatter.FormatTestResults(caseResults)
			} else {
				// Mixed (rare)
				fmt.Println("Warning: Mixed test suites and test cases in one run. Outputting all results.")
				formatter := output.NewFormatter(output.FormatConsole)
				for _, sr := range suiteResults {
					_ = formatter.FormatSuiteResult(sr)
				}
				_ = formatter.FormatTestResults(caseResults)
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

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

