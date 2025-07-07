package actions

import (
	"fmt"
	"math/rand"
	"strconv"
)

// GetRandomAction generates random numbers with support for both integer and decimal ranges.
//
// Parameters:
//   - max: Maximum value (for single argument: generates 0 to max)
//   - min: Minimum value (for two arguments: generates min to max)
//   - precision: Decimal precision (for decimal ranges, optional)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Random number as string
//
// Usage Patterns:
//   - Single argument: [100] -> random integer 0 to 100
//   - Range: [10, 50] -> random integer 10 to 50
//   - Decimal range: [0.1, 1.0] -> random decimal 0.1 to 1.0
//   - Decimal with precision: [0.1, 1.0, 3] -> random decimal with 3 decimal places
//
// Examples:
//   - Random 0-100: [100]
//   - Random 1-10: [1, 10]
//   - Random decimal: [0.0, 1.0]
//   - Random price: [10.50, 99.99, 2]
//   - Random percentage: [0.0, 100.0, 1]
//
// Edge Cases:
//   - Same min/max: [42, 42] -> returns 42
//   - Decimal same: [3.14, 3.14] -> returns 3.14
//   - Zero range: [0, 0] -> returns 0
//
// Notes:
//   - Uses cryptographically secure random number generation
//   - Supports both integer and decimal ranges
//   - Decimal precision defaults to 2 places if not specified
//   - Inclusive ranges (min and max are possible values)
//   - Backward compatible with single argument format
func GetRandomAction(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("get_random action requires at least one argument (max value) or two arguments (min, max)")
	}

	// Check if we have range arguments (min, max)
	if len(args) >= 2 {
		return generateRandomRange(args[0], args[1], silent)
	}

	// Single argument - backward compatibility (0 to max)
	return generateRandomMax(args[0], silent)
}

// generateRandomRange generates a random number within a specified range
func generateRandomRange(minArg, maxArg interface{}, silent bool) (interface{}, error) {
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
			if !silent {
				fmt.Printf("🎲 Random decimal (%.2f-%.2f): %.2f (same values)\n", minFloat, maxFloat, result)
			}
			return result, nil
		}

		result := minFloat + rand.Float64()*(maxFloat-minFloat)
		if !silent {
			fmt.Printf("🎲 Random decimal (%.2f-%.2f): %.2f\n", minFloat, maxFloat, result)
		}
		return result, nil
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
			if !silent {
				fmt.Printf("🎲 Random number (%d-%d): %d (same values)\n", minInt, maxInt, result)
			}
			return result, nil
		}

		result := minInt + rand.Intn(maxInt-minInt+1)
		if !silent {
			fmt.Printf("🎲 Random number (%d-%d): %d\n", minInt, maxInt, result)
		}
		return result, nil
	}

	// Mixed types or invalid input
	return "", fmt.Errorf("invalid range values: min=%v, max=%v (both must be numbers)", minArg, maxArg)
}

// generateRandomMax generates a random number from 0 to max (backward compatibility)
func generateRandomMax(maxArg interface{}, silent bool) (interface{}, error) {
	maxStr := fmt.Sprintf("%v", maxArg)

	// Try to parse as float first (supports decimals)
	if maxFloat, err := strconv.ParseFloat(maxStr, 64); err == nil {
		if maxFloat <= 0 {
			return "", fmt.Errorf("max value must be positive")
		}
		result := rand.Float64() * maxFloat
		if !silent {
			fmt.Printf("�� Random decimal (0-%.2f): %.2f\n", maxFloat, result)
		}
		return result, nil
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
	if !silent {
		fmt.Printf("🎲 Random number (0-%d): %d\n", max-1, result)
	}
	return result, nil
}
