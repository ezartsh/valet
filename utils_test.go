package valet

import (
	"testing"
)

func TestLookupPath(t *testing.T) {
	data := DataObject{
		"name": "John",
		"age":  float64(30),
		"address": map[string]any{
			"city":    "New York",
			"country": "USA",
			"zip": map[string]any{
				"code": "10001",
			},
		},
		"tags": []any{"go", "rust"},
	}

	tests := []struct {
		name     string
		path     string
		exists   bool
		expected any
	}{
		{"root level string", "name", true, "John"},
		{"root level number", "age", true, float64(30)},
		{"nested one level", "address.city", true, "New York"},
		{"nested two levels", "address.zip.code", true, "10001"},
		{"missing key", "missing", false, nil},
		{"missing nested key", "address.missing", false, nil},
		{"empty path", "", true, data},
		{"array value", "tags", true, []any{"go", "rust"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookupPath(data, tt.path)

			if result.Exists() != tt.exists {
				t.Errorf("lookupPath(%s).Exists() = %v, want %v", tt.path, result.Exists(), tt.exists)
			}

			if tt.exists && tt.expected != nil {
				switch expected := tt.expected.(type) {
				case string:
					if result.String() != expected {
						t.Errorf("lookupPath(%s) = %v, want %v", tt.path, result.String(), expected)
					}
				case float64:
					if result.Float() != expected {
						t.Errorf("lookupPath(%s) = %v, want %v", tt.path, result.Float(), expected)
					}
				}
			}
		})
	}
}

func TestLookupPath_EdgeCases(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		result := lookupPath(nil, "key")
		if result.Exists() {
			t.Error("lookupPath on nil should not exist")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		result := lookupPath(DataObject{}, "key")
		if result.Exists() {
			t.Error("lookupPath on empty data should not exist")
		}
	})

	t.Run("path with non-object intermediate", func(t *testing.T) {
		data := DataObject{
			"name": "John",
		}
		result := lookupPath(data, "name.first")
		if result.Exists() {
			t.Error("lookupPath through non-object should not exist")
		}
	})
}

func TestBuildFieldPath(t *testing.T) {
	tests := []struct {
		name     string
		path     []string
		expected string
	}{
		{"empty path", []string{}, ""},
		{"single path", []string{"user"}, "user"},
		{"multiple path", []string{"user", "profile", "name"}, "user.profile.name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFieldPath(tt.path)
			if result != tt.expected {
				t.Errorf("buildFieldPath(%v) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}
