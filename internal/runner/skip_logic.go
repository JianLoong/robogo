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
// Supports: bool, string (with variable substitution), and nil
func (tr *TestRunner) EvaluateSkip(skipCondition interface{}, context string) SkipInfo {
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
		// Substitute variables in the skip condition
		substitutedSkip := tr.substituteString(v)

		// For now, treat any non-empty string as skip=true
		// In the future, this could be enhanced to evaluate conditions like:
		// "${environment} == 'prod'" -> evaluate as boolean expression
		if substitutedSkip != "" {
			return SkipInfo{ShouldSkip: true, Reason: substitutedSkip}
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

// ShouldSkipStep determines if a step should be skipped
func (tr *TestRunner) ShouldSkipStep(step parser.Step, context string) SkipInfo {
	return tr.EvaluateSkip(step.Skip, context)
}

// ShouldSkipTestCase determines if a test case should be skipped
func (tr *TestRunner) ShouldSkipTestCase(testCase *parser.TestCase, context string) SkipInfo {
	return tr.EvaluateSkip(testCase.Skip, context)
}

// PrintSkipMessage prints a consistent skip message
func PrintSkipMessage(itemType, itemName, reason string, silent bool) {
	if silent {
		return
	}

	if reason == "" || reason == "(no reason provided)" {
		fmt.Printf("⏭️  %s skipped: %s\n", itemType, itemName)
	} else {
		fmt.Printf("⏭️  %s skipped: %s | Reason: %s\n", itemType, itemName, reason)
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
