package dtl

// these tests require the following variables to be set,
// although some test will only use a subset:
//
// * ARM_CLIENT_ID
// * ARM_CLIENT_SECRET
// * ARM_SUBSCRIPTION_ID
// * ARM_OBJECT_ID
//
// The subscription in question should have a resource group
// called "packer-acceptance-test" in "South Central US" region.
// This also requires a Devtest lab to be created with "packer-acceptance-test"
// name in "South Central US region. This can be achieved using the following
// az cli commands "
// az group create --name packer-acceptance-test --location "South Central US"
// az deployment group create \
//  --name ExampleDeployment \
//  --resource-group packer-acceptance-test \
//  --template-file acceptancetest.json \

// In addition, the PACKER_ACC variable should also be set to
// a non-empty value to enable Packer acceptance tests and the
// options "-v -timeout 90m" should be provided to the test
// command, e.g.:
//   go test -v -timeout 90m -run TestBuilderAcc_.*

import (
	"testing"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
)

const DeviceLoginAcceptanceTest = "DEVICELOGIN_TEST"

func TestBuilderAcc_ManagedDisk_Windows(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskWindows,
	})
}
func TestBuilderAcc_ManagedDisk_Linux_Artifacts(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskLinux,
	})
}

func testAccPreCheck(*testing.T) {}

const testBuilderAccManagedDiskWindows = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}",
	  "tenant_id": "{{env ` + "`ARM_TENANT_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",
	  "tenant_id": "{{user ` + "`tenant_id`" + `}}",

      "lab_name": "packer-acceptance-test",
	  "lab_resource_group_name":  "packer-acceptance-test",
	  "lab_virtual_network_name": "dtlpacker-acceptance-test",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindows-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccManagedDiskLinux = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}",
	  "tenant_id": "{{env ` + "`ARM_TENANT_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "lab_name": "packer-acceptance-test",
	  "lab_resource_group_name":  "packer-acceptance-test",
	  "lab_virtual_network_name": "dtlpacker-acceptance-test",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinux-{{timestamp}}",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2",


      "dtl_artifacts": [{
        "artifact_name": "linux-apt-package",
        "parameters" : [{
          "name": "packages",
          "value": "vim"
        },
        {
          "name":"update",
          "value": "true"
        },
        {
          "name": "options",
          "value": "--fix-broken"
		}]
		}]

	}]
}
`
