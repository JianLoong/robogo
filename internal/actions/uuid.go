package actions

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
	"github.com/google/uuid"
)

func uuidAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	id := uuid.New().String()
	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   id,
		Output: id,
	}, nil
}
