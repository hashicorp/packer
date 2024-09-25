package hcp_sbom

import (
	"encoding/json"
	"fmt"
	"io"
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

type MockCommunicator struct {
	packer.Communicator
}

func (m *MockCommunicator) Download(src string, dst io.Writer) error {
	_, err := dst.Write([]byte("mock SBOM content"))
	return err
}

func TestDownloadSBOM(t *testing.T) {
	ui := &MockUi{}
	comm := &MockCommunicator{}

	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Source is a dir, Dest is a dir",
			config: Config{
				Source:      "mock-source/",
				Destination: "test-dir/",
			},
			expectError: true,
		},
		{
			name: "Source is a json file, Destination is a dir",
			config: Config{
				Source:      "mock-source/sbom.json",
				Destination: "test-dir/",
			},
			expectError: true,
		},
		{
			name: "Source is a json file, Destination is a json file",
			config: Config{
				Source:      "mock-source/sbom.json",
				Destination: "sbom.json",
			},
			expectError: false,
		},
		{
			name: "Source is a json file, Destination is a json file in test-output-data",
			config: Config{
				Source:      "mock-source/sbom.json",
				Destination: "test-output-data/sbom.json",
			},
			expectError: false,
		},
		{
			name: "Source is a json file, Destination is test-output-data w/o /",
			config: Config{
				Source:      "mock-source/sbom.json",
				Destination: "test-output-data",
			},
			expectError: true,
		},
		{
			name: "Source is a json file, Destination is empty",
			config: Config{
				Source: "mock-source/sbom.json",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provisioner := &Provisioner{
				config: tt.config,
			}
			generatedData := map[string]interface{}{}
			destPath, err := provisioner.downloadSBOMForPacker(ui, comm, generatedData)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Fatalf("expected file to exist at %s", destPath)
				}

				os.RemoveAll(destPath)
			}
		})
	}
}

func TestValidateSBOM(t *testing.T) {
	provisioner := &Provisioner{}
	ui := &MockUi{}

	tests := []struct {
		name        string
		sbom        SBOM
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid SBOM",
			sbom: SBOM{
				BomFormat:   "CycloneDX",
				SpecVersion: "1.0",
			},
			expectError: false,
		},
		{
			name: "Invalid BomFormat",
			sbom: SBOM{
				BomFormat:   "InvalidFormat",
				SpecVersion: "1.0",
			},
			expectError: true,
			errorMsg:    "invalid bomFormat: InvalidFormat",
		},
		{
			name: "Empty SpecVersion",
			sbom: SBOM{
				BomFormat:   "CycloneDX",
				SpecVersion: "",
			},
			expectError: true,
			errorMsg:    "specVersion is required",
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
