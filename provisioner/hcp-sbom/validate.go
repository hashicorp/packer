package hcp_sbom

import (
	"bytes"
	"fmt"
	"github.com/CycloneDX/cyclonedx-go"
	spdxjson "github.com/spdx/tools-golang/json"
	"io"
)

// ErrorType represents the type of validation error.
type ErrorType string

const (
	ParsingErr    ErrorType = "parsing"
	ValidationErr ErrorType = "validation"
)

// ValidationError represents an error encountered while validating an SBOM.
type ValidationError struct {
	Type ErrorType
	Err  error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf(" %s error: %v", e.Type, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ValidateCycloneDX is a validation for CycloneDX in JSON format.
func ValidateCycloneDX(content io.Reader) error {
	decoder := cyclonedx.NewBOMDecoder(content, cyclonedx.BOMFileFormatJSON)
	bom := new(cyclonedx.BOM)
	if err := decoder.Decode(bom); err != nil {
		return &ValidationError{
			Type: ParsingErr,
			Err:  fmt.Errorf("error parsing CycloneDX SBOM: %w", err),
		}
	}

	if bom.BOMFormat != "CycloneDX" {
		return &ValidationError{
			Type: ValidationErr,
			Err:  fmt.Errorf("invalid bomFormat: %s, expected CycloneDX", bom.BOMFormat),
		}
	}
	if bom.SpecVersion.String() == "" {
		return &ValidationError{
			Type: ValidationErr,
			Err:  fmt.Errorf("specVersion is required"),
		}
	}

	return nil
}

// ValidateSPDX is a validation for SPDX in JSON format.
func ValidateSPDX(content io.Reader) error {
	doc, err := spdxjson.Read(content)
	if err != nil {
		return &ValidationError{
			Type: ParsingErr,
			Err:  fmt.Errorf("error parsing SPDX JSON file: %w", err),
		}
	}

	if doc.SPDXVersion == "" {
		return &ValidationError{
			Type: ValidationErr,
			Err:  fmt.Errorf("missing SPDXVersion"),
		}
	}

	return nil
}

// ValidateSBOM validates the SBOM file and returns the format of the SBOM.
func ValidateSBOM(content io.Reader) (string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, content); err != nil {
		return "", fmt.Errorf("failed to copy content: %s", err)
	}

	reader := bytes.NewReader(buf.Bytes())

	// Try validating as SPDX
	spdxErr := ValidateSPDX(reader)
	if spdxErr == nil {
		return "spdx", nil
	} else if vErr, ok := spdxErr.(*ValidationError); ok && vErr.Type == ValidationErr {
		return "", spdxErr
	}

	// Reset the reader's position
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset reader: %s", err)
	}

	cycloneDxErr := ValidateCycloneDX(reader)
	if cycloneDxErr == nil {
		return "cyclonedx", nil
	} else if vErr, ok := cycloneDxErr.(*ValidationError); ok && vErr.Type == ValidationErr {
		return "", spdxErr
	}

	return "", fmt.Errorf("error validating SBOM file: invalid SBOM format")
}
