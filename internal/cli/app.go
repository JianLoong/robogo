package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/spf13/cobra"
)

// App represents the main CLI application
type App struct {
	rootCmd  *cobra.Command
	executor *actions.ActionExecutor
	version  string
	commit   string
	date     string
}

// NewApp creates a new CLI application instance
func NewApp(version, commit, date string) *App {
	app := &App{
		version: version,
		commit:  commit,
		date:    date,
	}

	app.setupExecutor()
	app.setupCommands()
	app.setupGracefulShutdown()

	return app
}

// Run executes the CLI application
func (app *App) Run(args []string) error {
	// Set up panic recovery and cleanup
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n", r)
		}
		app.cleanup()
	}()

	// Ensure cleanup happens on normal exit too
	defer app.cleanup()

	app.rootCmd.SetArgs(args[1:]) // Remove program name
	return app.rootCmd.Execute()
}

// setupExecutor initializes the action executor
func (app *App) setupExecutor() {
	registry := actions.NewActionRegistry()
	app.executor = actions.NewActionExecutor(registry)
}

// setupCommands configures all CLI commands
func (app *App) setupCommands() {
	app.rootCmd = &cobra.Command{
		Use:   "robogo",
		Short: "Robogo - A modern, git-driven test automation framework",
		Long: `Robogo is a modern, git-driven test automation framework written in Go.
It provides fast, extensible, and developer-friendly test automation with YAML-based test cases or test suites.

Key Features:
- Modern variable management: define variables in a 'variables:' block and assign step outputs to variables (no 'variable' action needed)
- Secure secret management: supports inline and file-based secrets (single value per file)
- PostgreSQL database actions (connect, query, execute, close)
- Control flow: if, for, while loops
- Secret masking in output for security
- URL encoding for PostgreSQL connection strings
- Comprehensive test coverage`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", app.version, app.commit, app.date),
	}

	// Add subcommands
	app.rootCmd.AddCommand(app.createRunCommand())
	app.rootCmd.AddCommand(app.createListCommand())
	app.rootCmd.AddCommand(app.createCompletionsCommand())

	// Configure command behavior
	app.rootCmd.SilenceUsage = true
	app.rootCmd.SilenceErrors = true
}

// createRunCommand creates the run command
func (app *App) createRunCommand() *cobra.Command {
	var parallelEnabled bool
	var maxConcurrency int
	var variableDebug bool
	var silent bool

	cmd := &cobra.Command{
		Use:   "run [test-files...]",
		Short: "Run one or more test case or test suite files",
		Long:  `Run one or more test case or test suite files. You can specify multiple files or a directory.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runTests(cmd.Context(), args, RunOptions{
				ParallelEnabled: parallelEnabled,
				MaxConcurrency:  maxConcurrency,
				VariableDebug:   variableDebug,
				Silent:          silent,
				ParallelConfig:  app.createParallelConfig(parallelEnabled, maxConcurrency),
			})
		},
	}

	cmd.Flags().BoolVarP(&parallelEnabled, "parallel", "p", false, "Enable parallel execution")
	cmd.Flags().IntVarP(&maxConcurrency, "concurrency", "c", 4, "Maximum concurrency for parallel execution")
	cmd.Flags().BoolVarP(&variableDebug, "debug-vars", "d", false, "Enable variable resolution debugging")
	cmd.Flags().BoolVarP(&silent, "silent", "s", false, "Suppress output during execution")

	return cmd
}

// createListCommand creates the list command
func (app *App) createListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available actions",
		Long:  `List all available actions with their descriptions and examples.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listActions()
		},
	}

	return cmd
}

// createCompletionsCommand creates the completions command
func (app *App) createCompletionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completions [prefix]",
		Short: "Get action completions for autocomplete",
		Long:  `Get action completions for VS Code extension autocomplete.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}
			return app.getCompletions(prefix)
		},
	}

	return cmd
}

// runTests orchestrates the test execution process
func (app *App) runTests(ctx context.Context, paths []string, options RunOptions) error {
	// Setup action context for resource management
	actionCtx := actions.NewActionContext()
	defer actionCtx.Cleanup()
	ctx = actions.WithActionContext(ctx, actionCtx)

	// Show debug info if enabled
	if options.VariableDebug {
		fmt.Printf("🔍 Variable debugging enabled\n\n")
	}

	// Process files
	processor := NewFileProcessor(app.executor, options)
	results, err := processor.ProcessPaths(ctx, paths)
	if err != nil {
		return err
	}

	// Format and output results
	formatter := NewResultFormatter(options.Silent)
	return formatter.FormatResults(results)
}

// listActions lists all available actions
func (app *App) listActions() error {
	registry := actions.NewActionRegistry()
	actionList := registry.List()

	fmt.Printf("📋 Available Actions (%d total):\n\n", len(actionList))
	for _, action := range actionList {
		fmt.Printf("- %s: %s\n", action.Name, action.Description)
		if action.Example != "" {
			fmt.Printf("  Example: %s\n", action.Example)
		}
		fmt.Println()
	}

	return nil
}

// getCompletions gets action completions for autocomplete
func (app *App) getCompletions(prefix string) error {
	registry := actions.NewActionRegistry()
	completions := registry.GetCompletions(prefix)

	fmt.Printf("🔍 Completions for '%s':\n", prefix)
	for _, completion := range completions {
		fmt.Printf("  %s\n", completion)
	}

	return nil
}

// createParallelConfig creates parallel configuration if enabled
func (app *App) createParallelConfig(enabled bool, maxConcurrency int) *parser.ParallelConfig {
	if !enabled {
		return nil
	}

	return &parser.ParallelConfig{
		Enabled:        true,
		MaxConcurrency: maxConcurrency,
		TestCases:      true,
		Steps:          true,
		HTTPRequests:   true,
	}
}

// setupGracefulShutdown configures signal handlers for graceful shutdown
func (app *App) setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Fprintf(os.Stderr, "\nReceived shutdown signal, cleaning up resources...\n")
		app.cleanup()
		fmt.Fprintf(os.Stderr, "Resource cleanup completed\n")
		os.Exit(0)
	}()
}

// cleanup performs graceful shutdown of all resources
func (app *App) cleanup() {
	// Close all action resources using ActionContext
	if app.executor != nil && app.executor.ActionContext != nil {
		app.executor.ActionContext.Cleanup()
	}
}
