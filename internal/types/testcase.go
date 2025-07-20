package types

// Only keep the correct, single definition of TestCase and TestVariables here.
type TestCase struct {
	Name        string        `yaml:"testcase"`
	Description string        `yaml:"description,omitempty"`
	Setup       []Step        `yaml:"setup,omitempty"`
	Steps       []Step        `yaml:"steps"`
	Teardown    []Step        `yaml:"teardown,omitempty"`
	Variables   TestVariables `yaml:"variables,omitempty"`
}

type TestVariables struct {
	Vars map[string]any `yaml:"vars,omitempty"`
}
