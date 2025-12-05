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
	messages        map[string]string
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
func (v *StringValidator) Contains(substr string) *StringValidator {
	v.contains = substr
	return v
}

// Includes validates string contains all of the specified substrings
func (v *StringValidator) Includes(substrs ...string) *StringValidator {
	v.includes = substrs
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

// ASCII validates string contains only ASCII characters
func (v *StringValidator) ASCII() *StringValidator {
	v.ascii = true
	return v
}

// UUID validates UUID format (v1-v5)
func (v *StringValidator) UUID() *StringValidator {
	v.uuid = true
	return v
}

// IP validates IPv4 or IPv6 address
func (v *StringValidator) IP() *StringValidator {
	v.ip = true
	return v
}

// IPv4 validates IPv4 address only
func (v *StringValidator) IPv4() *StringValidator {
	v.ipv4 = true
	return v
}

// IPv6 validates IPv6 address only
func (v *StringValidator) IPv6() *StringValidator {
	v.ipv6 = true
	return v
}

// JSON validates string is valid JSON
func (v *StringValidator) JSON() *StringValidator {
	v.json = true
	return v
}

// HexColor validates hex color format (#RGB or #RRGGBB)
func (v *StringValidator) HexColor() *StringValidator {
	v.hexColor = true
	return v
}

// Base64 validates string is valid base64 encoded
func (v *StringValidator) Base64() *StringValidator {
	v.base64 = true
	return v
}

// MAC validates MAC address format
func (v *StringValidator) MAC() *StringValidator {
	v.mac = true
	return v
}

// ULID validates ULID format (26 characters, Crockford base32)
func (v *StringValidator) ULID() *StringValidator {
	v.ulid = true
	return v
}

// AlphaDash validates string contains only alphanumeric, dash, and underscore
func (v *StringValidator) AlphaDash() *StringValidator {
	v.alphaDash = true
	return v
}

// Digits validates string is numeric with exact length
func (v *StringValidator) Digits(length int) *StringValidator {
	v.digitsLen = length
	v.digitsSet = true
	return v
}

// Catch sets a fallback value to use if validation fails
func (v *StringValidator) Catch(value string) *StringValidator {
	v.catchValue = &value
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

	// Type check
	str, ok := value.(string)
	if !ok {
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be a string", fieldName)))
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
	// Apply custom transforms
	for _, transform := range v.transforms {
		str = transform(str)
	}

	// Empty string check for required
	if str == "" {
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

	length := utf8.RuneCountInString(str)

	// Min length
	if v.minSet && length < v.min {
		errors[fieldPath] = append(errors[fieldPath], v.msg("min", fmt.Sprintf("%s must be at least %d characters", fieldName, v.min)))
	}

	// Max length
	if v.maxSet && length > v.max {
		errors[fieldPath] = append(errors[fieldPath], v.msg("max", fmt.Sprintf("%s must be at most %d characters", fieldName, v.max)))
	}

	// Email
	if v.email && !isValidEmail(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("email", fmt.Sprintf("%s must be a valid email", fieldName)))
	}

	// URL
	if v.url {
		if !isValidURL(str) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be a valid URL", fieldName)))
		} else if v.urlOptions != nil {
			u, _ := url.Parse(str)
			if v.urlOptions.Http && !v.urlOptions.Https && u.Scheme != "http" {
				errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be an HTTP URL", fieldName)))
			} else if v.urlOptions.Https && !v.urlOptions.Http && u.Scheme != "https" {
				errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be an HTTPS URL", fieldName)))
			} else if v.urlOptions.Http && v.urlOptions.Https && u.Scheme != "http" && u.Scheme != "https" {
				errors[fieldPath] = append(errors[fieldPath], v.msg("url", fmt.Sprintf("%s must be an HTTP or HTTPS URL", fieldName)))
			}
		}
	}

	// StartsWith
	if v.startsWith != "" && !strings.HasPrefix(str, v.startsWith) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("startsWith", fmt.Sprintf("%s must start with %s", fieldName, v.startsWith)))
	}

	// EndsWith
	if v.endsWith != "" && !strings.HasSuffix(str, v.endsWith) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("endsWith", fmt.Sprintf("%s must end with %s", fieldName, v.endsWith)))
	}

	// Contains
	if v.contains != "" && !strings.Contains(str, v.contains) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("contains", fmt.Sprintf("%s must contain %s", fieldName, v.contains)))
	}

	// Alpha
	if v.alpha && !isAlpha(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("alpha", fmt.Sprintf("%s must contain only letters", fieldName)))
	}

	// AlphaNumeric
	if v.alphaNumeric && !isAlphaNumeric(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("alphaNumeric", fmt.Sprintf("%s must contain only letters and numbers", fieldName)))
	}

	// Regex
	if v.regex != nil && !v.regex.MatchString(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("regex", fmt.Sprintf("%s format is invalid", fieldName)))
	}

	// NotRegex
	if v.notRegex != nil && v.notRegex.MatchString(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("notRegex", fmt.Sprintf("%s format is invalid", fieldName)))
	}

	// In
	if len(v.in) > 0 && !contains(v.in, str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("in", fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(v.in, ", "))))
	}

	// NotIn
	if len(v.notIn) > 0 && contains(v.notIn, str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("notIn", fmt.Sprintf("%s must not be one of: %s", fieldName, strings.Join(v.notIn, ", "))))
	}

	// DoesntStartWith (array of prefixes)
	for _, prefix := range v.doesntStartWith {
		if strings.HasPrefix(str, prefix) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("doesntStartWith", fmt.Sprintf("%s must not start with %s", fieldName, prefix)))
			break
		}
	}

	// DoesntEndWith (array of suffixes)
	for _, suffix := range v.doesntEndWith {
		if strings.HasSuffix(str, suffix) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("doesntEndWith", fmt.Sprintf("%s must not end with %s", fieldName, suffix)))
			break
		}
	}

	// Includes (must contain all substrings)
	for _, substr := range v.includes {
		if !strings.Contains(str, substr) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("includes", fmt.Sprintf("%s must contain %s", fieldName, substr)))
		}
	}

	// UUID
	if v.uuid && !isValidUUID(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("uuid", fmt.Sprintf("%s must be a valid UUID", fieldName)))
	}

	// IP (v4 or v6)
	if v.ip && !isValidIP(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ip", fmt.Sprintf("%s must be a valid IP address", fieldName)))
	}

	// IPv4 only
	if v.ipv4 && !isValidIPv4(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ipv4", fmt.Sprintf("%s must be a valid IPv4 address", fieldName)))
	}

	// IPv6 only
	if v.ipv6 && !isValidIPv6(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ipv6", fmt.Sprintf("%s must be a valid IPv6 address", fieldName)))
	}

	// JSON
	if v.json && !isValidJSON(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("json", fmt.Sprintf("%s must be valid JSON", fieldName)))
	}

	// HexColor
	if v.hexColor && !isValidHexColor(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("hexColor", fmt.Sprintf("%s must be a valid hex color", fieldName)))
	}

	// ASCII
	if v.ascii && !isASCII(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ascii", fmt.Sprintf("%s must contain only ASCII characters", fieldName)))
	}

	// Base64
	if v.base64 && !isValidBase64(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("base64", fmt.Sprintf("%s must be valid base64", fieldName)))
	}

	// MAC address
	if v.mac && !isValidMAC(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("mac", fmt.Sprintf("%s must be a valid MAC address", fieldName)))
	}

	// ULID
	if v.ulid && !isValidULID(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("ulid", fmt.Sprintf("%s must be a valid ULID", fieldName)))
	}

	// AlphaDash (letters, numbers, dashes, underscores)
	if v.alphaDash && !isAlphaDash(str) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("alphaDash", fmt.Sprintf("%s must contain only letters, numbers, dashes, and underscores", fieldName)))
	}

	// Digits (exact length numeric string)
	if v.digitsSet && !isDigits(str, v.digitsLen) {
		errors[fieldPath] = append(errors[fieldPath], v.msg("digits", fmt.Sprintf("%s must be exactly %d digits", fieldName, v.digitsLen)))
	}

	// SameAs - cross-field equality check
	if v.sameAs != "" {
		otherValue := lookupPath(ctx.RootData, v.sameAs)
		if otherValue.Exists() {
			if otherStr, ok := otherValue.Value().(string); ok {
				if str != otherStr {
					errors[fieldPath] = append(errors[fieldPath], v.msg("sameAs", fmt.Sprintf("%s must match %s", fieldName, v.sameAs)))
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
					errors[fieldPath] = append(errors[fieldPath], v.msg("differentFrom", fmt.Sprintf("%s must be different from %s", fieldName, v.differentFrom)))
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
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error()))
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
