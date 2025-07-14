package internal

import "time"

// Core types - keep it simple
type TestCase struct {
	Name        string          `yaml:"testcase"`
	Description string          `yaml:"description,omitempty"`
	Steps       []Step          `yaml:"steps"`
	Variables   TestVariables   `yaml:"variables,omitempty"`
}

type TestVariables struct {
	Vars map[string]interface{} `yaml:"vars,omitempty"`
}

type Step struct {
	Name    string                 `yaml:"name"`
	Action  string                 `yaml:"action"`
	Args    []interface{}          `yaml:"args"`
	Options map[string]interface{} `yaml:"options,omitempty"`
	Result  string                 `yaml:"result,omitempty"`
}

type TestSuite struct {
	Name      string   `yaml:"testsuite"`
	TestCases []string `yaml:"testcases"`
}

type TestResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Steps    []StepResult  `json:"steps"`
	Error    string        `json:"error,omitempty"`
}

type StepResult struct {
	Name     string        `json:"name"`
	Action   string        `json:"action"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ActionFunc is defined in actions_registry.go to avoid duplication