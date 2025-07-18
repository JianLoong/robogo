package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
	_ "github.com/googleapis/go-sql-spanner"
)

func spannerAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 3 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "SPANNER_MISSING_ARGS").
			WithTemplate("spanner action requires at least 3 arguments: operation, database_path, query").
			Build()
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	dbPath := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultDatabaseTimeout)
	defer cancel()

	db, err := sql.Open("spanner", dbPath)
	if err != nil {
		log.Printf("[spanner/sql] failed to open database: %v", err)
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "SPANNER_CONNECTION_FAILED").
			WithTemplate("failed to open spanner database: %v").
			WithContext("database_path", dbPath).
			WithContext("error", err.Error()).
			Build(err)
	}
	defer db.Close()

	switch operation {
	case constants.OperationQuery, constants.OperationSelect:
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("[spanner/sql] query failed: %v", err)
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "SPANNER_QUERY_FAILED").
				WithTemplate("query failed: %v").
				WithContext("database_path", dbPath).
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "SPANNER_COLUMNS_FAILED").
				WithTemplate("failed to get columns: %v").
				WithContext("database_path", dbPath).
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
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
				return types.NewErrorBuilder(types.ErrorCategoryExecution, "SPANNER_SCAN_FAILED").
					WithTemplate("failed to scan row: %v").
					WithContext("database_path", dbPath).
					WithContext("query", query).
					WithContext("error", err.Error()).
					Build(err)
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
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "SPANNER_ROW_ITERATION_FAILED").
				WithTemplate("row iteration error: %v").
				WithContext("database_path", dbPath).
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
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
				}
			}
			// If marshaling fails, fall through to structured result
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   result,
		}

	case constants.OperationInsert, constants.OperationUpdate, constants.OperationDelete, constants.OperationExecute:
		res, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Printf("[spanner/sql] DML failed: %v", err)
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "SPANNER_DML_FAILED").
				WithTemplate("DML failed: %v").
				WithContext("database_path", dbPath).
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
		}
		affected, _ := res.RowsAffected()
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]any{"rows_affected": affected},
		}

	default:
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "SPANNER_UNKNOWN_OPERATION").
			WithTemplate("unsupported spanner operation: %s").
			WithContext("operation", operation).
			Build(operation)
	}
}
