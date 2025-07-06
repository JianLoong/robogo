package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigManager manages application configuration
type ConfigManager struct {
	config map[string]interface{}
	env    map[string]string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: make(map[string]interface{}),
		env:    make(map[string]string),
	}
}

// LoadFromFile loads configuration from a YAML file
func (cm *ConfigManager) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return NewConfigurationError("failed to read config file", "filepath", filepath)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return NewConfigurationError("failed to parse config file", "filepath", filepath)
	}

	// Merge with existing config
	cm.mergeConfig(config)
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (cm *ConfigManager) LoadFromEnv(prefix string) {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key, value := pair[0], pair[1]
		if strings.HasPrefix(key, prefix) {
			configKey := strings.TrimPrefix(key, prefix+"_")
			cm.config[configKey] = cm.parseEnvValue(value)
		}
	}
}

// Set sets a configuration value
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.config[key] = value
}

// Get retrieves a configuration value
func (cm *ConfigManager) Get(key string) (interface{}, bool) {
	value, exists := cm.config[key]
	return value, exists
}

// GetString retrieves a string configuration value
func (cm *ConfigManager) GetString(key string) (string, bool) {
	value, exists := cm.config[key]
	if !exists {
		return "", false
	}
	if str, ok := value.(string); ok {
		return str, true
	}
	return fmt.Sprintf("%v", value), true
}

// GetInt retrieves an integer configuration value
func (cm *ConfigManager) GetInt(key string) (int, bool) {
	value, exists := cm.config[key]
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetFloat retrieves a float configuration value
func (cm *ConfigManager) GetFloat(key string) (float64, bool) {
	value, exists := cm.config[key]
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetBool retrieves a boolean configuration value
func (cm *ConfigManager) GetBool(key string) (bool, bool) {
	value, exists := cm.config[key]
	if !exists {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	case string:
		lower := strings.ToLower(v)
		switch lower {
		case "true", "1", "yes", "on":
			return true, true
		case "false", "0", "no", "off":
			return false, true
		}
	}
	return false, false
}

// GetDuration retrieves a duration configuration value
func (cm *ConfigManager) GetDuration(key string) (time.Duration, bool) {
	value, exists := cm.config[key]
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case time.Duration:
		return v, true
	case int:
		return time.Duration(v) * time.Second, true
	case float64:
		return time.Duration(v * float64(time.Second)), true
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d, true
		}
		// Try parsing as seconds
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(f * float64(time.Second)), true
		}
	}
	return 0, false
}

// GetWithDefault retrieves a configuration value with a default
func (cm *ConfigManager) GetWithDefault(key string, defaultValue interface{}) interface{} {
	if value, exists := cm.config[key]; exists {
		return value
	}
	return defaultValue
}

// GetStringWithDefault retrieves a string configuration value with a default
func (cm *ConfigManager) GetStringWithDefault(key string, defaultValue string) string {
	if value, exists := cm.GetString(key); exists {
		return value
	}
	return defaultValue
}

// GetIntWithDefault retrieves an integer configuration value with a default
func (cm *ConfigManager) GetIntWithDefault(key string, defaultValue int) int {
	if value, exists := cm.GetInt(key); exists {
		return value
	}
	return defaultValue
}

// GetFloatWithDefault retrieves a float configuration value with a default
func (cm *ConfigManager) GetFloatWithDefault(key string, defaultValue float64) float64 {
	if value, exists := cm.GetFloat(key); exists {
		return value
	}
	return defaultValue
}

// GetBoolWithDefault retrieves a boolean configuration value with a default
func (cm *ConfigManager) GetBoolWithDefault(key string, defaultValue bool) bool {
	if value, exists := cm.GetBool(key); exists {
		return value
	}
	return defaultValue
}

// GetDurationWithDefault retrieves a duration configuration value with a default
func (cm *ConfigManager) GetDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value, exists := cm.GetDuration(key); exists {
		return value
	}
	return defaultValue
}

// ValidateRequired validates that required configuration keys are present
func (cm *ConfigManager) ValidateRequired(keys []string) error {
	var missing []string
	for _, key := range keys {
		if _, exists := cm.config[key]; !exists {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return NewConfigurationError(
			fmt.Sprintf("missing required configuration keys: %s", strings.Join(missing, ", ")),
			"missing_keys",
			missing,
		)
	}
	return nil
}

// ValidateOneOf validates that exactly one of the specified keys is present
func (cm *ConfigManager) ValidateOneOf(keys []string) error {
	var present []string
	for _, key := range keys {
		if _, exists := cm.config[key]; exists {
			present = append(present, key)
		}
	}

	if len(present) == 0 {
		return NewConfigurationError(
			fmt.Sprintf("exactly one of the following keys must be present: %s", strings.Join(keys, ", ")),
			"missing_keys",
			keys,
		)
	}

	if len(present) > 1 {
		return NewConfigurationError(
			fmt.Sprintf("only one of the following keys can be present: %s (found: %s)", strings.Join(keys, ", "), strings.Join(present, ", ")),
			"multiple_keys",
			present,
		)
	}

	return nil
}

// SaveToFile saves the current configuration to a YAML file
func (cm *ConfigManager) SaveToFile(configPath string) error {
	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return NewConfigurationError("failed to marshal config", "filepath", configPath)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewConfigurationError("failed to create config directory", "directory", dir)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return NewConfigurationError("failed to write config file", "filepath", configPath)
	}

	return nil
}

// GetConfig returns the entire configuration map
func (cm *ConfigManager) GetConfig() map[string]interface{} {
	return cm.config
}

// Clear clears all configuration
func (cm *ConfigManager) Clear() {
	cm.config = make(map[string]interface{})
}

// mergeConfig merges a new configuration map with the existing one
func (cm *ConfigManager) mergeConfig(newConfig map[string]interface{}) {
	for key, value := range newConfig {
		cm.config[key] = value
	}
}

// parseEnvValue parses an environment variable value
func (cm *ConfigManager) parseEnvValue(value string) interface{} {
	// Try to parse as different types
	if value == "true" || value == "false" {
		return value == "true"
	}

	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	if d, err := time.ParseDuration(value); err == nil {
		return d
	}

	// Default to string
	return value
}

// Global configuration manager instance
var globalConfig = NewConfigManager()

// Global configuration functions
func LoadGlobalConfig(filepath string) error {
	return globalConfig.LoadFromFile(filepath)
}

func LoadGlobalConfigFromEnv(prefix string) {
	globalConfig.LoadFromEnv(prefix)
}

func SetGlobalConfig(key string, value interface{}) {
	globalConfig.Set(key, value)
}

func GetGlobalConfig(key string) (interface{}, bool) {
	return globalConfig.Get(key)
}

func GetGlobalConfigString(key string) (string, bool) {
	return globalConfig.GetString(key)
}

func GetGlobalConfigInt(key string) (int, bool) {
	return globalConfig.GetInt(key)
}

func GetGlobalConfigFloat(key string) (float64, bool) {
	return globalConfig.GetFloat(key)
}

func GetGlobalConfigBool(key string) (bool, bool) {
	return globalConfig.GetBool(key)
}

func GetGlobalConfigDuration(key string) (time.Duration, bool) {
	return globalConfig.GetDuration(key)
}

func GetGlobalConfigWithDefault(key string, defaultValue interface{}) interface{} {
	return globalConfig.GetWithDefault(key, defaultValue)
}

func GetGlobalConfigStringWithDefault(key string, defaultValue string) string {
	return globalConfig.GetStringWithDefault(key, defaultValue)
}

func GetGlobalConfigIntWithDefault(key string, defaultValue int) int {
	return globalConfig.GetIntWithDefault(key, defaultValue)
}

func GetGlobalConfigFloatWithDefault(key string, defaultValue float64) float64 {
	return globalConfig.GetFloatWithDefault(key, defaultValue)
}

func GetGlobalConfigBoolWithDefault(key string, defaultValue bool) bool {
	return globalConfig.GetBoolWithDefault(key, defaultValue)
}

func GetGlobalConfigDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	return globalConfig.GetDurationWithDefault(key, defaultValue)
}
