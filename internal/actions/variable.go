package actions

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func variableAction(args []any, options map[string]any, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 2 {
		return types.NewErrorResult("variable action requires at least 2 arguments")
	}

	name := fmt.Sprintf("%v", args[0])
	value := args[1]

	vars.Set(name, value)

	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   value,
		Output: fmt.Sprintf("%v", value),
	}, nil
}
