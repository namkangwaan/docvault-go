package security

import (
	"bytes"
	"io"
	"net/http"
)

// MimeDetector detects MIME types from file content
type MimeDetector struct {
	enabled bool
}

// NewMimeDetector creates a new MIME detector
func NewMimeDetector(enabled bool) *MimeDetector {
	return &MimeDetector{enabled: enabled}
}

// DetectMimeType detects MIME type from file content
func (d *MimeDetector) DetectMimeType(data []byte) string {
	if !d.enabled {
		return ""
	}

	return http.DetectContentType(data)
}

// ValidateMimeType checks if detected MIME type matches expected type
func (d *MimeDetector) ValidateMimeType(data []byte, expectedExt string) (bool, string, string) {
	if !d.enabled {
		return true, "", ""
	}

	detectedMime := d.DetectMimeType(data)

	// Map of extensions to acceptable MIME types
	acceptableMimes := map[string][]string{
		"pdf":  {"application/pdf"},
		"docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/zip"},
		"doc":  {"application/msword", "application/x-cfb"},
		"xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/zip"},
		"xls":  {"application/vnd.ms-excel", "application/x-cfb"},
	}

	acceptable, exists := acceptableMimes[expectedExt]
	if !exists {
		return false, detectedMime, "unsupported file extension"
	}

	for _, mime := range acceptable {
		if detectedMime == mime {
			return true, detectedMime, ""
		}
	}

	return false, detectedMime, "MIME type mismatch: expected one of " + formatMimes(acceptable) + ", detected " + detectedMime
}

// formatMimes formats MIME types for error messages
func formatMimes(mimes []string) string {
	if len(mimes) == 0 {
		return ""
	}
	if len(mimes) == 1 {
		return mimes[0]
	}
	result := "["
	for i, mime := range mimes {
		if i > 0 {
			result += ", "
		}
		result += mime
	}
	result += "]"
	return result
}

// ReadFirstBytes reads the first n bytes from a reader
func ReadFirstBytes(reader io.Reader, n int) ([]byte, error) {
	buffer := make([]byte, n)
	bytesRead, err := io.ReadFull(reader, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return buffer[:bytesRead], nil
}

// ReaderWithPrefix creates a reader that starts with prefix data
func ReaderWithPrefix(prefix []byte, reader io.Reader) io.Reader {
	return io.MultiReader(bytes.NewReader(prefix), reader)
}
