package runner

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/parser"
)

// DefaultValidationContext implements ValidationContext
type DefaultValidationContext struct {
	testCase         *parser.TestCase
	testSuite        *parser.TestSuite
	currentStep      *parser.Step
	stepIndex        int
	phase            ValidationPhase
	contextData      map[string]interface{}
	variables        map[string]interface{}
	availableActions []string
}

// NewValidationContext creates a new validation context
func NewValidationContext() *DefaultValidationContext {
	return &DefaultValidationContext{
		contextData:      make(map[string]interface{}),
		variables:        make(map[string]interface{}),
		availableActions: []string{"http", "assert", "log", "variable", "postgres", "kafka", "rabbitmq"},
	}
}

func (ctx *DefaultValidationContext) WithTestCase(testCase *parser.TestCase) *DefaultValidationContext {
	ctx.testCase = testCase
	if testCase != nil && testCase.Variables.Regular != nil {
		for k, v := range testCase.Variables.Regular {
			ctx.variables[k] = v
		}
	}
	return ctx
}

func (ctx *DefaultValidationContext) WithTestSuite(testSuite *parser.TestSuite) *DefaultValidationContext {
	ctx.testSuite = testSuite
	return ctx
}

func (ctx *DefaultValidationContext) WithCurrentStep(step *parser.Step) *DefaultValidationContext {
	ctx.currentStep = step
	return ctx
}

func (ctx *DefaultValidationContext) WithPhase(phase ValidationPhase) *DefaultValidationContext {
	ctx.phase = phase
	return ctx
}

func (ctx *DefaultValidationContext) WithAvailableActions(actions []string) *DefaultValidationContext {
	ctx.availableActions = actions
	return ctx
}

func (ctx *DefaultValidationContext) GetTestCase() *parser.TestCase {
	return ctx.testCase
}

func (ctx *DefaultValidationContext) GetTestSuite() *parser.TestSuite {
	return ctx.testSuite
}

func (ctx *DefaultValidationContext) GetStep(index int) *parser.Step {
	if ctx.testCase == nil || index < 0 || index >= len(ctx.testCase.Steps) {
		return nil
	}
	return &ctx.testCase.Steps[index]
}

func (ctx *DefaultValidationContext) GetCurrentStep() *parser.Step {
	return ctx.currentStep
}

func (ctx *DefaultValidationContext) GetStepIndex() int {
	return ctx.stepIndex
}

func (ctx *DefaultValidationContext) GetVariable(name string) (interface{}, bool) {
	value, exists := ctx.variables[name]
	return value, exists
}

func (ctx *DefaultValidationContext) GetAvailableActions() []string {
	return ctx.availableActions
}

func (ctx *DefaultValidationContext) HasCircularDependency(steps []parser.Step) bool {
	// Build dependency graph
	dependencies := make(map[string][]string)
	variables := make(map[string]int) // variable -> step index that defines it

	for i, step := range steps {
		if step.Result != "" {
			variables[step.Result] = i
		}

		stepDeps := ctx.GetStepDependencies(step)
		if len(stepDeps) > 0 {
			dependencies[fmt.Sprintf("step_%d", i)] = stepDeps
		}
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range dependencies[node] {
			depKey := fmt.Sprintf("var_%s", dep)
			if !visited[depKey] {
				if hasCycle(depKey) {
					return true
				}
			} else if recStack[depKey] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range dependencies {
		if !visited[node] {
			if hasCycle(node) {
				return true
			}
		}
	}

	return false
}

func (ctx *DefaultValidationContext) GetStepDependencies(step parser.Step) []string {
	var dependencies []string

	// Extract variable references from arguments
	for _, arg := range step.Args {
		if str, ok := arg.(string); ok {
			deps := extractVariableReferences(str)
			dependencies = append(dependencies, deps...)
		}
	}

	// Extract from conditional statements
	if step.If != nil {
		deps := extractVariableReferences(step.If.Condition)
		dependencies = append(dependencies, deps...)
	}

	if step.For != nil {
		deps := extractVariableReferences(step.For.Condition)
		dependencies = append(dependencies, deps...)
	}

	if step.While != nil {
		deps := extractVariableReferences(step.While.Condition)
		dependencies = append(dependencies, deps...)
	}

	return dependencies
}


func (ctx *DefaultValidationContext) GetFieldValue(path string) (interface{}, bool) {
	// TODO: Implement field path resolution
	return nil, false
}

func (ctx *DefaultValidationContext) IsFieldRequired(path string) bool {
	// TODO: Implement field requirement checking
	return false
}

func (ctx *DefaultValidationContext) GetValidationPhase() ValidationPhase {
	return ctx.phase
}

func (ctx *DefaultValidationContext) GetContextData() map[string]interface{} {
	return ctx.contextData
}