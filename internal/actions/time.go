package actions

import (
	"fmt"
	"time"
)

// Predefined time formats for convenience
var timeFormats = map[string]string{
	"iso":       "2006-01-02T15:04:05Z07:00",
	"iso_date":  "2006-01-02",
	"iso_time":  "15:04:05",
	"datetime":  "2006-01-02 15:04:05",
	"date":      "2006-01-02",
	"time":      "15:04:05",
	"timestamp": "20060102150405",
	"unix":      "unix",    // Special case for Unix timestamp
	"unix_ms":   "unix_ms", // Special case for Unix timestamp in milliseconds
}

func GetTimeAction(args []interface{}) (string, error) {
	format := "datetime" // Default format

	if len(args) > 0 {
		format = fmt.Sprintf("%v", args[0])
	}

	// Check if it's a predefined format
	if predefinedFormat, exists := timeFormats[format]; exists {
		if predefinedFormat == "unix" {
			result := fmt.Sprintf("%d", time.Now().Unix())
			fmt.Printf("🕐 Unix timestamp: %s\n", result)
			return result, nil
		}
		if predefinedFormat == "unix_ms" {
			result := fmt.Sprintf("%d", time.Now().UnixMilli())
			fmt.Printf("🕐 Unix timestamp (ms): %s\n", result)
			return result, nil
		}
		result := time.Now().Format(predefinedFormat)
		fmt.Printf("🕐 Current time (%s): %s\n", format, result)
		return result, nil
	}

	// Use the format as a custom Go time format string
	result := time.Now().Format(format)
	fmt.Printf("🕐 Current time (custom format): %s\n", result)
	return result, nil
}
