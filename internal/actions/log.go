package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func logAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) == 0 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "LOG_MISSING_ARGS").
			WithTemplate("log action requires at least 1 argument").
			Build()
	}

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
			continue
		}
		parts[i] = fmt.Sprintf("%v", arg)
	}

	message := strings.Join(parts, " ")
	fmt.Println(message)
	os.Stdout.Sync() // Flush output immediately

	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   message,
	}
}
