package actions

import "context"

// Action represents a single action that can be executed
type Action interface {
	// Execute runs the action with the given arguments and options
	Execute(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)

	// ExecuteWithContext runs the action with context for cancellation and timeouts
	ExecuteWithContext(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)

	// GetMetadata returns metadata about the action
	GetMetadata() ActionMetadata
}

// ActionMetadata contains information about an action
type ActionMetadata struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Example     string          `json:"example"`
	Category    string          `json:"category,omitempty"`
	Version     string          `json:"version,omitempty"`
	Deprecated  bool            `json:"deprecated,omitempty"`
	Parameters  []ParameterInfo `json:"parameters,omitempty"`
	Returns     string          `json:"returns,omitempty"`
}

// ParameterInfo describes a parameter for an action
type ParameterInfo struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// ActionFunc is already defined in builtin.go

// ActionFuncWithContext represents a context-aware action function
type ActionFuncWithContext func(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)

// ActionWrapper wraps ActionFunc to implement the Action interface
type ActionWrapper struct {
	fn       ActionFunc
	fnCtx    ActionFuncWithContext // Optional context-aware function
	metadata ActionMetadata
}

// Execute implements the Action interface (backward compatibility)
func (aw *ActionWrapper) Execute(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if aw.fnCtx != nil {
		// Use context-aware function with background context for backward compatibility
		return aw.fnCtx(context.Background(), args, options, silent)
	}
	return aw.fn(args, options, silent)
}

// ExecuteWithContext implements the Action interface
func (aw *ActionWrapper) ExecuteWithContext(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if aw.fnCtx != nil {
		// Use context-aware function
		return aw.fnCtx(ctx, args, options, silent)
	}
	// Fallback to legacy function (ignore context for now)
	return aw.fn(args, options, silent)
}

// GetMetadata implements the Action interface
func (aw *ActionWrapper) GetMetadata() ActionMetadata {
	return aw.metadata
}

// NewAction creates a new Action from a function and metadata
func NewAction(fn ActionFunc, metadata ActionMetadata) Action {
	return &ActionWrapper{
		fn:       fn,
		metadata: metadata,
	}
}

// NewActionWithContext creates a new context-aware Action from a function and metadata
func NewActionWithContext(fn ActionFuncWithContext, metadata ActionMetadata) Action {
	return &ActionWrapper{
		fnCtx:    fn,
		metadata: metadata,
	}
}

// NewDualAction creates an action that supports both legacy and context-aware execution
func NewDualAction(fn ActionFunc, fnCtx ActionFuncWithContext, metadata ActionMetadata) Action {
	return &ActionWrapper{
		fn:       fn,
		fnCtx:    fnCtx,
		metadata: metadata,
	}
}

// ActionCategory represents the category of an action
const (
	CategoryBasic     = "basic"
	CategoryHTTP      = "http"
	CategoryDatabase  = "database"
	CategoryControl   = "control"
	CategoryVariable  = "variable"
	CategoryTDM       = "tdm"
	CategoryMessaging = "messaging"
	CategoryTemplate  = "template"
	CategoryUtility   = "utility"
)

// Common parameter types
const (
	ParamTypeString   = "string"
	ParamTypeNumber   = "number"
	ParamTypeBoolean  = "boolean"
	ParamTypeArray    = "array"
	ParamTypeObject   = "object"
	ParamTypeDuration = "duration"
	ParamTypeURL      = "url"
	ParamTypeFile     = "file"
)
