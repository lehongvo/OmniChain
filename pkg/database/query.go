package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryExecutor provides optimized query execution with prepared statements
type QueryExecutor struct {
	pool *pgxpool.Pool
	// Prepared statement cache
	preparedStmts map[string]*pgxpool.Conn
}

// NewQueryExecutor creates a new query executor with prepared statement support
func NewQueryExecutor(pool *pgxpool.Pool) *QueryExecutor {
	return &QueryExecutor{
		pool:          pool,
		preparedStmts: make(map[string]*pgxpool.Conn),
	}
}

// ExecuteWithTimeout executes a query with timeout
func (qe *QueryExecutor) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, query string, args ...interface{}) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return qe.pool.Query(ctx, query, args...)
}

// BatchInsert performs batch insert operation
func (qe *QueryExecutor) BatchInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// Build batch insert query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, formatColumns(columns))
	
	// Build VALUES clause
	valuePlaceholders := make([]string, len(values))
	allArgs := make([]interface{}, 0, len(values)*len(columns))
	argIndex := 1

	for i, row := range values {
		placeholders := make([]string, len(columns))
		for j := range columns {
			placeholders[j] = fmt.Sprintf("$%d", argIndex)
			allArgs = append(allArgs, row[j])
			argIndex++
		}
		valuePlaceholders[i] = fmt.Sprintf("(%s)", fmt.Sprintf("%s", placeholders))
	}

	query += fmt.Sprintf("%s", valuePlaceholders[0])
	for i := 1; i < len(valuePlaceholders); i++ {
		query += ", " + valuePlaceholders[i]
	}

	_, err := qe.pool.Exec(ctx, query, allArgs...)
	return err
}

// formatColumns formats column names for SQL
func formatColumns(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	result := columns[0]
	for i := 1; i < len(columns); i++ {
		result += ", " + columns[i]
	}
	return result
}

// BatchUpdate performs batch update operation
func (qe *QueryExecutor) BatchUpdate(ctx context.Context, table, idColumn string, updates map[string]interface{}) error {
	// This is a simplified version - in production, use proper batch update
	for id, data := range updates {
		// Build update query dynamically
		query := fmt.Sprintf("UPDATE %s SET ", table)
		args := make([]interface{}, 0)
		argIndex := 1

		updateParts := make([]string, 0)
		for col, val := range data.(map[string]interface{}) {
			updateParts = append(updateParts, fmt.Sprintf("%s = $%d", col, argIndex))
			args = append(args, val)
			argIndex++
		}

		query += fmt.Sprintf("%s", updateParts[0])
		for i := 1; i < len(updateParts); i++ {
			query += ", " + updateParts[i]
		}
		query += fmt.Sprintf(" WHERE %s = $%d", idColumn, argIndex)
		args = append(args, id)

		_, err := qe.pool.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

