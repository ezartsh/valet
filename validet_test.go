package valet

import (
	"context"
	"errors"
	"testing"
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
