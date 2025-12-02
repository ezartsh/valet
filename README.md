# Valet

[![Go Reference](https://pkg.go.dev/badge/github.com/ezartsh/valet.svg)](https://pkg.go.dev/github.com/ezartsh/valet)
[![Go Report Card](https://goreportcard.com/badge/github.com/ezartsh/valet)](https://goreportcard.com/report/github.com/ezartsh/valet)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://img.shields.io/badge/coverage-86.7%25-brightgreen)](https://github.com/ezartsh/valet)

A high-performance, Zod-inspired validation library for Go with a fluent builder API. Works with dynamic map data (`map[string]any`) - perfect for validating JSON payloads, API requests, and configuration data without requiring pre-defined structs.

## Features

- **Fluent Builder API** - Zod-like chainable validation rules
- **Zero Dependencies** - Only uses Go standard library
- **High Performance** - Optimized with sync.Pool, regex caching, and parallel execution
- **Type-Safe Generics** - Numeric validators use Go generics
- **Nested Validation** - Validate deeply nested objects and arrays
- **Conditional Validation** - `RequiredIf` and `RequiredUnless` rules
- **Custom Validators** - Add your own validation logic
- **Custom Error Messages** - Override default messages per field
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
    valet "github.com/ezartsh/valet"
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
    
    // Define validation schema with fluent API
    schema := valet.Schema{
        "name":  valet.String().Required().Min(2).Max(100),
        "email": valet.String().Required().Email(),
        "age":   valet.Float().Required().Min(18).Max(120),
    }
    
    // Validate
    if err := valet.Validate(data, schema); err != nil {
        fmt.Println("Validation failed:", err.Errors)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

## API Reference

### Validation Functions

```go
// Basic validation
err := valet.Validate(data, schema)
err := valet.Validate(data, schema, valet.Options{AbortEarly: true})

// Zod-like aliases
err := valet.Parse(data, schema)
data, err := valet.SafeParse(data, schema)

// With database validation
err := valet.ValidateWithDB(ctx, data, schema, dbChecker)
data, err := valet.ValidateWithDBContext(ctx, data, schema, opts)
```

### String Validator

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

### File Validator

```go
valet.File().
    Required().
    MaxSize(5 * 1024 * 1024).  // 5MB max
    MinSize(1024).             // 1KB min
    MimeTypes("image/jpeg", "image/png", "application/pdf").
    Extensions(".jpg", ".png", ".pdf").
    Image().                   // Must be image (jpeg, png, gif, webp)
    Document().                // Must be document (pdf, doc, docx, etc.)
    Custom(func(f FileInfo, lookup Lookup) error {
        // Custom file validation
        return nil
    })
```

### Database Validation

#### Using FuncAdapter (Simple)

```go
checker := valet.FuncAdapter(func(ctx context.Context, table, column string, values []any, wheres []WhereClause) (map[any]bool, error) {
    // Query your database
    // Return map of value -> exists
    result := make(map[any]bool)
    // ... query logic ...
    return result, nil
})

err := valet.ValidateWithDB(ctx, data, schema, checker)
```

#### Using SQL Adapter

```go
import "database/sql"

db, _ := sql.Open("mysql", "user:password@/dbname")
checker := valet.NewSQLAdapter(db)

schema := valet.Schema{
    "email":      valet.String().Required().Email().Unique("users", "email", nil),
    "category_id": valet.Float().Required().Exists("categories", "id"),
}

err := valet.ValidateWithDB(ctx, data, schema, checker)
```

#### Using GORM Adapter

```go
// Implement GormQuerier interface for your GORM instance
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
db, _ := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")
checker := valet.NewSQLXAdapter(db)
```

#### Using Bun Adapter

```go
sqldb, _ := sql.Open("postgres", dsn)
db := bun.NewDB(sqldb, pgdialect.New())
checker := valet.NewBunAdapter(db)
```

#### Where Clauses

```go
// Check exists with conditions
valet.Float().Exists("products", "id",
    valet.WhereEq("status", "active"),
    valet.WhereNot("deleted", true),
    valet.Where("stock", ">", 0),
)

// Unique with ignore (for updates)
valet.String().Unique("users", "email", currentUserEmail)
```

### Lookup Function

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
    
    // Check array
    if lookup("items").IsArray() {
        items := lookup("items").Array()
        // ...
    }
    
    return nil
})
```

### Options

```go
valet.Validate(data, schema, valet.Options{
    AbortEarly: true,           // Stop on first error
    DBChecker:  checker,        // Database checker for Exists/Unique
    Context:    ctx,            // Context for cancellation
})
```

### Error Handling

```go
err := valet.Validate(data, schema)
if err != nil {
    // err.Errors is map[string][]string
    for field, messages := range err.Errors {
        fmt.Printf("%s: %v\n", field, messages)
    }
    
    // Check if has errors
    if err.HasErrors() {
        // ...
    }
}
```

## Performance

### Benchmark Results

Tested on Intel Core i7-1355U, Go 1.21+

#### Individual Validators

| Validator | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| String_Required | 13.1 | 0 | 0 |
| String_MinMax | 17.9 | 0 | 0 |
| String_Email | 347.6 | 0 | 0 |
| Boolean_Required | 2.9 | 0 | 0 |
| Object_Required | 2.3 | 0 | 0 |
| Array_Required | 3.3 | 0 | 0 |

#### DB Validation

| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| Schema Only | 744 | 360 | 11 |
| Schema + DB Exists | 11,100 | 1,644 | 30 |
| Schema + DB Unique | 1,065 | 452 | 9 |
| Batching 10 items | 6,752 | 3,401 | 70 |
| Batching 100 items | 84,991 | 81,502 | 720 |
| Multi-table (4 tables) | 20,309 | 2,883 | 51 |

#### Nested Array Objects

| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| 5 items with DB check | 31,904 | 7,585 | 114 |
| 10 items with nested tags | 76,925 | 38,023 | 329 |
| 50 items deeply nested | 306,682 | 317,845 | 1,793 |

### Performance Optimizations

- **sync.Pool** for object reuse (DB check slices, batch groups, string builders)
- **Regex caching** with thread-safe cache
- **Parallel DB queries** for multi-table checks
- **Batched queries** to prevent N+1 problem
- **Pre-allocation** of maps and slices
- **Zero-allocation** validators for simple checks
- **Context cancellation** support

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
            Regex(`[A-Z]`).  // At least one uppercase
            Regex(`[0-9]`),  // At least one digit
        "age": valet.Float().Required().Min(18).Max(120),
        "role": valet.String().Required().In("user", "admin", "moderator"),
    }
    
    if err := valet.ValidateWithDB(r.Context(), data, schema, dbChecker); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(err.Errors)
        return
    }
    
    // Create user...
}
```

### E-commerce Order Validation

```go
schema := valet.Schema{
    "customer_id": valet.Float().Required().Exists("users", "id"),
    "shipping_address": valet.Object().Required().Shape(valet.Schema{
        "street":  valet.String().Required().Min(5),
        "city":    valet.String().Required(),
        "zip":     valet.String().Required().Regex(`^\d{5}$`),
        "country": valet.String().Required().In("US", "CA", "UK"),
    }),
    "items": valet.Array().Required().Min(1).Of(valet.Object().Shape(valet.Schema{
        "product_id": valet.Float().Required().Exists("products", "id",
            valet.WhereEq("status", "active"),
        ),
        "quantity": valet.Float().Required().Positive().Max(100),
        "price":    valet.Float().Required().Positive(),
    })),
    "coupon_code": valet.String().Exists("coupons", "code",
        valet.WhereEq("active", true),
    ),
}
```

### Conditional Validation

```go
schema := valet.Schema{
    "payment_type": valet.String().Required().In("card", "bank", "crypto"),
    
    // Required only if payment_type is "card"
    "card_number": valet.String().
        RequiredIf(func(d valet.DataObject) bool {
            return d["payment_type"] == "card"
        }).
        Regex(`^\d{16}$`),
    
    // Required unless payment_type is "crypto"
    "billing_address": valet.Object().
        RequiredUnless(func(d valet.DataObject) bool {
            return d["payment_type"] == "crypto"
        }).
        Shape(valet.Schema{
            "street": valet.String().Required(),
            "city":   valet.String().Required(),
        }),
}
```

## Notes

### JSON Number Types

When parsing JSON in Go, all numbers become `float64`. Use `valet.Float()` for JSON number validation:

```go
// JSON: {"age": 25}
// In Go: data["age"] is float64(25)
schema := valet.Schema{
    "age": valet.Float().Required().Int(), // Validates as integer
}
```

### Thread Safety

All validators are thread-safe and can be reused across goroutines. The internal caches (regex, sync.Pool) are protected by mutexes.

### DB Adapter Selection

| Adapter | Use Case |
|---------|----------|
| `FuncAdapter` | Simple cases, testing, custom implementations |
| `SQLAdapter` | Standard `database/sql` |
| `SQLXAdapter` | Using `jmoiron/sqlx` |
| `GormAdapter` | Using GORM ORM |
| `BunAdapter` | Using Bun ORM |

### Error Messages

Default error messages can be overridden per field:

```go
valet.String().
    Required().
    Min(5).
    Email().
    Message("required", "Please enter your email").
    Message("min", "Email is too short").
    Message("email", "Please enter a valid email address")
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Related Projects

- [go-validet](https://github.com/ezartsh/validet) - Struct-based validation library (original)
- [Zod](https://github.com/colinhacks/zod) - TypeScript-first schema validation (inspiration)
