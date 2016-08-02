// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
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
	"resource_group_name",
	"subscription_id",
}

func TestConfigShouldProvideReasonableDefaultValues(t *testing.T) {
	c, _, err := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())

	if err != nil {
		t.Error("Expected configuration creation to succeed, but it failed!\n")
		t.Fatalf(" errors: %s\n", err)
	}

	if c.UserName == "" {
		t.Error("Expected 'UserName' to be populated, but it was empty!")
	}

	if c.VMSize == "" {
		t.Error("Expected 'VMSize' to be populated, but it was empty!")
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
	builderValues["communicator"] = "ssh"

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

func TestConfigShouldNotDefaultImageVersionIfCustomImage(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
	}

	c, _, _ := newConfig(config, getPackerConfiguration())
	if c.ImageVersion != "" {
		t.Errorf("Expected 'ImageVersion' to empty, but got '%s'.", c.ImageVersion)
	}
}

func TestConfigShouldRejectCustomImageAndMarketPlace(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "ignore",
		"resource_group_name":    "ignore",
		"storage_account":        "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
	}
	packerConfiguration := getPackerConfiguration()
	marketPlace := []string{"image_publisher", "image_offer", "image_sku"}

	for _, x := range marketPlace {
		config[x] = "ignore"
		_, _, err := newConfig(config, packerConfiguration)
		if err == nil {
			t.Errorf("Expected Config to reject image_url and %s, but it did not", x)
		}
	}
}

func TestConfigVirtualNetworkNameIsOptional(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
		"virtual_network_name":   "MyVirtualNetwork",
	}

	c, _, _ := newConfig(config, getPackerConfiguration())
	if c.VirtualNetworkName != "MyVirtualNetwork" {
		t.Errorf("Expected Config to set virtual_network_name to MyVirtualNetwork, but got %q", c.VirtualNetworkName)
	}
	if c.VirtualNetworkResourceGroupName != "" {
		t.Errorf("Expected Config to leave virtual_network_resource_group_name to '', but got %q", c.VirtualNetworkResourceGroupName)
	}
	if c.VirtualNetworkSubnetName != "" {
		t.Errorf("Expected Config to leave virtual_network_subnet_name to '', but got %q", c.VirtualNetworkSubnetName)
	}
}

// The user can pass the value virtual_network_resource_group_name to avoid the lookup of
// a virtual network's resource group, or to help with disambiguation.  The value should
// only be set if virtual_network_name was set.
func TestConfigVirtualNetworkResourceGroupNameMustBeSetWithVirtualNetworkName(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":                 "ignore",
		"capture_container_name":              "ignore",
		"location":                            "ignore",
		"image_url":                           "ignore",
		"storage_account":                     "ignore",
		"resource_group_name":                 "ignore",
		"subscription_id":                     "ignore",
		"os_type":                             constants.Target_Linux,
		"communicator":                        "none",
		"virtual_network_resource_group_name": "MyVirtualNetworkRG",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Error("Expected Config to reject virtual_network_resource_group_name, if virtual_network_name is not set.")
	}
}

// The user can pass the value virtual_network_subnet_name to avoid the lookup of
// a virtual network subnet's name, or to help with disambiguation.  The value should
// only be set if virtual_network_name was set.
func TestConfigVirtualNetworkSubnetNameMustBeSetWithVirtualNetworkName(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":         "ignore",
		"capture_container_name":      "ignore",
		"location":                    "ignore",
		"image_url":                   "ignore",
		"storage_account":             "ignore",
		"resource_group_name":         "ignore",
		"subscription_id":             "ignore",
		"os_type":                     constants.Target_Linux,
		"communicator":                "none",
		"virtual_network_subnet_name": "MyVirtualNetworkRG",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Error("Expected Config to reject virtual_network_subnet_name, if virtual_network_name is not set.")
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
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
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
		c, _, err := newConfig(config, packerConfiguration)
		if err != nil {
			t.Fatal(err)
		}

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
		t.Error("Expected configuration creation to succeed, but it failed!\n")
		t.Fatalf(" -> %+v\n", builderValues)
	}

	// Take away a required element, and ensure construction fails.
	for _, v := range requiredConfigValues {
		originalValue := builderValues[v]
		delete(builderValues, v)

		_, _, err := newConfig(builderValues, getPackerConfiguration())
		if err == nil {
			t.Error("Expected configuration creation to fail, but it succeeded!\n")
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

func TestWinRMConfigShouldSetRoundTripDecorator(t *testing.T) {
	config := getArmBuilderConfiguration()
	config["communicator"] = "winrm"
	config["winrm_username"] = "username"
	config["winrm_password"] = "password"

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	if c.Comm.WinRMTransportDecorator == nil {
		t.Error("Expected WinRMTransportDecorator to be set, but it was nil")
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
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Linux,
		"communicator":           "none",
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
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"os_type":                constants.Target_Windows,
		"communicator":           "none",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("Expected test to fail, but it succeeded")
	}

	multiError, _ := err.(*packer.MultiError)
	if len(multiError.Errors) != 2 {
		t.Errorf("Expected to find 2 errors, but found %d errors", len(multiError.Errors))
	}

	if !strings.Contains(err.Error(), "client_id must be specified") {
		t.Error("Expected to find error for 'client_id must be specified")
	}
	if !strings.Contains(err.Error(), "client_secret must be specified") {
		t.Error("Expected to find error for 'client_secret must be specified")
	}
}

func TestConfigShouldRejectMalformedCaptureNamePrefix(t *testing.T) {
	config := map[string]string{
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"communicator":           "none",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	wellFormedCaptureNamePrefix := []string{
		"packer",
		"AbcdefghijklmnopqrstuvwX",
		"hypen-hypen",
		"0leading-number",
		"v1.core.local",
	}

	for _, x := range wellFormedCaptureNamePrefix {
		config["capture_name_prefix"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err != nil {
			t.Errorf("Expected test to pass, but it failed with the well-formed capture_name_prefix set to %q.", x)
		}
	}

	malformedCaptureNamePrefix := []string{
		"-leading-hypen",
		"trailing-hypen-",
		"trailing-period.",
		"_leading-underscore",
		"punc-!@#$%^&*()_+-=-punc",
		"There-are-too-many-characters-in-the-name-and-the-limit-is-twenty-four",
	}

	for _, x := range malformedCaptureNamePrefix {
		config["capture_name_prefix"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err == nil {
			t.Errorf("Expected test to fail, but it succeeded with the malformed capture_name_prefix set to %q.", x)
		}
	}
}

func TestConfigShouldRejectMalformedCaptureContainerName(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix": "ignore",
		"image_offer":         "ignore",
		"image_publisher":     "ignore",
		"image_sku":           "ignore",
		"location":            "ignore",
		"storage_account":     "ignore",
		"resource_group_name": "ignore",
		"subscription_id":     "ignore",
		"communicator":        "none",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	wellFormedCaptureContainerName := []string{
		"0leading",
		"aleading",
		"hype-hypen",
		"abcdefghijklmnopqrstuvwxyz0123456789-abcdefghijklmnopqrstuvwxyz", // 63 characters
	}

	for _, x := range wellFormedCaptureContainerName {
		config["capture_container_name"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err != nil {
			t.Errorf("Expected test to pass, but it failed with the well-formed capture_container_name set to %q.", x)
		}
	}

	malformedCaptureContainerName := []string{
		"No-Capitals",
		"double--hypens",
		"-leading-hypen",
		"trailing-hypen-",
		"punc-!@#$%^&*()_+-=-punc",
		"there-are-over-63-characters-in-this-string-and-that-is-a-bad-container-name",
	}

	for _, x := range malformedCaptureContainerName {
		config["capture_container_name"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err == nil {
			t.Errorf("Expected test to fail, but it succeeded with the malformed capture_container_name set to %q.", x)
		}
	}
}

func TestConfigShouldAcceptTags(t *testing.T) {
	config := map[string]interface{}{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"communicator":           "none",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
		"azure_tags": map[string]string{
			"tag01": "value01",
			"tag02": "value02",
		},
	}

	c, _, err := newConfig(config, getPackerConfiguration())

	if err != nil {
		t.Fatal(err)
	}

	if len(c.AzureTags) != 2 {
		t.Fatalf("expected to find 2 tags, but got %d", len(c.AzureTags))
	}

	if _, ok := c.AzureTags["tag01"]; !ok {
		t.Error("expected to find key=\"tag01\", but did not")
	}
	if _, ok := c.AzureTags["tag02"]; !ok {
		t.Error("expected to find key=\"tag02\", but did not")
	}

	value := c.AzureTags["tag01"]
	if *value != "value01" {
		t.Errorf("expected AzureTags[\"tag01\"] to have value \"value01\", but got %q", value)
	}

	value = c.AzureTags["tag02"]
	if *value != "value02" {
		t.Errorf("expected AzureTags[\"tag02\"] to have value \"value02\", but got %q", value)
	}
}

func TestConfigShouldRejectTagsInExcessOf15AcceptTags(t *testing.T) {
	tooManyTags := map[string]string{}
	for i := 0; i < 16; i++ {
		tooManyTags[fmt.Sprintf("tag%.2d", i)] = "ignored"
	}

	config := map[string]interface{}{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"communicator":           "none",
		// Does not matter for this test case, just pick one.
		"os_type":    constants.Target_Linux,
		"azure_tags": tooManyTags,
	}

	_, _, err := newConfig(config, getPackerConfiguration())

	if err == nil {
		t.Fatal("expected config to reject based on an excessive amount of tags (> 15)")
	}
}

func TestConfigShouldRejectExcessiveTagNameLength(t *testing.T) {
	nameTooLong := make([]byte, 513)
	for i := range nameTooLong {
		nameTooLong[i] = 'a'
	}

	tags := map[string]string{}
	tags[string(nameTooLong)] = "ignored"

	config := map[string]interface{}{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"communicator":           "none",
		// Does not matter for this test case, just pick one.
		"os_type":    constants.Target_Linux,
		"azure_tags": tags,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject tag name based on length (> 512)")
	}
}

func TestConfigShouldRejectExcessiveTagValueLength(t *testing.T) {
	valueTooLong := make([]byte, 257)
	for i := range valueTooLong {
		valueTooLong[i] = 'a'
	}

	tags := map[string]string{}
	tags["tag01"] = string(valueTooLong)

	config := map[string]interface{}{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"image_offer":            "ignore",
		"image_publisher":        "ignore",
		"image_sku":              "ignore",
		"location":               "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"communicator":           "none",
		// Does not matter for this test case, just pick one.
		"os_type":    constants.Target_Linux,
		"azure_tags": tags,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject tag value based on length (> 256)")
	}
}

func getArmBuilderConfiguration() map[string]string {
	m := make(map[string]string)
	for _, v := range requiredConfigValues {
		m[v] = "ignored00"
	}

	m["communicator"] = "none"
	m["os_type"] = constants.Target_Linux
	return m
}

func getArmBuilderConfigurationWithWindows() map[string]string {
	m := make(map[string]string)
	for _, v := range requiredConfigValues {
		m[v] = "ignored00"
	}

	m["object_id"] = "ignored00"
	m["tenant_id"] = "ignored00"
	m["winrm_username"] = "ignored00"
	m["communicator"] = "winrm"
	m["os_type"] = constants.Target_Windows
	return m
}

func getPackerConfiguration() interface{} {
	config := map[string]interface{}{
		"packer_build_name":    "azure-arm-vm",
		"packer_builder_type":  "azure-arm-vm",
		"packer_debug":         "false",
		"packer_force":         "false",
		"packer_template_path": "/home/jenkins/azure-arm-vm/template.json",
	}

	return config
}

func getPackerCommunicatorConfiguration() map[string]string {
	config := map[string]string{
		"ssh_timeout":   "1h",
		"winrm_timeout": "2h",
	}

	return config
}
