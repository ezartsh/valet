package valet

import (
	"fmt"
	"regexp"
	"strconv"
)

// NumberValidator validates numeric values with fluent API
type NumberValidator[T Number] struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	min            T
	minSet         bool
	max            T
	maxSet         bool
	minDigits      int
	minDigitsSet   bool
	maxDigits      int
	maxDigitsSet   bool
	positive       bool
	negative       bool
	integer        bool
	multipleOf     T
	multipleSet    bool
	in             []T
	notIn          []T
	regex          *regexp.Regexp
	notRegex       *regexp.Regexp
	exists         *ExistsRule
	unique         *UniqueRule
	customFn       func(value T, lookup Lookup) error
	messages       map[string]string
	defaultValue   *T
	nullable       bool
	coerce         bool
}

// Number constraint for numeric types
type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64
}

// Num creates a new number validator
func Num[T Number]() *NumberValidator[T] {
	return &NumberValidator[T]{
		messages: make(map[string]string),
	}
}

// Int creates an int64 validator (common for JSON)
func Int() *NumberValidator[int64] {
	return Num[int64]()
}

// Float creates a float64 validator (common for JSON)
func Float() *NumberValidator[float64] {
	return Num[float64]()
}

// Required marks the field as required
func (v *NumberValidator[T]) Required() *NumberValidator[T] {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *NumberValidator[T]) RequiredIf(fn func(data DataObject) bool) *NumberValidator[T] {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *NumberValidator[T]) RequiredUnless(fn func(data DataObject) bool) *NumberValidator[T] {
	v.requiredUnless = fn
	return v
}

// Min sets minimum value
func (v *NumberValidator[T]) Min(n T) *NumberValidator[T] {
	v.min = n
	v.minSet = true
	return v
}

// Max sets maximum value
func (v *NumberValidator[T]) Max(n T) *NumberValidator[T] {
	v.max = n
	v.maxSet = true
	return v
}

// MinDigits sets minimum number of digits
func (v *NumberValidator[T]) MinDigits(n int) *NumberValidator[T] {
	v.minDigits = n
	v.minDigitsSet = true
	return v
}

// MaxDigits sets maximum number of digits
func (v *NumberValidator[T]) MaxDigits(n int) *NumberValidator[T] {
	v.maxDigits = n
	v.maxDigitsSet = true
	return v
}

// Positive requires value > 0
func (v *NumberValidator[T]) Positive() *NumberValidator[T] {
	v.positive = true
	return v
}

// Negative requires value < 0
func (v *NumberValidator[T]) Negative() *NumberValidator[T] {
	v.negative = true
	return v
}

// Integer requires whole number (no decimals)
func (v *NumberValidator[T]) Integer() *NumberValidator[T] {
	v.integer = true
	return v
}

// MultipleOf requires value to be multiple of n
func (v *NumberValidator[T]) MultipleOf(n T) *NumberValidator[T] {
	v.multipleOf = n
	v.multipleSet = true
	return v
}

// In validates value is one of allowed values
func (v *NumberValidator[T]) In(values ...T) *NumberValidator[T] {
	v.in = values
	return v
}

// NotIn validates value is not one of disallowed values
func (v *NumberValidator[T]) NotIn(values ...T) *NumberValidator[T] {
	v.notIn = values
	return v
}

// Regex validates string representation matches pattern
func (v *NumberValidator[T]) Regex(pattern string) *NumberValidator[T] {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.regex = re
	}
	return v
}

// NotRegex validates string representation does NOT match pattern
func (v *NumberValidator[T]) NotRegex(pattern string) *NumberValidator[T] {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.notRegex = re
	}
	return v
}

// Exists adds database existence check
func (v *NumberValidator[T]) Exists(table, column string, wheres ...WhereClause) *NumberValidator[T] {
	v.exists = &ExistsRule{Table: table, Column: column, Where: wheres}
	return v
}

// Unique adds database uniqueness check
func (v *NumberValidator[T]) Unique(table, column string, ignore any, wheres ...WhereClause) *NumberValidator[T] {
	v.unique = &UniqueRule{Table: table, Column: column, Ignore: ignore, Where: wheres}
	return v
}

// Custom adds custom validation function
func (v *NumberValidator[T]) Custom(fn func(value T, lookup Lookup) error) *NumberValidator[T] {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *NumberValidator[T]) Message(rule, message string) *NumberValidator[T] {
	v.messages[rule] = message
	return v
}

// Default sets default value if field is empty/missing
func (v *NumberValidator[T]) Default(value T) *NumberValidator[T] {
	v.defaultValue = &value
	return v
}

// Nullable allows null values
func (v *NumberValidator[T]) Nullable() *NumberValidator[T] {
	v.nullable = true
	return v
}

// Coerce attempts to convert string to number
func (v *NumberValidator[T]) Coerce() *NumberValidator[T] {
	v.coerce = true
	return v
}

// Validate implements Validator interface
func (v *NumberValidator[T]) Validate(ctx *ValidationContext, value any) []string {
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

	// Coerce string to number if enabled
	if v.coerce {
		if str, ok := value.(string); ok {
			if f, err := strconv.ParseFloat(str, 64); err == nil {
				value = f
			}
		}
	}

	// Convert to target type
	num, ok := toNumber[T](value)
	if !ok {
		errors = append(errors, v.msg("type", fmt.Sprintf("%s must be a number", fieldName)))
		return errors
	}

	// Min
	if v.minSet && num < v.min {
		errors = append(errors, v.msg("min", fmt.Sprintf("%s must be at least %v", fieldName, v.min)))
	}

	// Max
	if v.maxSet && num > v.max {
		errors = append(errors, v.msg("max", fmt.Sprintf("%s must be at most %v", fieldName, v.max)))
	}

	// Positive
	if v.positive && num <= 0 {
		errors = append(errors, v.msg("positive", fmt.Sprintf("%s must be positive", fieldName)))
	}

	// Negative
	if v.negative && num >= 0 {
		errors = append(errors, v.msg("negative", fmt.Sprintf("%s must be negative", fieldName)))
	}

	// Integer check
	if v.integer {
		if f, ok := any(num).(float64); ok && f != float64(int64(f)) {
			errors = append(errors, v.msg("integer", fmt.Sprintf("%s must be an integer", fieldName)))
		}
	}

	// In
	if len(v.in) > 0 && !containsNum(v.in, num) {
		errors = append(errors, v.msg("in", fmt.Sprintf("%s must be one of the allowed values", fieldName)))
	}

	// NotIn
	if len(v.notIn) > 0 && containsNum(v.notIn, num) {
		errors = append(errors, v.msg("notIn", fmt.Sprintf("%s must not be one of the disallowed values", fieldName)))
	}

	// String representation for digit/regex checks
	numStr := fmt.Sprintf("%v", num)
	// Remove negative sign and decimal point for digit count
	digitStr := numStr
	if digitStr[0] == '-' {
		digitStr = digitStr[1:]
	}
	if idx := len(digitStr); idx > 0 {
		for i, c := range digitStr {
			if c == '.' {
				digitStr = digitStr[:i] + digitStr[i+1:]
				break
			}
		}
	}

	// MinDigits
	if v.minDigitsSet && len(digitStr) < v.minDigits {
		errors = append(errors, v.msg("minDigits", fmt.Sprintf("%s must have at least %d digits", fieldName, v.minDigits)))
	}

	// MaxDigits
	if v.maxDigitsSet && len(digitStr) > v.maxDigits {
		errors = append(errors, v.msg("maxDigits", fmt.Sprintf("%s must have at most %d digits", fieldName, v.maxDigits)))
	}

	// Regex on string representation
	if v.regex != nil && !v.regex.MatchString(numStr) {
		errors = append(errors, v.msg("regex", fmt.Sprintf("%s format is invalid", fieldName)))
	}

	// NotRegex on string representation
	if v.notRegex != nil && v.notRegex.MatchString(numStr) {
		errors = append(errors, v.msg("notRegex", fmt.Sprintf("%s format is invalid", fieldName)))
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(num, lookup); err != nil {
			errors = append(errors, v.msg("custom", err.Error()))
		}
	}

	return errors
}

// GetDBChecks returns database checks for this field
func (v *NumberValidator[T]) GetDBChecks(fieldPath string, value any) []DBCheck {
	var checks []DBCheck

	num, ok := toNumber[T](value)
	if !ok {
		return nil
	}

	if v.exists != nil {
		checks = append(checks, DBCheck{
			Field:    fieldPath,
			Value:    num,
			Rule:     *v.exists,
			IsUnique: false,
			Message:  v.messages["exists"],
		})
	}

	if v.unique != nil {
		checks = append(checks, DBCheck{
			Field:    fieldPath,
			Value:    num,
			Rule:     ExistsRule{Table: v.unique.Table, Column: v.unique.Column, Where: v.unique.Where},
			IsUnique: true,
			Ignore:   v.unique.Ignore,
			Message:  v.messages["unique"],
		})
	}

	return checks
}

func (v *NumberValidator[T]) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// toNumber converts any numeric type to target type
func toNumber[T Number](value any) (T, bool) {
	var zero T
	switch v := value.(type) {
	case int:
		return T(v), true
	case int32:
		return T(v), true
	case int64:
		return T(v), true
	case uint:
		return T(v), true
	case uint32:
		return T(v), true
	case uint64:
		return T(v), true
	case float32:
		return T(v), true
	case float64:
		return T(v), true
	}
	return zero, false
}

func containsNum[T Number](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
