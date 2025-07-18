package actions

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// httpAction performs an HTTP request. It always returns status code, headers, and raw body.
// If options["parse_json"] == true and the body is valid JSON, the parsed JSON is included in Data.
func httpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	fmt.Println("[DEBUG] Entered httpAction")
	if len(args) < 2 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "HTTP_MISSING_ARGS").
			WithTemplate("http action requires at least 2 arguments: method and URL").
			Build()
	}

	method := fmt.Sprintf("%v", args[0])
	url := fmt.Sprintf("%v", args[1])

	fmt.Printf("[DEBUG] Preparing HTTP request: %s %s\n", method, url)

	var body string
	if len(args) > 2 {
		body = fmt.Sprintf("%v", args[2])
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "HTTP_REQUEST_CREATE_FAILED").
			WithTemplate("failed to create request: %v").
			WithContext("method", method).
			WithContext("url", url).
			WithContext("error", err.Error()).
			Build(err)
	}

	if headers, ok := options["headers"].(map[string]any); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	client := &http.Client{Timeout: timeout}
	fmt.Println("[DEBUG] Sending HTTP request...")
	resp, err := client.Do(req)
	fmt.Println("[DEBUG] HTTP request completed")
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "HTTP_REQUEST_FAILED").
			WithTemplate("HTTP request failed: %v").
			WithContext("method", method).
			WithContext("url", url).
			WithContext("error", err.Error()).
			Build(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "HTTP_RESPONSE_READ_FAILED").
			WithTemplate("failed to read response body: %v").
			WithContext("method", method).
			WithContext("url", url).
			WithContext("status_code", resp.StatusCode).
			WithContext("error", err.Error()).
			Build(err)
	}

	respBodyStr := string(responseBody)
	result := map[string]any{
		"status_code": resp.StatusCode,
		"body":        respBodyStr,
		"headers":     resp.Header,
	}

	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   result,
	}
}
