package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// httpAction performs an HTTP request. It always returns status code, headers, and raw body.
func httpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {

	if len(args) < 2 {
		return types.MissingArgsError("http", 2, len(args))
	}

	// Check for unresolved variables in critical arguments (method and URL)
	if errorResult := validateArgsResolved("http", args[:2]); errorResult != nil {
		return *errorResult
	}

	method := fmt.Sprintf("%v", args[0])
	url := fmt.Sprintf("%v", args[1])

	// Extract request headers for context first (needed for body processing)
	var requestHeaders map[string]string
	if headers, ok := options["headers"].(map[string]any); ok {
		requestHeaders = make(map[string]string)
		for key, value := range headers {
			requestHeaders[key] = fmt.Sprintf("%v", value)
		}
	}

	var bodyReader io.Reader
	if len(args) > 2 {
		// Get the body argument
		bodyArg := args[2]

		// Always convert the body to a string first
		var bodyStr string

		// Check content type for special handling
		contentType := ""
		if headers, ok := options["headers"].(map[string]any); ok {
			for k, v := range headers {
				if strings.ToLower(k) == "content-type" {
					contentType = strings.ToLower(fmt.Sprintf("%v", v))
					break
				}
			}
		}

		// Special handling for JSON content type with map/slice data
		if (contentType == "application/json" || strings.HasPrefix(contentType, "application/json")) &&
			(isMap(bodyArg) || isSlice(bodyArg)) {
			// For JSON content type with structured data, serialize it
			jsonData, err := json.Marshal(bodyArg)
			if err != nil {
				return types.RequestError("JSON marshaling", err.Error())
			}
			bodyStr = string(jsonData)
		} else {
			// For all other cases, just convert to string
			bodyStr = fmt.Sprintf("%v", bodyArg)
		}

		// Create the body reader from the string
		bodyReader = strings.NewReader(bodyStr)

		// Log the request body for debugging if debug option is set
		if debugOpt, ok := options["debug"].(bool); ok && debugOpt {
			fmt.Printf("HTTP Request Body: %s\n", bodyStr)
		}
	}

	// Extract timeout for context
	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return types.RequestError(fmt.Sprintf("HTTP %s %s", method, url), err.Error())
	}

	if headers, ok := options["headers"].(map[string]any); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(req)

	if err != nil {
		return types.RequestError(fmt.Sprintf("HTTP %s %s", method, url), err.Error())
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.RequestError(fmt.Sprintf("HTTP %s %s response read", method, url), err.Error())
	}

	respBodyStr := string(responseBody)
	result := map[string]any{
		"status_code": resp.StatusCode,
		"body":        respBodyStr,
		"headers":     resp.Header,
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   result,
	}
}

// Helper functions to check types
func isMap(v any) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	kind := t.Kind()
	return kind == reflect.Map
}

func isSlice(v any) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	kind := t.Kind()
	return kind == reflect.Slice || kind == reflect.Array
}
