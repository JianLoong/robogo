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
		return types.NewErrorResult("log action requires at least 1 argument")
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}

	message := strings.Join(parts, " ")
	fmt.Println(message)
	os.Stdout.Sync() // Flush output immediately

	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   message,
		Output: message,
	}, nil
}
