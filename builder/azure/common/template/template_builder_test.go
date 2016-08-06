package template

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/approvals/go-approval-tests"
)

// Ensure that a Linux template is configured as expected.
//  * Include SSH configuration: authorized key, and key path.
func TestBuildLinux00(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
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

	err = approvaltests.VerifyJSONBytes(t, []byte(*doc))
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure that a user can specify a custom VHD when building a Linux template.
func TestBuildLinux01(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
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

	err = approvaltests.VerifyJSONBytes(t, []byte(*doc))
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure that a user can specify an existing Virtual Network
func TestBuildLinux02(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	testSubject.BuildLinux("--test-ssh-authorized-key--")
	testSubject.SetImageUrl("http://azure/custom.vhd", compute.Linux)

	err = testSubject.SetVirtualNetwork("--virtual-network-resource-group--", "--virtual-network--", "--subnet-name--")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := testSubject.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONBytes(t, []byte(*doc))
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure that a Windows template is configured as expected.
//  * Include WinRM configuration.
//  * Include KeyVault configuration, which is needed for WinRM.
func TestBuildWindows00(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
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

	err = approvaltests.VerifyJSONBytes(t, []byte(*doc))
	if err != nil {
		t.Fatal(err)
	}
}
