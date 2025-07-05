package actions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseManager manages database connections
type DatabaseManager struct {
	connections map[string]*sql.DB
	mutex       sync.RWMutex
}

// QueryResult represents the result of a database query
type QueryResult struct {
	Query        string                 `json:"query"`
	RowsAffected int64                  `json:"rows_affected,omitempty"`
	LastInsertID int64                  `json:"last_insert_id,omitempty"`
	Columns      []string               `json:"columns,omitempty"`
	Rows         [][]interface{}        `json:"rows,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Global database manager instance
var dbManager = &DatabaseManager{
	connections: make(map[string]*sql.DB),
}

// PostgresAction performs PostgreSQL database operations with comprehensive support for queries, transactions, and connection management.
//
// Parameters:
//   - operation: Database operation to perform (query, execute, connect, close, test)
//   - connection: Database connection string or connection parameters
//   - query: SQL query or statement to execute
//   - params: Query parameters (optional, for parameterized queries)
//   - options: Additional options (timeout, ssl_mode, etc.)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON result with operation status, data, and timing information
//
// Supported Operations:
//   - "connect": Establish database connection
//   - "query": Execute SELECT query and return results
//   - "execute": Execute INSERT/UPDATE/DELETE statement
//   - "close": Close database connection
//   - "test": Test connection without executing queries
//
// Examples:
//   - Connect: ["connect", "postgres://user:pass@localhost:5432/dbname"]
//   - Query: ["query", "SELECT * FROM users WHERE id = $1", [123]]
//   - Execute: ["execute", "INSERT INTO users (name, email) VALUES ($1, $2)", ["John", "john@example.com"]]
//   - Close: ["close"]
//
// Use Cases:
//   - Database testing and validation
//   - Data setup and teardown
//   - Integration testing with databases
//   - Performance testing of database operations
//   - Data verification and assertions
//
// Notes:
//   - Supports parameterized queries for security
//   - Automatic connection pooling and management
//   - Comprehensive error handling and timeout support
//   - Results available for assertions and variable storage
func PostgresAction(args []interface{}, silent bool) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("postgres action requires at least 2 arguments: operation and connection_string")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	switch operation {
	case "query", "select":
		return executeQuery(connectionString, args[2:], silent)
	case "execute", "insert", "update", "delete":
		return executeStatement(connectionString, args[2:], silent)
	case "connect":
		return testConnection(connectionString)
	case "close":
		return closeConnection(connectionString)
	default:
		return "", fmt.Errorf("unknown postgres operation: %s", operation)
	}
}

// executeQuery executes a SELECT query and returns results
func executeQuery(connectionString string, args []interface{}, silent bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("query operation requires a SQL query")
	}

	query := fmt.Sprintf("%v", args[0])
	var queryArgs []interface{}

	// Extract query parameters if provided
	if len(args) > 1 {
		if params, ok := args[1].([]interface{}); ok {
			queryArgs = params
		} else {
			queryArgs = args[1:]
		}
	}

	// Get or create database connection
	db, err := getConnection(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	startTime := time.Now()

	// Execute query
	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	duration := time.Since(startTime)

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get column names: %w", err)
	}

	// Prepare result container
	var resultRows [][]interface{}

	// Process rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the values slice
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert values to strings for JSON serialization
		rowData := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				rowData[i] = nil
			} else {
				rowData[i] = fmt.Sprintf("%v", val)
			}
		}

		resultRows = append(resultRows, rowData)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	// Create result object
	result := QueryResult{
		Query:    query,
		Columns:  columns,
		Rows:     resultRows,
		Duration: duration,
		Metadata: map[string]interface{}{
			"row_count": len(resultRows),
		},
	}

	// Convert to JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("🗄️  Query executed: %d rows returned in %v\n", len(resultRows), duration)
	}

	return string(jsonResult), nil
}

// executeStatement executes INSERT, UPDATE, DELETE statements
func executeStatement(connectionString string, args []interface{}, silent bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("execute operation requires a SQL statement")
	}

	query := fmt.Sprintf("%v", args[0])
	var queryArgs []interface{}

	// Extract query parameters if provided
	if len(args) > 1 {
		if params, ok := args[1].([]interface{}); ok {
			queryArgs = params
		} else {
			queryArgs = args[1:]
		}
	}

	// Get or create database connection
	db, err := getConnection(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	startTime := time.Now()

	// Execute statement
	result, err := db.Exec(query, queryArgs...)
	if err != nil {
		return "", fmt.Errorf("statement execution failed: %w", err)
	}

	duration := time.Since(startTime)

	// Get affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Get last insert ID (if applicable)
	var lastInsertID int64
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		lastInsertID, err = result.LastInsertId()
		if err != nil {
			// PostgreSQL doesn't support LastInsertId, so we'll ignore this error
			lastInsertID = 0
		}
	}

	// Create result object
	dbResult := QueryResult{
		Query:        query,
		RowsAffected: rowsAffected,
		LastInsertID: lastInsertID,
		Duration:     duration,
		Metadata: map[string]interface{}{
			"operation": getOperationType(query),
		},
	}

	// Convert to JSON
	jsonResult, err := json.Marshal(dbResult)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("🗄️  Statement executed: %d rows affected in %v\n", rowsAffected, duration)
	}

	return string(jsonResult), nil
}

// testConnection tests a database connection
func testConnection(connectionString string) (string, error) {
	db, err := getConnection(connectionString)
	if err != nil {
		return "", fmt.Errorf("connection test failed: %w", err)
	}

	// Test the connection with a simple query
	startTime := time.Now()
	err = db.Ping()
	duration := time.Since(startTime)

	if err != nil {
		return "", fmt.Errorf("connection ping failed: %w", err)
	}

	result := map[string]interface{}{
		"status":   "connected",
		"duration": duration.String(),
		"message":  "Database connection successful",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	fmt.Printf("🗄️  Connection test successful in %v\n", duration)
	return string(jsonResult), nil
}

// closeConnection closes a database connection
func closeConnection(connectionString string) (string, error) {
	dbManager.mutex.Lock()
	defer dbManager.mutex.Unlock()

	if db, exists := dbManager.connections[connectionString]; exists {
		err := db.Close()
		delete(dbManager.connections, connectionString)
		if err != nil {
			return "", fmt.Errorf("failed to close connection: %w", err)
		}

		result := map[string]interface{}{
			"status":  "closed",
			"message": "Database connection closed successfully",
		}

		jsonResult, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
		}

		fmt.Printf("🗄️  Connection closed successfully\n")
		return string(jsonResult), nil
	}

	return "", fmt.Errorf("no connection found for the given connection string")
}

// encodeConnectionString properly encodes a PostgreSQL connection string
func encodeConnectionString(connectionString string) (string, error) {
	// Parse the connection string to extract components
	if !strings.HasPrefix(connectionString, "postgres://") {
		return connectionString, nil // Not a postgres URL, return as-is
	}

	// Extract the parts after postgres://
	parts := strings.SplitN(connectionString[11:], "@", 2)
	if len(parts) != 2 {
		// Try to handle cases where there might be no @ symbol or malformed string
		fmt.Printf("⚠️  Warning: Connection string format may be malformed: %s\n", connectionString)
		return connectionString, nil
	}

	// Split userinfo into username and password
	userInfo := strings.SplitN(parts[0], ":", 2)
	if len(userInfo) != 2 {
		// Handle cases where there might be no password or malformed userinfo
		fmt.Printf("⚠️  Warning: User info format may be malformed: %s\n", parts[0])
		return connectionString, nil
	}

	username := userInfo[0]
	password := userInfo[1]
	hostPart := parts[1]

	// URL encode the username and password
	encodedUsername := url.QueryEscape(username)
	encodedPassword := url.QueryEscape(password)

	// Reconstruct the connection string
	encodedConnectionString := fmt.Sprintf("postgres://%s:%s@%s", encodedUsername, encodedPassword, hostPart)

	return encodedConnectionString, nil
}

// getConnection gets or creates a database connection
func getConnection(connectionString string) (*sql.DB, error) {
	dbManager.mutex.RLock()
	if db, exists := dbManager.connections[connectionString]; exists {
		dbManager.mutex.RUnlock()
		return db, nil
	}
	dbManager.mutex.RUnlock()

	// Create new connection
	dbManager.mutex.Lock()
	defer dbManager.mutex.Unlock()

	// Double-check after acquiring write lock
	if db, exists := dbManager.connections[connectionString]; exists {
		return db, nil
	}

	// Encode the connection string to handle special characters
	encodedConnectionString, err := encodeConnectionString(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to encode connection string: %w", err)
	}

	db, err := sql.Open("postgres", encodedConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbManager.connections[connectionString] = db
	return db, nil
}

// getOperationType determines the type of SQL operation
func getOperationType(query string) string {
	query = strings.ToUpper(strings.TrimSpace(query))
	if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	}
	return "UNKNOWN"
}

// CloseAllConnections closes all database connections
func CloseAllConnections() {
	dbManager.mutex.Lock()
	defer dbManager.mutex.Unlock()

	for connectionString, db := range dbManager.connections {
		db.Close()
		delete(dbManager.connections, connectionString)
	}
}
