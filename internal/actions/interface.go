package actions

// Action represents a single action that can be executed
type Action interface {
	// Execute runs the action with the given arguments and options
	Execute(args []interface{}, options map[string]interface{}, silent bool) (string, error)

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

// ActionWrapper wraps ActionFunc to implement the Action interface
type ActionWrapper struct {
	fn       ActionFunc
	metadata ActionMetadata
}

// Execute implements the Action interface
func (aw *ActionWrapper) Execute(args []interface{}, options map[string]interface{}, silent bool) (string, error) {
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
