package arm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/mitchellh/packer/builder/azure/common/approvals"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/builder/azure/common/template"
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
		t.Errorf("Expected the ParametersLink to be nil!")
	}

	if deployment.Properties.TemplateLink != nil {
		t.Errorf("Expected the TemplateLink to be nil!")
	}

	if deployment.Properties.Parameters == nil {
		t.Errorf("Expected the Parameters to not be nil!")
	}

	if deployment.Properties.Template == nil {
		t.Errorf("Expected the Template to not be nil!")
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

	bs, err := json.MarshalIndent(deployment.Properties.Template, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(string(bs))
	err = approvals.Verify(t, reader)
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

	bs, err := json.MarshalIndent(deployment.Properties.Template, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(string(bs))
	err = approvals.Verify(t, reader)
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
		t.Errorf("Expected the ParametersLink to be nil!")
	}

	if deployment.Properties.TemplateLink != nil {
		t.Errorf("Expected the TemplateLink to be nil!")
	}

	if deployment.Properties.Parameters == nil {
		t.Errorf("Expected the Parameters to not be nil!")
	}

	if deployment.Properties.Template == nil {
		t.Errorf("Expected the Template to not be nil!")
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

// Ensure the KeyVault template is correct.
func TestKeyVaultDeployment03(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfigurationWithWindows(), getPackerConfiguration())
	deployment, err := GetKeyVaultDeployment(c)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := json.MarshalIndent(deployment.Properties.Template, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(string(bs))
	err = approvals.Verify(t, reader)
	if err != nil {
		t.Fatal(err)
	}
}
