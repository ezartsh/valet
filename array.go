package valet

import (
	"fmt"
	"reflect"
	"sync"
)

// ArrayValidator validates array/slice values with fluent API
type ArrayValidator struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	min            int
	minSet         bool
	max            int
	maxSet         bool
	length         int
	lengthSet      bool
	element        Validator // Validator for each element
	unique         bool      // All elements must be unique
	exists         *ExistsRule
	customFn       func(value []any, lookup Lookup) error
	messages       map[string]MessageArg
	nullable       bool
	concurrent     int   // Number of goroutines for parallel validation (0 = sequential)
	contains       []any // Values that must be present in the array
	doesntContain  []any // Values that must not be present in the array
}

// Array creates a new array validator
func Array() *ArrayValidator {
	return &ArrayValidator{
		messages: make(map[string]MessageArg),
	}
}

// Required marks the field as required
func (v *ArrayValidator) Required(message ...MessageArg) *ArrayValidator {
	v.required = true
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
	return v
}

// RequiredIf makes field required based on condition
func (v *ArrayValidator) RequiredIf(fn func(data DataObject) bool, message ...MessageArg) *ArrayValidator {
	v.requiredIf = fn
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *ArrayValidator) RequiredUnless(fn func(data DataObject) bool, message ...MessageArg) *ArrayValidator {
	v.requiredUnless = fn
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
	return v
}

// Min sets minimum number of elements
func (v *ArrayValidator) Min(n int, message ...MessageArg) *ArrayValidator {
	v.min = n
	v.minSet = true
	if len(message) > 0 {
		v.messages["min"] = message[0]
	}
	return v
}

// Max sets maximum number of elements
func (v *ArrayValidator) Max(n int, message ...MessageArg) *ArrayValidator {
	v.max = n
	v.maxSet = true
	if len(message) > 0 {
		v.messages["max"] = message[0]
	}
	return v
}

// Length sets exact number of elements
func (v *ArrayValidator) Length(n int, message ...MessageArg) *ArrayValidator {
	v.length = n
	v.lengthSet = true
	if len(message) > 0 {
		v.messages["length"] = message[0]
	}
	return v
}

// Of sets the validator for each element
func (v *ArrayValidator) Of(validator Validator) *ArrayValidator {
	v.element = validator
	return v
}

// Unique requires all elements to be unique
func (v *ArrayValidator) Unique(message ...MessageArg) *ArrayValidator {
	v.unique = true
	if len(message) > 0 {
		v.messages["unique"] = message[0]
	}
	return v
}

// Distinct is an alias for Unique (Laravel naming)
func (v *ArrayValidator) Distinct(message ...MessageArg) *ArrayValidator {
	return v.Unique(message...)
}

// Contains requires array to contain the specified values
func (v *ArrayValidator) Contains(values ...any) *ArrayValidator {
	v.contains = values
	return v
}

// ContainsWithMessage requires array to contain specified values with custom message
func (v *ArrayValidator) ContainsWithMessage(message MessageArg, values ...any) *ArrayValidator {
	v.contains = values
	v.messages["contains"] = message
	return v
}

// DoesntContain requires array to NOT contain the specified values
func (v *ArrayValidator) DoesntContain(values ...any) *ArrayValidator {
	v.doesntContain = values
	return v
}

// DoesntContainWithMessage requires array to NOT contain specified values with custom message
func (v *ArrayValidator) DoesntContainWithMessage(message MessageArg, values ...any) *ArrayValidator {
	v.doesntContain = values
	v.messages["doesntContain"] = message
	return v
}

// Exists adds database existence check for each element
func (v *ArrayValidator) Exists(table, column string, wheres ...WhereClause) *ArrayValidator {
	v.exists = &ExistsRule{Table: table, Column: column, Where: wheres}
	return v
}

// Custom adds custom validation function
func (v *ArrayValidator) Custom(fn func(value []any, lookup Lookup) error) *ArrayValidator {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *ArrayValidator) Message(rule string, message MessageArg) *ArrayValidator {
	v.messages[rule] = message
	return v
}

// Nullable allows null values
func (v *ArrayValidator) Nullable() *ArrayValidator {
	v.nullable = true
	return v
}

// Nonempty is shorthand for Min(1)
func (v *ArrayValidator) Nonempty() *ArrayValidator {
	return v.Min(1)
}

// Concurrent enables parallel validation of array elements
// n specifies the maximum number of goroutines to use (0 = sequential)
func (v *ArrayValidator) Concurrent(n int) *ArrayValidator {
	if n < 0 {
		n = 0
	}
	v.concurrent = n
	return v
}

// Validate implements Validator interface
func (v *ArrayValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
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
		if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName), msgCtx))
			return errors
		}
		if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName), msgCtx))
			return errors
		}
		if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName), msgCtx))
			return errors
		}
		return nil
	}

	// Type check
	arr, ok := value.([]any)
	if !ok {
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be an array", fieldName), msgCtx))
		return errors
	}

	length := len(arr)
	msgCtx.Value = arr

	// Length check
	if v.lengthSet && length != v.length {
		msgCtx.Param = v.length
		errors[fieldPath] = append(errors[fieldPath], v.msg("length", fmt.Sprintf("%s must have exactly %d elements", fieldName, v.length), msgCtx))
	}

	// Min check
	if v.minSet && length < v.min {
		msgCtx.Param = v.min
		errors[fieldPath] = append(errors[fieldPath], v.msg("min", fmt.Sprintf("%s must have at least %d elements", fieldName, v.min), msgCtx))
	}

	// Max check
	if v.maxSet && length > v.max {
		msgCtx.Param = v.max
		errors[fieldPath] = append(errors[fieldPath], v.msg("max", fmt.Sprintf("%s must have at most %d elements", fieldName, v.max), msgCtx))
	}

	// Unique check
	if v.unique {
		seen := make(map[any]bool)
		for i, item := range arr {
			if seen[item] {
				elementPath := fmt.Sprintf("%s.%d", fieldPath, i)
				elemCtx := MessageContext{
					Field: fieldName,
					Path:  elementPath,
					Index: i,
					Value: item,
					Data:  DataAccessor(ctx.RootData),
				}
				errors[elementPath] = append(errors[elementPath], v.msg("unique", fmt.Sprintf("%s[%d] is a duplicate", fieldName, i), elemCtx))
			}
			seen[item] = true
		}
	}

	// Contains check - array must contain all specified values
	if len(v.contains) > 0 {
		for _, required := range v.contains {
			found := false
			for _, item := range arr {
				if equalValues(item, required) {
					found = true
					break
				}
			}
			if !found {
				msgCtx.Param = required
				errors[fieldPath] = append(errors[fieldPath], v.msg("contains", fmt.Sprintf("%s must contain %v", fieldName, required), msgCtx))
			}
		}
	}

	// DoesntContain check - array must not contain any of the specified values
	if len(v.doesntContain) > 0 {
		for _, forbidden := range v.doesntContain {
			for i, item := range arr {
				if equalValues(item, forbidden) {
					elementPath := fmt.Sprintf("%s.%d", fieldPath, i)
					elemCtx := MessageContext{
						Field: fieldName,
						Path:  elementPath,
						Index: i,
						Value: item,
						Param: forbidden,
						Data:  DataAccessor(ctx.RootData),
					}
					errors[elementPath] = append(errors[elementPath], v.msg("doesntContain", fmt.Sprintf("%s must not contain %v", fieldName, forbidden), elemCtx))
				}
			}
		}
	}

	// Validate each element
	if v.element != nil {
		if v.concurrent > 0 && len(arr) > 1 {
			// Concurrent validation
			var wg sync.WaitGroup
			var mu sync.Mutex
			sem := make(chan struct{}, v.concurrent)

			for i, item := range arr {
				wg.Add(1)
				go func(idx int, val any) {
					defer wg.Done()
					sem <- struct{}{}        // Acquire semaphore
					defer func() { <-sem }() // Release semaphore

					childCtx := &ValidationContext{
						Ctx:      ctx.Ctx,
						RootData: ctx.RootData,
						Path:     append(append([]string{}, ctx.Path...), fmt.Sprintf("%d", idx)),
						Options:  ctx.Options,
					}
					childErrors := v.element.Validate(childCtx, val)
					if len(childErrors) > 0 {
						mu.Lock()
						for path, errs := range childErrors {
							errors[path] = append(errors[path], errs...)
						}
						mu.Unlock()
					}
				}(i, item)
			}
			wg.Wait()
		} else {
			// Sequential validation
			for i, item := range arr {
				childCtx := &ValidationContext{
					Ctx:      ctx.Ctx,
					RootData: ctx.RootData,
					Path:     append(ctx.Path, fmt.Sprintf("%d", i)),
					Options:  ctx.Options,
				}
				childErrors := v.element.Validate(childCtx, item)
				// Merge child errors
				for path, errs := range childErrors {
					errors[path] = append(errors[path], errs...)
				}
			}
		}
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(arr, lookup); err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error(), msgCtx))
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// GetDBChecks returns database checks for array elements
func (v *ArrayValidator) GetDBChecks(fieldPath string, value any) []DBCheck {
	var checks []DBCheck

	arr, ok := value.([]any)
	if !ok {
		return nil
	}

	// If array has Exists rule, check each element
	if v.exists != nil {
		for i, item := range arr {
			checks = append(checks, DBCheck{
				Field:    fmt.Sprintf("%s.%d", fieldPath, i),
				Value:    item,
				Rule:     *v.exists,
				IsUnique: false,
			})
		}
	}

	// If array has element validator (Of), recursively collect DB checks
	if v.element != nil {
		if collector, ok := v.element.(DBCheckCollector); ok {
			for i, item := range arr {
				elementPath := fmt.Sprintf("%s.%d", fieldPath, i)
				nestedChecks := collector.GetDBChecks(elementPath, item)
				checks = append(checks, nestedChecks...)
			}
		}
	}

	return checks
}

func (v *ArrayValidator) msg(rule, defaultMsg string, msgCtx MessageContext) string {
	if msg, ok := v.messages[rule]; ok {
		msgCtx.Rule = rule
		return resolveMessage(msg, msgCtx)
	}
	return defaultMsg
}

// equalValues compares two values for equality using reflect.DeepEqual
func equalValues(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
