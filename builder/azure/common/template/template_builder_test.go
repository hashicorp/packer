package template

import (
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/mitchellh/packer/builder/azure/common/approvals"
)

// Ensure that a Linux template is configured as expected.
//  * Include SSH configuration: authorized key, and key path.
func TestBuildLinux00(t *testing.T) {
	testSubject, err := NewTemplateBuilder()
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildLinux("--test-ssh-authorized-key--")
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetMarketPlaceImage("Canonical", "UbuntuServer", "16.04", "latest")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := testSubject.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(*doc)

	err = approvals.Verify(t, reader)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure that a user can specify a custom VHD when building a Linux template.
func TestBuildLinux01(t *testing.T) {
	testSubject, err := NewTemplateBuilder()
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildLinux("--test-ssh-authorized-key--")
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetImageUrl("http://azure/custom.vhd", compute.Linux)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := testSubject.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(*doc)

	err = approvals.Verify(t, reader)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure that a Windows template is configured as expected.
//  * Include WinRM configuration.
//  * Include KeyVault configuration, which is needed for WinRM.
func TestBuildWindows00(t *testing.T) {
	testSubject, err := NewTemplateBuilder()
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildWindows("--test-key-vault-name", "--test-winrm-certificate-url--")
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetMarketPlaceImage("MicrosoftWindowsServer", "WindowsServer", "2012-R2-Datacenter", "latest")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := testSubject.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(*doc)

	err = approvals.Verify(t, reader)
	if err != nil {
		t.Fatal(err)
	}
}
