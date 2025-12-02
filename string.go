package valet

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// UrlOptions for URL validation
type UrlOptions struct {
	Http  bool
	Https bool
}

// StringValidator validates string values with fluent API
type StringValidator struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	min            int
	minSet         bool
	max            int
	maxSet         bool
	email          bool
	url            bool
	urlOptions     *UrlOptions
	alpha          bool
	alphaNumeric   bool
	regex          *regexp.Regexp
	regexPattern   string
	notRegex       *regexp.Regexp
	in             []string
	notIn          []string
	startsWith     string
	endsWith       string
	contains       string
	trim           bool
	lowercase      bool
	uppercase      bool
	exists         *ExistsRule
	unique         *UniqueRule
	customFn       func(value string, lookup Lookup) error
	messages       map[string]string
	defaultValue   *string
	nullable       bool
}

// String creates a new string validator
func String() *StringValidator {
	return &StringValidator{
		messages: make(map[string]string),
	}
}

// Required marks the field as required
func (v *StringValidator) Required() *StringValidator {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *StringValidator) RequiredIf(fn func(data DataObject) bool) *StringValidator {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *StringValidator) RequiredUnless(fn func(data DataObject) bool) *StringValidator {
	v.requiredUnless = fn
	return v
}

// Min sets minimum length
func (v *StringValidator) Min(n int) *StringValidator {
	v.min = n
	v.minSet = true
	return v
}

// Max sets maximum length
func (v *StringValidator) Max(n int) *StringValidator {
	v.max = n
	v.maxSet = true
	return v
}

// Length sets exact length (min = max = n)
func (v *StringValidator) Length(n int) *StringValidator {
	v.min = n
	v.max = n
	v.minSet = true
	v.maxSet = true
	return v
}

// Email validates email format
func (v *StringValidator) Email() *StringValidator {
	v.email = true
	return v
}

// URL validates URL format
func (v *StringValidator) URL() *StringValidator {
	v.url = true
	return v
}

// URLWithOptions validates URL with specific scheme requirements
func (v *StringValidator) URLWithOptions(opts UrlOptions) *StringValidator {
	v.url = true
	v.urlOptions = &opts
	return v
}

// StartsWith validates string starts with prefix
func (v *StringValidator) StartsWith(prefix string) *StringValidator {
	v.startsWith = prefix
	return v
}

// EndsWith validates string ends with suffix
func (v *StringValidator) EndsWith(suffix string) *StringValidator {
	v.endsWith = suffix
	return v
}

// Contains validates string contains substring
func (v *StringValidator) Contains(substr string) *StringValidator {
	v.contains = substr
	return v
}

// Alpha validates only letters
func (v *StringValidator) Alpha() *StringValidator {
	v.alpha = true
	return v
}

// AlphaNumeric validates only letters and numbers
func (v *StringValidator) AlphaNumeric() *StringValidator {
	v.alphaNumeric = true
	return v
}

// Regex validates against pattern
func (v *StringValidator) Regex(pattern string) *StringValidator {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.regex = re
	}
	v.regexPattern = pattern
	return v
}

// NotRegex validates value does NOT match pattern
func (v *StringValidator) NotRegex(pattern string) *StringValidator {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.notRegex = re
	}
	return v
}

// In validates value is one of allowed values
func (v *StringValidator) In(values ...string) *StringValidator {
	v.in = values
	return v
}

// NotIn validates value is not one of disallowed values
func (v *StringValidator) NotIn(values ...string) *StringValidator {
	v.notIn = values
	return v
}

// Trim trims whitespace before validation
func (v *StringValidator) Trim() *StringValidator {
	v.trim = true
	return v
}

// Lowercase converts to lowercase before validation
func (v *StringValidator) Lowercase() *StringValidator {
	v.lowercase = true
	return v
}

// Uppercase converts to uppercase before validation
func (v *StringValidator) Uppercase() *StringValidator {
	v.uppercase = true
	return v
}

// Exists adds database existence check
func (v *StringValidator) Exists(table, column string, wheres ...WhereClause) *StringValidator {
	v.exists = &ExistsRule{Table: table, Column: column, Where: wheres}
	return v
}

// Unique adds database uniqueness check
func (v *StringValidator) Unique(table, column string, ignore any, wheres ...WhereClause) *StringValidator {
	v.unique = &UniqueRule{Table: table, Column: column, Ignore: ignore, Where: wheres}
	return v
}

// Custom adds custom validation function
func (v *StringValidator) Custom(fn func(value string, lookup Lookup) error) *StringValidator {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *StringValidator) Message(rule, message string) *StringValidator {
	v.messages[rule] = message
	return v
}

// Default sets default value if field is empty/missing
func (v *StringValidator) Default(value string) *StringValidator {
	v.defaultValue = &value
	return v
}

// Nullable allows null values
func (v *StringValidator) Nullable() *StringValidator {
	v.nullable = true
	return v
}

// Validate implements Validator interface
func (v *StringValidator) Validate(ctx *ValidationContext, value any) []string {
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

	// Type check
	str, ok := value.(string)
	if !ok {
		errors = append(errors, v.msg("type", fmt.Sprintf("%s must be a string", fieldName)))
		return errors
	}

	// Transformations
	if v.trim {
		str = strings.TrimSpace(str)
	}
	if v.lowercase {
		str = strings.ToLower(str)
	}
	if v.uppercase {
		str = strings.ToUpper(str)
	}

	// Empty string check for required
	if str == "" {
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

	length := utf8.RuneCountInString(str)

	// Min length
	if v.minSet && length < v.min {
		errors = append(errors, v.msg("min", fmt.Sprintf("%s must be at least %d characters", fieldName, v.min)))
	}

	// Max length
	if v.maxSet && length > v.max {
		errors = append(errors, v.msg("max", fmt.Sprintf("%s must be at most %d characters", fieldName, v.max)))
	}

	// Email
	if v.email && !isValidEmail(str) {
		errors = append(errors, v.msg("email", fmt.Sprintf("%s must be a valid email", fieldName)))
	}

	// URL
	if v.url {
		if !isValidURL(str) {
			errors = append(errors, v.msg("url", fmt.Sprintf("%s must be a valid URL", fieldName)))
		} else if v.urlOptions != nil {
			u, _ := url.Parse(str)
			if v.urlOptions.Http && !v.urlOptions.Https && u.Scheme != "http" {
				errors = append(errors, v.msg("url", fmt.Sprintf("%s must be an HTTP URL", fieldName)))
			} else if v.urlOptions.Https && !v.urlOptions.Http && u.Scheme != "https" {
				errors = append(errors, v.msg("url", fmt.Sprintf("%s must be an HTTPS URL", fieldName)))
			} else if v.urlOptions.Http && v.urlOptions.Https && u.Scheme != "http" && u.Scheme != "https" {
				errors = append(errors, v.msg("url", fmt.Sprintf("%s must be an HTTP or HTTPS URL", fieldName)))
			}
		}
	}

	// StartsWith
	if v.startsWith != "" && !strings.HasPrefix(str, v.startsWith) {
		errors = append(errors, v.msg("startsWith", fmt.Sprintf("%s must start with %s", fieldName, v.startsWith)))
	}

	// EndsWith
	if v.endsWith != "" && !strings.HasSuffix(str, v.endsWith) {
		errors = append(errors, v.msg("endsWith", fmt.Sprintf("%s must end with %s", fieldName, v.endsWith)))
	}

	// Contains
	if v.contains != "" && !strings.Contains(str, v.contains) {
		errors = append(errors, v.msg("contains", fmt.Sprintf("%s must contain %s", fieldName, v.contains)))
	}

	// Alpha
	if v.alpha && !isAlpha(str) {
		errors = append(errors, v.msg("alpha", fmt.Sprintf("%s must contain only letters", fieldName)))
	}

	// AlphaNumeric
	if v.alphaNumeric && !isAlphaNumeric(str) {
		errors = append(errors, v.msg("alphaNumeric", fmt.Sprintf("%s must contain only letters and numbers", fieldName)))
	}

	// Regex
	if v.regex != nil && !v.regex.MatchString(str) {
		errors = append(errors, v.msg("regex", fmt.Sprintf("%s format is invalid", fieldName)))
	}

	// NotRegex
	if v.notRegex != nil && v.notRegex.MatchString(str) {
		errors = append(errors, v.msg("notRegex", fmt.Sprintf("%s format is invalid", fieldName)))
	}

	// In
	if len(v.in) > 0 && !contains(v.in, str) {
		errors = append(errors, v.msg("in", fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(v.in, ", "))))
	}

	// NotIn
	if len(v.notIn) > 0 && contains(v.notIn, str) {
		errors = append(errors, v.msg("notIn", fmt.Sprintf("%s must not be one of: %s", fieldName, strings.Join(v.notIn, ", "))))
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(str, lookup); err != nil {
			errors = append(errors, v.msg("custom", err.Error()))
		}
	}

	return errors
}

// GetDBChecks returns database checks for this field
func (v *StringValidator) GetDBChecks(fieldPath string, value any) []DBCheck {
	var checks []DBCheck

	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}

	if v.exists != nil {
		checks = append(checks, DBCheck{
			Field:    fieldPath,
			Value:    str,
			Rule:     *v.exists,
			IsUnique: false,
			Message:  v.messages["exists"],
		})
	}

	if v.unique != nil {
		checks = append(checks, DBCheck{
			Field:    fieldPath,
			Value:    str,
			Rule:     ExistsRule{Table: v.unique.Table, Column: v.unique.Column, Where: v.unique.Where},
			IsUnique: true,
			Ignore:   v.unique.Ignore,
			Message:  v.messages["unique"],
		})
	}

	return checks
}

func (v *StringValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// Helper functions
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var alphaRegex = regexp.MustCompile(`^[a-zA-Z]+$`)
var alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func isValidEmail(s string) bool {
	return emailRegex.MatchString(s)
}

func isValidURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isAlpha(s string) bool {
	return alphaRegex.MatchString(s)
}

func isAlphaNumeric(s string) bool {
	return alphaNumericRegex.MatchString(s)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
