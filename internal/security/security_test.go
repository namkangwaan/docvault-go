package security

import (
	"testing"
)

func TestMagicByteValidator(t *testing.T) {
	validator := NewMagicByteValidator(true)

	tests := []struct {
		name      string
		data      []byte
		extension string
		wantValid bool
	}{
		{
			name:      "Valid PDF",
			data:      []byte("%PDF-1.4\n"),
			extension: "pdf",
			wantValid: true,
		},
		{
			name:      "Invalid PDF",
			data:      []byte("not a pdf"),
			extension: "pdf",
			wantValid: false,
		},
		{
			name:      "Valid DOCX (ZIP header)",
			data:      []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00},
			extension: "docx",
			wantValid: true,
		},
		{
			name:      "File too small",
			data:      []byte{0x50},
			extension: "pdf",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := validator.Validate(tt.data, tt.extension)
			if valid != tt.wantValid {
				t.Errorf("Validate() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestMimeDetector(t *testing.T) {
	detector := NewMimeDetector(true)

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "PDF",
			data:     []byte("%PDF-1.4\n"),
			expected: "application/pdf",
		},
		{
			name:     "Plain text",
			data:     []byte("Hello, World!"),
			expected: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime := detector.DetectMimeType(tt.data)
			if mime != tt.expected {
				t.Errorf("DetectMimeType() = %v, want %v", mime, tt.expected)
			}
		})
	}
}

func TestMalwareScanner(t *testing.T) {
	scanner := NewMalwareScanner(true)

	tests := []struct {
		name     string
		data     []byte
		ext      string
		wantSafe bool
	}{
		{
			name:     "Clean PDF",
			data:     []byte("%PDF-1.4\nsome content"),
			ext:      "pdf",
			wantSafe: true,
		},
		{
			name:     "PDF with JavaScript",
			data:     []byte("%PDF-1.4\n/JavaScript\n/AA"),
			ext:      "pdf",
			wantSafe: false,
		},
		{
			name:     "DOCX with macros",
			data:     []byte("PK\x03\x04vbaProject.bin"),
			ext:      "docx",
			wantSafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.Scan(tt.data, tt.ext)
			if result.IsSafe != tt.wantSafe {
				t.Errorf("Scan() IsSafe = %v, want %v, threats: %v", result.IsSafe, tt.wantSafe, result.Threats)
			}
		})
	}
}
