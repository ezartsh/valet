package valet

import (
	"errors"
	"testing"
)

func TestNumberValidator_Required(t *testing.T) {
	schema := Schema{"age": Float().Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"nil value", nil, true},
		{"valid number", float64(25), false},
		{"zero value", float64(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"age": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_RequiredIf(t *testing.T) {
	schema := Schema{
		"hasDiscount": Bool(),
		"discount": Float().RequiredIf(func(data DataObject) bool {
			return data["hasDiscount"] == true
		}),
	}

	t.Run("condition met - value missing", func(t *testing.T) {
		err := Validate(DataObject{"hasDiscount": true}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}
	})

	t.Run("condition met - value present", func(t *testing.T) {
		err := Validate(DataObject{"hasDiscount": true, "discount": float64(10)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		err := Validate(DataObject{"hasDiscount": false}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_RequiredUnless(t *testing.T) {
	schema := Schema{
		"type": String().Required(),
		"price": Float().RequiredUnless(func(data DataObject) bool {
			return data["type"] == "free"
		}),
	}

	t.Run("condition met - value not required", func(t *testing.T) {
		err := Validate(DataObject{"type": "free"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met - value required", func(t *testing.T) {
		err := Validate(DataObject{"type": "paid"}, schema)
		if err == nil {
			t.Error("Expected error when condition not met")
		}
	})
}

func TestNumberValidator_MinMax(t *testing.T) {
	schema := Schema{"age": Float().Required().Min(18).Max(120)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"too small", 15, true},
		{"min boundary", 18, false},
		{"valid", 25, false},
		{"max boundary", 120, false},
		{"too large", 150, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"age": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_MinDigits(t *testing.T) {
	schema := Schema{"code": Float().Required().MinDigits(4)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"too few digits", 123, true},
		{"exact digits", 1234, false},
		{"more digits", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"code": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_MaxDigits(t *testing.T) {
	schema := Schema{"pin": Float().Required().MaxDigits(4)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"fewer digits", 123, false},
		{"exact digits", 1234, false},
		{"too many digits", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"pin": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Positive(t *testing.T) {
	schema := Schema{"amount": Float().Required().Positive()}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"positive", 10, false},
		{"zero", 0, true},
		{"negative", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"amount": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Negative(t *testing.T) {
	schema := Schema{"temperature": Float().Required().Negative()}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"negative", -10, false},
		{"zero", 0, true},
		{"positive", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"temperature": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Integer(t *testing.T) {
	schema := Schema{"count": Float().Required().Integer()}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"integer", 10, false},
		{"decimal", 10.5, true},
		{"negative integer", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"count": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_MultipleOf(t *testing.T) {
	schema := Schema{"quantity": Float().Required().MultipleOf(5)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"multiple of 5 - 10", 10, false},
		{"multiple of 5 - 15", 15, false},
		{"multiple of 5 - 0", 0, false},
		{"not multiple - 12", 12, true},
		{"not multiple - 7", 7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"quantity": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_In(t *testing.T) {
	schema := Schema{"rating": Float().Required().In(1, 2, 3, 4, 5)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid rating", 3, false},
		{"invalid rating", 6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"rating": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_NotIn(t *testing.T) {
	schema := Schema{"number": Float().Required().NotIn(0, 13, 666)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid number", 7, false},
		{"unlucky 13", 13, true},
		{"zero", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"number": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Regex(t *testing.T) {
	schema := Schema{"year": Float().Required().Regex(`^20\d{2}$`)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid year", 2024, false},
		{"invalid year", 1999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"year": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_NotRegex(t *testing.T) {
	schema := Schema{"code": Float().Required().NotRegex(`^0`)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid code", 123, false},
		{"starts with 0", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"code": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Coerce(t *testing.T) {
	schema := Schema{"age": Float().Coerce().Min(18)}

	t.Run("string to number", func(t *testing.T) {
		err := Validate(DataObject{"age": "25"}, schema)
		if err != nil {
			t.Errorf("Expected no error with coercion, got: %v", err.Errors)
		}
	})

	t.Run("invalid string", func(t *testing.T) {
		err := Validate(DataObject{"age": "not a number"}, schema)
		if err == nil {
			t.Error("Expected error for invalid string")
		}
	})
}

func TestNumberValidator_Nullable(t *testing.T) {
	schema := Schema{"score": Float().Nullable()}

	t.Run("null value allowed", func(t *testing.T) {
		err := Validate(DataObject{"score": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_Default(t *testing.T) {
	schema := Schema{"quantity": Float().Default(1).Min(1)}

	t.Run("nil uses default", func(t *testing.T) {
		err := Validate(DataObject{"quantity": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error with default, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_Custom(t *testing.T) {
	schema := Schema{
		"even": Float().Required().Custom(func(v float64, lookup Lookup) error {
			if int(v)%2 != 0 {
				return errors.New("must be even")
			}
			return nil
		}),
	}

	t.Run("even number", func(t *testing.T) {
		err := Validate(DataObject{"even": float64(4)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("odd number", func(t *testing.T) {
		err := Validate(DataObject{"even": float64(3)}, schema)
		if err == nil {
			t.Error("Expected error for odd number")
		}
	})
}

func TestNumberValidator_TypeCheck(t *testing.T) {
	schema := Schema{"age": Float().Required()}

	t.Run("non-number type", func(t *testing.T) {
		err := Validate(DataObject{"age": "twenty"}, schema)
		if err == nil {
			t.Error("Expected error for non-number type")
		}
	})
}

func TestNumberValidator_Message(t *testing.T) {
	schema := Schema{
		"age": Float().Required().Min(18).
			Message("required", "Age is required").
			Message("min", "Must be at least 18"),
	}

	t.Run("custom min message", func(t *testing.T) {
		err := Validate(DataObject{"age": float64(15)}, schema)
		if err == nil || err.Errors["age"][0] != "Must be at least 18" {
			t.Error("Expected custom min message")
		}
	})
}

func TestNumberValidator_Int(t *testing.T) {
	schema := Schema{"count": Int().Required().Min(0)}

	t.Run("valid int", func(t *testing.T) {
		err := Validate(DataObject{"count": float64(10)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestNumberValidator_GenericNum(t *testing.T) {
	schema := Schema{"value": Num[int64]().Required().Min(0)}

	t.Run("valid value", func(t *testing.T) {
		err := Validate(DataObject{"value": int64(10)}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

// Tests for new number validators

func TestNumberValidator_Between(t *testing.T) {
	schema := Schema{"score": Float().Required().Between(0, 100)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"below min", -1, true},
		{"at min", 0, false},
		{"in range", 50, false},
		{"at max", 100, false},
		{"above max", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"score": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Between(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Between_Integer(t *testing.T) {
	schema := Schema{"age": Int().Required().Between(18, 65)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"below min", 17, true},
		{"at min", 18, false},
		{"in range", 30, false},
		{"at max", 65, false},
		{"above max", 66, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"age": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Between(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Step(t *testing.T) {
	schema := Schema{"quantity": Float().Required().Step(5)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid - 0", 0, false},
		{"valid - 5", 5, false},
		{"valid - 10", 10, false},
		{"valid - 15", 15, false},
		{"invalid - 3", 3, true},
		{"invalid - 7", 7, true},
		{"invalid - 12", 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"quantity": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Step(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Step_Float(t *testing.T) {
	schema := Schema{"price": Float().Required().Step(0.25)}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid - 0", 0, false},
		{"valid - 0.25", 0.25, false},
		{"valid - 0.50", 0.50, false},
		{"valid - 1.00", 1.00, false},
		{"valid - 2.75", 2.75, false},
		{"invalid - 0.10", 0.10, true},
		{"invalid - 0.30", 0.30, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"price": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Step(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_Between_WithMessages(t *testing.T) {
	schema := Schema{
		"score": Float().Required().Between(0, 100).
			Message("min", "Score must be at least 0").
			Message("max", "Score must be at most 100"),
	}

	t.Run("below min with custom message", func(t *testing.T) {
		err := Validate(DataObject{"score": float64(-5)}, schema)
		if err == nil || err.Errors["score"][0] != "Score must be at least 0" {
			t.Error("Expected custom min message")
		}
	})

	t.Run("above max with custom message", func(t *testing.T) {
		err := Validate(DataObject{"score": float64(150)}, schema)
		if err == nil || err.Errors["score"][0] != "Score must be at most 100" {
			t.Error("Expected custom max message")
		}
	})
}

func TestNumberValidator_Combined(t *testing.T) {
	schema := Schema{
		"value": Float().Required().Between(0, 100).Step(5).Positive(),
	}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"valid - 0", 0, true}, // fails positive
		{"valid - 5", 5, false},
		{"valid - 50", 50, false},
		{"valid - 100", 100, false},
		{"invalid - negative", -5, true},
		{"invalid - not step", 7, true},
		{"invalid - above max", 105, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"value": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Combined(%v) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
