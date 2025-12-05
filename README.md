# Valet

[![Go Reference](https://pkg.go.dev/badge/github.com/ezartsh/valet.svg)](https://pkg.go.dev/github.com/ezartsh/valet)
[![Go Report Card](https://goreportcard.com/badge/github.com/ezartsh/valet)](https://goreportcard.com/report/github.com/ezartsh/valet)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://img.shields.io/badge/coverage-81.7%25-brightgreen)](https://github.com/ezartsh/valet)

A high-performance, Zod-inspired validation library for Go with a fluent builder API. Works with dynamic map data (`map[string]any`) - perfect for validating JSON payloads, API requests, and configuration data without requiring pre-defined structs.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Available Validation Rules](#available-validation-rules)
  - [String Rules](#string-rules)
  - [Number Rules](#number-rules)
  - [Boolean Rules](#boolean-rules)
  - [Array Rules](#array-rules)
  - [Object Rules](#object-rules)
  - [File Rules](#file-rules)
  - [Schema Helpers](#schema-helpers)
- [Custom Error Messages](#custom-error-messages)
- [Database Validation](#database-validation)
- [Performance](#performance)
- [Examples](#examples)
- [License](#license)

## Features

- **Fluent Builder API** - Zod-like chainable validation rules
- **Zero Dependencies** - Only uses Go standard library
- **High Performance** - Optimized with sync.Pool, regex caching, and parallel execution
- **Type-Safe Generics** - Numeric validators use Go generics
- **Nested Validation** - Validate deeply nested objects and arrays
- **Conditional Validation** - `RequiredIf` and `RequiredUnless` rules
- **Custom Validators** - Add your own validation logic
- **Custom Error Messages** - Inline messages with dynamic template functions
- **Integrated DB Validation** - Exists/Unique checks with batched queries (N+1 prevention)
- **Multiple DB Adapters** - Support for database/sql, GORM, sqlx, Bun, and custom implementations
- **Parallel DB Queries** - Multi-table checks execute concurrently
- **Context Support** - Full context.Context support for cancellation

## Requirements

- **Go 1.21+** (uses generics)
- **No external dependencies**

## Installation

```bash
go get github.com/ezartsh/valet
```

## Quick Start

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/ezartsh/valet"
)

func main() {
    // Parse JSON into map
    jsonData := []byte(`{
        "name": "John Doe",
        "email": "john@example.com",
        "age": 25
    }`)
    
    var data map[string]any
    json.Unmarshal(jsonData, &data)
    
    // Define validation schema
    schema := valet.Schema{
        "name":  valet.String().Required().Min(2).Max(100),
        "email": valet.String().Required().Email(),
        "age":   valet.Num[float64]().Required().Min(18).Max(120),
    }
    
    // Validate
    if err := valet.Validate(data, schema); err != nil {
        fmt.Println("Validation failed:", err.Errors)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

---

## Available Validation Rules

### String Rules

| Rule | Description |
|------|-------------|
| `Required()` | Field must be present and non-empty |
| `RequiredIf(fn)` | Required if condition is met |
| `RequiredUnless(fn)` | Required unless condition is met |
| `Min(n)` | Minimum string length |
| `Max(n)` | Maximum string length |
| `Length(n)` | Exact string length |
| `Email()` | Valid email format |
| `URL()` | Valid URL format |
| `UUID()` | Valid UUID format |
| `ULID()` | Valid ULID format |
| `IP()` | Valid IP address (IPv4 or IPv6) |
| `IPv4()` | Valid IPv4 address |
| `IPv6()` | Valid IPv6 address |
| `MAC()` | Valid MAC address |
| `JSON()` | Valid JSON string |
| `Base64()` | Valid Base64 encoded string |
| `HexColor()` | Valid hex color code |
| `Alpha()` | Contains only letters |
| `AlphaNumeric()` | Contains only letters and numbers |
| `AlphaDash()` | Letters, numbers, dashes, underscores |
| `ASCII()` | Contains only ASCII characters |
| `Digits(n)` | Contains only digits with exact length |
| `Regex(pattern)` | Must match regex pattern |
| `NotRegex(pattern)` | Must NOT match regex pattern |
| `In(values...)` | Must be one of specified values |
| `NotIn(values...)` | Must NOT be one of specified values |
| `StartsWith(prefix)` | Must start with prefix |
| `EndsWith(suffix)` | Must end with suffix |
| `DoesntStartWith(prefixes...)` | Must NOT start with any of prefixes |
| `DoesntEndWith(suffixes...)` | Must NOT end with any of suffixes |
| `Contains(substr)` | Must contain substring |
| `Includes(substrs...)` | Must contain all substrings |
| `SameAs(field)` | Must equal another field's value |
| `DifferentFrom(field)` | Must differ from another field's value |
| `Trim()` | Trim whitespace before validation |
| `Lowercase()` | Convert to lowercase before validation |
| `Uppercase()` | Convert to uppercase before validation |
| `Transform(fn)` | Apply custom transformation |
| `Exists(table, column)` | Value must exist in database |
| `Unique(table, column, ignore)` | Value must be unique in database |
| `Custom(fn)` | Custom validation function |
| `Nullable()` | Allow null values |
| `Default(value)` | Set default value if nil |

#### String Examples

```go
// Basic validations
valet.String().Required()
valet.String().Min(3).Max(100)
valet.String().Email()
valet.String().UUID()

// With custom messages
valet.String().Required("Name is required")
valet.String().Email("Please enter a valid email")

// Advanced validations
valet.String().Regex(`^[A-Z]{2}\d{4}$`)
valet.String().In("draft", "published", "archived")
valet.String().StartsWith("PRE_")

// Password confirmation
valet.String().SameAs("password")

// Transformations
valet.String().Trim().Lowercase().Email()

// Database checks
valet.String().Unique("users", "email", nil)
valet.String().Exists("categories", "slug")
```

### Number Rules

Number validators support Go generics. Use `Num[T]()` for any numeric type.

| Rule | Description |
|------|-------------|
| `Required()` | Field must be present |
| `RequiredIf(fn)` | Required if condition is met |
| `RequiredUnless(fn)` | Required unless condition is met |
| `Min(n)` | Minimum value |
| `Max(n)` | Maximum value |
| `Between(min, max)` | Value must be between min and max |
| `Positive()` | Must be greater than 0 |
| `Negative()` | Must be less than 0 |
| `Integer()` | Must be a whole number |
| `MultipleOf(n)` | Must be a multiple of value |
| `Step(n)` | Alias for MultipleOf |
| `MinDigits(n)` | Minimum number of digits |
| `MaxDigits(n)` | Maximum number of digits |
| `LessThan(field)` | Must be less than another field |
| `GreaterThan(field)` | Must be greater than another field |
| `LessThanOrEqual(field)` | Must be ≤ another field |
| `GreaterThanOrEqual(field)` | Must be ≥ another field |
| `In(values...)` | Must be one of specified values |
| `NotIn(values...)` | Must NOT be one of specified values |
| `Exists(table, column)` | Value must exist in database |
| `Unique(table, column, ignore)` | Value must be unique in database |
| `Coerce()` | Coerce string to number |
| `Custom(fn)` | Custom validation function |
| `Nullable()` | Allow null values |
| `Default(value)` | Set default value if nil |

#### Number Examples

```go
// Shorthand constructors
valet.Float()   // Num[float64]()
valet.Int()     // Num[int]()
valet.Int64()   // Num[int64]()

// Basic validations
valet.Float().Required().Min(0).Max(100)
valet.Int().Positive()
valet.Float().Between(0, 1)

// With custom messages
valet.Int().Min(18, "Must be at least 18 years old")

// Field comparisons
valet.Float().LessThan("max_price")
valet.Float().GreaterThan("min_price")

// Database checks
valet.Int().Exists("products", "id")
```

### Boolean Rules

| Rule | Description |
|------|-------------|
| `Required()` | Field must be present |
| `RequiredIf(fn)` | Required if condition is met |
| `RequiredUnless(fn)` | Required unless condition is met |
| `True()` | Must be true |
| `False()` | Must be false |
| `Coerce()` | Coerce string to boolean |
| `Custom(fn)` | Custom validation function |
| `Nullable()` | Allow null values |
| `Default(value)` | Set default value if nil |

#### Boolean Examples

```go
// Terms acceptance
valet.Bool().Required().True("You must accept the terms")

// Coerce from string ("true", "1", "yes", "on" -> true)
valet.Bool().Coerce().Required()

// Default value
valet.Bool().Default(false)
```

### Array Rules

| Rule | Description |
|------|-------------|
| `Required()` | Field must be present and non-empty |
| `RequiredIf(fn)` | Required if condition is met |
| `RequiredUnless(fn)` | Required unless condition is met |
| `Min(n)` | Minimum array length |
| `Max(n)` | Maximum array length |
| `Length(n)` | Exact array length |
| `Nonempty()` | Array must have at least one element |
| `Of(validator)` | Validate each element with schema |
| `Unique()` | All elements must be unique |
| `Distinct()` | Alias for Unique |
| `Contains(values...)` | Must contain specified values |
| `DoesntContain(values...)` | Must NOT contain specified values |
| `Exists(table, column)` | All elements must exist in database |
| `Concurrent(workers)` | Enable concurrent element validation |
| `Custom(fn)` | Custom validation function |
| `Nullable()` | Allow null values |

#### Array Examples

```go
// Array of strings (emails)
valet.Array().Of(valet.String().Email())

// Array of numbers
valet.Array().Min(1).Of(valet.Float().Positive())

// Array of objects
valet.Array().Of(valet.Object().Shape(valet.Schema{
    "product_id": valet.Int().Required().Exists("products", "id"),
    "quantity":   valet.Int().Required().Positive(),
}))

// Unique values
valet.Array().Unique()

// Concurrent validation for large arrays
valet.Array().Concurrent(4).Of(valet.Object().Shape(schema))
```

### Object Rules

| Rule | Description |
|------|-------------|
| `Required()` | Field must be present |
| `RequiredIf(fn)` | Required if condition is met |
| `RequiredUnless(fn)` | Required unless condition is met |
| `Shape(schema)` | Define nested validation schema |
| `Strict()` | Fail on unknown keys |
| `Passthrough()` | Allow unknown keys (default) |
| `Pick(fields...)` | Create validator with only specified fields |
| `Omit(fields...)` | Create validator excluding specified fields |
| `Partial()` | Make all fields optional |
| `Extend(schema)` | Extend schema with additional fields |
| `Merge(validator)` | Merge two object validators |
| `Custom(fn)` | Custom validation function |
| `Nullable()` | Allow null values |

#### Object Examples

```go
// Nested object
valet.Object().Shape(valet.Schema{
    "street":  valet.String().Required(),
    "city":    valet.String().Required(),
    "zip":     valet.String().Required().Digits(5),
    "country": valet.String().Required().In("US", "CA", "UK"),
})

// Strict mode (fail on unknown fields)
valet.Object().Strict().Shape(schema)

// Pick only specific fields
userSchema.Pick("id", "name")

// Omit fields
userSchema.Omit("password")

// Make all fields optional (for PATCH requests)
userSchema.Partial()

// Extend schema
baseSchema.Extend(valet.Schema{
    "email": valet.String().Email(),
})
```

### File Rules

For validating file uploads (`*multipart.FileHeader`).

| Rule | Description |
|------|-------------|
| `Required()` | File must be present |
| `RequiredIf(fn)` | Required if condition is met |
| `RequiredUnless(fn)` | Required unless condition is met |
| `Min(bytes)` | Minimum file size |
| `Max(bytes)` | Maximum file size |
| `Mimes(types...)` | Allowed MIME types |
| `Extensions(exts...)` | Allowed file extensions |
| `Image()` | Must be an image file |
| `Dimensions(opts)` | Image dimension constraints |
| `Custom(fn)` | Custom validation function |
| `Nullable()` | Allow null values |

#### File Examples

```go
// Image upload
valet.File().Required().Image().Max(5 * 1024 * 1024)

// PDF documents
valet.File().Mimes("application/pdf").Max(10 * 1024 * 1024)

// Image with dimensions
valet.File().Image().Dimensions(&valet.ImageDimensions{
    MinWidth:  200,
    MaxWidth:  2000,
    MinHeight: 200,
    MaxHeight: 2000,
})

// Avatar with aspect ratio
valet.File().Image().Dimensions(&valet.ImageDimensions{
    MinWidth: 100,
    Ratio:    "1/1",  // Square
})
```

### Schema Helpers

```go
valet.String().
    Required().                              // Field must be present and non-empty
    RequiredIf(func(d DataObject) bool {     // Required if condition is met
        return d["type"] == "premium"
    }).
    RequiredUnless(func(d DataObject) bool { // Required unless condition is met
        return d["role"] == "guest"
    }).
    Min(5).                                  // Minimum length (UTF-8 aware)
    Max(100).                                // Maximum length
    Length(10).                              // Exact length
    Regex(`^[a-zA-Z]+$`).                    // Must match regex pattern
    Email().                                 // Must be valid email format
    URL().                                   // Must be valid URL
    Alpha().                                 // Must contain only letters
    AlphaNumeric().                          // Must contain only letters and numbers
    Numeric().                               // Must contain only digits
    In("a", "b", "c").                       // Must be one of these values
    NotIn("x", "y", "z").                    // Must NOT be one of these values
    StartsWith("prefix").                    // Must start with string
    EndsWith("suffix").                      // Must end with string
    Contains("substring").                   // Must contain string
    UUID().                                  // Must be valid UUID
    Trim().                                  // Trim whitespace before validation
    ToLower().                               // Convert to lowercase before validation
    ToUpper().                               // Convert to uppercase before validation
    Exists("table", "column").               // DB existence check
    Exists("table", "column", where...).     // DB existence check with conditions
    Unique("table", "column", nil).          // DB uniqueness check
    Unique("table", "column", ignoreValue).  // DB uniqueness check (ignore value for updates)
    Custom(func(v string, lookup Lookup) error { // Custom validation
        if v == "forbidden" {
            return errors.New("this value is not allowed")
        }
        return nil
    }).
    Message("required", "Name is mandatory").  // Custom error message
    Nullable()                                 // Allow null values
```

### Number Validators

```go
// Float64 (default for JSON numbers)
valet.Float().
    Required().
    Min(0).                    // Minimum value
    Max(1000).                 // Maximum value
    Positive().                // Must be > 0
    Negative().                // Must be < 0
    NonNegative().             // Must be >= 0
    NonPositive().             // Must be <= 0
    MultipleOf(5).             // Must be multiple of value
    Int().                     // Must be integer (no decimals)
    In(1.0, 2.0, 3.0).         // Must be one of these values
    NotIn(0, -1).              // Must NOT be one of these values
    Exists("products", "id").  // DB existence check
    Unique("orders", "num", nil). // DB uniqueness check
    Custom(func(v float64, lookup Lookup) error {
        maxPrice := lookup("max_price").Float()
        if v > maxPrice {
            return errors.New("price exceeds maximum")
        }
        return nil
    }).
    Coerce().                  // Coerce string to number
    Nullable()

// Integer types
valet.Int().Required().Min(0).Max(100)
valet.Int64().Required().Min(0)
```

### Boolean Validator

```go
valet.Bool().
    Required().
    RequiredIf(func(d DataObject) bool {
        return d["show_terms"] == true
    }).
    True().                    // Must be true
    False().                   // Must be false
    Custom(func(v bool, lookup Lookup) error {
        termsShown := lookup("terms_shown").Bool()
        if termsShown && !v {
            return errors.New("you must agree to the terms")
        }
        return nil
    }).
    Coerce().                  // Coerce string "true"/"false" to bool
    Nullable()
```

### Object Validator

```go
valet.Object().
    Required().
    Shape(valet.Schema{      // Define nested schema
        "street":  valet.String().Required(),
        "city":    valet.String().Required(),
        "zip":     valet.String().Required().Length(5),
        "country": valet.String().Required(),
    }).
    Strict().                  // Fail on unknown keys
    Passthrough().             // Allow unknown keys (default)
    Custom(func(v DataObject, lookup Lookup) error {
        // Custom object validation
        return nil
    }).
    Nullable()
```

### Array Validator

```go
valet.Array().
    Required().
    Min(1).                    // Minimum length
    Max(10).                   // Maximum length
    Length(5).                 // Exact length
    Unique().                  // All elements must be unique
    Of(valet.String().Email()). // Validate each element
    Of(valet.Object().Shape(valet.Schema{  // Array of objects
        "product_id": valet.Float().Required().Exists("products", "id"),
        "quantity":   valet.Float().Required().Positive(),
    })).
    Exists("tags", "id").      // Check all elements exist in DB
    Custom(func(v []any, lookup Lookup) error {
        // Custom array validation
        return nil
    }).
    Nullable()
```

### Schema Helpers

Helper functions for common validation patterns.

| Helper | Description |
|--------|-------------|
| `Enum(values...)` | Value must be one of predefined options |
| `EnumInt(values...)` | Integer must be one of predefined options |
| `Literal(value)` | Value must be exactly the specified value |
| `Union(validators...)` | Value must match one of multiple validators |
| `Optional(validator)` | Make any validator optional |

#### Schema Helper Examples

```go
// String enum
valet.Enum("draft", "published", "archived")

// Integer enum
valet.EnumInt(1, 2, 3, 4, 5)

// Literal value
valet.Literal("active")

// Union (accept multiple types)
valet.Union(
    valet.String().Email(),
    valet.String().Regex(`^\+\d{10,15}$`),  // Phone
)

// Optional
valet.Optional(valet.String().Email())
```

---

## Custom Error Messages

Valet supports flexible custom error messages with two approaches:

### Inline String Messages

Pass a string message directly to any validation method:

```go
valet.String().
    Required("Name is required").
    Min(3, "Name must be at least 3 characters").
    Max(50, "Name cannot exceed 50 characters")
```

### Dynamic Message Functions

Use `MessageFunc` for dynamic messages with access to validation context:

```go
valet.String().Required(func(ctx valet.MessageContext) string {
    return fmt.Sprintf("The %s field is required", ctx.Field)
})
```

### MessageContext Fields

The `MessageContext` provides rich context for dynamic messages:

```go
type MessageContext struct {
    Field string       // Field name (e.g., "email")
    Path  string       // Full path (e.g., "users.0.email")
    Index int          // Array index if inside array (-1 otherwise)
    Value any          // The actual value being validated
    Rule  string       // The validation rule that failed
    Param any          // Rule parameter (e.g., 3 for Min(3))
    Data  DataAccessor // Root data with Get() method
}
```

### Using Data.Get()

Access other fields from the root data in your message:

```go
valet.String().Required(func(ctx valet.MessageContext) string {
    userName := ctx.Data.Get("user.name").String()
    return fmt.Sprintf("Email is required for user %s", userName)
})

// Access array items
valet.Float().Positive(func(ctx valet.MessageContext) string {
    itemName := ctx.Data.Get(fmt.Sprintf("items.%d.name", ctx.Index)).String()
    return fmt.Sprintf("Price for '%s' must be positive", itemName)
})
```

### Using Message() Method

Set messages for specific rules using the `Message()` method:

```go
valet.String().
    Required().
    Email().
    Message("required", "Please enter your email").
    Message("email", "Please enter a valid email address")
```

---

## Database Validation

### Setting Up a DB Checker

#### Using FuncAdapter (Simple)

```go
checker := valet.FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []valet.WhereClause) (map[any]bool, error) {
    result := make(map[any]bool)
    // Query your database and populate result
    // result[value] = true if value exists
    return result, nil
})

err := valet.ValidateWithDB(ctx, data, schema, checker)
```

#### Using SQL Adapter

```go
import "database/sql"

db, _ := sql.Open("mysql", "user:password@/dbname")
checker := valet.NewSQLAdapter(db)
```

#### Using GORM Adapter

```go
type GormDB struct {
    db *gorm.DB
}

func (g *GormDB) Raw(ctx context.Context, sql string, values ...interface{}) valet.GormResult {
    return g.db.WithContext(ctx).Raw(sql, values...)
}

checker := valet.NewGormAdapter(&GormDB{db: gormDB})
```

#### Using SQLX Adapter

```go
db, _ := sqlx.Connect("postgres", dsn)
checker := valet.NewSQLXAdapter(db)
```

#### Using Bun Adapter

```go
sqldb, _ := sql.Open("postgres", dsn)
db := bun.NewDB(sqldb, pgdialect.New())
checker := valet.NewBunAdapter(db)
```

### Where Clauses

Add conditions to database checks:

```go
// Equal
valet.WhereEq("status", "active")

// Not equal
valet.WhereNot("deleted", true)

// Custom operator
valet.Where("stock", ">", 0)

// Example usage
valet.Int().Exists("products", "id",
    valet.WhereEq("status", "active"),
    valet.Where("stock", ">", 0),
)
```

### Batched Queries

Valet automatically batches database queries to prevent N+1 problems:

```go
// This validates 100 items but only makes 1 query per table
schema := valet.Schema{
    "items": valet.Array().Of(valet.Object().Shape(valet.Schema{
        "product_id": valet.Int().Exists("products", "id"),
        "category":   valet.String().Exists("categories", "slug"),
    })),
}
```

---

## Lookup Function

Access other fields during validation:

```go
valet.Float().Custom(func(v float64, lookup valet.Lookup) error {
    // Get another field's value
    maxPrice := lookup("settings.max_price").Float()
    minQty := lookup("min_quantity").Int()
    isActive := lookup("is_active").Bool()
    name := lookup("name").String()
    
    // Check if field exists
    if lookup("optional_field").Exists() {
        // ...
    }
    
    // Get nested value
    city := lookup("address").Get("city").String()
    
    return nil
})
```

---

## Performance

### Benchmark Results

Tested on Intel Core i7-1355U, Go 1.21+

| Validator | ns/op | allocs/op |
|-----------|-------|-----------|
| String_Required | ~130 | 3 |
| String_Email | ~535 | 3 |
| String_UUID | ~783 | 7 |
| Number_Required | ~262 | 5 |
| Boolean_Required | ~85 | 2 |
| Object_Shape | ~1,498 | 14 |
| Array_Of_Object | ~3,425 | 46 |

### Performance Optimizations

- **sync.Pool** for object reuse
- **Regex caching** with thread-safe cache
- **Parallel DB queries** for multi-table checks
- **Batched queries** to prevent N+1 problem
- **Pre-compiled regex** for common patterns (UUID, MAC, ULID, etc.)
- **Zero-allocation** validators for simple checks

---

## Examples

### API Request Validation

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]any
    json.NewDecoder(r.Body).Decode(&data)
    
    schema := valet.Schema{
        "username": valet.String().Required().Min(3).Max(50).
            AlphaNumeric().
            Unique("users", "username", nil),
        "email": valet.String().Required().Email().
            Unique("users", "email", nil),
        "password": valet.String().Required().Min(8).
            Regex(`[A-Z]`, "Must contain uppercase").
            Regex(`[0-9]`, "Must contain number"),
        "age": valet.Float().Required().Min(18).Max(120),
    }
    
    if err := valet.ValidateWithDB(r.Context(), data, schema, dbChecker); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(err.Errors)
        return
    }
    
    // Create user...
}
```

### E-commerce Order

```go
schema := valet.Schema{
    "customer_id": valet.Int().Required().Exists("users", "id"),
    "shipping_address": valet.Object().Required().Shape(valet.Schema{
        "street":  valet.String().Required().Min(5),
        "city":    valet.String().Required(),
        "zip":     valet.String().Required().Digits(5),
        "country": valet.String().Required().In("US", "CA", "UK"),
    }),
    "items": valet.Array().Required().Min(1).Of(valet.Object().Shape(valet.Schema{
        "product_id": valet.Int().Required().Exists("products", "id",
            valet.WhereEq("status", "active"),
        ),
        "quantity": valet.Int().Required().Positive().Max(100),
    })),
    "coupon_code": valet.String().Nullable().Exists("coupons", "code",
        valet.WhereEq("active", true),
    ),
}
```

### Password Confirmation

```go
schema := valet.Schema{
    "password": valet.String().Required().Min(8),
    "password_confirmation": valet.String().Required().SameAs("password"),
}
```

### Conditional Validation

```go
schema := valet.Schema{
    "payment_type": valet.String().Required().In("card", "bank", "crypto"),
    
    "card_number": valet.String().
        RequiredIf(func(d valet.DataObject) bool {
            return d["payment_type"] == "card"
        }).
        Digits(16),
    
    "bank_account": valet.String().
        RequiredIf(func(d valet.DataObject) bool {
            return d["payment_type"] == "bank"
        }),
    
    "wallet_address": valet.String().
        RequiredIf(func(d valet.DataObject) bool {
            return d["payment_type"] == "crypto"
        }),
}
```

### Dynamic Error Messages with Array Context

```go
schema := valet.Schema{
    "items": valet.Array().Of(valet.Object().Shape(valet.Schema{
        "name": valet.String().Required(),
        "price": valet.Float().Required().Positive(func(ctx valet.MessageContext) string {
            name := ctx.Data.Get(fmt.Sprintf("items.%d.name", ctx.Index)).String()
            return fmt.Sprintf("Price for '%s' must be positive", name)
        }),
    })),
}
```

---

## Notes

### JSON Number Types

When parsing JSON in Go, all numbers become `float64`. Use `valet.Float()` for JSON number validation:

```go
// JSON: {"age": 25}
// In Go: data["age"] is float64(25)
schema := valet.Schema{
    "age": valet.Float().Required().Integer(), // Validates as integer
}
```

### Thread Safety

All validators are thread-safe and can be reused across goroutines. The internal caches (regex, sync.Pool) are protected by mutexes.

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
