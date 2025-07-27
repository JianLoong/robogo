package actions

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// tcpConnectAction tests TCP connectivity to a host and port
func tcpConnectAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.MissingArgsError("tcp_connect", 2, len(args))
	}

	// Check for unresolved variables in critical arguments
	if errorResult := validateArgsResolved("tcp_connect", args); errorResult != nil {
		return *errorResult
	}

	host := fmt.Sprintf("%v", args[0])
	portArg := fmt.Sprintf("%v", args[1])

	// Validate host
	if strings.TrimSpace(host) == "" {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "TCP_CONNECT_INVALID_HOST").
			WithTemplate("TCP connect host cannot be empty").
			WithSuggestion("Provide a valid hostname or IP address").
			Build("Empty host provided for TCP connection test")
	}

	// Parse and validate port
	port, err := strconv.Atoi(portArg)
	if err != nil || port < 1 || port > 65535 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "TCP_CONNECT_INVALID_PORT").
			WithTemplate("Invalid port number for TCP connection").
			WithSuggestion("Port must be a number between 1 and 65535").
			Build(fmt.Sprintf("Invalid port '%s' for TCP connection test", portArg))
	}

	// Parse timeout option
	timeout := 5 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		} else {
			return types.NewErrorBuilder(types.ErrorCategoryValidation, "TCP_CONNECT_INVALID_TIMEOUT").
				WithTemplate("Invalid timeout format for TCP connection").
				WithSuggestion("Use format like '5s', '500ms', '1m'").
				Build(fmt.Sprintf("Invalid timeout format '%s' for TCP connection test", timeoutStr))
		}
	}

	// Execute TCP connection test
	result := performTCPConnect(host, port, timeout)
	return result
}

// performTCPConnect executes the actual TCP connection test
func performTCPConnect(host string, port int, timeout time.Duration) types.ActionResult {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	
	fmt.Printf("🔌 Testing TCP connection to %s...\n", address)
	
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	responseTime := time.Since(start)

	if err != nil {
		// Connection failed - this is still a successful test result, just with connected=false
		fmt.Printf("❌ TCP connection failed to %s (%s)\n", address, responseTime)
		
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"connected":     false,
				"error":         err.Error(),
				"response_time": responseTime.String(),
				"host":          host,
				"port":          port,
				"address":       address,
			},
		}
	}

	// Connection successful
	defer conn.Close()
	
	localAddr := conn.LocalAddr().String()
	remoteAddr := conn.RemoteAddr().String()
	
	fmt.Printf("✅ TCP connection successful to %s (%s)\n", address, responseTime)
	
	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"connected":      true,
			"response_time":  responseTime.String(),
			"local_address":  localAddr,
			"remote_address": remoteAddr,
			"host":           host,
			"port":           port,
			"address":        address,
		},
	}
}