package actions

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/JianLoong/robogo/internal/util"
	_ "github.com/lib/pq"
)

// PostgreSQLManager manages PostgreSQL database connections
type PostgreSQLManager struct {
	connections map[string]*sql.DB
	mutex       sync.RWMutex
}

// QueryResult represents the result of a PostgreSQL query
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

// SimpleRow represents a single PostgreSQL row with column names
type SimpleRow struct {
	Data map[string]interface{} `json:"data"`
}

// SimpleResult represents a simplified PostgreSQL query result
type SimpleResult struct {
	Rows  []SimpleRow `json:"rows"`
	Count int         `json:"count"`
}

// BatchQueryResult represents the result of a batch PostgreSQL operation
type BatchQueryResult struct {
	ConnectionString string                 `json:"connection_string"`
	Query            string                 `json:"query"`
	Result           *QueryResult           `json:"result,omitempty"`
	Error            string                 `json:"error,omitempty"`
	Index            int                    `json:"index"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}


// PostgresAction performs PostgreSQL database operations with comprehensive support for queries, transactions, and connection management.
//
// Now accepts a context.Context parameter for resource cleanup and timeouts.
func PostgresAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewArgumentCountError("postgres", 2, len(args))
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	// Set timeout from options or default (30s)
	timeout := 30 * time.Second
	if optTimeout, ok := options["timeout"]; ok {
		if t, ok := optTimeout.(time.Duration); ok {
			timeout = t
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch operation {
	case "query", "select":
		return executeQuery(ctx, connectionString, args[2:], silent)
	case "execute", "insert", "update", "delete":
		return executeStatement(ctx, connectionString, args[2:], silent)
	case "connect":
		return testConnection(ctx, connectionString)
	case "close":
		return closeConnection(ctx, connectionString)
	case "batch":
		return executeBatchOperations(ctx, connectionString, args[2:], silent)
	default:
		return nil, util.NewArgumentValueError("postgres", 0, operation, "unknown postgres operation")
	}
}

// executeQuery executes a SELECT query and returns results
func executeQuery(ctx context.Context, connectionString string, args []interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewArgumentCountError("postgres", 1, len(args))
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
	actionCtx := GetActionContext(ctx)
	db, err := getConnectionWithManager(connectionString, actionCtx.PostgresManager)
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to get database connection").
			WithAction("postgres").
			WithCause(err).
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			}).
			Build()
	}

	startTime := time.Now()

	// Execute query with context
	rows, err := db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "query execution failed").
			WithAction("postgres").
			WithCause(err).
			Build().
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
	}
	defer rows.Close()

	duration := time.Since(startTime)

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to get column names").
			WithAction("postgres").
			WithCause(err).
			Build().
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
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
			return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to scan row").
			WithAction("postgres").
			WithCause(err).
			Build().
				WithDetails(map[string]interface{}{
					"query":             query,
					"connection_string": connectionString,
					"column_count":      len(columns),
				})
		}

		// Use the actual values without string conversion
		resultRows = append(resultRows, values)
	}

	if err := rows.Err(); err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "error iterating rows").
			WithAction("postgres").
			WithCause(err).
			Build().
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
				"rows_processed":    len(resultRows),
			})
	}

	// Transform to consistent format
	transformedResult := transformToConsistentFormat(columns, resultRows, query)

	// Add rich metadata
	result := map[string]interface{}{
		"query":    query,
		"columns":  transformedResult.(map[string]interface{})["columns"],
		"rows":     transformedResult.(map[string]interface{})["rows"],
		"duration": duration,
		"metadata": map[string]interface{}{
			"row_count": len(resultRows),
			"params":    queryArgs,
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
	resultMap, err := util.ConvertToMap(result)
	if err != nil {
		return nil, util.NewDatabaseError("failed to convert query result to map", err, "postgres").
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
				"row_count":         len(resultRows),
			})
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("🗄️  Query executed: %d rows returned in %v\n", len(resultRows), duration)
	}

	return resultMap, nil
}

// executeBatchOperations executes multiple database operations in parallel
//
// Parameters:
//   - connectionString: Database connection string
//   - args: Array of operations to execute in parallel
//   - silent: Whether to suppress output
//
// Returns: JSON array of results for all operations
//
// Examples:
//   - Batch queries: ["batch", [{"query": "SELECT * FROM users"}, {"query": "SELECT * FROM orders"}]]
//   - Mixed operations: ["batch", [{"operation": "query", "query": "SELECT COUNT(*) FROM users"}, {"operation": "execute", "query": "INSERT INTO logs (message) VALUES ($1)", "params": ["test"]}]]
//   - With concurrency: ["batch", [{"query": "SELECT * FROM table1"}, {"query": "SELECT * FROM table2"}], {"concurrency": 5}]
//
// Use Cases:
//   - Parallel data validation
//   - Batch data setup and teardown
//   - Performance testing with multiple queries
//   - Concurrent data operations
func executeBatchOperations(ctx context.Context, connectionString string, args []interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewValidationError("batch operation requires at least one operation to execute",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 1,
			}).WithAction("postgres")
	}

	// Parse operations
	var operations []map[string]interface{}
	var maxConcurrency int = 10 // Default concurrency limit

	// Extract operations and options
	for _, arg := range args {
		switch v := arg.(type) {
		case []interface{}:
			// This is the operations array
			for _, op := range v {
				if opMap, ok := op.(map[string]interface{}); ok {
					operations = append(operations, opMap)
				} else {
					return nil, util.NewValidationError("invalid operation format: expected map",
						map[string]interface{}{
							"operation_type": fmt.Sprintf("%T", op),
							"expected_type":  "map[string]interface{}",
						}).WithAction("postgres")
				}
			}
		case map[string]interface{}:
			// Check if this is options or a single operation
			if concurrency, ok := v["concurrency"]; ok {
				if concurrencyInt, ok := concurrency.(int); ok {
					maxConcurrency = concurrencyInt
				}
			} else {
				// Single operation
				operations = append(operations, v)
			}
		default:
			return nil, util.NewValidationError("invalid argument type: expected array or map",
				map[string]interface{}{
					"argument_type":  fmt.Sprintf("%T", arg),
					"expected_types": []string{"[]interface{}", "map[string]interface{}"},
				}).WithAction("postgres")
		}
	}

	if len(operations) == 0 {
		return nil, util.NewValidationError("no operations provided for batch execution",
			map[string]interface{}{
				"operations_count": len(operations),
				"required_count":   1,
			}).WithAction("postgres")
	}

	// Execute operations in parallel
	results := make([]BatchQueryResult, len(operations))
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	for i, operation := range operations {
		wg.Add(1)
		go func(index int, op map[string]interface{}) {
			defer wg.Done()
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			// Check for context cancellation
			select {
			case <-ctx.Done():
				results[index] = BatchQueryResult{
					ConnectionString: connectionString,
					Index:            index,
					Error:            "operation cancelled by context",
				}
				return
			default:
			}
			// Execute the operation
			result := executeSingleBatchOperation(ctx, connectionString, op, index)
			results[index] = result
		}(i, operation)
	}

	wg.Wait()

	// Convert results to JSON
	resultMap, err := util.ConvertToMap(results)
	if err != nil {
		return nil, util.NewDatabaseError("failed to convert batch results to map", err, "postgres").
			WithDetails(map[string]interface{}{
				"operations_count":  len(operations),
				"connection_string": connectionString,
			})
	}

	if !silent {
		fmt.Printf("🗄️  Batch database operations completed: %d operations, %d concurrent\n", len(operations), maxConcurrency)
	}

	return resultMap, nil
}

// executeSingleBatchOperation executes a single operation within a batch
func executeSingleBatchOperation(ctx context.Context, connectionString string, operation map[string]interface{}, index int) BatchQueryResult {
	result := BatchQueryResult{
		ConnectionString: connectionString,
		Index:            index,
		Metadata:         make(map[string]interface{}),
	}

	// Extract operation details
	opType, ok := operation["operation"].(string)
	if !ok {
		// Default to query if operation type not specified
		opType = "query"
	}

	query, ok := operation["query"].(string)
	if !ok {
		result.Error = "missing query in operation"
		return result
	}

	result.Query = query

	// Extract parameters if provided
	var params []interface{}
	if paramsInterface, ok := operation["params"]; ok {
		if paramsArray, ok := paramsInterface.([]interface{}); ok {
			params = paramsArray
		}
	}

	// Execute based on operation type
	switch strings.ToLower(opType) {
	case "query", "select":
		queryResult, err := executeQueryInternal(ctx, connectionString, query, params)
		if err != nil {
			result.Error = util.FormatRobogoError(err)
		} else {
			result.Result = queryResult
		}
	case "execute", "insert", "update", "delete":
		queryResult, err := executeStatementInternal(ctx, connectionString, query, params)
		if err != nil {
			result.Error = util.FormatRobogoError(err)
		} else {
			result.Result = queryResult
		}
	default:
		result.Error = fmt.Sprintf("unknown operation type: %s", opType)
	}

	return result
}

// executeQueryInternal is an internal version of executeQuery that returns QueryResult
func executeQueryInternal(ctx context.Context, connectionString, query string, params []interface{}) (*QueryResult, error) {
	// Get or create database connection
	actionCtx := GetActionContext(ctx)
	db, err := getConnectionWithManager(connectionString, actionCtx.PostgresManager)
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to get database connection").
			WithAction("postgres").
			WithCause(err).
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			}).
			Build()
	}

	startTime := time.Now()

	// Execute query with context
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "query execution failed").
			WithAction("postgres").
			WithCause(err).
			Build().
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
	}
	defer rows.Close()

	duration := time.Since(startTime)

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to get column names").
			WithAction("postgres").
			WithCause(err).
			Build().
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
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
			return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to scan row").
			WithAction("postgres").
			WithCause(err).
			Build().
				WithDetails(map[string]interface{}{
					"query":             query,
					"connection_string": connectionString,
					"column_count":      len(columns),
				})
		}

		// Use the actual values without string conversion
		resultRows = append(resultRows, values)
	}

	if err := rows.Err(); err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "error iterating rows").
			WithAction("postgres").
			WithCause(err).
			Build().
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
				"rows_processed":    len(resultRows),
			})
	}

	// Create result object (keeping original format for internal use)
	result := &QueryResult{
		Query:    query,
		Columns:  columns,
		Rows:     resultRows,
		Duration: duration,
		Metadata: map[string]interface{}{
			"row_count": len(resultRows),
		},
	}

	return result, nil
}

// executeStatementInternal is an internal version of executeStatement that returns QueryResult
func executeStatementInternal(ctx context.Context, connectionString, query string, params []interface{}) (*QueryResult, error) {
	// Get or create database connection
	actionCtx := GetActionContext(ctx)
	db, err := getConnectionWithManager(connectionString, actionCtx.PostgresManager)
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to get database connection").
			WithAction("postgres").
			WithCause(err).
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			}).
			Build()
	}

	startTime := time.Now()

	// Execute statement with context
	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, util.NewDatabaseError("statement execution failed", err, "postgres").
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
	}

	duration := time.Since(startTime)

	// Get affected rows and last insert ID
	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	// Create result object
	queryResult := &QueryResult{
		Query:        query,
		RowsAffected: rowsAffected,
		LastInsertID: lastInsertID,
		Duration:     duration,
		Metadata:     make(map[string]interface{}),
	}

	return queryResult, nil
}

// executeStatement executes INSERT, UPDATE, DELETE statements
func executeStatement(ctx context.Context, connectionString string, args []interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewValidationError("execute operation requires a SQL statement",
			map[string]interface{}{
				"provided_args": len(args),
				"required_args": 1,
			}).WithAction("postgres")
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
	actionCtx := GetActionContext(ctx)
	db, err := getConnectionWithManager(connectionString, actionCtx.PostgresManager)
	if err != nil {
		return nil, util.NewErrorBuilder(util.ErrorTypeDatabase, "failed to get database connection").
			WithAction("postgres").
			WithCause(err).
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			}).
			Build()
	}

	startTime := time.Now()

	// Execute statement with context
	result, err := db.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, util.NewDatabaseError("statement execution failed", err, "postgres").
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
	}

	duration := time.Since(startTime)

	// Get affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, util.NewDatabaseError("failed to get rows affected", err, "postgres").
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
			})
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
	resultMap, err := util.ConvertToMap(dbResult)
	if err != nil {
		return nil, util.NewDatabaseError("failed to convert statement result to map", err, "postgres").
			WithDetails(map[string]interface{}{
				"query":             query,
				"connection_string": connectionString,
				"rows_affected":     rowsAffected,
			})
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("🗄️  Statement executed: %d rows affected in %v\n", rowsAffected, duration)
	}

	return resultMap, nil
}

// testConnection tests a database connection
func testConnection(ctx context.Context, connectionString string) (interface{}, error) {
	actionCtx := GetActionContext(ctx)
	db, err := getConnectionWithManager(connectionString, actionCtx.PostgresManager)
	if err != nil {
		return nil, util.NewDatabaseError("connection test failed", err, "postgres").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			})
	}

	// Test the connection with a simple query
	startTime := time.Now()
	err = db.PingContext(ctx)
	duration := time.Since(startTime)

	if err != nil {
		return nil, util.NewDatabaseError("connection ping failed", err, "postgres").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			})
	}

	result := map[string]interface{}{
		"status":   "connected",
		"duration": duration.String(),
		"message":  "Database connection successful",
	}

	resultMap, err := util.ConvertToMap(result)
	if err != nil {
		return nil, util.NewDatabaseError("failed to convert test result to map", err, "postgres").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			})
	}

	fmt.Printf("🗄️  Connection test successful in %v\n", duration)
	return resultMap, nil
}

// closeConnection closes a PostgreSQL connection
func closeConnection(ctx context.Context, connectionString string) (interface{}, error) {
	actionCtx := GetActionContext(ctx)
	actionCtx.PostgresManager.mutex.Lock()
	defer actionCtx.PostgresManager.mutex.Unlock()

	if db, exists := actionCtx.PostgresManager.connections[connectionString]; exists {
		err := db.Close()
		delete(actionCtx.PostgresManager.connections, connectionString)
		if err != nil {
			return nil, util.NewDatabaseError("failed to close PostgreSQL connection", err, "postgres").
				WithDetails(map[string]interface{}{
					"connection_string": connectionString,
				})
		}

		result := map[string]interface{}{
			"status":  "closed",
			"message": "PostgreSQL connection closed successfully",
		}

		resultMap, err := util.ConvertToMap(result)
		if err != nil {
			return nil, util.NewDatabaseError("failed to convert close result to map", err, "postgres").
				WithDetails(map[string]interface{}{
					"connection_string": connectionString,
				})
		}

		fmt.Printf("🗄️  PostgreSQL connection closed successfully\n")
		return resultMap, nil
	}

	return nil, util.NewDatabaseError("no PostgreSQL connection found for the given connection string", nil, "postgres").
		WithDetails(map[string]interface{}{
			"connection_string": connectionString,
		})
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
		fmt.Printf("Warning: Connection string format may be malformed: %s\n", connectionString)
		return connectionString, nil
	}

	// Split userinfo into username and password
	userInfo := strings.SplitN(parts[0], ":", 2)
	if len(userInfo) != 2 {
		// Handle cases where there might be no password or malformed userinfo
		fmt.Printf("Warning: User info format may be malformed: %s\n", parts[0])
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


// getConnectionWithManager gets or creates a PostgreSQL connection using a specific manager
func getConnectionWithManager(connectionString string, manager *PostgreSQLManager) (*sql.DB, error) {
	manager.mutex.RLock()
	if db, exists := manager.connections[connectionString]; exists {
		manager.mutex.RUnlock()
		return db, nil
	}
	manager.mutex.RUnlock()

	// Create new connection
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	// Double-check after acquiring write lock
	if db, exists := manager.connections[connectionString]; exists {
		return db, nil
	}

	// Encode the connection string to handle special characters
	encodedConnectionString, err := encodeConnectionString(connectionString)
	if err != nil {
		return nil, util.NewDatabaseError("failed to encode PostgreSQL connection string", err, "postgres").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			})
	}

	db, err := sql.Open("postgres", encodedConnectionString)
	if err != nil {
		return nil, util.NewDatabaseError("failed to open PostgreSQL connection", err, "postgres").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			})
	}

	// Configure PostgreSQL-specific connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, util.NewDatabaseError("failed to ping PostgreSQL database", err, "postgres").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
			})
	}

	manager.connections[connectionString] = db
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


// CloseAll closes all PostgreSQL connections in the manager
func (pgm *PostgreSQLManager) CloseAll() error {
	pgm.mutex.Lock()
	defer pgm.mutex.Unlock()

	var lastErr error
	for connectionString, db := range pgm.connections {
		if err := db.Close(); err != nil {
			lastErr = err
		}
		delete(pgm.connections, connectionString)
	}
	return lastErr
}

// transformToConsistentFormat transforms database query results to a consistent format
// Returns a consistent structure with columns, rows, and value fields
func transformToConsistentFormat(columns []string, rows [][]interface{}, query string) interface{} {
	result := map[string]interface{}{
		"columns": columns,
		"rows":    rows,
	}

	// Add shortcut for single value or values array
	if len(rows) == 1 && len(columns) == 1 {
		// Single row, single column
		result["value"] = rows[0][0]
	} else if len(rows) > 1 && len(columns) == 1 {
		// Multiple rows, single column
		var values []interface{}
		for _, row := range rows {
			values = append(values, row[0])
		}
		result["values"] = values
	}

	return result
}

// isCountQuery checks if a query is a COUNT query
func isCountQuery(query string) bool {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	return strings.Contains(queryUpper, "SELECT COUNT(") ||
		(strings.HasPrefix(queryUpper, "SELECT COUNT(*)") && !strings.Contains(queryUpper, "FROM"))
}
