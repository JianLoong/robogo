package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
	_ "github.com/lib/pq"
)

// PostgreSQL action - simplified implementation with proper resource management
func postgresAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 3 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "POSTGRES_MISSING_ARGS").
			WithTemplate("postgres action requires at least 3 arguments: operation, connection_string, query").
			Build()
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	// Open connection for this operation only
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "POSTGRES_CONNECTION_FAILED").
			WithTemplate("failed to open postgres connection: %v").
			WithContext("connection_string", connectionString).
			WithContext("error", err.Error()).
			Build(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(constants.DefaultConnectionLifetime)

	if err = db.Ping(); err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "POSTGRES_PING_FAILED").
			WithTemplate("failed to ping postgres database: %v").
			WithContext("connection_string", connectionString).
			WithContext("error", err.Error()).
			Build(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultDatabaseTimeout)
	defer cancel()

	switch operation {
	case constants.OperationQuery, constants.OperationSelect:
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "POSTGRES_QUERY_FAILED").
				WithTemplate("failed to execute query: %v").
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "POSTGRES_COLUMNS_FAILED").
				WithTemplate("failed to get columns: %v").
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
		}

		var results [][]any
		for rows.Next() {
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return types.NewErrorBuilder(types.ErrorCategoryExecution, "POSTGRES_SCAN_FAILED").
					WithTemplate("failed to scan row: %v").
					WithContext("query", query).
					WithContext("error", err.Error()).
					Build(err)
			}
			results = append(results, values)
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

	case constants.OperationExecute, constants.OperationInsert, constants.OperationUpdate, constants.OperationDelete:
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "POSTGRES_EXECUTE_FAILED").
				WithTemplate("failed to execute statement: %v").
				WithContext("query", query).
				WithContext("error", err.Error()).
				Build(err)
		}
		rowsAffected, _ := result.RowsAffected()
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]any{"rows_affected": rowsAffected},
		}

	default:
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "POSTGRES_UNKNOWN_OPERATION").
			WithTemplate("unknown postgres operation: %s").
			WithContext("operation", operation).
			Build(operation)
	}
}
