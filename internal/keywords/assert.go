package keywords

import "fmt"

func AssertAction(args []interface{}) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("assert action requires at least 2 arguments")
	}
	if args[0] != args[1] {
		return "", fmt.Errorf("assertion failed: %v != %v", args[0], args[1])
	}
	msg := "Assertion passed"
	if len(args) > 2 {
		msg = fmt.Sprintf("%v", args[2])
	}
	fmt.Printf("✅ %s\n", msg)
	return msg, nil
} 