package types

type Step struct {
	Name    string         `yaml:"name"`
	Action  string         `yaml:"action"`
	Args    []any          `yaml:"args"`
	Options map[string]any `yaml:"options,omitempty"`
	Result  string         `yaml:"result,omitempty"`
	If      string         `yaml:"if,omitempty"`
	For     string         `yaml:"for,omitempty"`
	While   string         `yaml:"while,omitempty"`
	Retry   *RetryConfig   `yaml:"retry,omitempty"`
}

// RetryConfig defines retry behavior for a step
type RetryConfig struct {
	Attempts    int      `yaml:"attempts"`
	Delay       string   `yaml:"delay"`
	Backoff     string   `yaml:"backoff,omitempty"`     // "fixed", "linear", "exponential"
	RetryOn     []string `yaml:"retry_on,omitempty"`    // Specific error types to retry on
	StopOnSuccess bool   `yaml:"stop_on_success,omitempty"` // Stop retrying on first success
}
