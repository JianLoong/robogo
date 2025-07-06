package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ApplyOptions dynamically sets fields on a struct pointer from a map of options.
// It matches map keys to struct fields case-insensitively and handles basic type conversions.
func ApplyOptions(targetStructPtr interface{}, options map[string]interface{}) error {
	v := reflect.ValueOf(targetStructPtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}
	structVal := v.Elem()

	for key, val := range options {
		field := findField(structVal, key)
		if !field.IsValid() {
			// For now, we'll ignore unknown keys to allow for flexibility.
			// A stricter implementation could return an error here.
			continue
		}

		if !field.CanSet() {
			// This would happen for unexported fields (lowercase).
			continue
		}

		convertedVal, err := convertToType(val, field.Type())
		if err != nil {
			return fmt.Errorf("error setting option '%s': %w", key, err)
		}

		field.Set(convertedVal)
	}

	return nil
}

// findField finds a struct field that matches a key case-insensitively.
func findField(v reflect.Value, key string) reflect.Value {
	for i := 0; i < v.NumField(); i++ {
		if strings.EqualFold(v.Type().Field(i).Name, key) {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

// convertToType converts an interface{} value to a specific reflect.Type.
func convertToType(val interface{}, targetType reflect.Type) (reflect.Value, error) {
	valOf := reflect.ValueOf(val)

	// If the value is already assignable to the target type, no conversion is needed.
	if valOf.Type().AssignableTo(targetType) {
		return valOf, nil
	}

	// For most other cases, we'll work with the string representation of the value.
	valStr := fmt.Sprintf("%v", val)

	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(valStr), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle time.Duration, which is an int64 underneath.
		if targetType == reflect.TypeOf(time.Duration(0)) {
			// Try parsing as a formal duration string first (e.g., "5s", "100ms").
			d, err := time.ParseDuration(valStr)
			if err == nil {
				return reflect.ValueOf(d), nil
			}
			// If that fails, try parsing as a plain integer (assuming milliseconds).
			i, errInt := strconv.ParseInt(valStr, 10, 64)
			if errInt != nil {
				return reflect.Value{}, fmt.Errorf("cannot parse '%s' as duration string or integer milliseconds", valStr)
			}
			return reflect.ValueOf(time.Duration(i) * time.Millisecond), nil
		}
		// Handle standard integer types.
		i, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse '%s' as integer", valStr)
		}
		return reflect.ValueOf(i).Convert(targetType), nil

	case reflect.Bool:
		b, err := strconv.ParseBool(valStr)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse '%s' as boolean", valStr)
		}
		return reflect.ValueOf(b), nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse '%s' as float", valStr)
		}
		return reflect.ValueOf(f).Convert(targetType), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported conversion from type %T to %s", val, targetType.Name())
}
