package loops

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/templates"
	"github.com/JianLoong/robogo/internal/types"
)

// LoopParser handles parsing of loop specifications for different formats
type LoopParser struct{}

// NewLoopParser creates a new loop parser
func NewLoopParser() *LoopParser {
	return &LoopParser{}
}

// ParseIterations parses a range or array specification and returns the iteration values.
// Supports three formats:
// - Range: "1..5" creates iterations [1, 2, 3, 4, 5]
// - Array: "[item1,item2,item3]" creates iterations ["item1", "item2", "item3"]
// - Count: "3" creates iterations [1, 2, 3]
func (parser *LoopParser) ParseIterations(rangeOrArray string, step types.Step) ([]any, *types.StepResult, error) {
	if strings.Contains(rangeOrArray, "..") {
		return parser.ParseRange(rangeOrArray, step)
	} else if strings.HasPrefix(rangeOrArray, "[") && strings.HasSuffix(rangeOrArray, "]") {
		return parser.ParseArray(rangeOrArray, step)
	} else {
		return parser.ParseCount(rangeOrArray, step)
	}
}

// ParseRange parses a range specification like "1..5" and returns integers from start to end inclusive.
// Returns an error result if the range format is invalid or contains non-numeric values.
func (parser *LoopParser) ParseRange(rangeSpec string, step types.Step) ([]any, *types.StepResult, error) {
	parts := strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_RANGE_FORMAT").
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidRangeFormat)).
				WithContext("range_spec", rangeSpec).
				Build(rangeSpec),
		}, fmt.Errorf("invalid range format: %s", rangeSpec)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_START_VALUE").
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidStartValue)).
				WithContext("start_value", parts[0]).
				Build(parts[0]),
		}, fmt.Errorf("invalid start value: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_END_VALUE").
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidEndValue)).
				WithContext("end_value", parts[1]).
				Build(parts[1]),
		}, fmt.Errorf("invalid end value: %s", parts[1])
	}

	var iterations []any
	for i := start; i <= end; i++ {
		iterations = append(iterations, i)
	}
	return iterations, nil, nil
}

// ParseArray parses an array specification like "[item1,item2,item3]" and returns the items as strings.
// Items are trimmed of whitespace and returned in the order they appear.
func (parser *LoopParser) ParseArray(arraySpec string, step types.Step) ([]any, *types.StepResult, error) {
	arrayStr := arraySpec[1 : len(arraySpec)-1]
	items := strings.Split(arrayStr, ",")
	var iterations []any
	for _, item := range items {
		iterations = append(iterations, strings.TrimSpace(item))
	}
	return iterations, nil, nil
}

// ParseCount parses a count specification like "3" and returns integers from 1 to count inclusive.
// Returns an error result if the count is not a valid integer.
func (parser *LoopParser) ParseCount(countSpec string, step types.Step) ([]any, *types.StepResult, error) {
	count, err := strconv.Atoi(countSpec)
	if err != nil {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_COUNT_FORMAT").
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidCountFormat)).
				WithContext("count_spec", countSpec).
				Build(countSpec),
		}, fmt.Errorf("invalid count format: %s", countSpec)
	}

	var iterations []any
	for i := 1; i <= count; i++ {
		iterations = append(iterations, i)
	}
	return iterations, nil, nil
}
