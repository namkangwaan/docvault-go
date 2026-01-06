package security

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// DocumentValidator validates document structure
type DocumentValidator struct {
	enabled bool
}

// NewDocumentValidator creates a new document validator
func NewDocumentValidator(enabled bool) *DocumentValidator {
	return &DocumentValidator{enabled: enabled}
}

// ValidationResult represents document validation result
type ValidationResult struct {
	IsValid bool
	Errors  []string
	Warnings []string
}

// Validate validates document structure based on file type
func (v *DocumentValidator) Validate(data []byte, fileExt string) *ValidationResult {
	if !v.enabled {
		return &ValidationResult{IsValid: true}
	}

	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	switch fileExt {
	case "pdf":
		v.validatePDF(data, result)
	case "docx", "xlsx", "pptx":
		v.validateOOXML(data, result)
	}

	if len(result.Errors) > 0 {
		result.IsValid = false
	}

	return result
}

// validatePDF validates PDF structure
func (v *DocumentValidator) validatePDF(data []byte, result *ValidationResult) {
	// Check PDF header
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		result.Errors = append(result.Errors, "Invalid PDF header")
		return
	}

	// Check for EOF marker
	if !bytes.Contains(data, []byte("%%EOF")) {
		result.Warnings = append(result.Warnings, "PDF missing EOF marker")
	}

	// Basic structure checks
	requiredElements := []string{
		"/Type",
		"/Catalog",
	}

	for _, elem := range requiredElements {
		if !bytes.Contains(data, []byte(elem)) {
			result.Warnings = append(result.Warnings, "PDF missing element: "+elem)
		}
	}
}

// validateOOXML validates Office Open XML structure
func (v *DocumentValidator) validateOOXML(data []byte, result *ValidationResult) {
	// OOXML files are ZIP archives
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		result.Errors = append(result.Errors, "Invalid OOXML structure: not a valid ZIP archive")
		return
	}

	// Check for required files
	requiredFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
	}

	foundFiles := make(map[string]bool)
	for _, f := range zipReader.File {
		foundFiles[f.Name] = true
	}

	for _, required := range requiredFiles {
		if !foundFiles[required] {
			result.Errors = append(result.Errors, "Missing required OOXML file: "+required)
		}
	}

	// Check for ZIP bomb
	if v.isZipBomb(zipReader) {
		result.Errors = append(result.Errors, "Potential ZIP bomb detected (compression ratio too high)")
	}
}

// isZipBomb checks if the ZIP archive is a potential ZIP bomb
func (v *DocumentValidator) isZipBomb(zipReader *zip.Reader) bool {
	const maxCompressionRatio = 100 // 100:1 compression ratio threshold
	const maxTotalSize = 1024 * 1024 * 1024 // 1GB uncompressed size threshold

	var totalUncompressed uint64
	var totalCompressed uint64

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		totalUncompressed += f.UncompressedSize64
		totalCompressed += f.CompressedSize64

		// Check individual file size
		if f.UncompressedSize64 > maxTotalSize {
			return true
		}
	}

	// Check total uncompressed size
	if totalUncompressed > maxTotalSize {
		return true
	}

	// Check compression ratio
	if totalCompressed > 0 {
		ratio := totalUncompressed / totalCompressed
		if ratio > maxCompressionRatio {
			return true
		}
	}

	return false
}

// ValidateContentTypes validates [Content_Types].xml
func (v *DocumentValidator) ValidateContentTypes(zipReader *zip.Reader) error {
	for _, f := range zipReader.File {
		if f.Name == "[Content_Types].xml" {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open Content_Types.xml: %w", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return fmt.Errorf("failed to read Content_Types.xml: %w", err)
			}

			// Basic XML validation
			if !bytes.Contains(content, []byte("<Types")) {
				return fmt.Errorf("invalid Content_Types.xml structure")
			}

			return nil
		}
	}
	return fmt.Errorf("Content_Types.xml not found")
}

// GetOOXMLType determines the specific Office type from OOXML
func GetOOXMLType(zipReader *zip.Reader) string {
	for _, f := range zipReader.File {
		name := strings.ToLower(f.Name)
		if strings.Contains(name, "word/") {
			return "docx"
		}
		if strings.Contains(name, "xl/") {
			return "xlsx"
		}
		if strings.Contains(name, "ppt/") {
			return "pptx"
		}
	}
	return "unknown"
}
