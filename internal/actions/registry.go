package actions

import (
	"fmt"
	"sort"
	"strings"
)

// ActionRegistry manages available actions with proper metadata
type ActionRegistry struct {
	actions map[string]Action
}

// NewActionRegistry creates a new action registry and registers all built-in actions
func NewActionRegistry() *ActionRegistry {
	registry := &ActionRegistry{
		actions: make(map[string]Action),
	}

	// Register all built-in actions
	registry.registerBuiltinActions()

	return registry
}

// Register adds an action to the registry
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

// Get retrieves an action by name
func (ar *ActionRegistry) Get(name string) (Action, bool) {
	action, exists := ar.actions[name]
	return action, exists
}

// Execute executes an action by name
func (ar *ActionRegistry) Execute(name string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	action, exists := ar.actions[name]
	if !exists {
		return nil, fmt.Errorf("unknown action: %s", name)
	}

	return action.Execute(args, options, silent)
}

// List returns all registered actions
func (ar *ActionRegistry) List() []ActionMetadata {
	var metadata []ActionMetadata
	for _, action := range ar.actions {
		metadata = append(metadata, action.GetMetadata())
	}

	// Sort by name
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Name < metadata[j].Name
	})

	return metadata
}

// ListByCategory returns actions filtered by category
func (ar *ActionRegistry) ListByCategory(category string) []ActionMetadata {
	var metadata []ActionMetadata
	for _, action := range ar.actions {
		actionMeta := action.GetMetadata()
		if actionMeta.Category == category {
			metadata = append(metadata, actionMeta)
		}
	}

	// Sort by name
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Name < metadata[j].Name
	})

	return metadata
}

// Search returns actions matching a prefix
func (ar *ActionRegistry) Search(prefix string) []ActionMetadata {
	var matches []ActionMetadata
	for _, action := range ar.actions {
		metadata := action.GetMetadata()
		if strings.HasPrefix(metadata.Name, prefix) {
			matches = append(matches, metadata)
		}
	}

	// Sort by name
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// GetCompletions returns action names for autocomplete
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

// GetValidActions returns a list of valid action names
func (ar *ActionRegistry) GetValidActions() []string {
	var actions []string
	for name := range ar.actions {
		actions = append(actions, name)
	}
	sort.Strings(actions)
	return actions
}

// registerBuiltinActions registers all built-in actions with proper metadata
func (ar *ActionRegistry) registerBuiltinActions() {
	// Basic actions
	ar.registerBasicActions()

	// HTTP actions
	ar.registerHTTPActions()

	// Database actions
	ar.registerDatabaseActions()

	// Control flow actions
	ar.registerControlActions()

	// Variable actions
	ar.registerVariableActions()

	// TDM actions
	ar.registerTDMActions()

	// Messaging actions
	ar.registerMessagingActions()

	// Template actions
	ar.registerTemplateActions()

	// Utility actions
	ar.registerUtilityActions()
}

// registerBasicActions registers basic utility actions
func (ar *ActionRegistry) registerBasicActions() {
	ar.Register(NewActionWithContext(LogAction, ActionMetadata{
		Name:        "log",
		Description: "Log a message",
		Example:     `- action: log\n  args: ["message"]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "message", Type: ParamTypeString, Required: true, Description: "Message to log"},
		},
	}))

	ar.Register(NewActionWithContext(SleepAction, ActionMetadata{
		Name:        "sleep",
		Description: "Sleep for a duration",
		Example:     `- action: sleep\n  args: [2]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "duration", Type: ParamTypeDuration, Required: true, Description: "Duration to sleep"},
		},
	}))

	ar.Register(NewActionWithContext(AssertAction, ActionMetadata{
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

	ar.Register(NewActionWithContext(GetTimeAction, ActionMetadata{
		Name:        "get_time",
		Description: "Get current timestamp with optional format (iso, datetime, date, time, timestamp, unix, unix_ms, or custom Go format)",
		Example:     `- action: get_time\n  args: ["iso"]\n  result: timestamp`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "format", Type: ParamTypeString, Required: false, Description: "Time format", Default: "iso"},
		},
		Returns: "timestamp string",
	}))

	ar.Register(NewActionWithContext(GetRandomAction, ActionMetadata{
		Name:        "get_random",
		Description: "Get a random number",
		Example:     `- action: get_random\n  args: [100]\n  result: random_number`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "max", Type: ParamTypeNumber, Required: true, Description: "Maximum value (exclusive)"},
		},
		Returns: "random number",
	}))

	ar.Register(NewActionWithContext(ConcatAction, ActionMetadata{
		Name:        "concat",
		Description: "Concatenate strings",
		Example:     `- action: concat\n  args: ["Hello", " ", "World"]\n  result: message`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "strings", Type: ParamTypeArray, Required: true, Description: "Strings to concatenate"},
		},
		Returns: "concatenated string",
	}))

	ar.Register(NewActionWithContext(LengthAction, ActionMetadata{
		Name:        "length",
		Description: "Get length of string or array",
		Example:     `- action: length\n  args: ["Hello World"]\n  result: str_length`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "value", Type: ParamTypeString, Required: true, Description: "String or array to measure"},
		},
		Returns: "length as number",
	}))

	ar.Register(NewActionWithContext(SkipAction, ActionMetadata{
		Name:        "skip",
		Description: "Skip a test case with an optional reason",
		Example:     `- action: skip\n  args: ["Skipping this test case"]`,
		Category:    CategoryBasic,
		Parameters: []ParameterInfo{
			{Name: "reason", Type: ParamTypeString, Required: false, Description: "Reason for skipping"},
		},
	}))
}

// registerHTTPActions registers HTTP-related actions
func (ar *ActionRegistry) registerHTTPActions() {
	ar.Register(NewActionWithContext(HTTPAction, ActionMetadata{
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

	ar.Register(NewActionWithContext(HTTPGetAction, ActionMetadata{
		Name:        "http_get",
		Description: "Perform HTTP GET request",
		Example:     `- action: http_get\n  args: ["https://api.example.com/users"]\n  result: response`,
		Category:    CategoryHTTP,
		Parameters: []ParameterInfo{
			{Name: "url", Type: ParamTypeURL, Required: true, Description: "Target URL"},
			{Name: "headers", Type: ParamTypeObject, Required: false, Description: "Request headers"},
		},
		Returns: "HTTP response",
	}))

	ar.Register(NewActionWithContext(HTTPPostAction, ActionMetadata{
		Name:        "http_post",
		Description: "Perform HTTP POST request",
		Example:     `- action: http_post\n  args: ["https://api.example.com/users", '{"name": "John"}']\n  result: response`,
		Category:    CategoryHTTP,
		Parameters: []ParameterInfo{
			{Name: "url", Type: ParamTypeURL, Required: true, Description: "Target URL"},
			{Name: "body", Type: ParamTypeString, Required: false, Description: "Request body"},
			{Name: "headers", Type: ParamTypeObject, Required: false, Description: "Request headers"},
		},
		Returns: "HTTP response",
	}))

	ar.Register(NewActionWithContext(HTTPBatchActionWithContext, ActionMetadata{
		Name:        "http_batch",
		Description: "Perform multiple HTTP requests in parallel with concurrency control",
		Example:     `- action: http_batch\n  args: ["GET", ["https://api1.com", "https://api2.com"], {"concurrency": 5}]\n  result: batch_response`,
		Category:    CategoryHTTP,
		Parameters: []ParameterInfo{
			{Name: "method", Type: ParamTypeString, Required: true, Description: "HTTP method"},
			{Name: "urls", Type: ParamTypeArray, Required: true, Description: "Array of URLs"},
			{Name: "options", Type: ParamTypeObject, Required: false, Description: "Request options"},
		},
		Returns: "batch response",
	}))
}

// registerDatabaseActions registers database-related actions
func (ar *ActionRegistry) registerDatabaseActions() {
	ar.Register(NewActionWithContext(PostgresAction, ActionMetadata{
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

	ar.Register(NewActionWithContext(SpannerAction, ActionMetadata{
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

// registerControlActions registers control flow actions
func (ar *ActionRegistry) registerControlActions() {
	ar.Register(NewActionWithContext(ControlFlowAction, ActionMetadata{
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

// registerVariableActions registers variable management actions
func (ar *ActionRegistry) registerVariableActions() {
	ar.Register(NewActionWithContext(VariableAction, ActionMetadata{
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

// registerTDMActions registers Test Data Management actions
func (ar *ActionRegistry) registerTDMActions() {
	ar.Register(NewActionWithContext(TDMAction, ActionMetadata{
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

// registerMessagingActions registers messaging-related actions
func (ar *ActionRegistry) registerMessagingActions() {
	ar.Register(NewActionWithContext(RabbitMQActionWrapper, ActionMetadata{
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

	ar.Register(NewActionWithContext(KafkaActionWrapper, ActionMetadata{
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

// registerTemplateActions registers template-related actions
func (ar *ActionRegistry) registerTemplateActions() {
	ar.Register(NewActionWithContext(TemplateAction, ActionMetadata{
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

// registerUtilityActions registers utility actions
func (ar *ActionRegistry) registerUtilityActions() {
	// Add any additional utility actions here
}
