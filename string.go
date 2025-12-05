package valet

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pre-compiled regex patterns for better performance
var (
	macRegex       = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	ulidRegex      = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	alphaDashRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// UrlOptions for URL validation
type UrlOptions struct {
	Http  bool
	Https bool
}

// StringTransformFunc is a function that transforms a string
type StringTransformFunc func(string) string

// StringValidator validates string values with fluent API
type StringValidator struct {
	required        bool
	requiredIf      func(data DataObject) bool
	requiredUnless  func(data DataObject) bool
	min             int
	minSet          bool
	max             int
	maxSet          bool
	email           bool
	url             bool
	urlOptions      *UrlOptions
	alpha           bool
	alphaNumeric    bool
	regex           *regexp.Regexp
	regexPattern    string
	notRegex        *regexp.Regexp
	in              []string
	notIn           []string
	startsWith      string
	endsWith        string
	doesntStartWith []string
	doesntEndWith   []string
	contains        string
	trim            bool
	lowercase       bool
	uppercase       bool
	exists          *ExistsRule
	unique          *UniqueRule
	customFn        func(value string, lookup Lookup) error
	messages        map[string]MessageArg
	defaultValue    *string
	nullable        bool
	transforms      []StringTransformFunc
	sameAs          string
	differentFrom   string
	// New format validations
	uuid       bool
	ip         bool
	ipv4       bool
	ipv6       bool
	json       bool
	hexColor   bool
	ascii      bool
	base64     bool
	mac        bool
	ulid       bool
	alphaDash  bool
	digitsLen  int
	digitsSet  bool
	includes   []string
	catchValue *string
}

// String creates a new string validator
func String() *StringValidator {
	return &StringValidator{
		messages: make(map[string]MessageArg),
	}
}

// Required marks the field as required
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Required(message ...MessageArg) *StringValidator {
	v.required = true
	if len(message) > 0 {
		v.messages["required"] = message[0]
	}
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
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Min(n int, message ...MessageArg) *StringValidator {
	v.min = n
	v.minSet = true
	if len(message) > 0 {
		v.messages["min"] = message[0]
	}
	return v
}

// Max sets maximum length
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Max(n int, message ...MessageArg) *StringValidator {
	v.max = n
	v.maxSet = true
	if len(message) > 0 {
		v.messages["max"] = message[0]
	}
	return v
}

// Length sets exact length (min = max = n)
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Length(n int, message ...MessageArg) *StringValidator {
	v.min = n
	v.max = n
	v.minSet = true
	v.maxSet = true
	if len(message) > 0 {
		v.messages["length"] = message[0]
	}
	return v
}

// Email validates email format
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Email(message ...MessageArg) *StringValidator {
	v.email = true
	if len(message) > 0 {
		v.messages["email"] = message[0]
	}
	return v
}

// URL validates URL format
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) URL(message ...MessageArg) *StringValidator {
	v.url = true
	if len(message) > 0 {
		v.messages["url"] = message[0]
	}
	return v
}

// URLWithOptions validates URL with specific scheme requirements
func (v *StringValidator) URLWithOptions(opts UrlOptions) *StringValidator {
	v.url = true
	v.urlOptions = &opts
	return v
}

// StartsWith validates string starts with prefix
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) StartsWith(prefix string, message ...MessageArg) *StringValidator {
	v.startsWith = prefix
	if len(message) > 0 {
		v.messages["startsWith"] = message[0]
	}
	return v
}

// EndsWith validates string ends with suffix
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) EndsWith(suffix string, message ...MessageArg) *StringValidator {
	v.endsWith = suffix
	if len(message) > 0 {
		v.messages["endsWith"] = message[0]
	}
	return v
}

// DoesntStartWith validates string does NOT start with any of the prefixes
func (v *StringValidator) DoesntStartWith(prefixes ...string) *StringValidator {
	v.doesntStartWith = prefixes
	return v
}

// DoesntEndWith validates string does NOT end with any of the suffixes
func (v *StringValidator) DoesntEndWith(suffixes ...string) *StringValidator {
	v.doesntEndWith = suffixes
	return v
}

// Contains validates string contains substring
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Contains(substr string, message ...MessageArg) *StringValidator {
	v.contains = substr
	if len(message) > 0 {
		v.messages["contains"] = message[0]
	}
	return v
}

// Includes validates string contains all of the specified substrings
func (v *StringValidator) Includes(substrs ...string) *StringValidator {
	v.includes = substrs
	return v
}

// Alpha validates only letters
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Alpha(message ...MessageArg) *StringValidator {
	v.alpha = true
	if len(message) > 0 {
		v.messages["alpha"] = message[0]
	}
	return v
}

// AlphaNumeric validates only letters and numbers
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) AlphaNumeric(message ...MessageArg) *StringValidator {
	v.alphaNumeric = true
	if len(message) > 0 {
		v.messages["alphaNumeric"] = message[0]
	}
	return v
}

// ASCII validates string contains only ASCII characters
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) ASCII(message ...MessageArg) *StringValidator {
	v.ascii = true
	if len(message) > 0 {
		v.messages["ascii"] = message[0]
	}
	return v
}

// UUID validates UUID format (v1-v5)
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) UUID(message ...MessageArg) *StringValidator {
	v.uuid = true
	if len(message) > 0 {
		v.messages["uuid"] = message[0]
	}
	return v
}

// IP validates IPv4 or IPv6 address
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) IP(message ...MessageArg) *StringValidator {
	v.ip = true
	if len(message) > 0 {
		v.messages["ip"] = message[0]
	}
	return v
}

// IPv4 validates IPv4 address only
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) IPv4(message ...MessageArg) *StringValidator {
	v.ipv4 = true
	if len(message) > 0 {
		v.messages["ipv4"] = message[0]
	}
	return v
}

// IPv6 validates IPv6 address only
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) IPv6(message ...MessageArg) *StringValidator {
	v.ipv6 = true
	if len(message) > 0 {
		v.messages["ipv6"] = message[0]
	}
	return v
}

// JSON validates string is valid JSON
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) JSON(message ...MessageArg) *StringValidator {
	v.json = true
	if len(message) > 0 {
		v.messages["json"] = message[0]
	}
	return v
}

// HexColor validates hex color format (#RGB or #RRGGBB)
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) HexColor(message ...MessageArg) *StringValidator {
	v.hexColor = true
	if len(message) > 0 {
		v.messages["hexColor"] = message[0]
	}
	return v
}

// Base64 validates string is valid base64 encoded
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Base64(message ...MessageArg) *StringValidator {
	v.base64 = true
	if len(message) > 0 {
		v.messages["base64"] = message[0]
	}
	return v
}

// MAC validates MAC address format
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) MAC(message ...MessageArg) *StringValidator {
	v.mac = true
	if len(message) > 0 {
		v.messages["mac"] = message[0]
	}
	return v
}

// ULID validates ULID format (26 characters, Crockford base32)
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) ULID(message ...MessageArg) *StringValidator {
	v.ulid = true
	if len(message) > 0 {
		v.messages["ulid"] = message[0]
	}
	return v
}

// AlphaDash validates string contains only alphanumeric, dash, and underscore
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) AlphaDash(message ...MessageArg) *StringValidator {
	v.alphaDash = true
	if len(message) > 0 {
		v.messages["alphaDash"] = message[0]
	}
	return v
}

// Digits validates string is numeric with exact length
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Digits(length int, message ...MessageArg) *StringValidator {
	v.digitsLen = length
	v.digitsSet = true
	if len(message) > 0 {
		v.messages["digits"] = message[0]
	}
	return v
}

// Catch sets a fallback value to use if validation fails
func (v *StringValidator) Catch(value string) *StringValidator {
	v.catchValue = &value
	return v
}

// Regex validates against pattern
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) Regex(pattern string, message ...MessageArg) *StringValidator {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.regex = re
	}
	v.regexPattern = pattern
	if len(message) > 0 {
		v.messages["regex"] = message[0]
	}
	return v
}

// NotRegex validates value does NOT match pattern
// Optionally accepts a custom error message (string or MessageFunc)
func (v *StringValidator) NotRegex(pattern string, message ...MessageArg) *StringValidator {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err == nil {
		v.notRegex = re
	}
	if len(message) > 0 {
		v.messages["notRegex"] = message[0]
	}
	return v
}

// In validates value is one of allowed values
// Note: To add a custom message, use .Message("in", "your message") after this call
func (v *StringValidator) In(values ...string) *StringValidator {
	v.in = values
	return v
}

// InWithMessage validates value is one of allowed values with a custom message
func (v *StringValidator) InWithMessage(message MessageArg, values ...string) *StringValidator {
	v.in = values
	v.messages["in"] = message
	return v
}

// NotIn validates value is not one of disallowed values
// Note: To add a custom message, use .Message("notIn", "your message") after this call
func (v *StringValidator) NotIn(values ...string) *StringValidator {
	v.notIn = values
	return v
}

// NotInWithMessage validates value is not one of disallowed values with a custom message
func (v *StringValidator) NotInWithMessage(message MessageArg, values ...string) *StringValidator {
	v.notIn = values
	v.messages["notIn"] = message
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
// message can be a string or MessageFunc for dynamic messages
func (v *StringValidator) Message(rule string, message MessageArg) *StringValidator {
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

// Transform adds a transformation function to be applied before validation
func (v *StringValidator) Transform(fn StringTransformFunc) *StringValidator {
	v.transforms = append(v.transforms, fn)
	return v
}

// SameAs validates that this field equals another field's value
func (v *StringValidator) SameAs(fieldPath string) *StringValidator {
	v.sameAs = fieldPath
	return v
}

// DifferentFrom validates that this field differs from another field's value
func (v *StringValidator) DifferentFrom(fieldPath string) *StringValidator {
	v.differentFrom = fieldPath
	return v
}

// Validate implements Validator interface
func (v *StringValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
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

	// Type check
	str, ok := value.(string)
	if !ok {
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be a string", fieldName), msgCtx))
		return errors
	}

	// Update msgCtx with the string value
	msgCtx.Value = str

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
	// Apply custom transforms
	for _, transform := range v.transforms {
		str = transform(str)
	}

	// Empty string check for required
	if str == "" {
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

	length := utf8.RuneCountInString(str)

	// Min length (check for "length" message first if min == max, for Length() use case)
	if v.minSet && length < v.min {
		msgCtx.Param = v.min
		rule := "min"
		if v.maxSet && v.min == v.max {
			if _, ok := v.messages["length"]; ok {
				rule = "length"
			}
		}
		errors[fieldPath] = append(errors[fieldPath], v.msg(rule, fmt.Sprintf("%s must be at least %d characters", fieldName, v.min), msgCtx))
	}

	// Max length (check for "length" message first if min == max, for Length() use case)
	if v.maxSet && length > v.max {
		msgCtx.Param = v.max
		rule := "max"
		if v.minSet && v.min == v.max {
			if _, ok := v.messages["length"]; ok {
				rule = "length"
			}
		}
		errors[fieldPath] = append(errors[fieldPath], v.msg(rule, fmt.Sprintf("%s must be at most %d characters", fieldName, v.max), msgCtx))
	}

	// Email
	if v.email && !isValidEmail(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("email", fmt.Sprintf("%s must be a valid email", fieldName), msgCtx))
	}

	// URL
	if v.url {
		if !isValidURL(str) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be a valid URL", fieldName), msgCtx))
		} else if v.urlOptions != nil {
			u, _ := url.Parse(str)
			if v.urlOptions.Http && !v.urlOptions.Https && u.Scheme != "http" {
				errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be an HTTP URL", fieldName), msgCtx))
			} else if v.urlOptions.Https && !v.urlOptions.Http && u.Scheme != "https" {
				errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be an HTTPS URL", fieldName), msgCtx))
			} else if v.urlOptions.Http && v.urlOptions.Https && u.Scheme != "http" && u.Scheme != "https" {
				errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be an HTTP or HTTPS URL", fieldName), msgCtx))
			}
		}
	}

	// StartsWith
	if v.startsWith != "" && !strings.HasPrefix(str, v.startsWith) {
		msgCtx.Param = v.startsWith
		errors[fieldPath] = append(errors[fieldPath], v.msg("startsWith", fmt.Sprintf("%s must start with %s", fieldName, v.startsWith), msgCtx))
	}

	// EndsWith
	if v.endsWith != "" && !strings.HasSuffix(str, v.endsWith) {
		msgCtx.Param = v.endsWith
		errors[fieldPath] = append(errors[fieldPath], v.msg("endsWith", fmt.Sprintf("%s must end with %s", fieldName, v.endsWith), msgCtx))
	}

	// Contains
	if v.contains != "" && !strings.Contains(str, v.contains) {
		msgCtx.Param = v.contains
		errors[fieldPath] = append(errors[fieldPath], v.msg("contains", fmt.Sprintf("%s must contain %s", fieldName, v.contains), msgCtx))
	}

	// Alpha
	if v.alpha && !isAlpha(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("alpha", fmt.Sprintf("%s must contain only letters", fieldName), msgCtx))
	}

	// AlphaNumeric
	if v.alphaNumeric && !isAlphaNumeric(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("alphaNumeric", fmt.Sprintf("%s must contain only letters and numbers", fieldName), msgCtx))
	}

	// Regex
	if v.regex != nil && !v.regex.MatchString(str) {
		msgCtx.Param = v.regexPattern
		errors[fieldPath] = append(errors[fieldPath], v.msg("regex", fmt.Sprintf("%s format is invalid", fieldName), msgCtx))
	}

	// NotRegex
	if v.notRegex != nil && v.notRegex.MatchString(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("notRegex", fmt.Sprintf("%s format is invalid", fieldName), msgCtx))
	}

	// In
	if len(v.in) > 0 && !contains(v.in, str) {
		msgCtx.Param = v.in
		errors[fieldPath] = append(errors[fieldPath], v.msg("in", fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(v.in, ", ")), msgCtx))
	}

	// NotIn
	if len(v.notIn) > 0 && contains(v.notIn, str) {
		msgCtx.Param = v.notIn
		errors[fieldPath] = append(errors[fieldPath], v.msg("notIn", fmt.Sprintf("%s must not be one of: %s", fieldName, strings.Join(v.notIn, ", ")), msgCtx))
	}

	// DoesntStartWith (array of prefixes)
	for _, prefix := range v.doesntStartWith {
		if strings.HasPrefix(str, prefix) {
			msgCtx.Param = v.doesntStartWith
			errors[fieldPath] = append(errors[fieldPath], v.msg("doesntStartWith", fmt.Sprintf("%s must not start with %s", fieldName, prefix), msgCtx))
			break
		}
	}

	// DoesntEndWith (array of suffixes)
	for _, suffix := range v.doesntEndWith {
		if strings.HasSuffix(str, suffix) {
			msgCtx.Param = v.doesntEndWith
			errors[fieldPath] = append(errors[fieldPath], v.msg("doesntEndWith", fmt.Sprintf("%s must not end with %s", fieldName, suffix), msgCtx))
			break
		}
	}

	// Includes (must contain all substrings)
	for _, substr := range v.includes {
		if !strings.Contains(str, substr) {
			msgCtx.Param = v.includes
			errors[fieldPath] = append(errors[fieldPath], v.msg("includes", fmt.Sprintf("%s must contain %s", fieldName, substr), msgCtx))
		}
	}

	// UUID
	if v.uuid && !isValidUUID(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("uuid", fmt.Sprintf("%s must be a valid UUID", fieldName), msgCtx))
	}

	// IP (v4 or v6)
	if v.ip && !isValidIP(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ip", fmt.Sprintf("%s must be a valid IP address", fieldName), msgCtx))
	}

	// IPv4 only
	if v.ipv4 && !isValidIPv4(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ipv4", fmt.Sprintf("%s must be a valid IPv4 address", fieldName), msgCtx))
	}

	// IPv6 only
	if v.ipv6 && !isValidIPv6(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ipv6", fmt.Sprintf("%s must be a valid IPv6 address", fieldName), msgCtx))
	}

	// JSON
	if v.json && !isValidJSON(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("json", fmt.Sprintf("%s must be valid JSON", fieldName), msgCtx))
	}

	// HexColor
	if v.hexColor && !isValidHexColor(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("hexColor", fmt.Sprintf("%s must be a valid hex color", fieldName), msgCtx))
	}

	// ASCII
	if v.ascii && !isASCII(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ascii", fmt.Sprintf("%s must contain only ASCII characters", fieldName), msgCtx))
	}

	// Base64
	if v.base64 && !isValidBase64(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("base64", fmt.Sprintf("%s must be valid base64", fieldName), msgCtx))
	}

	// MAC address
	if v.mac && !isValidMAC(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("mac", fmt.Sprintf("%s must be a valid MAC address", fieldName), msgCtx))
	}

	// ULID
	if v.ulid && !isValidULID(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ulid", fmt.Sprintf("%s must be a valid ULID", fieldName), msgCtx))
	}

	// AlphaDash (letters, numbers, dashes, underscores)
	if v.alphaDash && !isAlphaDash(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("alphaDash", fmt.Sprintf("%s must contain only letters, numbers, dashes, and underscores", fieldName), msgCtx))
	}

	// Digits (exact length numeric string)
	if v.digitsSet && !isDigits(str, v.digitsLen) {
		msgCtx.Param = v.digitsLen
		errors[fieldPath] = append(errors[fieldPath], v.msg("digits", fmt.Sprintf("%s must be exactly %d digits", fieldName, v.digitsLen), msgCtx))
	}

	// SameAs - cross-field equality check
	if v.sameAs != "" {
		otherValue := lookupPath(ctx.RootData, v.sameAs)
		if otherValue.Exists() {
			if otherStr, ok := otherValue.Value().(string); ok {
				if str != otherStr {
					msgCtx.Param = v.sameAs
					errors[fieldPath] = append(errors[fieldPath], v.msg("sameAs", fmt.Sprintf("%s must match %s", fieldName, v.sameAs), msgCtx))
				}
			}
		}
	}

	// DifferentFrom - cross-field difference check
	if v.differentFrom != "" {
		otherValue := lookupPath(ctx.RootData, v.differentFrom)
		if otherValue.Exists() {
			if otherStr, ok := otherValue.Value().(string); ok {
				if str == otherStr {
					msgCtx.Param = v.differentFrom
					errors[fieldPath] = append(errors[fieldPath], v.msg("differentFrom", fmt.Sprintf("%s must be different from %s", fieldName, v.differentFrom), msgCtx))
				}
			}
		}
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(str, lookup); err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error(), msgCtx))
		}
	}

	if len(errors) == 0 {
		return nil
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

func (v *StringValidator) msg(rule, defaultMsg string, msgCtx MessageContext) string {
	if msg, ok := v.messages[rule]; ok {
		msgCtx.Rule = rule
		return resolveMessage(msg, msgCtx)
	}
	return defaultMsg
}

// Helper functions
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var alphaRegex = regexp.MustCompile(`^[a-zA-Z]+$`)
var alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

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

func isValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

func isValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

func isValidIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

func isValidIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}

func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

func isValidHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func isValidMAC(s string) bool {
	// MAC address formats: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
	return macRegex.MatchString(s)
}

func isValidULID(s string) bool {
	// ULID: 26 characters, Crockford base32 (excludes I, L, O, U)
	if len(s) != 26 {
		return false
	}
	return ulidRegex.MatchString(strings.ToUpper(s))
}

func isAlphaDash(s string) bool {
	// Only letters, numbers, dashes, and underscores
	return alphaDashRegex.MatchString(s)
}

func isDigits(s string, length int) bool {
	if len(s) != length {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
