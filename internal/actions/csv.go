package actions

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// csvParseAction parses CSV data from file or string
// Args: [source] - file path or CSV string content
// Options:
//   - delimiter: field separator (default: ",")
//   - skip_header: treat first row as headers (default: true)
//   - max_rows: limit rows parsed (default: unlimited, 0 = unlimited)
//   - trim_spaces: remove leading/trailing spaces (default: true)
//   - quote_char: quote character (default: '"')
func csvParseAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("csv_parse", 1, len(args))
	}

	// Validate arguments are resolved
	if errorResult := validateArgsResolved("csv_parse", args); errorResult != nil {
		return *errorResult
	}

	source := fmt.Sprintf("%v", args[0])
	if source == "" {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "EMPTY_SOURCE").
			WithTemplate("CSV parse source cannot be empty").
			WithSuggestion("Provide a valid file path or CSV string content").
			Build("empty source provided")
	}

	// Parse options
	delimiter := parseStringOption(options, "delimiter", ",")
	skipHeader := parseBoolOption(options, "skip_header", true)
	maxRows := parseIntOption(options, "max_rows", 0) // 0 = unlimited
	trimSpaces := parseBoolOption(options, "trim_spaces", true)
	quoteChar := parseStringOption(options, "quote_char", "\"")

	// Validate delimiter
	if len(delimiter) != 1 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_DELIMITER").
			WithTemplate("CSV delimiter must be exactly one character").
			WithContext("delimiter", delimiter).
			WithSuggestion("Use a single character like ',' or ';' or '\\t' for tab").
			Build(fmt.Sprintf("invalid delimiter: %s", delimiter))
	}

	// Validate quote character
	if len(quoteChar) != 1 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_QUOTE_CHAR").
			WithTemplate("CSV quote character must be exactly one character").
			WithContext("quote_char", quoteChar).
			WithSuggestion("Use a single character like '\"' or '\\''").
			Build(fmt.Sprintf("invalid quote character: %s", quoteChar))
	}

	// Validate max_rows
	if maxRows < 0 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_MAX_ROWS").
			WithTemplate("CSV max_rows must be 0 (unlimited) or positive integer").
			WithContext("max_rows", maxRows).
			WithSuggestion("Use 0 for unlimited rows or a positive number").
			Build(fmt.Sprintf("invalid max_rows: %d", maxRows))
	}

	// Determine if source is a file path or CSV content
	var reader io.Reader
	isFilePath := false

	// Check if it looks like a file path (doesn't contain newlines and commas suggest it's content)
	if !strings.Contains(source, "\n") && !strings.Contains(source, delimiter) {
		// Try to open as file first
		if file, err := os.Open(source); err == nil {
			reader = file
			isFilePath = true
			defer file.Close()
		}
	}

	// If not a file or file doesn't exist, treat as CSV content
	if reader == nil {
		reader = strings.NewReader(source)
	}

	fmt.Printf("📊 Parsing CSV %s...\n", func() string {
		if isFilePath {
			return fmt.Sprintf("file: %s", source)
		}
		return "content"
	}())

	// Parse CSV
	result := parseCSVData(reader, delimiter, skipHeader, maxRows, trimSpaces, rune(quoteChar[0]))
	return result
}

// parseCSVData performs the actual CSV parsing
func parseCSVData(reader io.Reader, delimiter string, skipHeader bool, maxRows int, trimSpaces bool, quoteChar rune) types.ActionResult {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = rune(delimiter[0])
	csvReader.TrimLeadingSpace = trimSpaces
	
	// Note: Go's csv.Reader uses '"' as quote character by default

	// Read all records
	allRecords, err := csvReader.ReadAll()
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "CSV_PARSE_ERROR").
			WithTemplate("Failed to parse CSV data").
			WithContext("error", err.Error()).
			WithSuggestion("Check CSV format, delimiters, and quote characters").
			WithSuggestion("Ensure CSV is well-formed with consistent column counts").
			Build(fmt.Sprintf("CSV parsing failed: %s", err.Error()))
	}

	if len(allRecords) == 0 {
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"rows":         []map[string]any{},
				"headers":      []string{},
				"row_count":    0,
				"column_count": 0,
				"has_header":   skipHeader,
			},
		}
	}

	var headers []string
	var dataRows [][]string
	startRow := 0

	// Handle headers
	if skipHeader && len(allRecords) > 0 {
		headers = allRecords[0]
		if trimSpaces {
			for i, header := range headers {
				headers[i] = strings.TrimSpace(header)
			}
		}
		startRow = 1
	} else {
		// Generate numeric headers
		if len(allRecords) > 0 {
			for i := 0; i < len(allRecords[0]); i++ {
				headers = append(headers, fmt.Sprintf("column_%d", i))
			}
		}
	}

	// Extract data rows
	for i := startRow; i < len(allRecords); i++ {
		if maxRows > 0 && len(dataRows) >= maxRows {
			break
		}
		
		row := allRecords[i]
		if trimSpaces {
			for j, cell := range row {
				row[j] = strings.TrimSpace(cell)
			}
		}
		dataRows = append(dataRows, row)
	}

	// Convert to array of objects
	var rows []map[string]any
	for _, row := range dataRows {
		rowObj := make(map[string]any)
		for i, cell := range row {
			if i < len(headers) {
				rowObj[headers[i]] = cell
			} else {
				// Handle rows with more columns than headers
				rowObj[fmt.Sprintf("column_%d", i)] = cell
			}
		}
		rows = append(rows, rowObj)
	}

	columnCount := len(headers)
	if len(allRecords) > 0 && len(allRecords[0]) > columnCount {
		columnCount = len(allRecords[0])
	}

	fmt.Printf("✅ CSV parsed: %d rows, %d columns\n", len(rows), columnCount)

	// Create result data
	resultData := map[string]any{
		"rows":         rows,
		"headers":      headers,
		"row_count":    len(rows),
		"column_count": columnCount,
		"has_header":   skipHeader,
		"raw_records":  allRecords, // Include raw data for advanced processing
	}

	// Convert to JSON and back to ensure jq compatibility
	jsonBytes, err := json.Marshal(resultData)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategorySystem, "JSON_MARSHAL_ERROR").
			WithTemplate("Failed to convert CSV data to JSON-compatible format").
			WithContext("error", err.Error()).
			Build(fmt.Sprintf("JSON marshaling failed: %s", err.Error()))
	}

	var jsonCompatibleData any
	if err := json.Unmarshal(jsonBytes, &jsonCompatibleData); err != nil {
		return types.NewErrorBuilder(types.ErrorCategorySystem, "JSON_UNMARSHAL_ERROR").
			WithTemplate("Failed to unmarshal JSON data").
			WithContext("error", err.Error()).
			Build(fmt.Sprintf("JSON unmarshaling failed: %s", err.Error()))
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   jsonCompatibleData,
	}
}

// Helper functions for CSV parsing
func parseStringOption(options map[string]any, key string, defaultValue string) string {
	if val, exists := options[key]; exists {
		return fmt.Sprintf("%v", val)
	}
	return defaultValue
}