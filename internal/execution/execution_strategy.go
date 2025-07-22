package execution

import "github.com/JianLoong/robogo/internal/types"

// ExecutionStrategy defines different ways to execute steps
type ExecutionStrategy interface {
	// Execute runs the step according to this strategy
	// All errors are encoded within the StepResult's ActionResult
	Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) *types.StepResult
	
	// CanHandle returns true if this strategy can handle the given step
	CanHandle(step types.Step) bool
	
	// Priority returns the priority of this strategy (higher = more specific)
	Priority() int
}