package types

// LoopContext provides information about loop execution context
type LoopContext struct {
	Type      string // "for" or "while"
	Iteration int    // Current iteration number (0-based)
	Index     int    // Current index in array/slice (for "for" loops)
	Item      any    // Current item value (for "for" loops)
}

// NewForLoopContext creates a loop context for "for" loops
func NewForLoopContext(iteration, index int, item any) *LoopContext {
	return &LoopContext{
		Type:      "for",
		Iteration: iteration,
		Index:     index,
		Item:      item,
	}
}

// NewWhileLoopContext creates a loop context for "while" loops
func NewWhileLoopContext(iteration int) *LoopContext {
	return &LoopContext{
		Type:      "while",
		Iteration: iteration,
	}
}