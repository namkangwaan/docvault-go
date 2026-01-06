package security

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/namkangwaan/docvault-go/internal/config"
)

// Validator is the main security validator
type Validator struct {
	config             *config.SecurityConfig
	magicByteValidator *MagicByteValidator
	mimeDetector       *MimeDetector
	malwareScanner     *MalwareScanner
	documentValidator  *DocumentValidator
	quarantine         *Quarantine
	tempScanPath       string
}

// NewValidator creates a new security validator
func NewValidator(cfg *config.SecurityConfig) (*Validator, error) {
	// Create temp scan directory
	if err := os.MkdirAll(cfg.TempScanPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create temp scan directory: %w", err)
	}

	quarantine, err := NewQuarantine(cfg.EnableQuarantine, cfg.QuarantinePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize quarantine: %w", err)
	}

	return &Validator{
		config:             cfg,
		magicByteValidator: NewMagicByteValidator(cfg.EnableMagicByteCheck),
		mimeDetector:       NewMimeDetector(cfg.EnableMimeCheck),
		malwareScanner:     NewMalwareScanner(cfg.EnableMalwareScanner),
		documentValidator:  NewDocumentValidator(cfg.EnableDocumentCheck),
		quarantine:         quarantine,
		tempScanPath:       cfg.TempScanPath,
	}, nil
}

// ValidateFile performs comprehensive security validation on a file
func (v *Validator) ValidateFile(reader io.Reader, fileName string) (*Report, error) {
	// Get file extension
	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	ext = strings.ToLower(ext)

	// Create temporary file for scanning
	tempFile := filepath.Join(v.tempScanPath, fmt.Sprintf("%s_%s", uuid.New().String(), fileName))
	f, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Copy data to temp file
	written, err := io.Copy(f, reader)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to copy data: %w", err)
	}
	f.Close()

	// Read file data
	data, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}

	// Create report
	report := NewReport(fileName, written, ext)

	// Layer 1: Magic Bytes Validation
	if v.config.EnableMagicByteCheck {
		passed, errMsg := v.magicByteValidator.Validate(data, ext)
		report.MagicBytesPassed = passed
		report.MagicBytesError = errMsg
	}

	// Layer 2: MIME Type Verification
	if v.config.EnableMimeCheck {
		passed, detected, errMsg := v.mimeDetector.ValidateMimeType(data, ext)
		report.MimeTypePassed = passed
		report.MimeTypeDetected = detected
		report.MimeTypeError = errMsg
	}

	// Layer 3: Malicious Content Scanner
	if v.config.EnableMalwareScanner {
		scanResult := v.malwareScanner.Scan(data, ext)
		report.MalwareScanPassed = scanResult.IsSafe
		report.MalwareThreats = scanResult.Threats

		// Check for executables
		if v.malwareScanner.ScanForExecutables(data) {
			report.MalwareScanPassed = false
			report.MalwareThreats = append(report.MalwareThreats, "Embedded executable detected")
		}

		// Check for scripts
		scriptThreats := v.malwareScanner.ScanForScripts(data)
		if len(scriptThreats) > 0 {
			report.MalwareScanPassed = false
			report.MalwareThreats = append(report.MalwareThreats, scriptThreats...)
		}
	}

	// Layer 4: Document Structure Validation
	if v.config.EnableDocumentCheck {
		validationResult := v.documentValidator.Validate(data, ext)
		report.DocumentValidPassed = validationResult.IsValid
		report.DocumentErrors = validationResult.Errors
		report.DocumentWarnings = validationResult.Warnings
	}

	// Finalize report
	report.Finalize()

	// Quarantine if failed
	if !report.OverallPassed && v.config.EnableQuarantine {
		quarantineID, err := v.quarantine.QuarantineFile(tempFile, report.Summary())
		if err != nil {
			// Log error but don't fail the validation
			report.QuarantineID = fmt.Sprintf("Error: %v", err)
		} else {
			report.QuarantineID = quarantineID
		}
	}

	return report, nil
}

// ValidateUpload validates an uploaded file and returns the validated data
func (v *Validator) ValidateUpload(reader io.Reader, fileName string) (io.Reader, *Report, error) {
	// Create temporary file
	tempFile := filepath.Join(v.tempScanPath, fmt.Sprintf("%s_%s", uuid.New().String(), fileName))
	f, err := os.Create(tempFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy data to temp file
	_, err = io.Copy(f, reader)
	if err != nil {
		f.Close()
		os.Remove(tempFile)
		return nil, nil, fmt.Errorf("failed to copy data: %w", err)
	}
	f.Close()

	// Open for reading
	f, err = os.Open(tempFile)
	if err != nil {
		os.Remove(tempFile)
		return nil, nil, fmt.Errorf("failed to open temp file: %w", err)
	}

	// Validate
	report, err := v.ValidateFile(f, fileName)
	if err != nil {
		f.Close()
		os.Remove(tempFile)
		return nil, nil, err
	}

	// Seek to beginning
	f.Seek(0, 0)

	// If validation failed, reject the file
	if !report.OverallPassed {
		f.Close()
		os.Remove(tempFile)
		return nil, report, fmt.Errorf("file validation failed: %s", report.Summary())
	}

	// Return reader that will clean up temp file when done
	return &tempFileReader{file: f, path: tempFile}, report, nil
}

// tempFileReader is a reader that deletes the temp file when closed
type tempFileReader struct {
	file *os.File
	path string
}

func (r *tempFileReader) Read(p []byte) (n int, err error) {
	return r.file.Read(p)
}

func (r *tempFileReader) Close() error {
	r.file.Close()
	return os.Remove(r.path)
}
