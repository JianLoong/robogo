package actions

import (
	"encoding/json"
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
func httpAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	fmt.Println("[DEBUG] Entered httpAction")
	if len(args) < 2 {
		msg := "http action requires at least 2 arguments: method and URL"
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
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
		msg := fmt.Sprintf("failed to create request: %v", err)
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	if headers, ok := options["headers"].(map[string]interface{}); ok {
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
		msg := fmt.Sprintf("HTTP request failed: %v", err)
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read response body: %v", err)
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	respBodyStr := string(responseBody)
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        respBodyStr,
		"headers":     resp.Header,
	}

	parseJSON, _ := options["parse_json"].(bool)
	if parseJSON {
		var parsedBody interface{}
		if err := json.Unmarshal(responseBody, &parsedBody); err == nil {
			result["json"] = parsedBody
		}
	}

	output := fmt.Sprintf("HTTP %s %s -> %d", method, url, resp.StatusCode)
	return types.ActionResult{
		Status: types.ActionStatusSuccess,
		Data:   result,
		Output: output,
	}, nil
}
