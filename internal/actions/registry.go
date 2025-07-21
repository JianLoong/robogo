package actions

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// Action function signature
type ActionFunc func(args []any, options map[string]any, vars *common.Variables) types.ActionResult

// Action registry - centralized registration of all actions
var ActionRegistry = map[string]ActionFunc{
	// Core actions
	"assert":   assertAction,
	"log":      logAction,
	"variable": variableAction,

	// Utility actions
	"uuid":  uuidAction,
	"time":  timeAction,
	"sleep": sleepAction,

	// Encoding actions
	"base64_encode": base64EncodeAction,
	"base64_decode": base64DecodeAction,
	"url_encode":    urlEncodeAction,
	"url_decode":    urlDecodeAction,
	"hash":          hashAction,

	// File actions
	"file_read": fileReadAction,

	// String actions
	"string_random":  stringRandomAction,
	"string_replace": stringReplaceAction,
	"string_format":  stringFormatAction,

	// HTTP actions
	"http": httpAction,

	// Database actions
	"postgres": postgresAction,
	"spanner":  spannerAction,

	// Messaging actions
	"kafka":         kafkaAction,
	"rabbitmq":      rabbitmqAction,
	"swift_message": swiftMessageAction,

	// JSON actions
	"json_build": jsonBuildAction,
	"json_parse": jsonParseAction,

	// XML actions
	"xml_build": xmlBuildAction,
	"xml_parse": xmlParseAction,

	// Query actions
	"jq":    jqAction,
	"xpath": xpathAction,
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
