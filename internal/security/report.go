package security

import (
	"fmt"
	"time"
)

// Report represents a security scan report
type Report struct {
	FileName            string
	FileSize            int64
	FileExtension       string
	ScanStartTime       time.Time
	ScanEndTime         time.Time
	ScanDuration        time.Duration
	MagicBytesPassed    bool
	MagicBytesError     string
	MimeTypePassed      bool
	MimeTypeDetected    string
	MimeTypeError       string
	MalwareScanPassed   bool
	MalwareThreats      []string
	DocumentValidPassed bool
	DocumentErrors      []string
	DocumentWarnings    []string
	AntivirusPassed     *bool
	AntivirusError      string
	OverallPassed       bool
	QuarantineID        string
}

// NewReport creates a new security report
func NewReport(fileName string, fileSize int64, fileExtension string) *Report {
	return &Report{
		FileName:            fileName,
		FileSize:            fileSize,
		FileExtension:       fileExtension,
		ScanStartTime:       time.Now(),
		MagicBytesPassed:    true,
		MimeTypePassed:      true,
		MalwareScanPassed:   true,
		DocumentValidPassed: true,
		OverallPassed:       true,
	}
}

// Finalize completes the report
func (r *Report) Finalize() {
	r.ScanEndTime = time.Now()
	r.ScanDuration = r.ScanEndTime.Sub(r.ScanStartTime)

	// Determine overall pass status
	r.OverallPassed = r.MagicBytesPassed &&
		r.MimeTypePassed &&
		r.MalwareScanPassed &&
		r.DocumentValidPassed

	if r.AntivirusPassed != nil && !*r.AntivirusPassed {
		r.OverallPassed = false
	}
}

// Summary returns a human-readable summary
func (r *Report) Summary() string {
	status := "PASSED"
	if !r.OverallPassed {
		status = "FAILED"
	}

	summary := fmt.Sprintf("Security Scan Report - %s\n", status)
	summary += fmt.Sprintf("File: %s (%.2f KB)\n", r.FileName, float64(r.FileSize)/1024)
	summary += fmt.Sprintf("Scan Duration: %v\n\n", r.ScanDuration)

	summary += fmt.Sprintf("Magic Bytes Check: %s\n", passFailStr(r.MagicBytesPassed))
	if r.MagicBytesError != "" {
		summary += fmt.Sprintf("  Error: %s\n", r.MagicBytesError)
	}

	summary += fmt.Sprintf("MIME Type Check: %s\n", passFailStr(r.MimeTypePassed))
	if r.MimeTypeDetected != "" {
		summary += fmt.Sprintf("  Detected: %s\n", r.MimeTypeDetected)
	}
	if r.MimeTypeError != "" {
		summary += fmt.Sprintf("  Error: %s\n", r.MimeTypeError)
	}

	summary += fmt.Sprintf("Malware Scan: %s\n", passFailStr(r.MalwareScanPassed))
	if len(r.MalwareThreats) > 0 {
		summary += "  Threats:\n"
		for _, threat := range r.MalwareThreats {
			summary += fmt.Sprintf("    - %s\n", threat)
		}
	}

	summary += fmt.Sprintf("Document Validation: %s\n", passFailStr(r.DocumentValidPassed))
	if len(r.DocumentErrors) > 0 {
		summary += "  Errors:\n"
		for _, err := range r.DocumentErrors {
			summary += fmt.Sprintf("    - %s\n", err)
		}
	}
	if len(r.DocumentWarnings) > 0 {
		summary += "  Warnings:\n"
		for _, warn := range r.DocumentWarnings {
			summary += fmt.Sprintf("    - %s\n", warn)
		}
	}

	if r.AntivirusPassed != nil {
		summary += fmt.Sprintf("Antivirus Scan: %s\n", passFailStr(*r.AntivirusPassed))
		if r.AntivirusError != "" {
			summary += fmt.Sprintf("  Error: %s\n", r.AntivirusError)
		}
	}

	if r.QuarantineID != "" {
		summary += fmt.Sprintf("\nFile quarantined with ID: %s\n", r.QuarantineID)
	}

	return summary
}

func passFailStr(passed bool) string {
	if passed {
		return "PASS"
	}
	return "FAIL"
}
