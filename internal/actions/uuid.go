package actions

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
	"github.com/google/uuid"
)

func uuidAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	id := uuid.New().String()
	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   id,
	}
}
