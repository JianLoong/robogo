package actions

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/util"
)

// HTTPResponse represents the response from an HTTP request with debugging information
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    HTTPHeaders       `json:"headers"`
	Body       string            `json:"body"`
	Duration   time.Duration     `json:"duration"`
	
	// Enhanced debugging fields
	Request    HTTPRequestInfo   `json:"request"`
	Redirects  []HTTPRedirect    `json:"redirects,omitempty"`
	TLSInfo    *TLSInfo         `json:"tls_info,omitempty"`
	Timing     HTTPTiming       `json:"timing"`
	Error      *HTTPError       `json:"error,omitempty"`
}

// HTTPHeaders provides easy access to common headers plus raw headers
type HTTPHeaders struct {
	// Common headers as direct fields for easy access
	ContentType     string `json:"content_type"`
	ContentLength   string `json:"content_length"`
	Authorization   string `json:"authorization"`
	UserAgent       string `json:"user_agent"`
	Accept          string `json:"accept"`
	AcceptEncoding  string `json:"accept_encoding"`
	CacheControl    string `json:"cache_control"`
	Connection      string `json:"connection"`
	Server          string `json:"server"`
	Date            string `json:"date"`
	Location        string `json:"location"`
	SetCookie       string `json:"set_cookie"`
	
	// Raw headers as map for full access
	Raw map[string]string `json:"raw"`
}

// HTTPRequestInfo contains information about the HTTP request
type HTTPRequestInfo struct {
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	Headers   HTTPHeaders `json:"headers"`
	BodySize  int64       `json:"body_size"`
}

// HTTPRedirect contains information about HTTP redirects
type HTTPRedirect struct {
	From       string `json:"from"`
	To         string `json:"to"`
	StatusCode int    `json:"status_code"`
}

// TLSInfo contains TLS connection information
type TLSInfo struct {
	Version            string   `json:"version"`
	CipherSuite        string   `json:"cipher_suite"`
	ServerCertificates []string `json:"server_certificates,omitempty"`
	ClientCertUsed     bool     `json:"client_cert_used"`
}

// HTTPTiming contains detailed timing information
type HTTPTiming struct {
	DNSLookup    time.Duration `json:"dns_lookup"`
	TCPConnect   time.Duration `json:"tcp_connect"`
	TLSHandshake time.Duration `json:"tls_handshake"`
	FirstByte    time.Duration `json:"first_byte"`
	Total        time.Duration `json:"total"`
}

// HTTPError contains detailed error information
type HTTPError struct {
	Message string                 `json:"message"`
	Type    string                 `json:"type"`
	Details map[string]interface{} `json:"details,omitempty"`
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
		return os.ReadFile(input)
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
		return tls.Certificate{}, util.NewSecurityError("failed to load certificate", err, "http").
			WithDetails(map[string]interface{}{
				"cert_input": certInput,
			})
	}

	keyData, err := loadCertificateData(keyInput)
	if err != nil {
		return tls.Certificate{}, util.NewSecurityError("failed to load private key", err, "http").
			WithDetails(map[string]interface{}{
				"key_input": keyInput,
			})
	}

	return tls.X509KeyPair(certData, keyData)
}

// getTLSVersion returns a human-readable TLS version string
func getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("TLS 0x%04x", version)
	}
}

// getCipherSuite returns a human-readable cipher suite string
func getCipherSuite(cipherSuite uint16) string {
	switch cipherSuite {
	case tls.TLS_RSA_WITH_RC4_128_SHA:
		return "TLS_RSA_WITH_RC4_128_SHA"
	case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA:
		return "TLS_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		return "TLS_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA256:
		return "TLS_RSA_WITH_AES_128_CBC_SHA256"
	case tls.TLS_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:
		return "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:
		return "TLS_ECDHE_RSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		return "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:
		return "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		return "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"
	default:
		return fmt.Sprintf("CIPHER_SUITE_0x%04x", cipherSuite)
	}
}

// newHTTPHeaders creates an HTTPHeaders struct from a map[string]string
func newHTTPHeaders(headerMap map[string]string) HTTPHeaders {
	headers := HTTPHeaders{
		Raw: headerMap,
	}
	
	// Populate common headers (case-insensitive lookup)
	for key, value := range headerMap {
		switch strings.ToLower(key) {
		case "content-type":
			headers.ContentType = value
		case "content-length":
			headers.ContentLength = value
		case "authorization":
			headers.Authorization = value
		case "user-agent":
			headers.UserAgent = value
		case "accept":
			headers.Accept = value
		case "accept-encoding":
			headers.AcceptEncoding = value
		case "cache-control":
			headers.CacheControl = value
		case "connection":
			headers.Connection = value
		case "server":
			headers.Server = value
		case "date":
			headers.Date = value
		case "location":
			headers.Location = value
		case "set-cookie":
			headers.SetCookie = value
		}
	}
	
	return headers
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
//
// HTTPActionWithContext performs HTTP requests with context support for cancellation and timeouts
func HTTPAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return httpActionInternal(ctx, args, options, silent)
}

func httpActionInternal(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("http action requires at least 2 arguments: method and url",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 2,
			}).WithAction("http")
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
		return nil, util.NewValidationError("invalid HTTP method",
			map[string]interface{}{
				"method":        method,
				"valid_methods": []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
			}).WithAction("http")
	}

	// Parse optional arguments
	var headers map[string]string
	var body string
	var timeout time.Duration = 30 * time.Second
	var certFile, keyFile, caFile string

	// Parse timeout from options
	if timeoutVal, ok := options["timeout"]; ok {
		if timeoutStr, ok := timeoutVal.(string); ok {
			timeout = util.ParseTimeout(timeoutStr, 30*time.Second)
		} else if timeoutDur, ok := timeoutVal.(time.Duration); ok {
			timeout = timeoutDur
		}
	}

	// Apply action timeout to context
	ctx, cancel := util.WithActionTimeout(ctx, timeout, "http")
	defer cancel()

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
			return nil, util.NewSecurityError("failed to load client certificate/key", err, "http").
				WithDetails(map[string]interface{}{
					"cert_file": certFile,
					"key_file":  keyFile,
				})
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	if caFile != "" {
		caData, err := loadCertificateData(caFile)
		if err != nil {
			return nil, util.NewSecurityError("failed to load CA certificate", err, "http").
				WithDetails(map[string]interface{}{
					"ca_file": caFile,
				})
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caData) {
			return nil, util.NewSecurityError("failed to append CA certificate", nil, "http").
				WithDetails(map[string]interface{}{
					"ca_file": caFile,
				})
		}
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		tlsConfig.RootCAs = caPool
	}

	// HTTP client (no timeout here, managed by context)
	transport := &http.Transport{}
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}
	client := &http.Client{
		Transport: transport,
	}

	// Create request with context
	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, util.NewNetworkError("failed to create request", err, "http").
			WithDetails(map[string]interface{}{
				"method": method,
				"url":    url,
			})
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

	// Execute request with detailed timing
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)
	
	if err != nil {
		// Create error response map instead of returning Go error
		// This maintains consistency with successful responses
		errorResponse := HTTPResponse{
			StatusCode: 0, // 0 indicates network error
			Headers:    newHTTPHeaders(make(map[string]string)),
			Body:       "",
			Duration:   duration,
			Request: HTTPRequestInfo{
				Method:   method,
				URL:      url,
				Headers:  newHTTPHeaders(make(map[string]string)),
				BodySize: int64(len(body)),
			},
			Timing: HTTPTiming{
				Total: duration,
			},
			Error: &HTTPError{
				Message: err.Error(),
				Type:    "network_error",
				Details: map[string]interface{}{
					"method":   method,
					"url":      url,
					"timeout":  timeout.String(),
					"duration": duration.String(),
				},
			},
		}
		
		// Convert to map for template engine compatibility
		errorMap, convertErr := util.ConvertToMap(errorResponse)
		if convertErr != nil {
			return nil, util.NewExecutionError("failed to convert error response to map", convertErr, "http")
		}
		
		// Enhanced output for debugging
		if !silent {
			fmt.Printf("❌ %s %s → ERROR (%v): %s\n", method, url, duration, err.Error())
		}
		
		return errorMap, nil
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, util.NewNetworkError("failed to read response body", err, "http").
			WithDetails(map[string]interface{}{
				"method":      method,
				"url":         url,
				"status_code": resp.StatusCode,
			})
	}

	// Build response headers map
	respHeadersMap := make(map[string]string)
	for k, v := range resp.Header {
		respHeadersMap[k] = strings.Join(v, ", ")
	}

	// Build request info for debugging
	requestHeadersMap := make(map[string]string)
	for k, v := range req.Header {
		requestHeadersMap[k] = strings.Join(v, ", ")
	}
	
	requestInfo := HTTPRequestInfo{
		Method:   method,
		URL:      url,
		Headers:  newHTTPHeaders(requestHeadersMap),
		BodySize: int64(len(body)),
	}
	
	// Build timing info
	timingInfo := HTTPTiming{
		Total: duration,
		// TODO: Add detailed timing metrics using httptrace
	}
	
	// Build TLS info if available
	var tlsInfo *TLSInfo
	if resp.TLS != nil {
		tlsInfo = &TLSInfo{
			Version:        getTLSVersion(resp.TLS.Version),
			CipherSuite:    getCipherSuite(resp.TLS.CipherSuite),
			ClientCertUsed: len(resp.TLS.PeerCertificates) > 0,
		}
		
		// Add server certificate info
		for _, cert := range resp.TLS.PeerCertificates {
			tlsInfo.ServerCertificates = append(tlsInfo.ServerCertificates, cert.Subject.String())
		}
	}
	
	// Check if this is an HTTP error status and add error information
	var httpError *HTTPError
	if resp.StatusCode >= 400 {
		errorType := "client_error"
		if resp.StatusCode >= 500 {
			errorType = "server_error"
		}
		
		httpError = &HTTPError{
			Message: fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
			Type:    errorType,
			Details: map[string]interface{}{
				"status_code":   resp.StatusCode,
				"status_text":   http.StatusText(resp.StatusCode),
				"method":        method,
				"url":           url,
				"response_body": string(respBody),
			},
		}
	}
	
	// Create enhanced response object
	httpResp := HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    newHTTPHeaders(respHeadersMap),
		Body:       string(respBody),
		Duration:   duration,
		Request:    requestInfo,
		Timing:     timingInfo,
		TLSInfo:    tlsInfo,
		Error:      httpError,
	}

	// Convert struct to map for template engine compatibility
	// The variable manager only supports map access, not Go struct field access
	responseMap, err := util.ConvertToMap(httpResp)
	if err != nil {
		return nil, util.NewExecutionError("failed to convert response to map", err, "http")
	}

	// Enhanced output for debugging
	if !silent {
		statusIcon := "→"
		if resp.StatusCode >= 400 {
			statusIcon = "❌"
		} else if resp.StatusCode >= 300 {
			statusIcon = "↗"
		} else {
			statusIcon = "✅"
		}
		
		fmt.Printf("%s %s %s %d (%v)\n", method, url, statusIcon, resp.StatusCode, duration)
		if len(respBody) > 0 {
			if len(respBody) > 500 {
				fmt.Printf("Response body (%d bytes): %s...\n", len(respBody), string(respBody[:500]))
			} else {
				fmt.Printf("Response body: %s\n", string(respBody))
			}
		}
		if tlsInfo != nil {
			fmt.Printf("TLS: %s with %s\n", tlsInfo.Version, tlsInfo.CipherSuite)
		}
		if httpError != nil {
			fmt.Printf("Error: %s\n", httpError.Message)
		}
	}

	return responseMap, nil
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
func HTTPGetAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewValidationError("http_get action requires at least 1 argument: url",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 1,
			}).WithAction("http_get")
	}

	url := fmt.Sprintf("%v", args[0])

	// Check if headers are provided
	if len(args) > 1 {
		if headers, ok := args[1].(map[string]interface{}); ok {
			return HTTPAction(ctx, []interface{}{"GET", url, headers}, options, silent)
		}
	}

	return HTTPAction(ctx, []interface{}{"GET", url}, options, silent)
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
func HTTPPostAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("http_post action requires at least 2 arguments: url and body",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 2,
			}).WithAction("http_post")
	}

	url := fmt.Sprintf("%v", args[0])
	body := fmt.Sprintf("%v", args[1])

	// Check if headers are provided
	if len(args) > 2 {
		if headers, ok := args[2].(map[string]interface{}); ok {
			return HTTPAction(ctx, []interface{}{"POST", url, body, headers}, options, silent)
		}
	}

	return HTTPAction(ctx, []interface{}{"POST", url, body}, options, silent)
}

// HTTPBatchResult represents the result of a batch HTTP operation
type HTTPBatchResult struct {
	URL      string       `json:"url"`
	Response HTTPResponse `json:"response"`
	Error    string       `json:"error,omitempty"`
}

// HTTPBatchAction performs multiple HTTP requests in parallel
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - urls: Array of URLs to request
//   - headers: Request headers (optional)
//   - body: Request body (optional, for POST/PUT requests)
//   - options: Additional options including concurrency limit
//   - silent: Whether to suppress output
//
// Returns: JSON array of results with status, headers, body, and timing for each request
//
// Examples:
//   - Parallel GET requests: ["GET", ["https://api1.com", "https://api2.com"], {"Authorization": "Bearer ${token}"}]
//   - Parallel POST requests: ["POST", ["https://api1.com/users", "https://api2.com/users"], {}, {"name": "John"}]
//   - With concurrency limit: ["GET", ["url1", "url2", "url3"], {}, {}, {"concurrency": 5}]
//
// Use Cases:
//   - Load testing multiple endpoints
//   - Batch API operations
//   - Performance testing
//   - Health checks across multiple services
func HTTPBatchAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("http_batch action requires at least 2 arguments: method and urls",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 2,
			}).WithAction("http_batch")
	}

	method := strings.ToUpper(fmt.Sprintf("%v", args[0]))

	// Parse URLs
	var urls []string
	switch v := args[1].(type) {
	case []interface{}:
		for _, url := range v {
			urls = append(urls, fmt.Sprintf("%v", url))
		}
	case []string:
		urls = v
	default:
		return nil, util.NewValidationError("urls must be an array",
			map[string]interface{}{
				"urls_type": fmt.Sprintf("%T", args[1]),
			}).WithAction("http_batch")
	}

	if len(urls) == 0 {
		return nil, util.NewValidationError("at least one URL is required",
			map[string]interface{}{
				"urls_count": len(urls),
			}).WithAction("http_batch")
	}

	// Parse optional arguments
	var headers map[string]string
	var body string
	var timeout time.Duration = 30 * time.Second
	var maxConcurrency int = 10 // Default concurrency limit

	for i := 2; i < len(args); i++ {
		switch v := args[i].(type) {
		case map[string]interface{}:
			// Check for concurrency limit
			if c, ok := v["concurrency"]; ok {
				if concurrency, ok := c.(int); ok {
					maxConcurrency = concurrency
				}
			}
			// Check for timeout
			if t, ok := v["timeout"]; ok {
				if timeoutStr, ok := t.(string); ok {
					if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
						timeout = parsedTimeout
					}
				}
			}
			// Treat other fields as headers
			if headers == nil {
				headers = make(map[string]string)
			}
			for k, val := range v {
				if k != "concurrency" && k != "timeout" {
					headers[k] = fmt.Sprintf("%v", val)
				}
			}
		case string:
			// This could be body
			if body == "" && (method == "POST" || method == "PUT" || method == "PATCH") {
				body = v
			}
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: timeout,
	}

	// Execute requests in parallel
	results := make([]HTTPBatchResult, len(urls))
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(index int, targetURL string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create request
			var req *http.Request
			var err error

			if body != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
				req, err = http.NewRequestWithContext(ctx, method, targetURL, strings.NewReader(body))
			} else {
				req, err = http.NewRequestWithContext(ctx, method, targetURL, nil)
			}

			if err != nil {
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: fmt.Sprintf("failed to create request: %v", err),
				}
				return
			}

			// Add headers
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			// Execute request
			startTime := time.Now()
			resp, err := client.Do(req)
			duration := time.Since(startTime)

			if err != nil {
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: fmt.Sprintf("request failed: %v", err),
				}
				return
			}
			defer resp.Body.Close()

			// Read response body
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: fmt.Sprintf("failed to read response body: %v", err),
				}
				return
			}

			// Build response headers map
			respHeadersMap := make(map[string]string)
			for k, v := range resp.Header {
				respHeadersMap[k] = strings.Join(v, ", ")
			}

			// Create response object
			response := HTTPResponse{
				StatusCode: resp.StatusCode,
				Headers:    newHTTPHeaders(respHeadersMap),
				Body:       string(respBody),
				Duration:   duration,
			}

			results[index] = HTTPBatchResult{
				URL:      targetURL,
				Response: response,
			}
		}(i, url)
	}

	wg.Wait()

	// Convert results to map format for template engine compatibility
	resultsMap, err := util.ConvertToMap(results)
	if err != nil {
		return nil, util.NewExecutionError("failed to convert batch results to map", err, "http_batch")
	}

	if !silent {
		fmt.Printf("Batch HTTP requests completed: %d URLs, %d concurrent\n", len(urls), maxConcurrency)
	}

	return resultsMap, nil
}

// HTTPGetActionWithContext performs HTTP GET requests with context support.
func HTTPGetActionWithContext(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewValidationError("http_get action requires at least 1 argument: url",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 1,
			}).WithAction("http_get")
	}

	url := fmt.Sprintf("%v", args[0])

	// Check if headers are provided
	if len(args) > 1 {
		if headers, ok := args[1].(map[string]interface{}); ok {
			return HTTPAction(ctx, []interface{}{"GET", url, headers}, options, silent)
		}
	}

	return HTTPAction(ctx, []interface{}{"GET", url}, options, silent)
}

// HTTPPostActionWithContext performs HTTP POST requests with context support.
func HTTPPostActionWithContext(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("http_post action requires at least 2 arguments: url and body",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 2,
			}).WithAction("http_post")
	}

	url := fmt.Sprintf("%v", args[0])
	body := fmt.Sprintf("%v", args[1])

	// Check if headers are provided
	if len(args) > 2 {
		if headers, ok := args[2].(map[string]interface{}); ok {
			return HTTPAction(ctx, []interface{}{"POST", url, body, headers}, options, silent)
		}
	}

	return HTTPAction(ctx, []interface{}{"POST", url, body}, options, silent)
}

// HTTPBatchActionWithContext performs multiple HTTP requests in parallel with context support.
func HTTPBatchActionWithContext(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("http_batch action requires at least 2 arguments: method and urls",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 2,
			}).WithAction("http_batch")
	}

	method := strings.ToUpper(fmt.Sprintf("%v", args[0]))

	// Parse URLs
	var urls []string
	switch v := args[1].(type) {
	case []interface{}:
		for _, url := range v {
			urls = append(urls, fmt.Sprintf("%v", url))
		}
	case []string:
		urls = v
	default:
		return nil, util.NewValidationError("urls must be an array",
			map[string]interface{}{
				"urls_type": fmt.Sprintf("%T", args[1]),
			}).WithAction("http_batch")
	}

	if len(urls) == 0 {
		return nil, util.NewValidationError("at least one URL is required",
			map[string]interface{}{
				"urls_count": len(urls),
			}).WithAction("http_batch")
	}

	// Parse optional arguments
	var headers map[string]string
	var body string
	var timeout time.Duration = 30 * time.Second
	var maxConcurrency int = 10 // Default concurrency limit

	for i := 2; i < len(args); i++ {
		switch v := args[i].(type) {
		case map[string]interface{}:
			// Check for concurrency limit
			if c, ok := v["concurrency"]; ok {
				if concurrency, ok := c.(int); ok {
					maxConcurrency = concurrency
				}
			}
			// Check for timeout
			if t, ok := v["timeout"]; ok {
				if timeoutStr, ok := t.(string); ok {
					if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
						timeout = parsedTimeout
					}
				}
			}
			// Treat other fields as headers
			if headers == nil {
				headers = make(map[string]string)
			}
			for k, val := range v {
				if k != "concurrency" && k != "timeout" {
					headers[k] = fmt.Sprintf("%v", val)
				}
			}
		case string:
			// This could be body
			if body == "" && (method == "POST" || method == "PUT" || method == "PATCH") {
				body = v
			}
		}
	}

	// Apply action timeout to context
	ctx, cancel := util.WithActionTimeout(ctx, timeout, "http_batch")
	defer cancel()

	// Create HTTP client (no timeout here, managed by context)
	client := &http.Client{}

	// Execute requests in parallel
	results := make([]HTTPBatchResult, len(urls))
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(index int, targetURL string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check for context cancellation
			select {
			case <-ctx.Done():
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: "operation cancelled by context",
				}
				return
			default:
			}

			// Create request with context
			var req *http.Request
			var err error

			if body != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
				req, err = http.NewRequestWithContext(ctx, method, targetURL, strings.NewReader(body))
			} else {
				req, err = http.NewRequestWithContext(ctx, method, targetURL, nil)
			}

			if err != nil {
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: fmt.Sprintf("failed to create request: %v", err),
				}
				return
			}

			// Add headers
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			// Execute request
			startTime := time.Now()
			resp, err := client.Do(req)
			duration := time.Since(startTime)

			if err != nil {
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: fmt.Sprintf("request failed: %v", err),
				}
				return
			}
			defer resp.Body.Close()

			// Read response body
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				results[index] = HTTPBatchResult{
					URL:   targetURL,
					Error: fmt.Sprintf("failed to read response body: %v", err),
				}
				return
			}

			// Build response headers map
			respHeadersMap := make(map[string]string)
			for k, v := range resp.Header {
				respHeadersMap[k] = strings.Join(v, ", ")
			}

			// Create response object
			response := HTTPResponse{
				StatusCode: resp.StatusCode,
				Headers:    newHTTPHeaders(respHeadersMap),
				Body:       string(respBody),
				Duration:   duration,
			}

			results[index] = HTTPBatchResult{
				URL:      targetURL,
				Response: response,
			}
		}(i, url)
	}

	wg.Wait()

	// Convert results to map format for template engine compatibility
	resultsMap, err := util.ConvertToMap(results)
	if err != nil {
		return nil, util.NewExecutionError("failed to convert batch results to map", err, "http_batch")
	}

	if !silent {
		fmt.Printf("Batch HTTP requests completed: %d URLs, %d concurrent\n", len(urls), maxConcurrency)
	}

	return resultsMap, nil
}
