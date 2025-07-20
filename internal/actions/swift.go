package actions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// swiftMessageAction generates a SWIFT message from a template file and data map.
func swiftMessageAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.MissingArgsError("swift_message", 2, len(args))
	}

	templateFile, ok := args[0].(string)
	if !ok {
		return types.InvalidArgError("swift_message", "template file name", "string")
	}

	dataMap, ok := args[1].(map[string]any)
	if !ok {
		return types.InvalidArgError("swift_message", "data map", "map[string]any")
	}

	templatePath := filepath.Join("templates", "swift", templateFile)
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			ErrorInfo: &types.ErrorInfo{
				Category: types.ErrorCategorySystem,
				Code:     "SWIFT_TEMPLATE_READ",
				Message:  fmt.Sprintf("Failed to read template file: %v", err),
			},
		}
	}

	tmpl, err := template.New(templateFile).Parse(string(tmplBytes))
	if err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			ErrorInfo: &types.ErrorInfo{
				Category: types.ErrorCategoryValidation,
				Code:     "SWIFT_TEMPLATE_PARSE",
				Message:  fmt.Sprintf("Failed to parse template: %v", err),
			},
		}
	}

	// Merge dataMap with current variables (vars)
	merged := make(map[string]any)
	for k, v := range vars.GetSnapshot() {
		merged[k] = v
	}
	for k, v := range dataMap {
		merged[k] = v
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, merged)
	if err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			ErrorInfo: &types.ErrorInfo{
				Category: types.ErrorCategoryExecution,
				Code:     "SWIFT_TEMPLATE_EXEC",
				Message:  fmt.Sprintf("Failed to execute template: %v", err),
			},
		}
	}

	result := buf.String()
	// Removed output_var support; result is returned in ActionResult.Data only
	return types.NewSuccessResultWithData(result)
}
