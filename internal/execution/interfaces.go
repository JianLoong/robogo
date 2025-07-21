package execution



// ConditionEvaluator defines the interface for evaluating conditions
type ConditionEvaluator interface {
	// Evaluate evaluates a condition expression and returns true/false
	Evaluate(condition string) (bool, error)
}

