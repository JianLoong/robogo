package execution

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// applyExtraction applies the specified extraction to the data
func (s *BasicExecutionStrategy) applyExtraction(data any, config *types.ExtractConfig) (any, error) {
	if data == nil {
		return nil, types.NewNilDataError()
	}

	switch config.Type {
	case "jq":
		return s.applyJQExtraction(data, config.Path)
	case "xpath":
		return s.applyXPathExtraction(data, config.Path)
	case "regex":
		return s.applyRegexExtraction(data, config.Path, config.Group)
	case "csv":
		result, err := s.applyCSVExtraction(data, config)
		if err != nil {
			return nil, err
		}
		// Convert to JSON-compatible format for jq processing
		jsonBytes, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return nil, types.NewExtractionError(fmt.Sprintf("Failed to convert CSV result to JSON: %s", marshalErr.Error()))
		}
		var jsonResult any
		if unmarshalErr := json.Unmarshal(jsonBytes, &jsonResult); unmarshalErr != nil {
			return nil, types.NewExtractionError(fmt.Sprintf("Failed to unmarshal JSON result: %s", unmarshalErr.Error()))
		}
		return jsonResult, nil
	default:
		return nil, types.NewUnsupportedExtractionTypeError(config.Type)
	}
}

// applyJQExtraction applies JQ extraction to data
func (s *BasicExecutionStrategy) applyJQExtraction(data any, path string) (any, error) {
	jqAction, exists := s.actionRegistry.Get("jq")
	if !exists {
		return nil, types.NewExtractionError("jq action not available")
	}
	
	result := jqAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetMessage())
	}
	
	return result.Data, nil
}

// applyXPathExtraction applies XPath extraction to data  
func (s *BasicExecutionStrategy) applyXPathExtraction(data any, path string) (any, error) {
	xpathAction, exists := s.actionRegistry.Get("xpath")
	if !exists {
		return nil, types.NewExtractionError("xpath action not available")
	}
	
	result := xpathAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetMessage())
	}
	
	return result.Data, nil
}

// applyRegexExtraction applies regex extraction to data
func (s *BasicExecutionStrategy) applyRegexExtraction(data any, pattern string, group int) (any, error) {
	// Convert data to string
	var text string
	switch v := data.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		text = fmt.Sprintf("%v", v)
	}
	
	// Apply regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, types.NewInvalidRegexPatternError(pattern, err.Error())
	}
	
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return nil, types.NewNoRegexMatchError(pattern)
	}
	
	// Default to group 1, or use specified group
	if group == 0 {
		group = 1
	}
	
	if group >= len(matches) {
		return nil, types.NewInvalidCaptureGroupError(group, len(matches)-1)
	}
	
	return matches[group], nil
}

// applyCSVExtraction applies CSV extraction to data
func (s *BasicExecutionStrategy) applyCSVExtraction(data any, config *types.ExtractConfig) (any, error) {
	// Check if data is already parsed CSV from file_read
	if dataMap, ok := data.(map[string]any); ok {
		if content, hasContent := dataMap["content"]; hasContent {
			if format, hasFormat := dataMap["format"]; hasFormat && format == "csv" {
				// This is already parsed CSV data from file_read action
				return s.processStructuredCSVData(content, config)
			}
		}
	}

	// Convert data to string for parsing
	var csvContent string
	switch v := data.(type) {
	case string:
		csvContent = v
	case []byte:
		csvContent = string(v)
	default:
		csvContent = fmt.Sprintf("%v", v)
	}

	// Set defaults
	delimiter := ","
	if config.Delimiter != "" {
		delimiter = config.Delimiter
	}
	
	hasHeader := true  // default to true
	// Note: YAML omitempty means HasHeader is false by default, but we want true by default
	// If explicitly set to false, it will be false, otherwise default to true

	// Parse CSV content
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.Comma = rune(delimiter[0])
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, types.NewExtractionError(fmt.Sprintf("CSV parsing failed: %s", err.Error()))
	}

	if len(records) == 0 {
		return nil, types.NewExtractionError("CSV data is empty")
	}

	var headers []string
	var dataRows [][]string
	startRow := 0

	// Handle headers
	if hasHeader && len(records) > 0 {
		headers = records[0]
		startRow = 1
	} else {
		// Generate numeric headers
		if len(records) > 0 {
			for i := 0; i < len(records[0]); i++ {
				headers = append(headers, fmt.Sprintf("column_%d", i))
			}
		}
	}

	// Extract data rows
	for i := startRow; i < len(records); i++ {
		dataRows = append(dataRows, records[i])
	}

	// Check if we have data rows
	if len(dataRows) == 0 {
		return nil, types.NewExtractionError("No data rows found after parsing CSV content")
	}

	// Apply specific extraction based on config
	
	// If row is specified, extract specific row
	if config.Row != nil {
		rowIndex := *config.Row
		if rowIndex >= len(dataRows) {
			return nil, types.NewExtractionError(fmt.Sprintf("Row index %d out of range (max: %d)", rowIndex, len(dataRows)-1))
		}
		
		row := dataRows[rowIndex]
		
		// If column is also specified, extract specific cell
		if config.Column != "" {
			return s.extractCSVCell(row, headers, config.Column)
		}
		
		// Return entire row as object
		rowObj := make(map[string]any)
		for i, cell := range row {
			if i < len(headers) {
				rowObj[headers[i]] = cell
			} else {
				rowObj[fmt.Sprintf("column_%d", i)] = cell
			}
		}
		return rowObj, nil
	}
	
	// If only column is specified, extract entire column
	if config.Column != "" {
		return s.extractCSVColumn(dataRows, headers, config.Column)
	}
	
	// If filter is specified, apply simple filtering
	if config.Filter != "" {
		return s.applyCSVFilter(dataRows, headers, config.Filter)
	}
	
	// Default: return all data as array of objects
	var rows []map[string]any
	for _, row := range dataRows {
		rowObj := make(map[string]any)
		for i, cell := range row {
			if i < len(headers) {
				rowObj[headers[i]] = cell
			} else {
				rowObj[fmt.Sprintf("column_%d", i)] = cell
			}
		}
		rows = append(rows, rowObj)
	}
	
	return rows, nil
}

// extractCSVCell extracts a specific cell value
func (s *BasicExecutionStrategy) extractCSVCell(row []string, headers []string, column string) (any, error) {
	// Try column name first
	for i, header := range headers {
		if header == column {
			if i < len(row) {
				return row[i], nil
			}
			return "", nil
		}
	}
	
	// Try column index
	if colIndex, err := strconv.Atoi(column); err == nil {
		if colIndex >= 0 && colIndex < len(row) {
			return row[colIndex], nil
		}
		return nil, types.NewExtractionError(fmt.Sprintf("Column index %d out of range (max: %d)", colIndex, len(row)-1))
	}
	
	return nil, types.NewExtractionError(fmt.Sprintf("Column '%s' not found", column))
}

// extractCSVColumn extracts an entire column
func (s *BasicExecutionStrategy) extractCSVColumn(dataRows [][]string, headers []string, column string) (any, error) {
	var columnIndex int = -1
	
	// Try column name first
	for i, header := range headers {
		if header == column {
			columnIndex = i
			break
		}
	}
	
	// Try column index if name not found
	if columnIndex == -1 {
		if colIndex, err := strconv.Atoi(column); err == nil {
			if colIndex >= 0 && colIndex < len(headers) {
				columnIndex = colIndex
			}
		}
	}
	
	if columnIndex == -1 {
		return nil, types.NewExtractionError(fmt.Sprintf("Column '%s' not found", column))
	}
	
	var columnValues []string
	for _, row := range dataRows {
		if columnIndex < len(row) {
			columnValues = append(columnValues, row[columnIndex])
		} else {
			columnValues = append(columnValues, "")
		}
	}
	
	return columnValues, nil
}

// applyCSVFilter applies simple filtering to CSV data
func (s *BasicExecutionStrategy) applyCSVFilter(dataRows [][]string, headers []string, filter string) (any, error) {
	// Simple filter format: "column operator value"
	// e.g., "age > 25", "name == John", "status != active"
	
	parts := strings.Fields(filter)
	if len(parts) != 3 {
		return nil, types.NewExtractionError("CSV filter must be in format: 'column operator value'")
	}
	
	columnName := parts[0]
	operator := parts[1]
	filterValue := parts[2]
	
	// Find column index
	var columnIndex int = -1
	for i, header := range headers {
		if header == columnName {
			columnIndex = i
			break
		}
	}
	
	if columnIndex == -1 {
		return nil, types.NewExtractionError(fmt.Sprintf("Filter column '%s' not found", columnName))
	}
	
	var filteredRows []map[string]any
	for _, row := range dataRows {
		if columnIndex >= len(row) {
			continue
		}
		
		cellValue := row[columnIndex]
		matches := false
		
		switch operator {
		case "==", "=":
			matches = cellValue == filterValue
		case "!=", "<>":
			matches = cellValue != filterValue
		case ">":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue > filterNum
				}
			}
		case "<":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue < filterNum
				}
			}
		case ">=":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue >= filterNum
				}
			}
		case "<=":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue <= filterNum
				}
			}
		case "contains":
			matches = strings.Contains(cellValue, filterValue)
		default:
			return nil, types.NewExtractionError(fmt.Sprintf("Unsupported CSV filter operator: %s", operator))
		}
		
		if matches {
			rowObj := make(map[string]any)
			for i, cell := range row {
				if i < len(headers) {
					rowObj[headers[i]] = cell
				} else {
					rowObj[fmt.Sprintf("column_%d", i)] = cell
				}
			}
			filteredRows = append(filteredRows, rowObj)
		}
	}
	
	return filteredRows, nil
}

// processStructuredCSVData processes already-parsed CSV data from file_read
func (s *BasicExecutionStrategy) processStructuredCSVData(content any, config *types.ExtractConfig) (any, error) {
	// Convert content to array of maps
	var rows []map[string]any
	
	switch v := content.(type) {
	case []any:
		for _, item := range v {
			if rowMap, ok := item.(map[string]any); ok {
				rows = append(rows, rowMap)
			}
		}
	case []map[string]any:
		rows = v
	default:
		return nil, types.NewExtractionError("Invalid structured CSV data format")
	}

	if len(rows) == 0 {
		return nil, types.NewExtractionError("No CSV data rows found")
	}

	// Get headers from first row keys
	var headers []string
	for key := range rows[0] {
		headers = append(headers, key)
	}

	// Apply extraction based on config
	if config.Row != nil {
		rowIndex := *config.Row
		if rowIndex >= len(rows) {
			return nil, types.NewExtractionError(fmt.Sprintf("Row index %d out of range (max: %d)", rowIndex, len(rows)-1))
		}
		
		row := rows[rowIndex]
		
		// If column is also specified, extract specific cell
		if config.Column != "" {
			// Try to resolve column name (handle numeric indices)
			actualColumnName := config.Column
			if colIndex, err := strconv.Atoi(config.Column); err == nil {
				// It's a numeric index, convert to actual column name
				if colIndex >= 0 && colIndex < len(headers) {
					actualColumnName = headers[colIndex]
				} else {
					return nil, types.NewExtractionError(fmt.Sprintf("Column index %d out of range (max: %d)", colIndex, len(headers)-1))
				}
			}
			
			if value, exists := row[actualColumnName]; exists {
				return value, nil
			}
			return nil, types.NewExtractionError(fmt.Sprintf("Column '%s' not found", actualColumnName))
		}
		
		// Return entire row
		return row, nil
	}
	
	// If only column is specified, extract entire column
	if config.Column != "" {
		// Try to resolve column name (handle numeric indices)
		actualColumnName := config.Column
		if colIndex, err := strconv.Atoi(config.Column); err == nil {
			// It's a numeric index, convert to actual column name
			if colIndex >= 0 && colIndex < len(headers) {
				actualColumnName = headers[colIndex]
			} else {
				return nil, types.NewExtractionError(fmt.Sprintf("Column index %d out of range (max: %d)", colIndex, len(headers)-1))
			}
		}
		
		var columnValues []any
		for _, row := range rows {
			if value, exists := row[actualColumnName]; exists {
				columnValues = append(columnValues, value)
			} else {
				columnValues = append(columnValues, "")
			}
		}
		return columnValues, nil
	}
	
	// If filter is specified, apply simple filtering
	if config.Filter != "" {
		return s.applyStructuredCSVFilter(rows, config.Filter)
	}
	
	// Default: return all rows
	return rows, nil
}

// applyStructuredCSVFilter applies filtering to already-structured CSV data
func (s *BasicExecutionStrategy) applyStructuredCSVFilter(rows []map[string]any, filter string) (any, error) {
	// Simple filter format: "column operator value"
	parts := strings.Fields(filter)
	if len(parts) != 3 {
		return nil, types.NewExtractionError("CSV filter must be in format: 'column operator value'")
	}
	
	columnName := parts[0]
	operator := parts[1]
	filterValue := parts[2]
	
	var filteredRows []map[string]any
	for _, row := range rows {
		value, exists := row[columnName]
		if !exists {
			continue
		}
		
		cellValue := fmt.Sprintf("%v", value)
		matches := false
		
		switch operator {
		case "==", "=":
			matches = cellValue == filterValue
		case "!=", "<>":
			matches = cellValue != filterValue
		case ">":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue > filterNum
				}
			}
		case "<":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue < filterNum
				}
			}
		case ">=":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue >= filterNum
				}
			}
		case "<=":
			if numValue, err := strconv.ParseFloat(cellValue, 64); err == nil {
				if filterNum, err := strconv.ParseFloat(filterValue, 64); err == nil {
					matches = numValue <= filterNum
				}
			}
		case "contains":
			matches = strings.Contains(cellValue, filterValue)
		default:
			return nil, types.NewExtractionError(fmt.Sprintf("Unsupported CSV filter operator: %s", operator))
		}
		
		if matches {
			filteredRows = append(filteredRows, row)
		}
	}
	
	return filteredRows, nil
}

