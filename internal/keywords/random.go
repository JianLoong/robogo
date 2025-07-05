package keywords

import (
	"fmt"
	"math/rand"
	"strconv"
)

func GetRandomAction(args []interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("get_random action requires at least one argument (max value)")
	}
	maxStr := fmt.Sprintf("%v", args[0])
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return "", fmt.Errorf("invalid max value: %s", maxStr)
	}
	if max <= 0 {
		return "", fmt.Errorf("max value must be positive")
	}
	result := rand.Intn(max)
	fmt.Printf("🎲 Random number (0-%d): %d\n", max-1, result)
	return strconv.Itoa(result), nil
} 