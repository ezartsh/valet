package valet

import (
	"context"
	"errors"
	"strings"
)

// Common errors
var (
	ErrNilDBConnection = errors.New("database connection is nil")
)

// DataObject represents the data to validate (parsed JSON)
type DataObject = map[string]any

// DataAccessor wraps DataObject to provide convenient access methods
type DataAccessor map[string]any

// Get retrieves a value from the data using dot-notation path
// Example: data.Get("user.profile.name") or data.Get("items.0.id")
func (d DataAccessor) Get(path string) LookupResult {
	if d == nil {
		return LookupResult{nil, false}
	}
	return lookupPath(map[string]any(d), path)
}

// Schema is a map of field names to validators
type Schema map[string]Validator

// Validator is the interface all validators must implement
type Validator interface {
	Validate(ctx *ValidationContext, value any) map[string][]string
}

// ValidationContext holds validation state
type ValidationContext struct {
	Ctx      context.Context
	RootData DataObject
	Path     []string
	Options  *Options
}

// FullPath returns the dot-notation path string from the path slice
func (ctx *ValidationContext) FullPath() string {
	if len(ctx.Path) == 0 {
		return ""
	}
	result := ctx.Path[0]
	for i := 1; i < len(ctx.Path); i++ {
		result += "." + ctx.Path[i]
	}
	return result
}

// Options for validation
type Options struct {
	AbortEarly bool
	DBChecker  DBChecker
	Context    context.Context
}

// ValidationError holds all validation errors
type ValidationError struct {
	Errors map[string][]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Lookup function for accessing other fields
type Lookup func(path string) LookupResult

// LookupResult wraps a value lookup
type LookupResult struct {
	value  any
	exists bool
}

func (r LookupResult) Value() any   { return r.value }
func (r LookupResult) Exists() bool { return r.exists }

func (r LookupResult) String() string {
	if s, ok := r.value.(string); ok {
		return s
	}
	return ""
}

func (r LookupResult) Int() int64 {
	switch v := r.value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}

func (r LookupResult) Float() float64 {
	switch v := r.value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

func (r LookupResult) Bool() bool {
	if b, ok := r.value.(bool); ok {
		return b
	}
	return false
}

// Get returns a nested value by key (for objects)
func (r LookupResult) Get(key string) LookupResult {
	if r.value == nil {
		return LookupResult{nil, false}
	}
	if m, ok := r.value.(map[string]any); ok {
		if v, exists := m[key]; exists {
			return LookupResult{v, true}
		}
	}
	return LookupResult{nil, false}
}

// IsArray returns true if the value is a slice/array
func (r LookupResult) IsArray() bool {
	if r.value == nil {
		return false
	}
	_, ok := r.value.([]any)
	return ok
}

// IsObject returns true if the value is an object/map
func (r LookupResult) IsObject() bool {
	if r.value == nil {
		return false
	}
	_, ok := r.value.(map[string]any)
	return ok
}

// Array returns the value as []any
func (r LookupResult) Array() []any {
	if r.value == nil {
		return nil
	}
	if arr, ok := r.value.([]any); ok {
		return arr
	}
	return nil
}

// DBChecker interface for database validation
type DBChecker interface {
	CheckExists(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error)
}

// WhereClause for DB queries
type WhereClause struct {
	Column   string
	Operator string
	Value    any
}

// Helper functions for where clauses
func Where(column, operator string, value any) WhereClause {
	return WhereClause{Column: column, Operator: operator, Value: value}
}

func WhereEq(column string, value any) WhereClause {
	return WhereClause{Column: column, Operator: "=", Value: value}
}

func WhereNot(column string, value any) WhereClause {
	return WhereClause{Column: column, Operator: "!=", Value: value}
}

// PathKey represents the current path in validation
type PathKey struct {
	Previous []string
	Current  string
}

// RequiredIfCondition for struct-based conditional requirement
type RequiredIfCondition struct {
	FieldPath string
	Value     any
}

// RequiredUnlessCondition for struct-based conditional requirement
type RequiredUnlessCondition struct {
	FieldPath string
	Value     any
}

// SchemaObject is an alias for nested schema definitions
type SchemaObject = map[string]any

// SchemaSliceObject is an alias for array of schema objects
type SchemaSliceObject = []SchemaObject

// ============================================================================
// MESSAGE CONTEXT AND TEMPLATING
// ============================================================================

// MessageContext provides contextual information for dynamic error messages
type MessageContext struct {
	Field string       // Field name (e.g., "email")
	Path  string       // Full path (e.g., "users.0.email")
	Index int          // Array index if inside array (-1 otherwise)
	Value any          // The actual value being validated
	Rule  string       // The validation rule that failed (e.g., "required", "min")
	Param any          // Rule parameter if applicable (e.g., 3 for Min(3))
	Data  DataAccessor // The root data object being validated (with Get method)
}

// MessageFunc is a function that generates a custom error message
type MessageFunc func(ctx MessageContext) string

// MessageArg can be either a string or a MessageFunc
// This allows flexible error message customization
type MessageArg interface{}

// resolveMessage resolves a MessageArg to a string
func resolveMessage(arg MessageArg, ctx MessageContext) string {
	switch m := arg.(type) {
	case string:
		return m
	case MessageFunc:
		return m(ctx)
	case func(MessageContext) string:
		return m(ctx)
	default:
		return ""
	}
}

// extractIndex extracts array index from path (returns -1 if not in array)
func extractIndex(path string) int {
	// Look for patterns like "field.0" or "field.0.subfield"
	parts := strings.Split(path, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		// Check if this part is a number
		idx := 0
		isNum := true
		for _, c := range parts[i] {
			if c < '0' || c > '9' {
				isNum = false
				break
			}
			idx = idx*10 + int(c-'0')
		}
		if isNum && len(parts[i]) > 0 {
			return idx
		}
	}
	return -1
}
