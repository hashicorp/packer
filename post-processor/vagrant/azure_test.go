package vagrant

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestAzureProvider_impl(t *testing.T) {
	var _ Provider = new(AzureProvider)
}

func TestAzureProvider_KeepInputArtifact(t *testing.T) {
	p := new(AzureProvider)

	if !p.KeepInputArtifact() {
		t.Fatal("should keep input artifact")
	}
}

func TestAzureProvider_ManagedImage(t *testing.T) {
	p := new(AzureProvider)
	ui := testUi()
	artifact := &packer.MockArtifact{
		StringValue: `Azure.ResourceManagement.VMImage:

OSType: Linux
ManagedImageResourceGroupName: packerruns
ManagedImageName: packer-1533651633
ManagedImageId: /subscriptions/e6229913-d9c3-4ddd-99a4-9e1ef3beaa1b/resourceGroups/packerruns/providers/Microsoft.Compute/images/packer-1533675589
ManagedImageLocation: westus`,
	}

	vagrantfile, _, err := p.Process(ui, artifact, "foo")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	result := `azure.location = "westus"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
	result = `azure.vm_managed_image_id = "/subscriptions/e6229913-d9c3-4ddd-99a4-9e1ef3beaa1b/resourceGroups/packerruns/providers/Microsoft.Compute/images/packer-1533675589"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
	// DO NOT set resource group in Vagrantfile, it should be separate from the image
	result = `azure.resource_group_name`
	if strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
	result = `azure.vm_operating_system`
	if strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
}

func TestAzureProvider_VHD(t *testing.T) {
	p := new(AzureProvider)
	ui := testUi()
	artifact := &packer.MockArtifact{
		IdValue: "https://packerbuildswest.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.96ed2120-591d-4900-95b0-ee8e985f2213.vhd",
		StringValue: `Azure.ResourceManagement.VMImage:

 OSType: Linux
 StorageAccountLocation: westus
 OSDiskUri: https://packerbuildswest.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.96ed2120-591d-4900-95b0-ee8e985f2213.vhd
 OSDiskUriReadOnlySas: https://packerbuildswest.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.96ed2120-591d-4900-95b0-ee8e985f2213.vhd?se=2018-09-07T18%3A36%3A34Z&sig=xUiFvwAviPYoP%2Bc91vErqvwYR1eK4x%2BAx7YLMe84zzU%3D&sp=r&sr=b&sv=2016-05-31
 TemplateUri: https://packerbuildswest.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.96ed2120-591d-4900-95b0-ee8e985f2213.json
 TemplateUriReadOnlySas: https://packerbuildswest.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.96ed2120-591d-4900-95b0-ee8e985f2213.json?se=2018-09-07T18%3A36%3A34Z&sig=lDxePyAUCZbfkB5ddiofimXfwk5INn%2F9E2BsnqIKC9Q%3D&sp=r&sr=b&sv=2016-05-31`,
	}

	vagrantfile, _, err := p.Process(ui, artifact, "foo")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	result := `azure.location = "westus"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
	result = `azure.vm_vhd_uri = "https://packerbuildswest.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.96ed2120-591d-4900-95b0-ee8e985f2213.vhd"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
	result = `azure.vm_operating_system = "Linux"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
	// DO NOT set resource group in Vagrantfile, it should be separate from the image
	result = `azure.resource_group_name`
	if strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
}
