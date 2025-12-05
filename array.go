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
	messages       map[string]string
	nullable       bool
	concurrent     int   // Number of goroutines for parallel validation (0 = sequential)
	contains       []any // Values that must be present in the array
	doesntContain  []any // Values that must not be present in the array
}

// Array creates a new array validator
func Array() *ArrayValidator {
	return &ArrayValidator{
		messages: make(map[string]string),
	}
}

// Required marks the field as required
func (v *ArrayValidator) Required() *ArrayValidator {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *ArrayValidator) RequiredIf(fn func(data DataObject) bool) *ArrayValidator {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *ArrayValidator) RequiredUnless(fn func(data DataObject) bool) *ArrayValidator {
	v.requiredUnless = fn
	return v
}

// Min sets minimum number of elements
func (v *ArrayValidator) Min(n int) *ArrayValidator {
	v.min = n
	v.minSet = true
	return v
}

// Max sets maximum number of elements
func (v *ArrayValidator) Max(n int) *ArrayValidator {
	v.max = n
	v.maxSet = true
	return v
}

// Length sets exact number of elements
func (v *ArrayValidator) Length(n int) *ArrayValidator {
	v.length = n
	v.lengthSet = true
	return v
}

// Of sets the validator for each element
func (v *ArrayValidator) Of(validator Validator) *ArrayValidator {
	v.element = validator
	return v
}

// Unique requires all elements to be unique
func (v *ArrayValidator) Unique() *ArrayValidator {
	v.unique = true
	return v
}

// Distinct is an alias for Unique (Laravel naming)
func (v *ArrayValidator) Distinct() *ArrayValidator {
	return v.Unique()
}

// Contains requires array to contain the specified values
func (v *ArrayValidator) Contains(values ...any) *ArrayValidator {
	v.contains = values
	return v
}

// DoesntContain requires array to NOT contain the specified values
func (v *ArrayValidator) DoesntContain(values ...any) *ArrayValidator {
	v.doesntContain = values
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
func (v *ArrayValidator) Message(rule, message string) *ArrayValidator {
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

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		return nil
	}

	// Type check
	arr, ok := value.([]any)
	if !ok {
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be an array", fieldName)))
		return errors
	}

	length := len(arr)

	// Length check
	if v.lengthSet && length != v.length {
		errors[fieldPath] = append(errors[fieldPath], v.msg("length", fmt.Sprintf("%s must have exactly %d elements", fieldName, v.length)))
	}

	// Min check
	if v.minSet && length < v.min {
		errors[fieldPath] = append(errors[fieldPath], v.msg("min", fmt.Sprintf("%s must have at least %d elements", fieldName, v.min)))
	}

	// Max check
	if v.maxSet && length > v.max {
		errors[fieldPath] = append(errors[fieldPath], v.msg("max", fmt.Sprintf("%s must have at most %d elements", fieldName, v.max)))
	}

	// Unique check
	if v.unique {
		seen := make(map[any]bool)
		for i, item := range arr {
			if seen[item] {
				elementPath := fmt.Sprintf("%s.%d", fieldPath, i)
				errors[elementPath] = append(errors[elementPath], v.msg("unique", fmt.Sprintf("%s[%d] is a duplicate", fieldName, i)))
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
				errors[fieldPath] = append(errors[fieldPath], v.msg("contains", fmt.Sprintf("%s must contain %v", fieldName, required)))
			}
		}
	}

	// DoesntContain check - array must not contain any of the specified values
	if len(v.doesntContain) > 0 {
		for _, forbidden := range v.doesntContain {
			for i, item := range arr {
				if equalValues(item, forbidden) {
					elementPath := fmt.Sprintf("%s.%d", fieldPath, i)
					errors[elementPath] = append(errors[elementPath], v.msg("doesntContain", fmt.Sprintf("%s must not contain %v", fieldName, forbidden)))
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
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error()))
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

func (v *ArrayValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// equalValues compares two values for equality using reflect.DeepEqual
func equalValues(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
