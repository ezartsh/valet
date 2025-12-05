package valet

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// ============================================================================
// COMPLEX VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkValidation_ComplexNested(b *testing.B) {
	jsonData := []byte(`{
		"name":        "tono",
		"email":       "test@example.com",
		"description": "some description",
		"store":       100,
		"url":         "https://www.example.com",
		"information": {
			"age":         25,
			"description": "ada",
			"job": {
				"level": "senior",
				"description": "developer"
			}
		},
		"tags": [1, 2, 3],
		"items": [
			{"title": "item1", "age": 12},
			{"title": "item2", "age": 30}
		]
	}`)

	request := map[string]interface{}{}
	json.Unmarshal(jsonData, &request)

	schema := Schema{
		"name":        String().Required().Min(2).Max(100),
		"email":       String().Required().Email(),
		"description": String().Required().Max(500),
		"url":         String().Required().URLWithOptions(UrlOptions{Https: true}),
		"store":       Float().Required().Min(1),
		"information": Object().Required().Shape(Schema{
			"age":         Float().Required().Min(1),
			"description": String().Required().Max(100),
			"job": Object().Required().Shape(Schema{
				"level":       String().Required().Max(50),
				"description": String().Required().Max(100),
			}),
		}),
		"tags": Array().Required().Min(1),
		"items": Array().Required().Of(Object().Shape(Schema{
			"title": String().Required(),
			"age":   Float().Required(),
		})),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(request, schema)
	}
}

func BenchmarkValidation_SimpleFlat(b *testing.B) {
	request := DataObject{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   float64(25),
	}

	schema := Schema{
		"name":  String().Required().Min(2).Max(100),
		"email": String().Required().Email(),
		"age":   Float().Required().Min(18).Max(120),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(request, schema)
	}
}

func BenchmarkValidation_WithErrors(b *testing.B) {
	request := DataObject{
		"name":  "J",
		"email": "invalid-email",
		"age":   float64(15),
	}

	schema := Schema{
		"name":  String().Required().Min(2).Max(100),
		"email": String().Required().Email(),
		"age":   Float().Required().Min(18).Max(120),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(request, schema)
	}
}

// ============================================================================
// STRING VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkString_Required(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"field"},
	}
	validator := String().Required()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "test value")
	}
}

func BenchmarkString_MinMax(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"field"},
	}
	validator := String().Min(5).Max(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "test value with some length")
	}
}

func BenchmarkString_Email(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"email"},
	}
	validator := String().Required().Email()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "test@example.com")
	}
}

func BenchmarkString_Regex(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"email"},
	}
	validator := String().Required().Regex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "test@example.com")
	}
}

func BenchmarkString_Regex_Cached(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"email"},
	}
	validator := String().Required().Regex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	// Warm up cache
	validator.Validate(ctx, "test@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "test@example.com")
	}
}

func BenchmarkString_Alpha(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"field"},
	}
	validator := String().Alpha()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "abcdefghij")
	}
}

func BenchmarkString_AlphaNumeric(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"field"},
	}
	validator := String().AlphaNumeric()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "abc123def456")
	}
}

func BenchmarkString_Url(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"url"},
	}
	validator := String().URLWithOptions(UrlOptions{Https: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "https://www.example.com/path")
	}
}

func BenchmarkString_In(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"status"},
	}
	validator := String().In("active", "inactive", "pending", "deleted", "archived")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "pending")
	}
}

func BenchmarkString_Custom(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"field"},
	}
	validator := String().Custom(func(v string, look Lookup) error {
		if len(v) < 5 {
			return errors.New("too short")
		}
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "test value")
	}
}

func BenchmarkString_RequiredIf(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{"type": "premium"},
		Path:     []string{"field"},
	}
	validator := String().RequiredIf(func(data DataObject) bool {
		return data["type"] == "premium"
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "some value")
	}
}

func BenchmarkString_Trim(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"field"},
	}
	validator := String().Trim().Min(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "   test value   ")
	}
}

// ============================================================================
// NUMERIC VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkNumeric_Required(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"age"},
	}
	validator := Float().Required()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(42))
	}
}

func BenchmarkNumeric_MinMax(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"value"},
	}
	validator := Float().Min(1).Max(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(50))
	}
}

func BenchmarkNumeric_Float64(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"price"},
	}
	validator := Float().Required().Min(1).Max(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, 123.456)
	}
}

func BenchmarkNumeric_Digits(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"code"},
	}
	validator := Float().MinDigits(3).MaxDigits(6)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(12345))
	}
}

func BenchmarkNumeric_In(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"rating"},
	}
	validator := Float().In(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(5))
	}
}

func BenchmarkNumeric_Custom(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"number"},
	}
	validator := Float().Custom(func(v float64, look Lookup) error {
		if int(v)%2 != 0 {
			return errors.New("must be even")
		}
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(42))
	}
}

func BenchmarkNumeric_Positive(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"amount"},
	}
	validator := Float().Positive()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(100))
	}
}

func BenchmarkNumeric_Integer(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"count"},
	}
	validator := Float().Integer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(42))
	}
}

func BenchmarkNumeric_Coerce(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"age"},
	}
	validator := Float().Coerce().Min(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "42")
	}
}

// ============================================================================
// BOOLEAN VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkBoolean_Required(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"active"},
	}
	validator := Bool().Required()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, true)
	}
}

func BenchmarkBoolean_RequiredIf(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{"type": "premium"},
		Path:     []string{"verified"},
	}
	validator := Bool().RequiredIf(func(data DataObject) bool {
		return data["type"] == "premium"
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, true)
	}
}

func BenchmarkBoolean_Custom(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"agree"},
	}
	validator := Bool().Custom(func(v bool, look Lookup) error {
		if !v {
			return errors.New("must be true")
		}
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, true)
	}
}

func BenchmarkBoolean_Coerce(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"active"},
	}
	validator := Bool().Coerce()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "true")
	}
}

// ============================================================================
// OBJECT VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkObject_Required(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"profile"},
	}
	validator := Object().Required()
	value := map[string]any{"name": "John", "age": 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkObject_Shape(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"user"},
	}
	validator := Object().Required().Shape(Schema{
		"name":  String().Required(),
		"email": String().Required().Email(),
	})
	value := map[string]any{"name": "John", "email": "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkObject_Custom(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"profile"},
	}
	validator := Object().Custom(func(v DataObject, look Lookup) error {
		if _, ok := v["email"]; !ok {
			return errors.New("email required")
		}
		return nil
	})
	value := map[string]any{"name": "John", "email": "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkObject_Strict(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"config"},
	}
	validator := Object().Strict().Shape(Schema{
		"key": String().Required(),
	})
	value := map[string]any{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

// ============================================================================
// ARRAY VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkArray_Required(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"tags"},
	}
	validator := Array().Required()
	value := []any{"tag1", "tag2", "tag3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_MinMax(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"tags"},
	}
	validator := Array().Min(1).Max(10)
	value := []any{"tag1", "tag2", "tag3", "tag4", "tag5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Of_String(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"tags"},
	}
	validator := Array().Required().Of(String().Min(2))
	value := []any{"tag1", "tag2", "tag3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Of_Number(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"scores"},
	}
	validator := Array().Required().Of(Float().Min(0).Max(100))
	value := []any{float64(85), float64(90), float64(75)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Of_Object(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}
	validator := Array().Required().Of(Object().Shape(Schema{
		"name":  String().Required(),
		"price": Float().Required(),
	}))
	value := []any{
		map[string]any{"name": "Item 1", "price": float64(10)},
		map[string]any{"name": "Item 2", "price": float64(20)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Unique(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"ids"},
	}
	validator := Array().Required().Unique()
	value := []any{1, 2, 3, 4, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Custom(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"tags"},
	}
	validator := Array().Custom(func(v []any, look Lookup) error {
		for _, item := range v {
			if s, ok := item.(string); ok && len(s) < 2 {
				return errors.New("too short")
			}
		}
		return nil
	})
	value := []any{"tag1", "tag2", "tag3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

// ============================================================================
// LOOKUP PATH BENCHMARKS
// ============================================================================

func BenchmarkLookupPath_Simple(b *testing.B) {
	data := DataObject{"name": "John", "age": 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lookupPath(data, "name")
	}
}

func BenchmarkLookupPath_Nested(b *testing.B) {
	data := DataObject{
		"user": map[string]any{
			"profile": map[string]any{
				"name": "John",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lookupPath(data, "user.profile.name")
	}
}

func BenchmarkLookupPath_DeepNested(b *testing.B) {
	data := DataObject{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": "deep",
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lookupPath(data, "level1.level2.level3.level4.value")
	}
}

// ============================================================================
// REGEX CACHE BENCHMARKS
// ============================================================================

func BenchmarkRegexCache_Hit(b *testing.B) {
	// Warm up cache
	globalRegexCache.GetOrCompile(`^[a-zA-Z]+$`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalRegexCache.GetOrCompile(`^[a-zA-Z]+$`)
	}
}

func BenchmarkRegexCache_Miss(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use unique pattern each time to force cache miss
		globalRegexCache.GetOrCompile(`^test` + string(rune(i%26+'a')) + `$`)
	}
}

// ============================================================================
// FULL SCHEMA VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkFullSchema_Small(b *testing.B) {
	data := DataObject{
		"name":  "John",
		"email": "john@example.com",
	}
	schema := Schema{
		"name":  String().Required().Min(2),
		"email": String().Required().Email(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkFullSchema_Medium(b *testing.B) {
	data := DataObject{
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      float64(30),
		"active":   true,
		"role":     "admin",
		"tags":     []any{"go", "rust"},
		"metadata": map[string]any{"key": "value"},
	}
	schema := Schema{
		"name":     String().Required().Min(2).Max(100),
		"email":    String().Required().Email(),
		"age":      Float().Required().Min(18).Max(120),
		"active":   Bool().Required(),
		"role":     String().Required().In("admin", "user", "guest"),
		"tags":     Array().Required().Min(1),
		"metadata": Object().Required(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkFullSchema_Large(b *testing.B) {
	data := DataObject{
		"user": map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
			"profile": map[string]any{
				"bio":     "Developer",
				"website": "https://example.com",
			},
		},
		"items": []any{
			map[string]any{"name": "Item 1", "price": float64(10)},
			map[string]any{"name": "Item 2", "price": float64(20)},
			map[string]any{"name": "Item 3", "price": float64(30)},
		},
		"tags":   []any{"tag1", "tag2", "tag3"},
		"active": true,
	}
	schema := Schema{
		"user": Object().Required().Shape(Schema{
			"name":  String().Required().Min(2),
			"email": String().Required().Email(),
			"profile": Object().Shape(Schema{
				"bio":     String().Max(500),
				"website": String().URL(),
			}),
		}),
		"items": Array().Required().Min(1).Of(Object().Shape(Schema{
			"name":  String().Required(),
			"price": Float().Required().Positive(),
		})),
		"tags":   Array().Required().Of(String().Min(2)),
		"active": Bool().Required(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// SAFE PARSE BENCHMARKS
// ============================================================================

func BenchmarkSafeParse_Valid(b *testing.B) {
	data := DataObject{
		"name":  "John",
		"email": "john@example.com",
	}
	schema := Schema{
		"name":  String().Required(),
		"email": String().Required().Email(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SafeParse(data, schema)
	}
}

func BenchmarkSafeParse_Invalid(b *testing.B) {
	data := DataObject{
		"name":  "",
		"email": "invalid",
	}
	schema := Schema{
		"name":  String().Required(),
		"email": String().Required().Email(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SafeParse(data, schema)
	}
}

// ============================================================================
// TIME VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkTime_Required(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"created_at"},
	}
	validator := Time().Required()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "2024-01-15T10:30:00Z")
	}
}

func BenchmarkTime_AfterBefore(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"event_date"},
	}
	validator := Time().Required().AfterNow()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "2030-12-31T23:59:59Z")
	}
}

// ============================================================================
// TRANSFORM BENCHMARKS
// ============================================================================

func BenchmarkString_Transform(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"email"},
	}
	validator := String().
		Transform(strings.ToLower).
		Transform(strings.TrimSpace).
		Email()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "  JOHN@EXAMPLE.COM  ")
	}
}

// ============================================================================
// CROSS-FIELD VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkString_SameAs(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{
			"password": "secret123",
		},
		Path: []string{"password_confirmation"},
	}
	validator := String().Required().SameAs("password")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, "secret123")
	}
}

func BenchmarkNumeric_LessThan(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{
			"max_price": float64(1000),
		},
		Path: []string{"min_price"},
	}
	validator := Float().Required().LessThan("max_price")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, float64(500))
	}
}

// ============================================================================
// CONCURRENT ARRAY VALIDATION BENCHMARKS
// ============================================================================

func BenchmarkArray_Sequential_SmallArray(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}
	validator := Array().Required().Of(String().Required().Min(2))
	value := []any{"tag1", "tag2", "tag3", "tag4", "tag5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Concurrent_SmallArray(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}
	validator := Array().Required().Of(String().Required().Min(2)).Concurrent(4)
	value := []any{"tag1", "tag2", "tag3", "tag4", "tag5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Sequential_LargeArray(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}
	validator := Array().Required().Of(String().Required().Email())
	value := make([]any, 100)
	for i := range value {
		value[i] = "user@example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

func BenchmarkArray_Concurrent_LargeArray(b *testing.B) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}
	validator := Array().Required().Of(String().Required().Email()).Concurrent(8)
	value := make([]any, 100)
	for i := range value {
		value[i] = "user@example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, value)
	}
}

// ============================================================================
// PATH CACHE BENCHMARKS
// ============================================================================

func BenchmarkPathCache_Hit(b *testing.B) {
	// Warm up cache
	globalPathCache.getSplitPath("user.profile.settings.theme")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalPathCache.getSplitPath("user.profile.settings.theme")
	}
}

func BenchmarkPathCache_Miss(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use unique path each time to force cache miss
		globalPathCache.getSplitPath("path" + string(rune(i%26+'a')) + ".nested")
	}
}

// ============================================================================
// POOL BENCHMARKS
// ============================================================================

func BenchmarkErrorMapPool_GetPut(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := GetErrorMap()
		m["field"] = []string{"error1", "error2"}
		PutErrorMap(m)
	}
}

func BenchmarkErrorMap_NoPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := make(map[string][]string, 8)
		m["field"] = []string{"error1", "error2"}
		_ = m
	}
}

func BenchmarkBuilderPool_GetPut(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := GetBuilder()
		builder.WriteString("field")
		builder.WriteByte('.')
		builder.WriteString("subfield")
		_ = builder.String()
		PutBuilder(builder)
	}
}

func BenchmarkBuildPath(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildPath("user", "profile", "settings", "theme")
	}
}

// ============================================================================
// LARGE SCHEMA BENCHMARKS
// ============================================================================

func BenchmarkFullSchema_50Fields(b *testing.B) {
	data := DataObject{}
	schema := Schema{}

	// Generate 50 fields
	for i := 0; i < 50; i++ {
		fieldName := "field" + string(rune(i%26+'a')) + string(rune(i/26%26+'a'))
		data[fieldName] = "valid_value_here"
		schema[fieldName] = String().Required().Min(5).Max(100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkFullSchema_100Fields(b *testing.B) {
	data := DataObject{}
	schema := Schema{}

	// Generate 100 fields
	for i := 0; i < 100; i++ {
		fieldName := "field" + string(rune(i%26+'a')) + string(rune(i/26%26+'a')) + string(rune(i/676%26+'a'))
		data[fieldName] = "valid_value_here"
		schema[fieldName] = String().Required().Min(5).Max(100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// NEW STRING VALIDATOR BENCHMARKS
// ============================================================================

func BenchmarkString_UUID(b *testing.B) {
	schema := Schema{"id": String().Required().UUID()}
	data := DataObject{"id": "550e8400-e29b-41d4-a716-446655440000"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_IPv4(b *testing.B) {
	schema := Schema{"ip": String().Required().IPv4()}
	data := DataObject{"ip": "192.168.1.1"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_IPv6(b *testing.B) {
	schema := Schema{"ip": String().Required().IPv6()}
	data := DataObject{"ip": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_JSON(b *testing.B) {
	schema := Schema{"config": String().Required().JSON()}
	data := DataObject{"config": `{"name": "test", "value": 123, "nested": {"key": "value"}}`}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_HexColor(b *testing.B) {
	schema := Schema{"color": String().Required().HexColor()}
	data := DataObject{"color": "#ff5733"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_DoesntStartWith(b *testing.B) {
	schema := Schema{"name": String().Required().DoesntStartWith("admin", "root", "system")}
	data := DataObject{"name": "john_doe"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_DoesntEndWith(b *testing.B) {
	schema := Schema{"file": String().Required().DoesntEndWith(".exe", ".bat", ".sh")}
	data := DataObject{"file": "document.pdf"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_Includes(b *testing.B) {
	schema := Schema{"bio": String().Required().Includes("developer", "engineer")}
	data := DataObject{"bio": "I am a software developer and systems engineer with 10 years of experience"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_MAC(b *testing.B) {
	schema := Schema{"mac": String().Required().MAC()}
	data := DataObject{"mac": "00:1A:2B:3C:4D:5E"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_ULID(b *testing.B) {
	schema := Schema{"id": String().Required().ULID()}
	data := DataObject{"id": "01ARZ3NDEKTSV4RRFFQ69G5FAV"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_AlphaDash(b *testing.B) {
	schema := Schema{"slug": String().Required().AlphaDash()}
	data := DataObject{"slug": "hello-world_123"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_Digits(b *testing.B) {
	schema := Schema{"otp": String().Required().Digits(6)}
	data := DataObject{"otp": "123456"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_ASCII(b *testing.B) {
	schema := Schema{"text": String().Required().ASCII()}
	data := DataObject{"text": "Hello World 123!@#$%"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_Base64(b *testing.B) {
	schema := Schema{"data": String().Required().Base64()}
	data := DataObject{"data": "SGVsbG8gV29ybGQhIFRoaXMgaXMgYSB0ZXN0IHN0cmluZy4="}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkString_AllNewValidators(b *testing.B) {
	schema := Schema{
		"id":    String().Required().UUID(),
		"ip":    String().Required().IPv4(),
		"mac":   String().Required().MAC(),
		"slug":  String().Required().AlphaDash(),
		"otp":   String().Required().Digits(6),
		"color": String().Required().HexColor(),
	}
	data := DataObject{
		"id":    "550e8400-e29b-41d4-a716-446655440000",
		"ip":    "192.168.1.1",
		"mac":   "00:1A:2B:3C:4D:5E",
		"slug":  "hello-world_123",
		"otp":   "123456",
		"color": "#ff5733",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// NEW NUMBER VALIDATOR BENCHMARKS
// ============================================================================

func BenchmarkNumber_Between(b *testing.B) {
	schema := Schema{"score": Float().Required().Between(0, 100)}
	data := DataObject{"score": float64(75)}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkNumber_Step(b *testing.B) {
	schema := Schema{"quantity": Float().Required().Step(5)}
	data := DataObject{"quantity": float64(25)}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkNumber_MultipleOf(b *testing.B) {
	schema := Schema{"price": Float().Required().MultipleOf(0.25)}
	data := DataObject{"price": float64(10.75)}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// NEW ARRAY VALIDATOR BENCHMARKS
// ============================================================================

func BenchmarkArray_Contains(b *testing.B) {
	schema := Schema{"roles": Array().Required().Contains("admin", "user")}
	data := DataObject{"roles": []any{"admin", "user", "guest", "editor"}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkArray_DoesntContain(b *testing.B) {
	schema := Schema{"tags": Array().Required().DoesntContain("spam", "nsfw", "banned")}
	data := DataObject{"tags": []any{"tech", "news", "sports", "entertainment"}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkArray_Distinct(b *testing.B) {
	schema := Schema{"items": Array().Required().Distinct()}
	data := DataObject{"items": []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkArray_AllNewValidators(b *testing.B) {
	schema := Schema{
		"roles": Array().Required().Contains("admin").DoesntContain("banned").Distinct(),
	}
	data := DataObject{"roles": []any{"admin", "user", "editor"}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// OBJECT UTILITY BENCHMARKS
// ============================================================================

func BenchmarkObject_Pick(b *testing.B) {
	fullSchema := Object().Shape(Schema{
		"name":     String().Required(),
		"email":    String().Required().Email(),
		"password": String().Required().Min(8),
		"age":      Int().Required(),
		"address":  String(),
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fullSchema.Pick("name", "email")
	}
}

func BenchmarkObject_Omit(b *testing.B) {
	fullSchema := Object().Shape(Schema{
		"name":     String().Required(),
		"email":    String().Required().Email(),
		"password": String().Required().Min(8),
		"age":      Int().Required(),
		"address":  String(),
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fullSchema.Omit("password")
	}
}

func BenchmarkObject_Partial(b *testing.B) {
	fullSchema := Object().Shape(Schema{
		"name":     String().Required(),
		"email":    String().Required().Email(),
		"password": String().Required().Min(8),
		"age":      Int().Required(),
		"address":  String(),
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fullSchema.Partial()
	}
}

func BenchmarkObject_Merge(b *testing.B) {
	baseSchema := Object().Shape(Schema{
		"name": String().Required(),
		"age":  Int().Required(),
	})
	extendedSchema := Object().Shape(Schema{
		"email": String().Required().Email(),
		"phone": String(),
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		baseSchema.Merge(extendedSchema)
	}
}

func BenchmarkObject_PickedValidation(b *testing.B) {
	fullSchema := Object().Shape(Schema{
		"name":     String().Required(),
		"email":    String().Required().Email(),
		"password": String().Required().Min(8),
	})
	pickedSchema := Schema{"user": fullSchema.Pick("name", "email")}
	data := DataObject{
		"user": map[string]any{
			"name":  "John",
			"email": "john@example.com",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, pickedSchema)
	}
}

// ============================================================================
// SCHEMA HELPER BENCHMARKS
// ============================================================================

func BenchmarkEnum_String(b *testing.B) {
	schema := Schema{"status": Enum("pending", "active", "completed", "archived").Required()}
	data := DataObject{"status": "active"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkEnumInt(b *testing.B) {
	schema := Schema{"priority": EnumInt(1, 2, 3, 4, 5).Required()}
	data := DataObject{"priority": float64(3)}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkLiteral(b *testing.B) {
	schema := Schema{"type": Literal("config").Required()}
	data := DataObject{"type": "config"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkUnion_TwoTypes(b *testing.B) {
	schema := Schema{"id": Union(String().UUID(), Int().Positive()).Required()}
	data := DataObject{"id": "550e8400-e29b-41d4-a716-446655440000"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkUnion_ThreeLiterals(b *testing.B) {
	schema := Schema{
		"action": Union(
			Literal("create"),
			Literal("update"),
			Literal("delete"),
		).Required(),
	}
	data := DataObject{"action": "update"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkOptional_Present(b *testing.B) {
	schema := Schema{
		"name":     String().Required(),
		"nickname": Optional(String().Min(2).Max(20)),
	}
	data := DataObject{"name": "John", "nickname": "Johnny"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

func BenchmarkOptional_Missing(b *testing.B) {
	schema := Schema{
		"name":     String().Required(),
		"nickname": Optional(String().Min(2).Max(20)),
	}
	data := DataObject{"name": "John"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}

// ============================================================================
// COMPLEX NEW FEATURES BENCHMARK
// ============================================================================

func BenchmarkComplexNewFeatures(b *testing.B) {
	schema := Schema{
		"id":       Union(String().UUID(), Int().Positive()).Required(),
		"status":   Enum("draft", "published", "archived").Required(),
		"type":     Literal("article"),
		"priority": EnumInt(1, 2, 3, 4, 5),
		"author": Object().Required().Shape(Schema{
			"name":  String().Required().AlphaDash(),
			"email": String().Required().Email(),
		}),
		"tags":   Array().Required().Min(1).Distinct().DoesntContain("spam"),
		"score":  Float().Between(0, 100),
		"config": Optional(String().JSON()),
	}
	data := DataObject{
		"id":       "550e8400-e29b-41d4-a716-446655440000",
		"status":   "published",
		"type":     "article",
		"priority": float64(3),
		"author": map[string]any{
			"name":  "john_doe",
			"email": "john@example.com",
		},
		"tags":   []any{"tech", "news"},
		"score":  float64(85),
		"config": `{"key": "value"}`,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(data, schema)
	}
}
