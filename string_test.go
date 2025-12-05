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
