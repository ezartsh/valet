package valet

import "fmt"

// ObjectValidator validates object/map values with fluent API
type ObjectValidator struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	schema         Schema
	strict         bool // Fail on unknown keys
	passthrough    bool // Allow unknown keys (default)
	customFn       func(value DataObject, lookup Lookup) error
	messages       map[string]string
	nullable       bool
}

// Object creates a new object validator
func Object() *ObjectValidator {
	return &ObjectValidator{
		messages:    make(map[string]string),
		passthrough: true,
	}
}

// Required marks the field as required
func (v *ObjectValidator) Required() *ObjectValidator {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *ObjectValidator) RequiredIf(fn func(data DataObject) bool) *ObjectValidator {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *ObjectValidator) RequiredUnless(fn func(data DataObject) bool) *ObjectValidator {
	v.requiredUnless = fn
	return v
}

// Shape sets the schema for object properties
func (v *ObjectValidator) Shape(schema Schema) *ObjectValidator {
	v.schema = schema
	return v
}

// Item is an alias for Shape (compatibility with go-validet)
func (v *ObjectValidator) Item(schema Schema) *ObjectValidator {
	return v.Shape(schema)
}

// Strict fails validation if unknown keys are present
func (v *ObjectValidator) Strict() *ObjectValidator {
	v.strict = true
	v.passthrough = false
	return v
}

// Passthrough allows unknown keys (default behavior)
func (v *ObjectValidator) Passthrough() *ObjectValidator {
	v.passthrough = true
	v.strict = false
	return v
}

// Extend creates a new validator with additional schema fields
func (v *ObjectValidator) Extend(additional Schema) *ObjectValidator {
	newValidator := &ObjectValidator{
		required:    v.required,
		requiredIf:  v.requiredIf,
		strict:      v.strict,
		passthrough: v.passthrough,
		customFn:    v.customFn,
		messages:    make(map[string]string),
		nullable:    v.nullable,
		schema:      make(Schema),
	}

	// Copy existing schema
	for k, val := range v.schema {
		newValidator.schema[k] = val
	}

	// Add new fields
	for k, val := range additional {
		newValidator.schema[k] = val
	}

	// Copy messages
	for k, val := range v.messages {
		newValidator.messages[k] = val
	}

	return newValidator
}

// Custom adds custom validation function
func (v *ObjectValidator) Custom(fn func(value DataObject, lookup Lookup) error) *ObjectValidator {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *ObjectValidator) Message(rule, message string) *ObjectValidator {
	v.messages[rule] = message
	return v
}

// Nullable allows null values
func (v *ObjectValidator) Nullable() *ObjectValidator {
	v.nullable = true
	return v
}

// Validate implements Validator interface
func (v *ObjectValidator) Validate(ctx *ValidationContext, value any) []string {
	var errors []string
	fieldName := ""
	if len(ctx.Path) > 0 {
		fieldName = ctx.Path[len(ctx.Path)-1]
	}

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.required {
			errors = append(errors, v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors = append(errors, v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors = append(errors, v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		return nil
	}

	// Type check
	obj, ok := value.(map[string]any)
	if !ok {
		errors = append(errors, v.msg("type", fmt.Sprintf("%s must be an object", fieldName)))
		return errors
	}

	// Strict mode - check for unknown keys
	if v.strict && v.schema != nil {
		for key := range obj {
			if _, exists := v.schema[key]; !exists {
				errors = append(errors, v.msg("strict", fmt.Sprintf("unknown field: %s", key)))
			}
		}
	}

	// Validate nested schema
	if v.schema != nil {
		for key, validator := range v.schema {
			childCtx := &ValidationContext{
				Ctx:      ctx.Ctx,
				RootData: ctx.RootData,
				Path:     append(ctx.Path, key),
				Options:  ctx.Options,
			}

			childValue := obj[key]
			childErrors := validator.Validate(childCtx, childValue)
			errors = append(errors, childErrors...)
		}
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(obj, lookup); err != nil {
			errors = append(errors, v.msg("custom", err.Error()))
		}
	}

	return errors
}

func (v *ObjectValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// GetDBChecks returns database checks from nested schema validators
func (v *ObjectValidator) GetDBChecks(fieldPath string, value any) []DBCheck {
	var checks []DBCheck

	obj, ok := value.(map[string]any)
	if !ok || v.schema == nil {
		return nil
	}

	// Recursively collect DB checks from nested validators
	for field, validator := range v.schema {
		nestedPath := fieldPath + "." + field
		fieldValue := obj[field]

		if collector, ok := validator.(DBCheckCollector); ok {
			nestedChecks := collector.GetDBChecks(nestedPath, fieldValue)
			checks = append(checks, nestedChecks...)
		}
	}

	return checks
}
