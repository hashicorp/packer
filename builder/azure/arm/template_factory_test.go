package arm

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/approvals/go-approval-tests"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/builder/azure/common/template"
)

// Ensure the link values are not set, and the concrete values are set.
func TestVirtualMachineDeployment00(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	if deployment.Properties.Mode != resources.Incremental {
		t.Errorf("Expected deployment.Properties.Mode to be %s, but got %s", resources.Incremental, deployment.Properties.Mode)
	}

	if deployment.Properties.ParametersLink != nil {
		t.Error("Expected the ParametersLink to be nil!")
	}

	if deployment.Properties.TemplateLink != nil {
		t.Error("Expected the TemplateLink to be nil!")
	}

	if deployment.Properties.Parameters == nil {
		t.Error("Expected the Parameters to not be nil!")
	}

	if deployment.Properties.Template == nil {
		t.Error("Expected the Template to not be nil!")
	}
}

// Ensure the Virtual Machine template is a valid JSON document.
func TestVirtualMachineDeployment01(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = json.Marshal(deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the Virtual Machine template parameters are correct.
func TestVirtualMachineDeployment02(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := json.Marshal(deployment.Properties.Parameters)
	if err != nil {
		t.Fatal(err)
	}

	var params template.TemplateParameters
	err = json.Unmarshal(bs, &params)
	if err != nil {
		t.Fatal(err)
	}

	if params.AdminUsername.Value != c.UserName {
		t.Errorf("Expected template parameter 'AdminUsername' to be %s, but got %s.", params.AdminUsername.Value, c.UserName)
	}
	if params.AdminPassword.Value != c.tmpAdminPassword {
		t.Errorf("Expected template parameter 'AdminPassword' to be %s, but got %s.", params.AdminPassword.Value, c.tmpAdminPassword)
	}
	if params.DnsNameForPublicIP.Value != c.tmpComputeName {
		t.Errorf("Expected template parameter 'DnsNameForPublicIP' to be %s, but got %s.", params.DnsNameForPublicIP.Value, c.tmpComputeName)
	}
	if params.OSDiskName.Value != c.tmpOSDiskName {
		t.Errorf("Expected template parameter 'OSDiskName' to be %s, but got %s.", params.OSDiskName.Value, c.tmpOSDiskName)
	}
	if params.StorageAccountBlobEndpoint.Value != c.storageAccountBlobEndpoint {
		t.Errorf("Expected template parameter 'StorageAccountBlobEndpoint' to be %s, but got %s.", params.StorageAccountBlobEndpoint.Value, c.storageAccountBlobEndpoint)
	}
	if params.VMSize.Value != c.VMSize {
		t.Errorf("Expected template parameter 'VMSize' to be %s, but got %s.", params.VMSize.Value, c.VMSize)
	}
	if params.VMName.Value != c.tmpComputeName {
		t.Errorf("Expected template parameter 'VMName' to be %s, but got %s.", params.VMName.Value, c.tmpComputeName)
	}
}

// Ensure the VM template is correct when using a market place image.
func TestVirtualMachineDeployment03(t *testing.T) {
	m := getArmBuilderConfiguration()
	m["image_publisher"] = "ImagePublisher"
	m["image_offer"] = "ImageOffer"
	m["image_sku"] = "ImageSku"
	m["image_version"] = "ImageVersion"

	c, _, _ := newConfig(m, getPackerConfiguration())
	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the VM template is correct when using a custom image.
func TestVirtualMachineDeployment04(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "https://localhost/custom.vhd",
		"resource_group_name":    "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVirtualMachineDeployment05(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":                 "ignore",
		"capture_container_name":              "ignore",
		"location":                            "ignore",
		"image_url":                           "https://localhost/custom.vhd",
		"resource_group_name":                 "ignore",
		"storage_account":                     "ignore",
		"subscription_id":                     "ignore",
		"os_type":                             constants.Target_Linux,
		"communicator":                        "none",
		"virtual_network_name":                "virtualNetworkName",
		"virtual_network_resource_group_name": "virtualNetworkResourceGroupName",
		"virtual_network_subnet_name":         "virtualNetworkSubnetName",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Verify that tags are properly applied to every resource
func TestVirtualMachineDeployment06(t *testing.T) {
	config := map[string]interface{}{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "https://localhost/custom.vhd",
		"resource_group_name":    "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
		"azure_tags": map[string]string{
			"tag01": "value01",
			"tag02": "value02",
			"tag03": "value03",
		},
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Verify that custom data are properly inserted
func TestVirtualMachineDeployment07(t *testing.T) {
	config := map[string]interface{}{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "https://localhost/custom.vhd",
		"resource_group_name":    "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	// The user specifies a configuration value for the setting custom_data_file.
	// The config type will read that file, and base64 encode it.  The encoded
	// contents are then assigned to Config's customData property, which are directly
	// injected into the template.
	//
	// I am not aware of an easy to mimic this situation in a test without having
	// a file on disk, which I am loathe to do.  The alternative is to inject base64
	// encoded data myself, which is what I am doing here.
	customData := `#cloud-config
growpart:
  mode: off
`
	base64CustomData := base64.StdEncoding.EncodeToString([]byte(customData))
	c.customData = base64CustomData

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the VM template is correct when building from a custom managed image.
func TestVirtualMachineDeployment08(t *testing.T) {
	config := map[string]interface{}{
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"os_type":                                  constants.Target_Linux,
		"communicator":                             "none",
		"custom_managed_image_resource_group_name": "CustomManagedImageResourceGroupName",
		"custom_managed_image_name":                "CustomManagedImageName",
		"managed_image_name":                       "ManagedImageName",
		"managed_image_resource_group_name":        "ManagedImageResourceGroupName",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the VM template is correct when building from a platform managed image.
func TestVirtualMachineDeployment09(t *testing.T) {
	config := map[string]interface{}{
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"os_type":                           constants.Target_Linux,
		"communicator":                      "none",
		"image_publisher":                   "--image-publisher--",
		"image_offer":                       "--image-offer--",
		"image_sku":                         "--image-sku--",
		"image_version":                     "--version--",
		"managed_image_name":                "ManagedImageName",
		"managed_image_resource_group_name": "ManagedImageResourceGroupName",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the VM template is correct when building with PublicIp and connect to Private Network
func TestVirtualMachineDeployment10(t *testing.T) {
	config := map[string]interface{}{
		"location":        "ignore",
		"subscription_id": "ignore",
		"os_type":         constants.Target_Linux,
		"communicator":    "none",
		"image_publisher": "--image-publisher--",
		"image_offer":     "--image-offer--",
		"image_sku":       "--image-sku--",
		"image_version":   "--version--",

		"virtual_network_resource_group_name":    "--virtual_network_resource_group_name--",
		"virtual_network_name":                   "--virtual_network_name--",
		"virtual_network_subnet_name":            "--virtual_network_subnet_name--",
		"private_virtual_network_with_public_ip": true,

		"managed_image_name":                "ManagedImageName",
		"managed_image_resource_group_name": "ManagedImageResourceGroupName",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	deployment, err := GetVirtualMachineDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the link values are not set, and the concrete values are set.
func TestKeyVaultDeployment00(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	deployment, err := GetKeyVaultDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	if deployment.Properties.Mode != resources.Incremental {
		t.Errorf("Expected deployment.Properties.Mode to be %s, but got %s", resources.Incremental, deployment.Properties.Mode)
	}

	if deployment.Properties.ParametersLink != nil {
		t.Error("Expected the ParametersLink to be nil!")
	}

	if deployment.Properties.TemplateLink != nil {
		t.Error("Expected the TemplateLink to be nil!")
	}

	if deployment.Properties.Parameters == nil {
		t.Error("Expected the Parameters to not be nil!")
	}

	if deployment.Properties.Template == nil {
		t.Error("Expected the Template to not be nil!")
	}
}

// Ensure the KeyVault template is a valid JSON document.
func TestKeyVaultDeployment01(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	deployment, err := GetKeyVaultDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = json.Marshal(deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}

// Ensure the KeyVault template parameters are correct.
func TestKeyVaultDeployment02(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfigurationWithWindows(), getPackerConfiguration())

	deployment, err := GetKeyVaultDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := json.Marshal(deployment.Properties.Parameters)
	if err != nil {
		t.Fatal(err)
	}

	var params template.TemplateParameters
	err = json.Unmarshal(bs, &params)
	if err != nil {
		t.Fatal(err)
	}

	if params.ObjectId.Value != c.ObjectID {
		t.Errorf("Expected template parameter 'ObjectId' to be %s, but got %s.", params.ObjectId.Value, c.ObjectID)
	}
	if params.TenantId.Value != c.TenantID {
		t.Errorf("Expected template parameter 'TenantId' to be %s, but got %s.", params.TenantId.Value, c.TenantID)
	}
	if params.KeyVaultName.Value != c.tmpKeyVaultName {
		t.Errorf("Expected template parameter 'KeyVaultName' to be %s, but got %s.", params.KeyVaultName.Value, c.tmpKeyVaultName)
	}
	if params.KeyVaultSecretValue.Value != c.winrmCertificate {
		t.Errorf("Expected template parameter 'KeyVaultSecretValue' to be %s, but got %s.", params.KeyVaultSecretValue.Value, c.winrmCertificate)
	}
}

// Ensure the KeyVault template is correct when tags are supplied.
func TestKeyVaultDeployment03(t *testing.T) {
	tags := map[string]interface{}{
		"azure_tags": map[string]string{
			"tag01": "value01",
			"tag02": "value02",
			"tag03": "value03",
		},
	}

	c, _, _ := newConfig(tags, getArmBuilderConfigurationWithWindows(), getPackerConfiguration())
	deployment, err := GetKeyVaultDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyJSONStruct(t, deployment.Properties.Template)
	if err != nil {
		t.Fatal(err)
	}
}
