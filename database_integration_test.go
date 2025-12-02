package valet

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
)

// ============================================================================
// MOCK DB CHECKER FOR TESTING
// ============================================================================

// MockDBChecker is a mock implementation for testing
type MockDBChecker struct {
	// ExistingValues maps table:column -> set of existing values
	ExistingValues map[string]map[any]bool
	QueryCount     int32 // Use int32 for atomic operations
}

func NewMockDBChecker() *MockDBChecker {
	return &MockDBChecker{
		ExistingValues: make(map[string]map[any]bool),
	}
}

func (m *MockDBChecker) AddExisting(table, column string, values ...any) {
	key := table + ":" + column
	if m.ExistingValues[key] == nil {
		m.ExistingValues[key] = make(map[any]bool)
	}
	for _, v := range values {
		m.ExistingValues[key][v] = true
	}
}

func (m *MockDBChecker) CheckExists(ctx context.Context, table string, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
	atomic.AddInt32(&m.QueryCount, 1) // Thread-safe increment

	key := table + ":" + column
	existing := m.ExistingValues[key]

	result := make(map[any]bool)
	for _, v := range values {
		if existing != nil && existing[v] {
			result[v] = true
		}
	}
	return result, nil
}

func (m *MockDBChecker) Reset() {
	atomic.StoreInt32(&m.QueryCount, 0)
}

// ============================================================================
// DB VALIDATOR TESTS
// ============================================================================

func TestDBValidator_ExistsCheck_Single(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))

	// Validate with DB check
	data := DataObject{
		"user_id": float64(1),
	}

	schema := Schema{
		"user_id": Float().Required().Exists("users", "id"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err.Errors)
	}

	if mock.QueryCount != 1 {
		t.Errorf("Expected 1 query, got %d", mock.QueryCount)
	}
}

func TestDBValidator_ExistsCheck_NotFound(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))

	data := DataObject{
		"user_id": float64(999),
	}

	schema := Schema{
		"user_id": Float().Required().Exists("users", "id"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err == nil {
		t.Error("Expected error for non-existent user_id")
	}
	if _, ok := err.Errors["user_id"]; !ok {
		t.Error("Expected error for user_id field")
	}
}

func TestDBValidator_BatchesMultipleValues_SingleQuery(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("categories", "id", float64(1), float64(2), float64(3), float64(4), float64(5))

	// Multiple fields checking same table should batch
	data := DataObject{
		"category_1": float64(1),
		"category_2": float64(2),
		"category_3": float64(3),
		"category_4": float64(6), // Does not exist
		"category_5": float64(7), // Does not exist
	}

	schema := Schema{
		"category_1": Float().Required().Exists("categories", "id"),
		"category_2": Float().Required().Exists("categories", "id"),
		"category_3": Float().Required().Exists("categories", "id"),
		"category_4": Float().Required().Exists("categories", "id"),
		"category_5": Float().Required().Exists("categories", "id"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)

	// Should have errors for category_4 and category_5
	if err == nil {
		t.Error("Expected errors for non-existent categories")
	}

	// Should batch into ONE query, not 5!
	if mock.QueryCount != 1 {
		t.Errorf("Expected 1 query (batched), got %d queries (N+1 problem!)", mock.QueryCount)
	}
}

func TestDBValidator_UniqueCheck(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "email", "taken@example.com", "also@taken.com")

	data := DataObject{
		"email": "taken@example.com",
	}

	schema := Schema{
		"email": String().Required().Email().Unique("users", "email", nil),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err == nil {
		t.Error("Expected error for taken email")
	}
}

func TestDBValidator_UniqueCheck_WithIgnore(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "email", "myemail@example.com")

	data := DataObject{
		"email": "myemail@example.com",
	}

	// When updating, ignore current user's email
	schema := Schema{
		"email": String().Required().Email().Unique("users", "email", "myemail@example.com"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err != nil {
		t.Errorf("Expected no error when ignoring own value, got: %v", err.Errors)
	}
}

func TestDBValidator_UniqueCheck_Available(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "email", "taken@example.com")

	data := DataObject{
		"email": "available@example.com",
	}

	schema := Schema{
		"email": String().Required().Email().Unique("users", "email", nil),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err != nil {
		t.Errorf("Expected no errors for available email, got: %v", err.Errors)
	}
}

func TestDBValidator_MixedChecks_BatchedByTable(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))
	mock.AddExisting("categories", "id", float64(10), float64(20), float64(30))
	mock.AddExisting("tags", "id", float64(100), float64(200))

	data := DataObject{
		"user_id":       float64(1),
		"user_id_2":     float64(2),
		"category_id":   float64(10),
		"category_id_2": float64(20),
		"tag_id":        float64(100),
	}

	schema := Schema{
		"user_id":       Float().Required().Exists("users", "id"),
		"user_id_2":     Float().Required().Exists("users", "id"),
		"category_id":   Float().Required().Exists("categories", "id"),
		"category_id_2": Float().Required().Exists("categories", "id"),
		"tag_id":        Float().Required().Exists("tags", "id"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err.Errors)
	}

	// Should batch by table: 3 queries (users, categories, tags), not 5
	if mock.QueryCount != 3 {
		t.Errorf("Expected 3 queries (batched by table), got %d", mock.QueryCount)
	}
}

func TestDBValidator_WithWhereClause(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("products", "id", float64(1), float64(2), float64(3))

	data := DataObject{
		"product_id": float64(1),
	}

	schema := Schema{
		"product_id": Float().Required().Exists("products", "id",
			WhereEq("status", "active"),
			WhereEq("tenant_id", 123),
		),
	}

	_, err := ValidateWithDBContext(context.Background(), data, schema, Options{DBChecker: mock})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Query should have been executed
	if mock.QueryCount != 1 {
		t.Errorf("Expected 1 query, got %d", mock.QueryCount)
	}
}

func TestDBValidator_StringExists(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("roles", "name", "admin", "user", "guest")

	data := DataObject{
		"role": "admin",
	}

	schema := Schema{
		"role": String().Required().Exists("roles", "name"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err.Errors)
	}
}

func TestDBValidator_StringUnique(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "username", "john", "jane")

	t.Run("taken username", func(t *testing.T) {
		mock.Reset()
		data := DataObject{"username": "john"}
		schema := Schema{
			"username": String().Required().Unique("users", "username", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, mock)
		if err == nil {
			t.Error("Expected error for taken username")
		}
	})

	t.Run("available username", func(t *testing.T) {
		mock.Reset()
		data := DataObject{"username": "newuser"}
		schema := Schema{
			"username": String().Required().Unique("users", "username", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, mock)
		if err != nil {
			t.Errorf("Expected no error for available username, got: %v", err.Errors)
		}
	})
}

func TestDBValidator_ArrayExists(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("tags", "id", float64(1), float64(2), float64(3))

	data := DataObject{
		"tag_ids": []any{float64(1), float64(2), float64(3)},
	}

	schema := Schema{
		"tag_ids": Array().Required().Exists("tags", "id"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err.Errors)
	}

	// Should batch all array values into single query
	if mock.QueryCount != 1 {
		t.Errorf("Expected 1 query for array, got %d", mock.QueryCount)
	}
}

func TestDBValidator_ArrayExists_SomeNotFound(t *testing.T) {
	mock := NewMockDBChecker()
	mock.AddExisting("tags", "id", float64(1), float64(2), float64(3))

	data := DataObject{
		"tag_ids": []any{float64(1), float64(2), float64(999)}, // 999 doesn't exist
	}

	schema := Schema{
		"tag_ids": Array().Required().Exists("tags", "id"),
	}

	err := ValidateWithDB(context.Background(), data, schema, mock)
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

// ============================================================================
// DB VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkDBValidator_Batched_vs_Individual(b *testing.B) {
	mock := NewMockDBChecker()
	for i := 1; i <= 1000; i++ {
		mock.AddExisting("items", "id", float64(i))
	}

	b.Run("Batched_100_items", func(b *testing.B) {
		data := make(DataObject)
		schema := make(Schema)

		for j := 1; j <= 100; j++ {
			key := fmt.Sprintf("item_%d", j)
			data[key] = float64(j)
			schema[key] = Float().Required().Exists("items", "id")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)

			if mock.QueryCount != 1 {
				b.Fatalf("Expected 1 query, got %d", mock.QueryCount)
			}
		}
	})
}

func BenchmarkIntegratedValidation(b *testing.B) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))
	mock.AddExisting("categories", "id", float64(10), float64(20), float64(30))

	data := DataObject{
		"name":        "Test Product",
		"user_id":     float64(1),
		"category_id": float64(10),
	}

	b.Run("SchemaOnly", func(b *testing.B) {
		schema := Schema{
			"name":        String().Required().Min(2).Max(100),
			"user_id":     Float().Required(),
			"category_id": Float().Required(),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Validate(data, schema)
		}
	})

	b.Run("SchemaWithDBExists", func(b *testing.B) {
		schema := Schema{
			"name":        String().Required().Min(2).Max(100),
			"user_id":     Float().Required().Exists("users", "id"),
			"category_id": Float().Required().Exists("categories", "id"),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})

	b.Run("SchemaWithDBUnique", func(b *testing.B) {
		mockUnique := NewMockDBChecker()
		mockUnique.AddExisting("users", "email", "other@example.com")

		emailData := DataObject{"email": "new@example.com"}
		schema := Schema{
			"email": String().Required().Email().Unique("users", "email", nil),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mockUnique.Reset()
			ValidateWithDB(context.Background(), emailData, schema, mockUnique)
		}
	})
}

func BenchmarkBatchingEfficiency(b *testing.B) {
	mock := NewMockDBChecker()
	for i := 1; i <= 1000; i++ {
		mock.AddExisting("items", "id", float64(i))
	}

	itemCounts := []int{1, 10, 50, 100, 500}

	for _, count := range itemCounts {
		b.Run(fmt.Sprintf("Items_%d", count), func(b *testing.B) {
			data := make(DataObject)
			schema := make(Schema)

			for j := 1; j <= count; j++ {
				key := fmt.Sprintf("item_%d", j)
				data[key] = float64(j)
				schema[key] = Float().Required().Exists("items", "id")
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mock.Reset()
				ValidateWithDB(context.Background(), data, schema, mock)

				if mock.QueryCount != 1 {
					b.Fatalf("Expected 1 query for %d items, got %d", count, mock.QueryCount)
				}
			}
		})
	}
}

func BenchmarkMultiTableBatching(b *testing.B) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))
	mock.AddExisting("categories", "id", float64(10), float64(20))
	mock.AddExisting("tags", "id", float64(100), float64(200), float64(300))
	mock.AddExisting("brands", "id", float64(1000))

	b.Run("4_Tables_4_Checks", func(b *testing.B) {
		data := DataObject{
			"user_id":     float64(1),
			"category_id": float64(10),
			"tag_id":      float64(100),
			"brand_id":    float64(1000),
		}

		schema := Schema{
			"user_id":     Float().Required().Exists("users", "id"),
			"category_id": Float().Required().Exists("categories", "id"),
			"tag_id":      Float().Required().Exists("tags", "id"),
			"brand_id":    Float().Required().Exists("brands", "id"),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)

			// Should be exactly 4 queries (one per table)
			if mock.QueryCount != 4 {
				b.Fatalf("Expected 4 queries, got %d", mock.QueryCount)
			}
		}
	})
}

func BenchmarkDBAdapters(b *testing.B) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))

	data := DataObject{
		"user_id": float64(1),
	}

	schema := Schema{
		"user_id": Float().Required().Exists("users", "id"),
	}

	b.Run("MockDBChecker", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})

	// FuncAdapter benchmark
	funcChecker := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		result := make(map[any]bool)
		for _, v := range values {
			if v == float64(1) || v == float64(2) || v == float64(3) {
				result[v] = true
			}
		}
		return result, nil
	})

	b.Run("FuncAdapter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ValidateWithDB(context.Background(), data, schema, funcChecker)
		}
	})
}

func BenchmarkAdapterOverhead(b *testing.B) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1))

	b.Run("DirectInterface", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			mock.CheckExists(context.Background(), "users", "id", []any{float64(1)}, nil)
		}
	})

	funcChecker := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		return map[any]bool{float64(1): true}, nil
	})

	b.Run("FuncAdapter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			funcChecker.CheckExists(context.Background(), "users", "id", []any{float64(1)}, nil)
		}
	})
}

func BenchmarkQueryBuilding(b *testing.B) {
	valueCounts := []int{1, 10, 100}

	for _, count := range valueCounts {
		b.Run(fmt.Sprintf("BuildQuery_%d_values", count), func(b *testing.B) {
			values := make([]any, count)
			for i := 0; i < count; i++ {
				values[i] = i + 1
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buildExistsQuery("users", "id", values, nil)
			}
		})
	}
}

// ============================================================================
// NESTED ARRAY OBJECT DB CHECK BENCHMARKS
// ============================================================================

func BenchmarkNestedArrayObjectDBCheck(b *testing.B) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1), float64(2), float64(3))
	mock.AddExisting("products", "id", float64(100), float64(101), float64(102), float64(103), float64(104))
	mock.AddExisting("tags", "id", float64(1), float64(2), float64(3), float64(4), float64(5))

	b.Run("5_items_with_product_id", func(b *testing.B) {
		items := make([]any, 5)
		for i := 0; i < 5; i++ {
			items[i] = map[string]any{
				"product_id": float64(100 + i),
				"quantity":   float64(i + 1),
			}
		}

		data := DataObject{
			"order_number": "ORD-001",
			"customer_id":  float64(1),
			"items":        items,
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"product_id": Float().Required().Exists("products", "id"),
				"quantity":   Float().Required().Positive(),
			})),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})

	b.Run("10_items_with_product_id_and_tags", func(b *testing.B) {
		items := make([]any, 10)
		for i := 0; i < 10; i++ {
			items[i] = map[string]any{
				"product_id": float64(100 + (i % 5)),
				"quantity":   float64(i + 1),
				"tag_ids":    []any{float64(1), float64(2)},
			}
		}

		data := DataObject{
			"order_number": "ORD-002",
			"customer_id":  float64(1),
			"items":        items,
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"product_id": Float().Required().Exists("products", "id"),
				"quantity":   Float().Required().Positive(),
				"tag_ids":    Array().Required().Min(1).Exists("tags", "id"),
			})),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})

	b.Run("50_items_deeply_nested", func(b *testing.B) {
		items := make([]any, 50)
		for i := 0; i < 50; i++ {
			items[i] = map[string]any{
				"product_id": float64(100 + (i % 5)),
				"quantity":   float64(i + 1),
				"tag_ids":    []any{float64(1), float64(2), float64(3)},
			}
		}

		data := DataObject{
			"order_number": "ORD-003",
			"customer_id":  float64(1),
			"items":        items,
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"product_id": Float().Required().Exists("products", "id"),
				"quantity":   Float().Required().Positive(),
				"tag_ids":    Array().Required().Min(1).Exists("tags", "id"),
			})),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})
}

func BenchmarkNestedVsFlat(b *testing.B) {
	mock := NewMockDBChecker()
	mock.AddExisting("users", "id", float64(1))
	mock.AddExisting("products", "id", float64(100), float64(101), float64(102), float64(103), float64(104))

	// Flat structure - product_ids as array
	b.Run("Flat_5_product_ids", func(b *testing.B) {
		data := DataObject{
			"order_number": "ORD-001",
			"customer_id":  float64(1),
			"product_ids":  []any{float64(100), float64(101), float64(102), float64(103), float64(104)},
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"product_ids":  Array().Required().Min(1).Exists("products", "id"),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})

	// Nested structure - items array with product_id in each object
	b.Run("Nested_5_items_with_product_id", func(b *testing.B) {
		items := make([]any, 5)
		for i := 0; i < 5; i++ {
			items[i] = map[string]any{
				"product_id": float64(100 + i),
				"quantity":   float64(i + 1),
			}
		}

		data := DataObject{
			"order_number": "ORD-001",
			"customer_id":  float64(1),
			"items":        items,
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"product_id": Float().Required().Exists("products", "id"),
				"quantity":   Float().Required().Positive(),
			})),
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.Reset()
			ValidateWithDB(context.Background(), data, schema, mock)
		}
	})
}
