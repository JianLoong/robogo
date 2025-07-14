package actions

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/JianLoong/robogo/internal/common"
)

// No persistent connections - open, use, close immediately for CLI tool simplicity

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

		return map[string]interface{}{
			"columns": columns,
			"rows":    results,
		}, nil

	case "execute", "insert", "update", "delete":
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("execute failed: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		return map[string]interface{}{
			"rows_affected": rowsAffected,
		}, nil

	default:
		return nil, fmt.Errorf("unknown postgres operation: %s", operation)
	}
}

// Spanner action - simplified implementation with proper resource management
func spannerAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("spanner action requires at least 3 arguments: operation, connection_string, query")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])
	query := fmt.Sprintf("%v", args[2])

	// Create client for this operation only
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer clientCancel()
	
	var client *spanner.Client
	var err error
	// Check if using emulator (either by connection string or environment variable)
	if strings.Contains(connectionString, "localhost:9010") || os.Getenv("SPANNER_EMULATOR_HOST") != "" {
		client, err = spanner.NewClient(clientCtx, connectionString,
			option.WithEndpoint("localhost:9010"),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	} else {
		client, err = spanner.NewClient(clientCtx, connectionString)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create spanner client: %w", err)
	}
	defer client.Close() // Always close client when done

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "query", "select":
		stmt := spanner.Statement{SQL: query}
		iter := client.Single().Query(ctx, stmt)
		defer func() {
			iter.Stop() // Always stop iterator to prevent resource leaks
		}()

		var results []map[string]interface{}
		var columns []string

		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to iterate results: %w", err)
			}

			if len(columns) == 0 {
				columns = row.ColumnNames()
			}

			rowMap := make(map[string]interface{})
			for i, col := range columns {
				// Try different common types for Spanner
				var intVal int64
				if err := row.Column(i, &intVal); err == nil {
					rowMap[col] = intVal
					continue
				}
				
				var strVal string
				if err := row.Column(i, &strVal); err == nil {
					rowMap[col] = strVal
					continue
				}
				
				var boolVal bool
				if err := row.Column(i, &boolVal); err == nil {
					rowMap[col] = boolVal
					continue
				}
				
				// If all specific types fail, store as nil
				rowMap[col] = nil
			}
			results = append(results, rowMap)
		}

		return map[string]interface{}{
			"columns": columns,
			"rows":    results,
		}, nil

	case "execute", "insert", "update", "delete":
		stmt := spanner.Statement{SQL: query}
		var rowsAffected int64

		_, err := client.ReadWriteTransaction(ctx, func(txnCtx context.Context, txn *spanner.ReadWriteTransaction) error {
			count, err := txn.Update(txnCtx, stmt)
			if err != nil {
				return err
			}
			rowsAffected = count
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("execute failed: %w", err)
		}

		return map[string]interface{}{
			"rows_affected": rowsAffected,
		}, nil

	default:
		return nil, fmt.Errorf("unknown spanner operation: %s", operation)
	}
}

// No cleanup needed - connections are closed after each operation