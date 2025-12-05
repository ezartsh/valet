package valet

import (
	"testing"
)

// Tests for EnumValidator

func TestEnumValidator_StringEnum(t *testing.T) {
	schema := Schema{"status": Enum("pending", "active", "completed")}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - pending", "pending", false},
		{"valid - active", "active", false},
		{"valid - completed", "completed", false},
		{"invalid - unknown", "unknown", true},
		{"invalid - empty", "", true},
		{"invalid - nil", nil, false}, // nil is allowed unless Required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"status": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enum(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestEnumValidator_Required(t *testing.T) {
	schema := Schema{"status": Enum("pending", "active").Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - pending", "pending", false},
		{"valid - active", "active", false},
		{"invalid - nil", nil, true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"status": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enum.Required(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestEnumValidator_IntEnum(t *testing.T) {
	schema := Schema{"priority": EnumInt(1, 2, 3, 4, 5)}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - 1", float64(1), false},
		{"valid - 3", float64(3), false},
		{"valid - 5", float64(5), false},
		{"invalid - 0", float64(0), true},
		{"invalid - 6", float64(6), true},
		{"invalid - negative", float64(-1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"priority": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnumInt(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestEnumValidator_CustomMessage(t *testing.T) {
	schema := Schema{
		"status": Enum("active", "inactive").Required().
			Message("enum", "Status must be either active or inactive"),
	}

	err := Validate(DataObject{"status": "unknown"}, schema)
	if err == nil {
		t.Error("Expected error for invalid status")
	}
	if err != nil && len(err.Errors["status"]) > 0 {
		if err.Errors["status"][0] != "Status must be either active or inactive" {
			t.Errorf("Expected custom message, got: %s", err.Errors["status"][0])
		}
	}
}

// Tests for LiteralValidator

func TestLiteralValidator_String(t *testing.T) {
	schema := Schema{"type": Literal("user")}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - exact match", "user", false},
		{"invalid - different value", "admin", true},
		{"invalid - nil", nil, false}, // nil allowed unless Required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"type": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Literal(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestLiteralValidator_Number(t *testing.T) {
	schema := Schema{"version": Literal(float64(1))}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - exact match", float64(1), false},
		{"invalid - different number", float64(2), true},
		{"invalid - zero", float64(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"version": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Literal(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestLiteralValidator_Boolean(t *testing.T) {
	schema := Schema{"enabled": Literal(true)}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - true", true, false},
		{"invalid - false", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"enabled": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Literal(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestLiteralValidator_Required(t *testing.T) {
	schema := Schema{"type": Literal("config").Required()}

	t.Run("nil value", func(t *testing.T) {
		err := Validate(DataObject{"type": nil}, schema)
		if err == nil {
			t.Error("Expected error for nil value")
		}
	})

	t.Run("valid value", func(t *testing.T) {
		err := Validate(DataObject{"type": "config"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestLiteralValidator_CustomMessage(t *testing.T) {
	schema := Schema{
		"type": Literal("admin").Required().
			Message("literal", "Type must be exactly 'admin'"),
	}

	err := Validate(DataObject{"type": "user"}, schema)
	if err == nil {
		t.Error("Expected error for non-matching value")
	}
	if err != nil && len(err.Errors["type"]) > 0 {
		if err.Errors["type"][0] != "Type must be exactly 'admin'" {
			t.Errorf("Expected custom message, got: %s", err.Errors["type"][0])
		}
	}
}

// Tests for UnionValidator

func TestUnionValidator_StringOrNumber(t *testing.T) {
	schema := Schema{"value": Union(String(), Float())}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - string", "hello", false},
		{"valid - number", float64(42), false},
		{"invalid - boolean", true, true},
		{"invalid - array", []any{1, 2, 3}, true},
		{"valid - nil", nil, false}, // nil allowed unless Required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"value": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Union(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestUnionValidator_WithConstraints(t *testing.T) {
	// Either an email string or a positive integer
	schema := Schema{
		"contact": Union(
			String().Email(),
			Int().Positive(),
		),
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - email", "test@example.com", false},
		{"valid - positive int", float64(123), false},
		{"invalid - non-email string", "hello", true},
		{"invalid - negative int", float64(-5), true},
		{"invalid - zero", float64(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"contact": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Union with constraints(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestUnionValidator_Literals(t *testing.T) {
	// Discriminated union: type must be "create" or "update" or "delete"
	schema := Schema{
		"action": Union(
			Literal("create"),
			Literal("update"),
			Literal("delete"),
		),
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - create", "create", false},
		{"valid - update", "update", false},
		{"valid - delete", "delete", false},
		{"invalid - read", "read", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"action": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Union of Literals(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestUnionValidator_Required(t *testing.T) {
	schema := Schema{"value": Union(String(), Int()).Required()}

	t.Run("nil value", func(t *testing.T) {
		err := Validate(DataObject{"value": nil}, schema)
		if err == nil {
			t.Error("Expected error for nil value")
		}
	})

	t.Run("valid string", func(t *testing.T) {
		err := Validate(DataObject{"value": "test"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestUnionValidator_CustomMessage(t *testing.T) {
	schema := Schema{
		"id": Union(String().UUID(), Int().Positive()).
			Message("union", "ID must be either a UUID or positive integer"),
	}

	err := Validate(DataObject{"id": "not-a-uuid"}, schema)
	if err == nil {
		t.Error("Expected error for invalid id")
	}
	if err != nil && len(err.Errors["id"]) > 0 {
		if err.Errors["id"][0] != "ID must be either a UUID or positive integer" {
			t.Errorf("Expected custom message, got: %s", err.Errors["id"][0])
		}
	}
}

// Tests for Optional wrapper

func TestOptional_String(t *testing.T) {
	schema := Schema{
		"name":     String().Required(),
		"nickname": Optional(String().Min(2).Max(20)),
	}

	tests := []struct {
		name    string
		data    DataObject
		wantErr bool
	}{
		{
			"with optional field",
			DataObject{"name": "John", "nickname": "Johnny"},
			false,
		},
		{
			"without optional field",
			DataObject{"name": "John"},
			false,
		},
		{
			"with nil optional field",
			DataObject{"name": "John", "nickname": nil},
			false,
		},
		{
			"with invalid optional field",
			DataObject{"name": "John", "nickname": "X"}, // too short
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Optional(%v) got error = %v, wantErr = %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

func TestOptional_Number(t *testing.T) {
	schema := Schema{
		"price":    Float().Required().Positive(),
		"discount": Optional(Float().Min(0).Max(100)),
	}

	tests := []struct {
		name    string
		data    DataObject
		wantErr bool
	}{
		{
			"with optional discount",
			DataObject{"price": float64(100), "discount": float64(10)},
			false,
		},
		{
			"without optional discount",
			DataObject{"price": float64(100)},
			false,
		},
		{
			"invalid discount - over max",
			DataObject{"price": float64(100), "discount": float64(150)},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Optional(%v) got error = %v, wantErr = %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

func TestOptional_Object(t *testing.T) {
	addressSchema := Object().Shape(Schema{
		"street": String().Required(),
		"city":   String().Required(),
	})

	schema := Schema{
		"name":    String().Required(),
		"address": Optional(addressSchema),
	}

	tests := []struct {
		name    string
		data    DataObject
		wantErr bool
	}{
		{
			"with valid address",
			DataObject{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
					"city":   "NYC",
				},
			},
			false,
		},
		{
			"without address",
			DataObject{"name": "John"},
			false,
		},
		{
			"with invalid address - missing city",
			DataObject{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Optional Object(%v) got error = %v, wantErr = %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

func TestOptional_Array(t *testing.T) {
	schema := Schema{
		"name": String().Required(),
		"tags": Optional(Array().Min(1).Of(String())),
	}

	tests := []struct {
		name    string
		data    DataObject
		wantErr bool
	}{
		{
			"with valid tags",
			DataObject{"name": "Product", "tags": []any{"tech", "new"}},
			false,
		},
		{
			"without tags",
			DataObject{"name": "Product"},
			false,
		},
		{
			"with empty tags - fails min",
			DataObject{"name": "Product", "tags": []any{}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Optional Array(%v) got error = %v, wantErr = %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

// Tests for combined schema helpers

func TestSchemaHelpers_Combined(t *testing.T) {
	// Complex schema using multiple helpers
	schema := Schema{
		"type":   Enum("user", "admin", "guest").Required(),
		"status": Literal("active"),
		"id":     Union(String().UUID(), Int().Positive()).Required(),
		"meta":   Optional(Object().Shape(Schema{"version": Int()})),
	}

	tests := []struct {
		name    string
		data    DataObject
		wantErr bool
	}{
		{
			"valid - all fields",
			DataObject{
				"type":   "admin",
				"status": "active",
				"id":     "550e8400-e29b-41d4-a716-446655440000",
				"meta":   map[string]any{"version": float64(1)},
			},
			false,
		},
		{
			"valid - integer id",
			DataObject{
				"type":   "user",
				"status": "active",
				"id":     float64(123),
			},
			false,
		},
		{
			"invalid - wrong type",
			DataObject{
				"type":   "superuser",
				"status": "active",
				"id":     float64(1),
			},
			true,
		},
		{
			"invalid - wrong status",
			DataObject{
				"type":   "user",
				"status": "inactive",
				"id":     float64(1),
			},
			true,
		},
		{
			"invalid - invalid id",
			DataObject{
				"type":   "user",
				"status": "active",
				"id":     "not-a-uuid",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Combined(%v) got error = %v, wantErr = %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

func TestEnumValidator_In_Alias(t *testing.T) {
	// Test that In() is an alias for the initial values
	schema := Schema{"role": Enum[string]().In("admin", "user", "guest").Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid - admin", "admin", false},
		{"valid - user", "user", false},
		{"invalid - unknown", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"role": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enum.In(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
