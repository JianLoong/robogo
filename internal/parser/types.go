package parser

import "time"

// TestCase represents a single test case
type TestCase struct {
	Name        string        `yaml:"testcase"`
	Description string        `yaml:"description,omitempty"`
	Steps       []Step        `yaml:"steps"`
	Timeout     time.Duration `yaml:"timeout,omitempty"`
}

// Step represents a single test step
// 'name' is optional but strongly recommended for clarity and reporting
type Step struct {
	Name   string        `yaml:"name,omitempty"`
	Action string        `yaml:"action"`
	Args   []interface{} `yaml:"args"`
	Result string        `yaml:"result,omitempty"`
}

// TestResult represents the result of running a test case
type TestResult struct {
	TestCase     TestCase
	Status       string
	Duration     time.Duration
	TotalSteps   int
	PassedSteps  int
	FailedSteps  int
	StepResults  []StepResult
	ErrorMessage string
}

// StepResult represents the result of a single step
type StepResult struct {
	Step      Step
	Status    string
	Duration  time.Duration
	Output    string
	Error     string
	Timestamp time.Time
} 