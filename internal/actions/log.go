package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

func logAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) == 0 {
		return types.MissingArgsError("log", 1, 0)
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
		parts[i] = fmt.Sprintf("%v", arg)
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
