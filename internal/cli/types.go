package cli

import (
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// RunOptions contains configuration for test execution
type RunOptions struct {
	OutputFormat     string
	ParallelEnabled  bool
	MaxConcurrency   int
	VariableDebug    bool
	Silent           bool
	ParallelConfig   *parser.ParallelConfig
}

// RunResults aggregates execution results
type RunResults struct {
	SuiteResults []*parser.TestSuiteResult
	CaseResults  []*parser.TestResult
}

// IsEmpty returns true if no results were collected
func (r *RunResults) IsEmpty() bool {
	return len(r.SuiteResults) == 0 && len(r.CaseResults) == 0
}

// HasOnlySuites returns true if only suite results exist
func (r *RunResults) HasOnlySuites() bool {
	return len(r.SuiteResults) > 0 && len(r.CaseResults) == 0
}

// HasOnlyCases returns true if only case results exist
func (r *RunResults) HasOnlyCases() bool {
	return len(r.CaseResults) > 0 && len(r.SuiteResults) == 0
}

// HasMixed returns true if both suites and cases exist
func (r *RunResults) HasMixed() bool {
	return len(r.SuiteResults) > 0 && len(r.CaseResults) > 0
}

// GrandTotal calculates aggregate statistics
type GrandTotal struct {
	TotalCases   int
	PassedCases  int
	FailedCases  int
	SkippedCases int
	Duration     time.Duration
}

// CalculateGrandTotal computes totals from suite results
func (r *RunResults) CalculateGrandTotal() GrandTotal {
	var total GrandTotal
	for _, sr := range r.SuiteResults {
		total.TotalCases += sr.TotalCases
		total.PassedCases += sr.PassedCases
		total.FailedCases += sr.FailedCases
		total.SkippedCases += sr.SkippedCases
		total.Duration += sr.Duration
	}
	return total
}