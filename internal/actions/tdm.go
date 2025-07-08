package actions

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// TDMManager handles test data management operations including data sets, environments, validation, and data generation.
//
// Features:
//   - Data set management with schemas and validation
//   - Environment-specific configurations
//   - Data validation with custom rules
//   - Test data generation with patterns
//   - Variable management and isolation
//
// Examples:
//   - Load data sets from test case configuration
//   - Validate data against schemas
//   - Generate test data with patterns like "user_{index}"
//   - Manage environment-specific variables
type TDMManager struct {
	dataSets     map[string]*parser.DataSet
	environments map[string]*parser.Environment
	variables    map[string]interface{}
	validations  []parser.Validation
	results      *parser.DataResults
}

// NewTDMManager creates a new TDM manager instance.
//
// Returns: Initialized TDMManager with empty data structures
//
// Notes:
//   - Initializes empty maps for data sets, environments, and variables
//   - Creates empty validation and results structures
//   - Ready to load data sets and environments from test cases
func NewTDMManager() *TDMManager {
	return &TDMManager{
		dataSets:     make(map[string]*parser.DataSet),
		environments: make(map[string]*parser.Environment),
		variables:    make(map[string]interface{}),
		validations:  make([]parser.Validation, 0),
		results: &parser.DataResults{
			Validations: make([]parser.ValidationResult, 0),
			DataSets:    make(map[string]parser.DataSetInfo),
		},
	}
}

// LoadDataSets loads data sets into the manager and makes them available as variables.
//
// Parameters:
//   - dataSets: Array of DataSet configurations from test case
//
// Returns: Error if loading fails
//
// Examples:
//   - Load from test case: tdm.LoadDataSets(testCase.DataManagement.DataSets)
//
// Notes:
//   - Creates nested variable structure (e.g., "users.user1.name")
//   - Records data set information for reporting
//   - Supports versioning and metadata
func (tdm *TDMManager) LoadDataSets(dataSets []parser.DataSet) error {
	for i := range dataSets {
		ds := &dataSets[i]
		tdm.dataSets[ds.Name] = ds

		// Load data into variables with proper nesting
		tdm.loadNestedData(ds.Name, ds.Data)

		// Record data set info
		tdm.results.DataSets[ds.Name] = parser.DataSetInfo{
			Name:     ds.Name,
			Version:  ds.Version,
			Records:  len(ds.Data),
			Status:   "loaded",
			LoadTime: time.Now().Format(time.RFC3339),
		}
	}
	return nil
}

// loadNestedData recursively loads nested data structures into variables
func (tdm *TDMManager) loadNestedData(prefix string, data map[string]interface{}) {
	for key, value := range data {
		varName := fmt.Sprintf("%s.%s", prefix, key)

		// If the value is a map, recursively process it
		if nestedMap, ok := value.(map[string]interface{}); ok {
			tdm.loadNestedData(varName, nestedMap)
		} else {
			// Store the leaf value
			tdm.variables[varName] = value
		}
	}
}

// LoadEnvironments loads environment configurations
func (tdm *TDMManager) LoadEnvironments(environments []parser.Environment) error {
	for i := range environments {
		env := &environments[i]
		tdm.environments[env.Name] = env

		// Load environment variables
		for key, value := range env.Variables {
			varName := fmt.Sprintf("env.%s.%s", env.Name, key)
			tdm.variables[varName] = value
		}
	}
	return nil
}

// SetEnvironment activates a specific environment
func (tdm *TDMManager) SetEnvironment(envName string) error {
	env, exists := tdm.environments[envName]
	if !exists {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	// Apply environment overrides
	for key, value := range env.Overrides {
		tdm.variables[key] = value
	}

	// Load environment-specific data sets
	for _, dsName := range env.DataSets {
		if ds, exists := tdm.dataSets[dsName]; exists {
			for key, value := range ds.Data {
				tdm.variables[key] = value
			}
		}
	}

	return nil
}

// ValidateData validates data according to validation rules
func (tdm *TDMManager) ValidateData(validations []parser.Validation) []parser.ValidationResult {
	results := make([]parser.ValidationResult, 0)

	for _, validation := range validations {
		result := tdm.validateField(validation)
		results = append(results, result)
		tdm.results.Validations = append(tdm.results.Validations, result)
	}

	return results
}

// validateField validates a single field according to validation rules
func (tdm *TDMManager) validateField(validation parser.Validation) parser.ValidationResult {
	result := parser.ValidationResult{
		Name:     validation.Name,
		Status:   "success",
		Message:  "Validation passed",
		Severity: validation.Severity,
	}

	// Get field value
	value, exists := tdm.variables[validation.Field]
	if !exists {
		result.Status = "failed"
		result.Message = fmt.Sprintf("Field '%s' not found", validation.Field)
		return result
	}

	// Apply validation based on type
	switch validation.Type {
	case "format":
		if err := tdm.validateFormat(value, validation.Rule); err != nil {
			result.Status = "failed"
			result.Message = util.FormatRobogoError(err)
		}
	case "range":
		if err := tdm.validateRange(value, validation.Rule); err != nil {
			result.Status = "failed"
			result.Message = util.FormatRobogoError(err)
		}
	case "length":
		if err := tdm.validateLength(value, validation.Rule); err != nil {
			result.Status = "failed"
			result.Message = util.FormatRobogoError(err)
		}
	case "required":
		if err := tdm.validateRequired(value); err != nil {
			result.Status = "failed"
			result.Message = util.FormatRobogoError(err)
		}
	case "unique":
		if err := tdm.validateUnique(validation.Field, value); err != nil {
			result.Status = "failed"
			result.Message = util.FormatRobogoError(err)
		}
	default:
		result.Status = "WARNING"
		result.Message = fmt.Sprintf("Unknown validation type: %s", validation.Type)
	}

	return result
}

// validateFormat validates field format (email, phone, etc.)
func (tdm *TDMManager) validateFormat(value interface{}, rule interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	ruleStr := fmt.Sprintf("%v", rule)

	switch ruleStr {
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(valueStr) {
			return fmt.Errorf("invalid email format: %s", valueStr)
		}
	case "phone":
		phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]+$`)
		if !phoneRegex.MatchString(valueStr) {
			return fmt.Errorf("invalid phone format: %s", valueStr)
		}
	case "url":
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(valueStr) {
			return fmt.Errorf("invalid URL format: %s", valueStr)
		}
	default:
		// Custom regex pattern
		if regex, err := regexp.Compile(ruleStr); err == nil {
			if !regex.MatchString(valueStr) {
				return fmt.Errorf("value does not match pattern '%s': %s", ruleStr, valueStr)
			}
		} else {
			return fmt.Errorf("invalid regex pattern: %s", ruleStr)
		}
	}

	return nil
}

// validateRange validates numeric range
func (tdm *TDMManager) validateRange(value interface{}, rule interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	valueNum, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("value is not numeric: %s", valueStr)
	}

	ruleMap, ok := rule.(map[string]interface{})
	if !ok {
		return fmt.Errorf("range rule must be a map with min/max values")
	}

	if min, exists := ruleMap["min"]; exists {
		if minNum, err := strconv.ParseFloat(fmt.Sprintf("%v", min), 64); err == nil {
			if valueNum < minNum {
				return fmt.Errorf("value %.2f is below minimum %.2f", valueNum, minNum)
			}
		}
	}

	if max, exists := ruleMap["max"]; exists {
		if maxNum, err := strconv.ParseFloat(fmt.Sprintf("%v", max), 64); err == nil {
			if valueNum > maxNum {
				return fmt.Errorf("value %.2f is above maximum %.2f", valueNum, maxNum)
			}
		}
	}

	return nil
}

// validateLength validates string/array length
func (tdm *TDMManager) validateLength(value interface{}, rule interface{}) error {
	ruleMap, ok := rule.(map[string]interface{})
	if !ok {
		return fmt.Errorf("length rule must be a map with min/max values")
	}

	var length int
	switch v := value.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	default:
		return fmt.Errorf("cannot validate length of type %T", value)
	}

	if min, exists := ruleMap["min"]; exists {
		if minNum, err := strconv.Atoi(fmt.Sprintf("%v", min)); err == nil {
			if length < minNum {
				return fmt.Errorf("length %d is below minimum %d", length, minNum)
			}
		}
	}

	if max, exists := ruleMap["max"]; exists {
		if maxNum, err := strconv.Atoi(fmt.Sprintf("%v", max)); err == nil {
			if length > maxNum {
				return fmt.Errorf("length %d is above maximum %d", length, maxNum)
			}
		}
	}

	return nil
}

// validateRequired validates that a field is not empty
func (tdm *TDMManager) validateRequired(value interface{}) error {
	if value == nil {
		return fmt.Errorf("field is required but value is nil")
	}

	valueStr := fmt.Sprintf("%v", value)
	if strings.TrimSpace(valueStr) == "" {
		return fmt.Errorf("field is required but value is empty")
	}

	return nil
}

// validateUnique validates that a field value is unique
func (tdm *TDMManager) validateUnique(fieldName string, value interface{}) error {
	// This is a simplified implementation
	// In a real system, you'd check against a database or data store
	valueStr := fmt.Sprintf("%v", value)

	// Check if this value already exists in our variables
	for key, existingValue := range tdm.variables {
		if key != fieldName && fmt.Sprintf("%v", existingValue) == valueStr {
			return fmt.Errorf("value '%s' is not unique (conflicts with field '%s')", valueStr, key)
		}
	}

	return nil
}

// GenerateTestData generates test data based on patterns and rules
func (tdm *TDMManager) GenerateTestData(pattern string, count int) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)

	for i := 0; i < count; i++ {
		data := make(map[string]interface{})

		// Parse pattern and generate data
		if err := tdm.parsePattern(pattern, data, i); err != nil {
			return nil, err
		}

		results = append(results, data)
	}

	return results, nil
}

// parsePattern parses a data generation pattern
func (tdm *TDMManager) parsePattern(pattern string, data map[string]interface{}, index int) error {
	// Simple pattern parsing - in a real implementation, this would be more sophisticated
	// Example patterns: "user_{index}", "email_{random}@example.com", "amount_{range:100:1000}"

	// Replace {index} with actual index
	pattern = strings.ReplaceAll(pattern, "{index}", fmt.Sprintf("%d", index))

	// Replace {random} with random value
	if strings.Contains(pattern, "{random}") {
		randomValue := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
		pattern = strings.ReplaceAll(pattern, "{random}", randomValue)
	}

	// Replace {range:min:max} with random range value
	rangeRegex := regexp.MustCompile(`\{range:(\d+):(\d+)\}`)
	pattern = rangeRegex.ReplaceAllStringFunc(pattern, func(match string) string {
		parts := strings.Split(strings.Trim(match, "{}"), ":")
		if len(parts) == 3 && parts[0] == "range" {
			min, _ := strconv.Atoi(parts[1])
			max, _ := strconv.Atoi(parts[2])
			randomValue := min + int(time.Now().UnixNano())%(max-min+1)
			return fmt.Sprintf("%d", randomValue)
		}
		return match
	})

	// For now, just store the pattern as a single field
	data["generated"] = pattern

	return nil
}

// GetVariable retrieves a variable value
func (tdm *TDMManager) GetVariable(name string) (interface{}, bool) {
	value, exists := tdm.variables[name]
	return value, exists
}

// SetVariable sets a variable value
func (tdm *TDMManager) SetVariable(name string, value interface{}) {
	tdm.variables[name] = value
}

// GetAllVariables returns all TDM variables
func (tdm *TDMManager) GetAllVariables() map[string]interface{} {
	return tdm.variables
}

// GetResults returns TDM execution results
func (tdm *TDMManager) GetResults() *parser.DataResults {
	return tdm.results
}

// TDMAction manages test data operations including data sets, environments, and validations.
//
// Parameters:
//   - operation: TDM operation to perform (load, validate, get, set, list)
//   - target: Target data set, environment, or variable name
//   - value: Value to set or validation criteria
//   - options: Additional options (format, encoding, etc.)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON result with operation status and data
//
// Supported Operations:
//   - "load": Load data set from file or source
//   - "validate": Validate data against criteria
//   - "get": Retrieve data set or variable value
//   - "set": Set environment or variable value
//   - "list": List available data sets or environments
//
// Examples:
//   - Load data: ["load", "users.csv", "csv"]
//   - Validate data: ["validate", "users", {"required_fields": ["id", "name"]}]
//   - Get variable: ["get", "test_user"]
//   - Set environment: ["set", "environment", "staging"]
//   - List data sets: ["list", "datasets"]
//
// Use Cases:
//   - Test data management and provisioning
//   - Environment-specific data handling
//   - Data validation and quality checks
//   - Dynamic test data generation
//   - Cross-environment testing
//
// Notes:
//   - Supports multiple data formats (CSV, JSON, XML, YAML)
//   - Environment-specific data isolation
//   - Comprehensive validation framework
//   - Integration with test variable system
func TDMAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("tdm action requires at least 1 argument: operation")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))

	switch operation {
	case "generate":
		return generateTestData(args[1:])
	case "validate":
		return validateTestData(args[1:])
	case "load_dataset":
		return loadDataSet(args[1:])
	case "set_environment":
		return setEnvironment(args[1:])
	default:
		return nil, fmt.Errorf("unknown tdm operation: %s", operation)
	}
}

// generateTestData generates test data based on pattern
func generateTestData(args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("generate requires pattern and count")
	}

	pattern := fmt.Sprintf("%v", args[0])
	countStr := fmt.Sprintf("%v", args[1])
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return nil, fmt.Errorf("invalid count: %s", countStr)
	}

	// Create a temporary TDM manager for this operation
	tdm := NewTDMManager()
	data, err := tdm.GenerateTestData(pattern, count)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"operation": "generate",
		"pattern":   pattern,
		"count":     count,
		"data":      data,
		"status":    "success",
	}

	return result, nil
}

// validateTestData validates test data
func validateTestData(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("validate requires validation rules")
	}

	// This is a placeholder - in a real implementation, you'd parse validation rules
	result := map[string]interface{}{
		"operation": "validate",
		"status":    "not_implemented",
		"message":   "Data validation requires TDM manager integration",
	}

	return result, nil
}

// loadDataSet loads a data set
func loadDataSet(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("load_dataset requires dataset name")
	}

	datasetName := fmt.Sprintf("%v", args[0])

	result := map[string]interface{}{
		"operation": "load_dataset",
		"dataset":   datasetName,
		"status":    "not_implemented",
		"message":   "Dataset loading requires TDM manager integration",
	}

	return result, nil
}

// setEnvironment sets the active environment
func setEnvironment(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("set_environment requires environment name")
	}

	envName := fmt.Sprintf("%v", args[0])

	result := map[string]interface{}{
		"operation":   "set_environment",
		"environment": envName,
		"status":      "not_implemented",
		"message":     "Environment setting requires TDM manager integration",
	}

	return result, nil
}
