package actions

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// pingAction performs ICMP ping to a host
// Args: [host] - hostname or IP address to ping
// Options:
//   - count: number of ping packets (default: 4)
//   - timeout: timeout duration per ping (default: "3s")
func pingAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("ping", 1, len(args))
	}

	// Validate arguments are resolved
	if errorResult := validateArgsResolved("ping", args); errorResult != nil {
		return *errorResult
	}

	host := fmt.Sprintf("%v", args[0])
	if host == "" {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "EMPTY_HOST").
			WithTemplate("Ping host cannot be empty").
			WithSuggestion("Provide a valid hostname or IP address").
			Build("empty host provided")
	}

	// Parse options
	count := 4
	if countVal, exists := options["count"]; exists {
		if countInt, ok := countVal.(int); ok {
			count = countInt
		} else if countStr, ok := countVal.(string); ok {
			if parsedCount, err := strconv.Atoi(countStr); err == nil {
				count = parsedCount
			}
		}
	}

	timeout := "3s"
	if timeoutVal, exists := options["timeout"]; exists {
		timeout = fmt.Sprintf("%v", timeoutVal)
	}

	// Validate count
	if count <= 0 || count > 100 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_COUNT").
			WithTemplate("Ping count must be between 1 and 100").
			WithContext("count", count).
			WithSuggestion("Use a reasonable ping count (1-10 for testing)").
			Build(fmt.Sprintf("invalid ping count: %d", count))
	}

	// Parse timeout
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_TIMEOUT").
			WithTemplate("Invalid timeout format for ping action").
			WithContext("timeout", timeout).
			WithContext("valid_examples", "3s, 1000ms, 5s").
			WithSuggestion("Use Go duration format: ns, us, ms, s, m, h").
			Build(fmt.Sprintf("invalid timeout format: %s", timeout))
	}

	// Resolve hostname to IP if needed
	resolvedIPs, err := net.LookupIP(host)
	var resolvedIP string
	if err == nil && len(resolvedIPs) > 0 {
		resolvedIP = resolvedIPs[0].String()
	} else {
		resolvedIP = host // Use original if resolution fails
	}

	// Execute ping command
	result := executePing(host, resolvedIP, count, timeoutDuration)
	
	return result
}

// executePing runs the actual ping command
func executePing(host, resolvedIP string, count int, timeout time.Duration) types.ActionResult {
	var cmd *exec.Cmd
	var args []string

	// Build command based on OS
	switch runtime.GOOS {
	case "windows":
		args = []string{"-n", strconv.Itoa(count), "-w", strconv.Itoa(int(timeout.Milliseconds())), host}
		cmd = exec.Command("ping", args...)
	case "darwin":
		args = []string{"-c", strconv.Itoa(count), "-W", strconv.Itoa(int(timeout.Milliseconds())), host}
		cmd = exec.Command("ping", args...)
	default: // Linux and others
		args = []string{"-c", strconv.Itoa(count), "-W", strconv.Itoa(int(timeout.Seconds())), host}
		cmd = exec.Command("ping", args...)
	}

	fmt.Printf("🏓 Pinging %s (%s) with %d packets...\n", host, resolvedIP, count)
	
	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	outputStr := string(output)
	
	if err != nil {
		// Check if it's a timeout or host unreachable
		if strings.Contains(outputStr, "timeout") || strings.Contains(outputStr, "Destination Host Unreachable") {
			return types.NewFailureBuilder(types.FailureCategoryResponse, "PING_TIMEOUT").
				WithTemplate("Ping operation timed out or host unreachable").
				WithContext("host", host).
				WithContext("resolved_ip", resolvedIP).
				WithContext("count", count).
				WithContext("timeout", timeout.String()).
				WithContext("output", outputStr).
				WithSuggestion("Check network connectivity and host availability").
				WithSuggestion("Verify firewall settings allow ICMP traffic").
				Build(fmt.Sprintf("ping failed for %s: %s", host, err.Error()))
		}
		
		return types.NewErrorBuilder(types.ErrorCategorySystem, "PING_COMMAND_FAILED").
			WithTemplate("Ping command execution failed").
			WithContext("host", host).
			WithContext("command", fmt.Sprintf("ping %s", strings.Join(args, " "))).
			WithContext("error", err.Error()).
			WithContext("output", outputStr).
			WithSuggestion("Ensure ping command is available on the system").
			Build(fmt.Sprintf("ping command failed: %s", err.Error()))
	}

	// Parse ping statistics
	stats := parsePingOutput(outputStr, runtime.GOOS)
	stats["host"] = host
	stats["resolved_ip"] = resolvedIP
	stats["count"] = count
	stats["timeout"] = timeout.String()
	stats["duration_ms"] = duration.Milliseconds()
	stats["raw_output"] = outputStr

	fmt.Printf("✅ Ping completed: %d packets transmitted, %v received\n", 
		count, stats["packets_received"])

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   stats,
	}
}

// parsePingOutput extracts statistics from ping command output
func parsePingOutput(output, os string) map[string]any {
	stats := make(map[string]any)
	
	lines := strings.Split(output, "\n")
	
	// Initialize defaults
	stats["packets_transmitted"] = 0
	stats["packets_received"] = 0
	stats["packet_loss_percent"] = 100.0
	stats["min_rtt_ms"] = 0.0
	stats["avg_rtt_ms"] = 0.0
	stats["max_rtt_ms"] = 0.0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse packet statistics (works for most systems)
		if strings.Contains(line, "packets transmitted") || strings.Contains(line, "Packets: Sent") {
			// Linux/Mac: "4 packets transmitted, 4 received, 0% packet loss"
			// Windows: "Packets: Sent = 4, Received = 4, Lost = 0 (0% loss)"
			
			if os == "windows" {
				if strings.Contains(line, "Sent =") {
					parts := strings.Split(line, ",")
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if strings.Contains(part, "Sent =") {
							if val := extractNumber(part); val >= 0 {
								stats["packets_transmitted"] = val
							}
						} else if strings.Contains(part, "Received =") {
							if val := extractNumber(part); val >= 0 {
								stats["packets_received"] = val
							}
						} else if strings.Contains(part, "% loss") {
							if val := extractFloat(part); val >= 0 {
								stats["packet_loss_percent"] = val
							}
						}
					}
				}
			} else {
				// Linux/Mac format
				if strings.Contains(line, "transmitted") {
					parts := strings.Fields(line)
					if len(parts) >= 4 {
						if val := parseInt(parts[0]); val >= 0 {
							stats["packets_transmitted"] = val
						}
						if val := parseInt(parts[3]); val >= 0 {
							stats["packets_received"] = val
						}
					}
					// Extract packet loss percentage
					if strings.Contains(line, "%") {
						if val := extractFloat(line); val >= 0 {
							stats["packet_loss_percent"] = val
						}
					}
				}
			}
		}
		
		// Parse RTT statistics
		if strings.Contains(line, "min/avg/max") || strings.Contains(line, "Minimum/Maximum/Average") {
			if os == "windows" {
				// Windows: "Minimum = 1ms, Maximum = 4ms, Average = 2ms"
				parts := strings.Split(line, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.Contains(part, "Minimum =") {
						if val := extractFloat(part); val >= 0 {
							stats["min_rtt_ms"] = val
						}
					} else if strings.Contains(part, "Maximum =") {
						if val := extractFloat(part); val >= 0 {
							stats["max_rtt_ms"] = val
						}
					} else if strings.Contains(part, "Average =") {
						if val := extractFloat(part); val >= 0 {
							stats["avg_rtt_ms"] = val
						}
					}
				}
			} else {
				// Linux/Mac: "rtt min/avg/max/mdev = 1.234/2.345/3.456/0.123 ms"
				if strings.Contains(line, "=") {
					parts := strings.Split(line, "=")
					if len(parts) >= 2 {
						values := strings.Fields(strings.TrimSpace(parts[1]))
						if len(values) >= 1 {
							rttValues := strings.Split(values[0], "/")
							if len(rttValues) >= 3 {
								if val := parseFloat(rttValues[0]); val >= 0 {
									stats["min_rtt_ms"] = val
								}
								if val := parseFloat(rttValues[1]); val >= 0 {
									stats["avg_rtt_ms"] = val
								}
								if val := parseFloat(rttValues[2]); val >= 0 {
									stats["max_rtt_ms"] = val
								}
							}
						}
					}
				}
			}
		}
	}
	
	return stats
}

// Helper functions for parsing
func extractNumber(s string) int {
	for _, part := range strings.Fields(s) {
		if val := parseInt(part); val >= 0 {
			return val
		}
	}
	return -1
}

func extractFloat(s string) float64 {
	for _, part := range strings.Fields(s) {
		if val := parseFloat(part); val >= 0 {
			return val
		}
	}
	return -1
}

func parseInt(s string) int {
	// Remove non-numeric characters except digits
	cleaned := strings.TrimFunc(s, func(r rune) bool {
		return r < '0' || r > '9'
	})
	if val, err := strconv.Atoi(cleaned); err == nil {
		return val
	}
	return -1
}

func parseFloat(s string) float64 {
	// Remove 'ms' and other suffixes, keep digits and decimal points
	cleaned := strings.TrimFunc(s, func(r rune) bool {
		return !((r >= '0' && r <= '9') || r == '.')
	})
	if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
		return val
	}
	return -1
}