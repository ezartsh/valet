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
