package runner

import (
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// SkipInfo contains information about whether something should be skipped
type SkipInfo struct {
	ShouldSkip bool
	Reason     string
}

// EvaluateSkip evaluates a skip condition and returns SkipInfo
// Supports: bool, string, and nil
func EvaluateSkip(skipCondition interface{}) SkipInfo {
	if skipCondition == nil {
		return SkipInfo{ShouldSkip: false, Reason: ""}
	}

	switch v := skipCondition.(type) {
	case bool:
		if v {
			return SkipInfo{ShouldSkip: true, Reason: "(no reason provided)"}
		}
		return SkipInfo{ShouldSkip: false, Reason: ""}
	case string:
		// For non-empty strings, treat as skip condition
		// Note: Variable substitution should be done by caller before calling this function
		if v != "" && v != "false" && v != "0" {
			return SkipInfo{ShouldSkip: true, Reason: v}
		}
		return SkipInfo{ShouldSkip: false, Reason: ""}
	default:
		// Unknown type - convert to string and evaluate
		strValue := fmt.Sprintf("%v", v)
		if strValue != "" && strValue != "false" && strValue != "0" {
			return SkipInfo{ShouldSkip: true, Reason: strValue}
		}
		return SkipInfo{ShouldSkip: false, Reason: ""}
	}
}

// PrintSkipMessage prints a consistent skip message
func PrintSkipMessage(itemType, itemName, reason string, silent bool) {
	if silent {
		return
	}

	if reason == "" || reason == "(no reason provided)" {
		fmt.Printf("%s skipped: %s\n", itemType, itemName)
	} else {
		fmt.Printf("%s skipped: %s | Reason: %s\n", itemType, itemName, reason)
	}
}

// CreateSkipResult creates a step result for a skipped step
func CreateSkipResult(step parser.Step, reason string) parser.StepResult {
	return parser.StepResult{
		Step:      step,
		Status:    parser.StatusSkipped,
		Duration:  0,
		Output:    "",
		Error:     reason,
		Timestamp: time.Now(),
	}
}

// CreateSkipTestCaseResult creates a test case result for a skipped test case
func CreateSkipTestCaseResult(testCase *parser.TestCase, reason string) parser.TestCaseResult {
	return parser.TestCaseResult{
		TestCase: testCase,
		Status:   parser.StatusSkipped,
		Error:    reason,
	}
}
