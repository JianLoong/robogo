package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

// GetTimeAction retrieves current time in various formats and timezones.
//
// Parameters:
//   - format: Time format specification (optional)
//   - timezone: Timezone specification (optional)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Formatted time string
//
// Supported Formats:
//   - "unix": Unix timestamp (seconds since epoch)
//   - "unix_ms": Unix timestamp in milliseconds
//   - "rfc3339": RFC3339 format (2006-01-02T15:04:05Z07:00)
//   - "rfc1123": RFC1123 format (Mon, 02 Jan 2006 15:04:05 MST)
//   - "iso8601": ISO8601 format (2006-01-02T15:04:05.000Z)
//   - "iso": ISO 8601 format (2006-01-02T15:04:05Z07:00)
//   - "iso_date": ISO date only (2006-01-02)
//   - "iso_time": ISO time only (15:04:05)
//   - "datetime": Standard datetime (2006-01-02 15:04:05)
//   - "date": Date only (2006-01-02)
//   - "time": Time only (15:04:05)
//   - "timestamp": Compact timestamp (20060102150405)
//   - Custom: Go time format string (e.g., "2006-01-02 15:04:05")
//
// Supported Timezones:
//   - "UTC": Coordinated Universal Time
//   - "Local": System local timezone
//   - IANA timezone names: "America/New_York", "Europe/London", etc.
//
// Examples:
//   - Current time: [] -> "2024-01-15T10:30:45Z"
//   - Unix timestamp: ["unix"] -> "1705311045"
//   - Custom format: ["2006-01-02 15:04:05"] -> "2024-01-15 10:30:45"
//   - With timezone: ["rfc3339", "America/New_York"] -> "2024-01-15T05:30:45-05:00"
//   - Unix milliseconds: ["unix_ms"] -> "1705311045123"
//
// Use Cases:
//   - Timestamp generation for API requests
//   - Log file naming with timestamps
//   - Performance measurement
//   - Time-based test data
//   - Timezone-specific testing
//
// Notes:
//   - Default format is RFC3339 if not specified
//   - Default timezone is UTC if not specified
//   - Custom formats use Go's time formatting syntax
//   - Timezone names must be valid IANA timezone identifiers
func GetTimeAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	now := time.Now()
	var format string
	var timezone string

	// Parse arguments
	if len(args) > 0 {
		format = fmt.Sprintf("%v", args[0])
	}

	// Relative time support: now, now+1h, now-2d, now+30m, now-15s
	if strings.HasPrefix(format, "now") {
		relative := format[3:] // e.g., +1h, -2d, etc.
		if len(relative) > 0 {
			sign := relative[0]
			delta := relative[1:]
			var dur time.Duration
			var err error
			if strings.HasSuffix(delta, "d") {
				// Days: convert to hours
				numStr := delta[:len(delta)-1]
				num, err := strconv.Atoi(numStr)
				if err != nil {
					return nil, fmt.Errorf("invalid day offset: %v", err)
				}
				dur = time.Duration(num) * 24 * time.Hour
			} else {
				// Use time.ParseDuration for h, m, s
				dur, err = time.ParseDuration(delta)
				if err != nil {
					return nil, fmt.Errorf("invalid duration offset: %v", err)
				}
			}
			if sign == '+' {
				now = now.Add(dur)
			} else if sign == '-' {
				now = now.Add(-dur)
			} else {
				return nil, fmt.Errorf("invalid relative time sign: %c", sign)
			}
		}
		// For 'now' and 'now+/-offset', second arg is format, third is timezone
		if len(args) > 1 {
			format = fmt.Sprintf("%v", args[1])
			if len(args) > 2 {
				timezone = fmt.Sprintf("%v", args[2])
			}
		} else {
			format = ""
		}
	} else {
		// Non-relative time: first is format, second is timezone
		if len(args) > 1 {
			timezone = fmt.Sprintf("%v", args[1])
		}
	}

	// Set timezone if specified
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone '%s': %w", timezone, err)
		}
		now = now.In(loc)
	}

	// Check if it's a predefined format first
	if predefinedFormat, exists := timeFormats[format]; exists {
		if predefinedFormat == "unix" {
			result := fmt.Sprintf("%d", now.Unix())
			if !silent {
				fmt.Printf("🕐 Unix timestamp: %s\n", result)
			}
			return result, nil
		} else if predefinedFormat == "unix_ms" {
			result := fmt.Sprintf("%d", now.UnixMilli())
			if !silent {
				fmt.Printf("🕐 Unix timestamp (ms): %s\n", result)
			}
			return result, nil
		} else {
			result := now.Format(predefinedFormat)
			if !silent {
				fmt.Printf("🕐 Current time (%s): %s\n", format, result)
			}
			return result, nil
		}
	}

	// Format time based on specification
	var result string
	switch format {
	case "":
		result = now.Format(time.RFC3339)
	case "unix":
		result = fmt.Sprintf("%d", now.Unix())
		if !silent {
			fmt.Printf("🕐 Unix timestamp: %s\n", result)
		}
		return result, nil
	case "unix_ms":
		result = fmt.Sprintf("%d", now.UnixMilli())
		if !silent {
			fmt.Printf("🕐 Unix timestamp (ms): %s\n", result)
		}
		return result, nil
	case "rfc3339":
		result = now.Format(time.RFC3339)
	case "rfc1123":
		result = now.Format(time.RFC1123)
	case "iso8601":
		result = now.Format("2006-01-02T15:04:05.000Z")
	default:
		// Try as custom format
		result = now.Format(format)
	}

	if !silent {
		fmt.Printf("🕐 Current time (%s): %s\n", format, result)
	}

	return result, nil
}
