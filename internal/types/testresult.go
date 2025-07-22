package types

import "time"

type TestResult struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"`
	Duration     time.Duration `json:"duration"`
	SetupSteps   []StepResult  `json:"setup_steps,omitempty"`
	Steps        []StepResult  `json:"steps"`
	TeardownSteps []StepResult `json:"teardown_steps,omitempty"`
	ErrorInfo    *ErrorInfo    `json:"error_info,omitempty"`
}

type StepResult struct {
	Name     string        `json:"name"`
	Action   string        `json:"action"`
	Duration time.Duration `json:"duration"`
	Result   ActionResult  `json:"result"`
}

// GetMessage returns the error message from ErrorInfo
func (tr *TestResult) GetMessage() string {
	if tr.ErrorInfo != nil {
		return tr.ErrorInfo.Message
	}
	return ""
}

// SetError sets the ErrorInfo for the test result
func (tr *TestResult) SetError(errorInfo *ErrorInfo) {
	tr.ErrorInfo = errorInfo
}
