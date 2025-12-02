package valet

import (
	"errors"
	"testing"
)

func TestStringValidator_Required(t *testing.T) {
	schema := Schema{"name": String().Required()}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"empty string", "", true},
		{"nil value", nil, true},
		{"valid string", "John", false},
		{"whitespace only", "   ", false}, // Not trimmed by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"name": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_RequiredIf(t *testing.T) {
	schema := Schema{
		"type": String().Required(),
		"value": String().RequiredIf(func(data DataObject) bool {
			return data["type"] == "custom"
		}),
	}

	t.Run("condition met - value missing", func(t *testing.T) {
		err := Validate(DataObject{"type": "custom"}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}
	})

	t.Run("condition met - value present", func(t *testing.T) {
		err := Validate(DataObject{"type": "custom", "value": "test"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		err := Validate(DataObject{"type": "default"}, schema)
		if err != nil {
			t.Errorf("Expected no error when condition not met, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_RequiredUnless(t *testing.T) {
	schema := Schema{
		"type": String().Required(),
		"value": String().RequiredUnless(func(data DataObject) bool {
			return data["type"] == "default"
		}),
	}

	t.Run("condition met - value not required", func(t *testing.T) {
		err := Validate(DataObject{"type": "default"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met - value required", func(t *testing.T) {
		err := Validate(DataObject{"type": "custom"}, schema)
		if err == nil {
			t.Error("Expected error when condition not met")
		}
	})
}

func TestStringValidator_MinMax(t *testing.T) {
	schema := Schema{"name": String().Required().Min(3).Max(10)}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"too short", "ab", true},
		{"min boundary", "abc", false},
		{"valid", "hello", false},
		{"max boundary", "1234567890", false},
		{"too long", "12345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"name": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Length(t *testing.T) {
	schema := Schema{"code": String().Required().Length(6)}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"too short", "12345", true},
		{"exact length", "123456", false},
		{"too long", "1234567", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"code": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Email(t *testing.T) {
	schema := Schema{"email": String().Required().Email()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid with subdomain", "test@mail.example.com", false},
		{"missing @", "testexample.com", true},
		{"missing domain", "test@", true},
		{"missing local", "@example.com", true},
		{"invalid chars", "test @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"email": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_URL(t *testing.T) {
	schema := Schema{"website": String().Required().URL()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"with path", "https://example.com/path", false},
		{"missing scheme", "example.com", true},
		{"invalid", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"website": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_URLWithOptions(t *testing.T) {
	t.Run("https only", func(t *testing.T) {
		schema := Schema{"url": String().Required().URLWithOptions(UrlOptions{Https: true})}

		err := Validate(DataObject{"url": "https://example.com"}, schema)
		if err != nil {
			t.Errorf("Expected no error for https, got: %v", err.Errors)
		}

		err = Validate(DataObject{"url": "http://example.com"}, schema)
		if err == nil {
			t.Error("Expected error for http when https required")
		}
	})

	t.Run("http only", func(t *testing.T) {
		schema := Schema{"url": String().Required().URLWithOptions(UrlOptions{Http: true})}

		err := Validate(DataObject{"url": "http://example.com"}, schema)
		if err != nil {
			t.Errorf("Expected no error for http, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Alpha(t *testing.T) {
	schema := Schema{"name": String().Required().Alpha()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"letters only", "John", false},
		{"with numbers", "John123", true},
		{"with space", "John Doe", true},
		{"with special", "John!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"name": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_AlphaNumeric(t *testing.T) {
	schema := Schema{"username": String().Required().AlphaNumeric()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"letters only", "John", false},
		{"with numbers", "John123", false},
		{"numbers only", "123456", false},
		{"with underscore", "John_Doe", true},
		{"with space", "John Doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"username": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Regex(t *testing.T) {
	schema := Schema{"phone": String().Required().Regex(`^\d{3}-\d{3}-\d{4}$`)}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid format", "123-456-7890", false},
		{"invalid format", "1234567890", true},
		{"partial match", "123-456-789", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"phone": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_NotRegex(t *testing.T) {
	schema := Schema{"text": String().Required().NotRegex(`badword`)}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"clean text", "hello world", false},
		{"contains bad word", "this has badword in it", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"text": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_In(t *testing.T) {
	schema := Schema{"status": String().Required().In("active", "inactive", "pending")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - active", "active", false},
		{"valid - inactive", "inactive", false},
		{"valid - pending", "pending", false},
		{"invalid", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"status": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_NotIn(t *testing.T) {
	schema := Schema{"username": String().Required().NotIn("admin", "root", "system")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid username", "john", false},
		{"reserved - admin", "admin", true},
		{"reserved - root", "root", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"username": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_StartsWith(t *testing.T) {
	schema := Schema{"code": String().Required().StartsWith("PRE_")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid prefix", "PRE_12345", false},
		{"missing prefix", "12345", true},
		{"wrong prefix", "POST_12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"code": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_EndsWith(t *testing.T) {
	schema := Schema{"file": String().Required().EndsWith(".txt")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid suffix", "document.txt", false},
		{"wrong suffix", "document.pdf", true},
		{"no suffix", "document", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"file": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Contains(t *testing.T) {
	schema := Schema{"text": String().Required().Contains("keyword")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"contains keyword", "this has keyword in it", false},
		{"missing keyword", "this has nothing", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"text": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Trim(t *testing.T) {
	schema := Schema{"name": String().Trim().Min(3)}

	t.Run("trimmed value too short", func(t *testing.T) {
		err := Validate(DataObject{"name": "  ab  "}, schema)
		if err == nil {
			t.Error("Expected error after trim")
		}
	})

	t.Run("trimmed value valid", func(t *testing.T) {
		err := Validate(DataObject{"name": "  abc  "}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Lowercase(t *testing.T) {
	schema := Schema{"code": String().Lowercase().In("abc", "def")}

	t.Run("uppercase converted", func(t *testing.T) {
		err := Validate(DataObject{"code": "ABC"}, schema)
		if err != nil {
			t.Errorf("Expected no error after lowercase, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Uppercase(t *testing.T) {
	schema := Schema{"code": String().Uppercase().In("ABC", "DEF")}

	t.Run("lowercase converted", func(t *testing.T) {
		err := Validate(DataObject{"code": "abc"}, schema)
		if err != nil {
			t.Errorf("Expected no error after uppercase, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Nullable(t *testing.T) {
	schema := Schema{"bio": String().Nullable().Max(100)}

	t.Run("null value allowed", func(t *testing.T) {
		err := Validate(DataObject{"bio": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})

	t.Run("valid value", func(t *testing.T) {
		err := Validate(DataObject{"bio": "Hello"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Default(t *testing.T) {
	schema := Schema{"role": String().Default("user").In("user", "admin")}

	t.Run("nil uses default", func(t *testing.T) {
		err := Validate(DataObject{"role": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error with default, got: %v", err.Errors)
		}
	})

	t.Run("provided value used", func(t *testing.T) {
		err := Validate(DataObject{"role": "admin"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_Custom(t *testing.T) {
	schema := Schema{
		"password": String().Required().Custom(func(v string, lookup Lookup) error {
			if v == "password123" {
				return errors.New("password is too common")
			}
			return nil
		}),
	}

	t.Run("custom validation fails", func(t *testing.T) {
		err := Validate(DataObject{"password": "password123"}, schema)
		if err == nil {
			t.Error("Expected error for common password")
		}
	})

	t.Run("custom validation passes", func(t *testing.T) {
		err := Validate(DataObject{"password": "secureP@ss123"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestStringValidator_TypeCheck(t *testing.T) {
	schema := Schema{"name": String().Required()}

	t.Run("non-string type", func(t *testing.T) {
		err := Validate(DataObject{"name": 123}, schema)
		if err == nil {
			t.Error("Expected error for non-string type")
		}
	})
}

func TestStringValidator_Message(t *testing.T) {
	schema := Schema{
		"email": String().Required().Email().
			Message("required", "Email is required").
			Message("email", "Please enter a valid email"),
	}

	t.Run("custom required message", func(t *testing.T) {
		err := Validate(DataObject{"email": ""}, schema)
		if err == nil || err.Errors["email"][0] != "Email is required" {
			t.Error("Expected custom required message")
		}
	})

	t.Run("custom email message", func(t *testing.T) {
		err := Validate(DataObject{"email": "invalid"}, schema)
		if err == nil || err.Errors["email"][0] != "Please enter a valid email" {
			t.Error("Expected custom email message")
		}
	})
}
