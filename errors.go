package valet

import "errors"

// Validation error types
var (
	ErrValidation       = errors.New("validation failed")
	ErrRequired         = errors.New("field is required")
	ErrInvalidType      = errors.New("invalid type")
	ErrMinLength        = errors.New("value is too short")
	ErrMaxLength        = errors.New("value is too long")
	ErrMinValue         = errors.New("value is too small")
	ErrMaxValue         = errors.New("value is too large")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidURL       = errors.New("invalid URL format")
	ErrInvalidFormat    = errors.New("invalid format")
	ErrNotInAllowed     = errors.New("value not in allowed list")
	ErrInDisallowed     = errors.New("value in disallowed list")
	ErrNotExists        = errors.New("value does not exist")
	ErrAlreadyExists    = errors.New("value already exists")
	ErrInvalidFile      = errors.New("invalid file")
	ErrFileTooSmall     = errors.New("file is too small")
	ErrFileTooLarge     = errors.New("file is too large")
	ErrInvalidMimeType  = errors.New("invalid MIME type")
	ErrInvalidExtension = errors.New("invalid file extension")
	ErrNotImage         = errors.New("file is not an image")
	ErrInvalidDimension = errors.New("invalid image dimensions")
)

// ValidationErrors wraps multiple field errors
type ValidationErrors struct {
	Errors map[string][]string
}

func (e *ValidationErrors) Error() string {
	return "validation failed"
}

// HasErrors returns true if there are any errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Add adds an error for a field
func (e *ValidationErrors) Add(field, message string) {
	if e.Errors == nil {
		e.Errors = make(map[string][]string)
	}
	e.Errors[field] = append(e.Errors[field], message)
}

// Get returns errors for a specific field
func (e *ValidationErrors) Get(field string) []string {
	if e.Errors == nil {
		return nil
	}
	return e.Errors[field]
}

// First returns the first error for a field
func (e *ValidationErrors) First(field string) string {
	errs := e.Get(field)
	if len(errs) > 0 {
		return errs[0]
	}
	return ""
}

// All returns all errors as a flat slice
func (e *ValidationErrors) All() []string {
	var all []string
	for _, errs := range e.Errors {
		all = append(all, errs...)
	}
	return all
}

// Fields returns all field names with errors
func (e *ValidationErrors) Fields() []string {
	fields := make([]string, 0, len(e.Errors))
	for field := range e.Errors {
		fields = append(fields, field)
	}
	return fields
}
