package actions

import (
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// ActionFunc defines the signature for action functions
type ActionFunc func(args []any, options map[string]any, vars *common.Variables) types.ActionResult

// ActionRegistry manages action registration and lookup without global state
type ActionRegistry struct {
	actions map[string]ActionFunc
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	registry := &ActionRegistry{
		actions: make(map[string]ActionFunc),
	}

	// Register all built-in actions
	registry.registerBuiltinActions()

	return registry
}

// Register registers a new action
func (registry *ActionRegistry) Register(name string, action ActionFunc) {
	registry.actions[name] = action
}

// Get retrieves an action by name
func (registry *ActionRegistry) Get(name string) (ActionFunc, bool) {
	action, exists := registry.actions[name]
	return action, exists
}

// Has checks if an action exists
func (registry *ActionRegistry) Has(name string) bool {
	_, exists := registry.actions[name]
	return exists
}

// GetRegisteredActions returns a list of all registered action names
func (registry *ActionRegistry) GetRegisteredActions() []string {
	names := make([]string, 0, len(registry.actions))
	for name := range registry.actions {
		names = append(names, name)
	}
	return names
}

// Unregister removes an action (useful for testing)
func (registry *ActionRegistry) Unregister(name string) {
	delete(registry.actions, name)
}

// Clone creates a copy of the registry
func (registry *ActionRegistry) Clone() *ActionRegistry {
	newRegistry := NewActionRegistry()
	// Clear the built-ins and copy from original
	newRegistry.actions = make(map[string]ActionFunc)
	for name, action := range registry.actions {
		newRegistry.actions[name] = action
	}
	return newRegistry
}

// registerBuiltinActions registers all built-in actions (based on existing registry)
func (registry *ActionRegistry) registerBuiltinActions() {
	// Core actions
	registry.Register("assert", assertAction)
	registry.Register("log", logAction)
	registry.Register("variable", variableAction)

	// Utility actions
	registry.Register("uuid", uuidAction)
	registry.Register("time", timeAction)
	registry.Register("sleep", sleepAction)
	registry.Register("ping", pingAction)

	// Security actions
	registry.Register("ssl_cert_check", sslCertCheckAction)

	// Encoding actions
	registry.Register("base64_encode", base64EncodeAction)
	registry.Register("base64_decode", base64DecodeAction)
	registry.Register("url_encode", urlEncodeAction)
	registry.Register("url_decode", urlDecodeAction)
	registry.Register("hash", hashAction)

	// File actions
	registry.Register("file_read", fileReadAction)
	registry.Register("scp", scpAction)

	// String actions
	registry.Register("string_random", stringRandomAction)
	registry.Register("string_replace", stringReplaceAction)
	registry.Register("string_format", stringFormatAction)
	registry.Register("string", stringAction)

	// Data processing actions
	registry.Register("jq", jqAction)
	registry.Register("xpath", xpathAction)

	// HTTP actions
	registry.Register("http", httpAction)

	// Database actions
	registry.Register("postgres", postgresAction)
	registry.Register("spanner", spannerAction)

	// Messaging actions
	registry.Register("kafka", kafkaAction)
	registry.Register("rabbitmq", rabbitmqAction)
	registry.Register("swift_message", swiftMessageAction)

	// JSON/XML/CSV actions
	registry.Register("json_parse", jsonParseAction)
	registry.Register("json_build", jsonBuildAction)
	registry.Register("xml_parse", xmlParseAction)
	registry.Register("xml_build", xmlBuildAction)
	registry.Register("csv_parse", csvParseAction)
}

// validateArgsResolved checks if any arguments contain unresolved variables
// Returns an ActionResult error if unresolved variables are found, nil otherwise
func validateArgsResolved(actionName string, args []any) *types.ActionResult {
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			if strings.Contains(str, "__UNRESOLVED") {
				errorResult := types.NewErrorBuilder(types.ErrorCategoryVariable, "UNRESOLVED_VARIABLE").
					WithTemplate("Action failed due to unresolved variable in argument").
					WithContext("action", actionName).
					WithContext("unresolved_value", str).
					WithContext("argument_index", i).
					Build(fmt.Sprintf("unresolved variable in %s argument: %s", actionName, str))
				return &errorResult
			}
		}
	}
	return nil
}
