package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ConversionError represents a type conversion error
type ConversionError struct {
	FromType string
	ToType   string
	Value    interface{}
	Message  string
}

func (e ConversionError) Error() string {
	return fmt.Sprintf("conversion error from %s to %s: %s (value: %v)",
		e.FromType, e.ToType, e.Message, e.Value)
}

// ToString converts any value to a string representation
func ToString(value interface{}) (string, error) {
	if value == nil {
		return "", ConversionError{
			FromType: "nil",
			ToType:   "string",
			Value:    value,
			Message:  "cannot convert nil to string",
		}
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ToInt converts any value to an integer
func ToInt(value interface{}) (int, error) {
	if value == nil {
		return 0, ConversionError{
			FromType: "nil",
			ToType:   "int",
			Value:    value,
			Message:  "cannot convert nil to int",
		}
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i, nil
		}
		// Try parsing as float first
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return int(f), nil
		}
		return 0, ConversionError{
			FromType: "string",
			ToType:   "int",
			Value:    v,
			Message:  "invalid number format",
		}
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, ConversionError{
			FromType: fmt.Sprintf("%T", value),
			ToType:   "int",
			Value:    value,
			Message:  "unsupported type",
		}
	}
}

// ToFloat converts any value to a float64
func ToFloat(value interface{}) (float64, error) {
	if value == nil {
		return 0, ConversionError{
			FromType: "nil",
			ToType:   "float64",
			Value:    value,
			Message:  "cannot convert nil to float64",
		}
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, nil
		}
		return 0, ConversionError{
			FromType: "string",
			ToType:   "float64",
			Value:    v,
			Message:  "invalid number format",
		}
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, ConversionError{
			FromType: fmt.Sprintf("%T", value),
			ToType:   "float64",
			Value:    value,
			Message:  "unsupported type",
		}
	}
}

// ToBool converts any value to a boolean
func ToBool(value interface{}) (bool, error) {
	if value == nil {
		return false, ConversionError{
			FromType: "nil",
			ToType:   "bool",
			Value:    value,
			Message:  "cannot convert nil to bool",
		}
	}

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
		default:
			return false, ConversionError{
				FromType: "string",
				ToType:   "bool",
				Value:    v,
				Message:  "invalid boolean format",
			}
		}
	default:
		return false, ConversionError{
			FromType: fmt.Sprintf("%T", value),
			ToType:   "bool",
			Value:    value,
			Message:  "unsupported type",
		}
	}
}

// ToMap converts any value to a map[string]interface{}
func ToMap(value interface{}) (map[string]interface{}, error) {
	if value == nil {
		return nil, ConversionError{
			FromType: "nil",
			ToType:   "map[string]interface{}",
			Value:    value,
			Message:  "cannot convert nil to map",
		}
	}

	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, ConversionError{
		FromType: fmt.Sprintf("%T", value),
		ToType:   "map[string]interface{}",
		Value:    value,
		Message:  "not a map type",
	}
}

// ToSlice converts any value to a []interface{}
func ToSlice(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, ConversionError{
			FromType: "nil",
			ToType:   "[]interface{}",
			Value:    value,
			Message:  "cannot convert nil to slice",
		}
	}

	if s, ok := value.([]interface{}); ok {
		return s, nil
	}

	return nil, ConversionError{
		FromType: fmt.Sprintf("%T", value),
		ToType:   "[]interface{}",
		Value:    value,
		Message:  "not a slice type",
	}
}

// IsNumeric checks if a value can be converted to a number
func IsNumeric(value interface{}) bool {
	switch value.(type) {
	case int, float64:
		return true
	case string:
		s := strings.TrimSpace(value.(string))
		_, err := strconv.ParseFloat(s, 64)
		return err == nil
	default:
		return false
	}
}

// IsEmpty checks if a value is considered empty
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	case int:
		return v == 0
	case float64:
		return v == 0
	case bool:
		return !v
	default:
		return false
	}
}

// ConvertToMap converts any struct or value to a map[string]interface{} for template engine compatibility
// This is essential for actions that return structured data to work with ${variable.field} syntax
func ConvertToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, ConversionError{
			FromType: "nil",
			ToType:   "map[string]interface{}",
			Value:    v,
			Message:  "cannot convert nil to map",
		}
	}

	// If it's already a map, return as-is
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}

	// For structs and other types, convert via JSON marshal/unmarshal
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, ConversionError{
			FromType: fmt.Sprintf("%T", v),
			ToType:   "map[string]interface{}",
			Value:    v,
			Message:  fmt.Sprintf("JSON marshal failed: %v", err),
		}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, ConversionError{
			FromType: fmt.Sprintf("%T", v),
			ToType:   "map[string]interface{}",
			Value:    v,
			Message:  fmt.Sprintf("JSON unmarshal failed: %v", err),
		}
	}

	return result, nil
}

// ConvertToMapOrKeepSimple converts structured data to maps but keeps simple types as-is
// This is useful for actions that may return either simple values or complex structures
func ConvertToMapOrKeepSimple(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	// Keep simple types as-is
	switch v.(type) {
	case string, int, int64, float64, bool:
		return v, nil
	case map[string]interface{}:
		return v, nil
	}

	// Convert structured types to maps
	return ConvertToMap(v)
}
