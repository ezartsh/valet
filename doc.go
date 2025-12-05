// Package valet provides a high-performance, Zod-inspired validation library
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
//   - Custom Error Messages - Inline messages with dynamic template functions
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
//	    "github.com/ezartsh/valet"
//	)
//
//	func main() {
//	    jsonData := []byte(`{"name": "John", "email": "john@example.com", "age": 25}`)
//
//	    var data map[string]any
//	    json.Unmarshal(jsonData, &data)
//
//	    schema := valet.Schema{
//	        "name":  valet.String().Required().Min(2).Max(100),
//	        "email": valet.String().Required().Email(),
//	        "age":   valet.Float().Required().Min(18).Max(120),
//	    }
//
//	    if err := valet.Validate(data, schema); err != nil {
//	        fmt.Println("Validation failed:", err.Errors)
//	    }
//	}
//
// # String Validator
//
// The String validator provides comprehensive string validation:
//
//	valet.String().
//	    Required().              // Field must be present and non-empty
//	    Min(3).                  // Minimum length
//	    Max(100).                // Maximum length
//	    Length(10).              // Exact length
//	    Email().                 // Valid email format
//	    URL().                   // Valid URL format
//	    UUID().                  // Valid UUID format
//	    ULID().                  // Valid ULID format
//	    IP().                    // Valid IP address (IPv4 or IPv6)
//	    IPv4().                  // Valid IPv4 address
//	    IPv6().                  // Valid IPv6 address
//	    MAC().                   // Valid MAC address
//	    JSON().                  // Valid JSON string
//	    Base64().                // Valid Base64 encoding
//	    HexColor().              // Valid hex color code
//	    Alpha().                 // Letters only
//	    AlphaNumeric().          // Letters and numbers only
//	    AlphaDash().             // Letters, numbers, dashes, underscores
//	    ASCII().                 // ASCII characters only
//	    Digits(4).               // Digits only with exact length
//	    Regex(`^[A-Z]+$`).       // Must match pattern
//	    NotRegex(`<script>`).    // Must NOT match pattern
//	    In("a", "b", "c").       // Must be one of values
//	    NotIn("x", "y").         // Must NOT be one of values
//	    StartsWith("pre_").      // Must start with prefix
//	    EndsWith(".pdf").        // Must end with suffix
//	    DoesntStartWith("_").    // Must NOT start with prefixes
//	    DoesntEndWith(".tmp").   // Must NOT end with suffixes
//	    Contains("@").           // Must contain substring
//	    Includes("http", "://"). // Must contain all substrings
//	    SameAs("password").      // Must equal another field
//	    DifferentFrom("old").    // Must differ from another field
//	    Trim().                  // Trim whitespace before validation
//	    Lowercase().             // Convert to lowercase
//	    Uppercase().             // Convert to uppercase
//	    Transform(fn).           // Custom transformation
//	    Exists("table", "col").  // Must exist in database
//	    Unique("table", "col", nil). // Must be unique in database
//	    Custom(fn).              // Custom validation function
//	    Nullable()               // Allow null values
//
// # Number Validators
//
// Number validators support Go generics for type-safe validation:
//
//	// Shorthand constructors
//	valet.Float()   // Num[float64]() - for JSON numbers
//	valet.Int()     // Num[int]()
//	valet.Int64()   // Num[int64]()
//
//	valet.Float().
//	    Required().              // Field must be present
//	    Min(0).                  // Minimum value
//	    Max(100).                // Maximum value
//	    Between(0, 100).         // Value must be between min and max
//	    Positive().              // Must be > 0
//	    Negative().              // Must be < 0
//	    Integer().               // Must be whole number
//	    MultipleOf(5).           // Must be multiple of value
//	    Step(0.5).               // Alias for MultipleOf
//	    MinDigits(4).            // Minimum number of digits
//	    MaxDigits(6).            // Maximum number of digits
//	    LessThan("max").         // Must be < another field
//	    GreaterThan("min").      // Must be > another field
//	    LessThanOrEqual("max").  // Must be <= another field
//	    GreaterThanOrEqual("min"). // Must be >= another field
//	    In(1, 2, 3).             // Must be one of values
//	    NotIn(0, -1).            // Must NOT be one of values
//	    Exists("table", "id").   // Must exist in database
//	    Unique("table", "num", nil). // Must be unique
//	    Coerce().                // Coerce string to number
//	    Custom(fn).              // Custom validation
//	    Nullable()               // Allow null values
//
// # Boolean Validator
//
// The Bool validator handles boolean values:
//
//	valet.Bool().
//	    Required().              // Field must be present
//	    True().                  // Must be true
//	    False().                 // Must be false
//	    Coerce().                // Coerce "true"/"false" strings
//	    Default(false).          // Default value if nil
//	    Nullable()               // Allow null values
//
// # Array Validator
//
// The Array validator handles slices with element validation:
//
//	valet.Array().
//	    Required().              // Field must be present
//	    Min(1).                  // Minimum length
//	    Max(10).                 // Maximum length
//	    Length(5).               // Exact length
//	    Nonempty().              // Must have at least one element
//	    Of(valet.String()).      // Validate each element
//	    Unique().                // All elements must be unique
//	    Distinct().              // Alias for Unique
//	    Contains("a", "b").      // Must contain values
//	    DoesntContain("x").      // Must NOT contain values
//	    Exists("table", "id").   // All elements must exist in DB
//	    Concurrent(4).           // Concurrent element validation
//	    Custom(fn).              // Custom validation
//	    Nullable()               // Allow null values
//
// # Object Validator
//
// The Object validator handles nested structures:
//
//	valet.Object().
//	    Required().              // Field must be present
//	    Shape(valet.Schema{      // Define nested schema
//	        "name": valet.String().Required(),
//	        "age":  valet.Float().Required(),
//	    }).
//	    Strict().                // Fail on unknown keys
//	    Passthrough().           // Allow unknown keys (default)
//	    Pick("name", "age").     // Only validate these fields
//	    Omit("password").        // Exclude these fields
//	    Partial().               // Make all fields optional
//	    Extend(schema).          // Add more fields
//	    Merge(other).            // Merge two validators
//	    Custom(fn).              // Custom validation
//	    Nullable()               // Allow null values
//
// # File Validator
//
// The File validator handles multipart file uploads:
//
//	valet.File().
//	    Required().              // File must be present
//	    Min(1024).               // Minimum size in bytes
//	    Max(5 * 1024 * 1024).    // Maximum size (5MB)
//	    Mimes("image/jpeg", "image/png"). // Allowed MIME types
//	    Extensions(".jpg", ".png"). // Allowed extensions
//	    Image().                 // Must be image file
//	    Dimensions(&valet.ImageDimensions{
//	        MinWidth: 100,
//	        MaxWidth: 2000,
//	        Ratio:    "16/9",
//	    }).
//	    Custom(fn).              // Custom validation
//	    Nullable()               // Allow null values
//
// # Schema Helpers
//
// Helper functions for common patterns:
//
//	valet.Enum("draft", "published")  // String enum
//	valet.EnumInt(1, 2, 3, 4, 5)      // Integer enum
//	valet.Literal("active")           // Exact value match
//	valet.Union(validator1, validator2) // Match any validator
//	valet.Optional(validator)         // Make validator optional
//
// # Custom Error Messages
//
// Valet supports inline custom error messages:
//
//	// String messages
//	valet.String().Required("Name is required")
//	valet.String().Min(3, "Name must be at least 3 characters")
//
//	// Dynamic messages with MessageFunc
//	valet.String().Required(func(ctx valet.MessageContext) string {
//	    return fmt.Sprintf("The %s field is required", ctx.Field)
//	})
//
// The MessageContext provides rich context for dynamic messages:
//
//	type MessageContext struct {
//	    Field string       // Field name (e.g., "email")
//	    Path  string       // Full path (e.g., "users.0.email")
//	    Index int          // Array index if inside array (-1 otherwise)
//	    Value any          // The actual value being validated
//	    Rule  string       // The validation rule that failed
//	    Param any          // Rule parameter (e.g., 3 for Min(3))
//	    Data  DataAccessor // Root data with Get() method
//	}
//
// Access other fields using Data.Get():
//
//	valet.Float().Positive(func(ctx valet.MessageContext) string {
//	    name := ctx.Data.Get(fmt.Sprintf("items.%d.name", ctx.Index)).String()
//	    return fmt.Sprintf("Price for '%s' must be positive", name)
//	})
//
// # Conditional Validation
//
// Validate fields based on conditions:
//
//	valet.String().RequiredIf(func(data valet.DataObject) bool {
//	    return data["type"] == "premium"
//	})
//
//	valet.String().RequiredUnless(func(data valet.DataObject) bool {
//	    return data["is_guest"] == true
//	})
//
// # Lookup Function
//
// Access other fields in custom validators:
//
//	valet.Float().Custom(func(v float64, lookup valet.Lookup) error {
//	    maxPrice := lookup("settings.max_price").Float()
//	    if v > maxPrice {
//	        return errors.New("price exceeds maximum")
//	    }
//	    return nil
//	})
//
// # Database Validation
//
// The package supports database validation with multiple adapters:
//
//	// SQL Adapter
//	checker := valet.NewSQLAdapter(db)
//
//	// GORM Adapter
//	checker := valet.NewGormAdapter(gormDB)
//
//	// SQLX Adapter
//	checker := valet.NewSQLXAdapter(sqlxDB)
//
//	// Bun Adapter
//	checker := valet.NewBunAdapter(bunDB)
//
//	// Function Adapter
//	checker := valet.FuncAdapter(func(ctx context.Context, table, column string,
//	    values []any, wheres []valet.WhereClause) (map[any]bool, error) {
//	    // Custom implementation
//	})
//
// Use with schema:
//
//	schema := valet.Schema{
//	    "email": valet.String().Required().Unique("users", "email", nil),
//	    "category_id": valet.Float().Required().Exists("categories", "id",
//	        valet.WhereEq("status", "active"),
//	    ),
//	}
//
//	err := valet.ValidateWithDB(ctx, data, schema, checker)
//
// # Where Clauses
//
// Add conditions to database checks:
//
//	valet.WhereEq("status", "active")    // status = 'active'
//	valet.WhereNot("deleted", true)      // deleted != true
//	valet.Where("stock", ">", 0)         // stock > 0
//
// # Performance
//
// The package is optimized for high performance:
//
//   - sync.Pool for object reuse
//   - Regex caching with thread-safe cache
//   - Pre-compiled regex for UUID, MAC, ULID, etc.
//   - Parallel DB queries for multi-table checks
//   - Batched queries to prevent N+1 problem
//   - Zero-allocation validators for simple checks
//   - Context cancellation support
package valet
