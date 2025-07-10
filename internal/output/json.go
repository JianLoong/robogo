package output

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// JSONFormatter handles JSON output formatting
type JSONFormatter struct{}

// StepResultWithStr adds human-readable duration to step results
type StepResultWithStr struct {
	parser.StepResult
	DurationStr string `json:"duration_str"`
}

// TestResultWithStr adds human-readable duration to test results
type TestResultWithStr struct {
	*parser.TestResult
	StepResults []StepResultWithStr `json:"step_results"`
}

// FormatTestResults outputs test results in JSON format
func (f *JSONFormatter) FormatTestResults(results []*parser.TestResult) error {
	var resultsWithStr []TestResultWithStr
	for _, r := range results {
		var stepsWithStr []StepResultWithStr
		for _, s := range r.StepResults {
			var durationStr string
			if s.Duration < time.Millisecond {
				durationStr = fmt.Sprintf("%dµs", s.Duration.Microseconds())
			} else {
				durationStr = s.Duration.String()
			}
			stepsWithStr = append(stepsWithStr, StepResultWithStr{
				StepResult:  s,
				DurationStr: durationStr,
			})
		}
		resultsWithStr = append(resultsWithStr, TestResultWithStr{
			TestResult:  r,
			StepResults: stepsWithStr,
		})
	}

	jsonBytes, err := json.MarshalIndent(resultsWithStr, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedSteps > 0 {
			return fmt.Errorf("test suite failed")
		}
	}
	return nil
}

// SuiteSummary represents a test suite summary for JSON output
type SuiteSummary struct {
	Name           string            `json:"suite_name"`
	Status         string            `json:"status"`
	Duration       time.Duration     `json:"duration"`
	TotalCases     int               `json:"total_cases"`
	PassedCases    int               `json:"passed_cases"`
	FailedCases    int               `json:"failed_cases"`
	SkippedCases   int               `json:"skipped_cases"`
	TotalSteps     int               `json:"total_steps"`
	PassedSteps    int               `json:"passed_steps"`
	FailedSteps    int               `json:"failed_steps"`
	SkippedSteps   int               `json:"skipped_steps"`
	SetupStatus    string            `json:"setup_status,omitempty"`
	TeardownStatus string            `json:"teardown_status,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	TestCases      []TestCaseSummary `json:"test_cases"`
}

// TestCaseSummary represents a test case summary for JSON output
type TestCaseSummary struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Steps    []StepSummary `json:"steps"`
}

// StepSummary represents a step summary for JSON output
type StepSummary struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// FormatSuiteResult outputs test suite results in JSON format
func (f *JSONFormatter) FormatSuiteResult(result *parser.TestSuiteResult) error {
	// Step summary calculation
	totalSteps, passedSteps, failedSteps, skippedSteps := 0, 0, 0, 0
	for _, caseResult := range result.CaseResults {
		if caseResult.Result != nil {
			for _, step := range caseResult.Result.StepResults {
				totalSteps++
				switch step.Status {
				case "PASSED", "passed":
					passedSteps++
				case "FAILED", "failed":
					failedSteps++
				case "SKIPPED", "skipped":
					skippedSteps++
				}
			}
		}
	}

	suite := SuiteSummary{
		Name:           result.TestSuite.Name,
		Status:         result.Status,
		Duration:       result.Duration,
		TotalCases:     result.TotalCases,
		PassedCases:    result.PassedCases,
		FailedCases:    result.FailedCases,
		SkippedCases:   result.SkippedCases,
		TotalSteps:     totalSteps,
		PassedSteps:    passedSteps,
		FailedSteps:    failedSteps,
		SkippedSteps:   skippedSteps,
		SetupStatus:    result.SetupStatus,
		TeardownStatus: result.TeardownStatus,
		ErrorMessage:   result.ErrorMessage,
	}

	for _, caseResult := range result.CaseResults {
		steps := []StepSummary{}
		if caseResult.Result != nil {
			for _, step := range caseResult.Result.StepResults {
				steps = append(steps, StepSummary{
					Name:     step.Step.Name,
					Status:   step.Status,
					Error:    step.Error,
					Duration: step.Duration,
				})
			}
		}
		dur := caseResult.Duration
		if dur == 0 && caseResult.Result != nil {
			dur = caseResult.Result.Duration
		}
		suite.TestCases = append(suite.TestCases, TestCaseSummary{
			Name:     caseResult.TestCase.Name,
			Status:   caseResult.Status,
			Error:    caseResult.Error,
			Duration: dur,
			Steps:    steps,
		})
	}

	jsonBytes, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test suite results to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))

	// Exit with non-zero code if test suite failed
	if result.Status == "failed" {
		return fmt.Errorf("test suite failed")
	}
	return nil
}

// FormatMultipleSuites outputs multiple test suite results in JSON format
func (f *JSONFormatter) FormatMultipleSuites(results []*parser.TestSuiteResult, grandTotal GrandTotal) error {
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test suite results to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))

	// Exit with non-zero code if any test failed
	for _, result := range results {
		if result.FailedCases > 0 {
			return fmt.Errorf("test suite failed")
		}
	}
	return nil
}