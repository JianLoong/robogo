package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
	_ "github.com/googleapis/go-sql-spanner"
)

func spannerAction(args []any, options map[string]any, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 3 {
		result, _ := types.NewErrorResult("spanner action requires at least 3 arguments: operation, database_path, query")
		return result, nil
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	dbPath := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("spanner", dbPath)
	if err != nil {
		log.Printf("[spanner/sql] failed to open database: %v", err)
		result, _ := types.NewErrorResult("failed to open spanner database: %v", err)
		return result, nil
	}
	defer db.Close()

	switch operation {
	case "query", "select":
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("[spanner/sql] query failed: %v", err)
			result, _ := types.NewErrorResult("query failed: %v", err)
			return result, nil
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			result, _ := types.NewErrorResult("failed to get columns: %v", err)
			return result, nil
		}

		var results [][]any
		rowCount := 0
		for rows.Next() {
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				log.Printf("[spanner/sql] failed to scan row: %v", err)
				result, _ := types.NewErrorResult("failed to scan row: %v", err)
				return result, nil
			}
			for i, v := range values {
				if b, ok := v.([]byte); ok {
					values[i] = string(b)
				}
			}
			results = append(results, values)
			rowCount++
		}
		if err := rows.Err(); err != nil {
			result, _ := types.NewErrorResult("row iteration error: %v", err)
			return result, nil
		}

		result := map[string]any{
			"columns": columns,
			"rows":    results,
		}
		if asJSON, ok := options["as_json"].(bool); ok && asJSON {
			jsonBytes, err := json.Marshal(result)
			if err == nil {
				return types.ActionResult{
					Status: types.ActionStatusPassed,
					Data:   map[string]any{"json_string": string(jsonBytes)},
				}, nil
			}
			// If marshaling fails, fall through to structured result
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   result,
		}, nil

	case "insert", "update", "delete", "execute":
		res, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Printf("[spanner/sql] DML failed: %v", err)
			result, _ := types.NewErrorResult("DML failed: %v", err)
			return result, nil
		}
		affected, _ := res.RowsAffected()
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]any{"rows_affected": affected},
		}, nil

	default:
		result, _ := types.NewErrorResult("unsupported spanner operation: %s", operation)
		return result, nil
	}
}
