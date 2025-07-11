package output

import (
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// Format represents the output format type
type Format string

const (
	FormatConsole Format = "console"
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

// NewFormatter creates a new formatter based on the format type
func NewFormatter(format Format) Formatter {
	return &ConsoleFormatter{}
}