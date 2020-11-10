package arm

// these tests require the following variables to be set,
// although some test will only use a subset:
//
// * ARM_CLIENT_ID
// * ARM_CLIENT_SECRET
// * ARM_SUBSCRIPTION_ID
// * ARM_STORAGE_ACCOUNT
//
// The subscription in question should have a resource group
// called "packer-acceptance-test" in "South Central US" region. The
// storage account referred to in the above variable should
// be inside this resource group and in "South Central US" as well.
//
// In addition, the PACKER_ACC variable should also be set to
// a non-empty value to enable Packer acceptance tests and the
// options "-v -timeout 90m" should be provided to the test
// command, e.g.:
//   go test -v -timeout 90m -run TestBuilderAcc_.*

import (
	"testing"

	"fmt"
	"os"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
)

const DeviceLoginAcceptanceTest = "DEVICELOGIN_TEST"

func TestBuilderAcc_ManagedDisk_Windows(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskWindows,
	})
}

func TestBuilderAcc_ManagedDisk_Windows_Build_Resource_Group(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskWindowsBuildResourceGroup,
	})
}

func TestBuilderAcc_ManagedDisk_Windows_Build_Resource_Group_Additional_Disk(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskWindowsBuildResourceGroupAdditionalDisk,
	})
}

func TestBuilderAcc_ManagedDisk_Windows_DeviceLogin(t *testing.T) {
	if os.Getenv(DeviceLoginAcceptanceTest) == "" {
		t.Skip(fmt.Sprintf(
			"Device Login Acceptance tests skipped unless env '%s' set, as its requires manual step during execution",
			DeviceLoginAcceptanceTest))
		return
	}
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskWindowsDeviceLogin,
	})
}

func TestBuilderAcc_ManagedDisk_Linux(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskLinux,
	})
}

func TestBuilderAcc_ManagedDisk_Linux_DeviceLogin(t *testing.T) {
	if os.Getenv(DeviceLoginAcceptanceTest) == "" {
		t.Skip(fmt.Sprintf(
			"Device Login Acceptance tests skipped unless env '%s' set, as its requires manual step during execution",
			DeviceLoginAcceptanceTest))
		return
	}
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskLinuxDeviceLogin,
	})
}

func TestBuilderAcc_ManagedDisk_Linux_AzureCLI(t *testing.T) {
	if os.Getenv("AZURE_CLI_AUTH") == "" {
		t.Skip("Azure CLI Acceptance tests skipped unless env 'AZURE_CLI_AUTH' is set, and an active `az login` session has been established")
		return
	}

	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccManagedDiskLinuxAzureCLI,
	})
}

func TestBuilderAcc_Blob_Windows(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBlobWindows,
	})
}

func TestBuilderAcc_Blob_Linux(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBlobLinux,
	})
}

func testAccPreCheck(*testing.T) {}

const testBuilderAccManagedDiskWindows = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

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
	  "async_resourcegroup_delete": "true",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`
const testBuilderAccManagedDiskWindowsBuildResourceGroup = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "build_resource_group_name" : "packer-acceptance-test",
	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindowsBuildResourceGroup-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",
	  "async_resourcegroup_delete": "true",

	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccManagedDiskWindowsBuildResourceGroupAdditionalDisk = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "build_resource_group_name" : "packer-acceptance-test",
	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindowsBuildResourceGroupAdditionDisk-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",
	  "async_resourcegroup_delete": "true",

	  "vm_size": "Standard_DS2_v2",
	  "disk_additional_size": [10,15]
	}]
}
`

const testBuilderAccManagedDiskWindowsDeviceLogin = `
{
	"variables": {
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindowsDeviceLogin-{{timestamp}}",

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
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinux-{{timestamp}}",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2",
	  "azure_tags": {
	    "env": "testing",
	    "builder": "packer"
	   }
	}]
}
`
const testBuilderAccManagedDiskLinuxDeviceLogin = `
{
	"variables": {
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinuxDeviceLogin-{{timestamp}}",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",
	  "async_resourcegroup_delete": "true",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccBlobWindows = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}",
	  "storage_account": "{{env ` + "`ARM_STORAGE_ACCOUNT`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "storage_account": "{{user ` + "`storage_account`" + `}}",
	  "resource_group_name": "packer-acceptance-test",
	  "capture_container_name": "test",
	  "capture_name_prefix": "testBuilderAccBlobWin",

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

const testBuilderAccBlobLinux = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}",
	  "storage_account": "{{env ` + "`ARM_STORAGE_ACCOUNT`" + `}}"
	},
	"builders": [{
	  "type": "test",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "storage_account": "{{user ` + "`storage_account`" + `}}",
	  "resource_group_name": "packer-acceptance-test",
	  "capture_container_name": "test",
	  "capture_name_prefix": "testBuilderAccBlobLinux",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`
const testBuilderAccManagedDiskLinuxAzureCLI = `
{
	"builders": [{
	  "type": "test",

	  "use_azure_cli_auth": true,

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinuxAzureCLI-{{timestamp}}",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2",
	  "azure_tags": {
	    "env": "testing",
	    "builder": "packer"
	   }
	}]
}
`
