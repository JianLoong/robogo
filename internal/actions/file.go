package actions

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
	"gopkg.in/yaml.v3"
)

// fileReadAction reads a file and returns its content
// Args: [file_path] - path to the file to read
// Options: format - force format detection (json, yaml, csv, text)
func fileReadAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("file_read", 1, len(args))
	}

	filePath := fmt.Sprintf("%v", args[0])

	// Security: Clean the path to prevent path traversal attacks
	cleanPath := filepath.Clean(filePath)

	// Security: Prevent absolute paths that could access system files
	if filepath.IsAbs(cleanPath) && !isAllowedAbsolutePath(cleanPath) {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "UNSAFE_FILE_PATH").
			WithTemplate("Absolute file paths are restricted for security").
			WithContext("file_path", filePath).
			WithContext("clean_path", cleanPath).
			WithSuggestion("Use relative paths from your test directory").
			WithSuggestion("Allowed absolute paths must be in current working directory").
			Build(fmt.Sprintf("unsafe absolute path: %s", cleanPath))
	}

	// Security: Prevent path traversal
	if strings.Contains(cleanPath, "..") {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "PATH_TRAVERSAL_DETECTED").
			WithTemplate("Path traversal detected in file path").
			WithContext("file_path", filePath).
			WithContext("clean_path", cleanPath).
			WithSuggestion("Use relative paths without '..' components").
			Build(fmt.Sprintf("path traversal detected: %s", cleanPath))
	}

	// Check if file exists and is readable
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return types.NewErrorBuilder(types.ErrorCategoryValidation, "FILE_NOT_FOUND").
				WithTemplate("File not found").
				WithContext("file_path", filePath).
				WithContext("clean_path", cleanPath).
				WithSuggestion("Check that the file path is correct").
				WithSuggestion("Ensure the file exists relative to your test location").
				Build(fmt.Sprintf("file not found: %s", cleanPath))
		}
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "FILE_ACCESS_ERROR").
			WithTemplate("Cannot access file").
			WithContext("file_path", filePath).
			WithContext("error", err.Error()).
			WithSuggestion("Check file permissions").
			Build(fmt.Sprintf("file access error: %s", err.Error()))
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "PATH_IS_DIRECTORY").
			WithTemplate("Path points to a directory, not a file").
			WithContext("file_path", filePath).
			WithSuggestion("Specify a file path, not a directory").
			Build(fmt.Sprintf("path is directory: %s", cleanPath))
	}

	// Read file content
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "FILE_READ_ERROR").
			WithTemplate("Failed to read file content").
			WithContext("file_path", filePath).
			WithContext("error", err.Error()).
			WithSuggestion("Check file permissions and disk space").
			Build(fmt.Sprintf("file read error: %s", err.Error()))
	}

	// Determine format (from options or file extension)
	format := determineFileFormat(cleanPath, options)

	// Parse content based on format
	parsedContent, parseErr := parseFileContent(content, format, cleanPath)
	if parseErr != nil {
		return *parseErr
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"content":    parsedContent,
			"file_path":  cleanPath,
			"format":     format,
			"size_bytes": len(content),
		},
	}
}

// isAllowedAbsolutePath checks if an absolute path is allowed (within current working directory)
func isAllowedAbsolutePath(path string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Check if the absolute path is within the current working directory
	relPath, err := filepath.Rel(cwd, path)
	if err != nil {
		return false
	}

	// If relative path starts with "..", it's outside the working directory
	return !strings.HasPrefix(relPath, "..")
}

// determineFileFormat determines the file format from extension or options
func determineFileFormat(filePath string, options map[string]any) string {
	// Check if format is explicitly specified in options
	if format, ok := options["format"].(string); ok {
		return strings.ToLower(format)
	}

	// Determine from file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".csv":
		return "csv"
	default:
		return "text"
	}
}

// parseFileContent parses file content based on format
func parseFileContent(content []byte, format, filePath string) (any, *types.ActionResult) {
	switch format {
	case "json":
		return parseJSONContent(content, filePath)
	case "yaml":
		return parseYAMLContent(content, filePath)
	case "csv":
		return parseCSVContent(content, filePath)
	case "text":
		return string(content), nil
	default:
		// Unknown format, return as text
		return string(content), nil
	}
}

// parseJSONContent parses JSON content
func parseJSONContent(content []byte, filePath string) (any, *types.ActionResult) {
	var result any
	if err := json.Unmarshal(content, &result); err != nil {
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_JSON").
			WithTemplate("Invalid JSON format in file").
			WithContext("file_path", filePath).
			WithContext("error", err.Error()).
			WithSuggestion("Verify the JSON syntax is valid").
			WithSuggestion("Use a JSON validator to check the file").
			Build(fmt.Sprintf("JSON parse error: %s", err.Error()))
		return nil, &errorResult
	}
	return result, nil
}

// parseYAMLContent parses YAML content
func parseYAMLContent(content []byte, filePath string) (any, *types.ActionResult) {
	var result any
	if err := yaml.Unmarshal(content, &result); err != nil {
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_YAML").
			WithTemplate("Invalid YAML format in file").
			WithContext("file_path", filePath).
			WithContext("error", err.Error()).
			WithSuggestion("Verify the YAML syntax is valid").
			WithSuggestion("Check indentation and structure").
			Build(fmt.Sprintf("YAML parse error: %s", err.Error()))
		return nil, &errorResult
	}
	return result, nil
}

// parseCSVContent parses CSV content into array of objects
func parseCSVContent(content []byte, filePath string) (any, *types.ActionResult) {
	reader := csv.NewReader(strings.NewReader(string(content)))

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_CSV").
			WithTemplate("Invalid CSV format in file").
			WithContext("file_path", filePath).
			WithContext("error", err.Error()).
			WithSuggestion("Verify the CSV format is valid").
			WithSuggestion("Check for proper comma separation and quoting").
			Build(fmt.Sprintf("CSV parse error: %s", err.Error()))
		return nil, &errorResult
	}

	if len(records) == 0 {
		return []any{}, nil
	}

	// First row is headers
	headers := records[0]
	var result []any

	// Convert remaining rows to objects using headers as keys
	for i := 1; i < len(records); i++ {
		row := records[i]
		obj := make(map[string]any)

		for j, header := range headers {
			if j < len(row) {
				obj[header] = row[j]
			} else {
				obj[header] = ""
			}
		}
		result = append(result, obj)
	}

	return result, nil
}
