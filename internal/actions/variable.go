package actions

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

func variableAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.MissingArgsError("variable", 2, len(args))
	}

	name := fmt.Sprintf("%v", args[0])
	value := args[1]

	vars.Set(name, value)

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   value,
	}
}
