package valet

import (
	"strings"
	"sync"
)

// Error map pool to reduce allocations
var errorMapPool = sync.Pool{
	New: func() any {
		return make(map[string][]string, 8)
	},
}

// GetErrorMap retrieves an error map from the pool
func GetErrorMap() map[string][]string {
	return errorMapPool.Get().(map[string][]string)
}

// PutErrorMap returns an error map to the pool after clearing it
func PutErrorMap(m map[string][]string) {
	// Clear the map before returning to pool
	for k := range m {
		delete(m, k)
	}
	errorMapPool.Put(m)
}

// String builder pool for efficient string construction
var builderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

// GetBuilder retrieves a string builder from the pool
func GetBuilder() *strings.Builder {
	b := builderPool.Get().(*strings.Builder)
	b.Reset()
	return b
}

// PutBuilder returns a string builder to the pool
func PutBuilder(b *strings.Builder) {
	builderPool.Put(b)
}

// String slice pool for path building
var stringSlicePool = sync.Pool{
	New: func() any {
		s := make([]string, 0, 8)
		return &s
	},
}

// GetStringSlice retrieves a string slice from the pool
func GetStringSlice() []string {
	return (*stringSlicePool.Get().(*[]string))[:0]
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(s []string) {
	s = s[:0]
	stringSlicePool.Put(&s)
}

// BuildPath efficiently builds a dot-notation path using pooled builder
func BuildPath(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	b := GetBuilder()
	b.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		b.WriteByte('.')
		b.WriteString(parts[i])
	}
	result := b.String()
	PutBuilder(b)
	return result
}

// JoinErrors efficiently joins error messages using pooled builder
func JoinErrors(errs []string, sep string) string {
	if len(errs) == 0 {
		return ""
	}
	if len(errs) == 1 {
		return errs[0]
	}

	b := GetBuilder()
	b.WriteString(errs[0])
	for i := 1; i < len(errs); i++ {
		b.WriteString(sep)
		b.WriteString(errs[i])
	}
	result := b.String()
	PutBuilder(b)
	return result
}
