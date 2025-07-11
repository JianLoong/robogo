package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// ActionRegistry manages available actions and their metadata.
type ActionRegistry struct {
	actions map[string]Action
}

// NewActionRegistry creates a new ActionRegistry and registers all built-in actions.
func NewActionRegistry() *ActionRegistry {
	registry := &ActionRegistry{
		actions: make(map[string]Action),
	}
	registry.registerBuiltinActions()
	return registry
}

// Register adds an action to the registry. Returns an error if the action name is empty or already registered.
func (ar *ActionRegistry) Register(action Action) error {
	metadata := action.GetMetadata()
	if metadata.Name == "" {
		return fmt.Errorf("action must have a name")
	}
	if _, exists := ar.actions[metadata.Name]; exists {
		return fmt.Errorf("action '%s' is already registered", metadata.Name)
	}
	ar.actions[metadata.Name] = action
	return nil
}

// Get retrieves an action by name. Returns the action and a boolean indicating existence.
func (ar *ActionRegistry) Get(name string) (Action, bool) {
	action, exists := ar.actions[name]
	return action, exists
}

// Execute runs an action by name with the provided context and arguments.
func (ar *ActionRegistry) Execute(ctx context.Context, name string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	action, exists := ar.actions[name]
	if !exists {
		return nil, fmt.Errorf("unknown action: %s", name)
	}
	return action.ExecuteWithContext(ctx, args, options, silent)
}

// List returns metadata for all registered actions, sorted by name.
func (ar *ActionRegistry) List() []ActionMetadata {
	var metadata []ActionMetadata
	for _, action := range ar.actions {
		metadata = append(metadata, action.GetMetadata())
	}
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Name < metadata[j].Name
	})
	return metadata
}

// ListByCategory returns metadata for actions in the specified category, sorted by name.
func (ar *ActionRegistry) ListByCategory(category string) []ActionMetadata {
	var metadata []ActionMetadata
	for _, action := range ar.actions {
		actionMeta := action.GetMetadata()
		if actionMeta.Category == category {
			metadata = append(metadata, actionMeta)
		}
	}
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Name < metadata[j].Name
	})
	return metadata
}

// Search returns metadata for actions whose names start with the given prefix, sorted by name.
func (ar *ActionRegistry) Search(prefix string) []ActionMetadata {
	var matches []ActionMetadata
	for _, action := range ar.actions {
		metadata := action.GetMetadata()
		if strings.HasPrefix(metadata.Name, prefix) {
			matches = append(matches, metadata)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})
	return matches
}

// GetCompletions returns a sorted list of action names that start with the given partial string.
func (ar *ActionRegistry) GetCompletions(partial string) []string {
	var completions []string
	for _, action := range ar.actions {
		metadata := action.GetMetadata()
		if strings.HasPrefix(metadata.Name, partial) {
			completions = append(completions, metadata.Name)
		}
	}
	sort.Strings(completions)
	return completions
}

// GetValidActions returns a sorted list of all valid action names.
func (ar *ActionRegistry) GetValidActions() []string {
	var actions []string
	for name := range ar.actions {
		actions = append(actions, name)
	}
	sort.Strings(actions)
	return actions
}

// registerBuiltinActions registers all built-in actions with their metadata.
func (ar *ActionRegistry) registerBuiltinActions() {
	ar.registerBasicActions()
	ar.registerHTTPActions()
	ar.registerDatabaseActions()
	ar.registerControlActions()
	ar.registerVariableActions()
	ar.registerTDMActions()
	ar.registerMessagingActions()
	ar.registerTemplateActions()
	ar.registerUtilityActions()
}

// The following methods register groups of related actions. Each action is registered with its metadata.
func (ar *ActionRegistry) registerBasicActions() {
	ar.Register(NewAction(LogAction, ActionMetadata{
		Name:        "log",
		Description: "Log a message",
		Example:     `- action: log\n  args: ["message"]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "message", Type: ParamTypeString, Required: true, Description: "Message to log"},
		},
	}))

	ar.Register(NewAction(SleepAction, ActionMetadata{
		Name:        "sleep",
		Description: "Sleep for a duration",
		Example:     `- action: sleep\n  args: [2]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "duration", Type: ParamTypeDuration, Required: true, Description: "Duration to sleep"},
		},
	}))

	ar.Register(NewAction(AssertAction, ActionMetadata{
		Name:        "assert",
		Description: "Assert a condition using comparison operators (==, !=, >, <, >=, <=, contains, starts_with, ends_with)",
		Example:     `- action: assert\n  args: ["value", ">", "0", "Value should be positive"]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "actual", Type: ParamTypeString, Required: true, Description: "Actual value"},
			{Name: "operator", Type: ParamTypeString, Required: true, Description: "Comparison operator"},
			{Name: "expected", Type: ParamTypeString, Required: true, Description: "Expected value"},
			{Name: "message", Type: ParamTypeString, Required: false, Description: "Assertion message"},
		},
	}))

	ar.Register(NewAction(GetTimeAction, ActionMetadata{
		Name:        "get_time",
		Description: "Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)",
		Example:     `- action: get_time\n  args: ["iso"]\n  result: timestamp`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "format", Type: ParamTypeString, Required: false, Description: "Time format", Default: "iso"},
		},
		Returns: "timestamp string",
	}))

	ar.Register(NewAction(GetRandomAction, ActionMetadata{
		Name:        "get_random",
		Description: "Get a random number",
		Example:     `- action: get_random\n  args: [100]\n  result: random_number`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "max", Type: ParamTypeNumber, Required: true, Description: "Maximum value (exclusive)"},
		},
		Returns: "random number",
	}))

	ar.Register(NewAction(ConcatAction, ActionMetadata{
		Name:        "concat",
		Description: "Concatenate strings",
		Example:     `- action: concat\n  args: ["Hello", " ", "World"]\n  result: message`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "strings", Type: ParamTypeArray, Required: true, Description: "Strings to concatenate"},
		},
		Returns: "concatenated string",
	}))

	ar.Register(NewAction(LengthAction, ActionMetadata{
		Name:        "length",
		Description: "Get length of string or array",
		Example:     `- action: length\n  args: ["Hello World"]\n  result: str_length`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "value", Type: ParamTypeString, Required: true, Description: "String or array to measure"},
		},
		Returns: "length as number",
	}))

	ar.Register(NewAction(SkipAction, ActionMetadata{
		Name:        "skip",
		Description: "Skip a test case with an optional reason",
		Example:     `- action: skip\n  args: ["Skipping this test case"]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "reason", Type: ParamTypeString, Required: false, Description: "Reason for skipping"},
		},
	}))

	ar.Register(NewAction(
		func(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
			return bytesToStringAction(ctx, args, options, silent)
		},
		ActionMetadata{
			Name:        "bytes_to_string",
			Description: "Convert a byte slice or any value to a string.",
			Example:     `- action: bytes_to_string\n  args: ["${some_bytes}"]\n  result: my_string`,
			Category:    CategoryBasic,
			Parameters: []ParameterInfo{
				{Name: "value", Type: ParamTypeString, Required: true, Description: "Value to convert to string (bytes or any)"},
			},
		},
	))

	ar.Register(NewAction(
		func(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
			return jsonExtractAction(ctx, args, options, silent)
		},
		ActionMetadata{
			Name:        "json_extract",
			Description: "Extract a value by key from a JSON string.",
			Example:     `- action: json_extract\n  args: ["${json_str}", "data"]\n  result: extracted_value`,
			Category:    CategoryBasic,
			Parameters: []ParameterInfo{
				{Name: "json", Type: ParamTypeString, Required: true, Description: "JSON string to extract from"},
				{Name: "key", Type: ParamTypeString, Required: true, Description: "Key to extract"},
			},
		},
	))
}

// The following register*Actions methods register actions for their respective domains.
// Each action is registered with its metadata for discoverability and documentation.
func (ar *ActionRegistry) registerHTTPActions() {
	ar.Register(NewAction(HTTPAction, ActionMetadata{
		Name:        "http",
		Description: "Perform HTTP request (GET, POST, PUT, DELETE, etc.). Supports client cert (cert/key) and custom CA (ca) options.",
		Example:     `- action: http\n  args: ["GET", "https://secure.example.com/api", {"cert": "client.crt", "key": "client.key", "ca": "ca.crt", "Authorization": "Bearer ..."}]\n  result: response`,
		Category:    CategoryHTTP,
		Parameters: []ParameterInfo{
			{Name: "method", Type: ParamTypeString, Required: true, Description: "HTTP method"},
			{Name: "url", Type: ParamTypeURL, Required: true, Description: "Target URL"},
			{Name: "headers", Type: ParamTypeObject, Required: false, Description: "Request headers"},
			{Name: "body", Type: ParamTypeString, Required: false, Description: "Request body"},
		},
		Returns: "HTTP response",
	}))
}

func (ar *ActionRegistry) registerDatabaseActions() {
	ar.Register(NewAction(PostgresAction, ActionMetadata{
		Name:        "postgres",
		Description: "PostgreSQL database operations (query, execute, connect, close, batch)",
		Example:     `- action: postgres\n  args: ["query", "postgres://user:pass@localhost/db", "SELECT * FROM users"]\n  result: query_result`,
		Category:    CategoryDatabase,
		Parameters: []ParameterInfo{
			{Name: "operation", Type: ParamTypeString, Required: true, Description: "Database operation"},
			{Name: "connection", Type: ParamTypeString, Required: true, Description: "Database connection string"},
			{Name: "query", Type: ParamTypeString, Required: false, Description: "SQL query"},
			{Name: "params", Type: ParamTypeArray, Required: false, Description: "Query parameters"},
		},
		Returns: "database result",
	}))

	ar.Register(NewAction(SpannerAction, ActionMetadata{
		Name:        "spanner",
		Description: "Google Cloud Spanner operations (connect, query, execute, close) with emulator support",
		Example:     `- action: spanner\n  args: ["connect", "projects/robogo-test-project/instances/robogo-test-instance/databases/robogo-test-db?useEmulator=true"]\n  result: connection_result`,
		Category:    CategoryDatabase,
		Parameters: []ParameterInfo{
			{Name: "operation", Type: ParamTypeString, Required: true, Description: "Spanner operation"},
			{Name: "connection", Type: ParamTypeString, Required: true, Description: "Spanner connection string"},
			{Name: "query", Type: ParamTypeString, Required: false, Description: "SQL query"},
			{Name: "params", Type: ParamTypeArray, Required: false, Description: "Query parameters"},
		},
		Returns: "spanner result",
	}))
}

func (ar *ActionRegistry) registerControlActions() {
	ar.Register(NewAction(ControlFlowAction, ActionMetadata{
		Name:        "control",
		Description: "Control flow operations (if, for, while)",
		Example:     `- action: control\n  args: ["if", "condition"]\n  result: condition_result`,
		Category:    CategoryControl,
		Parameters: []ParameterInfo{
			{Name: "type", Type: ParamTypeString, Required: true, Description: "Control flow type"},
			{Name: "condition", Type: ParamTypeString, Required: true, Description: "Condition to evaluate"},
		},
		Returns: "control result",
	}))
}

func (ar *ActionRegistry) registerVariableActions() {
	ar.Register(NewAction(VariableAction, ActionMetadata{
		Name:        "variable",
		Description: "Variable management operations (set_variable, get_variable, list_variables)",
		Example:     `- action: variable\n  args: ["set_variable", "my_var", "my_value"]\n  result: set_result`,
		Category:    CategoryVariable,
		Parameters: []ParameterInfo{
			{Name: "operation", Type: ParamTypeString, Required: true, Description: "Variable operation"},
			{Name: "name", Type: ParamTypeString, Required: false, Description: "Variable name"},
			{Name: "value", Type: ParamTypeString, Required: false, Description: "Variable value"},
		},
		Returns: "variable result",
	}))
}

func (ar *ActionRegistry) registerTDMActions() {
	ar.Register(NewAction(TDMAction, ActionMetadata{
		Name:        "tdm",
		Description: "Test Data Management operations (generate, validate, load_dataset, set_environment)",
		Example:     `- action: tdm\n  args: ["generate", "user_{index}", 5]\n  result: generated_data`,
		Category:    CategoryTDM,
		Parameters: []ParameterInfo{
			{Name: "operation", Type: ParamTypeString, Required: true, Description: "TDM operation"},
			{Name: "template", Type: ParamTypeString, Required: false, Description: "Data template"},
			{Name: "count", Type: ParamTypeNumber, Required: false, Description: "Number of records"},
		},
		Returns: "TDM result",
	}))
}

func (ar *ActionRegistry) registerMessagingActions() {
	ar.Register(NewAction(RabbitMQAction, ActionMetadata{
		Name:        "rabbitmq",
		Description: "RabbitMQ operations (connect, publish, consume, close)",
		Example:     `- action: rabbitmq\n  args: ["connect", "amqp://guest:guest@localhost:5672/", "my_connection"]`,
		Category:    CategoryMessaging,
		Parameters: []ParameterInfo{
			{Name: "operation", Type: ParamTypeString, Required: true, Description: "RabbitMQ operation"},
			{Name: "connection", Type: ParamTypeString, Required: true, Description: "Connection string"},
			{Name: "name", Type: ParamTypeString, Required: false, Description: "Connection name"},
		},
		Returns: "RabbitMQ result",
	}))

	ar.Register(NewAction(KafkaAction, ActionMetadata{
		Name:        "kafka",
		Description: "Kafka operations (publish, consume) with configurable options",
		Example:     `- action: kafka\n  args: ["publish", "localhost:9092", "test-topic", "message", {"acks": "all", "compression": "snappy"}]\n  result: publish_result`,
		Category:    CategoryMessaging,
		Parameters: []ParameterInfo{
			{Name: "operation", Type: ParamTypeString, Required: true, Description: "Kafka operation"},
			{Name: "broker", Type: ParamTypeString, Required: true, Description: "Kafka broker address"},
			{Name: "topic", Type: ParamTypeString, Required: false, Description: "Kafka topic"},
			{Name: "message", Type: ParamTypeString, Required: false, Description: "Message to publish"},
		},
		Returns: "Kafka result",
	}))
}

func (ar *ActionRegistry) registerTemplateActions() {
	ar.Register(NewAction(TemplateAction, ActionMetadata{
		Name:        "template",
		Description: "Render templates using Go template engine with inline template definitions",
		Example:     `- action: template\n  args: ["mt103", {"transaction_id": "123", "amount": "100.00"}]\n  result: rendered_message`,
		Category:    CategoryTemplate,
		Parameters: []ParameterInfo{
			{Name: "template_name", Type: ParamTypeString, Required: true, Description: "Template name"},
			{Name: "data", Type: ParamTypeObject, Required: true, Description: "Template data"},
		},
		Returns: "rendered template",
	}))
}

func (ar *ActionRegistry) registerUtilityActions() {
	// Add any additional utility actions here
}
