// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"strings"
	"testing"
)

func getFakeSasUrl(name string) string {
	return fmt.Sprintf("SAS-%s", name)
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
