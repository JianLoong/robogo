package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ArgParser provides utilities for parsing action arguments
type ArgParser struct {
	args    []interface{}
	options map[string]interface{}
}

// NewArgParser creates a new argument parser
func NewArgParser(args []interface{}, options map[string]interface{}) *ArgParser {
	return &ArgParser{
		args:    args,
		options: options,
	}
}

// RequireMinArgs validates that args has at least the minimum required arguments
func (ap *ArgParser) RequireMinArgs(min int) error {
	if len(ap.args) < min {
		return NewValidationError(
			fmt.Sprintf("requires at least %d argument(s), got %d", min, len(ap.args)),
			map[string]interface{}{
				"required": min,
				"provided": len(ap.args),
				"args":     ap.args,
			},
		)
	}
	return nil
}

// RequireExactArgs validates that args has exactly the required number of arguments
func (ap *ArgParser) RequireExactArgs(count int) error {
	if len(ap.args) != count {
		return NewValidationError(
			fmt.Sprintf("requires exactly %d argument(s), got %d", count, len(ap.args)),
			map[string]interface{}{
				"required": count,
				"provided": len(ap.args),
				"args":     ap.args,
			},
		)
	}
	return nil
}

// GetString extracts a string argument at the specified index
func (ap *ArgParser) GetString(index int) (string, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return "", err
	}

	value := ap.args[index]
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", NewValidationError(
		fmt.Sprintf("argument %d must be a string", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetInt extracts an integer argument at the specified index
func (ap *ArgParser) GetInt(index int) (int, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return 0, err
	}

	value := ap.args[index]
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i, nil
		}
	}

	return 0, NewValidationError(
		fmt.Sprintf("argument %d must be a number", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetFloat extracts a float argument at the specified index
func (ap *ArgParser) GetFloat(index int) (float64, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return 0, err
	}

	value := ap.args[index]
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, nil
		}
	}

	return 0, NewValidationError(
		fmt.Sprintf("argument %d must be a number", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetBool extracts a boolean argument at the specified index
func (ap *ArgParser) GetBool(index int) (bool, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return false, err
	}

	value := ap.args[index]
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		switch lower {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		}
	}

	return false, NewValidationError(
		fmt.Sprintf("argument %d must be a boolean", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetMap extracts a map argument at the specified index
func (ap *ArgParser) GetMap(index int) (map[string]interface{}, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return nil, err
	}

	value := ap.args[index]
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, NewValidationError(
		fmt.Sprintf("argument %d must be a map", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetSlice extracts a slice argument at the specified index
func (ap *ArgParser) GetSlice(index int) ([]interface{}, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return nil, err
	}

	value := ap.args[index]
	if s, ok := value.([]interface{}); ok {
		return s, nil
	}

	return nil, NewValidationError(
		fmt.Sprintf("argument %d must be a slice", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetDuration extracts a duration argument at the specified index
func (ap *ArgParser) GetDuration(index int) (time.Duration, error) {
	if err := ap.RequireMinArgs(index + 1); err != nil {
		return 0, err
	}

	value := ap.args[index]
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case int:
		return time.Duration(v) * time.Second, nil
	case float64:
		return time.Duration(v * float64(time.Second)), nil
	case string:
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d, nil
		}
		// Try parsing as seconds
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return time.Duration(f * float64(time.Second)), nil
		}
	}

	return 0, NewValidationError(
		fmt.Sprintf("argument %d must be a valid duration", index),
		map[string]interface{}{
			"index": index,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		},
	)
}

// GetOptionalString extracts an optional string argument at the specified index
func (ap *ArgParser) GetOptionalString(index int) (string, bool) {
	if len(ap.args) <= index {
		return "", false
	}

	if str, ok := ap.args[index].(string); ok {
		return str, true
	}
	return "", false
}

// GetOptionalInt extracts an optional integer argument at the specified index
func (ap *ArgParser) GetOptionalInt(index int) (int, bool) {
	if len(ap.args) <= index {
		return 0, false
	}

	value := ap.args[index]
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetOptionalFloat extracts an optional float argument at the specified index
func (ap *ArgParser) GetOptionalFloat(index int) (float64, bool) {
	if len(ap.args) <= index {
		return 0, false
	}

	value := ap.args[index]
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetOption extracts an option value by key
func (ap *ArgParser) GetOption(key string) (interface{}, bool) {
	if ap.options == nil {
		return nil, false
	}
	value, exists := ap.options[key]
	return value, exists
}

// GetOptionString extracts a string option value by key
func (ap *ArgParser) GetOptionString(key string) (string, bool) {
	if ap.options == nil {
		return "", false
	}
	if value, exists := ap.options[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetOptionInt extracts an integer option value by key
func (ap *ArgParser) GetOptionInt(key string) (int, bool) {
	if ap.options == nil {
		return 0, false
	}
	if value, exists := ap.options[key]; exists {
		switch v := value.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case string:
			if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				return i, true
			}
		}
	}
	return 0, false
}

// GetOptionFloat extracts a float option value by key
func (ap *ArgParser) GetOptionFloat(key string) (float64, bool) {
	if ap.options == nil {
		return 0, false
	}
	if value, exists := ap.options[key]; exists {
		switch v := value.(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case string:
			if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

// GetOptionBool extracts a boolean option value by key
func (ap *ArgParser) GetOptionBool(key string) (bool, bool) {
	if ap.options == nil {
		return false, false
	}
	if value, exists := ap.options[key]; exists {
		switch v := value.(type) {
		case bool:
			return v, true
		case int:
			return v != 0, true
		case float64:
			return v != 0, true
		case string:
			lower := strings.ToLower(strings.TrimSpace(v))
			switch lower {
			case "true", "1", "yes", "on":
				return true, true
			case "false", "0", "no", "off":
				return false, true
			}
		}
	}
	return false, false
}

// ValidateOneOf validates that a string value is one of the allowed values
func (ap *ArgParser) ValidateOneOf(index int, allowed []string) error {
	value, err := ap.GetString(index)
	if err != nil {
		return err
	}

	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}

	return NewValidationError(
		fmt.Sprintf("argument %d must be one of: %s", index, strings.Join(allowed, ", ")),
		map[string]interface{}{
			"index":   index,
			"value":   value,
			"allowed": allowed,
		},
	)
}
