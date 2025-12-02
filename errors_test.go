package valet

import (
	"testing"
)

func TestValidationErrors_Error(t *testing.T) {
	e := &ValidationErrors{
		Errors: map[string][]string{
			"name": {"is required"},
		},
	}

	if e.Error() != "validation failed" {
		t.Errorf("Error() = %s, want 'validation failed'", e.Error())
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	t.Run("has errors", func(t *testing.T) {
		e := &ValidationErrors{
			Errors: map[string][]string{
				"name": {"is required"},
			},
		}
		if !e.HasErrors() {
			t.Error("HasErrors() should return true")
		}
	})

	t.Run("no errors", func(t *testing.T) {
		e := &ValidationErrors{
			Errors: map[string][]string{},
		}
		if e.HasErrors() {
			t.Error("HasErrors() should return false")
		}
	})

	t.Run("nil errors", func(t *testing.T) {
		e := &ValidationErrors{}
		if e.HasErrors() {
			t.Error("HasErrors() should return false for nil")
		}
	})
}

func TestValidationErrors_Add(t *testing.T) {
	t.Run("add to new field", func(t *testing.T) {
		e := &ValidationErrors{}
		e.Add("name", "is required")

		if len(e.Errors["name"]) != 1 {
			t.Error("Expected 1 error for name")
		}
	})

	t.Run("add to existing field", func(t *testing.T) {
		e := &ValidationErrors{
			Errors: map[string][]string{
				"name": {"is required"},
			},
		}
		e.Add("name", "is too short")

		if len(e.Errors["name"]) != 2 {
			t.Error("Expected 2 errors for name")
		}
	})
}

func TestValidationErrors_Get(t *testing.T) {
	e := &ValidationErrors{
		Errors: map[string][]string{
			"name":  {"is required", "is too short"},
			"email": {"is invalid"},
		},
	}

	t.Run("existing field", func(t *testing.T) {
		errs := e.Get("name")
		if len(errs) != 2 {
			t.Errorf("Get(name) length = %d, want 2", len(errs))
		}
	})

	t.Run("missing field", func(t *testing.T) {
		errs := e.Get("missing")
		if errs != nil {
			t.Error("Get(missing) should return nil")
		}
	})

	t.Run("nil errors", func(t *testing.T) {
		e := &ValidationErrors{}
		errs := e.Get("name")
		if errs != nil {
			t.Error("Get on nil should return nil")
		}
	})
}

func TestValidationErrors_First(t *testing.T) {
	e := &ValidationErrors{
		Errors: map[string][]string{
			"name": {"is required", "is too short"},
		},
	}

	t.Run("existing field", func(t *testing.T) {
		first := e.First("name")
		if first != "is required" {
			t.Errorf("First(name) = %s, want 'is required'", first)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		first := e.First("missing")
		if first != "" {
			t.Errorf("First(missing) = %s, want empty string", first)
		}
	})
}

func TestValidationErrors_All(t *testing.T) {
	e := &ValidationErrors{
		Errors: map[string][]string{
			"name":  {"is required"},
			"email": {"is invalid", "is too long"},
		},
	}

	all := e.All()
	if len(all) != 3 {
		t.Errorf("All() length = %d, want 3", len(all))
	}
}

func TestValidationErrors_Fields(t *testing.T) {
	e := &ValidationErrors{
		Errors: map[string][]string{
			"name":  {"is required"},
			"email": {"is invalid"},
		},
	}

	fields := e.Fields()
	if len(fields) != 2 {
		t.Errorf("Fields() length = %d, want 2", len(fields))
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that error types are defined
	errors := []error{
		ErrValidation,
		ErrRequired,
		ErrInvalidType,
		ErrMinLength,
		ErrMaxLength,
		ErrMinValue,
		ErrMaxValue,
		ErrInvalidEmail,
		ErrInvalidURL,
		ErrInvalidFormat,
		ErrNotInAllowed,
		ErrInDisallowed,
		ErrNotExists,
		ErrAlreadyExists,
		ErrInvalidFile,
		ErrFileTooSmall,
		ErrFileTooLarge,
		ErrInvalidMimeType,
		ErrInvalidExtension,
		ErrNotImage,
		ErrInvalidDimension,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("Error type should not be nil")
		}
	}
}
