package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Variables - simple variable storage and substitution
type Variables struct {
	data map[string]interface{}
}

func NewVariables() *Variables {
	return &Variables{
		data: make(map[string]interface{}),
	}
}

func (v *Variables) Set(key string, value interface{}) {
	// If value is a JSON string, try to parse it
	if str, ok := value.(string); ok {
		var parsed interface{}
		if err := json.Unmarshal([]byte(str), &parsed); err == nil {
			v.data[key] = parsed
			return
		}
	}
	v.data[key] = value
}

func (v *Variables) Get(key string) interface{} {
	if val, exists := v.data[key]; exists {
		return val
	}
	return nil
}

func (v *Variables) Load(vars map[string]interface{}) {
	for k, val := range vars {
		v.data[k] = val
	}
}

// Simple variable substitution using ${variable} syntax
func (v *Variables) Substitute(template string) string {
	if template == "" {
		return template
	}

	result := template
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := result[start+2 : end]
		if varName == "" {
			result = result[:start] + result[end+1:]
			continue
		}

		// Handle dot notation (simple version)
		value := v.resolveDotNotation(varName)
		replacement := fmt.Sprintf("%v", value)
		if replacement == fmt.Sprintf("${%s}", varName) {
			replacement = "__UNRESOLVED__"
		}
		result = result[:start] + replacement + result[end+1:]
	}
	// Warn if unresolved marker remains
	if strings.Contains(result, "__UNRESOLVED__") {
		fmt.Printf("[WARN] Unresolved variable in template: %q\n", result)
	}
	return result
}

// Enhanced dot notation resolver with array indexing support
func (v *Variables) resolveDotNotation(varName string) interface{} {
	if !strings.Contains(varName, ".") && !strings.Contains(varName, "[") {
		// Simple variable
		if val, exists := v.data[varName]; exists {
			return val
		}
		return fmt.Sprintf("${%s}", varName) // Return unresolved
	}

	// Parse the variable path with array indices
	path := v.parsePath(varName)
	if len(path) == 0 {
		return fmt.Sprintf("${%s}", varName)
	}

	baseVar := path[0]
	current, exists := v.data[baseVar]
	if !exists {
		return fmt.Sprintf("${%s}", varName) // Return unresolved
	}

	// Navigate through the path
	for _, segment := range path[1:] {
		current = v.navigateSegment(current, segment)
		if current == nil {
			return fmt.Sprintf("${%s}", varName)
		}
	}

	return current
}

// Parse path with array indices like "response.rows[0][1]"
func (v *Variables) parsePath(varName string) []string {
	var path []string
	current := ""

	i := 0
	for i < len(varName) {
		char := varName[i]
		switch char {
		case '.':
			if current != "" {
				path = append(path, current)
				current = ""
			}
			i++
		case '[':
			if current != "" {
				path = append(path, current)
				current = ""
			}
			// Find the closing bracket
			j := i + 1
			for j < len(varName) && varName[j] != ']' {
				j++
			}
			if j < len(varName) {
				index := varName[i+1 : j]
				path = append(path, "["+index+"]")
				i = j + 1
			} else {
				i++
			}
		case ']':
			// Skip, handled in '[' case
			i++
		default:
			current += string(char)
			i++
		}
	}

	if current != "" {
		path = append(path, current)
	}

	return path
}

// Navigate a single segment (property or array index)
func (v *Variables) navigateSegment(current interface{}, segment string) interface{} {
	if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
		// Array index
		indexStr := segment[1 : len(segment)-1]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil
		}

		// Support []interface{}
		if arr, ok := current.([]interface{}); ok {
			if index >= 0 && index < len(arr) {
				return arr[index]
			}
		}
		// Support [][]interface{} (slice of slices)
		if arr, ok := current.([][]interface{}); ok {
			if index >= 0 && index < len(arr) {
				return arr[index]
			}
		}
		return nil
	} else {
		// Special property: length
		if segment == "length" {
			if arr, ok := current.([]interface{}); ok {
				return len(arr)
			}
			// Also support length on string
			if str, ok := current.(string); ok {
				return len(str)
			}
			// Support length on [][]interface{} (slice of slices)
			if arr, ok := current.([][]interface{}); ok {
				return len(arr)
			}
			fmt.Printf("[DEBUG] Length requested on type: %T, value: %v\n", current, current)
			return 0
		}

		// Property access
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[segment]; exists {
				return val
			}
		}
		return nil
	}
}

// Substitute variables in arguments
func (v *Variables) SubstituteArgs(args []interface{}) []interface{} {
	result := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			result[i] = v.Substitute(str)
		} else {
			result[i] = arg
		}
	}
	return result
}
