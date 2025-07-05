package actions

import (
	"fmt"
	"math/rand"
	"strconv"
)

func GetRandomAction(args []interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("get_random action requires at least one argument (max value) or two arguments (min, max)")
	}

	// Check if we have range arguments (min, max)
	if len(args) >= 2 {
		return generateRandomRange(args[0], args[1])
	}

	// Single argument - backward compatibility (0 to max)
	return generateRandomMax(args[0])
}

// generateRandomRange generates a random number within a specified range
func generateRandomRange(minArg, maxArg interface{}) (string, error) {
	minStr := fmt.Sprintf("%v", minArg)
	maxStr := fmt.Sprintf("%v", maxArg)

	// Try to parse as floats first (supports decimal ranges)
	minFloat, minErr := strconv.ParseFloat(minStr, 64)
	maxFloat, maxErr := strconv.ParseFloat(maxStr, 64)

	if minErr == nil && maxErr == nil {
		// Decimal range
		if minFloat > maxFloat {
			return "", fmt.Errorf("min value (%.2f) must be less than or equal to max value (%.2f)", minFloat, maxFloat)
		}

		// Handle edge case: same min and max values
		if minFloat == maxFloat {
			result := minFloat
			fmt.Printf("🎲 Random decimal (%.2f-%.2f): %.2f (same values)\n", minFloat, maxFloat, result)
			return fmt.Sprintf("%.2f", result), nil
		}

		result := minFloat + rand.Float64()*(maxFloat-minFloat)
		fmt.Printf("🎲 Random decimal (%.2f-%.2f): %.2f\n", minFloat, maxFloat, result)
		return fmt.Sprintf("%.2f", result), nil
	}

	// Try integer range
	minInt, minErr := strconv.Atoi(minStr)
	maxInt, maxErr := strconv.Atoi(maxStr)

	if minErr == nil && maxErr == nil {
		// Integer range
		if minInt > maxInt {
			return "", fmt.Errorf("min value (%d) must be less than or equal to max value (%d)", minInt, maxInt)
		}

		// Handle edge case: same min and max values
		if minInt == maxInt {
			result := minInt
			fmt.Printf("🎲 Random number (%d-%d): %d (same values)\n", minInt, maxInt, result)
			return strconv.Itoa(result), nil
		}

		result := minInt + rand.Intn(maxInt-minInt+1)
		fmt.Printf("🎲 Random number (%d-%d): %d\n", minInt, maxInt, result)
		return strconv.Itoa(result), nil
	}

	// Mixed types or invalid input
	return "", fmt.Errorf("invalid range values: min=%v, max=%v (both must be numbers)", minArg, maxArg)
}

// generateRandomMax generates a random number from 0 to max (backward compatibility)
func generateRandomMax(maxArg interface{}) (string, error) {
	maxStr := fmt.Sprintf("%v", maxArg)

	// Try to parse as float first (supports decimals)
	if maxFloat, err := strconv.ParseFloat(maxStr, 64); err == nil {
		if maxFloat <= 0 {
			return "", fmt.Errorf("max value must be positive")
		}
		result := rand.Float64() * maxFloat
		fmt.Printf("🎲 Random decimal (0-%.2f): %.2f\n", maxFloat, result)
		return fmt.Sprintf("%.2f", result), nil
	}

	// Fall back to integer parsing (backward compatibility)
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
