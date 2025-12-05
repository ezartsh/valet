package valet

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrNilDBConnection = errors.New("database connection is nil")
)

// DataObject represents the data to validate (parsed JSON)
type DataObject = map[string]any

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
