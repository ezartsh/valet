package valet

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// ImageDimensions defines constraints for image dimensions
type ImageDimensions struct {
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
	Width     int    // Exact width
	Height    int    // Exact height
	Ratio     string // e.g., "16/9", "1/1", "4/3"
}

// Common MIME type groups
var (
	ImageMimes = []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/bmp",
		"image/svg+xml",
		"image/webp",
		"image/tiff",
	}
	DocumentMimes = []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
	}
	VideoMimes = []string{
		"video/mp4",
		"video/mpeg",
		"video/quicktime",
		"video/x-msvideo",
		"video/webm",
	}
	AudioMimes = []string{
		"audio/mpeg",
		"audio/wav",
		"audio/ogg",
		"audio/mp4",
		"audio/webm",
	}
)

// FileValidator validates file uploads with fluent API
type FileValidator struct {
	required       bool
	requiredIf     func(data DataObject) bool
	requiredUnless func(data DataObject) bool
	min            int64
	minSet         bool
	max            int64
	maxSet         bool
	mimes          []string
	extensions     []string
	image          bool
	dimensions     *ImageDimensions
	customFn       func(file *multipart.FileHeader, lookup Lookup) error
	messages       map[string]string
	nullable       bool
}

// File creates a new file validator
func File() *FileValidator {
	return &FileValidator{
		messages: make(map[string]string),
	}
}

// Required marks the field as required
func (v *FileValidator) Required() *FileValidator {
	v.required = true
	return v
}

// RequiredIf makes field required based on condition
func (v *FileValidator) RequiredIf(fn func(data DataObject) bool) *FileValidator {
	v.requiredIf = fn
	return v
}

// RequiredUnless makes field required unless condition is met
func (v *FileValidator) RequiredUnless(fn func(data DataObject) bool) *FileValidator {
	v.requiredUnless = fn
	return v
}

// Min sets minimum file size in bytes
func (v *FileValidator) Min(bytes int64) *FileValidator {
	v.min = bytes
	v.minSet = true
	return v
}

// Max sets maximum file size in bytes
func (v *FileValidator) Max(bytes int64) *FileValidator {
	v.max = bytes
	v.maxSet = true
	return v
}

// Mimes sets allowed MIME types
func (v *FileValidator) Mimes(mimes ...string) *FileValidator {
	v.mimes = mimes
	return v
}

// Extensions sets allowed file extensions
func (v *FileValidator) Extensions(exts ...string) *FileValidator {
	v.extensions = exts
	return v
}

// Image requires file to be an image
func (v *FileValidator) Image() *FileValidator {
	v.image = true
	return v
}

// Dimensions sets image dimension constraints
func (v *FileValidator) Dimensions(d *ImageDimensions) *FileValidator {
	v.dimensions = d
	return v
}

// Custom adds custom validation function
func (v *FileValidator) Custom(fn func(file *multipart.FileHeader, lookup Lookup) error) *FileValidator {
	v.customFn = fn
	return v
}

// Message sets custom error message for a rule
func (v *FileValidator) Message(rule, message string) *FileValidator {
	v.messages[rule] = message
	return v
}

// Nullable allows null values
func (v *FileValidator) Nullable() *FileValidator {
	v.nullable = true
	return v
}

// Validate implements Validator interface
func (v *FileValidator) Validate(ctx *ValidationContext, value any) map[string][]string {
	errors := make(map[string][]string)
	fieldPath := ctx.FullPath()
	fieldName := ctx.Path[len(ctx.Path)-1]

	// Handle nil
	if value == nil {
		if v.nullable {
			return nil
		}
		if v.required {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		if v.requiredIf != nil && v.requiredIf(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		if v.requiredUnless != nil && !v.requiredUnless(ctx.RootData) {
			errors[fieldPath] = append(errors[fieldPath], v.msg("required", fmt.Sprintf("%s is required", fieldName)))
			return errors
		}
		return nil
	}

	// Type check - accept both pointer and value
	var file *multipart.FileHeader
	switch f := value.(type) {
	case *multipart.FileHeader:
		file = f
	case multipart.FileHeader:
		file = &f
	default:
		errors[fieldPath] = append(errors[fieldPath], v.msg("type", fmt.Sprintf("%s must be a file", fieldName)))
		return errors
	}

	// Min size
	if v.minSet && file.Size < v.min {
		errors[fieldPath] = append(errors[fieldPath], v.msg("min", fmt.Sprintf("%s must be at least %s", fieldName, formatFileSize(v.min))))
	}

	// Max size
	if v.maxSet && file.Size > v.max {
		errors[fieldPath] = append(errors[fieldPath], v.msg("max", fmt.Sprintf("%s must not be greater than %s", fieldName, formatFileSize(v.max))))
	}

	// MIME types
	if len(v.mimes) > 0 {
		detectedMime, err := detectMimeType(file)
		if err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("mimes", fmt.Sprintf("%s: unable to detect file type", fieldName)))
		} else {
			valid := false
			for _, mime := range v.mimes {
				if strings.EqualFold(detectedMime, mime) {
					valid = true
					break
				}
				// Handle wildcard mimes like "image/*"
				if strings.HasSuffix(mime, "/*") {
					prefix := strings.TrimSuffix(mime, "/*")
					if strings.HasPrefix(detectedMime, prefix+"/") {
						valid = true
						break
					}
				}
			}
			if !valid {
				errors[fieldPath] = append(errors[fieldPath], v.msg("mimes", fmt.Sprintf("%s must be a file of type: %s", fieldName, strings.Join(v.mimes, ", "))))
			}
		}
	}

	// Extensions
	if len(v.extensions) > 0 {
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
		valid := false
		for _, allowed := range v.extensions {
			if strings.EqualFold(ext, strings.TrimPrefix(allowed, ".")) {
				valid = true
				break
			}
		}
		if !valid {
			errors[fieldPath] = append(errors[fieldPath], v.msg("extensions", fmt.Sprintf("%s must be a file with extension: %s", fieldName, strings.Join(v.extensions, ", "))))
		}
	}

	// Image check
	if v.image {
		detectedMime, err := detectMimeType(file)
		if err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("image", fmt.Sprintf("%s must be an image", fieldName)))
		} else {
			isImage := false
			for _, imageMime := range ImageMimes {
				if strings.EqualFold(detectedMime, imageMime) {
					isImage = true
					break
				}
			}
			if !isImage {
				errors[fieldPath] = append(errors[fieldPath], v.msg("image", fmt.Sprintf("%s must be an image", fieldName)))
			}
		}
	}

	// Dimensions
	if v.dimensions != nil {
		width, height, err := getImageDimensions(file)
		if err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must be an image with valid dimensions", fieldName)))
		} else {
			d := v.dimensions
			if d.Width > 0 && width != d.Width {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have width of %d pixels", fieldName, d.Width)))
			}
			if d.Height > 0 && height != d.Height {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have height of %d pixels", fieldName, d.Height)))
			}
			if d.MinWidth > 0 && width < d.MinWidth {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have minimum width of %d pixels", fieldName, d.MinWidth)))
			}
			if d.MaxWidth > 0 && width > d.MaxWidth {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have maximum width of %d pixels", fieldName, d.MaxWidth)))
			}
			if d.MinHeight > 0 && height < d.MinHeight {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have minimum height of %d pixels", fieldName, d.MinHeight)))
			}
			if d.MaxHeight > 0 && height > d.MaxHeight {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have maximum height of %d pixels", fieldName, d.MaxHeight)))
			}
			if d.Ratio != "" && !checkAspectRatio(width, height, d.Ratio) {
				errors[fieldPath] = append(errors[fieldPath], v.msg("dimensions", fmt.Sprintf("%s must have aspect ratio of %s", fieldName, d.Ratio)))
			}
		}
	}

	// Custom validation
	if v.customFn != nil {
		lookup := func(path string) LookupResult {
			return lookupPath(ctx.RootData, path)
		}
		if err := v.customFn(file, lookup); err != nil {
			errors[fieldPath] = append(errors[fieldPath], v.msg("custom", err.Error()))
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func (v *FileValidator) msg(rule, defaultMsg string) string {
	if msg, ok := v.messages[rule]; ok {
		return msg
	}
	return defaultMsg
}

// Helper functions

func detectMimeType(fh *multipart.FileHeader) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:n])

	// Handle SVG detection
	if strings.Contains(contentType, "text/") {
		content := strings.ToLower(string(buffer[:n]))
		if strings.Contains(content, "<svg") || (strings.Contains(content, "<?xml") && strings.Contains(content, "svg")) {
			return "image/svg+xml", nil
		}
	}

	return contentType, nil
}

func getImageDimensions(fh *multipart.FileHeader) (int, int, error) {
	file, err := fh.Open()
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	buffer := make([]byte, 32)
	_, err = file.Read(buffer)
	if err != nil {
		return 0, 0, err
	}

	// PNG
	if buffer[0] == 0x89 && buffer[1] == 'P' && buffer[2] == 'N' && buffer[3] == 'G' {
		width := int(buffer[16])<<24 | int(buffer[17])<<16 | int(buffer[18])<<8 | int(buffer[19])
		height := int(buffer[20])<<24 | int(buffer[21])<<16 | int(buffer[22])<<8 | int(buffer[23])
		return width, height, nil
	}

	// JPEG
	if buffer[0] == 0xFF && buffer[1] == 0xD8 {
		return getJPEGDimensions(file)
	}

	// GIF
	if buffer[0] == 'G' && buffer[1] == 'I' && buffer[2] == 'F' {
		width := int(buffer[6]) | int(buffer[7])<<8
		height := int(buffer[8]) | int(buffer[9])<<8
		return width, height, nil
	}

	// WebP
	if buffer[0] == 'R' && buffer[1] == 'I' && buffer[2] == 'F' && buffer[3] == 'F' {
		return getWebPDimensions(file)
	}

	return 0, 0, fmt.Errorf("unsupported image format")
}

func getJPEGDimensions(file multipart.File) (int, int, error) {
	file.Seek(0, io.SeekStart)
	buf := make([]byte, 2)

	for {
		_, err := file.Read(buf)
		if err != nil {
			return 0, 0, err
		}

		if buf[0] != 0xFF {
			continue
		}

		marker := buf[1]
		if marker >= 0xC0 && marker <= 0xCF && marker != 0xC4 && marker != 0xC8 && marker != 0xCC {
			file.Read(buf)
			file.Read(make([]byte, 1))
			dimBuf := make([]byte, 4)
			file.Read(dimBuf)
			height := int(dimBuf[0])<<8 | int(dimBuf[1])
			width := int(dimBuf[2])<<8 | int(dimBuf[3])
			return width, height, nil
		}

		if marker != 0xD8 && marker != 0xD9 && marker != 0x00 {
			_, err := file.Read(buf)
			if err != nil {
				return 0, 0, err
			}
			length := int(buf[0])<<8 | int(buf[1])
			file.Seek(int64(length-2), io.SeekCurrent)
		}
	}
}

func getWebPDimensions(file multipart.File) (int, int, error) {
	file.Seek(0, io.SeekStart)
	header := make([]byte, 30)
	_, err := file.Read(header)
	if err != nil {
		return 0, 0, err
	}

	if header[12] == 'V' && header[13] == 'P' && header[14] == '8' {
		if header[15] == ' ' {
			width := int(header[26]) | int(header[27]&0x3F)<<8
			height := int(header[28]) | int(header[29]&0x3F)<<8
			return width, height, nil
		} else if header[15] == 'L' {
			bits := uint32(header[21]) | uint32(header[22])<<8 | uint32(header[23])<<16 | uint32(header[24])<<24
			width := int(bits&0x3FFF) + 1
			height := int((bits>>14)&0x3FFF) + 1
			return width, height, nil
		}
	}

	return 0, 0, fmt.Errorf("unsupported WebP format")
}

func checkAspectRatio(width, height int, ratio string) bool {
	parts := strings.Split(ratio, "/")
	if len(parts) != 2 {
		return false
	}

	var ratioW, ratioH int
	fmt.Sscanf(parts[0], "%d", &ratioW)
	fmt.Sscanf(parts[1], "%d", &ratioH)

	if ratioW == 0 || ratioH == 0 {
		return false
	}

	expectedRatio := float64(ratioW) / float64(ratioH)
	actualRatio := float64(width) / float64(height)

	tolerance := 0.01
	return actualRatio >= expectedRatio-tolerance && actualRatio <= expectedRatio+tolerance
}

func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
