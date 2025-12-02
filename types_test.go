package valet

import (
	"testing"
)

func TestLookupResult_String(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string value", "hello", "hello"},
		{"nil value", nil, ""},
		{"non-string", 123, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LookupResult{value: tt.value, exists: true}
			if r.String() != tt.expected {
				t.Errorf("String() = %s, want %s", r.String(), tt.expected)
			}
		})
	}
}

func TestLookupResult_Int(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected int64
	}{
		{"int value", 42, 42},
		{"int32 value", int32(42), 42},
		{"int64 value", int64(42), 42},
		{"float64 value", float64(42.9), 42},
		{"nil value", nil, 0},
		{"string value", "42", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LookupResult{value: tt.value, exists: true}
			if r.Int() != tt.expected {
				t.Errorf("Int() = %d, want %d", r.Int(), tt.expected)
			}
		})
	}
}

func TestLookupResult_Float(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected float64
	}{
		{"float64 value", float64(3.14), 3.14},
		{"float32 value", float32(3.14), float64(float32(3.14))},
		{"int value", 42, 42.0},
		{"int64 value", int64(42), 42.0},
		{"nil value", nil, 0},
		{"string value", "3.14", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LookupResult{value: tt.value, exists: true}
			if r.Float() != tt.expected {
				t.Errorf("Float() = %f, want %f", r.Float(), tt.expected)
			}
		})
	}
}

func TestLookupResult_Bool(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"true value", true, true},
		{"false value", false, false},
		{"nil value", nil, false},
		{"string value", "true", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LookupResult{value: tt.value, exists: true}
			if r.Bool() != tt.expected {
				t.Errorf("Bool() = %v, want %v", r.Bool(), tt.expected)
			}
		})
	}
}

func TestLookupResult_Exists(t *testing.T) {
	t.Run("exists true", func(t *testing.T) {
		r := LookupResult{value: "test", exists: true}
		if !r.Exists() {
			t.Error("Exists() should return true")
		}
	})

	t.Run("exists false", func(t *testing.T) {
		r := LookupResult{value: nil, exists: false}
		if r.Exists() {
			t.Error("Exists() should return false")
		}
	})
}

func TestLookupResult_Value(t *testing.T) {
	r := LookupResult{value: "test", exists: true}
	if r.Value() != "test" {
		t.Errorf("Value() = %v, want test", r.Value())
	}
}

func TestLookupResult_Get(t *testing.T) {
	t.Run("nested object", func(t *testing.T) {
		r := LookupResult{
			value: map[string]any{
				"name": "John",
				"age":  float64(30),
			},
			exists: true,
		}

		name := r.Get("name")
		if name.String() != "John" {
			t.Errorf("Get(name) = %s, want John", name.String())
		}

		age := r.Get("age")
		if age.Int() != 30 {
			t.Errorf("Get(age) = %d, want 30", age.Int())
		}
	})

	t.Run("missing key", func(t *testing.T) {
		r := LookupResult{
			value:  map[string]any{"name": "John"},
			exists: true,
		}

		missing := r.Get("missing")
		if missing.Exists() {
			t.Error("Get(missing) should not exist")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		r := LookupResult{value: nil, exists: false}
		result := r.Get("key")
		if result.Exists() {
			t.Error("Get on nil should not exist")
		}
	})

	t.Run("non-object value", func(t *testing.T) {
		r := LookupResult{value: "string", exists: true}
		result := r.Get("key")
		if result.Exists() {
			t.Error("Get on non-object should not exist")
		}
	})
}

func TestLookupResult_IsArray(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"array value", []any{1, 2, 3}, true},
		{"empty array", []any{}, true},
		{"nil value", nil, false},
		{"object value", map[string]any{}, false},
		{"string value", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LookupResult{value: tt.value, exists: true}
			if r.IsArray() != tt.expected {
				t.Errorf("IsArray() = %v, want %v", r.IsArray(), tt.expected)
			}
		})
	}
}

func TestLookupResult_IsObject(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"object value", map[string]any{"key": "value"}, true},
		{"empty object", map[string]any{}, true},
		{"nil value", nil, false},
		{"array value", []any{1, 2, 3}, false},
		{"string value", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LookupResult{value: tt.value, exists: true}
			if r.IsObject() != tt.expected {
				t.Errorf("IsObject() = %v, want %v", r.IsObject(), tt.expected)
			}
		})
	}
}

func TestLookupResult_Array(t *testing.T) {
	t.Run("array value", func(t *testing.T) {
		r := LookupResult{value: []any{1, 2, 3}, exists: true}
		arr := r.Array()
		if len(arr) != 3 {
			t.Errorf("Array() length = %d, want 3", len(arr))
		}
	})

	t.Run("nil value", func(t *testing.T) {
		r := LookupResult{value: nil, exists: false}
		arr := r.Array()
		if arr != nil {
			t.Error("Array() on nil should return nil")
		}
	})

	t.Run("non-array value", func(t *testing.T) {
		r := LookupResult{value: "string", exists: true}
		arr := r.Array()
		if arr != nil {
			t.Error("Array() on non-array should return nil")
		}
	})
}

func TestWhereClause(t *testing.T) {
	t.Run("Where", func(t *testing.T) {
		w := Where("status", "=", "active")
		if w.Column != "status" || w.Operator != "=" || w.Value != "active" {
			t.Error("Where() failed")
		}
	})

	t.Run("WhereEq", func(t *testing.T) {
		w := WhereEq("id", 1)
		if w.Column != "id" || w.Operator != "=" || w.Value != 1 {
			t.Error("WhereEq() failed")
		}
	})

	t.Run("WhereNot", func(t *testing.T) {
		w := WhereNot("deleted", true)
		if w.Column != "deleted" || w.Operator != "!=" || w.Value != true {
			t.Error("WhereNot() failed")
		}
	})
}

func TestValidationContext(t *testing.T) {
	ctx := &ValidationContext{
		RootData: DataObject{"name": "John"},
		Path:     []string{"user", "name"},
	}

	if ctx.RootData["name"] != "John" {
		t.Error("RootData not set correctly")
	}

	if len(ctx.Path) != 2 || ctx.Path[0] != "user" || ctx.Path[1] != "name" {
		t.Error("Path not set correctly")
	}
}

func TestOptions(t *testing.T) {
	opts := Options{
		AbortEarly: true,
	}

	if !opts.AbortEarly {
		t.Error("AbortEarly not set correctly")
	}
}
