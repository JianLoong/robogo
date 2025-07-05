package actions

import "fmt"

func LogAction(args []interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("log action requires at least one argument")
	}
	fmt.Println("📝", args[0])
	return fmt.Sprintf("Logged: %v", args[0]), nil
}
