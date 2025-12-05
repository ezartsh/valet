package valet

import (
	"fmt"
	"regexp"
	"strconv"
)

// NumberValidator validates numeric values with fluent API
type NumberValidator[T Number] struct {
	required        bool
	requiredIf      func(data DataObject) bool
	requiredUnless  func(data DataObject) bool
	min             T
	minSet          bool
	max             T
	maxSet          bool
	minDigits       int
	minDigitsSet    bool
	maxDigits       int
	maxDigitsSet    bool
	positive        bool
	negative        bool
	integer         bool
	multipleOf      T
	multipleSet     bool
	in              []T
	notIn           []T
	regex           *regexp.Regexp
	notRegex        *regexp.Regexp
	exists          *ExistsRule
	unique          *UniqueRule
	customFn        func(value T, lookup Lookup) error
	messages        map[string]MessageArg
	defaultValue    *T
	nullable        bool
	coerce          bool
	lessThan        string
	greaterThan     string
	lessThanOrEq    string
	greaterThanOrEq string
}

// Number constraint for numeric types
type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64
}

// Num creates a new number validator
func Num[T Number]() *NumberValidator[T] {
	return &NumberValidator[T]{
		messages: make(map[string]MessageArg),
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
func (v *NumberValidator[T]) Required(message ...MessageArg) *NumberValidator[T] {
	v.required = true
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
	return v
}

// RequiredIf makes field required based on condition
func (v *NumberValidator[T]) RequiredIf(fn func(data DataObject) bool, message ...MessageArg) *NumberValidator[T] {
	v.requiredIf = fn
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *NumberValidator[T]) RequiredUnless(fn func(data DataObject) bool, message ...MessageArg) *NumberValidator[T] {
	v.requiredUnless = fn
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
	return v
}

// Min sets minimum value
func (v *NumberValidator[T]) Min(n T, message ...MessageArg) *NumberValidator[T] {
	v.min = n
	v.minSet = true
	if len(message) > 0 {
		v.messages["min"] = message[0]
	}
	return v
}

// Max sets maximum value
func (v *NumberValidator[T]) Max(n T, message ...MessageArg) *NumberValidator[T] {
	v.max = n
	v.maxSet = true
	if len(message) > 0 {
		v.messages["max"] = message[0]
	}
	return v
}

// Between sets both minimum and maximum value (inclusive)
func (v *NumberValidator[T]) Between(min, max T, message ...MessageArg) *NumberValidator[T] {
	v.min = min
	v.minSet = true
	v.max = max
	v.maxSet = true
	if len(message) > 0 {
		v.messages["between"] = message[0]
	}
	return v
}

// Step is an alias for MultipleOf (Zod naming)
func (v *NumberValidator[T]) Step(n T, message ...MessageArg) *NumberValidator[T] {
	return v.MultipleOf(n, message...)
}

// MinDigits sets minimum number of digits
func (v *NumberValidator[T]) MinDigits(n int, message ...MessageArg) *NumberValidator[T] {
	v.minDigits = n
	v.minDigitsSet = true
	if len(message) > 0 {
		v.messages["minDigits"] = message[0]
	}
	return v
}

// MaxDigits sets maximum number of digits
func (v *NumberValidator[T]) MaxDigits(n int, message ...MessageArg) *NumberValidator[T] {
	v.maxDigits = n
	v.maxDigitsSet = true
	if len(message) > 0 {
		v.messages["maxDigits"] = message[0]
	}
	return v
}

// Positive requires value > 0
func (v *NumberValidator[T]) Positive(message ...MessageArg) *NumberValidator[T] {
	v.positive = true
	if len(message) > 0 {
		v.messages["positive"] = message[0]
	}
	return v
}

// Negative requires value < 0
func (v *NumberValidator[T]) Negative(message ...MessageArg) *NumberValidator[T] {
	v.negative = true
	if len(message) > 0 {
		v.messages["negative"] = message[0]
	}
	return v
}

// Integer requires whole number (no decimals)
func (v *NumberValidator[T]) Integer(message ...MessageArg) *NumberValidator[T] {
	v.integer = true
	if len(message) > 0 {
		v.messages["integer"] = message[0]
	}
	return v
}

// MultipleOf requires value to be multiple of n
func (v *NumberValidator[T]) MultipleOf(n T, message ...MessageArg) *NumberValidator[T] {
	v.multipleOf = n
	v.multipleSet = true
	if len(message) > 0 {
		v.messages["multipleOf"] = message[0]
	}
	return v
}

// In validates value is one of allowed values
func (v *NumberValidator[T]) In(values ...T) *NumberValidator[T] {
	v.in = values
	return v
}

// InWithMessage validates value is one of allowed values with custom message
func (v *NumberValidator[T]) InWithMessage(message MessageArg, values ...T) *NumberValidator[T] {
	v.in = values
	v.messages["in"] = message
	return v
}

// NotIn validates value is not one of disallowed values
func (v *NumberValidator[T]) NotIn(values ...T) *NumberValidator[T] {
	v.notIn = values
	return v
}

// NotInWithMessage validates value is not one of disallowed values with custom message
func (v *NumberValidator[T]) NotInWithMessage(message MessageArg, values ...T) *NumberValidator[T] {
	v.notIn = values
	v.messages["notIn"] = message
	return v
}

// Regex validates string representation matches pattern
func (v *NumberValidator[T]) Regex(pattern string, message ...MessageArg) *NumberValidator[T] {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.regex = re
	}
	if len(message) > 0 {
		v.messages["regex"] = message[0]
	}
	return v
}

// NotRegex validates string representation does NOT match pattern
func (v *NumberValidator[T]) NotRegex(pattern string, message ...MessageArg) *NumberValidator[T] {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.notRegex = re
	}
	if len(message) > 0 {
		v.messages["notRegex"] = message[0]
	}
	return v
}

// Exists adds database existence check
func (v *NumberValidator[T]) Exists(table, column string, wheres ...WhereClause) *NumberValidator[T] {
	v.exists = &ExistsRule{Table: table, Column: column, Where: wheres}
	return v
}

// ExistsWithMessage adds database existence check with custom message
func (v *NumberValidator[T]) ExistsWithMessage(message MessageArg, table, column string, wheres ...WhereClause) *NumberValidator[T] {
	v.exists = &ExistsRule{Table: table, Column: column, Where: wheres}
	v.messages["exists"] = message
	return v
}

// Unique adds database uniqueness check
func (v *NumberValidator[T]) Unique(table, column string, ignore any, wheres ...WhereClause) *NumberValidator[T] {
	v.unique = &UniqueRule{Table: table, Column: column, Ignore: ignore, Where: wheres}
	return v
}

// UniqueWithMessage adds database uniqueness check with custom message
func (v *NumberValidator[T]) UniqueWithMessage(message MessageArg, table, column string, ignore any, wheres ...WhereClause) *NumberValidator[T] {
	v.unique = &UniqueRule{Table: table, Column: column, Ignore: ignore, Where: wheres}
	v.messages["unique"] = message
	return v
}

// Custom adds custom validation function
func (v *NumberValidator[T]) Custom(fn func(value T, lookup Lookup) error) *NumberValidator[T] {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *NumberValidator[T]) Message(rule string, message MessageArg) *NumberValidator[T] {
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

// LessThan validates value is less than another field's value
func (v *NumberValidator[T]) LessThan(fieldPath string, message ...MessageArg) *NumberValidator[T] {
	v.lessThan = fieldPath
	if len(message) > 0 {
		v.messages["lessThan"] = message[0]
	}
	return v
}

// GreaterThan validates value is greater than another field's value
func (v *NumberValidator[T]) GreaterThan(fieldPath string, message ...MessageArg) *NumberValidator[T] {
	v.greaterThan = fieldPath
	if len(message) > 0 {
		v.messages["greaterThan"] = message[0]
	}
	return v
}

// LessThanOrEqual validates value is less than or equal to another field's value
func (v *NumberValidator[T]) LessThanOrEqual(fieldPath string, message ...MessageArg) *NumberValidator[T] {
	v.lessThanOrEq = fieldPath
	if len(message) > 0 {
		v.messages["lessThanOrEqual"] = message[0]
	}
	return v
}

// GreaterThanOrEqual validates value is greater than or equal to another field's value
func (v *NumberValidator[T]) GreaterThanOrEqual(fieldPath string, message ...MessageArg) *NumberValidator[T] {
	v.greaterThanOrEq = fieldPath
	if len(message) > 0 {
		v.messages["greaterThanOrEqual"] = message[0]
	}
	return v
}

// Validate implements Validator interface
func (v *NumberValidator[T]) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Create base message context
	msgCtx := MessageContext{
		Field: fieldName,
		Path:  fieldPath,
		Index: extractIndex(fieldPath),
		Value: value,
		Data:  DataAccessor(ctx.RootData),
	}

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.defaultValue != nil {
			value = *v.defaultValue
		} else if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName), msgCtx))
			return errors
		} else if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName), msgCtx))
			return errors
		} else if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName), msgCtx))
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
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be a number", fieldName), msgCtx))
		return errors
	}

	// Update msgCtx with actual value
	msgCtx.Value = num

	// Min
	if v.minSet && num < v.min {
		msgCtx.Param = v.min
		errors[fieldPath] = append(errors[fieldPath], v.msg("min", fmt.Sprintf("%s must be at least %v", fieldName, v.min), msgCtx))
	}

	// Max
	if v.maxSet && num > v.max {
		msgCtx.Param = v.max
		errors[fieldPath] = append(errors[fieldPath], v.msg("max", fmt.Sprintf("%s must be at most %v", fieldName, v.max), msgCtx))
	}

	// Positive
	if v.positive && num <= 0 {
		errors[fieldPath] = append(errors[fieldPath], v.msg("positive", fmt.Sprintf("%s must be positive", fieldName), msgCtx))
	}

	// Negative
	if v.negative && num >= 0 {
		errors[fieldPath] = append(errors[fieldPath], v.msg("negative", fmt.Sprintf("%s must be negative", fieldName), msgCtx))
	}

	// MultipleOf / Step
	if v.multipleSet {
		// Convert to float64 for modulo calculation
		numFloat := float64(num)
		stepFloat := float64(v.multipleOf)
		if stepFloat != 0 {
			remainder := numFloat - stepFloat*float64(int64(numFloat/stepFloat))
			// Use a small epsilon for float comparison
			if remainder > 1e-9 && remainder < stepFloat-1e-9 {
				msgCtx.Param = v.multipleOf
				errors[fieldPath] = append(errors[fieldPath], v.msg("multipleOf", fmt.Sprintf("%s must be a multiple of %v", fieldName, v.multipleOf), msgCtx))
			}
		}
	}

	// Integer check
	if v.integer {
		if f, ok := any(num).(float64); ok && f != float64(int64(f)) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("integer", fmt.Sprintf("%s must be an integer", fieldName), msgCtx))
		}
	}

	// In
	if len(v.in) > 0 && !containsNum(v.in, num) {
		msgCtx.Param = v.in
		errors[fieldPath] = append(errors[fieldPath], v.msg("in", fmt.Sprintf("%s must be one of the allowed values", fieldName), msgCtx))
	}

	// NotIn
	if len(v.notIn) > 0 && containsNum(v.notIn, num) {
		msgCtx.Param = v.notIn
		errors[fieldPath] = append(errors[fieldPath], v.msg("notIn", fmt.Sprintf("%s must not be one of the disallowed values", fieldName), msgCtx))
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
		msgCtx.Param = v.minDigits
		errors[fieldPath] = append(errors[fieldPath], v.msg("minDigits", fmt.Sprintf("%s must have at least %d digits", fieldName, v.minDigits), msgCtx))
	}

	// MaxDigits
	if v.maxDigitsSet && len(digitStr) > v.maxDigits {
		msgCtx.Param = v.maxDigits
		errors[fieldPath] = append(errors[fieldPath], v.msg("maxDigits", fmt.Sprintf("%s must have at most %d digits", fieldName, v.maxDigits), msgCtx))
	}

	// Regex on string representation
	if v.regex != nil && !v.regex.MatchString(numStr) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("regex", fmt.Sprintf("%s format is invalid", fieldName), msgCtx))
	}

	// NotRegex on string representation
	if v.notRegex != nil && v.notRegex.MatchString(numStr) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("notRegex", fmt.Sprintf("%s format is invalid", fieldName), msgCtx))
	}

	// LessThan - cross-field comparison
	if v.lessThan != "" {
		otherValue := lookupPath(ctx.RootData, v.lessThan)
		if otherValue.Exists() {
			if otherNum, ok := toNumber[T](otherValue.Value()); ok {
				if num >= otherNum {
					msgCtx.Param = v.lessThan
					errors[fieldPath] = append(errors[fieldPath], v.msg("lessThan", fmt.Sprintf("%s must be less than %s", fieldName, v.lessThan), msgCtx))
				}
			}
		}
	}

	// GreaterThan - cross-field comparison
	if v.greaterThan != "" {
		otherValue := lookupPath(ctx.RootData, v.greaterThan)
		if otherValue.Exists() {
			if otherNum, ok := toNumber[T](otherValue.Value()); ok {
				if num <= otherNum {
					msgCtx.Param = v.greaterThan
					errors[fieldPath] = append(errors[fieldPath], v.msg("greaterThan", fmt.Sprintf("%s must be greater than %s", fieldName, v.greaterThan), msgCtx))
				}
			}
		}
	}

	// LessThanOrEqual - cross-field comparison
	if v.lessThanOrEq != "" {
		otherValue := lookupPath(ctx.RootData, v.lessThanOrEq)
		if otherValue.Exists() {
			if otherNum, ok := toNumber[T](otherValue.Value()); ok {
				if num > otherNum {
					msgCtx.Param = v.lessThanOrEq
					errors[fieldPath] = append(errors[fieldPath], v.msg("lessThanOrEqual", fmt.Sprintf("%s must be less than or equal to %s", fieldName, v.lessThanOrEq), msgCtx))
				}
			}
		}
	}

	// GreaterThanOrEqual - cross-field comparison
	if v.greaterThanOrEq != "" {
		otherValue := lookupPath(ctx.RootData, v.greaterThanOrEq)
		if otherValue.Exists() {
			if otherNum, ok := toNumber[T](otherValue.Value()); ok {
				if num < otherNum {
					msgCtx.Param = v.greaterThanOrEq
					errors[fieldPath] = append(errors[fieldPath], v.msg("greaterThanOrEqual", fmt.Sprintf("%s must be greater than or equal to %s", fieldName, v.greaterThanOrEq), msgCtx))
				}
			}
		}
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(num, lookup); err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error(), msgCtx))
		}
	}

	if len(errors) == 0 {
		return nil
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

func (v *NumberValidator[T]) msg(rule, defaultMsg string, msgCtx MessageContext) string {
	if msg, ok := v.messages[rule]; ok {
		msgCtx.Rule = rule
		return resolveMessage(msg, msgCtx)
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
