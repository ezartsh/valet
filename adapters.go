package valet

import (
	"context"
	"database/sql"
	"strings"
)

// DBQuerier is a minimal interface that both *sql.DB and *sql.Tx satisfy
type DBQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// SQLAdapter implements DBChecker for standard database/sql
type SQLAdapter struct {
	db DBQuerier
}

// NewSQLAdapter creates a checker for database/sql compatible connections
func NewSQLAdapter(db DBQuerier) *SQLAdapter {
	return &SQLAdapter{db: db}
}

// NewSQLChecker is an alias for NewSQLAdapter (convenience)
func NewSQLChecker(db *sql.DB) *SQLAdapter {
	return NewSQLAdapter(db)
}

// CheckExists implements DBChecker using a single IN query
func (s *SQLAdapter) CheckExists(ctx context.Context, table string, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
	if len(values) == 0 {
		return make(map[any]bool), nil
	}

	if s.db == nil {
		return nil, ErrNilDBConnection
	}

	query, args := buildExistsQuery(table, column, values, wheres)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[any]bool, len(values))
	for rows.Next() {
		var val any
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		result[val] = true
	}

	return result, rows.Err()
}

// FuncAdapter allows using a simple function as DBChecker
type FuncAdapter func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error)

func (f FuncAdapter) CheckExists(ctx context.Context, table string, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
	return f(ctx, table, column, values, wheres)
}

// SQLXQuerier interface for sqlx compatibility
type SQLXQuerier interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// SQLXAdapter implements DBChecker for jmoiron/sqlx
type SQLXAdapter struct {
	db SQLXQuerier
}

// NewSQLXAdapter creates a checker for sqlx
func NewSQLXAdapter(db SQLXQuerier) *SQLXAdapter {
	return &SQLXAdapter{db: db}
}

func (s *SQLXAdapter) CheckExists(ctx context.Context, table string, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
	if len(values) == 0 {
		return make(map[any]bool), nil
	}

	if s.db == nil {
		return nil, ErrNilDBConnection
	}

	query, args := buildExistsQuery(table, column, values, wheres)

	var results []interface{}
	if err := s.db.SelectContext(ctx, &results, query, args...); err != nil {
		return nil, err
	}

	resultMap := make(map[any]bool, len(results))
	for _, v := range results {
		resultMap[v] = true
	}
	return resultMap, nil
}

// GormQuerier is a simple interface for GORM-like ORMs
type GormQuerier interface {
	Raw(ctx context.Context, sql string, values ...interface{}) GormResult
}

// GormResult represents the result of a GORM query
type GormResult interface {
	Scan(dest interface{}) error
}

// GormAdapter implements DBChecker for GORM-like ORMs
type GormAdapter struct {
	querier GormQuerier
}

// NewGormAdapter creates an adapter for GORM-like ORMs
func NewGormAdapter(q GormQuerier) *GormAdapter {
	return &GormAdapter{querier: q}
}

func (g *GormAdapter) CheckExists(ctx context.Context, table string, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
	if len(values) == 0 {
		return make(map[any]bool), nil
	}

	if g.querier == nil {
		return nil, ErrNilDBConnection
	}

	query, args := buildExistsQuery(table, column, values, wheres)

	var results []interface{}
	if err := g.querier.Raw(ctx, query, args...).Scan(&results); err != nil {
		return nil, err
	}

	resultMap := make(map[any]bool, len(results))
	for _, v := range results {
		resultMap[v] = true
	}
	return resultMap, nil
}

// BunQuerier interface for uptrace/bun compatibility
type BunQuerier interface {
	NewRaw(query string, args ...interface{}) BunRawQuery
}

type BunRawQuery interface {
	Scan(ctx context.Context, dest ...interface{}) error
}

// BunAdapter implements DBChecker for uptrace/bun
type BunAdapter struct {
	db BunQuerier
}

// NewBunAdapter creates a checker for bun ORM
func NewBunAdapter(db BunQuerier) *BunAdapter {
	return &BunAdapter{db: db}
}

func (b *BunAdapter) CheckExists(ctx context.Context, table string, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
	if len(values) == 0 {
		return make(map[any]bool), nil
	}

	if b.db == nil {
		return nil, ErrNilDBConnection
	}

	query, args := buildExistsQuery(table, column, values, wheres)

	var results []interface{}
	if err := b.db.NewRaw(query, args...).Scan(ctx, &results); err != nil {
		return nil, err
	}

	resultMap := make(map[any]bool, len(results))
	for _, v := range results {
		resultMap[v] = true
	}
	return resultMap, nil
}

// buildExistsQuery builds the SQL query for existence check
func buildExistsQuery(table, column string, values []any, wheres []WhereClause) (string, []any) {
	var query strings.Builder
	query.WriteString("SELECT ")
	query.WriteString(column)
	query.WriteString(" FROM ")
	query.WriteString(table)
	query.WriteString(" WHERE ")
	query.WriteString(column)
	query.WriteString(" IN (")

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	query.WriteString(strings.Join(placeholders, ","))
	query.WriteString(")")

	args := make([]any, len(values))
	copy(args, values)

	for _, w := range wheres {
		query.WriteString(" AND ")
		query.WriteString(w.Column)
		query.WriteString(" ")
		query.WriteString(w.Operator)
		query.WriteString(" ?")
		args = append(args, w.Value)
	}

	return query.String(), args
}
