package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
	_ "github.com/lib/pq"
)

// PostgreSQL action - simplified implementation with proper resource management
func postgresAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 3 {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  "postgres action requires at least 3 arguments: operation, connection_string, query",
		}, fmt.Errorf("postgres action requires at least 3 arguments: operation, connection_string, query")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	// Open connection for this operation only
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  fmt.Sprintf("failed to open postgres connection: %v", err),
		}, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(1 * time.Second)

	if err = db.Ping(); err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  fmt.Sprintf("failed to ping postgres database: %v", err),
		}, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "query", "select":
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  fmt.Sprintf("failed to execute query: %v", err),
			}, fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  fmt.Sprintf("failed to get columns: %v", err),
			}, fmt.Errorf("failed to get columns: %w", err)
		}

		var results [][]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return types.ActionResult{
					Status: types.ActionStatusError,
					Error:  fmt.Sprintf("failed to scan row: %v", err),
				}, fmt.Errorf("failed to scan row: %w", err)
			}
			results = append(results, values)
		}

		result := map[string]interface{}{
			"columns": columns,
			"rows":    results,
		}

		if jsonBytes, err := json.Marshal(result); err == nil {
			return types.ActionResult{
				Status: types.ActionStatusSuccess,
				Data:   string(jsonBytes),
			}, nil
		}
		return types.ActionResult{
			Status: types.ActionStatusSuccess,
			Data:   result,
		}, nil

	case "execute", "insert", "update", "delete":
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  fmt.Sprintf("failed to execute statement: %v", err),
			}, fmt.Errorf("failed to execute statement: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		return types.ActionResult{
			Status: types.ActionStatusSuccess,
			Data:   map[string]interface{}{"rows_affected": rowsAffected},
		}, nil

	default:
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  fmt.Sprintf("unknown postgres operation: %s", operation),
		}, fmt.Errorf("unknown postgres operation: %s", operation)
	}
}
