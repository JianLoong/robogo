package cli

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/output"
)

// ResultFormatter handles formatting and output of test results
type ResultFormatter struct {
	silent bool
}

// NewResultFormatter creates a new result formatter
func NewResultFormatter(silent bool) *ResultFormatter {
	return &ResultFormatter{
		silent: silent,
	}
}

// FormatResults formats and outputs the test results based on their type
func (rf *ResultFormatter) FormatResults(results *RunResults) error {
	if results.IsEmpty() {
		return fmt.Errorf("no test results to format")
	}

	// If silent mode is enabled, only show exit code, no output
	if rf.silent {
		return nil
	}

	formatter := output.NewFormatter()

	if results.HasOnlySuites() {
		return rf.formatSuiteResults(formatter, results)
	}

	if results.HasOnlyCases() {
		return formatter.FormatTestResults(results.CaseResults)
	}

	if results.HasMixed() {
		return rf.formatMixedResults(formatter, results)
	}

	return fmt.Errorf("unexpected result state")
}

// formatSuiteResults formats suite-only results
func (rf *ResultFormatter) formatSuiteResults(formatter output.Formatter, results *RunResults) error {
	if len(results.SuiteResults) == 1 {
		return formatter.FormatSuiteResult(results.SuiteResults[0])
	}

	// Multiple suites
	grandTotal := results.CalculateGrandTotal()
	return formatter.FormatMultipleSuites(results.SuiteResults, output.GrandTotal{
		TotalCases:   grandTotal.TotalCases,
		PassedCases:  grandTotal.PassedCases,
		FailedCases:  grandTotal.FailedCases,
		SkippedCases: grandTotal.SkippedCases,
		Duration:     grandTotal.Duration,
	})
}

// formatMixedResults formats mixed suite and case results
func (rf *ResultFormatter) formatMixedResults(formatter output.Formatter, results *RunResults) error {
	fmt.Println("Warning: Mixed test suites and test cases in one run. Outputting all results.")

	for _, suiteResult := range results.SuiteResults {
		if err := formatter.FormatSuiteResult(suiteResult); err != nil {
			return err
		}
	}

	if len(results.CaseResults) > 0 {
		if err := formatter.FormatTestResults(results.CaseResults); err != nil {
			return err
		}
	}

	return nil
}
