package types

import "time"

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
	Duration time.Duration `json:"duration"`
	Result   ActionResult  `json:"result"`
}
