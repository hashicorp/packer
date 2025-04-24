package hcp_sbom

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/CycloneDX/cyclonedx-go"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	spdxjson "github.com/spdx/tools-golang/json"
)

// ValidationError represents an error encountered while validating an SBOM.
type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ValidateCycloneDX is a validation for CycloneDX in JSON format.
func validateCycloneDX(content []byte) error {
	decoder := cyclonedx.NewBOMDecoder(bytes.NewBuffer(content), cyclonedx.BOMFileFormatJSON)
	bom := new(cyclonedx.BOM)
	if err := decoder.Decode(bom); err != nil {
		return fmt.Errorf("error parsing CycloneDX SBOM: %w", err)
	}

	if !strings.EqualFold(bom.BOMFormat, "CycloneDX") {
		return &ValidationError{
			Err: fmt.Errorf("invalid bomFormat: %q, expected CycloneDX", bom.BOMFormat),
		}
	}
	if bom.SpecVersion.String() == "" {
		return &ValidationError{
			Err: fmt.Errorf("specVersion is required"),
		}
	}

	return nil
}

// validateSPDX is a validation for SPDX in JSON format.
func validateSPDX(content []byte) error {
	doc, err := spdxjson.Read(bytes.NewBuffer(content))
	if err != nil {
		return fmt.Errorf("error parsing SPDX JSON file: %w", err)
	}

	if doc.SPDXVersion == "" {
		return &ValidationError{
			Err: fmt.Errorf("missing SPDXVersion"),
		}
	}

	return nil
}

// validateSBOM validates the SBOM file and returns the format of the SBOM.
func validateSBOM(content []byte) (hcpPackerModels.HashicorpCloudPacker20230101SbomFormat, error) {
	// Try validating as SPDX
	spdxErr := validateSPDX(content)
	if spdxErr == nil {
		return hcpPackerModels.HashicorpCloudPacker20230101SbomFormatSPDX, nil
	}

	if vErr, ok := spdxErr.(*ValidationError); ok {
		return "", vErr
	}

	cycloneDxErr := validateCycloneDX(content)
	if cycloneDxErr == nil {
		return hcpPackerModels.HashicorpCloudPacker20230101SbomFormatCYCLONEDX, nil
	}

	if vErr, ok := cycloneDxErr.(*ValidationError); ok {
		return "", vErr
	}

	return "", fmt.Errorf("error validating SBOM file: invalid SBOM format")
}
