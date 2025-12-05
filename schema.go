package valet

import (
	"fmt"
	"reflect"
	"strings"
)

// ============================================================================
// OPTIONAL VALIDATOR
// ============================================================================

// OptionalValidator wraps another validator and makes it optional
type OptionalValidator struct {
	inner Validator
}

// Optional creates an optional wrapper for any validator
func Optional(validator Validator) *OptionalValidator {
	return &OptionalValidator{inner: validator}
}

// Validate implements Validator interface
func (v *OptionalValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
	// If value is nil or empty, it's valid (optional field)
	if value == nil {
		return nil
	}

	// For strings, empty is also valid
	if str, ok := value.(string); ok && str == "" {
		return nil
	}

	// Otherwise, delegate to inner validator
	return v.inner.Validate(ctx, value)
}

// GetDBChecks returns database checks from inner validator
func (v *OptionalValidator) GetDBChecks(fieldPath string, value any) []DBCheck {
	if collector, ok := v.inner.(DBCheckCollector); ok {
		return collector.GetDBChecks(fieldPath, value)
	}
	return nil
}

// ============================================================================
// ENUM VALIDATOR
// ============================================================================

// EnumValidator validates value is one of a fixed set of allowed values
type EnumValidator[T comparable] struct {
	values       []T
	required     bool
	messages     map[string]string
	nullable     bool
	defaultValue *T
}

// Enum creates a new enum validator with the allowed values
func Enum[T comparable](values ...T) *EnumValidator[T] {
	return &EnumValidator[T]{
		values:   values,
		messages: make(map[string]string),
	}
}

// EnumInt creates a new enum validator for integer values (convenience function)
func EnumInt(values ...int) *EnumValidator[int] {
	return &EnumValidator[int]{
		values:   values,
		messages: make(map[string]string),
	}
}

// In sets the allowed values (can be used instead of passing to Enum())
func (v *EnumValidator[T]) In(values ...T) *EnumValidator[T] {
	v.values = values
	return v
}

// Required marks the field as required
func (v *EnumValidator[T]) Required() *EnumValidator[T] {
	v.required = true
	return v
}

// Nullable allows null values
func (v *EnumValidator[T]) Nullable() *EnumValidator[T] {
	v.nullable = true
	return v
}

// Default sets default value if field is empty/missing
func (v *EnumValidator[T]) Default(value T) *EnumValidator[T] {
	v.defaultValue = &value
	return v
}

// Message sets custom error message for a rule
func (v *EnumValidator[T]) Message(rule, message string) *EnumValidator[T] {
	v.messages[rule] = message
	return v
}

// Validate implements Validator interface
func (v *EnumValidator[T]) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.defaultValue != nil {
			value = *v.defaultValue
		} else if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else {
			return nil
		}
	}

	// Type check and convert
	var typedValue T
	switch val := value.(type) {
	case T:
		typedValue = val
	default:
		// Try to convert from compatible types
		converted, ok := convertToType[T](value)
		if !ok {
			errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s has invalid type", fieldName)))
			return errors
		}
		typedValue = converted
	}

	// Check if value is in allowed values
	found := false
	for _, allowed := range v.values {
		if typedValue == allowed {
			found = true
			break
		}
	}

	if !found {
		// Format allowed values for error message
		allowedStrs := make([]string, len(v.values))
		for i, val := range v.values {
			allowedStrs[i] = fmt.Sprintf("%v", val)
		}
		errors[fieldPath] = append(errors[fieldPath], v.msg("enum", fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(allowedStrs, ", "))))
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func (v *EnumValidator[T]) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// ============================================================================
// LITERAL VALIDATOR
// ============================================================================

// LiteralValidator validates value matches exactly one specific value
type LiteralValidator[T comparable] struct {
	value    T
	required bool
	messages map[string]string
	nullable bool
}

// Literal creates a new literal validator for an exact value match
func Literal[T comparable](value T) *LiteralValidator[T] {
	return &LiteralValidator[T]{
		value:    value,
		messages: make(map[string]string),
	}
}

// Required marks the field as required
func (v *LiteralValidator[T]) Required() *LiteralValidator[T] {
	v.required = true
	return v
}

// Nullable allows null values
func (v *LiteralValidator[T]) Nullable() *LiteralValidator[T] {
	v.nullable = true
	return v
}

// Message sets custom error message for a rule
func (v *LiteralValidator[T]) Message(rule, message string) *LiteralValidator[T] {
	v.messages[rule] = message
	return v
}

// Validate implements Validator interface
func (v *LiteralValidator[T]) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		return nil
	}

	// Type check and convert
	var typedValue T
	switch val := value.(type) {
	case T:
		typedValue = val
	default:
		// Try to convert from compatible types
		converted, ok := convertToType[T](value)
		if !ok {
			errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s has invalid type", fieldName)))
			return errors
		}
		typedValue = converted
	}

	// Check exact match
	if typedValue != v.value {
		errors[fieldPath] = append(errors[fieldPath], v.msg("literal", fmt.Sprintf("%s must be exactly %v", fieldName, v.value)))
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func (v *LiteralValidator[T]) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// ============================================================================
// UNION VALIDATOR
// ============================================================================

// UnionValidator validates value against multiple validators (any of)
type UnionValidator struct {
	validators []Validator
	required   bool
	messages   map[string]string
	nullable   bool
}

// Union creates a new union validator that accepts any of the provided validators
func Union(validators ...Validator) *UnionValidator {
	return &UnionValidator{
		validators: validators,
		messages:   make(map[string]string),
	}
}

// Required marks the field as required
func (v *UnionValidator) Required() *UnionValidator {
	v.required = true
	return v
}

// Nullable allows null values
func (v *UnionValidator) Nullable() *UnionValidator {
	v.nullable = true
	return v
}

// Message sets custom error message for a rule
func (v *UnionValidator) Message(rule, message string) *UnionValidator {
	v.messages[rule] = message
	return v
}

// Validate implements Validator interface
func (v *UnionValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		return nil
	}

	// Try each validator - if any succeeds, the value is valid
	for _, validator := range v.validators {
		errs := validator.Validate(ctx, value)
		if len(errs) == 0 {
			return nil // One validator passed
		}
	}

	// All validators failed
	errors[fieldPath] = append(errors[fieldPath], v.msg("union", fmt.Sprintf("%s does not match any of the expected types", fieldName)))
	return errors
}

func (v *UnionValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// GetDBChecks returns database checks from all validators in the union
func (v *UnionValidator) GetDBChecks(fieldPath string, value any) []DBCheck {
	var checks []DBCheck
	for _, validator := range v.validators {
		if collector, ok := validator.(DBCheckCollector); ok {
			checks = append(checks, collector.GetDBChecks(fieldPath, value)...)
		}
	}
	return checks
}

// ============================================================================
// ANY VALIDATOR
// ============================================================================

// AnyValidator accepts any value (passthrough)
type AnyValidator struct {
	required bool
	nullable bool
	messages map[string]string
}

// Any creates a new validator that accepts any value
func Any() *AnyValidator {
	return &AnyValidator{
		messages: make(map[string]string),
	}
}

// Required marks the field as required
func (v *AnyValidator) Required() *AnyValidator {
	v.required = true
	return v
}

// Nullable allows null values
func (v *AnyValidator) Nullable() *AnyValidator {
	v.nullable = true
	return v
}

// Message sets custom error message for a rule
func (v *AnyValidator) Message(rule, message string) *AnyValidator {
	v.messages[rule] = message
	return v
}

// Validate implements Validator interface
func (v *AnyValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	if value == nil {
		if v.nullable {
			return nil
		}
		if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
	}

	return nil
}

func (v *AnyValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// convertToType attempts to convert a value to the target type
func convertToType[T comparable](value any) (T, bool) {
	var zero T

	// Use reflection to check if types are compatible
	targetType := reflect.TypeOf(zero)
	valueType := reflect.TypeOf(value)

	if valueType == nil {
		return zero, false
	}

	// Direct type match
	if valueType == targetType {
		return value.(T), true
	}

	// Handle numeric conversions
	switch any(zero).(type) {
	case int:
		switch v := value.(type) {
		case float64:
			return any(int(v)).(T), true
		case int64:
			return any(int(v)).(T), true
		case int:
			return any(v).(T), true
		}
	case int64:
		switch v := value.(type) {
		case float64:
			return any(int64(v)).(T), true
		case int:
			return any(int64(v)).(T), true
		case int64:
			return any(v).(T), true
		}
	case float64:
		switch v := value.(type) {
		case int:
			return any(float64(v)).(T), true
		case int64:
			return any(float64(v)).(T), true
		case float64:
			return any(v).(T), true
		}
	case string:
		if s, ok := value.(string); ok {
			return any(s).(T), true
		}
	case bool:
		if b, ok := value.(bool); ok {
			return any(b).(T), true
		}
	}

	return zero, false
}
