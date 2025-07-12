package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// LogAction outputs messages to the console with optional formatting and verbosity control.
//
// Parameters:
//   - message: Message to log (can be string, number, boolean, or object)
//   - ...args: Additional arguments to log (optional)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Options:
//   - format: Output format ("auto", "raw", "pretty", "compact")
//     - "auto" (default): Intelligent parsing and formatting
//     - "raw": Simple string representation without parsing
//     - "pretty": Force pretty JSON formatting with indentation
//     - "compact": Single-line JSON formatting
//
// Returns: The logged message as string
//
// Supported Types:
//   - Strings: Logged as-is
//   - Numbers: Converted to string representation
//   - Booleans: Logged as "true"/"false"
//   - Objects: Formatted according to format option
//   - Arrays: Formatted according to format option
//
// Examples:
//   - Simple message: ["Hello World"] -> "Hello World"
//   - With variables: ["User logged in: ", "${user_name}"] -> "User logged in: John"
//   - Auto format: ["Data:", "${response}"] -> "Data: {pretty JSON}"
//   - Raw format: ["Data:", "${response}"] + format:"raw" -> "Data: map[...]"
//   - Pretty format: ["Data:", "${response}"] + format:"pretty" -> "Data: {\n  \"formatted\": \"json\"\n}"
//   - Compact format: ["Data:", "${response}"] + format:"compact" -> "Data: {\"formatted\":\"json\"}"
//
// Use Cases:
//   - Debug information (format: "pretty")
//   - Production logs (format: "compact")
//   - Simple messages (format: "raw")
//   - Smart defaults (format: "auto")
//
// Notes:
//   - Messages are displayed in console output (unless silent=true)
//   - Supports variable substitution with ${variable} syntax
//   - Format option controls object/array rendering
//   - Messages are included in test reports
func LogAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("log action requires at least one argument")
	}

	// Get format option, default to "auto"
	format := "auto"
	if f, ok := options["format"].(string); ok {
		format = strings.ToLower(f)
	}

	// Convert all arguments to string representation with chosen formatting
	var message string
	for i, arg := range args {
		if i > 0 {
			message += " "
		}
		
		// Format the argument based on format option
		formatted := formatArgumentWithMode(arg, format)
		message += formatted
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("%s\n", message)
	}

	return message, nil
}

// formatArgumentWithMode formats an argument based on the specified format mode
func formatArgumentWithMode(arg interface{}, format string) string {
	switch format {
	case "raw":
		// Raw format: no parsing, just direct string representation
		return fmt.Sprintf("%v", arg)
	case "pretty":
		return formatArgumentPretty(arg)
	case "compact":
		return formatArgumentCompact(arg)
	case "auto":
		return formatArgumentAuto(arg)
	default:
		// Unknown format, fall back to auto
		return formatArgumentAuto(arg)
	}
}

// formatArgumentAuto intelligently formats any argument (original behavior)
func formatArgumentAuto(arg interface{}) string {
	// First check if it's already a structured type
	switch v := arg.(type) {
	case map[string]interface{}:
		if jsonBytes, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	case []interface{}:
		if jsonBytes, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	case string:
		// Try to parse string as potential structured data
		return parseAndFormatString(v, "pretty")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatArgumentPretty forces pretty JSON formatting
func formatArgumentPretty(arg interface{}) string {
	// Try to convert to structured data first
	switch v := arg.(type) {
	case map[string]interface{}:
		if jsonBytes, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	case []interface{}:
		if jsonBytes, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	case string:
		// Try to parse and format as pretty JSON
		return parseAndFormatString(v, "pretty")
	default:
		// For non-structured types, try to marshal as JSON if possible
		if jsonBytes, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// formatArgumentCompact forces compact JSON formatting
func formatArgumentCompact(arg interface{}) string {
	// Try to convert to structured data first
	switch v := arg.(type) {
	case map[string]interface{}:
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	case []interface{}:
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	case string:
		// Try to parse and format as compact JSON
		return parseAndFormatString(v, "compact")
	default:
		// For non-structured types, try to marshal as compact JSON if possible
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// parseAndFormatString attempts to parse a string and format it according to the specified format
func parseAndFormatString(s string, format string) string {
	// Remove any leading/trailing whitespace
	s = strings.TrimSpace(s)
	
	// Try to parse as JSON first
	if (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
	   (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(s), &jsonData); err == nil {
			return formatDataAsJSON(jsonData, format)
		}
	}
	
	// Try to parse Go map representation: map[key:value key2:value2]
	if strings.HasPrefix(s, "map[") && strings.HasSuffix(s, "]") {
		if parsed := parseGoMapString(s); parsed != nil {
			return formatDataAsJSON(parsed, format)
		}
	}
	
	// Try to parse Go slice representation: [item1 item2 item3]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") && !strings.Contains(s, "{") {
		if parsed := parseGoSliceString(s); parsed != nil {
			return formatDataAsJSON(parsed, format)
		}
	}
	
	// Return as-is if no parsing worked
	return s
}

// formatDataAsJSON formats data as JSON according to the specified format
func formatDataAsJSON(data interface{}, format string) string {
	switch format {
	case "pretty":
		if jsonBytes, err := json.MarshalIndent(data, "", "  "); err == nil {
			return string(jsonBytes)
		}
	case "compact":
		if jsonBytes, err := json.Marshal(data); err == nil {
			return string(jsonBytes)
		}
	default: // "auto" or fallback
		if jsonBytes, err := json.MarshalIndent(data, "", "  "); err == nil {
			return string(jsonBytes)
		}
	}
	
	// Fallback to string representation
	return fmt.Sprintf("%v", data)
}

// parseGoMapString parses Go's map[key:value] string representation into a map
func parseGoMapString(s string) map[string]interface{} {
	// Remove "map[" and "]"
	content := s[4 : len(s)-1]
	if content == "" {
		return make(map[string]interface{})
	}
	
	result := make(map[string]interface{})
	
	// Handle nested maps by tracking bracket depth
	depth := 0
	start := 0
	inKey := true
	currentKey := ""
	
	for i, char := range content {
		switch char {
		case '[':
			depth++
		case ']':
			depth--
		case ':':
			if depth == 0 && inKey {
				currentKey = strings.TrimSpace(content[start:i])
				start = i + 1
				inKey = false
			}
		case ' ':
			if depth == 0 && !inKey && (i == len(content)-1 || content[i+1] != ' ') {
				// End of value
				value := strings.TrimSpace(content[start:i])
				if value != "" {
					result[currentKey] = parseValue(value)
				}
				start = i + 1
				inKey = true
			}
		}
	}
	
	// Handle last key-value pair
	if !inKey && start < len(content) {
		value := strings.TrimSpace(content[start:])
		if value != "" {
			result[currentKey] = parseValue(value)
		}
	}
	
	return result
}

// parseGoSliceString parses Go's [item1 item2] string representation into a slice
func parseGoSliceString(s string) []interface{} {
	// Remove "[" and "]"
	content := s[1 : len(s)-1]
	if content == "" {
		return []interface{}{}
	}
	
	// Split by spaces, but handle nested structures
	var result []interface{}
	depth := 0
	start := 0
	
	for i, char := range content {
		switch char {
		case '[', '{':
			depth++
		case ']', '}':
			depth--
		case ' ':
			if depth == 0 {
				item := strings.TrimSpace(content[start:i])
				if item != "" {
					result = append(result, parseValue(item))
				}
				start = i + 1
			}
		}
	}
	
	// Handle last item
	if start < len(content) {
		item := strings.TrimSpace(content[start:])
		if item != "" {
			result = append(result, parseValue(item))
		}
	}
	
	return result
}

// parseValue attempts to parse a string value into its appropriate type
func parseValue(s string) interface{} {
	s = strings.TrimSpace(s)
	
	// Try boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	
	// Try number
	if num, err := strconv.Atoi(s); err == nil {
		return num
	}
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num
	}
	
	// Try nested map
	if strings.HasPrefix(s, "map[") && strings.HasSuffix(s, "]") {
		if parsed := parseGoMapString(s); parsed != nil {
			return parsed
		}
	}
	
	// Try nested slice
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		if parsed := parseGoSliceString(s); parsed != nil {
			return parsed
		}
	}
	
	// Return as string
	return s
}
