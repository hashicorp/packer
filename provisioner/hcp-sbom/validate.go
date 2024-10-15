package hcp_sbom

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/CycloneDX/cyclonedx-go"
	spdxjson "github.com/spdx/tools-golang/json"
)

// ValidateCycloneDX is a validation for CycloneDX in JSON format.
func ValidateCycloneDX(content io.Reader) error {
	decoder := cyclonedx.NewBOMDecoder(content, cyclonedx.BOMFileFormatJSON)
	bom := new(cyclonedx.BOM)
	if err := decoder.Decode(bom); err != nil {
		return fmt.Errorf("error parsing CycloneDX SBOM: %w", err)
	}

	if bom.BOMFormat != "CycloneDX" {
		return fmt.Errorf("invalid bomFormat: %s, expected CycloneDX", bom.BOMFormat)
	}
	if bom.SpecVersion.String() == "" {
		return fmt.Errorf("specVersion is required")
	}

	return nil
}

// ValidateSPDX is a validation for SPDX in JSON format.
func ValidateSPDX(content io.Reader) error {
	doc, err := spdxjson.Read(content)
	if err != nil {
		return fmt.Errorf("error parsing SPDX JSON file: %w", err)
	}

	if doc.SPDXVersion == "" {
		return fmt.Errorf("SPDX validation error: missing SPDXVersion")
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

	spdxErr := ValidateSPDX(reader)
	if spdxErr == nil {
		return "spdx", nil
	}
	if !strings.Contains(spdxErr.Error(), "error parsing") {
		return "", spdxErr
	}

	// Reset the reader's position
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset reader: %s", err)
	}

	cycloneDxErr := ValidateCycloneDX(reader)
	if cycloneDxErr == nil {
		return "cyclonedx", nil
	}
	if !strings.Contains(cycloneDxErr.Error(), "error parsing") {
		return "", cycloneDxErr
	}

	return "", fmt.Errorf("error validating SBOM file: invalid SBOM format")
}
