package valet_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ezartsh/valet"
)

func Example_basicValidation() {
	// Parse JSON into map
	jsonData := []byte(`{
		"name": "John Doe",
		"email": "john@example.com",
		"age": 25
	}`)

	var data map[string]any
	json.Unmarshal(jsonData, &data)

	// Define validation schema
	schema := valet.Schema{
		"name":  valet.String().Required().Min(2).Max(100),
		"email": valet.String().Required().Email(),
		"age":   valet.Float().Required().Min(18).Max(120),
	}

	// Validate
	if err := valet.Validate(data, schema); err != nil {
		fmt.Println("Validation failed:", err.Errors)
	} else {
		fmt.Println("Validation passed!")
	}
	// Output: Validation passed!
}

func Example_nestedObjectValidation() {
	data := map[string]any{
		"user": map[string]any{
			"name":  "John",
			"email": "john@example.com",
		},
		"address": map[string]any{
			"street": "123 Main St",
			"city":   "New York",
			"zip":    "10001",
		},
	}

	schema := valet.Schema{
		"user": valet.Object().Required().Shape(valet.Schema{
			"name":  valet.String().Required().Min(2),
			"email": valet.String().Required().Email(),
		}),
		"address": valet.Object().Required().Shape(valet.Schema{
			"street": valet.String().Required().Min(5),
			"city":   valet.String().Required(),
			"zip":    valet.String().Required().Length(5),
		}),
	}

	if err := valet.Validate(data, schema); err != nil {
		fmt.Println("Validation failed:", err.Errors)
	} else {
		fmt.Println("Validation passed!")
	}
	// Output: Validation passed!
}

func Example_arrayValidation() {
	data := map[string]any{
		"tags": []any{"go", "validation", "api"},
		"items": []any{
			map[string]any{"name": "Item 1", "price": float64(10.99)},
			map[string]any{"name": "Item 2", "price": float64(20.50)},
		},
	}

	schema := valet.Schema{
		"tags": valet.Array().Required().Min(1).Of(valet.String().Min(2)),
		"items": valet.Array().Required().Min(1).Of(valet.Object().Shape(valet.Schema{
			"name":  valet.String().Required(),
			"price": valet.Float().Required().Positive(),
		})),
	}

	if err := valet.Validate(data, schema); err != nil {
		fmt.Println("Validation failed:", err.Errors)
	} else {
		fmt.Println("Validation passed!")
	}
	// Output: Validation passed!
}

func Example_conditionalValidation() {
	data := map[string]any{
		"payment_type": "card",
		"card_number":  "1234567890123456",
	}

	schema := valet.Schema{
		"payment_type": valet.String().Required().In("card", "bank", "crypto"),
		"card_number": valet.String().
			RequiredIf(func(d valet.DataObject) bool {
				return d["payment_type"] == "card"
			}).
			Regex(`^\d{16}$`),
	}

	if err := valet.Validate(data, schema); err != nil {
		fmt.Println("Validation failed:", err.Errors)
	} else {
		fmt.Println("Validation passed!")
	}
	// Output: Validation passed!
}

func Example_databaseValidation() {
	// Create a mock database checker
	checker := valet.FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []valet.WhereClause) (map[any]bool, error) {
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

	data := map[string]any{
		"user_id": float64(1),
	}

	schema := valet.Schema{
		"user_id": valet.Float().Required().Exists("users", "id"),
	}

	if err := valet.ValidateWithDB(context.Background(), data, schema, checker); err != nil {
		fmt.Println("Validation failed:", err.Errors)
	} else {
		fmt.Println("Validation passed!")
	}
	// Output: Validation passed!
}

func Example_customValidation() {
	data := map[string]any{
		"password":         "SecurePass123",
		"confirm_password": "SecurePass123",
	}

	schema := valet.Schema{
		"password": valet.String().Required().Min(8),
		"confirm_password": valet.String().Required().
			Custom(func(v string, lookup valet.Lookup) error {
				password := lookup("password").String()
				if v != password {
					return fmt.Errorf("passwords do not match")
				}
				return nil
			}),
	}

	if err := valet.Validate(data, schema); err != nil {
		fmt.Println("Validation failed:", err.Errors)
	} else {
		fmt.Println("Validation passed!")
	}
	// Output: Validation passed!
}

func Example_errorHandling() {
	data := map[string]any{
		"name":  "J",          // Too short
		"email": "invalid",    // Invalid email
		"age":   float64(150), // Too high
	}

	schema := valet.Schema{
		"name":  valet.String().Required().Min(2).Max(100),
		"email": valet.String().Required().Email(),
		"age":   valet.Float().Required().Min(18).Max(120),
	}

	if err := valet.Validate(data, schema); err != nil {
		fmt.Println("Has errors:", err.HasErrors())
		fmt.Println("Error count:", len(err.Errors))
	}
	// Output:
	// Has errors: true
	// Error count: 3
}
