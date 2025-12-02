# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-12-02

### Added

- **Fluent Builder API** - Zod-inspired chainable validation rules
- **String Validator** with support for:
  - Required, Min, Max, Length
  - Email, URL, UUID, Alpha, AlphaNumeric, Numeric
  - Regex pattern matching
  - In/NotIn value lists
  - StartsWith, EndsWith, Contains
  - Trim, ToLower, ToUpper transformations
  - DB Exists/Unique checks
  - Custom validation functions
  - Custom error messages
  - Nullable support

- **Number Validators** (Float, Int, Int64) with support for:
  - Required, Min, Max
  - Positive, Negative, NonNegative, NonPositive
  - MultipleOf, Int (integer check)
  - In/NotIn value lists
  - DB Exists/Unique checks
  - Coercion from string
  - Custom validation functions

- **Boolean Validator** with support for:
  - Required, True, False
  - Coercion from string
  - Custom validation functions

- **Object Validator** with support for:
  - Required, Shape (nested schema)
  - Strict mode (fail on unknown keys)
  - Passthrough mode (allow unknown keys)
  - Recursive DB check collection
  - Custom validation functions

- **Array Validator** with support for:
  - Required, Min, Max, Length
  - Unique elements
  - Element validation (Of)
  - DB Exists check for all elements
  - Recursive DB check collection from nested objects
  - Custom validation functions

- **File Validator** with support for:
  - Required, MaxSize, MinSize
  - MimeTypes, Extensions
  - Image, Document presets
  - Custom validation functions

- **Database Integration**:
  - Batched queries (N+1 prevention)
  - Parallel multi-table execution
  - Multiple adapters: SQL, SQLX, GORM, Bun, FuncAdapter
  - Where clause support
  - Unique with ignore (for updates)
  - Context cancellation support

- **Performance Optimizations**:
  - sync.Pool for object reuse
  - Regex caching
  - Pre-allocation of maps and slices
  - Zero-allocation validators
  - Parallel DB query execution

- **Conditional Validation**:
  - RequiredIf
  - RequiredUnless

- **Lookup Function** for accessing other fields during validation

- **Comprehensive Test Suite** with 86%+ coverage

### Performance

- String_Required: 13.1 ns/op, 0 allocs
- Boolean_Required: 2.9 ns/op, 0 allocs
- Object_Required: 2.3 ns/op, 0 allocs
- Array_Required: 3.3 ns/op, 0 allocs
- Schema validation: 744 ns/op
- DB Exists check: 11,100 ns/op
- Batching 100 items: 84,991 ns/op

[Unreleased]: https://github.com/ezartsh/valet/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/ezartsh/valet/releases/tag/v1.0.0
