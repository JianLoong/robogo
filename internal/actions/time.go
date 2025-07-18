package actions

import (
	"fmt"
	"strconv"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func timeAction(args []any, options map[string]any, vars *common.Variables) (types.ActionResult, error) {
	format := "2006-01-02T15:04:05Z07:00" // RFC3339 format
	if len(args) > 0 {
		format = fmt.Sprintf("%v", args[0])
	}

	var timestamp string
	if format == "Unix" {
		// Handle Unix timestamp (seconds since epoch)
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	} else {
		// Use Go time format
		timestamp = time.Now().Format(format)
	}

	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   timestamp,
	}, nil
}
