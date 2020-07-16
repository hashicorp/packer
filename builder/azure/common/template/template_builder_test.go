package template

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	approvaltests "github.com/approvals/go-approval-tests"
)

// Ensure that a Linux template is configured as expected.
//  * Include SSH configuration: authorized key, and key path.
func TestBuildLinux00(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildLinux("--test-ssh-authorized-key--", true)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetMarketPlaceImage("Canonical", "UbuntuServer", "16.04", "latest", compute.CachingTypesReadWrite)
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

	err = testSubject.BuildLinux("--test-ssh-authorized-key--", false)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetImageUrl("http://azure/custom.vhd", compute.Linux, compute.CachingTypesReadWrite)
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

	testSubject.BuildLinux("--test-ssh-authorized-key--", true)
	testSubject.SetImageUrl("http://azure/custom.vhd", compute.Linux, compute.CachingTypesReadWrite)
	testSubject.SetOSDiskSizeGB(100)

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

	err = testSubject.SetMarketPlaceImage("MicrosoftWindowsServer", "WindowsServer", "2012-R2-Datacenter", "latest", compute.CachingTypesReadWrite)
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

// Windows build with additional disk for an managed build
func TestBuildWindows01(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildWindows("--test-key-vault-name", "--test-winrm-certificate-url--")
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetManagedMarketplaceImage("MicrosoftWindowsServer", "WindowsServer", "2012-R2-Datacenter", "latest", "2015-1", "1", "Premium_LRS", compute.CachingTypesReadWrite)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetAdditionalDisks([]int32{32, 64}, "datadisk", true, compute.CachingTypesReadWrite)
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

// Windows build with additional disk for an unmanaged build
func TestBuildWindows02(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildWindows("--test-key-vault-name", "--test-winrm-certificate-url--")
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetAdditionalDisks([]int32{32, 64}, "datadisk", false, compute.CachingTypesReadWrite)
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

// Shared Image Gallery Build
func TestSharedImageGallery00(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildLinux("--test-ssh-authorized-key--", false)
	if err != nil {
		t.Fatal(err)
	}

	imageID := "/subscriptions/ignore/resourceGroups/ignore/providers/Microsoft.Compute/galleries/ignore/images/ignore"
	err = testSubject.SetSharedGalleryImage("westcentralus", imageID, compute.CachingTypesReadOnly)
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

// Linux build with Network Security Group
func TestNetworkSecurityGroup00(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.BuildLinux("--test-ssh-authorized-key--", false)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetMarketPlaceImage("Canonical", "UbuntuServer", "16.04", "latest", compute.CachingTypesReadWrite)
	if err != nil {
		t.Fatal(err)
	}

	err = testSubject.SetNetworkSecurityGroup([]string{"127.0.0.1", "192.168.100.0/24"}, 123)
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

// Linux with user assigned managed identity configured
func TestSetIdentity00(t *testing.T) {
	testSubject, err := NewTemplateBuilder(BasicTemplate)
	if err != nil {
		t.Fatal(err)
	}

	if err = testSubject.BuildLinux("--test-ssh-authorized-key--", true); err != nil {
		t.Fatal(err)
	}

	if err = testSubject.SetMarketPlaceImage("Canonical", "UbuntuServer", "16.04", "latest", compute.CachingTypesReadWrite); err != nil {
		t.Fatal(err)
	}

	if err = testSubject.SetIdentity([]string{"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.ManagedIdentity/userAssignedIdentities/id"}); err != nil {
		t.Fatal(err)
	}

	doc, err := testSubject.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	if err = approvaltests.VerifyJSONBytes(t, []byte(*doc)); err != nil {
		t.Fatal(err)
	}
}
