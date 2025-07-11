package output

import (
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// GrandTotal represents aggregated results from multiple test suites
type GrandTotal struct {
	TotalCases   int
	PassedCases  int
	FailedCases  int
	SkippedCases int
	Duration     time.Duration
}

// Formatter interface for different output formats
type Formatter interface {
	FormatTestResults(results []*parser.TestResult) error
	FormatSuiteResult(result *parser.TestSuiteResult) error
	FormatMultipleSuites(results []*parser.TestSuiteResult, grandTotal GrandTotal) error
}

// NewFormatter creates a new console formatter
func NewFormatter() Formatter {
	return &ConsoleFormatter{}
}