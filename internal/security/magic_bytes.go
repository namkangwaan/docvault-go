package security

import (
	"bytes"
	"encoding/hex"
)

// MagicBytes defines known file signatures
type MagicBytes struct {
	Extension string
	Signature []byte
	Offset    int
}

var knownSignatures = []MagicBytes{
	// PDF
	{Extension: "pdf", Signature: []byte{0x25, 0x50, 0x44, 0x46}, Offset: 0}, // %PDF
	
	// Microsoft Office (DOCX, XLSX, PPTX) - ZIP format
	{Extension: "docx", Signature: []byte{0x50, 0x4B, 0x03, 0x04}, Offset: 0}, // PK..
	{Extension: "xlsx", Signature: []byte{0x50, 0x4B, 0x03, 0x04}, Offset: 0}, // PK..
	{Extension: "pptx", Signature: []byte{0x50, 0x4B, 0x03, 0x04}, Offset: 0}, // PK..
	
	// Old Microsoft Office formats
	{Extension: "doc", Signature: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, Offset: 0},
	{Extension: "xls", Signature: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, Offset: 0},
}

// MagicByteValidator validates file signatures
type MagicByteValidator struct {
	enabled bool
}

// NewMagicByteValidator creates a new magic byte validator
func NewMagicByteValidator(enabled bool) *MagicByteValidator {
	return &MagicByteValidator{enabled: enabled}
}

// Validate checks if file signature matches expected type
func (v *MagicByteValidator) Validate(data []byte, extension string) (bool, string) {
	if !v.enabled {
		return true, ""
	}

	if len(data) < 8 {
		return false, "file too small to validate"
	}

	// Check if signature matches the claimed extension
	for _, sig := range knownSignatures {
		if sig.Extension != extension {
			continue
		}

		if len(data) < sig.Offset+len(sig.Signature) {
			continue
		}

		if bytes.Equal(data[sig.Offset:sig.Offset+len(sig.Signature)], sig.Signature) {
			return true, ""
		}
	}

	// Check what the actual file type might be
	detectedType := v.detectFileType(data)
	if detectedType != "" && detectedType != extension {
		return false, "file signature mismatch: claimed " + extension + ", detected " + detectedType
	}

	return false, "unknown or invalid file signature for extension: " + extension
}

// detectFileType attempts to detect the actual file type from magic bytes
func (v *MagicByteValidator) detectFileType(data []byte) string {
	for _, sig := range knownSignatures {
		if len(data) < sig.Offset+len(sig.Signature) {
			continue
		}

		if bytes.Equal(data[sig.Offset:sig.Offset+len(sig.Signature)], sig.Signature) {
			return sig.Extension
		}
	}
	return ""
}

// GetFileSignature returns the hex representation of file signature
func GetFileSignature(data []byte, length int) string {
	if len(data) < length {
		length = len(data)
	}
	return hex.EncodeToString(data[:length])
}
