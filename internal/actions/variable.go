package actions

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func variableAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 2 {
		msg := "variable action requires at least 2 arguments"
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	name := fmt.Sprintf("%v", args[0])
	value := args[1]

	vars.Set(name, value)

	msg := fmt.Sprintf("Set variable %s = %v", name, value)
	return types.ActionResult{
		Status: types.ActionStatusSuccess,
		Data:   msg,
		Output: msg,
	}, nil
}
