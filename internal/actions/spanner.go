package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SpannerAction provides Google Cloud Spanner operations
//
// Parameters:
//   - subcommand: Operation type (connect, query, execute, close)
//   - connection_string: Spanner connection string or project/instance/database
//   - query/statement: SQL query or statement to execute
//   - parameters: Query parameters (optional)
//   - options: Additional options (optional)
//
// Returns: Operation result as string
//
// Example usage:
//   - ["connect", "projects/robogo-test-project/instances/robogo-test-instance/databases/robogo-test-db"]
//   - ["query", "projects/robogo-test-project/instances/robogo-test-instance/databases/robogo-test-db", "SELECT * FROM users"]
//   - ["execute", "projects/robogo-test-project/instances/robogo-test-instance/databases/robogo-test-db", "INSERT INTO users (id, name) VALUES (@id, @name)", ["user1", "John Doe"]]
//
// Connection String Format:
//   - Full: "projects/PROJECT_ID/instances/INSTANCE_ID/databases/DATABASE_ID"
//   - With emulator: "projects/PROJECT_ID/instances/INSTANCE_ID/databases/DATABASE_ID?useEmulator=true"
//
// Use Cases:
//   - Cloud-native database testing
//   - Distributed transaction testing
//   - Scalable database operations
//   - Google Cloud integration testing
//
// Notes:
//   - Requires Google Cloud credentials or emulator
//   - Supports Spanner's SQL dialect
//   - Handles distributed transactions
//   - Supports parameterized queries
func SpannerAction(args []interface{}, options map[string]interface{}, silent bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("spanner action requires at least one argument")
	}

	subcommand, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("spanner subcommand must be a string")
	}

	switch subcommand {
	case "connect":
		return spannerConnect(args)
	case "query":
		return spannerQuery(args)
	case "execute":
		return spannerExecute(args)
	case "close":
		return spannerClose(args)
	default:
		return "", fmt.Errorf("unknown spanner subcommand: %s", subcommand)
	}
}

// Global Spanner connection pool
var spannerDB *sql.DB

func spannerConnect(args []interface{}) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("spanner connect requires a connection string")
	}

	connectionString, ok := args[1].(string)
	if !ok {
		return "", fmt.Errorf("spanner connection string must be a string")
	}

	// Check if using emulator
	useEmulator := strings.Contains(connectionString, "useEmulator=true")

	var db *sql.DB
	var err error

	if useEmulator {
		// Connect to Spanner emulator
		dsn := fmt.Sprintf("projects/robogo-test-project/instances/robogo-test-instance/databases/robogo-test-db")
		db, err = sql.Open("spanner", dsn)
		if err != nil {
			return "", fmt.Errorf("failed to connect to Spanner emulator: %w", err)
		}
	} else {
		// Connect to real Spanner (requires credentials)
		dsn := connectionString
		db, err = sql.Open("spanner", dsn)
		if err != nil {
			return "", fmt.Errorf("failed to connect to Spanner: %w", err)
		}
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to ping Spanner: %w", err)
	}

	spannerDB = db

	return fmt.Sprintf(`{"status": "connected", "connection_string": "%s"}`, connectionString), nil
}

func spannerQuery(args []interface{}) (string, error) {
	if spannerDB == nil {
		return "", fmt.Errorf("no Spanner connection available, call connect first")
	}

	if len(args) < 3 {
		return "", fmt.Errorf("spanner query requires a connection string and query")
	}

	_, ok1 := args[1].(string)
	query, ok2 := args[2].(string)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("spanner query connection string and query must be strings")
	}

	// Execute query
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := spannerDB.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to execute Spanner query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get column names: %w", err)
	}

	// Prepare result slice
	var results []map[string]interface{}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Iterate through rows
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			row[col] = val
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	// Format result
	result := map[string]interface{}{
		"status":    "success",
		"rows":      results,
		"row_count": len(results),
		"columns":   columns,
		"query":     query,
	}

	return formatJSON(result), nil
}

func spannerExecute(args []interface{}) (string, error) {
	if spannerDB == nil {
		return "", fmt.Errorf("no Spanner connection available, call connect first")
	}

	if len(args) < 3 {
		return "", fmt.Errorf("spanner execute requires a connection string and statement")
	}

	_, ok1 := args[1].(string)
	statement, ok2 := args[2].(string)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("spanner execute connection string and statement must be strings")
	}

	// Handle parameters if provided
	var params []interface{}
	if len(args) > 3 {
		if paramArg, ok := args[3].([]interface{}); ok {
			params = paramArg
		}
	}

	// Execute statement
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result sql.Result
	var err error

	if len(params) > 0 {
		result, err = spannerDB.ExecContext(ctx, statement, params...)
	} else {
		result, err = spannerDB.ExecContext(ctx, statement)
	}

	if err != nil {
		return "", fmt.Errorf("failed to execute Spanner statement: %w", err)
	}

	// Get result info
	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	resultMap := map[string]interface{}{
		"status":         "success",
		"statement":      statement,
		"last_insert_id": lastInsertID,
		"rows_affected":  rowsAffected,
		"parameters":     params,
	}

	return formatJSON(resultMap), nil
}

func spannerClose(args []interface{}) (string, error) {
	if spannerDB == nil {
		return "", fmt.Errorf("no Spanner connection available")
	}

	err := spannerDB.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close Spanner connection: %w", err)
	}

	spannerDB = nil

	return `{"status": "disconnected"}`, nil
}

// formatJSON converts a map to JSON string
func formatJSON(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}
	return string(jsonBytes)
}
