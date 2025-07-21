package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadDotEnv loads environment variables from a .env file
// Returns error if file exists but cannot be read, returns nil if file doesn't exist
func LoadDotEnv(filepath string) error {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// File doesn't exist, which is OK - just skip loading
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("[WARN] Invalid .env line %d: %s\n", lineNumber, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// Only set if not already set (existing env vars take precedence)
		if os.Getenv(key) == "" {
			err := os.Setenv(key, value)
			if err != nil {
				fmt.Printf("[WARN] Failed to set environment variable %s: %v\n", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

// LoadDotEnvWithDefault attempts to load .env file from the current directory
// This is a convenience function for the common case
func LoadDotEnvWithDefault() error {
	return LoadDotEnv(".env")
}