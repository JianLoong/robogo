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
		return types.MissingArgsError("postgres", 3, len(args))
	}

	// Check for unresolved variables in critical arguments
	if errorResult := validateArgsResolved("postgres", args[:3]); errorResult != nil {
		return *errorResult
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	// Open connection for this operation only
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return types.DatabaseConnectionError("PostgreSQL", err.Error())
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(constants.DefaultConnectionLifetime)

	if err = db.Ping(); err != nil {
		return types.DatabaseConnectionError("PostgreSQL", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultDatabaseTimeout)
	defer cancel()

	switch operation {
	case constants.OperationQuery, constants.OperationSelect:
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return types.DatabaseQueryError("PostgreSQL", err.Error())
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return types.DatabaseQueryError("PostgreSQL", err.Error())
		}

		var results [][]any
		for rows.Next() {
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return types.DatabaseQueryError("PostgreSQL", err.Error())
			}
			
			// Convert each value to a JSON-compatible type
			jsonValues := make([]any, len(values))
			for i, val := range values {
				if val == nil {
					jsonValues[i] = nil
				} else if bytes, ok := val.([]byte); ok {
					jsonValues[i] = string(bytes)
				} else {
					jsonValues[i] = val
				}
			}
			results = append(results, jsonValues)
		}

		// Create the initial result structure
		resultData := map[string]any{
			"columns": columns,
			"rows":    results,
		}
		
		// Marshal and unmarshal to ensure JSON compatibility for jq
		jsonBytes, err := json.Marshal(resultData)
		if err != nil {
			return types.DatabaseQueryError("PostgreSQL", fmt.Sprintf("JSON marshal error: %v", err))
		}
		
		var jsonCompatibleResult map[string]any
		if err := json.Unmarshal(jsonBytes, &jsonCompatibleResult); err != nil {
			return types.DatabaseQueryError("PostgreSQL", fmt.Sprintf("JSON unmarshal error: %v", err))
		}
		
		result := jsonCompatibleResult
		if asJSON, ok := options["as_json"].(bool); ok && asJSON {
			jsonBytes, err := json.Marshal(result)
			if err == nil {
				return types.ActionResult{
					Status: constants.ActionStatusPassed,
					Data:   map[string]any{"json_string": string(jsonBytes)},
				}
			}
			// If marshaling fails, fall through to structured result
		}
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   result,
		}

	case constants.OperationExecute, constants.OperationInsert, constants.OperationUpdate, constants.OperationDelete:
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			return types.DatabaseExecuteError("PostgreSQL", err.Error())
		}
		rowsAffected, _ := result.RowsAffected()
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   map[string]any{"rows_affected": rowsAffected},
		}

	default:
		return types.UnknownOperationError("postgres", operation)
	}
}
