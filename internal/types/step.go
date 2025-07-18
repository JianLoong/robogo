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
}
