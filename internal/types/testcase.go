package types

// Only keep the correct, single definition of TestCase and TestVariables here.
type TestCase struct {
	Name        string        `yaml:"testcase"`
	Description string        `yaml:"description,omitempty"`
	Steps       []Step        `yaml:"steps"`
	Variables   TestVariables `yaml:"variables,omitempty"`
}

type TestVariables struct {
	Vars map[string]interface{} `yaml:"vars,omitempty"`
}
