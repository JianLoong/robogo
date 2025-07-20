package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

func logAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) == 0 {
		return types.MissingArgsError("log", 1, 0)
	}

	// Get format option (default: "pretty")
	format := "pretty"
	if f, ok := options["format"]; ok {
		format = fmt.Sprintf("%v", f)
	}

	var unresolvedArgs []int
	parts := make([]string, len(args))

	for i, arg := range args {
		if arg == nil {
			fmt.Printf("[WARN] logAction: argument %d is nil\n", i)
			parts[i] = "<nil>"
			continue
		}
		if str, ok := arg.(string); ok && str == "__UNRESOLVED__" {
			fmt.Printf("[WARN] logAction: argument %d is unresolved\n", i)
			parts[i] = "<unresolved>"
			unresolvedArgs = append(unresolvedArgs, i)
			continue
		}
		parts[i] = formatLogValue(arg, format)
	}

	message := strings.Join(parts, " ")
	fmt.Println(message)
	os.Stdout.Sync() // Flush output immediately

	// Fail if any variables were unresolved for consistency with other actions
	if len(unresolvedArgs) > 0 {
		return types.UnresolvedVariableError(len(unresolvedArgs), unresolvedArgs)
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   message,
	}
}

// formatLogValue formats a value for logging based on the specified format
func formatLogValue(arg any, format string) string {
	// Handle simple types with basic formatting
	if isSimpleType(arg) {
		return fmt.Sprintf("%v", arg)
	}

	// Handle complex types based on format
	switch format {
	case "raw":
		return fmt.Sprintf("%v", arg)
	case "compact":
		if jsonBytes, err := json.Marshal(arg); err == nil {
			return string(jsonBytes)
		}
	case "pretty", "":
		if jsonBytes, err := json.MarshalIndent(arg, "", "  "); err == nil {
			return string(jsonBytes)
		}
	}

	// Fallback to default formatting
	return fmt.Sprintf("%v", arg)
}

// isSimpleType checks if a value is a simple type that doesn't need JSON formatting
func isSimpleType(arg any) bool {
	if arg == nil {
		return true
	}

	switch reflect.TypeOf(arg).Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		 reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	default:
		return false
	}
}
