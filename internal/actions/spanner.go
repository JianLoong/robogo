package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
	"google.golang.org/api/iterator"
)

// spannerAction executes Spanner queries generically, returning results as JSON-compatible maps.
func spannerAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 3 {
		return types.NewErrorResult("spanner action requires at least 3 arguments: operation, database_path, query")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	databasePath := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := spanner.NewClient(ctx, databasePath)
	if err != nil {
		return types.NewErrorResult("failed to create spanner client: %v", err)
	}
	defer client.Close()

	switch operation {
	case "query", "select":
		iter := client.Single().Query(ctx, spanner.NewStatement(query))
		defer iter.Stop()

		var columns []string
		var results [][]interface{}
		firstRow := true
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return types.NewErrorResult("failed to iterate results: %v", err)
			}

			if firstRow {
				columns = row.ColumnNames()
				firstRow = false
			}

			rowVals := make([]interface{}, len(columns))
			for i := range columns {
				var val interface{}
				if err := trySpannerNullTypes(row, i, &val); err != nil {
					return types.NewErrorResult("failed to decode column %s: %v", columns[i], err)
				}
				rowVals[i] = val
			}
			results = append(results, rowVals)
		}

		result := map[string]interface{}{
			"columns": columns,
			"rows":    results,
		}

		if jsonBytes, err := json.Marshal(result); err == nil {
			return types.ActionResult{
				Status: types.ActionStatusPassed,
				Data:   string(jsonBytes),
			}, nil
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   result,
		}, nil

	case "execute", "insert", "update", "delete":
		_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			_, err := txn.Update(ctx, spanner.NewStatement(query))
			return err
		})
		if err != nil {
			return types.NewErrorResult("failed to execute statement: %v", err)
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]interface{}{"result": "ok"},
		}, nil

	default:
		return types.NewErrorResult("unknown spanner operation: %s", operation)
	}
}

// trySpannerNullTypes attempts to decode a Spanner column into a supported Null type, then assigns the value to *out.
func trySpannerNullTypes(row *spanner.Row, i int, out *interface{}) error {
	// Try NullInt64
	var nInt spanner.NullInt64
	if err := row.Column(i, &nInt); err == nil {
		if nInt.Valid {
			*out = nInt.Int64
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullString
	var nStr spanner.NullString
	if err := row.Column(i, &nStr); err == nil {
		if nStr.Valid {
			*out = nStr.StringVal
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullBool
	var nBool spanner.NullBool
	if err := row.Column(i, &nBool); err == nil {
		if nBool.Valid {
			*out = nBool.Bool
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullFloat64
	var nFloat spanner.NullFloat64
	if err := row.Column(i, &nFloat); err == nil {
		if nFloat.Valid {
			*out = nFloat.Float64
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullTime
	var nTime spanner.NullTime
	if err := row.Column(i, &nTime); err == nil {
		if nTime.Valid {
			*out = nTime.Time
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullDate
	var nDate spanner.NullDate
	if err := row.Column(i, &nDate); err == nil {
		if nDate.Valid {
			*out = nDate.Date
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullNumeric
	var nNum spanner.NullNumeric
	if err := row.Column(i, &nNum); err == nil {
		if nNum.Valid {
			*out = nNum.Numeric.String()
		} else {
			*out = nil
		}
		return nil
	}
	// Try NullJSON
	var nJSON spanner.NullJSON
	if err := row.Column(i, &nJSON); err == nil {
		if nJSON.Valid {
			*out = nJSON.Value
		} else {
			*out = nil
		}
		return nil
	}
	// Try []byte (for BYTES)
	var b []byte
	if err := row.Column(i, &b); err == nil {
		*out = b
		return nil
	}
	return fmt.Errorf("unsupported or unknown Spanner type")
}
