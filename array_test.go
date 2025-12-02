package valet

import (
	"errors"
	"testing"
)

func TestArrayValidator_Required(t *testing.T) {
	schema := Schema{"tags": Array().Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"nil value", nil, true},
		{"empty array", []any{}, false},
		{"valid array", []any{"a", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"tags": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_RequiredIf(t *testing.T) {
	schema := Schema{
		"hasItems": Bool(),
		"items": Array().RequiredIf(func(data DataObject) bool {
			return data["hasItems"] == true
		}),
	}

	t.Run("condition met - value missing", func(t *testing.T) {
		err := Validate(DataObject{"hasItems": true}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}
	})

	t.Run("condition met - value present", func(t *testing.T) {
		err := Validate(DataObject{"hasItems": true, "items": []any{"item1"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		err := Validate(DataObject{"hasItems": false}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestArrayValidator_RequiredUnless(t *testing.T) {
	schema := Schema{
		"isEmpty": Bool(),
		"items": Array().RequiredUnless(func(data DataObject) bool {
			return data["isEmpty"] == true
		}),
	}

	t.Run("condition met - value not required", func(t *testing.T) {
		err := Validate(DataObject{"isEmpty": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met - value required", func(t *testing.T) {
		err := Validate(DataObject{"isEmpty": false}, schema)
		if err == nil {
			t.Error("Expected error when condition not met")
		}
	})
}

func TestArrayValidator_Min(t *testing.T) {
	schema := Schema{"tags": Array().Required().Min(2)}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"too few", []any{"a"}, true},
		{"min boundary", []any{"a", "b"}, false},
		{"more than min", []any{"a", "b", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"tags": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Max(t *testing.T) {
	schema := Schema{"tags": Array().Required().Max(3)}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"fewer than max", []any{"a", "b"}, false},
		{"max boundary", []any{"a", "b", "c"}, false},
		{"too many", []any{"a", "b", "c", "d"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"tags": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Length(t *testing.T) {
	schema := Schema{"coordinates": Array().Required().Length(2)}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"too few", []any{1}, true},
		{"exact length", []any{1, 2}, false},
		{"too many", []any{1, 2, 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"coordinates": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Of(t *testing.T) {
	t.Run("string elements", func(t *testing.T) {
		schema := Schema{"emails": Array().Required().Of(String().Email())}

		err := Validate(DataObject{
			"emails": []any{"a@example.com", "b@example.com"},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}

		err = Validate(DataObject{
			"emails": []any{"valid@example.com", "invalid"},
		}, schema)
		if err == nil {
			t.Error("Expected error for invalid email in array")
		}
	})

	t.Run("number elements", func(t *testing.T) {
		schema := Schema{"scores": Array().Required().Of(Float().Min(0).Max(100))}

		err := Validate(DataObject{
			"scores": []any{float64(85), float64(90), float64(75)},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}

		err = Validate(DataObject{
			"scores": []any{float64(85), float64(150)},
		}, schema)
		if err == nil {
			t.Error("Expected error for score > 100")
		}
	})

	t.Run("object elements", func(t *testing.T) {
		schema := Schema{
			"users": Array().Required().Of(Object().Shape(Schema{
				"name": String().Required(),
			})),
		}

		err := Validate(DataObject{
			"users": []any{
				map[string]any{"name": "John"},
				map[string]any{"name": "Jane"},
			},
		}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}

		err = Validate(DataObject{
			"users": []any{
				map[string]any{"name": "John"},
				map[string]any{},
			},
		}, schema)
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})
}

func TestArrayValidator_Unique(t *testing.T) {
	schema := Schema{"ids": Array().Required().Unique()}

	t.Run("unique values", func(t *testing.T) {
		err := Validate(DataObject{"ids": []any{1, 2, 3}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("duplicate values", func(t *testing.T) {
		err := Validate(DataObject{"ids": []any{1, 2, 2, 3}}, schema)
		if err == nil {
			t.Error("Expected error for duplicate values")
		}
	})
}

func TestArrayValidator_Nonempty(t *testing.T) {
	schema := Schema{"items": Array().Required().Nonempty()}

	t.Run("empty array", func(t *testing.T) {
		err := Validate(DataObject{"items": []any{}}, schema)
		if err == nil {
			t.Error("Expected error for empty array")
		}
	})

	t.Run("non-empty array", func(t *testing.T) {
		err := Validate(DataObject{"items": []any{"item"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestArrayValidator_Nullable(t *testing.T) {
	schema := Schema{"tags": Array().Nullable()}

	t.Run("null value allowed", func(t *testing.T) {
		err := Validate(DataObject{"tags": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})
}

func TestArrayValidator_Custom(t *testing.T) {
	schema := Schema{
		"numbers": Array().Required().Custom(func(v []any, lookup Lookup) error {
			sum := 0
			for _, n := range v {
				if num, ok := n.(float64); ok {
					sum += int(num)
				}
			}
			if sum > 100 {
				return errors.New("sum must not exceed 100")
			}
			return nil
		}),
	}

	t.Run("custom validation fails", func(t *testing.T) {
		err := Validate(DataObject{"numbers": []any{float64(50), float64(60)}}, schema)
		if err == nil {
			t.Error("Expected error for sum > 100")
		}
	})

	t.Run("custom validation passes", func(t *testing.T) {
		err := Validate(DataObject{"numbers": []any{float64(30), float64(40)}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestArrayValidator_TypeCheck(t *testing.T) {
	schema := Schema{"items": Array().Required()}

	t.Run("non-array type", func(t *testing.T) {
		err := Validate(DataObject{"items": "not an array"}, schema)
		if err == nil {
			t.Error("Expected error for non-array type")
		}
	})

	t.Run("object type", func(t *testing.T) {
		err := Validate(DataObject{"items": map[string]any{"key": "value"}}, schema)
		if err == nil {
			t.Error("Expected error for object type")
		}
	})
}

func TestArrayValidator_Message(t *testing.T) {
	schema := Schema{
		"tags": Array().Required().Min(1).
			Message("required", "Tags are required").
			Message("min", "At least one tag is required"),
	}

	t.Run("custom min message", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{}}, schema)
		if err == nil || err.Errors["tags"][0] != "At least one tag is required" {
			t.Error("Expected custom min message")
		}
	})
}

func TestArrayValidator_CombinedRules(t *testing.T) {
	schema := Schema{
		"tags": Array().Required().Min(1).Max(5).Unique().Of(String().Min(2).Max(20)),
	}

	t.Run("valid array", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{"go", "rust", "python"}}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("too many elements", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{"a", "b", "c", "d", "e", "f"}}, schema)
		if err == nil {
			t.Error("Expected error for too many elements")
		}
	})

	t.Run("duplicate elements", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{"go", "go"}}, schema)
		if err == nil {
			t.Error("Expected error for duplicates")
		}
	})

	t.Run("element too short", func(t *testing.T) {
		err := Validate(DataObject{"tags": []any{"go", "a"}}, schema)
		if err == nil {
			t.Error("Expected error for short element")
		}
	})
}
