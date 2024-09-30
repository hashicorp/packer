package hcp_sbom

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type MockUi struct {
	packer.Ui
}

func (m *MockUi) Say(message string) {
	fmt.Println(message)
}

func (m *MockUi) Error(message string) {
	fmt.Println("ERROR:", message)
}

func TestValidateSBOM(t *testing.T) {
	provisioner := &Provisioner{}
	ui := &MockUi{}

	tests := []struct {
		name        string
		sbom        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid SBOM",
			sbom: map[string]interface{}{
				"bomFormat":   "CycloneDX",
				"specVersion": "1.0",
			},
			expectError: false,
		},
		{
			name: "Invalid BomFormat",
			sbom: map[string]interface{}{
				"bomFormat":   "InvalidFormat",
				"specVersion": "1.0",
			},
			expectError: true,
			errorMsg:    "invalid bomFormat: InvalidFormat, expected CycloneDX",
		},
		{
			name: "Empty SpecVersion",
			sbom: map[string]interface{}{
				"bomFormat":   "CycloneDX",
				"specVersion": "",
			},
			expectError: true,
			errorMsg:    "failed to decode CycloneDX SBOM: invalid specification version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.sbom)
			filePath := "test-sbom.json"
			os.WriteFile(filePath, data, 0644)
			defer os.Remove(filePath)

			err := provisioner.validateSBOM(ui, filePath)
			if tt.expectError {
				if err == nil || err.Error() != tt.errorMsg {
					t.Fatalf("expected error %v, got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}
