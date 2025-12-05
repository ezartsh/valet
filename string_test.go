package valet

import (
	"errors"
	"fmt"
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

// Tests for new string validators

func TestStringValidator_UUID(t *testing.T) {
	schema := Schema{"id": String().Required().UUID()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uuid v1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", false},
		{"valid uuid lowercase", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", false},
		{"valid uuid uppercase", "A0EEBC99-9C0B-4EF8-BB6D-6BB9BD380A11", false},
		{"invalid - too short", "550e8400-e29b-41d4-a716", true},
		{"invalid - no dashes", "550e8400e29b41d4a716446655440000", true},
		{"invalid - wrong format", "not-a-uuid", true},
		{"invalid - empty", "", true},
		{"invalid - wrong characters", "550e8400-e29b-41d4-a716-44665544000g", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"id": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("UUID(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_IPv4(t *testing.T) {
	schema := Schema{"ip": String().Required().IPv4()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ipv4", "192.168.1.1", false},
		{"valid ipv4 localhost", "127.0.0.1", false},
		{"valid ipv4 zeros", "0.0.0.0", false},
		{"valid ipv4 broadcast", "255.255.255.255", false},
		{"invalid - out of range", "256.1.1.1", true},
		{"invalid - ipv6", "::1", true},
		{"invalid - missing octets", "192.168.1", true},
		{"invalid - extra octets", "192.168.1.1.1", true},
		{"invalid - letters", "192.168.1.abc", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"ip": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("IPv4(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_IPv6(t *testing.T) {
	schema := Schema{"ip": String().Required().IPv6()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ipv6 loopback", "::1", false},
		{"valid ipv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"valid ipv6 compressed", "2001:db8:85a3::8a2e:370:7334", false},
		{"valid ipv6 all zeros", "::", false},
		{"invalid - ipv4", "192.168.1.1", true},
		{"invalid - random string", "not-an-ip", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"ip": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("IPv6(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_IP(t *testing.T) {
	schema := Schema{"ip": String().Required().IP()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ipv4", "192.168.1.1", false},
		{"valid ipv6", "::1", false},
		{"valid ipv6 full", "2001:db8::1", false},
		{"invalid - random string", "not-an-ip", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"ip": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("IP(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_JSON(t *testing.T) {
	schema := Schema{"data": String().Required().JSON()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid json object", `{"name": "John", "age": 30}`, false},
		{"valid json array", `[1, 2, 3]`, false},
		{"valid json string", `"hello"`, false},
		{"valid json number", `123`, false},
		{"valid json boolean", `true`, false},
		{"valid json null", `null`, false},
		{"valid json nested", `{"users": [{"name": "John"}]}`, false},
		{"invalid - missing quotes", `{name: "John"}`, true},
		{"invalid - trailing comma", `{"name": "John",}`, true},
		{"invalid - random string", `not json`, true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"data": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_HexColor(t *testing.T) {
	schema := Schema{"color": String().Required().HexColor()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid 6-digit hex", "#ff5733", false},
		{"valid 6-digit uppercase", "#FF5733", false},
		{"valid 3-digit hex", "#f53", false},
		{"valid 3-digit uppercase", "#F53", false},
		{"invalid - no hash", "ff5733", true},
		{"invalid - wrong length", "#ff573", true},
		{"invalid - invalid chars", "#gg5733", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"color": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexColor(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_DoesntStartWith(t *testing.T) {
	schema := Schema{"name": String().Required().DoesntStartWith("admin", "root", "system")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - regular name", "john_doe", false},
		{"valid - contains but not starts", "user_admin", false},
		{"invalid - starts with admin", "admin_user", true},
		{"invalid - starts with root", "root_access", true},
		{"invalid - starts with system", "system_config", true},
		{"invalid - exact match", "admin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"name": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoesntStartWith(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_DoesntEndWith(t *testing.T) {
	schema := Schema{"filename": String().Required().DoesntEndWith(".exe", ".bat", ".sh")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - txt file", "document.txt", false},
		{"valid - contains but not ends", "exe_file.txt", false},
		{"invalid - exe file", "virus.exe", true},
		{"invalid - bat file", "script.bat", true},
		{"invalid - sh file", "install.sh", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"filename": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoesntEndWith(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Includes(t *testing.T) {
	schema := Schema{"bio": String().Required().Includes("developer", "engineer")}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - contains all", "I am a software developer and engineer", false},
		{"invalid - missing developer", "I am an engineer", true},
		{"invalid - missing engineer", "I am a developer", true},
		{"invalid - missing both", "I am a designer", true},
		{"valid - case sensitive match", "developer engineer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"bio": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Includes(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_ASCII(t *testing.T) {
	schema := Schema{"text": String().Required().ASCII()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - letters", "Hello World", false},
		{"valid - numbers", "12345", false},
		{"valid - special chars", "!@#$%^&*()", false},
		{"valid - mixed", "Hello123!@#", false},
		{"invalid - unicode", "Hello 世界", true},
		{"invalid - emoji", "Hello 😀", true},
		{"invalid - accented", "café", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"text": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ASCII(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Base64(t *testing.T) {
	schema := Schema{"data": String().Required().Base64()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - simple", "SGVsbG8gV29ybGQ=", false},
		{"valid - with padding", "SGVsbG8=", false},
		{"valid - with numbers", "MTIzNDU2Nzg5MA==", false},
		{"invalid - special chars", "Hello!@#", true},
		{"invalid - spaces", "SGVsbG8g V29ybGQ=", true},
		{"invalid - unicode", "世界", true},
		{"invalid - no padding when needed", "SGVsbG8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"data": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_MAC(t *testing.T) {
	schema := Schema{"mac": String().Required().MAC()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - colon separated", "00:1A:2B:3C:4D:5E", false},
		{"valid - lowercase", "00:1a:2b:3c:4d:5e", false},
		{"valid - hyphen separated", "00-1A-2B-3C-4D-5E", false},
		{"invalid - wrong separator", "00.1A.2B.3C.4D.5E", true},
		{"invalid - too short", "00:1A:2B:3C:4D", true},
		{"invalid - too long", "00:1A:2B:3C:4D:5E:6F", true},
		{"invalid - wrong chars", "00:1A:2B:3C:4D:GG", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"mac": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("MAC(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_ULID(t *testing.T) {
	schema := Schema{"id": String().Required().ULID()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ulid", "01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"valid ulid lowercase", "01arz3ndektsv4rrffq69g5fav", false},
		{"invalid - too short", "01ARZ3NDEKTSV4RRFFQ69G5FA", true},
		{"invalid - too long", "01ARZ3NDEKTSV4RRFFQ69G5FAVX", true},
		{"invalid - contains I", "01ARZ3NDEKTSV4RRFFQ69G5FAI", true},
		{"invalid - contains L", "01ARZ3NDEKTSV4RRFFQ69G5FAL", true},
		{"invalid - contains O", "01ARZ3NDEKTSV4RRFFQ69G5FAO", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"id": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ULID(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_AlphaDash(t *testing.T) {
	schema := Schema{"slug": String().Required().AlphaDash()}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - letters only", "hello", false},
		{"valid - with numbers", "hello123", false},
		{"valid - with underscore", "hello_world", false},
		{"valid - with dash", "hello-world", false},
		{"valid - mixed", "hello_world-123", false},
		{"invalid - with space", "hello world", true},
		{"invalid - with special char", "hello@world", true},
		{"invalid - with dot", "hello.world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"slug": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlphaDash(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_Digits(t *testing.T) {
	schema := Schema{"code": String().Required().Digits(6)}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid - exact digits", "123456", false},
		{"valid - with zeros", "012345", false},
		{"invalid - too short", "12345", true},
		{"invalid - too long", "1234567", true},
		{"invalid - with letters", "12345a", true},
		{"invalid - with special char", "12345!", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(DataObject{"code": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Digits(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringValidator_CustomMessages(t *testing.T) {
	t.Run("UUID custom message", func(t *testing.T) {
		schema := Schema{"id": String().Required().UUID().Message("uuid", "Invalid UUID format")}
		err := Validate(DataObject{"id": "not-a-uuid"}, schema)
		if err == nil || err.Errors["id"][0] != "Invalid UUID format" {
			t.Error("Expected custom UUID message")
		}
	})

	t.Run("IP custom message", func(t *testing.T) {
		schema := Schema{"ip": String().Required().IP().Message("ip", "Invalid IP address")}
		err := Validate(DataObject{"ip": "not-an-ip"}, schema)
		if err == nil || err.Errors["ip"][0] != "Invalid IP address" {
			t.Error("Expected custom IP message")
		}
	})

	t.Run("JSON custom message", func(t *testing.T) {
		schema := Schema{"data": String().Required().JSON().Message("json", "Invalid JSON")}
		err := Validate(DataObject{"data": "not json"}, schema)
		if err == nil || err.Errors["data"][0] != "Invalid JSON" {
			t.Error("Expected custom JSON message")
		}
	})
}

func TestStringValidator_ChainedValidators(t *testing.T) {
	t.Run("UUID with min/max", func(t *testing.T) {
		schema := Schema{"id": String().Required().UUID().Min(36).Max(36)}
		err := Validate(DataObject{"id": "550e8400-e29b-41d4-a716-446655440000"}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("AlphaDash with length constraints", func(t *testing.T) {
		schema := Schema{"slug": String().Required().AlphaDash().Min(3).Max(20)}
		tests := []struct {
			value   string
			wantErr bool
		}{
			{"ab", true},                    // too short
			{"abc", false},                  // valid
			{"hello-world_123", false},      // valid
			{"aaaaaaaaaaaaaaaaaaaaa", true}, // too long (21 chars)
		}
		for _, tt := range tests {
			err := Validate(DataObject{"slug": tt.value}, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChainedValidators(%q) got error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		}
	})
}

// TestStringValidator_InlineCustomMessages tests the new inline custom message feature
func TestStringValidator_InlineCustomMessages(t *testing.T) {
	t.Run("Required with custom message", func(t *testing.T) {
		schema := Schema{"name": String().Required("Name is required")}
		err := Validate(DataObject{"name": ""}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["name"][0] != "Name is required" {
			t.Errorf("Expected 'Name is required', got: %s", err.Errors["name"][0])
		}
	})

	t.Run("Min with custom message", func(t *testing.T) {
		schema := Schema{"name": String().Required().Min(3, "Name must be at least 3 characters")}
		err := Validate(DataObject{"name": "ab"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["name"][0] != "Name must be at least 3 characters" {
			t.Errorf("Expected custom min message, got: %s", err.Errors["name"][0])
		}
	})

	t.Run("Max with custom message", func(t *testing.T) {
		schema := Schema{"name": String().Required().Max(5, "Name too long")}
		err := Validate(DataObject{"name": "toolongname"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["name"][0] != "Name too long" {
			t.Errorf("Expected 'Name too long', got: %s", err.Errors["name"][0])
		}
	})

	t.Run("Email with custom message", func(t *testing.T) {
		schema := Schema{"email": String().Required().Email("Please enter a valid email")}
		err := Validate(DataObject{"email": "not-an-email"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["email"][0] != "Please enter a valid email" {
			t.Errorf("Expected custom email message, got: %s", err.Errors["email"][0])
		}
	})

	t.Run("URL with custom message", func(t *testing.T) {
		schema := Schema{"website": String().Required().URL("Invalid website URL")}
		err := Validate(DataObject{"website": "not-a-url"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["website"][0] != "Invalid website URL" {
			t.Errorf("Expected custom URL message, got: %s", err.Errors["website"][0])
		}
	})

	t.Run("Regex with custom message", func(t *testing.T) {
		schema := Schema{"code": String().Required().Regex(`^\d{4}$`, "Code must be 4 digits")}
		err := Validate(DataObject{"code": "abc"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["code"][0] != "Code must be 4 digits" {
			t.Errorf("Expected custom regex message, got: %s", err.Errors["code"][0])
		}
	})

	t.Run("In with custom message", func(t *testing.T) {
		schema := Schema{"status": String().Required().InWithMessage("Invalid status", "active", "inactive")}
		err := Validate(DataObject{"status": "pending"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["status"][0] != "Invalid status" {
			t.Errorf("Expected 'Invalid status', got: %s", err.Errors["status"][0])
		}
	})

	t.Run("UUID with custom message", func(t *testing.T) {
		schema := Schema{"id": String().Required().UUID("Invalid UUID format")}
		err := Validate(DataObject{"id": "not-a-uuid"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["id"][0] != "Invalid UUID format" {
			t.Errorf("Expected 'Invalid UUID format', got: %s", err.Errors["id"][0])
		}
	})

	t.Run("StartsWith with custom message", func(t *testing.T) {
		schema := Schema{"code": String().Required().StartsWith("PRE_", "Code must start with PRE_")}
		err := Validate(DataObject{"code": "POST_123"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["code"][0] != "Code must start with PRE_" {
			t.Errorf("Expected custom startsWith message, got: %s", err.Errors["code"][0])
		}
	})

	t.Run("EndsWith with custom message", func(t *testing.T) {
		schema := Schema{"file": String().Required().EndsWith(".pdf", "Only PDF files allowed")}
		err := Validate(DataObject{"file": "document.doc"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["file"][0] != "Only PDF files allowed" {
			t.Errorf("Expected custom endsWith message, got: %s", err.Errors["file"][0])
		}
	})

	t.Run("Contains with custom message", func(t *testing.T) {
		schema := Schema{"bio": String().Required().Contains("experience", "Bio must mention experience")}
		err := Validate(DataObject{"bio": "I am a developer"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["bio"][0] != "Bio must mention experience" {
			t.Errorf("Expected custom contains message, got: %s", err.Errors["bio"][0])
		}
	})

	t.Run("Alpha with custom message", func(t *testing.T) {
		schema := Schema{"name": String().Required().Alpha("Name must contain only letters")}
		err := Validate(DataObject{"name": "John123"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["name"][0] != "Name must contain only letters" {
			t.Errorf("Expected custom alpha message, got: %s", err.Errors["name"][0])
		}
	})

	t.Run("AlphaNumeric with custom message", func(t *testing.T) {
		schema := Schema{"username": String().Required().AlphaNumeric("Username must be alphanumeric")}
		err := Validate(DataObject{"username": "user@name"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["username"][0] != "Username must be alphanumeric" {
			t.Errorf("Expected custom alphaNumeric message, got: %s", err.Errors["username"][0])
		}
	})

	t.Run("Length with custom message", func(t *testing.T) {
		schema := Schema{"code": String().Required().Length(6, "Code must be exactly 6 characters")}
		err := Validate(DataObject{"code": "1234"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["code"][0] != "Code must be exactly 6 characters" {
			t.Errorf("Expected custom length message, got: %s", err.Errors["code"][0])
		}
	})

	t.Run("IP with custom message", func(t *testing.T) {
		schema := Schema{"ip": String().Required().IP("Invalid IP address format")}
		err := Validate(DataObject{"ip": "invalid"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["ip"][0] != "Invalid IP address format" {
			t.Errorf("Expected custom IP message, got: %s", err.Errors["ip"][0])
		}
	})

	t.Run("JSON with custom message", func(t *testing.T) {
		schema := Schema{"data": String().Required().JSON("Must be valid JSON")}
		err := Validate(DataObject{"data": "not json"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["data"][0] != "Must be valid JSON" {
			t.Errorf("Expected custom JSON message, got: %s", err.Errors["data"][0])
		}
	})

	t.Run("Multiple validators with custom messages", func(t *testing.T) {
		schema := Schema{
			"email": String().Required("Email is required").Email("Invalid email format"),
		}

		// Test required message
		err := Validate(DataObject{"email": ""}, schema)
		if err == nil || err.Errors["email"][0] != "Email is required" {
			t.Errorf("Expected 'Email is required', got: %v", err)
		}

		// Test email message
		err = Validate(DataObject{"email": "invalid"}, schema)
		if err == nil || err.Errors["email"][0] != "Invalid email format" {
			t.Errorf("Expected 'Invalid email format', got: %v", err)
		}
	})
}

// TestStringValidator_MessageFunc tests dynamic message functions
func TestStringValidator_MessageFunc(t *testing.T) {
	t.Run("MessageFunc with field context", func(t *testing.T) {
		schema := Schema{
			"username": String().Required(func(ctx MessageContext) string {
				return "The " + ctx.Field + " field is required"
			}),
		}
		err := Validate(DataObject{"username": ""}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["username"][0] != "The username field is required" {
			t.Errorf("Expected dynamic message with field, got: %s", err.Errors["username"][0])
		}
	})

	t.Run("MessageFunc with value context", func(t *testing.T) {
		schema := Schema{
			"email": String().Required().Email(func(ctx MessageContext) string {
				return "'" + ctx.Value.(string) + "' is not a valid email address"
			}),
		}
		err := Validate(DataObject{"email": "bad-email"}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["email"][0] != "'bad-email' is not a valid email address" {
			t.Errorf("Expected dynamic message with value, got: %s", err.Errors["email"][0])
		}
	})

	t.Run("MessageFunc with array index context", func(t *testing.T) {
		schema := Schema{
			"users": Array().Of(Object().Shape(Schema{
				"name": String().Required(func(ctx MessageContext) string {
					if ctx.Index >= 0 {
						return fmt.Sprintf("User #%d name is required", ctx.Index+1)
					}
					return "Name is required"
				}),
			})),
		}
		err := Validate(DataObject{
			"users": []any{
				DataObject{"name": "John"},
				DataObject{"name": ""},
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		// Check that the error message contains the index
		errMsg := err.Errors["users.1.name"][0]
		if errMsg != "User #2 name is required" {
			t.Errorf("Expected 'User #2 name is required', got: %s", errMsg)
		}
	})

	t.Run("MessageFunc accessing data context", func(t *testing.T) {
		schema := Schema{
			"confirm_email": String().Required(func(ctx MessageContext) string {
				if ctx.Data != nil {
					if email, ok := ctx.Data["email"].(string); ok {
						return "Please confirm your email: " + email
					}
				}
				return "Confirmation required"
			}),
		}
		err := Validate(DataObject{
			"email":         "user@example.com",
			"confirm_email": "",
		}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["confirm_email"][0] != "Please confirm your email: user@example.com" {
			t.Errorf("Expected message with data context, got: %s", err.Errors["confirm_email"][0])
		}
	})

	t.Run("MessageFunc using Data.Get() for nested access", func(t *testing.T) {
		schema := Schema{
			"profile": Object().Shape(Schema{
				"bio": String().Required(func(ctx MessageContext) string {
					// Use Get() to access nested data
					name := ctx.Data.Get("profile.name").String()
					if name != "" {
						return "Bio is required for user: " + name
					}
					return "Bio is required"
				}),
			}),
		}
		err := Validate(DataObject{
			"profile": DataObject{
				"name": "John Doe",
				"bio":  "",
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["profile.bio"][0] != "Bio is required for user: John Doe" {
			t.Errorf("Expected message with Get(), got: %s", err.Errors["profile.bio"][0])
		}
	})

	t.Run("Data.Get() with array index", func(t *testing.T) {
		schema := Schema{
			"items": Array().Of(Object().Shape(Schema{
				"price": Num[float64]().Required().Positive(func(ctx MessageContext) string {
					// Get item name using Get()
					path := fmt.Sprintf("items.%d.name", ctx.Index)
					itemName := ctx.Data.Get(path).String()
					if itemName != "" {
						return fmt.Sprintf("Price for '%s' must be positive", itemName)
					}
					return "Price must be positive"
				}),
			})),
		}
		err := Validate(DataObject{
			"items": []any{
				DataObject{"name": "Apple", "price": 1.5},
				DataObject{"name": "Banana", "price": -0.5},
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err.Errors["items.1.price"][0] != "Price for 'Banana' must be positive" {
			t.Errorf("Expected message with array item name, got: %s", err.Errors["items.1.price"][0])
		}
	})

	t.Run("Data.Get() deeply nested value", func(t *testing.T) {
		schema := Schema{
			"settings": Object().Shape(Schema{
				"notifications": Object().Shape(Schema{
					"email": Bool().Required(func(ctx MessageContext) string {
						userName := ctx.Data.Get("user.profile.name").String()
						if userName != "" {
							return fmt.Sprintf("Email notification setting required for %s", userName)
						}
						return "Email notification setting is required"
					}),
				}),
			}),
		}
		err := Validate(DataObject{
			"user": DataObject{
				"profile": DataObject{
					"name": "Alice",
				},
			},
			"settings": DataObject{
				"notifications": DataObject{},
			},
		}, schema)
		if err == nil {
			t.Fatal("Expected error")
		}
		expectedMsg := "Email notification setting required for Alice"
		if err.Errors["settings.notifications.email"][0] != expectedMsg {
			t.Errorf("Expected '%s', got: %s", expectedMsg, err.Errors["settings.notifications.email"][0])
		}
	})
}

// TestStringValidator_DefaultMessagesUnchanged verifies default messages still work
func TestStringValidator_DefaultMessagesUnchanged(t *testing.T) {
	tests := []struct {
		name     string
		schema   Schema
		data     DataObject
		expected string
	}{
		{
			name:     "Required default message",
			schema:   Schema{"name": String().Required()},
			data:     DataObject{"name": ""},
			expected: "name is required",
		},
		{
			name:     "Email default message",
			schema:   Schema{"email": String().Required().Email()},
			data:     DataObject{"email": "invalid"},
			expected: "email must be a valid email",
		},
		{
			name:     "Min default message",
			schema:   Schema{"name": String().Required().Min(5)},
			data:     DataObject{"name": "ab"},
			expected: "name must be at least 5 characters",
		},
		{
			name:     "Max default message",
			schema:   Schema{"name": String().Required().Max(3)},
			data:     DataObject{"name": "toolong"},
			expected: "name must be at most 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, tt.schema)
			if err == nil {
				t.Fatal("Expected error")
			}
			for _, errs := range err.Errors {
				if len(errs) > 0 && errs[0] == tt.expected {
					return
				}
			}
			t.Errorf("Expected message '%s', got: %v", tt.expected, err.Errors)
		})
	}
}
