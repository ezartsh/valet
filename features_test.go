package valet

import (
	"strings"
	"testing"
	"time"
)

// ============================================================================
// TIME VALIDATOR TESTS
// ============================================================================

func TestTimeValidator_Required(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid time string", "2024-01-15T10:30:00Z", false},
		{"nil value", nil, true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: DataObject{},
				Path:     []string{"created_at"},
			}
			validator := Time().Required()
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().Required().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_Format(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		value   string
		wantErr bool
	}{
		{"RFC3339 valid", time.RFC3339, "2024-01-15T10:30:00Z", false},
		{"RFC3339 invalid", time.RFC3339, "2024-01-15", true},
		{"date only format", "2006-01-02", "2024-01-15", false},
		{"date only format invalid", "2006-01-02", "2024-01-15T10:30:00Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: DataObject{},
				Path:     []string{"date"},
			}
			validator := Time().Required().Format(tt.format)
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().Format().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_After(t *testing.T) {
	referenceTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"after reference", "2024-01-16T00:00:00Z", false},
		{"same as reference", "2024-01-15T00:00:00Z", true},
		{"before reference", "2024-01-14T00:00:00Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: DataObject{},
				Path:     []string{"date"},
			}
			validator := Time().Required().After(referenceTime)
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().After().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_Before(t *testing.T) {
	referenceTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"before reference", "2024-01-14T00:00:00Z", false},
		{"same as reference", "2024-01-15T00:00:00Z", true},
		{"after reference", "2024-01-16T00:00:00Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: DataObject{},
				Path:     []string{"date"},
			}
			validator := Time().Required().Before(referenceTime)
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().Before().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_Between(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"within range", "2024-06-15T12:00:00Z", false},
		{"at start", "2024-01-01T00:00:00Z", false},
		{"at end", "2024-12-31T23:59:59Z", false},
		{"before start", "2023-12-31T23:59:59Z", true},
		{"after end", "2025-01-01T00:00:00Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: DataObject{},
				Path:     []string{"date"},
			}
			validator := Time().Required().Between(startTime, endTime)
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().Between().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_AfterField(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    string
		wantErr  bool
	}{
		{
			"after field value",
			DataObject{"start_date": "2024-01-01T00:00:00Z"},
			"2024-01-15T00:00:00Z",
			false,
		},
		{
			"before field value",
			DataObject{"start_date": "2024-01-15T00:00:00Z"},
			"2024-01-01T00:00:00Z",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"end_date"},
			}
			validator := Time().Required().AfterField("start_date")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().AfterField().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_Nullable(t *testing.T) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"optional_date"},
	}
	validator := Time().Nullable()
	errs := validator.Validate(ctx, nil)
	if len(errs) > 0 {
		t.Errorf("Time().Nullable().Validate() should not error on nil, got %v", errs)
	}
}

func TestTimeValidator_TimezoneWithFieldComparison(t *testing.T) {
	// Test that timezone is respected when comparing with other fields
	jakarta, _ := time.LoadLocation("Asia/Jakarta")

	tests := []struct {
		name     string
		rootData DataObject
		value    string
		timezone *time.Location
		format   string
		wantErr  bool
	}{
		{
			"timezone aware comparison - after field",
			DataObject{"start_date": "2024-01-15 10:00:00"},
			"2024-01-15 12:00:00",
			jakarta,
			"2006-01-02 15:04:05",
			false,
		},
		{
			"timezone aware comparison - before field (should fail)",
			DataObject{"start_date": "2024-01-15 12:00:00"},
			"2024-01-15 10:00:00",
			jakarta,
			"2006-01-02 15:04:05",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"end_date"},
			}
			validator := Time().Required().Format(tt.format).Timezone(tt.timezone).AfterField("start_date")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Time().Timezone().AfterField().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestTimeValidator_FieldComparisonWithTimeValue(t *testing.T) {
	// Test that time.Time values in root data are also supported
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	ctx := &ValidationContext{
		RootData: DataObject{"start_date": startTime},
		Path:     []string{"end_date"},
	}

	// End date is after start date
	validator := Time().Required().AfterField("start_date")
	errs := validator.Validate(ctx, "2024-01-15T12:00:00Z")
	if len(errs) > 0 {
		t.Errorf("Time().AfterField() should pass when end > start, got %v", errs)
	}

	// End date is before start date - should fail
	errs = validator.Validate(ctx, "2024-01-15T08:00:00Z")
	if len(errs) == 0 {
		t.Error("Time().AfterField() should fail when end < start")
	}
}

// ============================================================================
// STRING TRANSFORM TESTS
// ============================================================================

func TestStringValidator_Transform(t *testing.T) {
	tests := []struct {
		name       string
		transforms []StringTransformFunc
		value      string
		expected   string
	}{
		{
			"single transform lowercase",
			[]StringTransformFunc{strings.ToLower},
			"HELLO",
			"hello",
		},
		{
			"multiple transforms",
			[]StringTransformFunc{strings.ToLower, strings.TrimSpace},
			"  HELLO  ",
			"hello",
		},
		{
			"custom transform",
			[]StringTransformFunc{func(s string) string { return s + "_suffix" }},
			"hello",
			"hello_suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: DataObject{},
				Path:     []string{"field"},
			}
			validator := String().Required().Min(1)
			for _, transform := range tt.transforms {
				validator = validator.Transform(transform)
			}
			// Validation should pass since transforms are applied
			errs := validator.Validate(ctx, tt.value)
			if len(errs) > 0 {
				t.Errorf("String().Transform().Validate() unexpected error = %v", errs)
			}
		})
	}
}

// ============================================================================
// STRING CROSS-FIELD VALIDATION TESTS
// ============================================================================

func TestStringValidator_SameAs(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    string
		wantErr  bool
	}{
		{
			"same value",
			DataObject{"password": "secret123"},
			"secret123",
			false,
		},
		{
			"different value",
			DataObject{"password": "secret123"},
			"different",
			true,
		},
		{
			"field not exists",
			DataObject{},
			"secret123",
			false, // No error if field doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"password_confirmation"},
			}
			validator := String().Required().SameAs("password")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("String().SameAs().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_DifferentFrom(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    string
		wantErr  bool
	}{
		{
			"different value",
			DataObject{"old_password": "old123"},
			"new456",
			false,
		},
		{
			"same value",
			DataObject{"old_password": "same123"},
			"same123",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"new_password"},
			}
			validator := String().Required().DifferentFrom("old_password")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("String().DifferentFrom().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// NUMBER CROSS-FIELD VALIDATION TESTS
// ============================================================================

func TestNumberValidator_LessThan(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    float64
		wantErr  bool
	}{
		{
			"less than field",
			DataObject{"max": float64(100)},
			50,
			false,
		},
		{
			"equal to field",
			DataObject{"max": float64(100)},
			100,
			true,
		},
		{
			"greater than field",
			DataObject{"max": float64(100)},
			150,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"value"},
			}
			validator := Float().Required().LessThan("max")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Float().LessThan().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_GreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    float64
		wantErr  bool
	}{
		{
			"greater than field",
			DataObject{"min": float64(10)},
			50,
			false,
		},
		{
			"equal to field",
			DataObject{"min": float64(10)},
			10,
			true,
		},
		{
			"less than field",
			DataObject{"min": float64(10)},
			5,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"value"},
			}
			validator := Float().Required().GreaterThan("min")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Float().GreaterThan().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_LessThanOrEqual(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    float64
		wantErr  bool
	}{
		{
			"less than field",
			DataObject{"max": float64(100)},
			50,
			false,
		},
		{
			"equal to field",
			DataObject{"max": float64(100)},
			100,
			false,
		},
		{
			"greater than field",
			DataObject{"max": float64(100)},
			150,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"value"},
			}
			validator := Float().Required().LessThanOrEqual("max")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Float().LessThanOrEqual().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestNumberValidator_GreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		name     string
		rootData DataObject
		value    float64
		wantErr  bool
	}{
		{
			"greater than field",
			DataObject{"min": float64(10)},
			50,
			false,
		},
		{
			"equal to field",
			DataObject{"min": float64(10)},
			10,
			false,
		},
		{
			"less than field",
			DataObject{"min": float64(10)},
			5,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				RootData: tt.rootData,
				Path:     []string{"value"},
			}
			validator := Float().Required().GreaterThanOrEqual("min")
			errs := validator.Validate(ctx, tt.value)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("Float().GreaterThanOrEqual().Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// ARRAY CONCURRENT VALIDATION TESTS
// ============================================================================

func TestArrayValidator_Concurrent(t *testing.T) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}

	validator := Array().Required().Of(String().Required().Min(2)).Concurrent(4)
	value := []any{"ab", "cd", "ef", "gh", "ij"}

	errs := validator.Validate(ctx, value)
	if len(errs) > 0 {
		t.Errorf("Array().Concurrent().Validate() unexpected error = %v", errs)
	}
}

func TestArrayValidator_Concurrent_WithErrors(t *testing.T) {
	ctx := &ValidationContext{
		RootData: DataObject{},
		Path:     []string{"items"},
	}

	validator := Array().Required().Of(String().Required().Min(5)).Concurrent(4)
	value := []any{"ab", "cd", "valid_long", "ef"}

	errs := validator.Validate(ctx, value)
	// Should have errors for "ab", "cd", "ef"
	if len(errs) != 3 {
		t.Errorf("Array().Concurrent().Validate() expected 3 errors, got %d: %v", len(errs), errs)
	}
}

// ============================================================================
// PATH CACHE TESTS
// ============================================================================

func TestPathCache(t *testing.T) {
	// Test cache miss then hit
	path := "user.profile.settings"

	parts1 := globalPathCache.getSplitPath(path)
	parts2 := globalPathCache.getSplitPath(path)

	if len(parts1) != 3 || len(parts2) != 3 {
		t.Errorf("PathCache returned wrong parts: %v, %v", parts1, parts2)
	}

	if parts1[0] != "user" || parts1[1] != "profile" || parts1[2] != "settings" {
		t.Errorf("PathCache returned wrong values: %v", parts1)
	}
}

// ============================================================================
// POOL TESTS
// ============================================================================

func TestErrorMapPool(t *testing.T) {
	m := GetErrorMap()
	m["field1"] = []string{"error1", "error2"}
	m["field2"] = []string{"error3"}

	// Return to pool
	PutErrorMap(m)

	// Get again - should be cleared
	m2 := GetErrorMap()
	if len(m2) != 0 {
		t.Errorf("Error map from pool should be empty, got %v", m2)
	}
}

func TestBuilderPool(t *testing.T) {
	b := GetBuilder()
	b.WriteString("hello")
	b.WriteString(" world")

	result := b.String()
	if result != "hello world" {
		t.Errorf("Builder expected 'hello world', got %s", result)
	}

	PutBuilder(b)

	// Get again - should be reset
	b2 := GetBuilder()
	if b2.Len() != 0 {
		t.Errorf("Builder from pool should be empty, got length %d", b2.Len())
	}
}

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"user"}, "user"},
		{"two parts", []string{"user", "name"}, "user.name"},
		{"multiple parts", []string{"user", "profile", "settings", "theme"}, "user.profile.settings.theme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPath(tt.parts...)
			if result != tt.expected {
				t.Errorf("BuildPath() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name     string
		errs     []string
		sep      string
		expected string
	}{
		{"empty", []string{}, ", ", ""},
		{"single", []string{"error1"}, ", ", "error1"},
		{"multiple", []string{"error1", "error2", "error3"}, ", ", "error1, error2, error3"},
		{"custom separator", []string{"a", "b"}, " | ", "a | b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinErrors(tt.errs, tt.sep)
			if result != tt.expected {
				t.Errorf("JoinErrors() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestIntegration_OrderValidation(t *testing.T) {
	data := DataObject{
		"start_date": "2024-01-01T00:00:00Z",
		"end_date":   "2024-12-31T23:59:59Z",
		"min_price":  float64(10),
		"max_price":  float64(100),
		"items": []any{
			map[string]any{"name": "Item 1", "price": float64(25)},
			map[string]any{"name": "Item 2", "price": float64(50)},
		},
	}

	schema := Schema{
		"start_date": Time().Required(),
		"end_date":   Time().Required().AfterField("start_date"),
		"min_price":  Float().Required().Positive(),
		"max_price":  Float().Required().GreaterThan("min_price"),
		"items": Array().Required().Min(1).Concurrent(2).Of(Object().Shape(Schema{
			"name":  String().Required().Min(1),
			"price": Float().Required().Positive(),
		})),
	}

	err := Validate(data, schema)
	if err != nil {
		t.Errorf("Integration validation failed: %v", err.Errors)
	}
}

func TestIntegration_PasswordChange(t *testing.T) {
	data := DataObject{
		"old_password":     "old123",
		"new_password":     "new456",
		"confirm_password": "new456",
	}

	schema := Schema{
		"old_password":     String().Required().Min(6),
		"new_password":     String().Required().Min(6).DifferentFrom("old_password"),
		"confirm_password": String().Required().SameAs("new_password"),
	}

	err := Validate(data, schema)
	if err != nil {
		t.Errorf("Password change validation failed: %v", err.Errors)
	}
}

func TestIntegration_PasswordChange_SameAsOld(t *testing.T) {
	data := DataObject{
		"old_password":     "same123",
		"new_password":     "same123",
		"confirm_password": "same123",
	}

	schema := Schema{
		"old_password":     String().Required().Min(6),
		"new_password":     String().Required().Min(6).DifferentFrom("old_password"),
		"confirm_password": String().Required().SameAs("new_password"),
	}

	err := Validate(data, schema)
	if err == nil {
		t.Error("Expected error when new password same as old")
	}
}
