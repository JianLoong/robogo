package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func logAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) == 0 {
		msg := "log action requires at least 1 argument"
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}

	message := strings.Join(parts, " ")
	fmt.Println(message)
	os.Stdout.Sync() // Flush output immediately

	return types.ActionResult{
		Status: types.ActionStatusSuccess,
		Data:   message,
		Output: message,
	}, nil
}
