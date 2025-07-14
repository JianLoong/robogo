package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
)

// HTTP action - handles web requests
func httpAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("http action requires at least 2 arguments: method and URL")
	}

	method := fmt.Sprintf("%v", args[0])
	url := fmt.Sprintf("%v", args[1])

	var body string
	if len(args) > 2 {
		body = fmt.Sprintf("%v", args[2])
	}

	// Create request
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers from options
	if headers, ok := options["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	// Set timeout
	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse JSON response for easier data access
	var parsedBody interface{}
	respBodyStr := string(responseBody)
	if json.Unmarshal(responseBody, &parsedBody) == nil {
		// JSON parsing successful - include both raw and parsed
		result := map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        respBodyStr,
			"json":        parsedBody, // Parsed JSON for easy access
			"headers":     resp.Header,
		}
		return result, nil
	} else {
		// Not JSON - return as string
		result := map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        respBodyStr,
			"headers":     resp.Header,
		}
		return result, nil
	}
}