// Package validet provides a high-performance, Zod-inspired validation library
// for Go with a fluent builder API.
//
// It works with dynamic map data (map[string]any) - perfect for validating
// JSON payloads, API requests, and configuration data without requiring
// pre-defined structs.
//
// # Features
//
//   - Fluent Builder API - Zod-like chainable validation rules
//   - Zero Dependencies - Only uses Go standard library
//   - High Performance - Optimized with sync.Pool, regex caching, and parallel execution
//   - Type-Safe Generics - Numeric validators use Go generics
//   - Nested Validation - Validate deeply nested objects and arrays
//   - Conditional Validation - RequiredIf and RequiredUnless rules
//   - Custom Validators - Add your own validation logic
//   - Integrated DB Validation - Exists/Unique checks with batched queries
//   - Multiple DB Adapters - Support for database/sql, GORM, sqlx, Bun
//
// # Quick Start
//
//	package main
//
//	import (
//	    "encoding/json"
//	    "fmt"
//	    validet "github.com/ezartsh/valet"
//	)
//
//	func main() {
//	    jsonData := []byte(`{"name": "John", "email": "john@example.com", "age": 25}`)
//
//	    var data map[string]any
//	    json.Unmarshal(jsonData, &data)
//
//	    schema := validet.Schema{
//	        "name":  validet.String().Required().Min(2).Max(100),
//	        "email": validet.String().Required().Email(),
//	        "age":   validet.Float().Required().Min(18).Max(120),
//	    }
//
//	    if err := validet.Validate(data, schema); err != nil {
//	        fmt.Println("Validation failed:", err.Errors)
//	    }
//	}
//
// # Validators
//
// The package provides the following validators:
//
//   - String() - String validation with email, URL, regex, etc.
//   - Float() - Float64 validation (default for JSON numbers)
//   - Int() - Integer validation
//   - Int64() - Int64 validation
//   - Bool() - Boolean validation
//   - Object() - Nested object validation
//   - Array() - Array/slice validation
//   - File() - File upload validation
//
// # Database Validation
//
// The package supports database validation with batched queries:
//
//	checker := validet.NewSQLAdapter(db)
//
//	schema := validet.Schema{
//	    "email":      validet.String().Required().Unique("users", "email", nil),
//	    "category_id": validet.Float().Required().Exists("categories", "id"),
//	}
//
//	err := validet.ValidateWithDB(ctx, data, schema, checker)
//
// # Performance
//
// The package is optimized for high performance:
//
//   - sync.Pool for object reuse
//   - Regex caching with thread-safe cache
//   - Parallel DB queries for multi-table checks
//   - Batched queries to prevent N+1 problem
//   - Zero-allocation validators for simple checks
package valet
