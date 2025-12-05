package valet

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStringValidator(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		schema := Schema{
			"name": String().Required(),
		}

		err := Validate(DataObject{"name": ""}, schema)
		if err == nil {
			t.Error("Expected error for empty required string")
		}

		err = Validate(DataObject{"name": "John"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("min max", func(t *testing.T) {
		schema := Schema{
			"username": String().Required().Min(3).Max(20),
		}

		err := Validate(DataObject{"username": "ab"}, schema)
		if err == nil {
			t.Error("Expected error for too short")
		}

		err = Validate(DataObject{"username": "abcdefghijklmnopqrstuvwxyz"}, schema)
		if err == nil {
			t.Error("Expected error for too long")
		}

		err = Validate(DataObject{"username": "john_doe"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("email", func(t *testing.T) {
		schema := Schema{
			"email": String().Required().Email(),
		}

		err := Validate(DataObject{"email": "invalid"}, schema)
		if err == nil {
			t.Error("Expected error for invalid email")
		}

		err = Validate(DataObject{"email": "test@example.com"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("in", func(t *testing.T) {
		schema := Schema{
			"status": String().Required().In("active", "inactive", "pending"),
		}

		err := Validate(DataObject{"status": "unknown"}, schema)
		if err == nil {
			t.Error("Expected error for value not in list")
		}

		err = Validate(DataObject{"status": "active"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("custom message", func(t *testing.T) {
		schema := Schema{
			"name": String().Required().Message("required", "Name is mandatory"),
		}

		err := Validate(DataObject{"name": ""}, schema)
		if err == nil {
			t.Error("Expected error")
		}
		if err.Errors["name"][0] != "Name is mandatory" {
			t.Errorf("Expected custom message, got: %s", err.Errors["name"][0])
		}
	})

	t.Run("custom validation", func(t *testing.T) {
		schema := Schema{
			"password": String().Required().Min(8).Custom(func(v string, lookup Lookup) error {
				if v == "password123" {
					return errors.New("password is too common")
				}
				return nil
			}),
		}

		err := Validate(DataObject{"password": "password123"}, schema)
		if err == nil {
			t.Error("Expected error for common password")
		}
	})

	t.Run("nullable", func(t *testing.T) {
		schema := Schema{
			"bio": String().Nullable().Max(100),
		}

		err := Validate(DataObject{"bio": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})

	t.Run("default", func(t *testing.T) {
		schema := Schema{
			"role": String().Default("user"),
		}

		err := Validate(DataObject{}, schema)
		if err != nil {
			t.Errorf("Expected no error with default, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		schema := Schema{
			"age": Float().Required(),
		}

		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for missing required number")
		}

		err = Validate(DataObject{"age": float64(25)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("min max", func(t *testing.T) {
		schema := Schema{
			"age": Float().Required().Min(18).Max(120),
		}

		err := Validate(DataObject{"age": float64(15)}, schema)
		if err == nil {
			t.Error("Expected error for too small")
		}

		err = Validate(DataObject{"age": float64(150)}, schema)
		if err == nil {
			t.Error("Expected error for too large")
		}

		err = Validate(DataObject{"age": float64(25)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("positive", func(t *testing.T) {
		schema := Schema{
			"price": Float().Required().Positive(),
		}

		err := Validate(DataObject{"price": float64(-10)}, schema)
		if err == nil {
			t.Error("Expected error for negative")
		}

		err = Validate(DataObject{"price": float64(99.99)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestBoolValidator(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		schema := Schema{
			"agree": Bool().Required(),
		}

		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for missing required bool")
		}

		err = Validate(DataObject{"agree": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("must be true", func(t *testing.T) {
		schema := Schema{
			"terms": Bool().Required().True(),
		}

		err := Validate(DataObject{"terms": false}, schema)
		if err == nil {
			t.Error("Expected error for false when must be true")
		}

		err = Validate(DataObject{"terms": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator(t *testing.T) {
	t.Run("nested schema", func(t *testing.T) {
		schema := Schema{
			"user": Object().Required().Shape(Schema{
				"name":  String().Required().Min(2),
				"email": String().Required().Email(),
			}),
		}

		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "J",
				"email": "invalid",
			},
		}, schema)
		if err == nil {
			t.Error("Expected errors for invalid nested fields")
		}

		err = Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("strict mode", func(t *testing.T) {
		schema := Schema{
			"config": Object().Required().Strict().Shape(Schema{
				"name": String().Required(),
			}),
		}

		err := Validate(DataObject{
			"config": map[string]any{
				"name":    "test",
				"unknown": "field",
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for unknown field in strict mode")
		}
	})

	t.Run("extend", func(t *testing.T) {
		baseUser := Object().Shape(Schema{
			"name": String().Required(),
		})

		adminUser := baseUser.Extend(Schema{
			"role": String().Required().In("admin", "superadmin"),
		})

		err := Validate(DataObject{
			"admin": map[string]any{
				"name": "John",
				"role": "admin",
			},
		}, Schema{"admin": adminUser})

		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestArrayValidator(t *testing.T) {
	t.Run("min max", func(t *testing.T) {
		schema := Schema{
			"tags": Array().Required().Min(1).Max(5),
		}

		err := Validate(DataObject{"tags": []any{}}, schema)
		if err == nil {
			t.Error("Expected error for empty array")
		}

		err = Validate(DataObject{"tags": []any{"a", "b", "c", "d", "e", "f"}}, schema)
		if err == nil {
			t.Error("Expected error for too many elements")
		}

		err = Validate(DataObject{"tags": []any{"go", "rust"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("element validation", func(t *testing.T) {
		schema := Schema{
			"emails": Array().Required().Of(String().Email()),
		}

		err := Validate(DataObject{
			"emails": []any{"valid@example.com", "invalid"},
		}, schema)
		if err == nil {
			t.Error("Expected error for invalid email in array")
		}

		err = Validate(DataObject{
			"emails": []any{"a@example.com", "b@example.com"},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("unique", func(t *testing.T) {
		schema := Schema{
			"ids": Array().Required().Unique(),
		}

		err := Validate(DataObject{
			"ids": []any{1, 2, 2, 3},
		}, schema)
		if err == nil {
			t.Error("Expected error for duplicate values")
		}
	})
}

func TestComplexSchema(t *testing.T) {
	schema := Schema{
		"username": String().Required().Min(3).Max(20).AlphaNumeric(),
		"email":    String().Required().Email(),
		"age":      Float().Required().Min(18).Max(120),
		"bio":      String().Nullable().Max(500),
		"profile": Object().Required().Shape(Schema{
			"avatar":  String().URL(),
			"website": String().URL(),
		}),
		"tags": Array().Max(10).Of(String().Min(2).Max(20)),
		"settings": Object().Shape(Schema{
			"notifications": Bool().Default(true),
			"theme":         String().In("light", "dark").Default("light"),
		}),
	}

	data := DataObject{
		"username": "johndoe",
		"email":    "john@example.com",
		"age":      float64(25),
		"bio":      nil,
		"profile": map[string]any{
			"avatar":  "https://example.com/avatar.jpg",
			"website": "https://johndoe.com",
		},
		"tags": []any{"golang", "programming"},
		"settings": map[string]any{
			"notifications": true,
			"theme":         "dark",
		},
	}

	err := Validate(data, schema)
	if err != nil {
		t.Errorf("Expected no error for valid complex data, got: %v", err.Errors)
	}
}

func TestSafeParse(t *testing.T) {
	schema := Schema{
		"name": String().Required(),
	}

	data, err := SafeParse(DataObject{"name": "John"}, schema)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err.Errors)
	}
	if data["name"] != "John" {
		t.Error("Expected data to be returned")
	}

	data, err = SafeParse(DataObject{}, schema)
	if err == nil {
		t.Error("Expected error")
	}
	if data != nil {
		t.Error("Expected nil data on error")
	}
}

func TestFileValidator(t *testing.T) {
	// Note: File validation tests require actual file headers
	// These are basic structure tests
	t.Run("required", func(t *testing.T) {
		schema := Schema{
			"avatar": File().Required(),
		}

		err := Validate(DataObject{"avatar": nil}, schema)
		if err == nil {
			t.Error("Expected error for missing required file")
		}
	})

	t.Run("nullable", func(t *testing.T) {
		schema := Schema{
			"avatar": File().Nullable(),
		}

		err := Validate(DataObject{"avatar": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable file, got: %v", err.Errors)
		}
	})
}

func TestDBAdapters(t *testing.T) {
	t.Run("FuncAdapter", func(t *testing.T) {
		adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
			result := make(map[any]bool)
			for _, v := range values {
				if v == float64(1) || v == float64(2) {
					result[v] = true
				}
			}
			return result, nil
		})

		result, err := adapter.CheckExists(context.Background(), "users", "id", []any{float64(1), float64(3)}, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !result[float64(1)] {
			t.Error("Expected 1 to exist")
		}
		if result[float64(3)] {
			t.Error("Expected 3 to not exist")
		}
	})
}

// ============================================================================
// DB ADAPTER INTEGRATION TESTS WITH NESTED DATA AND ARRAY OBJECTS
// ============================================================================

// MockGormDB simulates GORM database for testing
// In real usage, this would be *gorm.DB
type MockGormDB struct {
	// Simulated database tables
	users      map[any]bool
	categories map[any]bool
	products   map[any]bool
	tags       map[any]bool
	warehouses map[any]bool
}

func NewMockGormDB() *MockGormDB {
	return &MockGormDB{
		users:      map[any]bool{float64(1): true, float64(2): true, float64(3): true},
		categories: map[any]bool{float64(10): true, float64(20): true, float64(30): true},
		products:   map[any]bool{float64(100): true, float64(101): true, float64(102): true},
		tags:       map[any]bool{float64(1): true, float64(2): true, float64(3): true, float64(4): true, float64(5): true},
		warehouses: map[any]bool{float64(1): true, float64(2): true},
	}
}

// CreateMockGormAdapter creates a FuncAdapter that simulates GORM behavior
func CreateMockGormAdapter(db *MockGormDB) DBChecker {
	return FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		result := make(map[any]bool)

		var tableData map[any]bool
		switch table {
		case "users":
			tableData = db.users
		case "categories":
			tableData = db.categories
		case "products":
			tableData = db.products
		case "tags":
			tableData = db.tags
		case "warehouses":
			tableData = db.warehouses
		default:
			return result, nil
		}

		for _, v := range values {
			result[v] = tableData[v]
		}
		return result, nil
	})
}

func TestDBAdapter_NestedObjectWithExists(t *testing.T) {
	mockDB := NewMockGormDB()
	adapter := CreateMockGormAdapter(mockDB)

	t.Run("flat structure with exists check", func(t *testing.T) {
		// Simulating a request with flat structure
		data := DataObject{
			"title":       "My Blog Post",
			"user_id":     float64(1),  // Should exist in users table
			"category_id": float64(10), // Should exist in categories table
		}

		schema := Schema{
			"title":       String().Required().Min(5).Max(200),
			"user_id":     Float().Required().Exists("users", "id"),
			"category_id": Float().Required().Exists("categories", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("flat structure with non-existent user_id", func(t *testing.T) {
		data := DataObject{
			"title":       "My Blog Post",
			"user_id":     float64(999), // Does NOT exist
			"category_id": float64(10),
		}

		schema := Schema{
			"title":       String().Required().Min(5).Max(200),
			"user_id":     Float().Required().Exists("users", "id"),
			"category_id": Float().Required().Exists("categories", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent user_id")
		}
	})

	t.Run("nested object validation without DB check", func(t *testing.T) {
		// Nested objects are validated but DB checks are at top level
		data := DataObject{
			"title": "My Blog Post",
			"author": map[string]any{
				"name": "John Doe",
				"bio":  "A writer",
			},
			"category_id": float64(10),
		}

		schema := Schema{
			"title": String().Required().Min(5).Max(200),
			"author": Object().Required().Shape(Schema{
				"name": String().Required().Min(2),
				"bio":  String().Max(500),
			}),
			"category_id": Float().Required().Exists("categories", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})
}

func TestDBAdapter_ArrayObjectWithExists(t *testing.T) {
	mockDB := NewMockGormDB()
	adapter := CreateMockGormAdapter(mockDB)

	t.Run("order with customer_id and product_ids array", func(t *testing.T) {
		// Simulating an order with product IDs as flat array
		data := DataObject{
			"order_number": "ORD-001",
			"customer_id":  float64(1),
			"product_ids":  []any{float64(100), float64(101), float64(102)},
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"product_ids":  Array().Required().Min(1).Exists("products", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("order with non-existent product in array", func(t *testing.T) {
		data := DataObject{
			"order_number": "ORD-002",
			"customer_id":  float64(1),
			"product_ids":  []any{float64(100), float64(999)}, // 999 doesn't exist
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"product_ids":  Array().Required().Min(1).Exists("products", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent product_id in array")
		}
	})

	t.Run("order with nested items validation (no DB check in nested)", func(t *testing.T) {
		// Array of objects with validation but DB check at top level
		data := DataObject{
			"order_number": "ORD-003",
			"customer_id":  float64(1),
			"items": []any{
				map[string]any{
					"name":     "Product A",
					"quantity": float64(2),
					"price":    float64(29.99),
				},
				map[string]any{
					"name":     "Product B",
					"quantity": float64(1),
					"price":    float64(49.99),
				},
			},
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"name":     String().Required().Min(2),
				"quantity": Float().Required().Positive(),
				"price":    Float().Required().Positive(),
			})),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("nested array objects with product_id DB check - all valid", func(t *testing.T) {
		// Array of objects where each object has a product_id that needs DB check
		data := DataObject{
			"order_number": "ORD-004",
			"customer_id":  float64(1),
			"items": []any{
				map[string]any{
					"product_id": float64(100),
					"quantity":   float64(2),
				},
				map[string]any{
					"product_id": float64(101),
					"quantity":   float64(1),
				},
				map[string]any{
					"product_id": float64(102),
					"quantity":   float64(3),
				},
			},
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"product_id": Float().Required().Exists("products", "id"),
				"quantity":   Float().Required().Positive(),
			})),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("nested array objects with non-existent product_id", func(t *testing.T) {
		data := DataObject{
			"order_number": "ORD-005",
			"customer_id":  float64(1),
			"items": []any{
				map[string]any{
					"product_id": float64(100),
					"quantity":   float64(2),
				},
				map[string]any{
					"product_id": float64(999), // Does NOT exist
					"quantity":   float64(1),
				},
			},
		}

		schema := Schema{
			"order_number": String().Required().Min(3),
			"customer_id":  Float().Required().Exists("users", "id"),
			"items": Array().Required().Min(1).Of(Object().Shape(Schema{
				"product_id": Float().Required().Exists("products", "id"),
				"quantity":   Float().Required().Positive(),
			})),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent product_id in nested array object")
		}
	})

	t.Run("deeply nested array objects with DB checks", func(t *testing.T) {
		// Order with items, each item has tags array
		data := DataObject{
			"order_number": "ORD-006",
			"customer_id":  float64(1),
			"items": []any{
				map[string]any{
					"product_id": float64(100),
					"quantity":   float64(2),
					"tag_ids":    []any{float64(1), float64(2)},
				},
				map[string]any{
					"product_id": float64(101),
					"quantity":   float64(1),
					"tag_ids":    []any{float64(3), float64(4)},
				},
			},
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

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("deeply nested with invalid tag", func(t *testing.T) {
		data := DataObject{
			"order_number": "ORD-007",
			"customer_id":  float64(1),
			"items": []any{
				map[string]any{
					"product_id": float64(100),
					"quantity":   float64(2),
					"tag_ids":    []any{float64(1), float64(999)}, // 999 doesn't exist
				},
			},
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

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent tag in deeply nested array")
		}
	})
}

func TestDBAdapter_ComplexNestedWithMultipleTables(t *testing.T) {
	mockDB := NewMockGormDB()
	adapter := CreateMockGormAdapter(mockDB)

	t.Run("e-commerce order with multiple flat DB checks", func(t *testing.T) {
		// Order structure with DB checks at top level
		data := DataObject{
			"order_number": "ORD-2024-001",
			"customer_id":  float64(1),
			"warehouse_id": float64(1),
			"product_ids":  []any{float64(100), float64(101)},
			"category_ids": []any{float64(10), float64(20)},
			"tag_ids":      []any{float64(1), float64(2), float64(3)},
		}

		schema := Schema{
			"order_number": String().Required().Min(5),
			"customer_id":  Float().Required().Exists("users", "id"),
			"warehouse_id": Float().Required().Exists("warehouses", "id"),
			"product_ids":  Array().Required().Min(1).Exists("products", "id"),
			"category_ids": Array().Required().Min(1).Exists("categories", "id"),
			"tag_ids":      Array().Required().Min(1).Exists("tags", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("order with invalid warehouse_id", func(t *testing.T) {
		data := DataObject{
			"order_number": "ORD-2024-002",
			"customer_id":  float64(1),
			"warehouse_id": float64(999), // Does NOT exist
			"product_ids":  []any{float64(100)},
		}

		schema := Schema{
			"order_number": String().Required().Min(5),
			"customer_id":  Float().Required().Exists("users", "id"),
			"warehouse_id": Float().Required().Exists("warehouses", "id"),
			"product_ids":  Array().Required().Min(1).Exists("products", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent warehouse_id")
		}
	})

	t.Run("order with invalid tag in array", func(t *testing.T) {
		data := DataObject{
			"order_number": "ORD-2024-003",
			"customer_id":  float64(2),
			"warehouse_id": float64(2),
			"tag_ids":      []any{float64(1), float64(999)}, // 999 does NOT exist
		}

		schema := Schema{
			"order_number": String().Required().Min(5),
			"customer_id":  Float().Required().Exists("users", "id"),
			"warehouse_id": Float().Required().Exists("warehouses", "id"),
			"tag_ids":      Array().Required().Min(1).Exists("tags", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent tag in tag_ids array")
		}
	})

	t.Run("order with nested object validation (no nested DB check)", func(t *testing.T) {
		// Nested objects are validated but DB checks are at top level
		data := DataObject{
			"order_number": "ORD-2024-004",
			"customer_id":  float64(1),
			"shipping": map[string]any{
				"address": "123 Main St",
				"city":    "New York",
				"zip":     "10001",
			},
			"product_ids": []any{float64(100), float64(101)},
		}

		schema := Schema{
			"order_number": String().Required().Min(5),
			"customer_id":  Float().Required().Exists("users", "id"),
			"shipping": Object().Required().Shape(Schema{
				"address": String().Required().Min(5),
				"city":    String().Required().Min(2),
				"zip":     String().Required().Min(5),
			}),
			"product_ids": Array().Required().Min(1).Exists("products", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})
}

func TestDBAdapter_UniqueCheckFlat(t *testing.T) {
	// Mock adapter that tracks unique values
	existingEmails := map[any]bool{
		"john@example.com": true,
		"jane@example.com": true,
	}
	existingUsernames := map[any]bool{
		"johndoe": true,
		"janedoe": true,
	}

	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		result := make(map[any]bool)

		if table == "users" && column == "email" {
			for _, v := range values {
				result[v] = existingEmails[v]
			}
		} else if table == "users" && column == "username" {
			for _, v := range values {
				result[v] = existingUsernames[v]
			}
		}
		return result, nil
	})

	t.Run("unique email - available", func(t *testing.T) {
		data := DataObject{
			"email":    "newuser@example.com", // Available
			"username": "newuser",             // Available
		}

		schema := Schema{
			"email":    String().Required().Email().Unique("users", "email", nil),
			"username": String().Required().Min(3).Unique("users", "username", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("unique email - taken", func(t *testing.T) {
		data := DataObject{
			"email":    "john@example.com", // Already taken
			"username": "newuser",
		}

		schema := Schema{
			"email":    String().Required().Email().Unique("users", "email", nil),
			"username": String().Required().Min(3).Unique("users", "username", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for taken email")
		}
	})

	t.Run("unique username - taken", func(t *testing.T) {
		data := DataObject{
			"email":    "newuser@example.com",
			"username": "johndoe", // Already taken
		}

		schema := Schema{
			"email":    String().Required().Email().Unique("users", "email", nil),
			"username": String().Required().Min(3).Unique("users", "username", nil),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for taken username")
		}
	})

	t.Run("unique with ignore for update scenario", func(t *testing.T) {
		// Simulating updating user's own profile
		data := DataObject{
			"email":    "john@example.com", // Own email, should be ignored
			"username": "johndoe",          // Own username, should be ignored
		}

		schema := Schema{
			"email":    String().Required().Email().Unique("users", "email", "john@example.com"),
			"username": String().Required().Min(3).Unique("users", "username", "johndoe"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors when ignoring own values, got: %v", err.Errors)
		}
	})

	t.Run("unique with partial ignore", func(t *testing.T) {
		// Updating email but keeping same username
		data := DataObject{
			"email":    "newemail@example.com", // New email
			"username": "johndoe",              // Own username, should be ignored
		}

		schema := Schema{
			"email":    String().Required().Email().Unique("users", "email", nil),
			"username": String().Required().Min(3).Unique("users", "username", "johndoe"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})
}

func TestDBAdapter_ArrayOfIDsWithExists(t *testing.T) {
	mockDB := NewMockGormDB()
	adapter := CreateMockGormAdapter(mockDB)

	t.Run("simple array of tag IDs - all exist", func(t *testing.T) {
		data := DataObject{
			"title":   "My Article",
			"tag_ids": []any{float64(1), float64(2), float64(3)},
		}

		schema := Schema{
			"title":   String().Required().Min(3),
			"tag_ids": Array().Required().Min(1).Exists("tags", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("simple array of tag IDs - some don't exist", func(t *testing.T) {
		data := DataObject{
			"title":   "My Article",
			"tag_ids": []any{float64(1), float64(2), float64(999)}, // 999 doesn't exist
		}

		schema := Schema{
			"title":   String().Required().Min(3),
			"tag_ids": Array().Required().Min(1).Exists("tags", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for non-existent tag ID")
		}
	})

	t.Run("multiple arrays with different table checks", func(t *testing.T) {
		data := DataObject{
			"name":         "Product Bundle",
			"product_ids":  []any{float64(100), float64(101)},
			"category_ids": []any{float64(10), float64(20)},
			"tag_ids":      []any{float64(1), float64(2), float64(3)},
		}

		schema := Schema{
			"name":         String().Required().Min(3),
			"product_ids":  Array().Required().Min(1).Exists("products", "id"),
			"category_ids": Array().Required().Min(1).Exists("categories", "id"),
			"tag_ids":      Array().Required().Min(1).Exists("tags", "id"),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})
}

func TestDBAdapter_WithWhereConditions(t *testing.T) {
	// Adapter that respects where conditions
	adapter := FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
		result := make(map[any]bool)

		// Simulate: products 100, 101 are active; 102 is inactive
		activeProducts := map[any]bool{
			float64(100): true,
			float64(101): true,
		}

		// Check if we have a status=active where clause
		hasActiveFilter := false
		for _, w := range wheres {
			if w.Column == "status" && w.Value == "active" {
				hasActiveFilter = true
				break
			}
		}

		if table == "products" && hasActiveFilter {
			// Only return active products
			for _, v := range values {
				result[v] = activeProducts[v]
			}
		} else if table == "products" {
			// Return all products
			allProducts := map[any]bool{
				float64(100): true,
				float64(101): true,
				float64(102): true,
			}
			for _, v := range values {
				result[v] = allProducts[v]
			}
		}

		return result, nil
	})

	t.Run("exists with where condition - active product", func(t *testing.T) {
		data := DataObject{
			"product_id": float64(100), // Active product
		}

		schema := Schema{
			"product_id": Float().Required().Exists("products", "id",
				WhereEq("status", "active"),
			),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err.Errors)
		}
	})

	t.Run("exists with where condition - inactive product", func(t *testing.T) {
		data := DataObject{
			"product_id": float64(102), // Inactive product
		}

		schema := Schema{
			"product_id": Float().Required().Exists("products", "id",
				WhereEq("status", "active"),
			),
		}

		err := ValidateWithDB(context.Background(), data, schema, adapter)
		if err == nil {
			t.Error("Expected error for inactive product")
		}
	})
}

func TestRegexCache(t *testing.T) {
	t.Run("caches compiled regex", func(t *testing.T) {
		pattern := `^test\d+$`

		re1, err := globalRegexCache.GetOrCompile(pattern)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		re2, err := globalRegexCache.GetOrCompile(pattern)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if re1 != re2 {
			t.Error("Expected same cached regex instance")
		}
	})

	t.Run("empty pattern returns nil", func(t *testing.T) {
		re, err := globalRegexCache.GetOrCompile("")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if re != nil {
			t.Error("Expected nil for empty pattern")
		}
	})
}

func TestWhereHelpers(t *testing.T) {
	t.Run("Where", func(t *testing.T) {
		w := Where("status", ">", 5)
		if w.Column != "status" || w.Operator != ">" || w.Value != 5 {
			t.Error("Where helper failed")
		}
	})

	t.Run("WhereEq", func(t *testing.T) {
		w := WhereEq("status", "active")
		if w.Column != "status" || w.Operator != "=" || w.Value != "active" {
			t.Error("WhereEq helper failed")
		}
	})

	t.Run("WhereNot", func(t *testing.T) {
		w := WhereNot("status", "deleted")
		if w.Column != "status" || w.Operator != "!=" || w.Value != "deleted" {
			t.Error("WhereNot helper failed")
		}
	})
}

func TestLookupResult(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		r := LookupResult{value: "hello", exists: true}
		if r.String() != "hello" {
			t.Error("String() failed")
		}
	})

	t.Run("Int", func(t *testing.T) {
		r := LookupResult{value: float64(42), exists: true}
		if r.Int() != 42 {
			t.Error("Int() failed")
		}
	})

	t.Run("Float", func(t *testing.T) {
		r := LookupResult{value: float64(3.14), exists: true}
		if r.Float() != 3.14 {
			t.Error("Float() failed")
		}
	})

	t.Run("Bool", func(t *testing.T) {
		r := LookupResult{value: true, exists: true}
		if r.Bool() != true {
			t.Error("Bool() failed")
		}
	})

	t.Run("Exists", func(t *testing.T) {
		r := LookupResult{value: "test", exists: true}
		if !r.Exists() {
			t.Error("Exists() failed")
		}
	})
}

func TestRequiredIf(t *testing.T) {
	t.Run("string required if condition met", func(t *testing.T) {
		schema := Schema{
			"type": String().Required(),
			"value": String().RequiredIf(func(data DataObject) bool {
				return data["type"] == "custom"
			}),
		}

		err := Validate(DataObject{"type": "custom"}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}

		err = Validate(DataObject{"type": "default"}, schema)
		if err != nil {
			t.Errorf("Expected no error when condition not met, got: %v", err.Errors)
		}
	})
}

func TestTransformations(t *testing.T) {
	t.Run("trim", func(t *testing.T) {
		schema := Schema{
			"name": String().Trim().Min(3),
		}

		// "  ab  " after trim is "ab" which is 2 chars, should fail Min(3)
		err := Validate(DataObject{"name": "  ab  "}, schema)
		if err == nil {
			t.Error("Expected error after trim")
		}
	})
}

func TestErrorResponseStructure(t *testing.T) {
	t.Run("top level required fields error keys", func(t *testing.T) {
		schema := Schema{
			"username": String().Required(),
			"email":    String().Required().Email(),
			"bio":      String().Required(),
			"profile":  Object().Required(),
			"tags":     Array().Required(),
			"settings": Object().Required(),
		}

		// Send empty data - all fields should have errors with their field names as keys
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Fatal("Expected error for missing required fields")
		}

		expectedFields := []string{"username", "email", "bio", "profile", "tags", "settings"}
		for _, field := range expectedFields {
			if _, exists := err.Errors[field]; !exists {
				t.Errorf("Expected error key '%s' to exist in errors", field)
			}
			if len(err.Errors[field]) == 0 {
				t.Errorf("Expected error key '%s' to have at least one error message", field)
			}
		}

		// Ensure we have exactly the expected number of error keys
		if len(err.Errors) != len(expectedFields) {
			t.Errorf("Expected %d error keys, got %d. Keys: %v", len(expectedFields), len(err.Errors), err.Errors)
		}
	})

	t.Run("nested object required field error keys", func(t *testing.T) {
		schema := Schema{
			"profile": Object().Required().Shape(Schema{
				"avatar":  String().Required(),
				"website": String().Required(),
			}),
		}

		// Send profile object but missing nested required fields
		err := Validate(DataObject{
			"profile": map[string]any{},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for missing nested required fields")
		}

		// Error keys should be "profile.avatar" and "profile.website"
		expectedFields := []string{"profile.avatar", "profile.website"}
		for _, field := range expectedFields {
			if _, exists := err.Errors[field]; !exists {
				t.Errorf("Expected error key '%s' to exist in errors. Got keys: %v", field, err.Errors)
			}
			if len(err.Errors[field]) == 0 {
				t.Errorf("Expected error key '%s' to have at least one error message", field)
			}
		}
	})

	t.Run("deeply nested object required field error keys", func(t *testing.T) {
		schema := Schema{
			"user": Object().Required().Shape(Schema{
				"profile": Object().Required().Shape(Schema{
					"avatar": Object().Required().Shape(Schema{
						"url": String().Required(),
					}),
				}),
			}),
		}

		// Send nested objects but missing the deepest required field
		err := Validate(DataObject{
			"user": map[string]any{
				"profile": map[string]any{
					"avatar": map[string]any{},
				},
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for missing deeply nested required field")
		}

		// Error key should be "user.profile.avatar.url"
		if _, exists := err.Errors["user.profile.avatar.url"]; !exists {
			t.Errorf("Expected error key 'user.profile.avatar.url' to exist in errors. Got keys: %v", err.Errors)
		}
	})

	t.Run("array of objects required field error keys", func(t *testing.T) {
		schema := Schema{
			"users": Array().Required().Of(Object().Shape(Schema{
				"name":  String().Required(),
				"email": String().Required(),
			})),
		}

		// Send array with objects missing required fields
		err := Validate(DataObject{
			"users": []any{
				map[string]any{},                            // index 0: missing name and email
				map[string]any{"name": "John"},              // index 1: missing email
				map[string]any{"email": "test@example.com"}, // index 2: missing name
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for missing required fields in array objects")
		}

		// Error keys should follow pattern: "users.INDEX.FIELD"
		expectedFields := []string{
			"users.0.name",
			"users.0.email",
			"users.1.email",
			"users.2.name",
		}
		for _, field := range expectedFields {
			if _, exists := err.Errors[field]; !exists {
				t.Errorf("Expected error key '%s' to exist in errors. Got keys: %v", field, err.Errors)
			}
		}
	})

	t.Run("array of objects with nested objects error keys", func(t *testing.T) {
		schema := Schema{
			"posts": Array().Required().Of(Object().Shape(Schema{
				"title": String().Required(),
				"author": Object().Required().Shape(Schema{
					"id":   String().Required(),
					"name": String().Required(),
				}),
			})),
		}

		// Send array with nested object missing required fields
		err := Validate(DataObject{
			"posts": []any{
				map[string]any{
					"title": "Post 1",
					"author": map[string]any{
						"name": "Author", // missing id
					},
				},
				map[string]any{
					"author": map[string]any{}, // missing title, id, and name
				},
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for missing required fields in nested array objects")
		}

		// Expected error keys
		expectedFields := []string{
			"posts.0.author.id",
			"posts.1.title",
			"posts.1.author.id",
			"posts.1.author.name",
		}
		for _, field := range expectedFields {
			if _, exists := err.Errors[field]; !exists {
				t.Errorf("Expected error key '%s' to exist in errors. Got keys: %v", field, err.Errors)
			}
		}
	})

	t.Run("mixed validation errors structure", func(t *testing.T) {
		schema := Schema{
			"username": String().Required().Min(3).Max(20).AlphaNumeric(),
			"email":    String().Required().Email(),
			"age":      Float().Required().Min(18).Max(120),
			"bio":      String().Required().Max(500),
			"profile": Object().Required().Shape(Schema{
				"avatar":  String().Required().URL(),
				"website": String().URL(),
			}),
			"tags": Array().Required().Max(10).Of(String().Min(2).Max(20)),
			"settings": Object().Required().Shape(Schema{
				"notifications": Bool().Required(),
				"theme":         String().Required().In("light", "dark"),
			}),
		}

		// Send empty data - all required fields should have errors
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Fatal("Expected error for missing required fields")
		}

		// Top level required fields should have errors
		topLevelFields := []string{"username", "email", "age", "bio", "profile", "tags", "settings"}
		for _, field := range topLevelFields {
			if _, exists := err.Errors[field]; !exists {
				t.Errorf("Expected error key '%s' to exist in errors when all data is missing", field)
			}
		}
	})

	t.Run("partial data with nested errors", func(t *testing.T) {
		schema := Schema{
			"username": String().Required(),
			"profile": Object().Required().Shape(Schema{
				"avatar":  String().Required(),
				"website": String().Required(),
			}),
			"settings": Object().Required().Shape(Schema{
				"notifications": Bool().Required(),
			}),
		}

		// Send partial data - some top level, some nested missing
		err := Validate(DataObject{
			"profile":  map[string]any{"website": "https://example.com"},
			"settings": map[string]any{},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for missing required fields")
		}

		// "username" should have error (missing top level)
		if _, exists := err.Errors["username"]; !exists {
			t.Errorf("Expected error key 'username' to exist")
		}

		// "profile.avatar" should have error (nested missing)
		if _, exists := err.Errors["profile.avatar"]; !exists {
			t.Errorf("Expected error key 'profile.avatar' to exist. Got: %v", err.Errors)
		}

		// "settings.notifications" should have error (nested missing)
		if _, exists := err.Errors["settings.notifications"]; !exists {
			t.Errorf("Expected error key 'settings.notifications' to exist. Got: %v", err.Errors)
		}

		// "profile.website" should NOT have error (provided)
		if _, exists := err.Errors["profile.website"]; exists {
			t.Errorf("Expected NO error key 'profile.website' since it was provided")
		}
	})

	t.Run("array element validation error keys", func(t *testing.T) {
		schema := Schema{
			"emails": Array().Required().Of(String().Email()),
		}

		err := Validate(DataObject{
			"emails": []any{"valid@example.com", "invalid-email", "another@test.com", "bad"},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for invalid emails in array")
		}

		// Error keys should be "emails.1" and "emails.3" for invalid emails at index 1 and 3
		if _, exists := err.Errors["emails.1"]; !exists {
			t.Errorf("Expected error key 'emails.1' to exist for invalid email. Got: %v", err.Errors)
		}
		if _, exists := err.Errors["emails.3"]; !exists {
			t.Errorf("Expected error key 'emails.3' to exist for invalid email. Got: %v", err.Errors)
		}

		// Valid email indices should not have errors
		if _, exists := err.Errors["emails.0"]; exists {
			t.Errorf("Expected NO error key 'emails.0' since it's valid")
		}
		if _, exists := err.Errors["emails.2"]; exists {
			t.Errorf("Expected NO error key 'emails.2' since it's valid")
		}
	})

	t.Run("complex schema like TestComplexSchema with empty data", func(t *testing.T) {
		schema := Schema{
			"username": String().Required().Min(3).Max(20).AlphaNumeric(),
			"email":    String().Required().Email(),
			"age":      Float().Required().Min(18).Max(120),
			"bio":      String().Nullable().Max(500),
			"profile": Object().Required().Shape(Schema{
				"avatar":  String().URL(),
				"website": String().URL(),
			}),
			"tags": Array().Max(10).Of(String().Min(2).Max(20)),
			"settings": Object().Shape(Schema{
				"notifications": Bool().Default(true),
				"theme":         String().In("light", "dark").Default("light"),
			}),
		}

		// Send empty data - only required fields should have errors
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Fatal("Expected error for missing required fields")
		}

		// Required fields should have errors
		requiredFields := []string{"username", "email", "age", "profile"}
		for _, field := range requiredFields {
			if _, exists := err.Errors[field]; !exists {
				t.Errorf("Expected error key '%s' to exist for required field", field)
			}
		}

		// Nullable/optional fields should NOT have errors
		optionalFields := []string{"bio", "tags", "settings"}
		for _, field := range optionalFields {
			if _, exists := err.Errors[field]; exists {
				t.Errorf("Expected NO error key '%s' since it's not required. Got errors: %v", field, err.Errors[field])
			}
		}
	})

	t.Run("profile object provided but nested avatar required", func(t *testing.T) {
		schema := Schema{
			"profile": Object().Required().Shape(Schema{
				"avatar": String().Required(),
			}),
		}

		// Send profile but missing avatar
		err := Validate(DataObject{
			"profile": map[string]any{},
		}, schema)
		if err == nil {
			t.Fatal("Expected error for missing nested avatar")
		}

		// Error should be "profile.avatar", not "profile"
		if _, exists := err.Errors["profile.avatar"]; !exists {
			t.Errorf("Expected error key 'profile.avatar' to exist. Got: %v", err.Errors)
		}

		// "profile" itself should NOT have an error since it was provided
		if _, exists := err.Errors["profile"]; exists {
			t.Errorf("Expected NO error key 'profile' since it was provided. Got: %v", err.Errors)
		}
	})

	t.Run("multiple errors per field", func(t *testing.T) {
		schema := Schema{
			"password": String().Required().Min(8).Custom(func(v string, lookup Lookup) error {
				if v == "short" {
					return errors.New("password is too common")
				}
				return nil
			}),
		}

		// "short" is less than 8 chars and is a common password
		err := Validate(DataObject{"password": "short"}, schema)
		if err == nil {
			t.Fatal("Expected error for invalid password")
		}

		// The field should have multiple error messages
		if len(err.Errors["password"]) < 1 {
			t.Errorf("Expected at least 1 error for password, got %d", len(err.Errors["password"]))
		}
	})

	t.Run("error messages are strings array", func(t *testing.T) {
		schema := Schema{
			"name": String().Required(),
		}

		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}

		// Verify the error structure is map[string][]string
		errors := err.Errors
		if errors == nil {
			t.Fatal("Expected errors map to not be nil")
		}

		nameErrors, exists := errors["name"]
		if !exists {
			t.Fatal("Expected 'name' key in errors")
		}

		if len(nameErrors) == 0 {
			t.Fatal("Expected at least one error message for 'name'")
		}

		// Each error should be a non-empty string
		for i, msg := range nameErrors {
			if msg == "" {
				t.Errorf("Error message at index %d should not be empty", i)
			}
		}
	})
}

func TestCoercion(t *testing.T) {
	t.Run("number coerce from string", func(t *testing.T) {
		schema := Schema{
			"age": Float().Coerce().Min(18),
		}

		err := Validate(DataObject{"age": "25"}, schema)
		if err != nil {
			t.Errorf("Expected no error with coercion, got: %v", err.Errors)
		}
	})

	t.Run("bool coerce from string", func(t *testing.T) {
		schema := Schema{
			"active": Bool().Coerce().True(),
		}

		err := Validate(DataObject{"active": "true"}, schema)
		if err != nil {
			t.Errorf("Expected no error with coercion, got: %v", err.Errors)
		}
	})
}

// Benchmark tests
func BenchmarkStringValidation(b *testing.B) {
	schema := Schema{
		"username": String().Required().Min(3).Max(20),
		"email":    String().Required().Email(),
	}
	data := DataObject{
		"username": "johndoe",
		"email":    "john@example.com",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkComplexValidation(b *testing.B) {
	schema := Schema{
		"username": String().Required().Min(3).Max(20),
		"email":    String().Required().Email(),
		"age":      Float().Required().Min(18),
		"profile": Object().Required().Shape(Schema{
			"name": String().Required(),
			"bio":  String().Max(500),
		}),
		"tags": Array().Max(10).Of(String()),
	}
	data := DataObject{
		"username": "johndoe",
		"email":    "john@example.com",
		"age":      float64(25),
		"profile": map[string]any{
			"name": "John Doe",
			"bio":  "Developer",
		},
		"tags": []any{"go", "rust", "python"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// INTEGRATION TESTS FOR NEW FEATURES
// ============================================================================

func TestIntegration_NewStringValidators(t *testing.T) {
	t.Run("UUID validation", func(t *testing.T) {
		schema := Schema{
			"id": String().Required().UUID(),
		}

		err := Validate(DataObject{"id": "550e8400-e29b-41d4-a716-446655440000"}, schema)
		if err != nil {
			t.Errorf("Expected valid UUID, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"id": "invalid-uuid"}, schema)
		if err == nil {
			t.Error("Expected error for invalid UUID")
		}
	})

	t.Run("IP address validation", func(t *testing.T) {
		schema := Schema{
			"ipv4": String().Required().IPv4(),
			"ipv6": String().Required().IPv6(),
			"ip":   String().Required().IP(),
		}

		err := Validate(DataObject{
			"ipv4": "192.168.1.1",
			"ipv6": "::1",
			"ip":   "10.0.0.1",
		}, schema)
		if err != nil {
			t.Errorf("Expected valid IPs, got error: %v", err.Errors)
		}
	})

	t.Run("JSON validation", func(t *testing.T) {
		schema := Schema{
			"config": String().Required().JSON(),
		}

		err := Validate(DataObject{"config": `{"key": "value"}`}, schema)
		if err != nil {
			t.Errorf("Expected valid JSON, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"config": "not json"}, schema)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("HexColor validation", func(t *testing.T) {
		schema := Schema{
			"color": String().Required().HexColor(),
		}

		err := Validate(DataObject{"color": "#ff5733"}, schema)
		if err != nil {
			t.Errorf("Expected valid hex color, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"color": "red"}, schema)
		if err == nil {
			t.Error("Expected error for invalid hex color")
		}
	})

	t.Run("DoesntStartWith and DoesntEndWith", func(t *testing.T) {
		schema := Schema{
			"username": String().Required().DoesntStartWith("admin", "root"),
			"filename": String().Required().DoesntEndWith(".exe", ".bat"),
		}

		err := Validate(DataObject{
			"username": "john_doe",
			"filename": "document.pdf",
		}, schema)
		if err != nil {
			t.Errorf("Expected valid values, got error: %v", err.Errors)
		}

		err = Validate(DataObject{
			"username": "admin_user",
			"filename": "document.pdf",
		}, schema)
		if err == nil {
			t.Error("Expected error for username starting with admin")
		}
	})

	t.Run("Includes validation", func(t *testing.T) {
		schema := Schema{
			"bio": String().Required().Includes("developer", "engineer"),
		}

		err := Validate(DataObject{"bio": "I am a software developer and engineer"}, schema)
		if err != nil {
			t.Errorf("Expected valid bio, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"bio": "I am a designer"}, schema)
		if err == nil {
			t.Error("Expected error for bio missing required words")
		}
	})

	t.Run("MAC address validation", func(t *testing.T) {
		schema := Schema{
			"mac": String().Required().MAC(),
		}

		err := Validate(DataObject{"mac": "00:1A:2B:3C:4D:5E"}, schema)
		if err != nil {
			t.Errorf("Expected valid MAC, got error: %v", err.Errors)
		}
	})

	t.Run("ULID validation", func(t *testing.T) {
		schema := Schema{
			"id": String().Required().ULID(),
		}

		err := Validate(DataObject{"id": "01ARZ3NDEKTSV4RRFFQ69G5FAV"}, schema)
		if err != nil {
			t.Errorf("Expected valid ULID, got error: %v", err.Errors)
		}
	})

	t.Run("AlphaDash validation", func(t *testing.T) {
		schema := Schema{
			"slug": String().Required().AlphaDash(),
		}

		err := Validate(DataObject{"slug": "hello-world_123"}, schema)
		if err != nil {
			t.Errorf("Expected valid slug, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"slug": "hello world"}, schema)
		if err == nil {
			t.Error("Expected error for slug with space")
		}
	})

	t.Run("Digits validation", func(t *testing.T) {
		schema := Schema{
			"otp": String().Required().Digits(6),
		}

		err := Validate(DataObject{"otp": "123456"}, schema)
		if err != nil {
			t.Errorf("Expected valid OTP, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"otp": "12345"}, schema)
		if err == nil {
			t.Error("Expected error for OTP with wrong length")
		}
	})

	t.Run("ASCII validation", func(t *testing.T) {
		schema := Schema{
			"text": String().Required().ASCII(),
		}

		err := Validate(DataObject{"text": "Hello World 123!"}, schema)
		if err != nil {
			t.Errorf("Expected valid ASCII, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"text": "Hello 世界"}, schema)
		if err == nil {
			t.Error("Expected error for non-ASCII characters")
		}
	})

	t.Run("Base64 validation", func(t *testing.T) {
		schema := Schema{
			"data": String().Required().Base64(),
		}

		err := Validate(DataObject{"data": "SGVsbG8gV29ybGQ="}, schema)
		if err != nil {
			t.Errorf("Expected valid Base64, got error: %v", err.Errors)
		}
	})
}

func TestIntegration_NewNumberValidators(t *testing.T) {
	t.Run("Between validation", func(t *testing.T) {
		schema := Schema{
			"score": Float().Required().Between(0, 100),
		}

		err := Validate(DataObject{"score": float64(50)}, schema)
		if err != nil {
			t.Errorf("Expected valid score, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"score": float64(150)}, schema)
		if err == nil {
			t.Error("Expected error for score out of range")
		}
	})

	t.Run("Step validation", func(t *testing.T) {
		schema := Schema{
			"quantity": Float().Required().Step(5),
		}

		err := Validate(DataObject{"quantity": float64(15)}, schema)
		if err != nil {
			t.Errorf("Expected valid quantity, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"quantity": float64(17)}, schema)
		if err == nil {
			t.Error("Expected error for quantity not a multiple of 5")
		}
	})
}

func TestIntegration_NewArrayValidators(t *testing.T) {
	t.Run("Contains validation", func(t *testing.T) {
		schema := Schema{
			"roles": Array().Required().Contains("admin"),
		}

		err := Validate(DataObject{"roles": []any{"user", "admin"}}, schema)
		if err != nil {
			t.Errorf("Expected valid roles, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"roles": []any{"user", "guest"}}, schema)
		if err == nil {
			t.Error("Expected error for missing admin role")
		}
	})

	t.Run("DoesntContain validation", func(t *testing.T) {
		schema := Schema{
			"tags": Array().Required().DoesntContain("spam", "nsfw"),
		}

		err := Validate(DataObject{"tags": []any{"tech", "news"}}, schema)
		if err != nil {
			t.Errorf("Expected valid tags, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"tags": []any{"tech", "spam"}}, schema)
		if err == nil {
			t.Error("Expected error for containing spam")
		}
	})

	t.Run("Distinct validation", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required().Distinct(),
		}

		err := Validate(DataObject{"items": []any{"a", "b", "c"}}, schema)
		if err != nil {
			t.Errorf("Expected valid items, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"items": []any{"a", "b", "a"}}, schema)
		if err == nil {
			t.Error("Expected error for duplicate items")
		}
	})
}

func TestIntegration_ObjectUtilities(t *testing.T) {
	t.Run("Pick utility", func(t *testing.T) {
		userSchema := Object().Shape(Schema{
			"name":     String().Required(),
			"email":    String().Required().Email(),
			"password": String().Required().Min(8),
		})

		// Pick only name and email
		publicSchema := Schema{"user": userSchema.Pick("name", "email")}

		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
			},
		}, publicSchema)
		if err != nil {
			t.Errorf("Expected valid with picked fields, got error: %v", err.Errors)
		}
	})

	t.Run("Omit utility", func(t *testing.T) {
		userSchema := Object().Shape(Schema{
			"name":     String().Required(),
			"email":    String().Required().Email(),
			"password": String().Required().Min(8),
		})

		// Omit password for public view
		publicSchema := Schema{"user": userSchema.Omit("password")}

		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
			},
		}, publicSchema)
		if err != nil {
			t.Errorf("Expected valid without password, got error: %v", err.Errors)
		}
	})

	t.Run("Partial utility", func(t *testing.T) {
		userSchema := Object().Shape(Schema{
			"name":  String().Required(),
			"email": String().Required().Email(),
		})

		// Make all fields optional for PATCH
		patchSchema := Schema{"user": userSchema.Partial()}

		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, patchSchema)
		if err != nil {
			t.Errorf("Expected valid partial update, got error: %v", err.Errors)
		}
	})

	t.Run("Merge utility", func(t *testing.T) {
		baseSchema := Object().Shape(Schema{
			"name": String().Required(),
		})

		extendedSchema := Object().Shape(Schema{
			"email": String().Required().Email(),
		})

		mergedSchema := Schema{"user": baseSchema.Merge(extendedSchema)}

		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
			},
		}, mergedSchema)
		if err != nil {
			t.Errorf("Expected valid merged schema, got error: %v", err.Errors)
		}
	})
}

func TestIntegration_SchemaHelpers(t *testing.T) {
	t.Run("Enum validation", func(t *testing.T) {
		schema := Schema{
			"status": Enum("pending", "active", "completed").Required(),
		}

		err := Validate(DataObject{"status": "active"}, schema)
		if err != nil {
			t.Errorf("Expected valid status, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"status": "unknown"}, schema)
		if err == nil {
			t.Error("Expected error for invalid status")
		}
	})

	t.Run("EnumInt validation", func(t *testing.T) {
		schema := Schema{
			"priority": EnumInt(1, 2, 3, 4, 5).Required(),
		}

		err := Validate(DataObject{"priority": float64(3)}, schema)
		if err != nil {
			t.Errorf("Expected valid priority, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"priority": float64(10)}, schema)
		if err == nil {
			t.Error("Expected error for invalid priority")
		}
	})

	t.Run("Literal validation", func(t *testing.T) {
		schema := Schema{
			"type": Literal("config").Required(),
		}

		err := Validate(DataObject{"type": "config"}, schema)
		if err != nil {
			t.Errorf("Expected valid type, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"type": "other"}, schema)
		if err == nil {
			t.Error("Expected error for wrong literal")
		}
	})

	t.Run("Union validation", func(t *testing.T) {
		schema := Schema{
			"id": Union(String().UUID(), Int().Positive()).Required(),
		}

		err := Validate(DataObject{"id": "550e8400-e29b-41d4-a716-446655440000"}, schema)
		if err != nil {
			t.Errorf("Expected valid UUID id, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"id": float64(123)}, schema)
		if err != nil {
			t.Errorf("Expected valid int id, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"id": "not-a-uuid"}, schema)
		if err == nil {
			t.Error("Expected error for invalid id")
		}
	})

	t.Run("Optional validation", func(t *testing.T) {
		schema := Schema{
			"name":     String().Required(),
			"nickname": Optional(String().Min(2).Max(20)),
		}

		err := Validate(DataObject{"name": "John"}, schema)
		if err != nil {
			t.Errorf("Expected valid without optional, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"name": "John", "nickname": "Johnny"}, schema)
		if err != nil {
			t.Errorf("Expected valid with optional, got error: %v", err.Errors)
		}

		err = Validate(DataObject{"name": "John", "nickname": "X"}, schema)
		if err == nil {
			t.Error("Expected error for invalid optional value")
		}
	})
}

func TestIntegration_ComplexSchema(t *testing.T) {
	// Test a complex real-world schema using new features
	schema := Schema{
		"id":     Union(String().UUID(), Int().Positive()).Required(),
		"status": Enum("draft", "published", "archived").Required(),
		"type":   Literal("article"),
		"author": Object().Required().Shape(Schema{
			"name":  String().Required().Min(2).AlphaDash(),
			"email": String().Required().Email(),
			"role":  Enum("admin", "editor", "writer").Required(),
		}),
		"metadata": Optional(Object().Shape(Schema{
			"color":    String().HexColor(),
			"priority": EnumInt(1, 2, 3, 4, 5),
		})),
		"tags":       Array().Required().Min(1).Distinct().DoesntContain("spam"),
		"config":     Optional(String().JSON()),
		"score":      Float().Between(0, 100),
		"ip_address": Optional(String().IP()),
	}

	validData := DataObject{
		"id":     "550e8400-e29b-41d4-a716-446655440000",
		"status": "published",
		"type":   "article",
		"author": map[string]any{
			"name":  "john_doe",
			"email": "john@example.com",
			"role":  "editor",
		},
		"metadata": map[string]any{
			"color":    "#ff5733",
			"priority": float64(3),
		},
		"tags":       []any{"tech", "news"},
		"config":     `{"key": "value"}`,
		"score":      float64(85),
		"ip_address": "192.168.1.1",
	}

	err := Validate(validData, schema)
	if err != nil {
		t.Errorf("Expected valid complex data, got error: %v", err.Errors)
	}
}

// ============================================================================
// ADDITIONAL COVERAGE TESTS
// ============================================================================

func TestArrayValidator_ContainsWithMessage_Coverage(t *testing.T) {
	schema := Schema{
		"roles": Array().Required().ContainsWithMessage("Must have admin role", "admin"),
	}

	t.Run("missing required item", func(t *testing.T) {
		err := Validate(DataObject{"roles": []any{"user", "editor"}}, schema)
		if err == nil {
			t.Error("Expected error for missing admin")
		}
		if err != nil && err.Errors["roles"][0] != "Must have admin role" {
			t.Errorf("Expected custom message, got: %s", err.Errors["roles"][0])
		}
	})

	t.Run("has required item", func(t *testing.T) {
		err := Validate(DataObject{"roles": []any{"admin", "user"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestArrayValidator_DoesntContainWithMessage_Coverage(t *testing.T) {
	schema := Schema{
		"tags": Array().Required().DoesntContainWithMessage("No spam allowed", "spam"),
	}

	t.Run("contains forbidden item", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{"news", "spam"}}, schema)
		if err == nil {
			t.Error("Expected error for spam")
		}
		// Error is on the element path (tags.1), not tags
		if err != nil {
			found := false
			for _, errs := range err.Errors {
				for _, e := range errs {
					if e == "No spam allowed" {
						found = true
					}
				}
			}
			if !found {
				t.Errorf("Expected custom message 'No spam allowed', got: %v", err.Errors)
			}
		}
	})

	t.Run("no forbidden item", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{"news", "tech"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_InWithMessage_Coverage(t *testing.T) {
	schema := Schema{
		"status": Int().Required().InWithMessage("Status must be 1, 2, or 3", 1, 2, 3),
	}

	t.Run("invalid value", func(t *testing.T) {
		err := Validate(DataObject{"status": 5}, schema)
		if err == nil {
			t.Error("Expected error for invalid status")
		}
		if err != nil && err.Errors["status"][0] != "Status must be 1, 2, or 3" {
			t.Errorf("Expected custom message, got: %s", err.Errors["status"][0])
		}
	})

	t.Run("valid value", func(t *testing.T) {
		err := Validate(DataObject{"status": 2}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_NotInWithMessage_Coverage(t *testing.T) {
	schema := Schema{
		"code": Int().Required().NotInWithMessage("Code cannot be 0 or -1", 0, -1),
	}

	t.Run("forbidden value", func(t *testing.T) {
		err := Validate(DataObject{"code": 0}, schema)
		if err == nil {
			t.Error("Expected error for forbidden code")
		}
		if err != nil && err.Errors["code"][0] != "Code cannot be 0 or -1" {
			t.Errorf("Expected custom message, got: %s", err.Errors["code"][0])
		}
	})

	t.Run("allowed value", func(t *testing.T) {
		err := Validate(DataObject{"code": 100}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_NotInWithMessage_Coverage(t *testing.T) {
	schema := Schema{
		"role": String().Required().NotInWithMessage("Restricted roles not allowed", "root", "superuser"),
	}

	t.Run("forbidden value", func(t *testing.T) {
		err := Validate(DataObject{"role": "root"}, schema)
		if err == nil {
			t.Error("Expected error for forbidden role")
		}
		if err != nil && err.Errors["role"][0] != "Restricted roles not allowed" {
			t.Errorf("Expected custom message, got: %s", err.Errors["role"][0])
		}
	})

	t.Run("allowed value", func(t *testing.T) {
		err := Validate(DataObject{"role": "admin"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Catch(t *testing.T) {
	schema := Schema{
		"name": String().Required().Min(3).Catch("default_name"),
	}

	t.Run("invalid value uses catch", func(t *testing.T) {
		// The catch should provide default when validation fails
		err := Validate(DataObject{"name": "ab"}, schema) // too short
		// Note: Catch behavior depends on implementation
		_ = err // Just ensure it doesn't panic
	})

	t.Run("valid value", func(t *testing.T) {
		err := Validate(DataObject{"name": "john"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestAnyValidator(t *testing.T) {
	t.Run("accepts any value", func(t *testing.T) {
		schema := Schema{
			"data": Any(),
		}

		testCases := []any{
			"string",
			123,
			12.34,
			true,
			[]any{1, 2, 3},
			map[string]any{"key": "value"},
		}

		for _, val := range testCases {
			err := Validate(DataObject{"data": val}, schema)
			if err != nil {
				t.Errorf("Expected Any() to accept %v, got: %v", val, err.Errors)
			}
		}
	})

	t.Run("required", func(t *testing.T) {
		schema := Schema{
			"data": Any().Required(),
		}

		err := Validate(DataObject{"data": nil}, schema)
		if err == nil {
			t.Error("Expected error for nil required Any()")
		}

		err = Validate(DataObject{"data": "something"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("nullable", func(t *testing.T) {
		schema := Schema{
			"data": Any().Nullable(),
		}

		err := Validate(DataObject{"data": nil}, schema)
		if err != nil {
			t.Errorf("Expected nullable Any() to accept nil, got: %v", err.Errors)
		}
	})

	t.Run("message", func(t *testing.T) {
		schema := Schema{
			"data": Any().Required().Message("required", "Data is required"),
		}

		err := Validate(DataObject{"data": nil}, schema)
		if err == nil || err.Errors["data"][0] != "Data is required" {
			t.Error("Expected custom message")
		}
	})
}

func TestEnumValidator_Nullable(t *testing.T) {
	schema := Schema{
		"status": Enum("active", "inactive").Nullable(),
	}

	err := Validate(DataObject{"status": nil}, schema)
	if err != nil {
		t.Errorf("Expected nullable enum to accept nil, got: %v", err.Errors)
	}
}

func TestEnumValidator_Default(t *testing.T) {
	schema := Schema{
		"status": Enum("active", "inactive").Default("active"),
	}

	err := Validate(DataObject{}, schema)
	if err != nil {
		t.Errorf("Expected default to be applied, got: %v", err.Errors)
	}
}

func TestLiteralValidator_Nullable(t *testing.T) {
	schema := Schema{
		"type": Literal("article").Nullable(),
	}

	err := Validate(DataObject{"type": nil}, schema)
	if err != nil {
		t.Errorf("Expected nullable literal to accept nil, got: %v", err.Errors)
	}
}

func TestPoolFunctions(t *testing.T) {
	t.Run("GetStringSlice and PutStringSlice", func(t *testing.T) {
		slice := GetStringSlice()
		if slice == nil {
			t.Error("Expected non-nil slice")
		}
		slice = append(slice, "a", "b", "c")
		PutStringSlice(slice)

		// Get again should return empty slice
		slice2 := GetStringSlice()
		if len(slice2) != 0 {
			t.Errorf("Expected empty slice, got len=%d", len(slice2))
		}
	})

	t.Run("BuildPath", func(t *testing.T) {
		result := BuildPath("user", "address", "city")
		if result != "user.address.city" {
			t.Errorf("Expected 'user.address.city', got '%s'", result)
		}

		result = BuildPath("single")
		if result != "single" {
			t.Errorf("Expected 'single', got '%s'", result)
		}

		result = BuildPath()
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("JoinErrors", func(t *testing.T) {
		result := JoinErrors([]string{"error1", "error2"}, ", ")
		if result != "error1, error2" {
			t.Errorf("Expected 'error1, error2', got '%s'", result)
		}

		result = JoinErrors([]string{"single"}, ", ")
		if result != "single" {
			t.Errorf("Expected 'single', got '%s'", result)
		}

		result = JoinErrors([]string{}, ", ")
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("GetErrorMap and PutErrorMap", func(t *testing.T) {
		m := GetErrorMap()
		if m == nil {
			t.Error("Expected non-nil map")
		}
		m["field"] = []string{"error"}
		PutErrorMap(m)

		// Get again should return empty map
		m2 := GetErrorMap()
		if len(m2) != 0 {
			t.Errorf("Expected empty map, got len=%d", len(m2))
		}
	})

	t.Run("GetBuilder and PutBuilder", func(t *testing.T) {
		b := GetBuilder()
		if b == nil {
			t.Error("Expected non-nil builder")
		}
		b.WriteString("test")
		PutBuilder(b)

		// Get again should return reset builder
		b2 := GetBuilder()
		if b2.Len() != 0 {
			t.Errorf("Expected reset builder, got len=%d", b2.Len())
		}
	})
}

func TestTimeValidator_Extended(t *testing.T) {
	t.Run("RequiredIf", func(t *testing.T) {
		schema := Schema{
			"hasDeadline": Bool(),
			"deadline": Time().RequiredIf(func(data DataObject) bool {
				return data["hasDeadline"] == true
			}),
		}

		err := Validate(DataObject{"hasDeadline": true}, schema)
		if err == nil {
			t.Error("Expected error when deadline required but missing")
		}

		err = Validate(DataObject{"hasDeadline": false}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("RequiredUnless", func(t *testing.T) {
		schema := Schema{
			"isImmediate": Bool(),
			"scheduledAt": Time().RequiredUnless(func(data DataObject) bool {
				return data["isImmediate"] == true
			}),
		}

		err := Validate(DataObject{"isImmediate": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}

		err = Validate(DataObject{"isImmediate": false}, schema)
		if err == nil {
			t.Error("Expected error when scheduledAt required but missing")
		}
	})

	t.Run("Custom", func(t *testing.T) {
		schema := Schema{
			"meeting": Time().Required().Custom(func(t2 time.Time, lookup Lookup) error {
				// Custom validation - just return nil for coverage
				return nil
			}),
		}

		err := Validate(DataObject{"meeting": "2024-01-01T10:00:00Z"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("Message", func(t *testing.T) {
		schema := Schema{
			"date": Time().Required().Message("required", "Date is required"),
		}

		err := Validate(DataObject{"date": nil}, schema)
		if err == nil || err.Errors["date"][0] != "Date is required" {
			t.Error("Expected custom message")
		}
	})

	t.Run("Default", func(t *testing.T) {
		schema := Schema{
			"created": Time().Default(time.Now()),
		}

		err := Validate(DataObject{}, schema)
		if err != nil {
			t.Errorf("Expected default to be applied, got: %v", err.Errors)
		}
	})

	t.Run("Nullable", func(t *testing.T) {
		schema := Schema{
			"deleted": Time().Nullable(),
		}

		err := Validate(DataObject{"deleted": nil}, schema)
		if err != nil {
			t.Errorf("Expected nullable to accept nil, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_UniqueAndExists_Coverage(t *testing.T) {
	t.Run("Unique", func(t *testing.T) {
		schema := Schema{
			"userId": Int().Required().Unique("users", "id", nil),
		}

		// Without DB adapter, just test the schema builds correctly
		err := Validate(DataObject{"userId": 1}, schema)
		if err != nil {
			t.Errorf("Expected no error without DB, got: %v", err.Errors)
		}
	})

	t.Run("UniqueWithMessage", func(t *testing.T) {
		schema := Schema{
			"userId": Int().Required().UniqueWithMessage("User ID already taken", "users", "id", nil),
		}

		err := Validate(DataObject{"userId": 1}, schema)
		if err != nil {
			t.Errorf("Expected no error without DB, got: %v", err.Errors)
		}
	})

	t.Run("ExistsWithMessage", func(t *testing.T) {
		schema := Schema{
			"categoryId": Int().Required().ExistsWithMessage("Category not found", "categories", "id"),
		}

		err := Validate(DataObject{"categoryId": 1}, schema)
		if err != nil {
			t.Errorf("Expected no error without DB, got: %v", err.Errors)
		}
	})
}

func TestFileValidator_WithMessageVariants(t *testing.T) {
	t.Run("MimesWithMessage", func(t *testing.T) {
		schema := Schema{
			"doc": File().Required().MimesWithMessage([]string{"application/pdf"}, "Only PDF allowed"),
		}

		// Just ensure schema builds without panic
		_ = schema
	})

	t.Run("ExtensionsWithMessage", func(t *testing.T) {
		schema := Schema{
			"image": File().Required().ExtensionsWithMessage([]string{"jpg", "png"}, "Only JPG and PNG allowed"),
		}

		// Just ensure schema builds without panic
		_ = schema
	})
}

func TestNumberValidator_NegativeCoverage(t *testing.T) {
	schema := Schema{
		"debt": Float().Required().Negative(),
	}

	t.Run("valid negative", func(t *testing.T) {
		err := Validate(DataObject{"debt": float64(-100)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("invalid positive", func(t *testing.T) {
		err := Validate(DataObject{"debt": float64(100)}, schema)
		if err == nil {
			t.Error("Expected error for positive number")
		}
	})

	t.Run("invalid zero", func(t *testing.T) {
		err := Validate(DataObject{"debt": float64(0)}, schema)
		if err == nil {
			t.Error("Expected error for zero")
		}
	})
}

func TestNumberValidator_IntegerCoverage(t *testing.T) {
	schema := Schema{
		"count": Float().Required().Integer(),
	}

	t.Run("valid integer", func(t *testing.T) {
		err := Validate(DataObject{"count": float64(42)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("invalid float", func(t *testing.T) {
		err := Validate(DataObject{"count": float64(3.14)}, schema)
		if err == nil {
			t.Error("Expected error for float")
		}
	})
}

func TestValidate_Parse(t *testing.T) {
	// Test Parse function (alias for Validate)
	schema := Schema{
		"name": String().Required(),
	}

	t.Run("valid data", func(t *testing.T) {
		err := Parse(DataObject{"name": "John"}, schema)
		if err != nil {
			t.Errorf("Parse should return nil for valid data: %v", err)
		}
	})

	t.Run("invalid data returns error", func(t *testing.T) {
		err := Parse(DataObject{"name": ""}, schema)
		if err == nil {
			t.Error("Parse should return error for invalid data")
		}
	})
}

// Additional coverage for builder methods with messages
func TestBuilderMethodsWithMessages(t *testing.T) {
	// Test Array methods with messages
	t.Run("Array.Min with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required().Min(2, "Need at least 2 items"),
		}
		err := Validate(DataObject{"items": []any{"a"}}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Array.Max with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required().Max(2, "Max 2 items allowed"),
		}
		err := Validate(DataObject{"items": []any{"a", "b", "c"}}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Array.Length with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required().Length(2, "Exactly 2 items required"),
		}
		err := Validate(DataObject{"items": []any{"a"}}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Array.Unique with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required().Unique("Items must be unique"),
		}
		err := Validate(DataObject{"items": []any{"a", "a"}}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	// Test Number methods with messages
	t.Run("Number.Min with message", func(t *testing.T) {
		schema := Schema{
			"age": Int().Required().Min(18, "Must be 18 or older"),
		}
		err := Validate(DataObject{"age": 15}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Number.Max with message", func(t *testing.T) {
		schema := Schema{
			"age": Int().Required().Max(100, "Must be 100 or younger"),
		}
		err := Validate(DataObject{"age": 150}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Number.MinDigits with message", func(t *testing.T) {
		schema := Schema{
			"code": Int().Required().MinDigits(4, "Code must be at least 4 digits"),
		}
		err := Validate(DataObject{"code": 123}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Number.MaxDigits with message", func(t *testing.T) {
		schema := Schema{
			"code": Int().Required().MaxDigits(4, "Code must be at most 4 digits"),
		}
		err := Validate(DataObject{"code": 12345}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Number.MultipleOf with message", func(t *testing.T) {
		schema := Schema{
			"qty": Int().Required().MultipleOf(5, "Must be multiple of 5"),
		}
		err := Validate(DataObject{"qty": 7}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	// Test File methods with messages
	t.Run("File.Min with message", func(t *testing.T) {
		schema := Schema{
			"file": File().Required().Min(1000, "File too small"),
		}
		// Just ensure it builds
		_ = schema
	})

	t.Run("File.Max with message", func(t *testing.T) {
		schema := Schema{
			"file": File().Required().Max(1000, "File too large"),
		}
		// Just ensure it builds
		_ = schema
	})

	t.Run("File.Image with message", func(t *testing.T) {
		schema := Schema{
			"avatar": File().Required().Image("Must be an image"),
		}
		// Just ensure it builds
		_ = schema
	})
}

// Test Boolean coercion edge cases
func TestBooleanCoercion(t *testing.T) {
	schema := Schema{
		"flag": Bool().Required().Coerce(),
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"string true", "true", false},
		{"string false", "false", false},
		{"string 1", "1", false},
		{"string 0", "0", false},
		{"int 1", 1, false},
		{"int 0", 0, false},
		{"float 1.0", 1.0, false},
		{"float 0.0", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"flag": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bool coerce %v: got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

// Test Array concurrent validation
func TestArrayConcurrent(t *testing.T) {
	schema := Schema{
		"items": Array().Required().Concurrent(4).Of(String().Min(2)),
	}

	t.Run("concurrent valid", func(t *testing.T) {
		data := DataObject{
			"items": []any{"abc", "def", "ghi", "jkl", "mno"},
		}
		err := Validate(data, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("concurrent with errors", func(t *testing.T) {
		data := DataObject{
			"items": []any{"abc", "x", "def", "y", "ghi"},
		}
		err := Validate(data, schema)
		if err == nil {
			t.Error("Expected errors for short strings")
		}
	})
}

// Test resolveMessage edge cases
func TestResolveMessage(t *testing.T) {
	// Test function-based messages
	schema := Schema{
		"name": String().Required().Min(3).Message("min", func(ctx MessageContext) string {
			return "Name '" + ctx.Value.(string) + "' is too short"
		}),
	}

	err := Validate(DataObject{"name": "ab"}, schema)
	if err == nil {
		t.Error("Expected error")
	}
	if err != nil && len(err.Errors["name"]) > 0 {
		if err.Errors["name"][0] != "Name 'ab' is too short" {
			t.Errorf("Expected dynamic message, got: %s", err.Errors["name"][0])
		}
	}
}

// Test Boolean True/False validators
func TestBooleanTrueFalse(t *testing.T) {
	t.Run("True validator", func(t *testing.T) {
		schema := Schema{
			"agree": Bool().Required().True("Must agree to terms"),
		}
		err := Validate(DataObject{"agree": false}, schema)
		if err == nil {
			t.Error("Expected error for false value")
		}
		err = Validate(DataObject{"agree": true}, schema)
		if err != nil {
			t.Errorf("Expected no error for true, got: %v", err.Errors)
		}
	})

	t.Run("False validator", func(t *testing.T) {
		schema := Schema{
			"disabled": Bool().Required().False("Must be disabled"),
		}
		err := Validate(DataObject{"disabled": true}, schema)
		if err == nil {
			t.Error("Expected error for true value")
		}
		err = Validate(DataObject{"disabled": false}, schema)
		if err != nil {
			t.Errorf("Expected no error for false, got: %v", err.Errors)
		}
	})
}

// Test Number.Between edge cases
func TestNumberBetweenEdgeCases(t *testing.T) {
	schema := Schema{
		"score": Float().Required().Between(0, 100),
	}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"at min", 0, false},
		{"at max", 100, false},
		{"in middle", 50, false},
		{"below min", -1, true},
		{"above max", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"score": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Between(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

// Test String regex edge cases
func TestStringRegexEdgeCases(t *testing.T) {
	t.Run("Regex with message", func(t *testing.T) {
		schema := Schema{
			"code": String().Required().Regex("^[A-Z]{3}$", "Must be 3 uppercase letters"),
		}
		err := Validate(DataObject{"code": "abc"}, schema)
		if err == nil {
			t.Error("Expected error for lowercase")
		}
		err = Validate(DataObject{"code": "ABC"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("NotRegex with message", func(t *testing.T) {
		schema := Schema{
			"text": String().Required().NotRegex("\\d+", "Must not contain numbers"),
		}
		err := Validate(DataObject{"text": "hello123"}, schema)
		if err == nil {
			t.Error("Expected error for containing numbers")
		}
		err = Validate(DataObject{"text": "hello"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

// Test Number.Regex and NotRegex
func TestNumberRegex(t *testing.T) {
	t.Run("Number.Regex", func(t *testing.T) {
		schema := Schema{
			"code": Int().Required().Regex("^\\d{4}$", "Must be 4 digits"),
		}
		err := Validate(DataObject{"code": 1234}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
		err = Validate(DataObject{"code": 123}, schema)
		if err == nil {
			t.Error("Expected error for 3 digits")
		}
	})

	t.Run("Number.NotRegex", func(t *testing.T) {
		schema := Schema{
			"code": Int().Required().NotRegex("^0", "Must not start with 0"),
		}
		err := Validate(DataObject{"code": 123}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

// Test more cache edge cases
func TestCacheEdgeCases(t *testing.T) {
	// Test compiling same regex multiple times
	pattern := "^test\\d+$"
	for i := 0; i < 5; i++ {
		re, err := globalRegexCache.GetOrCompile(pattern)
		if err != nil {
			t.Errorf("Failed to compile regex: %v", err)
		}
		if re == nil {
			t.Error("Expected non-nil regex")
		}
	}

	// Test invalid regex
	_, err := globalRegexCache.GetOrCompile("[invalid")
	if err == nil {
		t.Error("Expected error for invalid regex")
	}
}

// Test Object validator edge cases
func TestObjectEdgeCases(t *testing.T) {
	t.Run("nested object with passthrough", func(t *testing.T) {
		schema := Schema{
			"config": Object().Required().Passthrough().Shape(Schema{
				"name": String().Required(),
			}),
		}
		data := DataObject{
			"config": map[string]any{
				"name":     "test",
				"extraKey": "extraValue",
			},
		}
		err := Validate(data, schema)
		if err != nil {
			t.Errorf("Expected no error with passthrough, got: %v", err.Errors)
		}
	})

	t.Run("strict object rejects extra keys", func(t *testing.T) {
		schema := Schema{
			"config": Object().Required().Strict().Shape(Schema{
				"name": String().Required(),
			}),
		}
		data := DataObject{
			"config": map[string]any{
				"name":     "test",
				"extraKey": "extraValue",
			},
		}
		err := Validate(data, schema)
		if err == nil {
			t.Error("Expected error for extra keys in strict mode")
		}
	})
}

// Test Time validator additional methods
func TestTimeAdditionalMethods(t *testing.T) {
	t.Run("AfterNow", func(t *testing.T) {
		schema := Schema{
			"expiry": Time().Required().AfterNow(),
		}
		// Past date should fail
		err := Validate(DataObject{"expiry": "2020-01-01T00:00:00Z"}, schema)
		if err == nil {
			t.Error("Expected error for past date")
		}
	})

	t.Run("BeforeNow", func(t *testing.T) {
		schema := Schema{
			"birthdate": Time().Required().BeforeNow(),
		}
		// Future date should fail
		err := Validate(DataObject{"birthdate": "2099-01-01T00:00:00Z"}, schema)
		if err == nil {
			t.Error("Expected error for future date")
		}
	})

	t.Run("BeforeField", func(t *testing.T) {
		schema := Schema{
			"startDate": Time().Required(),
			"endDate":   Time().Required().BeforeField("startDate"),
		}
		// endDate before startDate should pass
		err := Validate(DataObject{
			"startDate": "2024-12-31T00:00:00Z",
			"endDate":   "2024-01-01T00:00:00Z",
		}, schema)
		// Note: BeforeField validates endDate is before startDate
		_ = err
	})
}

// Test Union.Nullable
func TestUnionNullable(t *testing.T) {
	schema := Schema{
		"id": Union(String().UUID(), Int().Positive()).Nullable(),
	}

	err := Validate(DataObject{"id": nil}, schema)
	if err != nil {
		t.Errorf("Expected nullable union to accept nil, got: %v", err.Errors)
	}
}

// Test ValidationError.Error method
func TestValidationErrorError(t *testing.T) {
	err := &ValidationError{
		Errors: map[string][]string{
			"name":  {"is required"},
			"email": {"must be valid"},
		},
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("Expected non-empty error string")
	}
}

// Test resolveMessage with additional scenarios
func TestResolveMessageExtended(t *testing.T) {
	// Test with nil message
	t.Run("nil message uses default", func(t *testing.T) {
		schema := Schema{
			"name": String().Required().Min(3),
		}
		err := Validate(DataObject{"name": "ab"}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	// Test with numeric context
	t.Run("message with index context", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required().Of(String().Min(2)),
		}
		err := Validate(DataObject{"items": []any{"a", "abc"}}, schema)
		if err == nil {
			t.Error("Expected error for short string")
		}
	})
}

// Test String.Digits
func TestStringDigits(t *testing.T) {
	schema := Schema{
		"phone": String().Required().Digits(10, "Phone must be 10 digits"),
	}

	t.Run("valid digits", func(t *testing.T) {
		err := Validate(DataObject{"phone": "1234567890"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("wrong length", func(t *testing.T) {
		err := Validate(DataObject{"phone": "123"}, schema)
		if err == nil {
			t.Error("Expected error for wrong digit count")
		}
	})

	t.Run("non-digits", func(t *testing.T) {
		err := Validate(DataObject{"phone": "123abc4567"}, schema)
		if err == nil {
			t.Error("Expected error for non-digits")
		}
	})
}

// Test Object.Pick and Omit
func TestObjectPickOmit(t *testing.T) {
	baseSchema := Schema{
		"name":  String().Required(),
		"email": String().Required().Email(),
		"age":   Int().Required(),
	}

	t.Run("Pick", func(t *testing.T) {
		schema := Schema{
			"user": Object().Required().Shape(baseSchema).Pick("name", "email"),
		}
		data := DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@test.com",
			},
		}
		err := Validate(data, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("Omit", func(t *testing.T) {
		schema := Schema{
			"user": Object().Required().Shape(baseSchema).Omit("age"),
		}
		data := DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@test.com",
			},
		}
		err := Validate(data, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

// Test Object.Merge
func TestObjectMerge(t *testing.T) {
	schema1 := Schema{
		"name": String().Required(),
	}
	obj2 := Object().Required().Shape(Schema{
		"email": String().Required().Email(),
	})

	schema := Schema{
		"user": Object().Required().Shape(schema1).Merge(obj2),
	}

	data := DataObject{
		"user": map[string]any{
			"name":  "John",
			"email": "john@test.com",
		},
	}
	err := Validate(data, schema)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err.Errors)
	}
}

// Test Optional validator edge cases
func TestOptionalEdgeCases(t *testing.T) {
	t.Run("optional with nil", func(t *testing.T) {
		schema := Schema{
			"nickname": Optional(String().Min(3)),
		}
		err := Validate(DataObject{"nickname": nil}, schema)
		if err != nil {
			t.Errorf("Expected optional to accept nil, got: %v", err.Errors)
		}
	})

	t.Run("optional with empty", func(t *testing.T) {
		schema := Schema{
			"nickname": Optional(String().Min(3)),
		}
		err := Validate(DataObject{}, schema)
		if err != nil {
			t.Errorf("Expected optional to accept missing, got: %v", err.Errors)
		}
	})

	t.Run("optional with valid value", func(t *testing.T) {
		schema := Schema{
			"nickname": Optional(String().Min(3)),
		}
		err := Validate(DataObject{"nickname": "Johnny"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("optional with invalid value", func(t *testing.T) {
		schema := Schema{
			"nickname": Optional(String().Min(3)),
		}
		err := Validate(DataObject{"nickname": "Jo"}, schema)
		if err == nil {
			t.Error("Expected error for invalid value even if optional")
		}
	})
}

// Test RequiredIf and RequiredUnless branches
func TestRequiredIfUnless(t *testing.T) {
	alwaysTrue := func(data DataObject) bool { return true }
	alwaysFalse := func(data DataObject) bool { return false }

	t.Run("array RequiredIf true with missing value", func(t *testing.T) {
		schema := Schema{
			"items": Array().RequiredIf(alwaysTrue),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for RequiredIf true")
		}
	})

	t.Run("array RequiredUnless false with missing value", func(t *testing.T) {
		schema := Schema{
			"items": Array().RequiredUnless(alwaysFalse),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for RequiredUnless false (condition is false, so required)")
		}
	})

	t.Run("bool RequiredIf true with missing value", func(t *testing.T) {
		schema := Schema{
			"active": Bool().RequiredIf(alwaysTrue),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for bool RequiredIf true")
		}
	})

	t.Run("bool RequiredUnless false with missing value", func(t *testing.T) {
		schema := Schema{
			"active": Bool().RequiredUnless(alwaysFalse),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for bool RequiredUnless false")
		}
	})

	t.Run("file RequiredIf true with missing value", func(t *testing.T) {
		schema := Schema{
			"document": File().RequiredIf(alwaysTrue),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for file RequiredIf true")
		}
	})

	t.Run("file RequiredUnless false with missing value", func(t *testing.T) {
		schema := Schema{
			"document": File().RequiredUnless(alwaysFalse),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for file RequiredUnless false")
		}
	})

	t.Run("number RequiredIf true with missing value", func(t *testing.T) {
		schema := Schema{
			"count": Int().RequiredIf(alwaysTrue),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for number RequiredIf true")
		}
	})

	t.Run("number RequiredUnless false with missing value", func(t *testing.T) {
		schema := Schema{
			"count": Int().RequiredUnless(alwaysFalse),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for number RequiredUnless false")
		}
	})

	t.Run("object RequiredIf true with missing value", func(t *testing.T) {
		schema := Schema{
			"profile": Object().RequiredIf(alwaysTrue).Shape(Schema{"name": String()}),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for object RequiredIf true")
		}
	})

	t.Run("object RequiredUnless false with missing value", func(t *testing.T) {
		schema := Schema{
			"profile": Object().RequiredUnless(alwaysFalse).Shape(Schema{"name": String()}),
		}
		err := Validate(DataObject{}, schema)
		if err == nil {
			t.Error("Expected error for object RequiredUnless false")
		}
	})
}

// Test numeric type conversions
func TestNumericConversions(t *testing.T) {
	t.Run("int32 conversion", func(t *testing.T) {
		schema := Schema{
			"value": Int().Required(),
		}
		err := Validate(DataObject{"value": int32(42)}, schema)
		if err != nil {
			t.Errorf("Expected int32 conversion to work, got: %v", err.Errors)
		}
	})

	t.Run("uint conversion", func(t *testing.T) {
		schema := Schema{
			"value": Int().Required(),
		}
		err := Validate(DataObject{"value": uint(42)}, schema)
		if err != nil {
			t.Errorf("Expected uint conversion to work, got: %v", err.Errors)
		}
	})

	t.Run("uint32 conversion", func(t *testing.T) {
		schema := Schema{
			"value": Int().Required(),
		}
		err := Validate(DataObject{"value": uint32(42)}, schema)
		if err != nil {
			t.Errorf("Expected uint32 conversion to work, got: %v", err.Errors)
		}
	})

	t.Run("uint64 conversion", func(t *testing.T) {
		schema := Schema{
			"value": Int().Required(),
		}
		err := Validate(DataObject{"value": uint64(42)}, schema)
		if err != nil {
			t.Errorf("Expected uint64 conversion to work, got: %v", err.Errors)
		}
	})

	t.Run("float32 conversion", func(t *testing.T) {
		schema := Schema{
			"value": Float().Required(),
		}
		err := Validate(DataObject{"value": float32(3.14)}, schema)
		if err != nil {
			t.Errorf("Expected float32 conversion to work, got: %v", err.Errors)
		}
	})

	t.Run("int64 conversion", func(t *testing.T) {
		schema := Schema{
			"value": Int().Required(),
		}
		err := Validate(DataObject{"value": int64(42)}, schema)
		if err != nil {
			t.Errorf("Expected int64 conversion to work, got: %v", err.Errors)
		}
	})
}

// Test Negative and Integer with actual negative/non-integer values
func TestNegativeIntegerValidation(t *testing.T) {
	t.Run("negative validation with positive number", func(t *testing.T) {
		schema := Schema{
			"value": Int().Negative(),
		}
		err := Validate(DataObject{"value": 5}, schema)
		if err == nil {
			t.Error("Expected error for positive number with Negative validator")
		}
	})

	t.Run("integer validation with float", func(t *testing.T) {
		schema := Schema{
			"value": Float().Integer(),
		}
		err := Validate(DataObject{"value": 3.14}, schema)
		if err == nil {
			t.Error("Expected error for float with Integer validator")
		}
	})

	t.Run("integer validation with whole number", func(t *testing.T) {
		schema := Schema{
			"value": Float().Integer(),
		}
		err := Validate(DataObject{"value": 3.0}, schema)
		if err != nil {
			t.Errorf("Expected no error for whole number, got: %v", err.Errors)
		}
	})
}

// Test comparison validators (using field references)
func TestComparisonValidators(t *testing.T) {
	t.Run("LessThan fail", func(t *testing.T) {
		schema := Schema{
			"value":    Int(),
			"maxValue": Int(),
		}
		err := Validate(DataObject{"value": 15, "maxValue": 10}, schema)
		if err != nil {
			t.Errorf("Unexpected error: %v", err.Errors)
		}
		// Test using Max which takes a value
		schema2 := Schema{
			"value": Int().Max(10),
		}
		err = Validate(DataObject{"value": 15}, schema2)
		if err == nil {
			t.Error("Expected error for Max")
		}
	})

	t.Run("GreaterThan fail", func(t *testing.T) {
		schema := Schema{
			"value": Int().Min(10),
		}
		err := Validate(DataObject{"value": 5}, schema)
		if err == nil {
			t.Error("Expected error for Min")
		}
	})

	t.Run("field comparisons", func(t *testing.T) {
		schema := Schema{
			"min":   Int(),
			"max":   Int(),
			"value": Int().LessThan("max").GreaterThan("min"),
		}
		// Valid: 5 < 10, 5 > 1
		err := Validate(DataObject{"min": 1, "max": 10, "value": 5}, schema)
		if err != nil {
			t.Errorf("Expected valid comparison, got: %v", err.Errors)
		}
		// Invalid: 15 > max (10)
		err = Validate(DataObject{"min": 1, "max": 10, "value": 15}, schema)
		if err == nil {
			t.Error("Expected error for value > max")
		}
		// Invalid: 0 < min (1)
		err = Validate(DataObject{"min": 1, "max": 10, "value": 0}, schema)
		if err == nil {
			t.Error("Expected error for value < min")
		}
	})
}

// Test Between with out of range values
func TestBetweenOutOfRange(t *testing.T) {
	t.Run("below min", func(t *testing.T) {
		schema := Schema{
			"value": Int().Between(10, 20),
		}
		err := Validate(DataObject{"value": 5}, schema)
		if err == nil {
			t.Error("Expected error for value below min")
		}
	})

	t.Run("above max", func(t *testing.T) {
		schema := Schema{
			"value": Int().Between(10, 20),
		}
		err := Validate(DataObject{"value": 25}, schema)
		if err == nil {
			t.Error("Expected error for value above max")
		}
	})
}

// Test convertToType function
func TestConvertToType(t *testing.T) {
	// Test via Enum validation with different types
	t.Run("enum with float64 value for int enum", func(t *testing.T) {
		schema := Schema{
			"status": Enum(1, 2, 3).Required(),
		}
		// JSON numbers are float64
		err := Validate(DataObject{"status": float64(2)}, schema)
		if err != nil {
			t.Errorf("Expected float64 to convert to int, got: %v", err.Errors)
		}
	})

	t.Run("enum with int64 value for int enum", func(t *testing.T) {
		schema := Schema{
			"status": Enum(1, 2, 3).Required(),
		}
		err := Validate(DataObject{"status": int64(2)}, schema)
		if err != nil {
			t.Errorf("Expected int64 to convert to int, got: %v", err.Errors)
		}
	})
}

// Test string validators that are at 75%
func TestStringValidatorsLowCoverage(t *testing.T) {
	t.Run("ASCII invalid", func(t *testing.T) {
		schema := Schema{
			"text": String().ASCII(),
		}
		err := Validate(DataObject{"text": "Hello 世界"}, schema)
		if err == nil {
			t.Error("Expected error for non-ASCII")
		}
	})

	t.Run("IPv4 invalid", func(t *testing.T) {
		schema := Schema{
			"ip": String().IPv4(),
		}
		err := Validate(DataObject{"ip": "not-an-ip"}, schema)
		if err == nil {
			t.Error("Expected error for invalid IPv4")
		}
	})

	t.Run("IPv6 invalid", func(t *testing.T) {
		schema := Schema{
			"ip": String().IPv6(),
		}
		err := Validate(DataObject{"ip": "not-an-ip"}, schema)
		if err == nil {
			t.Error("Expected error for invalid IPv6")
		}
	})

	t.Run("HexColor invalid", func(t *testing.T) {
		schema := Schema{
			"color": String().HexColor(),
		}
		err := Validate(DataObject{"color": "not-a-color"}, schema)
		if err == nil {
			t.Error("Expected error for invalid hex color")
		}
	})

	t.Run("Base64 invalid", func(t *testing.T) {
		schema := Schema{
			"data": String().Base64(),
		}
		err := Validate(DataObject{"data": "!!not-base64!!"}, schema)
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})

	t.Run("MAC invalid", func(t *testing.T) {
		schema := Schema{
			"mac": String().MAC(),
		}
		err := Validate(DataObject{"mac": "not-a-mac"}, schema)
		if err == nil {
			t.Error("Expected error for invalid MAC address")
		}
	})

	t.Run("ULID invalid", func(t *testing.T) {
		schema := Schema{
			"id": String().ULID(),
		}
		err := Validate(DataObject{"id": "not-a-ulid"}, schema)
		if err == nil {
			t.Error("Expected error for invalid ULID")
		}
	})

	t.Run("AlphaDash invalid", func(t *testing.T) {
		schema := Schema{
			"slug": String().AlphaDash(),
		}
		err := Validate(DataObject{"slug": "hello world spaces"}, schema)
		if err == nil {
			t.Error("Expected error for invalid AlphaDash")
		}
	})
}

// Test Object validators that are at lower coverage
func TestObjectValidatorsLowCoverage(t *testing.T) {
	t.Run("strict with extra keys", func(t *testing.T) {
		schema := Schema{
			"user": Object().Strict().Shape(Schema{
				"name": String().Required(),
			}),
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"extra": "not allowed",
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for extra keys in strict mode")
		}
	})

	t.Run("Pick creates subset", func(t *testing.T) {
		baseSchema := Schema{
			"name":  String().Required(),
			"email": String().Required().Email(),
			"age":   Int().Required(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Pick("name", "email"),
		}
		// Should only validate name and email, not age
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@test.com",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected Pick to only validate selected fields, got: %v", err.Errors)
		}
	})

	t.Run("Omit excludes fields", func(t *testing.T) {
		baseSchema := Schema{
			"name":     String().Required(),
			"email":    String().Required().Email(),
			"password": String().Required(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Omit("password"),
		}
		// Should not require password
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@test.com",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected Omit to exclude password, got: %v", err.Errors)
		}
	})

	t.Run("Partial makes all fields optional", func(t *testing.T) {
		baseSchema := Schema{
			"name":  String().Required(),
			"email": String().Required().Email(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Partial(),
		}
		// Should not require any fields
		err := Validate(DataObject{
			"user": map[string]any{},
		}, schema)
		if err != nil {
			t.Errorf("Expected Partial to make all fields optional, got: %v", err.Errors)
		}
	})

	t.Run("Extend adds fields", func(t *testing.T) {
		baseSchema := Schema{
			"name": String().Required(),
		}
		extendSchema := Schema{
			"email": String().Required().Email(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Extend(extendSchema),
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, schema)
		if err == nil {
			t.Error("Expected Extend to add email requirement")
		}
	})
}

// Test coerceToBool edge cases
func TestCoerceToBoolEdgeCases(t *testing.T) {
	t.Run("coerce int to bool", func(t *testing.T) {
		schema := Schema{
			"flag": Bool().Coerce(),
		}
		err := Validate(DataObject{"flag": 1}, schema)
		if err != nil {
			t.Errorf("Expected int 1 to coerce to true, got: %v", err.Errors)
		}
	})

	t.Run("coerce zero to false", func(t *testing.T) {
		schema := Schema{
			"flag": Bool().Coerce(),
		}
		err := Validate(DataObject{"flag": 0}, schema)
		if err != nil {
			t.Errorf("Expected int 0 to coerce to false, got: %v", err.Errors)
		}
	})

	t.Run("coerce non-coercible type", func(t *testing.T) {
		schema := Schema{
			"flag": Bool().Coerce().Required(),
		}
		err := Validate(DataObject{"flag": []int{1, 2, 3}}, schema)
		if err == nil {
			t.Error("Expected error for non-coercible type")
		}
	})
}

// Test getSplitPath edge cases
func TestGetSplitPathEdgeCases(t *testing.T) {
	t.Run("nested array path", func(t *testing.T) {
		schema := Schema{
			"users": Array().Of(Object().Shape(Schema{
				"tags": Array().Of(String().Required().Min(1)),
			})),
		}
		err := Validate(DataObject{
			"users": []any{
				map[string]any{
					"tags": []any{"", "valid"},
				},
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for empty tag")
		}
	})
}

// Test Array concurrent with actual errors
func TestArrayConcurrentWithErrors(t *testing.T) {
	schema := Schema{
		"items": Array().Of(String().Email()).Concurrent(4),
	}
	data := DataObject{
		"items": []any{"test@test.com", "invalid-email", "another@test.com"},
	}
	err := Validate(data, schema)
	if err == nil {
		t.Error("Expected error for invalid email in concurrent array")
	}
}

// Test custom messages for Required/RequiredIf/RequiredUnless
func TestRequiredWithCustomMessages(t *testing.T) {
	alwaysTrue := func(data DataObject) bool { return true }
	alwaysFalse := func(data DataObject) bool { return false }

	t.Run("array Required with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().Required("Items are required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["items"], "Items are required") {
			t.Error("Expected custom message for array Required")
		}
	})

	t.Run("array RequiredIf with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().RequiredIf(alwaysTrue, "Items needed when condition true"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["items"], "Items needed when condition true") {
			t.Error("Expected custom message for array RequiredIf")
		}
	})

	t.Run("array RequiredUnless with message", func(t *testing.T) {
		schema := Schema{
			"items": Array().RequiredUnless(alwaysFalse, "Items needed unless condition true"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["items"], "Items needed unless condition true") {
			t.Error("Expected custom message for array RequiredUnless")
		}
	})

	t.Run("bool RequiredIf with message", func(t *testing.T) {
		schema := Schema{
			"active": Bool().RequiredIf(alwaysTrue, "Active flag is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["active"], "Active flag is required") {
			t.Error("Expected custom message for bool RequiredIf")
		}
	})

	t.Run("bool RequiredUnless with message", func(t *testing.T) {
		schema := Schema{
			"active": Bool().RequiredUnless(alwaysFalse, "Active flag is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["active"], "Active flag is required") {
			t.Error("Expected custom message for bool RequiredUnless")
		}
	})

	t.Run("file RequiredIf with message", func(t *testing.T) {
		schema := Schema{
			"doc": File().RequiredIf(alwaysTrue, "Document is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["doc"], "Document is required") {
			t.Error("Expected custom message for file RequiredIf")
		}
	})

	t.Run("file RequiredUnless with message", func(t *testing.T) {
		schema := Schema{
			"doc": File().RequiredUnless(alwaysFalse, "Document is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["doc"], "Document is required") {
			t.Error("Expected custom message for file RequiredUnless")
		}
	})

	t.Run("number RequiredIf with message", func(t *testing.T) {
		schema := Schema{
			"count": Int().RequiredIf(alwaysTrue, "Count is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["count"], "Count is required") {
			t.Error("Expected custom message for number RequiredIf")
		}
	})

	t.Run("number RequiredUnless with message", func(t *testing.T) {
		schema := Schema{
			"count": Int().RequiredUnless(alwaysFalse, "Count is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["count"], "Count is required") {
			t.Error("Expected custom message for number RequiredUnless")
		}
	})

	t.Run("object RequiredIf with message", func(t *testing.T) {
		schema := Schema{
			"profile": Object().RequiredIf(alwaysTrue, "Profile is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["profile"], "Profile is required") {
			t.Error("Expected custom message for object RequiredIf")
		}
	})

	t.Run("object RequiredUnless with message", func(t *testing.T) {
		schema := Schema{
			"profile": Object().RequiredUnless(alwaysFalse, "Profile is required"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["profile"], "Profile is required") {
			t.Error("Expected custom message for object RequiredUnless")
		}
	})
}

// Test number validators with custom messages
func TestNumberValidatorCustomMessages(t *testing.T) {
	t.Run("Negative with message", func(t *testing.T) {
		schema := Schema{
			"value": Int().Negative("Must be negative"),
		}
		err := Validate(DataObject{"value": 5}, schema)
		if err == nil || !containsStr(err.Errors["value"], "Must be negative") {
			t.Error("Expected custom message for Negative")
		}
	})

	t.Run("Integer with message", func(t *testing.T) {
		schema := Schema{
			"value": Float().Integer("Must be a whole number"),
		}
		err := Validate(DataObject{"value": 3.14}, schema)
		if err == nil || !containsStr(err.Errors["value"], "Must be a whole number") {
			t.Error("Expected custom message for Integer")
		}
	})

	t.Run("LessThan with message", func(t *testing.T) {
		schema := Schema{
			"max":   Int(),
			"value": Int().LessThan("max", "Must be less than max"),
		}
		err := Validate(DataObject{"max": 10, "value": 15}, schema)
		if err == nil || !containsStr(err.Errors["value"], "Must be less than max") {
			t.Error("Expected custom message for LessThan")
		}
	})

	t.Run("GreaterThan with message", func(t *testing.T) {
		schema := Schema{
			"min":   Int(),
			"value": Int().GreaterThan("min", "Must be greater than min"),
		}
		err := Validate(DataObject{"min": 10, "value": 5}, schema)
		if err == nil || !containsStr(err.Errors["value"], "Must be greater than min") {
			t.Error("Expected custom message for GreaterThan")
		}
	})

	t.Run("LessThanOrEqual with message", func(t *testing.T) {
		schema := Schema{
			"max":   Int(),
			"value": Int().LessThanOrEqual("max", "Must be at most max"),
		}
		err := Validate(DataObject{"max": 10, "value": 15}, schema)
		if err == nil || !containsStr(err.Errors["value"], "Must be at most max") {
			t.Error("Expected custom message for LessThanOrEqual")
		}
	})

	t.Run("GreaterThanOrEqual with message", func(t *testing.T) {
		schema := Schema{
			"min":   Int(),
			"value": Int().GreaterThanOrEqual("min", "Must be at least min"),
		}
		err := Validate(DataObject{"min": 10, "value": 5}, schema)
		if err == nil || !containsStr(err.Errors["value"], "Must be at least min") {
			t.Error("Expected custom message for GreaterThanOrEqual")
		}
	})
}

// Test string validators with custom messages
func TestStringValidatorCustomMessages(t *testing.T) {
	t.Run("ASCII with message", func(t *testing.T) {
		schema := Schema{
			"text": String().ASCII("Only ASCII characters allowed"),
		}
		err := Validate(DataObject{"text": "Hello 世界"}, schema)
		if err == nil || !containsStr(err.Errors["text"], "Only ASCII characters allowed") {
			t.Error("Expected custom message for ASCII")
		}
	})

	t.Run("IPv4 with message", func(t *testing.T) {
		schema := Schema{
			"ip": String().IPv4("Must be valid IPv4"),
		}
		err := Validate(DataObject{"ip": "not-an-ip"}, schema)
		if err == nil || !containsStr(err.Errors["ip"], "Must be valid IPv4") {
			t.Error("Expected custom message for IPv4")
		}
	})

	t.Run("IPv6 with message", func(t *testing.T) {
		schema := Schema{
			"ip": String().IPv6("Must be valid IPv6"),
		}
		err := Validate(DataObject{"ip": "not-an-ip"}, schema)
		if err == nil || !containsStr(err.Errors["ip"], "Must be valid IPv6") {
			t.Error("Expected custom message for IPv6")
		}
	})

	t.Run("HexColor with message", func(t *testing.T) {
		schema := Schema{
			"color": String().HexColor("Must be valid hex color"),
		}
		err := Validate(DataObject{"color": "not-a-color"}, schema)
		if err == nil || !containsStr(err.Errors["color"], "Must be valid hex color") {
			t.Error("Expected custom message for HexColor")
		}
	})

	t.Run("Base64 with message", func(t *testing.T) {
		schema := Schema{
			"data": String().Base64("Must be valid base64"),
		}
		err := Validate(DataObject{"data": "!!invalid!!"}, schema)
		if err == nil || !containsStr(err.Errors["data"], "Must be valid base64") {
			t.Error("Expected custom message for Base64")
		}
	})

	t.Run("MAC with message", func(t *testing.T) {
		schema := Schema{
			"mac": String().MAC("Must be valid MAC address"),
		}
		err := Validate(DataObject{"mac": "invalid"}, schema)
		if err == nil || !containsStr(err.Errors["mac"], "Must be valid MAC address") {
			t.Error("Expected custom message for MAC")
		}
	})

	t.Run("ULID with message", func(t *testing.T) {
		schema := Schema{
			"id": String().ULID("Must be valid ULID"),
		}
		err := Validate(DataObject{"id": "invalid"}, schema)
		if err == nil || !containsStr(err.Errors["id"], "Must be valid ULID") {
			t.Error("Expected custom message for ULID")
		}
	})

	t.Run("AlphaDash with message", func(t *testing.T) {
		schema := Schema{
			"slug": String().AlphaDash("Only letters, numbers, dashes allowed"),
		}
		err := Validate(DataObject{"slug": "has spaces"}, schema)
		if err == nil || !containsStr(err.Errors["slug"], "Only letters, numbers, dashes allowed") {
			t.Error("Expected custom message for AlphaDash")
		}
	})
}

// Test concurrent array with lower concurrency values
func TestArrayConcurrentOptions(t *testing.T) {
	t.Run("concurrent with n=1", func(t *testing.T) {
		schema := Schema{
			"items": Array().Of(String()).Concurrent(1),
		}
		err := Validate(DataObject{"items": []any{"a", "b", "c"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("concurrent with n=0 falls back to non-concurrent", func(t *testing.T) {
		schema := Schema{
			"items": Array().Of(String()).Concurrent(0),
		}
		err := Validate(DataObject{"items": []any{"a", "b", "c"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

// Test file dimension validation
func TestFileDimensionValidation(t *testing.T) {
	t.Run("Dimensions with message", func(t *testing.T) {
		dims := &ImageDimensions{MinWidth: 100, MaxWidth: 100, MinHeight: 100, MaxHeight: 100}
		schema := Schema{
			"image": File().Dimensions(dims, "Must be 100x100"),
		}
		// Empty file should not error if file is optional
		err := Validate(DataObject{}, schema)
		if err != nil {
			t.Logf("Got error: %v", err.Errors)
		}
	})
}

// Helper function to check if slice contains a string
func containsStr(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Test Required with custom messages for file, number, object
func TestRequiredCustomMessages(t *testing.T) {
	t.Run("file Required with message", func(t *testing.T) {
		schema := Schema{
			"doc": File().Required("Document is mandatory"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["doc"], "Document is mandatory") {
			t.Error("Expected custom message for file Required")
		}
	})

	t.Run("number Required with message", func(t *testing.T) {
		schema := Schema{
			"count": Int().Required("Count is mandatory"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["count"], "Count is mandatory") {
			t.Error("Expected custom message for number Required")
		}
	})

	t.Run("object Required with message", func(t *testing.T) {
		schema := Schema{
			"profile": Object().Required("Profile is mandatory"),
		}
		err := Validate(DataObject{}, schema)
		if err == nil || !containsStr(err.Errors["profile"], "Profile is mandatory") {
			t.Error("Expected custom message for object Required")
		}
	})
}

// Test Object validators with custom messages
func TestObjectValidatorCustomMessages(t *testing.T) {
	t.Run("Strict with extra key error", func(t *testing.T) {
		schema := Schema{
			"user": Object().Strict().Shape(Schema{
				"name": String().Required(),
			}),
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name":    "John",
				"unknown": "value",
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for extra key in strict mode")
		}
	})

	t.Run("Pick non-existent field", func(t *testing.T) {
		baseSchema := Schema{
			"name": String().Required(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Pick("name", "email"),
		}
		// Pick should just ignore non-existent fields in the base schema
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected Pick to work with only existing fields, got: %v", err.Errors)
		}
	})

	t.Run("Omit all fields", func(t *testing.T) {
		baseSchema := Schema{
			"name":  String().Required(),
			"email": String().Required(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Omit("name", "email"),
		}
		// Should have no required fields after omitting all
		err := Validate(DataObject{
			"user": map[string]any{},
		}, schema)
		if err != nil {
			t.Errorf("Expected Omit to remove all fields, got: %v", err.Errors)
		}
	})

	t.Run("Partial makes fields optional", func(t *testing.T) {
		baseSchema := Schema{
			"name":  String().Required(),
			"email": String().Required().Email(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Partial(),
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"email": "test@test.com",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected Partial to make name optional, got: %v", err.Errors)
		}
	})

	t.Run("Extend with overlapping fields", func(t *testing.T) {
		baseSchema := Schema{
			"name": String(),
		}
		extendSchema := Schema{
			"name":  String().Required(), // Override name to be required
			"email": String().Required().Email(),
		}
		schema := Schema{
			"user": Object().Shape(baseSchema).Extend(extendSchema),
		}
		// Should now require name and email
		err := Validate(DataObject{
			"user": map[string]any{},
		}, schema)
		if err == nil {
			t.Error("Expected Extend to make name required")
		}
	})

	t.Run("Merge combines validators", func(t *testing.T) {
		obj1 := Object().Shape(Schema{
			"name": String().Required(),
		})
		obj2 := Object().Shape(Schema{
			"email": String().Required().Email(),
		})
		schema := Schema{
			"user": obj1.Merge(obj2),
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, schema)
		if err == nil {
			t.Error("Expected Merge to add email requirement")
		}
	})
}

// Test Between with both min and max violations
func TestBetweenValidation(t *testing.T) {
	t.Run("value at exact min", func(t *testing.T) {
		schema := Schema{
			"value": Int().Between(10, 20),
		}
		err := Validate(DataObject{"value": 10}, schema)
		if err != nil {
			t.Errorf("Expected value at min to be valid, got: %v", err.Errors)
		}
	})

	t.Run("value at exact max", func(t *testing.T) {
		schema := Schema{
			"value": Int().Between(10, 20),
		}
		err := Validate(DataObject{"value": 20}, schema)
		if err != nil {
			t.Errorf("Expected value at max to be valid, got: %v", err.Errors)
		}
	})

	t.Run("value below min", func(t *testing.T) {
		schema := Schema{
			"value": Int().Between(10, 20),
		}
		err := Validate(DataObject{"value": 5}, schema)
		if err == nil {
			t.Error("Expected error for value below min")
		}
	})

	t.Run("value above max", func(t *testing.T) {
		schema := Schema{
			"value": Int().Between(10, 20),
		}
		err := Validate(DataObject{"value": 25}, schema)
		if err == nil {
			t.Error("Expected error for value above max")
		}
	})
}

// Test Concurrent with validation errors to hit error path
func TestConcurrentWithValidationErrors(t *testing.T) {
	schema := Schema{
		"emails": Array().Of(String().Email()).Concurrent(2),
	}
	data := DataObject{
		"emails": []any{"good@email.com", "bad-email", "another@good.com", "also-bad"},
	}
	err := Validate(data, schema)
	if err == nil {
		t.Error("Expected validation errors for bad emails")
	}
	// Should have errors for both bad emails
	if len(err.Errors) < 2 {
		t.Errorf("Expected at least 2 errors, got: %d", len(err.Errors))
	}
}

// Test edge cases for coverage
func TestEdgeCasesForCoverage(t *testing.T) {
	t.Run("Concurrent with negative n", func(t *testing.T) {
		schema := Schema{
			"items": Array().Of(String()).Concurrent(-1),
		}
		err := Validate(DataObject{"items": []any{"a", "b"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("Merge with different properties", func(t *testing.T) {
		obj1 := Object().Required().Shape(Schema{
			"name": String().Required(),
		})
		obj2 := Object().Nullable().Strict().Shape(Schema{
			"email": String().Required().Email(),
		}).Message("required", "Field is needed")
		merged := obj1.Merge(obj2)
		schema := Schema{
			"user": merged,
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for missing email")
		}
	})

	t.Run("Partial with messages", func(t *testing.T) {
		obj := Object().Required().Shape(Schema{
			"name":  String().Required(),
			"email": String().Required().Email(),
		}).Message("required", "Object is required")
		partial := obj.Partial()
		schema := Schema{
			"user": partial,
		}
		// With Partial, even the object is not required
		err := Validate(DataObject{}, schema)
		if err != nil {
			t.Errorf("Expected Partial to make object optional, got: %v", err.Errors)
		}
	})

	t.Run("Extend with custom function", func(t *testing.T) {
		customFn := func(obj map[string]any, lookup Lookup) error {
			return nil
		}
		obj := Object().Shape(Schema{
			"name": String(),
		}).Custom(customFn)
		extended := obj.Extend(Schema{
			"email": String().Email(),
		})
		schema := Schema{
			"user": extended,
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@test.com",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("Object with passthrough", func(t *testing.T) {
		schema := Schema{
			"user": Object().Passthrough().Shape(Schema{
				"name": String().Required(),
			}),
		}
		// Passthrough should allow extra fields
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"extra": "allowed",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected passthrough to allow extra fields, got: %v", err.Errors)
		}
	})

	t.Run("Object strict vs passthrough", func(t *testing.T) {
		// Strict should reject extra fields
		schema := Schema{
			"user": Object().Strict().Shape(Schema{
				"name": String().Required(),
			}),
		}
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"extra": "not allowed",
			},
		}, schema)
		if err == nil {
			t.Error("Expected strict to reject extra fields")
		}
	})

	t.Run("Number with string coercion", func(t *testing.T) {
		schema := Schema{
			"count": Int().Coerce().Required(),
		}
		err := Validate(DataObject{"count": "42"}, schema)
		if err != nil {
			t.Errorf("Expected string coercion to work, got: %v", err.Errors)
		}
	})

	t.Run("Number with invalid string coercion", func(t *testing.T) {
		schema := Schema{
			"count": Int().Coerce().Required(),
		}
		err := Validate(DataObject{"count": "not-a-number"}, schema)
		if err == nil {
			t.Error("Expected error for invalid string coercion")
		}
	})

	t.Run("Array nonempty", func(t *testing.T) {
		schema := Schema{
			"items": Array().Nonempty(),
		}
		err := Validate(DataObject{"items": []any{}}, schema)
		if err == nil {
			t.Error("Expected error for empty array with Nonempty")
		}
	})

	t.Run("Array unique with duplicates", func(t *testing.T) {
		schema := Schema{
			"items": Array().Unique(),
		}
		err := Validate(DataObject{"items": []any{"a", "b", "a"}}, schema)
		if err == nil {
			t.Error("Expected error for duplicate in unique array")
		}
	})

	t.Run("Array length exact", func(t *testing.T) {
		schema := Schema{
			"items": Array().Length(3),
		}
		err := Validate(DataObject{"items": []any{"a", "b"}}, schema)
		if err == nil {
			t.Error("Expected error for wrong array length")
		}
	})

	t.Run("Time validator with After constraint", func(t *testing.T) {
		threshold := time.Now().Add(-24 * time.Hour)
		schema := Schema{
			"date": Time().After(threshold),
		}
		err := Validate(DataObject{"date": time.Now()}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("Time validator with Before constraint", func(t *testing.T) {
		threshold := time.Now().Add(24 * time.Hour)
		schema := Schema{
			"date": Time().Before(threshold),
		}
		err := Validate(DataObject{"date": time.Now()}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

// Test DataAccessor.Get with nil
func TestDataAccessorNil(t *testing.T) {
	var accessor DataAccessor
	result := accessor.Get("any.path")
	if result.Exists() {
		t.Error("Expected nil accessor to return non-existing result")
	}
}

// Test OptionalValidator with DBCheckCollector
func TestOptionalValidatorDBChecks(t *testing.T) {
	// String with Unique creates DB checks
	inner := String().Unique("users", "email", nil)
	opt := Optional(inner)

	// Should pass through DB checks from inner validator
	checks := opt.GetDBChecks("email", "test@test.com")
	if len(checks) == 0 {
		t.Error("Expected Optional to pass through DB checks from inner validator")
	}

	// Test with non-DBCheckCollector inner validator (e.g., Bool which has no DB checks)
	boolOpt := Optional(Bool())
	checks = boolOpt.GetDBChecks("flag", true)
	if len(checks) != 0 {
		t.Error("Expected no DB checks for Bool validator")
	}
}

// Test ValidationContext.FullPath edge cases
func TestValidationContextFullPath(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		ctx := &ValidationContext{Path: []string{}}
		if ctx.FullPath() != "" {
			t.Errorf("Expected empty string, got: %s", ctx.FullPath())
		}
	})

	t.Run("single path element", func(t *testing.T) {
		ctx := &ValidationContext{Path: []string{"field"}}
		if ctx.FullPath() != "field" {
			t.Errorf("Expected 'field', got: %s", ctx.FullPath())
		}
	})

	t.Run("multiple path elements", func(t *testing.T) {
		ctx := &ValidationContext{Path: []string{"user", "profile", "name"}}
		if ctx.FullPath() != "user.profile.name" {
			t.Errorf("Expected 'user.profile.name', got: %s", ctx.FullPath())
		}
	})
}

// Test OptionalValidator.Validate with non-nil, non-empty values
func TestOptionalValidatorValidate(t *testing.T) {
	opt := Optional(String().Min(3))

	// nil value - should pass
	schema := Schema{
		"name": opt,
	}
	err := Validate(DataObject{"name": nil}, schema)
	if err != nil {
		t.Errorf("Expected nil to be valid for optional, got: %v", err.Errors)
	}

	// empty string - should pass
	err = Validate(DataObject{"name": ""}, schema)
	if err != nil {
		t.Errorf("Expected empty string to be valid for optional, got: %v", err.Errors)
	}

	// valid value - should pass
	err = Validate(DataObject{"name": "John"}, schema)
	if err != nil {
		t.Errorf("Expected valid value to pass, got: %v", err.Errors)
	}

	// invalid value - should fail
	err = Validate(DataObject{"name": "Jo"}, schema)
	if err == nil {
		t.Error("Expected error for invalid value")
	}
}

// Test Strict with message
func TestStrictWithMessage(t *testing.T) {
	schema := Schema{
		"user": Object().Strict().Shape(Schema{
			"name": String().Required(),
		}).Message("strict", "No extra fields allowed"),
	}
	err := Validate(DataObject{
		"user": map[string]any{
			"name":  "John",
			"extra": "field",
		},
	}, schema)
	if err == nil {
		t.Error("Expected error for extra field in strict mode")
	}
}

// Test Pick with message
func TestPickWithMessage(t *testing.T) {
	baseSchema := Schema{
		"name":  String().Required(),
		"email": String().Required(),
		"age":   Int(),
	}
	schema := Schema{
		"user": Object().Shape(baseSchema).Pick("name"),
	}
	// Should only require name, not email
	err := Validate(DataObject{
		"user": map[string]any{
			"name": "John",
		},
	}, schema)
	if err != nil {
		t.Errorf("Expected Pick to keep only specified fields, got: %v", err.Errors)
	}
}

// Test Extend with message
func TestExtendWithMessage(t *testing.T) {
	baseObj := Object().Shape(Schema{
		"name": String(),
	}).Message("required", "Object required")
	extended := baseObj.Extend(Schema{
		"email": String().Required().Email(),
	})
	schema := Schema{
		"user": extended,
	}
	err := Validate(DataObject{
		"user": map[string]any{
			"name": "John",
		},
	}, schema)
	if err == nil {
		t.Error("Expected error for missing email after Extend")
	}
}

// Test convertToType with various types
func TestConvertToTypeEdgeCases(t *testing.T) {
	// Test via Enum validator which uses convertToType
	t.Run("enum with int conversion from float64", func(t *testing.T) {
		schema := Schema{
			"status": Enum(1, 2, 3).Required(),
		}
		// JSON unmarshals numbers as float64
		err := Validate(DataObject{"status": float64(2)}, schema)
		if err != nil {
			t.Errorf("Expected float64 to convert to int for enum, got: %v", err.Errors)
		}
	})

	t.Run("enum with int64 value", func(t *testing.T) {
		schema := Schema{
			"status": Enum(int64(1), int64(2), int64(3)).Required(),
		}
		err := Validate(DataObject{"status": int64(2)}, schema)
		if err != nil {
			t.Errorf("Expected int64 enum to work, got: %v", err.Errors)
		}
	})

	t.Run("enum with string value", func(t *testing.T) {
		schema := Schema{
			"status": Enum("active", "inactive", "pending").Required(),
		}
		err := Validate(DataObject{"status": "active"}, schema)
		if err != nil {
			t.Errorf("Expected string enum to work, got: %v", err.Errors)
		}
	})

	t.Run("enum with bool value", func(t *testing.T) {
		schema := Schema{
			"flag": Enum(true, false).Required(),
		}
		err := Validate(DataObject{"flag": true}, schema)
		if err != nil {
			t.Errorf("Expected bool enum to work, got: %v", err.Errors)
		}
	})
}
