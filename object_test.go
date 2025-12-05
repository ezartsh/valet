package valet

import (
	"errors"
	"testing"
)

func TestObjectValidator_Required(t *testing.T) {
	schema := Schema{"user": Object().Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"nil value", nil, true},
		{"empty object", map[string]any{}, false},
		{"valid object", map[string]any{"name": "John"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"user": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectValidator_RequiredIf(t *testing.T) {
	schema := Schema{
		"hasProfile": Bool(),
		"profile": Object().RequiredIf(func(data DataObject) bool {
			return data["hasProfile"] == true
		}),
	}

	t.Run("condition met - value missing", func(t *testing.T) {
		err := Validate(DataObject{"hasProfile": true}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}
	})

	t.Run("condition met - value present", func(t *testing.T) {
		err := Validate(DataObject{"hasProfile": true, "profile": map[string]any{}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		err := Validate(DataObject{"hasProfile": false}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator_RequiredUnless(t *testing.T) {
	schema := Schema{
		"anonymous": Bool(),
		"profile": Object().RequiredUnless(func(data DataObject) bool {
			return data["anonymous"] == true
		}),
	}

	t.Run("condition met - value not required", func(t *testing.T) {
		err := Validate(DataObject{"anonymous": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met - value required", func(t *testing.T) {
		err := Validate(DataObject{"anonymous": false}, schema)
		if err == nil {
			t.Error("Expected error when condition not met")
		}
	})
}

func TestObjectValidator_Shape(t *testing.T) {
	schema := Schema{
		"user": Object().Required().Shape(Schema{
			"name":  String().Required().Min(2),
			"email": String().Required().Email(),
			"age":   Float().Min(0),
		}),
	}

	t.Run("valid nested object", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
				"age":   float64(25),
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("invalid nested fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "J",
				"email": "invalid",
			},
		}, schema)
		if err == nil {
			t.Error("Expected errors for invalid nested fields")
		}
	})

	t.Run("missing required nested field", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"email": "john@example.com",
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for missing required field")
		}
	})
}

func TestObjectValidator_Item(t *testing.T) {
	// Item is an alias for Shape
	schema := Schema{
		"config": Object().Required().Item(Schema{
			"key": String().Required(),
		}),
	}

	t.Run("valid with Item", func(t *testing.T) {
		err := Validate(DataObject{
			"config": map[string]any{"key": "value"},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator_Strict(t *testing.T) {
	schema := Schema{
		"config": Object().Required().Strict().Shape(Schema{
			"name": String().Required(),
		}),
	}

	t.Run("no unknown keys", func(t *testing.T) {
		err := Validate(DataObject{
			"config": map[string]any{"name": "test"},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("unknown keys present", func(t *testing.T) {
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
}

func TestObjectValidator_Passthrough(t *testing.T) {
	schema := Schema{
		"config": Object().Required().Passthrough().Shape(Schema{
			"name": String().Required(),
		}),
	}

	t.Run("unknown keys allowed", func(t *testing.T) {
		err := Validate(DataObject{
			"config": map[string]any{
				"name":    "test",
				"unknown": "field",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error in passthrough mode, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator_Extend(t *testing.T) {
	baseUser := Object().Shape(Schema{
		"name": String().Required(),
	})

	adminUser := baseUser.Extend(Schema{
		"role": String().Required().In("admin", "superadmin"),
	})

	t.Run("extended schema", func(t *testing.T) {
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

	t.Run("missing extended field", func(t *testing.T) {
		err := Validate(DataObject{
			"admin": map[string]any{
				"name": "John",
			},
		}, Schema{"admin": adminUser})
		if err == nil {
			t.Error("Expected error for missing role")
		}
	})

	t.Run("invalid extended field", func(t *testing.T) {
		err := Validate(DataObject{
			"admin": map[string]any{
				"name": "John",
				"role": "user",
			},
		}, Schema{"admin": adminUser})
		if err == nil {
			t.Error("Expected error for invalid role")
		}
	})
}

func TestObjectValidator_Nullable(t *testing.T) {
	schema := Schema{"metadata": Object().Nullable()}

	t.Run("null value allowed", func(t *testing.T) {
		err := Validate(DataObject{"metadata": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})

	t.Run("valid object", func(t *testing.T) {
		err := Validate(DataObject{"metadata": map[string]any{"key": "value"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator_Custom(t *testing.T) {
	schema := Schema{
		"user": Object().Required().Custom(func(v DataObject, lookup Lookup) error {
			if v["password"] != v["confirmPassword"] {
				return errors.New("passwords do not match")
			}
			return nil
		}),
	}

	t.Run("custom validation fails", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"password":        "secret",
				"confirmPassword": "different",
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for mismatched passwords")
		}
	})

	t.Run("custom validation passes", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"password":        "secret",
				"confirmPassword": "secret",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator_TypeCheck(t *testing.T) {
	schema := Schema{"user": Object().Required()}

	t.Run("non-object type", func(t *testing.T) {
		err := Validate(DataObject{"user": "not an object"}, schema)
		if err == nil {
			t.Error("Expected error for non-object type")
		}
	})

	t.Run("array type", func(t *testing.T) {
		err := Validate(DataObject{"user": []any{1, 2, 3}}, schema)
		if err == nil {
			t.Error("Expected error for array type")
		}
	})
}

func TestObjectValidator_Message(t *testing.T) {
	schema := Schema{
		"user": Object().Required().
			Message("required", "User object is required").
			Message("type", "User must be an object"),
	}

	t.Run("custom required message", func(t *testing.T) {
		err := Validate(DataObject{"user": nil}, schema)
		if err == nil || err.Errors["user"][0] != "User object is required" {
			t.Error("Expected custom required message")
		}
	})
}

func TestObjectValidator_DeepNesting(t *testing.T) {
	schema := Schema{
		"level1": Object().Required().Shape(Schema{
			"level2": Object().Required().Shape(Schema{
				"level3": Object().Required().Shape(Schema{
					"value": String().Required(),
				}),
			}),
		}),
	}

	t.Run("valid deep nesting", func(t *testing.T) {
		err := Validate(DataObject{
			"level1": map[string]any{
				"level2": map[string]any{
					"level3": map[string]any{
						"value": "deep",
					},
				},
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("missing deep field", func(t *testing.T) {
		err := Validate(DataObject{
			"level1": map[string]any{
				"level2": map[string]any{
					"level3": map[string]any{},
				},
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for missing deep field")
		}
	})
}

// Tests for new object utilities

func TestObjectValidator_Pick(t *testing.T) {
	fullSchema := Object().Shape(Schema{
		"name":    String().Required(),
		"email":   String().Required().Email(),
		"age":     Int().Required().Min(0),
		"address": String(),
	})

	pickedSchema := Schema{"user": fullSchema.Pick("name", "email")}

	t.Run("valid picked fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
			},
		}, pickedSchema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("missing picked required field", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, pickedSchema)
		if err == nil {
			t.Error("Expected error for missing email")
		}
	})

	t.Run("extra fields ignored", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
				"age":   float64(30), // This field is not in picked schema
			},
		}, pickedSchema)
		if err != nil {
			t.Errorf("Expected no error with extra fields, got: %v", err.Errors)
		}
	})
}

func TestObjectValidator_Omit(t *testing.T) {
	fullSchema := Object().Shape(Schema{
		"name":     String().Required(),
		"email":    String().Required().Email(),
		"password": String().Required().Min(8),
		"age":      Int().Min(0),
	})

	omittedSchema := Schema{"user": fullSchema.Omit("password")}

	t.Run("valid without omitted field", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
			},
		}, omittedSchema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("still validates other required fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
				// missing email
			},
		}, omittedSchema)
		if err == nil {
			t.Error("Expected error for missing email")
		}
	})
}

func TestObjectValidator_Partial(t *testing.T) {
	fullSchema := Object().Shape(Schema{
		"name":  String().Required(),
		"email": String().Required().Email(),
		"age":   Int().Required().Min(0),
	})

	partialSchema := Schema{"user": fullSchema.Partial()}

	t.Run("valid with all fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
				"age":   float64(30),
			},
		}, partialSchema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("valid with no fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{},
		}, partialSchema)
		if err != nil {
			t.Errorf("Expected no error with empty object, got: %v", err.Errors)
		}
	})

	t.Run("valid with some fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, partialSchema)
		if err != nil {
			t.Errorf("Expected no error with partial fields, got: %v", err.Errors)
		}
	})

	t.Run("still validates format when present", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"email": "invalid-email",
			},
		}, partialSchema)
		if err == nil {
			t.Error("Expected error for invalid email format")
		}
	})
}

func TestObjectValidator_Merge(t *testing.T) {
	baseSchema := Object().Shape(Schema{
		"name": String().Required(),
		"age":  Int().Min(0),
	})

	extendedFields := Object().Shape(Schema{
		"email": String().Required().Email(),
		"phone": String(),
	})

	mergedSchema := Schema{"user": baseSchema.Merge(extendedFields)}

	t.Run("valid with all fields", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name":  "John",
				"email": "john@example.com",
				"age":   float64(30),
				"phone": "123-456-7890",
			},
		}, mergedSchema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("missing required from base", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"email": "john@example.com",
			},
		}, mergedSchema)
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})

	t.Run("missing required from extension", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"name": "John",
			},
		}, mergedSchema)
		if err == nil {
			t.Error("Expected error for missing email")
		}
	})
}

func TestObjectValidator_Pick_Empty(t *testing.T) {
	fullSchema := Object().Shape(Schema{
		"name":  String().Required(),
		"email": String().Required(),
	})

	// Pick with no fields should create empty schema
	pickedSchema := Schema{"user": fullSchema.Pick()}

	err := Validate(DataObject{
		"user": map[string]any{},
	}, pickedSchema)
	if err != nil {
		t.Errorf("Expected no error with empty pick, got: %v", err.Errors)
	}
}

func TestObjectValidator_Omit_All(t *testing.T) {
	fullSchema := Object().Shape(Schema{
		"name":  String().Required(),
		"email": String().Required(),
	})

	// Omit all fields
	omittedSchema := Schema{"user": fullSchema.Omit("name", "email")}

	err := Validate(DataObject{
		"user": map[string]any{},
	}, omittedSchema)
	if err != nil {
		t.Errorf("Expected no error with all omitted, got: %v", err.Errors)
	}
}

func TestObjectValidator_Chained_Utilities(t *testing.T) {
	baseSchema := Object().Shape(Schema{
		"id":       Int().Required(),
		"name":     String().Required(),
		"email":    String().Required().Email(),
		"password": String().Required().Min(8),
		"role":     String().Required(),
	})

	// Create a "safe user" schema by omitting password and picking only certain fields
	// First omit password, then pick the fields we want
	safeUserSchema := baseSchema.Omit("password")

	schema := Schema{"user": safeUserSchema}

	t.Run("validates without password", func(t *testing.T) {
		err := Validate(DataObject{
			"user": map[string]any{
				"id":    float64(1),
				"name":  "John",
				"email": "john@example.com",
				"role":  "admin",
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}
