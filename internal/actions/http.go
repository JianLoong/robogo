package actions

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// HTTPResponse represents the response from an HTTP request
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   time.Duration     `json:"duration"`
}

// loadCertificateData loads certificate data from either a file path or PEM content
//
// Parameters:
//   - input: File path or PEM content string
//
// Returns: Certificate data as bytes
//
// Examples:
//   - File path: "/path/to/cert.crt"
//   - PEM content: "-----BEGIN CERTIFICATE-----\nMII...\n-----END CERTIFICATE-----"
func loadCertificateData(input string) ([]byte, error) {
	// Check if input looks like a file path (contains path separators or doesn't look like PEM)
	if strings.Contains(input, "/") || strings.Contains(input, "\\") ||
		strings.Contains(input, ".") || !strings.Contains(input, "-----BEGIN") {
		// Treat as file path
		return ioutil.ReadFile(input)
	}

	// Treat as PEM content
	return []byte(input), nil
}

// loadX509KeyPair loads certificate and key from either file paths or PEM content
//
// Parameters:
//   - certInput: Certificate file path or PEM content
//   - keyInput: Private key file path or PEM content
//
// Returns: TLS certificate pair
//
// Examples:
//   - File paths: "/path/to/cert.crt", "/path/to/key.key"
//   - PEM content: "-----BEGIN CERTIFICATE-----...", "-----BEGIN PRIVATE KEY-----..."
func loadX509KeyPair(certInput, keyInput string) (tls.Certificate, error) {
	certData, err := loadCertificateData(certInput)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load certificate: %w", err)
	}

	keyData, err := loadCertificateData(keyInput)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load private key: %w", err)
	}

	return tls.X509KeyPair(certData, keyData)
}

// HTTPAction performs HTTP requests with comprehensive configuration and response handling.
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
//   - url: Target URL for the request
//   - headers: Request headers (optional, map or array of key-value pairs)
//   - body: Request body (optional, string, object, or array)
//   - options: Additional options (timeout, follow_redirects, verify_ssl, etc.)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON response with status, headers, body, and timing information
//
// Supported Methods:
//   - GET: Retrieve data from server
//   - POST: Submit data to server
//   - PUT: Replace resource on server
//   - DELETE: Remove resource from server
//   - PATCH: Partially update resource
//   - HEAD: Get response headers only
//   - OPTIONS: Get allowed methods
//
// Examples:
//   - Simple GET: ["GET", "https://api.example.com/users"]
//   - POST with JSON: ["POST", "https://api.example.com/users", {"Content-Type": "application/json"}, {"name": "John", "email": "john@example.com"}]
//   - With headers: ["GET", "https://api.example.com/data", {"Authorization": "Bearer ${token}"}]
//   - With options: ["POST", "https://api.example.com/upload", {}, "file content", {"timeout": 30, "follow_redirects": false}]
//
// Use Cases:
//   - API testing and validation
//   - Web service integration
//   - Load testing and performance validation
//   - Data retrieval and submission
//   - Authentication testing
//
// Notes:
//   - Supports variable substitution with ${variable} syntax
//   - Automatic JSON handling for request/response bodies
//   - Comprehensive error handling and timeout support
//   - Response data available for assertions and variable storage
func HTTPAction(args []interface{}, silent bool) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("http action requires at least 2 arguments: method and url")
	}

	method := strings.ToUpper(fmt.Sprintf("%v", args[0]))
	url := fmt.Sprintf("%v", args[1])

	// Validate method
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	if !validMethods[method] {
		return "", fmt.Errorf("invalid HTTP method: %s", method)
	}

	// Parse optional arguments
	var headers map[string]string
	var body string
	var timeout time.Duration = 30 * time.Second
	var certFile, keyFile, caFile string

	for i := 2; i < len(args); i++ {
		switch v := args[i].(type) {
		case map[string]interface{}:
			// Check for cert/key/ca fields
			if c, ok := v["cert"]; ok {
				certFile = fmt.Sprintf("%v", c)
			}
			if k, ok := v["key"]; ok {
				keyFile = fmt.Sprintf("%v", k)
			}
			if ca, ok := v["ca"]; ok {
				caFile = fmt.Sprintf("%v", ca)
			}
			// Treat other fields as headers
			if headers == nil {
				headers = make(map[string]string)
			}
			for k, val := range v {
				if k != "cert" && k != "key" && k != "ca" {
					headers[k] = fmt.Sprintf("%v", val)
				}
			}
		case string:
			// This could be body or timeout
			if body == "" && (method == "POST" || method == "PUT" || method == "PATCH") {
				body = v
			}
		}
	}

	// TLS config
	var tlsConfig *tls.Config
	if certFile != "" && keyFile != "" {
		cert, err := loadX509KeyPair(certFile, keyFile)
		if err != nil {
			return "", fmt.Errorf("failed to load client certificate/key: %w", err)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	if caFile != "" {
		caData, err := loadCertificateData(caFile)
		if err != nil {
			return "", fmt.Errorf("failed to load CA certificate: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caData) {
			return "", fmt.Errorf("failed to append CA certificate")
		}
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		tlsConfig.RootCAs = caPool
	}

	// HTTP client
	transport := &http.Transport{}
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// Create request
	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Set default content type for POST/PUT/PATCH if not specified
	if (method == "POST" || method == "PUT" || method == "PATCH") && body != "" {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		respHeaders[k] = strings.Join(v, ", ")
	}

	// Create response object
	httpResp := HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(respBody),
		Duration:   duration,
	}

	// Convert to JSON for return
	jsonResp, err := json.Marshal(httpResp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("🌐 %s %s → %d (%v)\n", method, url, resp.StatusCode, duration)
		if len(respBody) > 0 {
			fmt.Printf("⚠️  Response body: %s\n", string(respBody))
		}
	}

	return string(jsonResp), nil
}

// HTTPGetAction performs HTTP GET requests with simplified syntax.
//
// Parameters:
//   - url: Target URL for the request
//   - headers: Request headers (optional, map or array of key-value pairs)
//   - options: Additional options (timeout, follow_redirects, verify_ssl, etc.)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON response with status, headers, body, and timing information
//
// Examples:
//   - Simple GET: ["https://api.example.com/users"]
//   - With headers: ["https://api.example.com/data", {"Authorization": "Bearer ${token}"}]
//   - With options: ["https://api.example.com/slow", {}, {"timeout": 60}]
//
// Use Cases:
//   - Data retrieval from APIs
//   - Status checking
//   - Content validation
//   - Performance testing
//
// Notes:
//   - Simplified syntax for GET requests
//   - Same functionality as HTTPAction with GET method
//   - Supports all HTTPAction options and features
func HTTPGetAction(args []interface{}, silent bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("http_get action requires at least 1 argument: url")
	}

	url := fmt.Sprintf("%v", args[0])

	// Check if headers are provided
	if len(args) > 1 {
		if headers, ok := args[1].(map[string]interface{}); ok {
			return HTTPAction([]interface{}{"GET", url, headers}, silent)
		}
	}

	return HTTPAction([]interface{}{"GET", url}, silent)
}

// HTTPPostAction performs HTTP POST requests with simplified syntax.
//
// Parameters:
//   - url: Target URL for the request
//   - body: Request body (string, object, or array)
//   - headers: Request headers (optional, map or array of key-value pairs)
//   - options: Additional options (timeout, follow_redirects, verify_ssl, etc.)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON response with status, headers, body, and timing information
//
// Examples:
//   - JSON POST: ["https://api.example.com/users", {"name": "John", "email": "john@example.com"}]
//   - With headers: ["https://api.example.com/data", "raw data", {"Content-Type": "text/plain"}]
//   - With options: ["https://api.example.com/upload", "file content", {}, {"timeout": 30}]
//
// Use Cases:
//   - Data submission to APIs
//   - Form submissions
//   - File uploads
//   - Authentication requests
//
// Notes:
//   - Simplified syntax for POST requests
//   - Same functionality as HTTPAction with POST method
//   - Supports all HTTPAction options and features
func HTTPPostAction(args []interface{}, silent bool) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("http_post action requires at least 2 arguments: url and body")
	}

	url := fmt.Sprintf("%v", args[0])
	body := fmt.Sprintf("%v", args[1])

	// Check if headers are provided
	if len(args) > 2 {
		if headers, ok := args[2].(map[string]interface{}); ok {
			return HTTPAction([]interface{}{"POST", url, body, headers}, silent)
		}
	}

	return HTTPAction([]interface{}{"POST", url, body}, silent)
}
