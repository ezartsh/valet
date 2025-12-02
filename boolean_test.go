package valet

import (
	"errors"
	"testing"
)

func TestBoolValidator_Required(t *testing.T) {
	schema := Schema{"active": Bool().Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"nil value", nil, true},
		{"true", true, false},
		{"false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"active": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoolValidator_RequiredIf(t *testing.T) {
	schema := Schema{
		"type": String().Required(),
		"confirmed": Bool().RequiredIf(func(data DataObject) bool {
			return data["type"] == "important"
		}),
	}

	t.Run("condition met - value missing", func(t *testing.T) {
		err := Validate(DataObject{"type": "important"}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}
	})

	t.Run("condition met - value present", func(t *testing.T) {
		err := Validate(DataObject{"type": "important", "confirmed": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		err := Validate(DataObject{"type": "normal"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestBoolValidator_RequiredUnless(t *testing.T) {
	schema := Schema{
		"skip": Bool(),
		"agree": Bool().RequiredUnless(func(data DataObject) bool {
			return data["skip"] == true
		}),
	}

	t.Run("condition met - value not required", func(t *testing.T) {
		err := Validate(DataObject{"skip": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met - value required", func(t *testing.T) {
		err := Validate(DataObject{"skip": false}, schema)
		if err == nil {
			t.Error("Expected error when condition not met")
		}
	})
}

func TestBoolValidator_True(t *testing.T) {
	schema := Schema{"terms": Bool().Required().True()}

	tests := []struct {
		name    string
		value   bool
		wantErr bool
	}{
		{"true", true, false},
		{"false", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"terms": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoolValidator_False(t *testing.T) {
	schema := Schema{"disabled": Bool().Required().False()}

	tests := []struct {
		name    string
		value   bool
		wantErr bool
	}{
		{"false", false, false},
		{"true", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"disabled": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoolValidator_Coerce(t *testing.T) {
	schema := Schema{"active": Bool().Coerce()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"string true", "true", false},
		{"string false", "false", false},
		{"string 1", "1", false},
		{"string 0", "0", false},
		{"string yes", "yes", false},
		{"string no", "no", false},
		{"number 1", 1, false},
		{"number 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"active": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoolValidator_Nullable(t *testing.T) {
	schema := Schema{"flag": Bool().Nullable()}

	t.Run("null value allowed", func(t *testing.T) {
		err := Validate(DataObject{"flag": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})
}

func TestBoolValidator_Default(t *testing.T) {
	schema := Schema{"enabled": Bool().Default(true).True()}

	t.Run("nil uses default", func(t *testing.T) {
		err := Validate(DataObject{"enabled": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error with default, got: %v", err.Errors)
		}
	})
}

func TestBoolValidator_Custom(t *testing.T) {
	schema := Schema{
		"agree": Bool().Required().Custom(func(v bool, lookup Lookup) error {
			if !v {
				return errors.New("you must agree")
			}
			return nil
		}),
	}

	t.Run("custom validation fails", func(t *testing.T) {
		err := Validate(DataObject{"agree": false}, schema)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("custom validation passes", func(t *testing.T) {
		err := Validate(DataObject{"agree": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestBoolValidator_TypeCheck(t *testing.T) {
	schema := Schema{"flag": Bool().Required()}

	t.Run("non-bool type", func(t *testing.T) {
		err := Validate(DataObject{"flag": "yes"}, schema)
		if err == nil {
			t.Error("Expected error for non-bool type")
		}
	})
}

func TestBoolValidator_Message(t *testing.T) {
	schema := Schema{
		"terms": Bool().Required().True().
			Message("required", "Terms acceptance is required").
			Message("true", "You must accept the terms"),
	}

	t.Run("custom true message", func(t *testing.T) {
		err := Validate(DataObject{"terms": false}, schema)
		if err == nil || err.Errors["terms"][0] != "You must accept the terms" {
			t.Error("Expected custom true message")
		}
	})
}
