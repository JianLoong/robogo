package actions

import (
	"sort"
	"strings"
)

// ActionInfo contains information about an action
type ActionInfo struct {
	Name        string
	Description string
	Example     string
}

// ActionRegistry manages available actions
type ActionRegistry struct {
	actions map[string]ActionInfo
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	registry := &ActionRegistry{
		actions: make(map[string]ActionInfo),
	}

	// Register built-in actions
	registry.registerBuiltinActions()

	return registry
}

// registerBuiltinActions registers all built-in actions
func (ar *ActionRegistry) registerBuiltinActions() {
	ar.Register(ActionInfo{
		Name:        "log",
		Description: "Log a message",
		Example:     `- action: log\n  args: ["message"]`,
	})

	ar.Register(ActionInfo{
		Name:        "sleep",
		Description: "Sleep for a duration",
		Example:     `- action: sleep\n  args: [2]`,
	})

	ar.Register(ActionInfo{
		Name:        "assert",
		Description: "Assert a condition using comparison operators (==, !=, >, <, >=, <=, contains, starts_with, ends_with)",
		Example:     `- action: assert\n  args: ["value", ">", "0", "Value should be positive"]`,
	})

	ar.Register(ActionInfo{
		Name:        "get_time",
		Description: "Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)",
		Example:     `- action: get_time\n  args: ["iso"]\n  result: timestamp`,
	})

	ar.Register(ActionInfo{
		Name:        "get_random",
		Description: "Get a random number",
		Example:     `- action: get_random\n  args: [100]\n  result: random_number`,
	})

	ar.Register(ActionInfo{
		Name:        "concat",
		Description: "Concatenate strings",
		Example:     `- action: concat\n  args: ["Hello", " ", "World"]\n  result: message`,
	})

	ar.Register(ActionInfo{
		Name:        "length",
		Description: "Get length of string or array",
		Example:     `- action: length\n  args: ["Hello World"]\n  result: str_length`,
	})

	ar.Register(ActionInfo{
		Name:        "http",
		Description: "Perform HTTP request (GET, POST, PUT, DELETE, etc.). Supports client cert (cert/key) and custom CA (ca) options.",
		Example:     `- action: http\n  args: [\"GET\", \"https://secure.example.com/api\", {\"cert\": \"client.crt\", \"key\": \"client.key\", \"ca\": \"ca.crt\", \"Authorization\": \"Bearer ...\"}]\n  result: response`,
	})

	ar.Register(ActionInfo{
		Name:        "http_get",
		Description: "Perform HTTP GET request",
		Example:     `- action: http_get\n  args: ["https://api.example.com/users"]\n  result: response`,
	})

	ar.Register(ActionInfo{
		Name:        "http_post",
		Description: "Perform HTTP POST request",
		Example:     `- action: http_post\n  args: ["https://api.example.com/users", '{"name": "John"}']\n  result: response`,
	})

	// Control flow actions
	ar.Register(ActionInfo{
		Name:        "control",
		Description: "Control flow operations (if, for, while)",
		Example:     `- action: control\n  args: ["if", "condition"]\n  result: condition_result`,
	})

	// Database actions
	ar.Register(ActionInfo{
		Name:        "postgres",
		Description: "PostgreSQL database operations (query, execute, connect, close)",
		Example:     `- action: postgres\n  args: ["query", "postgres://user:pass@localhost/db", "SELECT * FROM users"]\n  result: query_result`,
	})

	// Variable management actions
	ar.Register(ActionInfo{
		Name:        "variable",
		Description: "Variable management operations (set_variable, get_variable, list_variables)",
		Example:     `- action: variable\n  args: ["set_variable", "my_var", "my_value"]\n  result: set_result`,
	})
}

// Register adds an action to the registry
func (ar *ActionRegistry) Register(info ActionInfo) {
	ar.actions[info.Name] = info
}

// Get retrieves action information
func (ar *ActionRegistry) Get(name string) (ActionInfo, bool) {
	info, exists := ar.actions[name]
	return info, exists
}

// List returns all registered actions
func (ar *ActionRegistry) List() []ActionInfo {
	var actions []ActionInfo
	for _, info := range ar.actions {
		actions = append(actions, info)
	}

	// Sort by name
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Name < actions[j].Name
	})

	return actions
}

// Search returns actions matching a prefix
func (ar *ActionRegistry) Search(prefix string) []ActionInfo {
	var matches []ActionInfo
	for _, info := range ar.actions {
		if strings.HasPrefix(info.Name, prefix) {
			matches = append(matches, info)
		}
	}

	return matches
}

// GetCompletions returns action names for autocomplete
func (ar *ActionRegistry) GetCompletions(partial string) []string {
	var completions []string
	for _, info := range ar.actions {
		if strings.HasPrefix(info.Name, partial) {
			completions = append(completions, info.Name)
		}
	}

	return completions
}
