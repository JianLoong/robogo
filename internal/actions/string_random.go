package actions

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// Character sets for random string generation
const (
	charsetNumeric      = "0123456789"
	charsetLowercase    = "abcdefghijklmnopqrstuvwxyz"
	charsetUppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetAlphabetic   = charsetLowercase + charsetUppercase
	charsetAlphanumeric = charsetAlphabetic + charsetNumeric
	charsetHex          = "0123456789abcdef"
	charsetSpecial      = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	charsetAll          = charsetAlphanumeric + charsetSpecial
)

// stringRandomAction generates a random string
// Args: [length, charset] - length (int) and charset type (string)
// Supported charsets: numeric, lowercase, uppercase, alphabetic, alphanumeric, hex, special, all, custom
func stringRandomAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("string_random", 1, len(args))
	}

	// Parse length
	lengthArg := fmt.Sprintf("%v", args[0])
	length := 0
	if _, err := fmt.Sscanf(lengthArg, "%d", &length); err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_LENGTH").
			WithTemplate("Invalid length for random string generation").
			WithContext("length", lengthArg).
			WithSuggestion("Use a positive integer for length (e.g., 8, 16, 32)").
			Build(fmt.Sprintf("invalid length: %s", lengthArg))
	}

	if length <= 0 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_LENGTH").
			WithTemplate("Length must be positive for random string generation").
			WithContext("length", length).
			WithSuggestion("Use a positive integer greater than 0").
			Build(fmt.Sprintf("length must be positive: %d", length))
	}

	if length > 10000 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "LENGTH_TOO_LARGE").
			WithTemplate("Length is too large for random string generation").
			WithContext("length", length).
			WithContext("max_length", 10000).
			WithSuggestion("Use a length less than or equal to 10000").
			Build(fmt.Sprintf("length too large: %d", length))
	}

	// Determine charset
	charset := "alphanumeric" // default
	if len(args) > 1 {
		charset = strings.ToLower(fmt.Sprintf("%v", args[1]))
	}

	// Get character set
	chars, err := getCharacterSet(charset, options)
	if err != nil {
		return *err
	}

	// Generate random string
	randomString, genErr := generateRandomString(length, chars)
	if genErr != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "RANDOM_GENERATION_ERROR").
			WithTemplate("Failed to generate random string").
			WithContext("length", length).
			WithContext("charset", charset).
			WithContext("error", genErr.Error()).
			WithSuggestion("Try again or use a different charset").
			Build(fmt.Sprintf("random generation error: %s", genErr.Error()))
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"value":   randomString,
			"length":  length,
			"charset": charset,
		},
	}
}

// getCharacterSet returns the character set based on the type
func getCharacterSet(charset string, options map[string]any) (string, *types.ActionResult) {
	switch charset {
	case "numeric", "numbers":
		return charsetNumeric, nil
	case "lowercase", "lower":
		return charsetLowercase, nil
	case "uppercase", "upper":
		return charsetUppercase, nil
	case "alphabetic", "alpha", "letters":
		return charsetAlphabetic, nil
	case "alphanumeric", "alphanum":
		return charsetAlphanumeric, nil
	case "hex", "hexadecimal":
		return charsetHex, nil
	case "special", "symbols":
		return charsetSpecial, nil
	case "all":
		return charsetAll, nil
	case "custom":
		// Custom charset from options
		if customChars, ok := options["custom_chars"].(string); ok {
			if len(customChars) == 0 {
				errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "EMPTY_CUSTOM_CHARSET").
					WithTemplate("Custom charset cannot be empty").
					WithSuggestion("Provide a non-empty string for custom_chars option").
					Build("custom charset is empty")
				return "", &errorResult
			}
			return customChars, nil
		}
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "MISSING_CUSTOM_CHARSET").
			WithTemplate("Custom charset requires custom_chars option").
			WithSuggestion("Add custom_chars option with desired characters").
			WithSuggestion("Example: custom_chars: 'ABC123'").
			Build("custom charset specified but custom_chars option not provided")
		return "", &errorResult
	default:
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "UNSUPPORTED_CHARSET").
			WithTemplate("Unsupported charset for random string generation").
			WithContext("charset", charset).
			WithContext("supported_charsets", "numeric, lowercase, uppercase, alphabetic, alphanumeric, hex, special, all, custom").
			WithSuggestion("Use one of the supported charset types").
			Build(fmt.Sprintf("unsupported charset: %s", charset))
		return "", &errorResult
	}
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int, charset string) (string, error) {
	if len(charset) == 0 {
		return "", fmt.Errorf("charset cannot be empty")
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}