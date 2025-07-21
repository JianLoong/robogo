package common

import (
	"net/url"
	"regexp"
)

// MaskConnectionString masks passwords and sensitive information in connection strings
// Supports various formats: postgres://, mysql://, amqp://, etc.
func MaskConnectionString(connectionString string) string {
	if connectionString == "" {
		return connectionString
	}

	// Try to parse as URL first (most common case)
	if parsedURL, err := url.Parse(connectionString); err == nil && parsedURL.Scheme != "" {
		return maskURLConnectionString(parsedURL)
	}

	// Fallback to regex-based masking for non-URL formats
	return maskWithRegex(connectionString)
}

// maskURLConnectionString masks passwords in URL-format connection strings
func maskURLConnectionString(parsedURL *url.URL) string {
	if parsedURL.User == nil {
		return parsedURL.String()
	}

	// Create a copy to avoid modifying the original
	maskedURL := *parsedURL
	
	username := parsedURL.User.Username()
	if _, hasPassword := parsedURL.User.Password(); hasPassword {
		// Mask the password
		maskedURL.User = url.UserPassword(username, "***")
	}

	return maskedURL.String()
}

// maskWithRegex masks passwords using regex patterns for various connection string formats
func maskWithRegex(connectionString string) string {
	// Patterns for different connection string formats
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		// PostgreSQL: password=xxx or pwd=xxx
		{
			regex:       regexp.MustCompile(`(?i)(password|pwd)=([^;\s]+)`),
			replacement: "${1}=***",
		},
		// MySQL: password=xxx
		{
			regex:       regexp.MustCompile(`(?i)(password)=([^;&\s]+)`),
			replacement: "${1}=***",
		},
		// Generic key=value patterns for passwords
		{
			regex:       regexp.MustCompile(`(?i)(pass|passwd|password|pwd|secret|token|key)=([^;\s&]+)`),
			replacement: "${1}=***",
		},
		// AMQP format: amqp://user:pass@host
		{
			regex:       regexp.MustCompile(`(amqp://[^:]+:)([^@]+)(@)`),
			replacement: "${1}***${3}",
		},
	}

	masked := connectionString
	for _, pattern := range patterns {
		masked = pattern.regex.ReplaceAllString(masked, pattern.replacement)
	}

	return masked
}

// MaskSensitiveData masks various types of sensitive data in strings
// This is a more general function for other sensitive information
func MaskSensitiveData(data string, sensitiveKeys []string) string {
	if data == "" {
		return data
	}

	masked := data
	for _, key := range sensitiveKeys {
		// Case-insensitive pattern matching
		pattern := regexp.MustCompile(`(?i)(` + regexp.QuoteMeta(key) + `[=:]\s*)([^\s,;}&]+)`)
		masked = pattern.ReplaceAllString(masked, "${1}***")
	}

	return masked
}

// Common sensitive keywords that should be masked
var DefaultSensitiveKeys = []string{
	"password", "pass", "passwd", "pwd",
	"secret", "token", "key", "apikey", "api_key",
	"auth", "authorization", "credential", "cred",
}