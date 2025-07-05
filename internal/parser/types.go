package parser

import "time"

// TestCase represents a single test case
type TestCase struct {
	Name        string        `yaml:"testcase"`
	Description string        `yaml:"description,omitempty"`
	Steps       []Step        `yaml:"steps"`
	Timeout     time.Duration `yaml:"timeout,omitempty"`
	Variables   Variables     `yaml:"variables,omitempty"`
}

// Variables represents test case variables
type Variables struct {
	Regular map[string]interface{} `yaml:"vars,omitempty"`    // Regular variables
	Secrets map[string]Secret      `yaml:"secrets,omitempty"` // Secret variables
}

// Secret represents a secret variable
// Supports inline value or file-based secret
// If both value and file are set, value takes precedence
type Secret struct {
	Value      string `yaml:"value,omitempty"`
	File       string `yaml:"file,omitempty"`
	MaskOutput bool   `yaml:"mask_output,omitempty"`
}

// Step represents a single test step
// 'name' is optional but strongly recommended for clarity and reporting
type Step struct {
	Name   string        `yaml:"name,omitempty"`
	Action string        `yaml:"action"`
	Args   []interface{} `yaml:"args"`
	Result string        `yaml:"result,omitempty"`

	// Control flow fields
	If    *ConditionalBlock `yaml:"if,omitempty"`    // If statement
	For   *LoopBlock        `yaml:"for,omitempty"`   // For loop
	While *LoopBlock        `yaml:"while,omitempty"` // While loop
}

// ConditionalBlock represents an if/else block
type ConditionalBlock struct {
	Condition string `yaml:"condition"`      // Condition to evaluate
	Then      []Step `yaml:"then"`           // Steps to execute if true
	Else      []Step `yaml:"else,omitempty"` // Steps to execute if false
}

// LoopBlock represents a for or while loop
type LoopBlock struct {
	Condition     string `yaml:"condition"`                // For: range/array, While: condition
	Steps         []Step `yaml:"steps"`                    // Steps to execute in loop
	MaxIterations int    `yaml:"max_iterations,omitempty"` // Prevent infinite loops
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
