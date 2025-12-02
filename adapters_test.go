package valet

import (
	"context"
	"errors"
	"testing"
)

func TestFuncAdapter(t *testing.T) {
	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		result := make(map[any]bool)
		// Simulate: only values 1, 2, 3 exist in database
		for _, v := range values {
			if num, ok := v.(float64); ok && num >= 1 && num <= 3 {
				result[v] = true
			}
		}
		return result, nil
	})

	t.Run("check existing values", func(t *testing.T) {
		result, err := adapter.CheckExists(context.Background(), "users", "id", []any{float64(1), float64(2)}, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !result[float64(1)] || !result[float64(2)] {
			t.Error("Expected 1 and 2 to exist")
		}
	})

	t.Run("check non-existing values", func(t *testing.T) {
		result, err := adapter.CheckExists(context.Background(), "users", "id", []any{float64(5), float64(6)}, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result[float64(5)] || result[float64(6)] {
			t.Error("Expected 5 and 6 to not exist")
		}
	})

	t.Run("mixed values", func(t *testing.T) {
		result, err := adapter.CheckExists(context.Background(), "users", "id", []any{float64(1), float64(5)}, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !result[float64(1)] {
			t.Error("Expected 1 to exist")
		}
		if result[float64(5)] {
			t.Error("Expected 5 to not exist")
		}
	})

	t.Run("empty values", func(t *testing.T) {
		result, err := adapter.CheckExists(context.Background(), "users", "id", []any{}, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Error("Expected empty result for empty values")
		}
	})
}

func TestBuildExistsQuery(t *testing.T) {
	t.Run("simple query", func(t *testing.T) {
		query, args := buildExistsQuery("users", "id", []any{1, 2, 3}, nil)

		expectedQuery := "SELECT id FROM users WHERE id IN (?,?,?)"
		if query != expectedQuery {
			t.Errorf("Query = %s, want %s", query, expectedQuery)
		}

		if len(args) != 3 {
			t.Errorf("Args length = %d, want 3", len(args))
		}
	})

	t.Run("with where clauses", func(t *testing.T) {
		wheres := []WhereClause{
			{Column: "status", Operator: "=", Value: "active"},
			{Column: "deleted", Operator: "!=", Value: true},
		}
		query, args := buildExistsQuery("users", "email", []any{"a@test.com"}, wheres)

		if len(args) != 3 { // 1 value + 2 where values
			t.Errorf("Args length = %d, want 3", len(args))
		}

		// Check query contains WHERE clauses
		if query == "" {
			t.Error("Query should not be empty")
		}
	})

	t.Run("empty values", func(t *testing.T) {
		query, args := buildExistsQuery("users", "id", []any{}, nil)

		expectedQuery := "SELECT id FROM users WHERE id IN ()"
		if query != expectedQuery {
			t.Errorf("Query = %s, want %s", query, expectedQuery)
		}

		if len(args) != 0 {
			t.Errorf("Args length = %d, want 0", len(args))
		}
	})
}

func TestSQLAdapter_CheckExists_EmptyValues(t *testing.T) {
	// Test that empty values returns empty map without DB call
	adapter := &SQLAdapter{db: nil}
	result, err := adapter.CheckExists(context.Background(), "users", "id", []any{}, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty values")
	}
}

func TestSQLXAdapter_CheckExists_EmptyValues(t *testing.T) {
	adapter := &SQLXAdapter{db: nil}
	result, err := adapter.CheckExists(context.Background(), "users", "id", []any{}, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty values")
	}
}

func TestGormAdapter_CheckExists_EmptyValues(t *testing.T) {
	adapter := &GormAdapter{querier: nil}
	result, err := adapter.CheckExists(context.Background(), "users", "id", []any{}, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty values")
	}
}

func TestBunAdapter_CheckExists_EmptyValues(t *testing.T) {
	adapter := &BunAdapter{db: nil}
	result, err := adapter.CheckExists(context.Background(), "users", "id", []any{}, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty values")
	}
}

func TestNewSQLAdapter(t *testing.T) {
	adapter := NewSQLAdapter(nil)
	if adapter == nil {
		t.Error("NewSQLAdapter should return non-nil adapter")
	}
}

func TestNewSQLChecker(t *testing.T) {
	adapter := NewSQLChecker(nil)
	if adapter == nil {
		t.Error("NewSQLChecker should return non-nil adapter")
	}
}

func TestNewSQLXAdapter(t *testing.T) {
	adapter := NewSQLXAdapter(nil)
	if adapter == nil {
		t.Error("NewSQLXAdapter should return non-nil adapter")
	}
}

func TestNewGormAdapter(t *testing.T) {
	adapter := NewGormAdapter(nil)
	if adapter == nil {
		t.Error("NewGormAdapter should return non-nil adapter")
	}
}

func TestNewBunAdapter(t *testing.T) {
	adapter := NewBunAdapter(nil)
	if adapter == nil {
		t.Error("NewBunAdapter should return non-nil adapter")
	}
}

func TestNilDBConnectionError(t *testing.T) {
	ctx := context.Background()
	values := []any{float64(1), float64(2)}

	t.Run("SQLAdapter nil connection", func(t *testing.T) {
		adapter := NewSQLAdapter(nil)
		_, err := adapter.CheckExists(ctx, "users", "id", values, nil)
		if err != ErrNilDBConnection {
			t.Errorf("Expected ErrNilDBConnection, got %v", err)
		}
	})

	t.Run("SQLXAdapter nil connection", func(t *testing.T) {
		adapter := NewSQLXAdapter(nil)
		_, err := adapter.CheckExists(ctx, "users", "id", values, nil)
		if err != ErrNilDBConnection {
			t.Errorf("Expected ErrNilDBConnection, got %v", err)
		}
	})

	t.Run("GormAdapter nil connection", func(t *testing.T) {
		adapter := NewGormAdapter(nil)
		_, err := adapter.CheckExists(ctx, "users", "id", values, nil)
		if err != ErrNilDBConnection {
			t.Errorf("Expected ErrNilDBConnection, got %v", err)
		}
	})

	t.Run("BunAdapter nil connection", func(t *testing.T) {
		adapter := NewBunAdapter(nil)
		_, err := adapter.CheckExists(ctx, "users", "id", values, nil)
		if err != ErrNilDBConnection {
			t.Errorf("Expected ErrNilDBConnection, got %v", err)
		}
	})

	t.Run("Empty values returns empty map without error", func(t *testing.T) {
		adapter := NewSQLAdapter(nil)
		result, err := adapter.CheckExists(ctx, "users", "id", []any{}, nil)
		if err != nil {
			t.Errorf("Expected no error for empty values, got %v", err)
		}
		if len(result) != 0 {
			t.Error("Expected empty result for empty values")
		}
	})
}

// ============================================================================
// EXAMPLE TESTS - Demonstrating how to use each adapter
// ============================================================================

// ExampleFuncAdapter demonstrates using FuncAdapter for simple cases
func ExampleFuncAdapter() {
	// FuncAdapter wraps a simple function as DBChecker
	checker := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		// Simulate database lookup
		existing := map[any]bool{
			float64(1): true,
			float64(2): true,
			float64(3): true,
		}

		result := make(map[any]bool)
		for _, v := range values {
			result[v] = existing[v]
		}
		return result, nil
	})

	// Use with validation
	data := DataObject{"user_id": float64(1)}
	schema := Schema{
		"user_id": Float().Required().Exists("users", "id"),
	}

	_ = ValidateWithDB(context.Background(), data, schema, checker)
	// Output validation result
}

// ExampleSQLAdapter demonstrates using SQLAdapter with database/sql
func ExampleSQLAdapter() {
	// In real usage:
	// db, _ := sql.Open("mysql", "user:password@/dbname")
	// checker := NewSQLAdapter(db)

	// Example schema with DB validation
	/*
		data := DataObject{
			"email":   "user@example.com",
			"role_id": float64(1),
		}

		schema := Schema{
			"email":   String().Required().Email().Unique("users", "email", nil),
			"role_id": Float().Required().Exists("roles", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, checker)
		if err != nil {
			// Handle validation errors
		}
	*/
}

// ExampleGormAdapter demonstrates using GormAdapter with GORM
func ExampleGormAdapter() {
	// In real usage with GORM:
	// db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// checker := NewGormAdapter(db)

	// Example schema with DB validation
	/*
		data := DataObject{
			"username": "john_doe",
			"category_id": float64(5),
		}

		schema := Schema{
			"username":    String().Required().Min(3).Unique("users", "username", nil),
			"category_id": Float().Required().Exists("categories", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, checker)
		if err != nil {
			// Handle validation errors
		}
	*/
}

// ExampleSQLXAdapter demonstrates using SQLXAdapter with sqlx
func ExampleSQLXAdapter() {
	// In real usage with sqlx:
	// db, _ := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")
	// checker := NewSQLXAdapter(db)

	// Example schema with DB validation
	/*
		data := DataObject{
			"product_id": float64(100),
			"tag_ids":    []any{float64(1), float64(2), float64(3)},
		}

		schema := Schema{
			"product_id": Float().Required().Exists("products", "id"),
			"tag_ids":    Array().Required().Exists("tags", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, checker)
		if err != nil {
			// Handle validation errors
		}
	*/
}

// ExampleBunAdapter demonstrates using BunAdapter with Bun ORM
func ExampleBunAdapter() {
	// In real usage with Bun:
	// sqldb, _ := sql.Open("postgres", dsn)
	// db := bun.NewDB(sqldb, pgdialect.New())
	// checker := NewBunAdapter(db)

	// Example schema with DB validation
	/*
		data := DataObject{
			"author_id": float64(1),
			"status":    "published",
		}

		schema := Schema{
			"author_id": Float().Required().Exists("authors", "id"),
			"status":    String().Required().In("draft", "published", "archived"),
		}

		err := ValidateWithDB(context.Background(), data, schema, checker)
		if err != nil {
			// Handle validation errors
		}
	*/
}

// ============================================================================
// MOCK ADAPTER TESTS - Testing adapter behavior with mocks
// ============================================================================

// MockSQLRows simulates sql.Rows for testing
type MockSQLRows struct {
	values  []any
	index   int
	columns []string
}

func (m *MockSQLRows) Next() bool {
	if m.index < len(m.values) {
		m.index++
		return true
	}
	return false
}

func (m *MockSQLRows) Scan(dest ...any) error {
	if m.index > 0 && m.index <= len(m.values) {
		if ptr, ok := dest[0].(*any); ok {
			*ptr = m.values[m.index-1]
		}
	}
	return nil
}

func (m *MockSQLRows) Close() error {
	return nil
}

func (m *MockSQLRows) Err() error {
	return nil
}

func (m *MockSQLRows) Columns() ([]string, error) {
	return m.columns, nil
}

func TestAdapterWithWhereClause(t *testing.T) {
	// Test that where clauses are properly passed to adapters
	var capturedWheres []WhereClause

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		capturedWheres = wheres
		return map[any]bool{float64(1): true}, nil
	})

	wheres := []WhereClause{
		Where("status", "=", "active"),
		WhereEq("tenant_id", 123),
		WhereNot("deleted", true),
	}

	adapter.CheckExists(context.Background(), "users", "id", []any{float64(1)}, wheres)

	if len(capturedWheres) != 3 {
		t.Errorf("Expected 3 where clauses, got %d", len(capturedWheres))
	}

	// Verify first where clause
	if capturedWheres[0].Column != "status" || capturedWheres[0].Operator != "=" || capturedWheres[0].Value != "active" {
		t.Error("First where clause not captured correctly")
	}

	// Verify WhereEq
	if capturedWheres[1].Column != "tenant_id" || capturedWheres[1].Operator != "=" || capturedWheres[1].Value != 123 {
		t.Error("WhereEq clause not captured correctly")
	}

	// Verify WhereNot
	if capturedWheres[2].Column != "deleted" || capturedWheres[2].Operator != "!=" || capturedWheres[2].Value != true {
		t.Error("WhereNot clause not captured correctly")
	}
}

func TestAdapterTableColumnCapture(t *testing.T) {
	var capturedTable, capturedColumn string

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		capturedTable = table
		capturedColumn = column
		return make(map[any]bool), nil
	})

	adapter.CheckExists(context.Background(), "products", "sku", []any{"ABC123"}, nil)

	if capturedTable != "products" {
		t.Errorf("Expected table 'products', got '%s'", capturedTable)
	}
	if capturedColumn != "sku" {
		t.Errorf("Expected column 'sku', got '%s'", capturedColumn)
	}
}

func TestAdapterContextPropagation(t *testing.T) {
	type ctxKey string
	key := ctxKey("test_key")

	var capturedCtx context.Context

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		capturedCtx = ctx
		return make(map[any]bool), nil
	})

	ctx := context.WithValue(context.Background(), key, "test_value")
	adapter.CheckExists(ctx, "users", "id", []any{1}, nil)

	if capturedCtx.Value(key) != "test_value" {
		t.Error("Context not properly propagated to adapter")
	}
}

func TestAdapterErrorHandling(t *testing.T) {
	expectedErr := errors.New("database connection failed")

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		return nil, expectedErr
	})

	_, err := adapter.CheckExists(context.Background(), "users", "id", []any{1}, nil)

	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestAdapterBatchValues(t *testing.T) {
	var capturedValues []any

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		capturedValues = values
		result := make(map[any]bool)
		for _, v := range values {
			result[v] = true
		}
		return result, nil
	})

	inputValues := []any{float64(1), float64(2), float64(3), "abc", "def"}
	adapter.CheckExists(context.Background(), "items", "id", inputValues, nil)

	if len(capturedValues) != 5 {
		t.Errorf("Expected 5 values, got %d", len(capturedValues))
	}

	// Verify all values captured
	for i, v := range inputValues {
		if capturedValues[i] != v {
			t.Errorf("Value at index %d: expected %v, got %v", i, v, capturedValues[i])
		}
	}
}

// ============================================================================
// INTEGRATION TEST - Full validation flow with adapters
// ============================================================================

func TestAdapterIntegrationWithValidation(t *testing.T) {
	// Create a mock adapter that simulates a real database
	mockDB := map[string]map[string]map[any]bool{
		"users": {
			"id":    {float64(1): true, float64(2): true, float64(3): true},
			"email": {"john@example.com": true, "jane@example.com": true},
		},
		"roles": {
			"id": {float64(1): true, float64(2): true},
		},
		"categories": {
			"id": {float64(10): true, float64(20): true, float64(30): true},
		},
	}

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		result := make(map[any]bool)
		if tableData, ok := mockDB[table]; ok {
			if columnData, ok := tableData[column]; ok {
				for _, v := range values {
					result[v] = columnData[v]
				}
			}
		}
		return result, nil
	})

	t.Run("all exists checks pass", func(t *testing.T) {
		data := DataObject{
			"user_id":     float64(1),
			"role_id":     float64(2),
			"category_id": float64(10),
		}

		schema := Schema{
			"user_id":     Float().Required().Exists("users", "id"),
			"role_id":     Float().Required().Exists("roles", "id"),
			"category_id": Float().Required().Exists("categories", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("exists check fails", func(t *testing.T) {
		data := DataObject{
			"user_id": float64(999), // Does not exist
		}

		schema := Schema{
			"user_id": Float().Required().Exists("users", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent user_id")
		}
	})

	t.Run("unique check passes", func(t *testing.T) {
		data := DataObject{
			"email": "newuser@example.com", // Does not exist, so unique
		}

		schema := Schema{
			"email": String().Required().Email().Unique("users", "email", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors for unique email, got: %v", err.Errors)
		}
	})

	t.Run("unique check fails", func(t *testing.T) {
		data := DataObject{
			"email": "john@example.com", // Already exists
		}

		schema := Schema{
			"email": String().Required().Email().Unique("users", "email", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for taken email")
		}
	})

	t.Run("unique check with ignore", func(t *testing.T) {
		data := DataObject{
			"email": "john@example.com", // Exists but should be ignored (updating own record)
		}

		schema := Schema{
			"email": String().Required().Email().Unique("users", "email", "john@example.com"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors when ignoring own email, got: %v", err.Errors)
		}
	})

	t.Run("mixed validation and db checks", func(t *testing.T) {
		data := DataObject{
			"name":    "Jo",         // Too short (min 3)
			"user_id": float64(999), // Does not exist
		}

		schema := Schema{
			"name":    String().Required().Min(3),
			"user_id": Float().Required().Exists("users", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected validation errors")
		}

		// Should have error for name (validation fails first, DB checks may not run)
		if _, ok := err.Errors["name"]; !ok {
			t.Error("Expected error for name field")
		}
	})
}
