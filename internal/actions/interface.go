package actions

import "context"

// Action represents a single action that can be executed
type Action interface {
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

// ActionFuncWithContext represents a context-aware action function
type ActionFuncWithContext func(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)

// ActionWrapper wraps ActionFuncWithContext to implement the Action interface
type ActionWrapper struct {
	fn       ActionFuncWithContext
	metadata ActionMetadata
}

// ExecuteWithContext implements the Action interface
func (aw *ActionWrapper) ExecuteWithContext(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return aw.fn(ctx, args, options, silent)
}

// GetMetadata implements the Action interface
func (aw *ActionWrapper) GetMetadata() ActionMetadata {
	return aw.metadata
}

// NewAction creates a new context-aware Action from a function and metadata
func NewAction(fn ActionFuncWithContext, metadata ActionMetadata) Action {
	return &ActionWrapper{
		fn:       fn,
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
