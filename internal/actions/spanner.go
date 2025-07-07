package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SpannerManager manages Spanner connections
type SpannerManager struct {
	connections map[string]*spanner.Client
	mutex       sync.RWMutex
}

// SpannerQueryResult represents the result of a Spanner query
type SpannerQueryResult struct {
	Query        string                   `json:"query"`
	RowsAffected int64                    `json:"rows_affected,omitempty"`
	Columns      []string                 `json:"columns,omitempty"`
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	Duration     time.Duration            `json:"duration"`
	Error        string                   `json:"error,omitempty"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
}

// Global Spanner manager instance
var spannerManager = &SpannerManager{
	connections: make(map[string]*spanner.Client),
}

// SpannerAction performs Google Cloud Spanner operations with comprehensive support for queries, transactions, and connection management.
//
// Parameters:
//   - operation: Spanner operation to perform (query, execute, connect, close, test)
//   - connection: Spanner connection string (projects/project-id/instances/instance-id/databases/database-id)
//   - query: SQL query or statement to execute
//   - params: Query parameters (optional, for parameterized queries)
//   - options: Additional options (timeout, emulator settings, etc.)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON result with operation status, data, and timing information
//
// Supported Operations:
//   - "connect": Establish Spanner connection
//   - "query": Execute SELECT query and return results
//   - "execute": Execute INSERT/UPDATE/DELETE statement
//   - "close": Close Spanner connection
//   - "test": Test connection without executing queries
//
// Examples:
//   - Connect: ["connect", "projects/your-project-id/instances/test-instance/databases/test-db"]
//   - Query: ["query", "SELECT * FROM users WHERE id = @id", {"id": 123}]
//   - Execute: ["execute", "INSERT INTO users (name, email) VALUES (@name, @email)", {"name": "John", "email": "john@example.com"}]
//   - Close: ["close"]
//
// Use Cases:
//   - Spanner testing and validation
//   - Data setup and teardown
//   - Integration testing with Spanner
//   - Performance testing of Spanner operations
//   - Data verification and assertions
//
// Notes:
//   - Supports parameterized queries for security
//   - Automatic connection pooling and management
//   - Comprehensive error handling and timeout support
//   - Results available for assertions and variable storage
//   - Supports both real Spanner and emulator connections
func SpannerAction(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("spanner action requires at least 2 arguments: operation and connection_string")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	switch operation {
	case "query", "select":
		return executeSpannerQuery(connectionString, args[2:], silent)
	case "execute", "insert", "update", "delete":
		return executeSpannerStatement(connectionString, args[2:], silent)
	case "connect":
		return testSpannerConnection(connectionString)
	case "close":
		return closeSpannerConnection(connectionString)
	default:
		return nil, fmt.Errorf("unknown spanner operation: %s", operation)
	}
}

// executeSpannerQuery executes a SELECT query and returns results
func executeSpannerQuery(connectionString string, args []interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("query operation requires a SQL query")
	}

	query := fmt.Sprintf("%v", args[0])
	var queryParams map[string]interface{}

	// Extract query parameters if provided
	if len(args) > 1 {
		if params, ok := args[1].(map[string]interface{}); ok {
			queryParams = params
		}
	}

	// Get or create Spanner client
	client, err := getSpannerClient(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to get Spanner client: %w", err)
	}

	startTime := time.Now()

	// Execute query
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stmt := spanner.Statement{SQL: query, Params: queryParams}
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	duration := time.Since(startTime)

	// Get column names from first row
	var results []map[string]interface{}
	var columns []string

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate query results: %w", err)
		}

		// Get column names from first row
		if len(columns) == 0 {
			columns = row.ColumnNames()
		}

		// Convert row to map
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			// Try different types in order of preference
			// Start with int64 since that's what we're getting
			var intVal int64
			if err := row.Column(i, &intVal); err == nil {
				rowMap[col] = intVal
				continue
			}

			// Try string
			var strVal string
			if err := row.Column(i, &strVal); err == nil {
				rowMap[col] = strVal
				continue
			}

			// Try float64
			var floatVal float64
			if err := row.Column(i, &floatVal); err == nil {
				rowMap[col] = floatVal
				continue
			}

			// Try bool
			var boolVal bool
			if err := row.Column(i, &boolVal); err == nil {
				rowMap[col] = boolVal
				continue
			}

			// Try time.Time
			var timeVal time.Time
			if err := row.Column(i, &timeVal); err == nil {
				rowMap[col] = timeVal.Format(time.RFC3339)
				continue
			}

			// Last resort: try interface{} for unknown types
			var interfaceVal interface{}
			if err := row.Column(i, &interfaceVal); err != nil {
				return nil, fmt.Errorf("failed to get column value for %s: %w", col, err)
			}
			rowMap[col] = interfaceVal
		}
		results = append(results, rowMap)
	}

	// Transform to consistent format
	transformedResult := transformSpannerToConsistentFormat(columns, results, query)

	// Add rich metadata
	result := map[string]interface{}{
		"query":    query,
		"columns":  transformedResult.(map[string]interface{})["columns"],
		"rows":     transformedResult.(map[string]interface{})["rows"],
		"duration": duration,
		"metadata": map[string]interface{}{
			"row_count": len(results),
			"params":    queryParams,
		},
	}

	// Copy value/values fields if they exist
	if value, exists := transformedResult.(map[string]interface{})["value"]; exists {
		result["value"] = value
	}
	if values, exists := transformedResult.(map[string]interface{})["values"]; exists {
		result["values"] = values
	}

	// Convert to JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return jsonResult, nil
}

// executeSpannerStatement executes an INSERT/UPDATE/DELETE statement
func executeSpannerStatement(connectionString string, args []interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("execute operation requires a SQL statement")
	}

	statement := fmt.Sprintf("%v", args[0])
	var statementParams map[string]interface{}

	// Extract statement parameters if provided
	if len(args) > 1 {
		if params, ok := args[1].(map[string]interface{}); ok {
			statementParams = params
		}
	}

	// Get or create Spanner client
	client, err := getSpannerClient(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to get Spanner client: %w", err)
	}

	startTime := time.Now()

	// Execute statement
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stmt := spanner.Statement{SQL: statement, Params: statementParams}
	var rowsAffected int64

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		count, err := txn.Update(ctx, stmt)
		if err != nil {
			return err
		}
		rowsAffected = count
		return nil
	})

	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("failed to execute Spanner statement: %w", err)
	}

	// Create result
	result := &SpannerQueryResult{
		Query:        statement,
		RowsAffected: rowsAffected,
		Duration:     duration,
		Metadata: map[string]interface{}{
			"params": statementParams,
		},
	}

	// Convert to JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return jsonResult, nil
}

// testSpannerConnection tests the Spanner connection
func testSpannerConnection(connectionString string) (interface{}, error) {
	client, err := getSpannerClient(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Spanner: %w", err)
	}

	// Test the connection with a simple query
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stmt := spanner.Statement{SQL: "SELECT 1"}
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	_, err = iter.Next()
	if err != nil && err != iterator.Done {
		return nil, fmt.Errorf("failed to test Spanner connection: %w", err)
	}

	result := map[string]interface{}{
		"status":            "connected",
		"connection_string": connectionString,
		"message":           "Spanner connection test successful",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return jsonResult, nil
}

// closeSpannerConnection closes the Spanner connection
func closeSpannerConnection(connectionString string) (interface{}, error) {
	spannerManager.mutex.Lock()
	defer spannerManager.mutex.Unlock()

	if client, exists := spannerManager.connections[connectionString]; exists {
		client.Close()
		delete(spannerManager.connections, connectionString)
	}

	result := map[string]interface{}{
		"status":            "closed",
		"connection_string": connectionString,
		"message":           "Spanner connection closed successfully",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return jsonResult, nil
}

// getSpannerClient gets or creates a Spanner client
func getSpannerClient(connectionString string) (*spanner.Client, error) {
	spannerManager.mutex.Lock()
	defer spannerManager.mutex.Unlock()

	// Check if connection already exists
	if client, exists := spannerManager.connections[connectionString]; exists {
		return client, nil
	}

	// Check if using emulator
	useEmulator := strings.Contains(connectionString, "useEmulator=true") ||
		strings.Contains(connectionString, "localhost:9010")

	var client *spanner.Client
	var err error

	if useEmulator {
		// Connect to Spanner emulator
		ctx := context.Background()
		client, err = spanner.NewClient(ctx, connectionString, option.WithEndpoint("localhost:9010"),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Spanner emulator: %w", err)
		}
	} else {
		// Connect to real Spanner (requires credentials)
		ctx := context.Background()
		client, err = spanner.NewClient(ctx, connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Spanner: %w", err)
		}
	}

	// Store the connection
	spannerManager.connections[connectionString] = client

	return client, nil
}

// CloseAllSpannerConnections closes all Spanner connections
func CloseAllSpannerConnections() {
	spannerManager.mutex.Lock()
	defer spannerManager.mutex.Unlock()

	for connectionString, client := range spannerManager.connections {
		client.Close()
		delete(spannerManager.connections, connectionString)
	}
}

// transformSpannerToConsistentFormat transforms the result to a consistent format
func transformSpannerToConsistentFormat(columns []string, rows []map[string]interface{}, query string) interface{} {
	// Transform to consistent format
	transformedResult := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
	}

	// Add value/values fields if they exist
	if len(rows) == 1 && len(columns) == 1 {
		// Single row, single column
		transformedResult["value"] = rows[0][columns[0]]
	} else if len(rows) > 1 && len(columns) == 1 {
		// Multiple rows, single column
		var values []interface{}
		for _, row := range rows {
			values = append(values, row[columns[0]])
		}
		transformedResult["values"] = values
	}

	return transformedResult
}
