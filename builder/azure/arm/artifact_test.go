package arm

import (
	"fmt"
	"strings"
	"testing"
)

func getFakeSasUrl(name string) string {
	return fmt.Sprintf("SAS-%s", name)
}

func TestArtifactId(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
							},
						},
					},
				},
				Location: "southcentralus",
			},
		},
	}

	artifact, err := NewArtifact(&template, getFakeSasUrl)
	if err != nil {
		t.Fatalf("err=%s", err)
	}

	expected := "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd"

	result := artifact.Id()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}

func TestArtifactString(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
							},
						},
					},
				},
				Location: "southcentralus",
			},
		},
	}

	artifact, err := NewArtifact(&template, getFakeSasUrl)
	if err != nil {
		t.Fatalf("err=%s", err)
	}

	testSubject := artifact.String()
	if !strings.Contains(testSubject, "OSDiskUri: https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd") {
		t.Errorf("Expected String() output to contain OSDiskUri")
	}
	if !strings.Contains(testSubject, "OSDiskUriReadOnlySas: SAS-Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd") {
		t.Errorf("Expected String() output to contain OSDiskUriReadOnlySas")
	}
	if !strings.Contains(testSubject, "TemplateUri: https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json") {
		t.Errorf("Expected String() output to contain TemplateUri")
	}
	if !strings.Contains(testSubject, "TemplateUriReadOnlySas: SAS-Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json") {
		t.Errorf("Expected String() output to contain TemplateUriReadOnlySas")
	}
	if !strings.Contains(testSubject, "StorageAccountLocation: southcentralus") {
		t.Errorf("Expected String() output to contain StorageAccountLocation")
	}
}

func TestAdditionalDiskArtifactString(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
							},
						},
						DataDisks: []CaptureDisk{
							{
								Image: CaptureUri{
									Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
								},
							},
						},
					},
				},
				Location: "southcentralus",
			},
		},
	}

	artifact, err := NewArtifact(&template, getFakeSasUrl)
	if err != nil {
		t.Fatalf("err=%s", err)
	}

	testSubject := artifact.String()
	if !strings.Contains(testSubject, "OSDiskUri: https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd") {
		t.Errorf("Expected String() output to contain OSDiskUri")
	}
	if !strings.Contains(testSubject, "OSDiskUriReadOnlySas: SAS-Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd") {
		t.Errorf("Expected String() output to contain OSDiskUriReadOnlySas")
	}
	if !strings.Contains(testSubject, "TemplateUri: https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json") {
		t.Errorf("Expected String() output to contain TemplateUri")
	}
	if !strings.Contains(testSubject, "TemplateUriReadOnlySas: SAS-Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json") {
		t.Errorf("Expected String() output to contain TemplateUriReadOnlySas")
	}
	if !strings.Contains(testSubject, "StorageAccountLocation: southcentralus") {
		t.Errorf("Expected String() output to contain StorageAccountLocation")
	}
	if !strings.Contains(testSubject, "AdditionalDiskUri (datadisk-1): https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd") {
		t.Errorf("Expected String() output to contain AdditionalDiskUri")
	}
	if !strings.Contains(testSubject, "AdditionalDiskUriReadOnlySas (datadisk-1): SAS-Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd") {
		t.Errorf("Expected String() output to contain AdditionalDiskUriReadOnlySas")
	}
}

func TestArtifactProperties(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
							},
						},
					},
				},
				Location: "southcentralus",
			},
		},
	}

	testSubject, err := NewArtifact(&template, getFakeSasUrl)
	if err != nil {
		t.Fatalf("err=%s", err)
	}

	if testSubject.OSDiskUri != "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd" {
		t.Errorf("Expected template to be 'https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd', but got %s", testSubject.OSDiskUri)
	}
	if testSubject.OSDiskUriReadOnlySas != "SAS-Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd" {
		t.Errorf("Expected template to be 'SAS-Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd', but got %s", testSubject.OSDiskUriReadOnlySas)
	}
	if testSubject.TemplateUri != "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json" {
		t.Errorf("Expected template to be 'https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json', but got %s", testSubject.TemplateUri)
	}
	if testSubject.TemplateUriReadOnlySas != "SAS-Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json" {
		t.Errorf("Expected template to be 'SAS-Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json', but got %s", testSubject.TemplateUriReadOnlySas)
	}
	if testSubject.StorageAccountLocation != "southcentralus" {
		t.Errorf("Expected StorageAccountLocation to be 'southcentral', but got %s", testSubject.StorageAccountLocation)
	}
}

func TestAdditionalDiskArtifactProperties(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
							},
						},
						DataDisks: []CaptureDisk{
							{
								Image: CaptureUri{
									Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
								},
							},
						},
					},
				},
				Location: "southcentralus",
			},
		},
	}

	testSubject, err := NewArtifact(&template, getFakeSasUrl)
	if err != nil {
		t.Fatalf("err=%s", err)
	}

	if testSubject.OSDiskUri != "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd" {
		t.Errorf("Expected template to be 'https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd', but got %s", testSubject.OSDiskUri)
	}
	if testSubject.OSDiskUriReadOnlySas != "SAS-Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd" {
		t.Errorf("Expected template to be 'SAS-Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd', but got %s", testSubject.OSDiskUriReadOnlySas)
	}
	if testSubject.TemplateUri != "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json" {
		t.Errorf("Expected template to be 'https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json', but got %s", testSubject.TemplateUri)
	}
	if testSubject.TemplateUriReadOnlySas != "SAS-Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json" {
		t.Errorf("Expected template to be 'SAS-Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json', but got %s", testSubject.TemplateUriReadOnlySas)
	}
	if testSubject.StorageAccountLocation != "southcentralus" {
		t.Errorf("Expected StorageAccountLocation to be 'southcentral', but got %s", testSubject.StorageAccountLocation)
	}
	if testSubject.AdditionalDisks == nil {
		t.Errorf("Expected AdditionalDisks to be not nil")
	}
	if len(*testSubject.AdditionalDisks) != 1 {
		t.Errorf("Expected AdditionalDisks to have one additional disk, but got %d", len(*testSubject.AdditionalDisks))
	}
	if (*testSubject.AdditionalDisks)[0].AdditionalDiskUri != "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd" {
		t.Errorf("Expected additional disk uri to be 'https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd', but got %s", (*testSubject.AdditionalDisks)[0].AdditionalDiskUri)
	}
	if (*testSubject.AdditionalDisks)[0].AdditionalDiskUriReadOnlySas != "SAS-Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd" {
		t.Errorf("Expected additional disk sas to be 'SAS-Images/images/packer-datadisk-1.4085bb15-3644-4641-b9cd-f575918640b4.vhd', but got %s", (*testSubject.AdditionalDisks)[0].AdditionalDiskUriReadOnlySas)
	}
}

func TestArtifactOverHypenatedCaptureUri(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/pac-ker-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd",
							},
						},
					},
				},
				Location: "southcentralus",
			},
		},
	}

	testSubject, err := NewArtifact(&template, getFakeSasUrl)
	if err != nil {
		t.Fatalf("err=%s", err)
	}

	if testSubject.TemplateUri != "https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/pac-ker-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json" {
		t.Errorf("Expected template to be 'https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/pac-ker-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json', but got %s", testSubject.TemplateUri)
	}
}

func TestArtifactRejectMalformedTemplates(t *testing.T) {
	template := CaptureTemplate{}

	_, err := NewArtifact(&template, getFakeSasUrl)
	if err == nil {
		t.Fatalf("Expected artifact creation to fail, but it succeeded.")
	}
}

func TestArtifactRejectMalformedStorageUri(t *testing.T) {
	template := CaptureTemplate{
		Resources: []CaptureResources{
			{
				Properties: CaptureProperties{
					StorageProfile: CaptureStorageProfile{
						OSDisk: CaptureDisk{
							Image: CaptureUri{
								Uri: "bark",
							},
						},
					},
				},
			},
		},
	}

	_, err := NewArtifact(&template, getFakeSasUrl)
	if err == nil {
		t.Fatalf("Expected artifact creation to fail, but it succeeded.")
	}
}
