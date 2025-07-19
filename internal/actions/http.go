package actions

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// httpAction performs an HTTP request. It always returns status code, headers, and raw body.
// If options["parse_json"] == true and the body is valid JSON, the parsed JSON is included in Data.
func httpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {

	if len(args) < 2 {
		return types.MissingArgsError("http", 2, len(args))
	}

	method := fmt.Sprintf("%v", args[0])
	url := fmt.Sprintf("%v", args[1])

	var body string
	if len(args) > 2 {
		body = fmt.Sprintf("%v", args[2])
	}

	// Extract request headers for context
	var requestHeaders map[string]string
	if headers, ok := options["headers"].(map[string]any); ok {
		requestHeaders = make(map[string]string)
		for key, value := range headers {
			requestHeaders[key] = fmt.Sprintf("%v", value)
		}
	}

	// Extract timeout for context
	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
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
