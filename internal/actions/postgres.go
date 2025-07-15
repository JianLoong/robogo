package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/JianLoong/robogo/internal/common"
)

// PostgreSQL action - simplified implementation with proper resource management
func postgresAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("postgres action requires at least 3 arguments: operation, connection_string, query")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	// Open connection for this operation only
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	// Ensure connection is closed immediately after use
	defer func() {
		db.Close()
	}()
	
	// Set aggressive connection limits to ensure cleanup
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(1 * time.Second)
	
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "query", "select":
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log but don't override original error
				fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
			}
		}()

		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get columns: %w", err)
		}

		var results [][]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			results = append(results, values)
		}

		result := map[string]interface{}{
			"columns": columns,
			"rows":    results,
		}
		
		// Return JSON string directly  
		if jsonBytes, err := json.Marshal(result); err == nil {
			return string(jsonBytes), nil
		}
		
		// Fallback to map if JSON fails
		return result, nil

	case "execute", "insert", "update", "delete":
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("execute failed: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		
		execResult := map[string]interface{}{
			"rows_affected": rowsAffected,
		}
		
		// Return JSON string directly
		if jsonBytes, err := json.Marshal(execResult); err == nil {
			return string(jsonBytes), nil
		}
		
		// Fallback to map if JSON fails
		return execResult, nil

	default:
		return nil, fmt.Errorf("unknown postgres operation: %s", operation)
	}
}