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
		return types.MissingArgsError("spanner", 3, len(args))
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	dbPath := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultDatabaseTimeout)
	defer cancel()

	db, err := sql.Open("spanner", dbPath)
	if err != nil {
		log.Printf("[spanner/sql] failed to open database: %v", err)
		return types.DatabaseConnectionError("Cloud Spanner", err.Error())
	}
	defer db.Close()

	switch operation {
	case constants.OperationQuery, constants.OperationSelect:
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("[spanner/sql] query failed: %v", err)
			return types.DatabaseQueryError("Cloud Spanner", err.Error())
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return types.DatabaseQueryError("Cloud Spanner", err.Error())
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
				return types.DatabaseQueryError("Cloud Spanner", err.Error())
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
			return types.DatabaseQueryError("Cloud Spanner", err.Error())
		}

		// Create the initial result structure
		resultData := map[string]any{
			"columns": columns,
			"rows":    results,
		}
		
		// Marshal and unmarshal to ensure JSON compatibility for jq
		jsonBytes, err := json.Marshal(resultData)
		if err != nil {
			return types.DatabaseQueryError("Cloud Spanner", fmt.Sprintf("JSON marshal error: %v", err))
		}
		
		var jsonCompatibleResult map[string]any
		if err := json.Unmarshal(jsonBytes, &jsonCompatibleResult); err != nil {
			return types.DatabaseQueryError("Cloud Spanner", fmt.Sprintf("JSON unmarshal error: %v", err))
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

	case constants.OperationInsert, constants.OperationUpdate, constants.OperationDelete, constants.OperationExecute:
		res, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Printf("[spanner/sql] DML failed: %v", err)
			return types.DatabaseExecuteError("Cloud Spanner", err.Error())
		}
		affected, _ := res.RowsAffected()
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   map[string]any{"rows_affected": affected},
		}

	default:
		return types.UnknownOperationError("spanner", operation)
	}
}
