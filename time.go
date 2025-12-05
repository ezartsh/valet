package valet

import (
	"fmt"
	"time"
)

// TimeValidator validates time values with fluent API
type TimeValidator struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	format         string
	after          *time.Time
	afterField     string
	before         *time.Time
	beforeField    string
	betweenStart   *time.Time
	betweenEnd     *time.Time
	customFn       func(value time.Time, lookup Lookup) error
	messages       map[string]string
	defaultValue   *time.Time
	nullable       bool
	timezone       *time.Location
}

// Time creates a new time validator
func Time() *TimeValidator {
	return &TimeValidator{
		messages: make(map[string]string),
		format:   time.RFC3339, // Default format
	}
}

// Required marks the field as required
func (v *TimeValidator) Required() *TimeValidator {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *TimeValidator) RequiredIf(fn func(data DataObject) bool) *TimeValidator {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *TimeValidator) RequiredUnless(fn func(data DataObject) bool) *TimeValidator {
	v.requiredUnless = fn
	return v
}

// Format sets the expected time format (default: RFC3339)
func (v *TimeValidator) Format(format string) *TimeValidator {
	v.format = format
	return v
}

// After validates time is after the specified time
func (v *TimeValidator) After(t time.Time) *TimeValidator {
	v.after = &t
	return v
}

// AfterField validates time is after another field's time value
func (v *TimeValidator) AfterField(fieldPath string) *TimeValidator {
	v.afterField = fieldPath
	return v
}

// AfterNow validates time is after current time
func (v *TimeValidator) AfterNow() *TimeValidator {
	now := time.Now()
	v.after = &now
	return v
}

// Before validates time is before the specified time
func (v *TimeValidator) Before(t time.Time) *TimeValidator {
	v.before = &t
	return v
}

// BeforeField validates time is before another field's time value
func (v *TimeValidator) BeforeField(fieldPath string) *TimeValidator {
	v.beforeField = fieldPath
	return v
}

// BeforeNow validates time is before current time
func (v *TimeValidator) BeforeNow() *TimeValidator {
	now := time.Now()
	v.before = &now
	return v
}

// Between validates time is between start and end (inclusive)
func (v *TimeValidator) Between(start, end time.Time) *TimeValidator {
	v.betweenStart = &start
	v.betweenEnd = &end
	return v
}

// Timezone sets the timezone for parsing
func (v *TimeValidator) Timezone(loc *time.Location) *TimeValidator {
	v.timezone = loc
	return v
}

// Custom adds custom validation function
func (v *TimeValidator) Custom(fn func(value time.Time, lookup Lookup) error) *TimeValidator {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *TimeValidator) Message(rule, message string) *TimeValidator {
	v.messages[rule] = message
	return v
}

// Default sets default value if field is empty/missing
func (v *TimeValidator) Default(value time.Time) *TimeValidator {
	v.defaultValue = &value
	return v
}

// Nullable allows null values
func (v *TimeValidator) Nullable() *TimeValidator {
	v.nullable = true
	return v
}

// Validate implements Validator interface
func (v *TimeValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.defaultValue != nil {
			value = v.defaultValue.Format(v.format)
		} else if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		} else {
			return nil
		}
	}

	// Parse time from string or time.Time
	var t time.Time
	var err error

	switch val := value.(type) {
	case time.Time:
		t = val
	case string:
		if val == "" {
			if v.required {
				errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
				return errors
			}
			return nil
		}
		if v.timezone != nil {
			t, err = time.ParseInLocation(v.format, val, v.timezone)
		} else {
			t, err = time.Parse(v.format, val)
		}
		if err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("format", fmt.Sprintf("%s must be a valid time format", fieldName)))
			return errors
		}
	default:
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be a time value", fieldName)))
		return errors
	}

	// Create lookup function
	lookup := func(path string) LookupResult {
		return lookupPath(ctx.RootData, path)
	}

	// After validation
	if v.after != nil && !t.After(*v.after) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("after", fmt.Sprintf("%s must be after %s", fieldName, v.after.Format(v.format))))
	}

	// AfterField validation
	if v.afterField != "" {
		afterResult := lookup(v.afterField)
		if afterResult.Exists() {
			if afterStr, ok := afterResult.Value().(string); ok {
				var afterTime time.Time
				var parseErr error
				if v.timezone != nil {
					afterTime, parseErr = time.ParseInLocation(v.format, afterStr, v.timezone)
				} else {
					afterTime, parseErr = time.Parse(v.format, afterStr)
				}
				if parseErr == nil {
					if !t.After(afterTime) {
						errors[fieldPath] = append(errors[fieldPath], v.msg("afterField", fmt.Sprintf("%s must be after %s", fieldName, v.afterField)))
					}
				}
			} else if afterTime, ok := afterResult.Value().(time.Time); ok {
				if !t.After(afterTime) {
					errors[fieldPath] = append(errors[fieldPath], v.msg("afterField", fmt.Sprintf("%s must be after %s", fieldName, v.afterField)))
				}
			}
		}
	}

	// Before validation
	if v.before != nil && !t.Before(*v.before) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("before", fmt.Sprintf("%s must be before %s", fieldName, v.before.Format(v.format))))
	}

	// BeforeField validation
	if v.beforeField != "" {
		beforeResult := lookup(v.beforeField)
		if beforeResult.Exists() {
			if beforeStr, ok := beforeResult.Value().(string); ok {
				var beforeTime time.Time
				var parseErr error
				if v.timezone != nil {
					beforeTime, parseErr = time.ParseInLocation(v.format, beforeStr, v.timezone)
				} else {
					beforeTime, parseErr = time.Parse(v.format, beforeStr)
				}
				if parseErr == nil {
					if !t.Before(beforeTime) {
						errors[fieldPath] = append(errors[fieldPath], v.msg("beforeField", fmt.Sprintf("%s must be before %s", fieldName, v.beforeField)))
					}
				}
			} else if beforeTime, ok := beforeResult.Value().(time.Time); ok {
				if !t.Before(beforeTime) {
					errors[fieldPath] = append(errors[fieldPath], v.msg("beforeField", fmt.Sprintf("%s must be before %s", fieldName, v.beforeField)))
				}
			}
		}
	}

	// Between validation
	if v.betweenStart != nil && v.betweenEnd != nil {
		if t.Before(*v.betweenStart) || t.After(*v.betweenEnd) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("between", fmt.Sprintf("%s must be between %s and %s", fieldName, v.betweenStart.Format(v.format), v.betweenEnd.Format(v.format))))
		}
	}

	// Custom validation
	if v.customFn != nil {
		if err := v.customFn(t, lookup); err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error()))
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func (v *TimeValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}
