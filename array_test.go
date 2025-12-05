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

// Tests for new array validators

func TestArrayValidator_Contains(t *testing.T) {
	schema := Schema{"roles": Array().Required().Contains("admin")}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"contains admin", []any{"user", "admin", "guest"}, false},
		{"only admin", []any{"admin"}, false},
		{"missing admin", []any{"user", "guest"}, true},
		{"empty array", []any{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"roles": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Contains(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Contains_Multiple(t *testing.T) {
	schema := Schema{"permissions": Array().Required().Contains("read", "write")}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"contains both", []any{"read", "write", "delete"}, false},
		{"missing write", []any{"read", "execute"}, true},
		{"missing read", []any{"write", "execute"}, true},
		{"missing both", []any{"execute", "delete"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"permissions": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Contains(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Contains_Numbers(t *testing.T) {
	schema := Schema{"numbers": Array().Required().Contains(float64(1), float64(2))}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"contains both", []any{float64(1), float64(2), float64(3)}, false},
		{"missing 2", []any{float64(1), float64(3)}, true},
		{"missing 1", []any{float64(2), float64(3)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"numbers": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Contains(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_DoesntContain(t *testing.T) {
	schema := Schema{"words": Array().Required().DoesntContain("forbidden")}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"no forbidden words", []any{"hello", "world"}, false},
		{"contains forbidden", []any{"hello", "forbidden", "world"}, true},
		{"only forbidden", []any{"forbidden"}, true},
		{"empty array", []any{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"words": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoesntContain(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_DoesntContain_Multiple(t *testing.T) {
	schema := Schema{"tags": Array().Required().DoesntContain("spam", "nsfw", "banned")}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"all clean", []any{"tech", "news", "sports"}, false},
		{"contains spam", []any{"tech", "spam"}, true},
		{"contains nsfw", []any{"nsfw", "art"}, true},
		{"contains banned", []any{"banned"}, true},
		{"contains multiple forbidden", []any{"spam", "nsfw"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"tags": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoesntContain(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Distinct(t *testing.T) {
	// Distinct is an alias for Unique
	schema := Schema{"items": Array().Required().Distinct()}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"all unique", []any{"a", "b", "c"}, false},
		{"has duplicate", []any{"a", "b", "a"}, true},
		{"multiple duplicates", []any{"a", "a", "b", "b"}, true},
		{"single element", []any{"a"}, false},
		{"empty array", []any{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"items": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Distinct(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Contains_DoesntContain_Combined(t *testing.T) {
	schema := Schema{
		"permissions": Array().Required().
			Contains("read").
			DoesntContain("admin", "superuser"),
	}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"valid - has read, no forbidden", []any{"read", "write"}, false},
		{"invalid - missing read", []any{"write", "execute"}, true},
		{"invalid - has admin", []any{"read", "admin"}, true},
		{"invalid - has superuser", []any{"read", "superuser"}, true},
		{"invalid - missing read and has admin", []any{"admin", "write"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"permissions": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Combined(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestArrayValidator_Contains_WithCustomMessage(t *testing.T) {
	schema := Schema{
		"roles": Array().Required().Contains("admin").
			Message("contains", "Must include admin role"),
	}

	err := Validate(DataObject{"roles": []any{"user", "guest"}}, schema)
	if err == nil {
		t.Error("Expected error for missing admin")
	}
	if err != nil && len(err.Errors["roles"]) > 0 {
		if err.Errors["roles"][0] != "Must include admin role" {
			t.Errorf("Expected custom message, got: %s", err.Errors["roles"][0])
		}
	}
}

func TestArrayValidator_DoesntContain_WithCustomMessage(t *testing.T) {
	schema := Schema{
		"tags": Array().Required().DoesntContain("spam").
			Message("doesntContain", "Tags cannot include spam"),
	}

	err := Validate(DataObject{"tags": []any{"tech", "spam"}}, schema)
	if err == nil {
		t.Error("Expected error for containing spam")
	}
	if err != nil && len(err.Errors["tags"]) > 0 {
		if err.Errors["tags"][0] != "Tags cannot include spam" {
			t.Errorf("Expected custom message, got: %s", err.Errors["tags"][0])
		}
	}
}

func TestArrayValidator_All_Combined(t *testing.T) {
	schema := Schema{
		"items": Array().Required().
			Min(2).Max(10).
			Distinct().
			Contains("required-item").
			DoesntContain("forbidden-item").
			Of(String().Min(3)),
	}

	tests := []struct {
		name    string
		value   []any
		wantErr bool
	}{
		{"valid", []any{"required-item", "another", "valid"}, false},
		{"too few items", []any{"required-item"}, true},
		{"missing required", []any{"item1", "item2", "item3"}, true},
		{"has forbidden", []any{"required-item", "forbidden-item"}, true},
		{"has duplicate", []any{"required-item", "required-item"}, true},
		{"element too short", []any{"required-item", "ab"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"items": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllCombined(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
