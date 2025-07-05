package actions

import (
	"fmt"
	"time"
)

func SleepAction(args []interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("sleep action requires duration argument")
	}
	duration, err := parseDuration(args[0])
	if err != nil {
		return "", err
	}
	fmt.Printf("😴 Sleeping for %v\n", duration)
	time.Sleep(duration)
	return fmt.Sprintf("Slept for %v", duration), nil
}
