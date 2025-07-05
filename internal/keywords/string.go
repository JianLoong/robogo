package keywords

import (
	"fmt"
	"strconv"
	"strings"
)

func ConcatAction(args []interface{}) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("concat action requires at least 2 arguments")
	}
	var parts []string
	for _, arg := range args {
		parts = append(parts, fmt.Sprintf("%v", arg))
	}
	result := strings.Join(parts, "")
	fmt.Printf("🔗 Concatenated: %s\n", result)
	return result, nil
}

func LengthAction(args []interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("length action requires one argument")
	}
	var length int
	switch v := args[0].(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	default:
		length = len(fmt.Sprintf("%v", v))
	}
	result := strconv.Itoa(length)
	fmt.Printf("📏 Length: %s\n", result)
	return result, nil
} 