package actions

import "github.com/JianLoong/robogo/internal/common"

// Action function signature
type ActionFunc func(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error)

// Action registry - centralized registration of all actions
var ActionRegistry = map[string]ActionFunc{
	// Core actions
	"assert":   assertAction,
	"log":      logAction,
	"variable": variableAction,
	
	// HTTP actions
	"http": httpAction,
	
	// Database actions
	"postgres": postgresAction,
	"spanner":  spannerAction,
	
	// Messaging actions
	"kafka":    kafkaAction,
	"rabbitmq": rabbitmqAction,
	
}

// Helper function to get action
func GetAction(name string) (ActionFunc, bool) {
	action, exists := ActionRegistry[name]
	return action, exists
}

// List all available actions
func ListActions() []string {
	actions := make([]string, 0, len(ActionRegistry))
	for name := range ActionRegistry {
		actions = append(actions, name)
	}
	return actions
}

// No persistent connections to cleanup - all connections are closed after each operation