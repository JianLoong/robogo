package parser

import (
	"fmt"
	"strings"
)

// VerbosityLevel represents the level of verbosity
type VerbosityLevel string

const (
	VerbosityNone     VerbosityLevel = "none"
	VerbosityBasic    VerbosityLevel = "basic"
	VerbosityDetailed VerbosityLevel = "detailed"
	VerbosityDebug    VerbosityLevel = "debug"
)

// ParseVerbosity parses a verbosity value and returns the appropriate level
func ParseVerbosity(value interface{}) (VerbosityLevel, error) {
	if value == nil {
		return VerbosityNone, nil
	}

	switch v := value.(type) {
	case bool:
		if v {
			return VerbosityBasic, nil
		}
		return VerbosityNone, nil
	case string:
		level := VerbosityLevel(strings.ToLower(v))
		switch level {
		case VerbosityNone, VerbosityBasic, VerbosityDetailed, VerbosityDebug:
			return level, nil
		default:
			return VerbosityNone, fmt.Errorf("invalid verbosity level: %s", v)
		}
	default:
		return VerbosityNone, fmt.Errorf("unsupported verbosity type: %T", value)
	}
}

// IsVerbose checks if a verbosity level is verbose (not none)
func IsVerbose(level VerbosityLevel) bool {
	return level != VerbosityNone
}

// GetVerbosityLevel determines the verbosity level for a step
// It checks step-level verbosity first, then falls back to test case level
func GetVerbosityLevel(step *Step, testCase *TestCase) VerbosityLevel {
	// Check step-level verbosity first
	if step.Verbose != nil {
		if level, err := ParseVerbosity(step.Verbose); err == nil {
			return level
		}
	}

	// Fall back to test case level
	if testCase.Verbose != nil {
		if level, err := ParseVerbosity(testCase.Verbose); err == nil {
			return level
		}
	}

	return VerbosityNone
}

// FormatVerboseOutput formats verbose output based on the level
func FormatVerboseOutput(level VerbosityLevel, action string, args []interface{}, output string, duration string) string {
	if !IsVerbose(level) {
		return ""
	}

	var result strings.Builder

	switch level {
	case VerbosityBasic:
		result.WriteString(fmt.Sprintf("🔍 %s: %s\n", action, duration))
		if output != "" {
			result.WriteString(fmt.Sprintf("📝 Output: %s\n", output))
		}

	case VerbosityDetailed:
		result.WriteString(fmt.Sprintf("🔍 Verbose %s Operation:\n", action))
		result.WriteString(fmt.Sprintf("   Args: %v\n", args))
		result.WriteString(fmt.Sprintf("   Duration: %s\n", duration))
		if output != "" {
			result.WriteString(fmt.Sprintf("   Output: %s\n", output))
		}

	case VerbosityDebug:
		result.WriteString(fmt.Sprintf("🐛 Debug %s Operation:\n", action))
		result.WriteString(fmt.Sprintf("   Args: %v\n", args))
		result.WriteString(fmt.Sprintf("   Duration: %s\n", duration))
		result.WriteString(fmt.Sprintf("   Verbosity Level: %s\n", level))
		if output != "" {
			result.WriteString(fmt.Sprintf("   Output: %s\n", output))
		}
	}

	return result.String()
}
