package types

import (
	"time"
)

// StepContext provides comprehensive context information for step execution
type StepContext struct {
	StepNumber  int                    `json:"step_number"`
	StepName    string                 `json:"step_name"`
	ActionName  string                 `json:"action_name"`
	LoopContext *LoopContext           `json:"loop_context,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Arguments   []interface{}          `json:"arguments,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Conditions  *ConditionContext      `json:"conditions,omitempty"`
}

// LoopContext provides information about loop execution context
type LoopContext struct {
	Type          string      `json:"type"`                     // "for", "while"
	Iteration     int         `json:"iteration"`                // 1-based iteration number
	Index         int         `json:"index,omitempty"`          // 0-based index (for "for" loops)
	Item          interface{} `json:"item,omitempty"`           // Current item (for "for" loops)
	Condition     string      `json:"condition,omitempty"`      // Loop condition
	MaxIterations int         `json:"max_iterations,omitempty"` // Maximum allowed iterations
}

// ConditionContext provides information about conditional execution
type ConditionContext struct {
	IfCondition    string `json:"if_condition,omitempty"`
	ForCondition   string `json:"for_condition,omitempty"`
	WhileCondition string `json:"while_condition,omitempty"`
	ResultVariable string `json:"result_variable,omitempty"`
}

// NewStepContext creates a new StepContext with basic information
func NewStepContext(stepNumber int, stepName, actionName string) *StepContext {
	return &StepContext{
		StepNumber: stepNumber,
		StepName:   stepName,
		ActionName: actionName,
		Timestamp:  time.Now(),
		Variables:  make(map[string]interface{}),
	}
}

// WithLoopContext adds loop context information
func (sc *StepContext) WithLoopContext(loopCtx *LoopContext) *StepContext {
	sc.LoopContext = loopCtx
	return sc
}

// WithArguments adds argument information
func (sc *StepContext) WithArguments(args []interface{}) *StepContext {
	sc.Arguments = args
	return sc
}

// WithOptions adds option information
func (sc *StepContext) WithOptions(options map[string]interface{}) *StepContext {
	sc.Options = options
	return sc
}

// WithConditions adds condition information
func (sc *StepContext) WithConditions(conditions *ConditionContext) *StepContext {
	sc.Conditions = conditions
	return sc
}

// WithVariables adds variable snapshot
func (sc *StepContext) WithVariables(variables map[string]interface{}) *StepContext {
	sc.Variables = variables
	return sc
}

// NewForLoopContext creates a LoopContext for a for loop
func NewForLoopContext(iteration, index int, item interface{}, condition string) *LoopContext {
	return &LoopContext{
		Type:      "for",
		Iteration: iteration,
		Index:     index,
		Item:      item,
		Condition: condition,
	}
}

// NewWhileLoopContext creates a LoopContext for a while loop
func NewWhileLoopContext(iteration int, condition string, maxIterations int) *LoopContext {
	return &LoopContext{
		Type:          "while",
		Iteration:     iteration,
		Condition:     condition,
		MaxIterations: maxIterations,
	}
}

// NewConditionContext creates a ConditionContext with the provided conditions
func NewConditionContext(ifCond, forCond, whileCond, resultVar string) *ConditionContext {
	return &ConditionContext{
		IfCondition:    ifCond,
		ForCondition:   forCond,
		WhileCondition: whileCond,
		ResultVariable: resultVar,
	}
}

// ToMap converts StepContext to a map for error context enrichment
func (sc *StepContext) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"step_number": sc.StepNumber,
		"step_name":   sc.StepName,
		"action_name": sc.ActionName,
		"timestamp":   sc.Timestamp,
	}

	if len(sc.Arguments) > 0 {
		result["arguments"] = sc.Arguments
	}

	if len(sc.Options) > 0 {
		result["options"] = sc.Options
	}

	if len(sc.Variables) > 0 {
		result["variables"] = sc.Variables
	}

	if sc.LoopContext != nil {
		result["loop_context"] = map[string]interface{}{
			"type":      sc.LoopContext.Type,
			"iteration": sc.LoopContext.Iteration,
		}

		if sc.LoopContext.Index >= 0 {
			result["loop_context"].(map[string]interface{})["index"] = sc.LoopContext.Index
		}

		if sc.LoopContext.Item != nil {
			result["loop_context"].(map[string]interface{})["item"] = sc.LoopContext.Item
		}

		if sc.LoopContext.Condition != "" {
			result["loop_context"].(map[string]interface{})["condition"] = sc.LoopContext.Condition
		}

		if sc.LoopContext.MaxIterations > 0 {
			result["loop_context"].(map[string]interface{})["max_iterations"] = sc.LoopContext.MaxIterations
		}
	}

	if sc.Conditions != nil {
		conditions := make(map[string]interface{})

		if sc.Conditions.IfCondition != "" {
			conditions["if_condition"] = sc.Conditions.IfCondition
		}

		if sc.Conditions.ForCondition != "" {
			conditions["for_condition"] = sc.Conditions.ForCondition
		}

		if sc.Conditions.WhileCondition != "" {
			conditions["while_condition"] = sc.Conditions.WhileCondition
		}

		if sc.Conditions.ResultVariable != "" {
			conditions["result_variable"] = sc.Conditions.ResultVariable
		}

		if len(conditions) > 0 {
			result["conditions"] = conditions
		}
	}

	return result
}

