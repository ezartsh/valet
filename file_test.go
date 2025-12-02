package valet

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
)

// createTestFileHeader creates a multipart.FileHeader for testing
func createTestFileHeader(t *testing.T, filename string, content []byte, size int64) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "application/octet-stream")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("Failed to create form part: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("Failed to read form: %v", err)
	}

	files := form.File["file"]
	if len(files) == 0 {
		t.Fatalf("No file found in form")
	}

	files[0].Size = size
	return files[0]
}

// createPNGHeader creates PNG file header bytes
func createPNGHeader(width, height int) []byte {
	header := []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		'I', 'H', 'D', 'R', // IHDR chunk type
		byte(width >> 24), byte(width >> 16), byte(width >> 8), byte(width), // Width
		byte(height >> 24), byte(height >> 16), byte(height >> 8), byte(height), // Height
		0x08, 0x06, 0x00, 0x00, 0x00, // Bit depth, color type, compression, filter, interlace
	}
	return header
}

// createJPEGHeader creates JPEG file header bytes
func createJPEGHeader() []byte {
	return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
}

func TestFileValidator_Required(t *testing.T) {
	schema := Schema{"avatar": File().Required()}

	t.Run("nil value", func(t *testing.T) {
		err := Validate(DataObject{"avatar": nil}, schema)
		if err == nil {
			t.Error("Expected error for nil file")
		}
	})

	t.Run("valid file", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", []byte("content"), 7)
		err := Validate(DataObject{"avatar": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestFileValidator_RequiredIf(t *testing.T) {
	schema := Schema{
		"hasAvatar": Bool(),
		"avatar": File().RequiredIf(func(data DataObject) bool {
			return data["hasAvatar"] == true
		}),
	}

	t.Run("condition met - value missing", func(t *testing.T) {
		err := Validate(DataObject{"hasAvatar": true}, schema)
		if err == nil {
			t.Error("Expected error when condition met but value missing")
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		err := Validate(DataObject{"hasAvatar": false}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestFileValidator_RequiredUnless(t *testing.T) {
	schema := Schema{
		"useDefault": Bool(),
		"avatar": File().RequiredUnless(func(data DataObject) bool {
			return data["useDefault"] == true
		}),
	}

	t.Run("condition met - value not required", func(t *testing.T) {
		err := Validate(DataObject{"useDefault": true}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("condition not met - value required", func(t *testing.T) {
		err := Validate(DataObject{"useDefault": false}, schema)
		if err == nil {
			t.Error("Expected error when condition not met")
		}
	})
}

func TestFileValidator_MinMax(t *testing.T) {
	schema := Schema{"file": File().Required().Min(100).Max(1024)}

	t.Run("too small", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", []byte("small"), 50)
		err := Validate(DataObject{"file": fh}, schema)
		if err == nil {
			t.Error("Expected error for file too small")
		}
	})

	t.Run("valid size", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", make([]byte, 500), 500)
		err := Validate(DataObject{"file": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("too large", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", make([]byte, 2000), 2000)
		err := Validate(DataObject{"file": fh}, schema)
		if err == nil {
			t.Error("Expected error for file too large")
		}
	})
}

func TestFileValidator_Mimes(t *testing.T) {
	schema := Schema{"image": File().Required().Mimes("image/png", "image/jpeg")}

	t.Run("valid mime - png", func(t *testing.T) {
		content := createPNGHeader(100, 100)
		fh := createTestFileHeader(t, "test.png", content, int64(len(content)))
		err := Validate(DataObject{"image": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error for PNG, got: %v", err.Errors)
		}
	})

	t.Run("invalid mime", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", []byte("text content"), 12)
		err := Validate(DataObject{"image": fh}, schema)
		if err == nil {
			t.Error("Expected error for invalid mime type")
		}
	})
}

func TestFileValidator_Extensions(t *testing.T) {
	schema := Schema{"doc": File().Required().Extensions("pdf", "doc", "docx")}

	t.Run("valid extension", func(t *testing.T) {
		fh := createTestFileHeader(t, "document.pdf", []byte("%PDF-1.4"), 8)
		err := Validate(DataObject{"doc": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		fh := createTestFileHeader(t, "document.txt", []byte("text"), 4)
		err := Validate(DataObject{"doc": fh}, schema)
		if err == nil {
			t.Error("Expected error for invalid extension")
		}
	})
}

func TestFileValidator_Image(t *testing.T) {
	schema := Schema{"photo": File().Required().Image()}

	t.Run("valid image - png", func(t *testing.T) {
		content := createPNGHeader(100, 100)
		fh := createTestFileHeader(t, "photo.png", content, int64(len(content)))
		err := Validate(DataObject{"photo": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error for image, got: %v", err.Errors)
		}
	})

	t.Run("not an image", func(t *testing.T) {
		fh := createTestFileHeader(t, "file.txt", []byte("not an image"), 12)
		err := Validate(DataObject{"photo": fh}, schema)
		if err == nil {
			t.Error("Expected error for non-image file")
		}
	})
}

func TestFileValidator_Dimensions(t *testing.T) {
	t.Run("exact dimensions", func(t *testing.T) {
		schema := Schema{
			"icon": File().Required().Dimensions(&ImageDimensions{
				Width:  100,
				Height: 100,
			}),
		}

		content := createPNGHeader(100, 100)
		fh := createTestFileHeader(t, "icon.png", content, int64(len(content)))
		err := Validate(DataObject{"icon": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error for correct dimensions, got: %v", err.Errors)
		}
	})

	t.Run("wrong dimensions", func(t *testing.T) {
		schema := Schema{
			"icon": File().Required().Dimensions(&ImageDimensions{
				Width:  100,
				Height: 100,
			}),
		}

		content := createPNGHeader(200, 200)
		fh := createTestFileHeader(t, "icon.png", content, int64(len(content)))
		err := Validate(DataObject{"icon": fh}, schema)
		if err == nil {
			t.Error("Expected error for wrong dimensions")
		}
	})

	t.Run("min dimensions", func(t *testing.T) {
		schema := Schema{
			"banner": File().Required().Dimensions(&ImageDimensions{
				MinWidth:  800,
				MinHeight: 400,
			}),
		}

		content := createPNGHeader(1024, 512)
		fh := createTestFileHeader(t, "banner.png", content, int64(len(content)))
		err := Validate(DataObject{"banner": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})

	t.Run("max dimensions", func(t *testing.T) {
		schema := Schema{
			"thumb": File().Required().Dimensions(&ImageDimensions{
				MaxWidth:  200,
				MaxHeight: 200,
			}),
		}

		content := createPNGHeader(100, 100)
		fh := createTestFileHeader(t, "thumb.png", content, int64(len(content)))
		err := Validate(DataObject{"thumb": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestFileValidator_Nullable(t *testing.T) {
	schema := Schema{"attachment": File().Nullable()}

	t.Run("null value allowed", func(t *testing.T) {
		err := Validate(DataObject{"attachment": nil}, schema)
		if err != nil {
			t.Errorf("Expected no error for nullable, got: %v", err.Errors)
		}
	})
}

func TestFileValidator_Custom(t *testing.T) {
	schema := Schema{
		"file": File().Required().Custom(func(f *multipart.FileHeader, lookup Lookup) error {
			if f.Filename == "forbidden.txt" {
				return errors.New("this filename is not allowed")
			}
			return nil
		}),
	}

	t.Run("custom validation fails", func(t *testing.T) {
		fh := createTestFileHeader(t, "forbidden.txt", []byte("content"), 7)
		err := Validate(DataObject{"file": fh}, schema)
		if err == nil {
			t.Error("Expected error for forbidden filename")
		}
	})

	t.Run("custom validation passes", func(t *testing.T) {
		fh := createTestFileHeader(t, "allowed.txt", []byte("content"), 7)
		err := Validate(DataObject{"file": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err.Errors)
		}
	})
}

func TestFileValidator_TypeCheck(t *testing.T) {
	schema := Schema{"file": File().Required()}

	t.Run("non-file type", func(t *testing.T) {
		err := Validate(DataObject{"file": "not a file"}, schema)
		if err == nil {
			t.Error("Expected error for non-file type")
		}
	})
}

func TestFileValidator_Message(t *testing.T) {
	schema := Schema{
		"avatar": File().Required().Max(1024).
			Message("required", "Avatar is required").
			Message("max", "Avatar must be less than 1KB"),
	}

	t.Run("custom required message", func(t *testing.T) {
		err := Validate(DataObject{"avatar": nil}, schema)
		if err == nil || err.Errors["avatar"][0] != "Avatar is required" {
			t.Error("Expected custom required message")
		}
	})
}

func TestFileValidator_ValueType(t *testing.T) {
	schema := Schema{"file": File().Required()}

	t.Run("pointer type", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", []byte("content"), 7)
		err := Validate(DataObject{"file": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error for pointer type, got: %v", err.Errors)
		}
	})

	t.Run("value type", func(t *testing.T) {
		fh := createTestFileHeader(t, "test.txt", []byte("content"), 7)
		err := Validate(DataObject{"file": *fh}, schema)
		if err != nil {
			t.Errorf("Expected no error for value type, got: %v", err.Errors)
		}
	})
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 bytes"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestCheckAspectRatio(t *testing.T) {
	tests := []struct {
		width, height int
		ratio         string
		expected      bool
	}{
		{1920, 1080, "16/9", true},
		{1000, 1000, "1/1", true},
		{800, 600, "4/3", true},
		{1920, 1080, "4/3", false},
		{100, 100, "16/9", false},
	}

	for _, tt := range tests {
		t.Run(tt.ratio, func(t *testing.T) {
			result := checkAspectRatio(tt.width, tt.height, tt.ratio)
			if result != tt.expected {
				t.Errorf("checkAspectRatio(%d, %d, %s) = %v, want %v",
					tt.width, tt.height, tt.ratio, result, tt.expected)
			}
		})
	}
}

func TestMimeTypeGroups(t *testing.T) {
	t.Run("ImageMimes contains common types", func(t *testing.T) {
		expected := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
		for _, mime := range expected {
			found := false
			for _, m := range ImageMimes {
				if m == mime {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ImageMimes should contain %s", mime)
			}
		}
	})

	t.Run("DocumentMimes contains common types", func(t *testing.T) {
		expected := []string{"application/pdf", "text/plain"}
		for _, mime := range expected {
			found := false
			for _, m := range DocumentMimes {
				if m == mime {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("DocumentMimes should contain %s", mime)
			}
		}
	})
}

// TestFileValidationWithRealFiles tests with actual files if available
func TestFileValidationWithRealFiles(t *testing.T) {
	// Skip if test files don't exist
	testFiles := []string{"test_file_png.png", "test_file_jpg.jpg"}
	for _, filename := range testFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Skip("Test files not available")
			return
		}
	}

	t.Run("real PNG file", func(t *testing.T) {
		fh := createRealFileHeader(t, "test_file_png.png")
		schema := Schema{"image": File().Required().Image()}
		err := Validate(DataObject{"image": fh}, schema)
		if err != nil {
			t.Errorf("Expected no error for real PNG, got: %v", err.Errors)
		}
	})
}

func createRealFileHeader(t *testing.T, filePath string) *multipart.FileHeader {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", filePath, err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filepath.Base(filePath)+`"`)
	h.Set("Content-Type", "application/octet-stream")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("Failed to create form part: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("Failed to read form: %v", err)
	}

	files := form.File["file"]
	if len(files) == 0 {
		t.Fatalf("No file found in form")
	}

	files[0].Size = info.Size()
	return files[0]
}
