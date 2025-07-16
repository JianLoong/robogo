package types

type Step struct {
	Name    string                 `yaml:"name"`
	Action  string                 `yaml:"action"`
	Args    []interface{}          `yaml:"args"`
	Options map[string]interface{} `yaml:"options,omitempty"`
	Result  string                 `yaml:"result,omitempty"`
	If      string                 `yaml:"if,omitempty"`
	For     string                 `yaml:"for,omitempty"`
	While   string                 `yaml:"while,omitempty"`
}
