package parser

import (
	"time"
)

// TestSuite represents a collection of test cases with shared configuration
type TestSuite struct {
	Name        string                 `yaml:"testsuite"`
	Description string                 `yaml:"description,omitempty"`
	Variables   *Variables             `yaml:"variables,omitempty"`
	Setup       []Step                 `yaml:"setup,omitempty"`
	TestCases   []TestCaseReference    `yaml:"testcases"`
	Teardown    []Step                 `yaml:"teardown,omitempty"`
	Parallel    bool                   `yaml:"parallel,omitempty"`
	Options     map[string]interface{} `yaml:"options,omitempty"`
	FailFast    bool                   `yaml:"fail_fast,omitempty"`
}

// TestCaseReference represents a test case file reference with optional overrides
type TestCaseReference struct {
	File      string                 `yaml:"file"`
	Variables *Variables             `yaml:"variables,omitempty"`
	Options   map[string]interface{} `yaml:"options,omitempty"`
}

// TestSuiteResult represents the result of running a test suite
type TestSuiteResult struct {
	TestSuite    *TestSuite
	Status       string
	Duration     time.Duration
	TotalCases   int
	PassedCases  int
	FailedCases  int
	SkippedCases int
	// Step-level aggregation
	TotalSteps     int
	PassedSteps    int
	FailedSteps    int
	SkippedSteps   int
	CaseResults    []TestCaseResult
	SetupStatus    string
	TeardownStatus string
	ErrorMessage   string
}

// TestCaseResult represents the result of a single test case within a suite
type TestCaseResult struct {
	File     string
	TestCase *TestCase
	Result   *TestResult
	Status   string // "passed", "failed", "skipped"
	Duration time.Duration
	Error    string
}
