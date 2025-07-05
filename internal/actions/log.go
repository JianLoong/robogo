package actions

import "fmt"

func LogAction(args []interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("log action requires at least one argument")
	}

	// Get the message to log
	message := fmt.Sprintf("%v", args[0])

	// Return the message - printing and masking will be handled by the runner
	return message, nil
}
