package types

type Step struct {
	Name     string         `yaml:"name"`
	Action   string         `yaml:"action,omitempty"`
	Steps    []Step         `yaml:"steps,omitempty"`
	Args     []any          `yaml:"args,omitempty"`
	Options  map[string]any `yaml:"options,omitempty"`
	Result   string         `yaml:"result,omitempty"`
	Extract  *ExtractConfig `yaml:"extract,omitempty"`
	If       string         `yaml:"if,omitempty"`
	For      string         `yaml:"for,omitempty"`
	While    string         `yaml:"while,omitempty"`
	Retry    *RetryConfig   `yaml:"retry,omitempty"`
	Continue bool           `yaml:"continue,omitempty"`
	NoLog           bool     `yaml:"no_log,omitempty"`           // Suppress logging for sensitive steps
	SensitiveFields []string `yaml:"sensitive_fields,omitempty"` // Custom fields to mask in logs and output
	Summary         *bool    `yaml:"summary,omitempty"`          // Include step in summary table (default: true)
}

// ExtractConfig defines data extraction from action results
type ExtractConfig struct {
	Type      string `yaml:"type"`               // "jq", "xpath", "regex", "csv"
	Path      string `yaml:"path"`               // The extraction expression
	Group     int    `yaml:"group,omitempty"`    // For regex: which capture group (default: 1)
	
	// CSV-specific options
	Row       *int   `yaml:"row,omitempty"`      // For csv: specific row index (0-based), nil means not specified
	Column    string `yaml:"column,omitempty"`   // For csv: column name or index
	Delimiter string `yaml:"delimiter,omitempty"` // For csv: field separator (default: ",")
	HasHeader bool   `yaml:"has_header,omitempty"` // For csv: first row contains headers (default: true)
	Filter    string `yaml:"filter,omitempty"`   // For csv: simple filtering expression
}

// RetryConfig defines retry behavior for a step
type RetryConfig struct {
	Attempts      int    `yaml:"attempts"`                  // Number of retry attempts
	Delay         string `yaml:"delay"`                     // Base delay between retries (e.g., "1s", "500ms")
	Backoff       string `yaml:"backoff,omitempty"`         // "fixed", "linear", "exponential"
	StopOnSuccess bool   `yaml:"stop_on_success,omitempty"` // Stop retrying on first success
	RetryIf       string `yaml:"retry_if,omitempty"`        // Condition to determine if retry should continue
	// Can use extracted values, e.g., "${author} == 'Yours Truly'"
	RetryOn []string `yaml:"retry_on,omitempty"` // Specific error types to retry on
	// e.g., ["assertion_failed", "http_error", "timeout"]
}
