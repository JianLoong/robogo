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
		return types.NewErrorResult("postgres action requires at least 3 arguments: operation, connection_string, query")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	// Open connection for this operation only
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return types.NewErrorResult("failed to open postgres connection: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(1 * time.Second)

	if err = db.Ping(); err != nil {
		return types.NewErrorResult("failed to ping postgres database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "query", "select":
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return types.NewErrorResult("failed to execute query: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return types.NewErrorResult("failed to get columns: %v", err)
		}

		var results [][]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return types.NewErrorResult("failed to scan row: %v", err)
			}
			results = append(results, values)
		}

		result := map[string]interface{}{
			"columns": columns,
			"rows":    results,
		}
		if asJSON, ok := options["as_json"].(bool); ok && asJSON {
			jsonBytes, err := json.Marshal(result)
			if err == nil {
				return types.ActionResult{
					Status: types.ActionStatusPassed,
					Data:   map[string]interface{}{"json_string": string(jsonBytes)},
					Output: string(jsonBytes),
				}, nil
			}
			// If marshaling fails, fall through to structured result
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   result,
		}, nil

	case "execute", "insert", "update", "delete":
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			return types.NewErrorResult("failed to execute statement: %v", err)
		}
		rowsAffected, _ := result.RowsAffected()
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]interface{}{"rows_affected": rowsAffected},
		}, nil

	default:
		return types.NewErrorResult("unknown postgres operation: %s", operation)
	}
}
