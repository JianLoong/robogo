package util

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation error: %s (value: %v)", e.Message, e.Value)
}

// RequireArgs validates that args has at least the minimum required arguments
func RequireArgs(args []interface{}, min int) error {
	if len(args) < min {
		return ValidationError{
			Message: fmt.Sprintf("requires at least %d argument(s), got %d", min, len(args)),
			Value:   args,
		}
	}
	return nil
}

// RequireArgsExact validates that args has exactly the required number of arguments
func RequireArgsExact(args []interface{}, count int) error {
	if len(args) != count {
		return ValidationError{
			Message: fmt.Sprintf("requires exactly %d argument(s), got %d", count, len(args)),
			Value:   args,
		}
	}
	return nil
}

// RequireString validates and extracts a string argument at the specified index
func RequireString(args []interface{}, index int) (string, error) {
	if err := RequireArgs(args, index+1); err != nil {
		return "", err
	}

	value := args[index]
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", ValidationError{
		Field:   fmt.Sprintf("args[%d]", index),
		Message: "must be a string",
		Value:   value,
	}
}

// RequireInt validates and extracts an integer argument at the specified index
func RequireInt(args []interface{}, index int) (int, error) {
	if err := RequireArgs(args, index+1); err != nil {
		return 0, err
	}

	value := args[index]
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
	}

	return 0, ValidationError{
		Field:   fmt.Sprintf("args[%d]", index),
		Message: "must be a number",
		Value:   value,
	}
}

// RequireFloat validates and extracts a float argument at the specified index
func RequireFloat(args []interface{}, index int) (float64, error) {
	if err := RequireArgs(args, index+1); err != nil {
		return 0, err
	}

	value := args[index]
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
	}

	return 0, ValidationError{
		Field:   fmt.Sprintf("args[%d]", index),
		Message: "must be a number",
		Value:   value,
	}
}

// RequireMap validates and extracts a map argument at the specified index
func RequireMap(args []interface{}, index int) (map[string]interface{}, error) {
	if err := RequireArgs(args, index+1); err != nil {
		return nil, err
	}

	value := args[index]
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, ValidationError{
		Field:   fmt.Sprintf("args[%d]", index),
		Message: "must be a map",
		Value:   value,
	}
}

// RequireSlice validates and extracts a slice argument at the specified index
func RequireSlice(args []interface{}, index int) ([]interface{}, error) {
	if err := RequireArgs(args, index+1); err != nil {
		return nil, err
	}

	value := args[index]
	if s, ok := value.([]interface{}); ok {
		return s, nil
	}

	return nil, ValidationError{
		Field:   fmt.Sprintf("args[%d]", index),
		Message: "must be a slice",
		Value:   value,
	}
}

// ValidateNotEmpty validates that a string is not empty after trimming whitespace
func ValidateNotEmpty(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{
			Field:   fieldName,
			Message: "cannot be empty",
			Value:   value,
		}
	}
	return nil
}

// ValidateRange validates that a numeric value is within the specified range
func ValidateRange(value float64, min, max float64, fieldName string) error {
	if value < min || value > max {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be between %.2f and %.2f", min, max),
			Value:   value,
		}
	}
	return nil
}

// ValidateOneOf validates that a string value is one of the allowed values
func ValidateOneOf(value string, allowed []string, fieldName string) error {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}
	return ValidationError{
		Field:   fieldName,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
		Value:   value,
	}
}
