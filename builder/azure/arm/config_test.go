// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/packer"
)

// List of configuration parameters that are required by the ARM builder.
var requiredConfigValues = []string{
	"capture_name_prefix",
	"capture_container_name",
	"client_id",
	"client_secret",
	"image_offer",
	"image_publisher",
	"image_sku",
	"location",
	"os_type",
	"storage_account",
	"subscription_id",
	"tenant_id",
}

func TestConfigShouldProvideReasonableDefaultValues(t *testing.T) {
	c, _, err := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())

	if err != nil {
		t.Errorf("Expected configuration creation to succeed, but it failed!\n")
		t.Fatalf(" errors: %s\n", err)
	}

	if c.UserName == "" {
		t.Errorf("Expected 'UserName' to be populated, but it was empty!")
	}

	if c.VMSize == "" {
		t.Errorf("Expected 'VMSize' to be populated, but it was empty!")
	}

	if c.ObjectID != "" {
		t.Errorf("Expected 'ObjectID' to be nil, but it was '%s'!", c.ObjectID)
	}
}

func TestConfigShouldBeAbleToOverrideDefaultedValues(t *testing.T) {
	builderValues := getArmBuilderConfiguration()
	builderValues["ssh_password"] = "override_password"
	builderValues["ssh_username"] = "override_username"
	builderValues["vm_size"] = "override_vm_size"

	c, _, err := newConfig(builderValues, getPackerConfiguration())

	if err != nil {
		t.Fatalf("newConfig failed: %s", err)
	}

	if c.Password != "override_password" {
		t.Errorf("Expected 'Password' to be set to 'override_password', but found '%s'!", c.Password)
	}

	if c.Comm.SSHPassword != "override_password" {
		t.Errorf("Expected 'c.Comm.SSHPassword' to be set to 'override_password', but found '%s'!", c.Comm.SSHPassword)
	}

	if c.UserName != "override_username" {
		t.Errorf("Expected 'UserName' to be set to 'override_username', but found '%s'!", c.UserName)
	}

	if c.Comm.SSHUsername != "override_username" {
		t.Errorf("Expected 'c.Comm.SSHUsername' to be set to 'override_username', but found '%s'!", c.Comm.SSHUsername)
	}

	if c.VMSize != "override_vm_size" {
		t.Errorf("Expected 'vm_size' to be set to 'override_vm_size', but found '%s'!", c.VMSize)
	}
}

func TestConfigShouldDefaultVMSizeToStandardA1(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())

	if c.VMSize != "Standard_A1" {
		t.Errorf("Expected 'VMSize' to default to 'Standard_A1', but got '%s'.", c.VMSize)
	}
}

func TestConfigShouldDefaultImageVersionToLatest(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())

	if c.ImageVersion != "latest" {
		t.Errorf("Expected 'ImageVersion' to default to 'latest', but got '%s'.", c.ImageVersion)
	}
}

func TestConfigShouldDefaultToPublicCloud(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())

	if c.CloudEnvironmentName != "Public" {
		t.Errorf("Expected 'CloudEnvironmentName' to default to 'Public', but got '%s'.", c.CloudEnvironmentName)
	}

	if c.cloudEnvironment == nil || c.cloudEnvironment.Name != "AzurePublicCloud" {
		t.Errorf("Expected 'cloudEnvironment' to be set to 'AzurePublicCloud', but got '%s'.", c.cloudEnvironment)
	}
}

func TestConfigInstantiatesCorrectAzureEnvironment(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
	}

	// user input is fun :)
	var table = []struct {
		name            string
		environmentName string
	}{
		{"China", "AzureChinaCloud"},
		{"ChinaCloud", "AzureChinaCloud"},
		{"AzureChinaCloud", "AzureChinaCloud"},
		{"aZuReChInAcLoUd", "AzureChinaCloud"},

		{"USGovernment", "AzureUSGovernmentCloud"},
		{"USGovernmentCloud", "AzureUSGovernmentCloud"},
		{"AzureUSGovernmentCloud", "AzureUSGovernmentCloud"},
		{"aZuReUsGoVeRnMeNtClOuD", "AzureUSGovernmentCloud"},

		{"Public", "AzurePublicCloud"},
		{"PublicCloud", "AzurePublicCloud"},
		{"AzurePublicCloud", "AzurePublicCloud"},
		{"aZuRePuBlIcClOuD", "AzurePublicCloud"},
	}

	packerConfiguration := getPackerConfiguration()

	for _, x := range table {
		config["cloud_environment_name"] = x.name
		c, _, _ := newConfig(config, packerConfiguration)

		if c.cloudEnvironment == nil || c.cloudEnvironment.Name != x.environmentName {
			t.Errorf("Expected 'cloudEnvironment' to be set to '%s', but got '%s'.", x.environmentName, c.cloudEnvironment)
		}
	}
}

func TestUserShouldProvideRequiredValues(t *testing.T) {
	builderValues := getArmBuilderConfiguration()

	// Ensure we can successfully create a config.
	_, _, err := newConfig(builderValues, getPackerConfiguration())
	if err != nil {
		t.Errorf("Expected configuration creation to succeed, but it failed!\n")
		t.Fatalf(" -> %+v\n", builderValues)
	}

	// Take away a required element, and ensure construction fails.
	for _, v := range requiredConfigValues {
		originalValue := builderValues[v]
		delete(builderValues, v)

		_, _, err := newConfig(builderValues, getPackerConfiguration())
		if err == nil {
			t.Errorf("Expected configuration creation to fail, but it succeeded!\n")
			t.Fatalf(" -> %+v\n", builderValues)
		}

		builderValues[v] = originalValue
	}
}

func TestSystemShouldDefineRuntimeValues(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())

	if c.Password == "" {
		t.Errorf("Expected Password to not be empty, but it was '%s'!", c.Password)
	}

	if c.tmpComputeName == "" {
		t.Errorf("Expected tmpComputeName to not be empty, but it was '%s'!", c.tmpComputeName)
	}

	if c.tmpDeploymentName == "" {
		t.Errorf("Expected tmpDeploymentName to not be empty, but it was '%s'!", c.tmpDeploymentName)
	}

	if c.tmpResourceGroupName == "" {
		t.Errorf("Expected tmpResourceGroupName to not be empty, but it was '%s'!", c.tmpResourceGroupName)
	}

	if c.tmpOSDiskName == "" {
		t.Errorf("Expected tmpOSDiskName to not be empty, but it was '%s'!", c.tmpOSDiskName)
	}
}

func TestConfigShouldTransformToTemplateParameters(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	templateParameters := c.toTemplateParameters()

	if templateParameters.AdminUsername.Value != c.UserName {
		t.Errorf("Expected AdminUsername to be equal to config's AdminUsername, but they were '%s' and '%s' respectively.", templateParameters.AdminUsername.Value, c.UserName)
	}

	if templateParameters.DnsNameForPublicIP.Value != c.tmpComputeName {
		t.Errorf("Expected DnsNameForPublicIP to be equal to config's DnsNameForPublicIP, but they were '%s' and '%s' respectively.", templateParameters.DnsNameForPublicIP.Value, c.tmpComputeName)
	}

	if templateParameters.ImageOffer.Value != c.ImageOffer {
		t.Errorf("Expected ImageOffer to be equal to config's ImageOffer, but they were '%s' and '%s' respectively.", templateParameters.ImageOffer.Value, c.ImageOffer)
	}

	if templateParameters.ImagePublisher.Value != c.ImagePublisher {
		t.Errorf("Expected ImagePublisher to be equal to config's ImagePublisher, but they were '%s' and '%s' respectively.", templateParameters.ImagePublisher.Value, c.ImagePublisher)
	}

	if templateParameters.ImageSku.Value != c.ImageSku {
		t.Errorf("Expected ImageSku to be equal to config's ImageSku, but they were '%s' and '%s' respectively.", templateParameters.ImageSku.Value, c.ImageSku)
	}

	if templateParameters.OSDiskName.Value != c.tmpOSDiskName {
		t.Errorf("Expected OSDiskName to be equal to config's OSDiskName, but they were '%s' and '%s' respectively.", templateParameters.OSDiskName.Value, c.tmpOSDiskName)
	}

	if templateParameters.StorageAccountBlobEndpoint.Value != c.storageAccountBlobEndpoint {
		t.Errorf("Expected StorageAccountBlobEndpoint to be equal to config's storageAccountBlobEndpoint, but they were '%s' and '%s' respectively.", templateParameters.StorageAccountBlobEndpoint.Value, c.storageAccountBlobEndpoint)
	}

	if templateParameters.VMName.Value != c.tmpComputeName {
		t.Errorf("Expected VMName to be equal to config's VMName, but they were '%s' and '%s' respectively.", templateParameters.VMName.Value, c.tmpComputeName)
	}

	if templateParameters.VMSize.Value != c.VMSize {
		t.Errorf("Expected VMSize to be equal to config's VMSize, but they were '%s' and '%s' respectively.", templateParameters.VMSize.Value, c.VMSize)
	}
}

func TestConfigShouldTransformToTemplateParametersLinux(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	c.OSType = constants.Target_Linux
	templateParameters := c.toTemplateParameters()

	if templateParameters.KeyVaultSecretValue != nil {
		t.Errorf("Expected KeyVaultSecretValue to be empty for an os_type == '%s', but it was not.", c.OSType)
	}

	if templateParameters.ObjectId != nil {
		t.Errorf("Expected ObjectId to be empty for an os_type == '%s', but it was not.", c.OSType)
	}

	if templateParameters.TenantId != nil {
		t.Errorf("Expected TenantId to be empty for an os_type == '%s', but it was not.", c.OSType)
	}
}

func TestConfigShouldTransformToTemplateParametersWindows(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	c.OSType = constants.Target_Windows
	templateParameters := c.toTemplateParameters()

	if templateParameters.SshAuthorizedKey != nil {
		t.Errorf("Expected SshAuthorizedKey to be empty for an os_type == '%s', but it was not.", c.OSType)
	}

	if templateParameters.KeyVaultName == nil {
		t.Errorf("Expected KeyVaultName to not be empty for an os_type == '%s', but it was not.", c.OSType)
	}

	if templateParameters.KeyVaultSecretValue == nil {
		t.Errorf("Expected KeyVaultSecretValue to not be empty for an os_type == '%s', but it was not.", c.OSType)
	}

	if templateParameters.ObjectId == nil {
		t.Errorf("Expected ObjectId to not be empty for an os_type == '%s', but it was not.", c.OSType)
	}

	if templateParameters.TenantId == nil {
		t.Errorf("Expected TenantId to not be empty for an os_type == '%s', but it was not.", c.OSType)
	}
}

func TestConfigShouldTransformToVirtualMachineCaptureParameters(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	parameters := c.toVirtualMachineCaptureParameters()

	if *parameters.DestinationContainerName != c.CaptureContainerName {
		t.Errorf("Expected DestinationContainerName to be equal to config's CaptureContainerName, but they were '%s' and '%s' respectively.", *parameters.DestinationContainerName, c.CaptureContainerName)
	}

	if *parameters.VhdPrefix != c.CaptureNamePrefix {
		t.Errorf("Expected DestinationContainerName to be equal to config's CaptureContainerName, but they were '%s' and '%s' respectively.", *parameters.VhdPrefix, c.CaptureNamePrefix)
	}

	if *parameters.OverwriteVhds != false {
		t.Error("Expected OverwriteVhds to be false, but it was not.")
	}
}

func TestConfigShouldSupportPackersConfigElements(t *testing.T) {
	c, _, err := newConfig(
		getArmBuilderConfiguration(),
		getPackerConfiguration(),
		getPackerCommunicatorConfiguration())

	if err != nil {
		t.Fatal(err)
	}

	if c.Comm.SSHTimeout != 1*time.Hour {
		t.Errorf("Expected Comm.SSHTimeout to be a duration of an hour, but got '%s' instead.", c.Comm.SSHTimeout)
	}

	if c.Comm.WinRMTimeout != 2*time.Hour {
		t.Errorf("Expected Comm.WinRMTimeout to be a durationof two hours, but got '%s' instead.", c.Comm.WinRMTimeout)
	}
}

func TestUserDeviceLoginIsEnabledForLinux(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatalf("failed to use device login for Linux: %s", err)
	}
}

func TestUseDeviceLoginIsDisabledForWindows(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Windows,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatalf("Expected test to fail, but it succeeded")
	}

	multiError, _ := err.(*packer.MultiError)
	if len(multiError.Errors) != 3 {
		t.Errorf("Expected to find 3 errors, but found %d errors", len(multiError.Errors))
	}

	if !strings.Contains(err.Error(), "client_id must be specified") {
		t.Errorf("Expected to find error for 'client_id must be specified")
	}
	if !strings.Contains(err.Error(), "client_secret must be specified") {
		t.Errorf("Expected to find error for 'client_secret must be specified")
	}
	if !strings.Contains(err.Error(), "tenant_id must be specified") {
		t.Errorf("Expected to find error for 'tenant_id must be specified")
	}
}

func getArmBuilderConfiguration() map[string]string {
	m := make(map[string]string)
	for _, v := range requiredConfigValues {
		m[v] = fmt.Sprintf("%s00", v)
	}

	m["os_type"] = constants.Target_Linux
	return m
}

func getPackerConfiguration() interface{} {
	var doc = `{
		"packer_user_variables": {
			"sa": "my_storage_account"
		},
		"packer_build_name": "azure-arm-vm",
		"packer_builder_type": "azure-arm-vm",
		"packer_debug": "false",
		"packer_force": "false",
		"packer_template_path": "/home/jenkins/azure-arm-vm/template.json"
	}`

	var config interface{}
	json.Unmarshal([]byte(doc), &config)

	return config
}

func getPackerCommunicatorConfiguration() map[string]string {
	config := map[string]string{
		"ssh_timeout":   "1h",
		"winrm_timeout": "2h",
	}

	return config
}
