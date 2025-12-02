package valet

import "fmt"

// BoolValidator validates boolean values with fluent API
type BoolValidator struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	mustBeTrue     bool
	mustBeFalse    bool
	customFn       func(value bool, lookup Lookup) error
	messages       map[string]string
	defaultValue   *bool
	nullable       bool
	coerce         bool
}

// Bool creates a new boolean validator
func Bool() *BoolValidator {
	return &BoolValidator{
		messages: make(map[string]string),
	}
}

// Required marks the field as required
func (v *BoolValidator) Required() *BoolValidator {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *BoolValidator) RequiredIf(fn func(data DataObject) bool) *BoolValidator {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *BoolValidator) RequiredUnless(fn func(data DataObject) bool) *BoolValidator {
	v.requiredUnless = fn
	return v
}

// True requires value to be true
func (v *BoolValidator) True() *BoolValidator {
	v.mustBeTrue = true
	return v
}

// False requires value to be false
func (v *BoolValidator) False() *BoolValidator {
	v.mustBeFalse = true
	return v
}

// Custom adds custom validation function
func (v *BoolValidator) Custom(fn func(value bool, lookup Lookup) error) *BoolValidator {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *BoolValidator) Message(rule, message string) *BoolValidator {
	v.messages[rule] = message
	return v
}

// Default sets default value if field is empty/missing
func (v *BoolValidator) Default(value bool) *BoolValidator {
	v.defaultValue = &value
	return v
}

// Nullable allows null values
func (v *BoolValidator) Nullable() *BoolValidator {
	v.nullable = true
	return v
}

// Coerce attempts to convert string/number to boolean
func (v *BoolValidator) Coerce() *BoolValidator {
	v.coerce = true
	return v
}

// Validate implements Validator interface
func (v *BoolValidator) Validate(ctx *ValidationContext, value any) []string {
	var errors []string
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.defaultValue != nil {
			value = *v.defaultValue
		} else if v.required {
			errors = append(errors, v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors = append(errors, v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors = append(errors, v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else {
			return nil
		}
	}

	// Coerce to boolean if enabled
	if v.coerce {
		value = coerceToBool(value)
	}

	// Type check
	b, ok := value.(bool)
	if !ok {
		errors = append(errors, v.msg("type", fmt.Sprintf("%s must be a boolean", fieldName)))
		return errors
	}

	// Must be true
	if v.mustBeTrue && !b {
		errors = append(errors, v.msg("true", fmt.Sprintf("%s must be true", fieldName)))
	}

	// Must be false
	if v.mustBeFalse && b {
		errors = append(errors, v.msg("false", fmt.Sprintf("%s must be false", fieldName)))
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(b, lookup); err != nil {
			errors = append(errors, v.msg("custom", err.Error()))
		}
	}

	return errors
}

func (v *BoolValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

func coerceToBool(value any) any {
	switch v := value.(type) {
	case string:
		switch v {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off", "":
			return false
		}
	case int, int32, int64, float32, float64:
		return v != 0
	}
	return value
}
