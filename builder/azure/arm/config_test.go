package arm

import (
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/hashicorp/packer/builder/azure/common/constants"
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

	if c.managedImageStorageAccountType == "" {
		t.Errorf("Expected 'managedImageStorageAccountType' to be populated, but it was empty!")
	}

	if c.diskCachingType == "" {
		t.Errorf("Expected 'diskCachingType' to be populated, but it was empty!")
	}
}

func TestConfigShouldBeAbleToOverrideDefaultedValues(t *testing.T) {
	builderValues := getArmBuilderConfiguration()
	builderValues["ssh_password"] = "override_password"
	builderValues["ssh_username"] = "override_username"
	builderValues["vm_size"] = "override_vm_size"
	builderValues["communicator"] = "ssh"
	builderValues["managed_image_storage_account_type"] = "Premium_LRS"
	builderValues["disk_caching_type"] = "None"

	c, _, err := newConfig(builderValues, getPackerConfiguration())

	if err != nil {
		t.Fatalf("newConfig failed: %s", err)
	}

	if c.Password != "override_password" {
		t.Errorf("Expected 'Password' to be set to 'override_password', but found %q!", c.Password)
	}

	if c.Comm.SSHPassword != "override_password" {
		t.Errorf("Expected 'c.Comm.SSHPassword' to be set to 'override_password', but found %q!", c.Comm.SSHPassword)
	}

	if c.UserName != "override_username" {
		t.Errorf("Expected 'UserName' to be set to 'override_username', but found %q!", c.UserName)
	}

	if c.Comm.SSHUsername != "override_username" {
		t.Errorf("Expected 'c.Comm.SSHUsername' to be set to 'override_username', but found %q!", c.Comm.SSHUsername)
	}

	if c.VMSize != "override_vm_size" {
		t.Errorf("Expected 'vm_size' to be set to 'override_vm_size', but found %q!", c.VMSize)
	}

	if c.managedImageStorageAccountType != compute.StorageAccountTypesPremiumLRS {
		t.Errorf("Expected 'managed_image_storage_account_type' to be set to 'Premium_LRS', but found %q!", c.managedImageStorageAccountType)
	}

	if c.diskCachingType != compute.CachingTypesNone {
		t.Errorf("Expected 'disk_caching_type' to be set to 'None', but found %q!", c.diskCachingType)
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

func TestConfigShouldNormalizeOSTypeCase(t *testing.T) {
	config := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"subscription_id":        "ignore",
		"communicator":           "none",
	}

	os_types := map[string][]string{
		constants.Target_Linux:   {"linux", "LiNuX"},
		constants.Target_Windows: {"windows", "WiNdOWs"},
	}

	for k, v := range os_types {
		for _, os_type := range v {
			config["os_type"] = os_type
			c, _, err := newConfig(config, getPackerConfiguration())
			if err != nil {
				t.Fatalf("Expected config to accept the value %q, but it did not", os_type)
			}

			if c.OSType != k {
				t.Fatalf("Expected config to normalize the value %q to %q, but it did not", os_type, constants.Target_Linux)
			}
		}
	}

	bad_os_types := []string{"", "does-not-exist"}
	for _, os_type := range bad_os_types {
		config["os_type"] = os_type
		_, _, err := newConfig(config, getPackerConfiguration())
		if err == nil {
			t.Fatalf("Expected config to not accept the value %q, but it did", os_type)
		}
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

func TestConfigAllowedInboundIpAddressesIsOptional(t *testing.T) {
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

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	if c.AllowedInboundIpAddresses != nil {
		t.Errorf("Expected Config to set allowed_inbound_ip_addresses to nil, but got %v", c.AllowedInboundIpAddresses)
	}
}

func TestConfigShouldAcceptCorrectInboundIpAddresses(t *testing.T) {
	ipValue0 := "127.0.0.1"
	ipValue1 := "127.0.0.2"
	cidrValue2 := "192.168.100.0/24"
	cidrValue3 := "10.10.1.16/32"
	config := map[string]interface{}{
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

	config["allowed_inbound_ip_addresses"] = ipValue0
	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	if c.AllowedInboundIpAddresses == nil || len(c.AllowedInboundIpAddresses) != 1 ||
		c.AllowedInboundIpAddresses[0] != ipValue0 {
		t.Errorf("Expected 'allowed_inbound_ip_addresses' to have one element (%s), but got '%v'.", ipValue0, c.AllowedInboundIpAddresses)
	}

	config["allowed_inbound_ip_addresses"] = cidrValue2
	c, _, err = newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	if c.AllowedInboundIpAddresses == nil || len(c.AllowedInboundIpAddresses) != 1 ||
		c.AllowedInboundIpAddresses[0] != cidrValue2 {
		t.Errorf("Expected 'allowed_inbound_ip_addresses' to have one element (%s), but got '%v'.", cidrValue2, c.AllowedInboundIpAddresses)
	}

	config["allowed_inbound_ip_addresses"] = []string{ipValue0, cidrValue2, ipValue1, cidrValue3}
	c, _, err = newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	if c.AllowedInboundIpAddresses == nil || len(c.AllowedInboundIpAddresses) != 4 ||
		c.AllowedInboundIpAddresses[0] != ipValue0 || c.AllowedInboundIpAddresses[1] != cidrValue2 ||
		c.AllowedInboundIpAddresses[2] != ipValue1 || c.AllowedInboundIpAddresses[3] != cidrValue3 {
		t.Errorf("Expected 'allowed_inbound_ip_addresses' to have four elements (%s %s %s %s), but got '%v'.", ipValue0, cidrValue2, ipValue1,
			cidrValue3, c.AllowedInboundIpAddresses)
	}
}

func TestConfigShouldRejectIncorrectInboundIpAddresses(t *testing.T) {
	config := map[string]interface{}{
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

	config["allowed_inbound_ip_addresses"] = []string{"127.0.0.1", "127.0.0.two"}
	c, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Errorf("Expected configuration creation to fail, but it succeeded with the malformed allowed_inbound_ip_addresses set to %v", c.AllowedInboundIpAddresses)
	}

	config["allowed_inbound_ip_addresses"] = []string{"192.168.100.1000/24", "10.10.1.16/32"}
	c, _, err = newConfig(config, getPackerConfiguration())
	if err == nil {
		// 192.168.100.1000/24 is invalid
		t.Errorf("Expected configuration creation to fail, but it succeeded with the malformed allowed_inbound_ip_addresses set to %v", c.AllowedInboundIpAddresses)
	}
}

func TestConfigShouldRejectInboundIpAddressesWithVirtualNetwork(t *testing.T) {
	config := map[string]interface{}{
		"capture_name_prefix":          "ignore",
		"capture_container_name":       "ignore",
		"location":                     "ignore",
		"image_url":                    "ignore",
		"storage_account":              "ignore",
		"resource_group_name":          "ignore",
		"subscription_id":              "ignore",
		"os_type":                      constants.Target_Linux,
		"communicator":                 "none",
		"allowed_inbound_ip_addresses": "127.0.0.1",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	config["virtual_network_name"] = "some_vnet_name"
	_, _, err = newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Errorf("Expected configuration creation to fail, but it succeeded with allowed_inbound_ip_addresses and virtual_network_name both specified")
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
		"hyphen-hyphen",
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
		"-leading-hyphen",
		"trailing-hyphen-",
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
		"hype-hyphen",
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
		"double--hyphens",
		"-leading-hyphen",
		"trailing-hyphen-",
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

func TestConfigShouldRejectMalformedManagedImageOSDiskSnapshotName(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                         "ignore",
		"image_publisher":                     "ignore",
		"image_sku":                           "ignore",
		"location":                            "ignore",
		"subscription_id":                     "ignore",
		"communicator":                        "none",
		"managed_image_resource_group_name":   "ignore",
		"managed_image_name":                  "ignore",
		"managed_image_os_disk_snapshot_name": "ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	wellFormedManagedImageOSDiskSnapshotName := []string{
		"AbcdefghijklmnopqrstuvwX",
		"underscore_underscore",
		"0leading_number",
		"really_loooooooooooooooooooooooooooooooooooooooooooooooooong",
	}

	for _, x := range wellFormedManagedImageOSDiskSnapshotName {
		config["managed_image_os_disk_snapshot_name"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err != nil {
			t.Errorf("Expected test to pass, but it failed with the well-formed managed_image_os_disk_snapshot_name set to %q.", x)
		}
	}

	malformedManagedImageOSDiskSnapshotName := []string{
		"-leading-hyphen",
		"trailing-hyphen-",
		"trailing-period.",
		"punc-!@#$%^&*()_+-=-punc",
		"really_looooooooooooooooooooooooooooooooooooooooooooooooooooooong_exceeding_80_char_limit",
	}

	for _, x := range malformedManagedImageOSDiskSnapshotName {
		config["managed_image_os_disk_snapshot_name"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err == nil {
			t.Errorf("Expected test to fail, but it succeeded with the malformed managed_image_os_disk_snapshot_name set to %q.", x)
		}
	}
}

func TestConfigShouldRejectMalformedManagedImageDataDiskSnapshotPrefix(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"managed_image_resource_group_name": "ignore",
		"managed_image_name":                "ignore",
		"managed_image_data_disk_snapshot_prefix": "ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	wellFormedManagedImageDataDiskSnapshotPrefix := []string{
		"min_ten_chars",
		"AbcdefghijklmnopqrstuvwX",
		"underscore_underscore",
		"0leading_number",
		"less_than_sixty_characters",
	}

	for _, x := range wellFormedManagedImageDataDiskSnapshotPrefix {
		config["managed_image_data_disk_snapshot_prefix"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err != nil {
			t.Errorf("Expected test to pass, but it failed with the well-formed managed_image_data_disk_snapshot_prefix set to %q.", x)
		}
	}

	malformedManagedImageDataDiskSnapshotPrefix := []string{
		"-leading-hyphen",
		"trailing-hyphen-",
		"trailing-period.",
		"punc-!@#$%^&*()_+-=-punc",
		"really_looooooooooooooooooooooooooooooooooooooooooooooooooooooong_exceeding_60_char_limit",
	}

	for _, x := range malformedManagedImageDataDiskSnapshotPrefix {
		config["managed_image_data_disk_snapshot_prefix"] = x
		_, _, err := newConfig(config, getPackerConfiguration())

		if err == nil {
			t.Errorf("Expected test to fail, but it succeeded with the malformed managed_image_data_disk_snapshot_prefix set to %q.", x)
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
		t.Errorf("expected AzureTags[\"tag01\"] to have value \"value01\", but got %q", *value)
	}

	value = c.AzureTags["tag02"]
	if *value != "value02" {
		t.Errorf("expected AzureTags[\"tag02\"] to have value \"value02\", but got %q", *value)
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

func TestConfigZoneResilientShouldDefaultToFalse(t *testing.T) {
	config := map[string]interface{}{
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
		"build_resource_group_name":         "ignore",
		"image_publisher":                   "igore",
		"image_offer":                       "ignore",
		"image_sku":                         "ignore",
		"os_type":                           "linux",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	p := c.toImageParameters()
	if *p.ImageProperties.StorageProfile.ZoneResilient {
		t.Fatal("expected zone resilient default to be false")
	}
}

func TestConfigZoneResilientSetFromConfig(t *testing.T) {
	config := map[string]interface{}{
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
		"build_resource_group_name":         "ignore",
		"image_publisher":                   "igore",
		"image_offer":                       "ignore",
		"image_sku":                         "ignore",
		"os_type":                           "linux",
		"managed_image_zone_resilient":      true,
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	p := c.toImageParameters()
	if *p.ImageProperties.StorageProfile.ZoneResilient == false {
		t.Fatal("expected managed image zone resilient to be true from config")
	}
}

func TestConfigShouldRejectMissingCustomDataFile(t *testing.T) {
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
		"os_type":          constants.Target_Linux,
		"custom_data_file": "/this/file/does/not/exist",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject missing custom data file")
	}
}

func TestConfigShouldRejectManagedImageOSDiskSnapshotNameWithoutManagedImageName(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                         "ignore",
		"image_publisher":                     "ignore",
		"image_sku":                           "ignore",
		"location":                            "ignore",
		"subscription_id":                     "ignore",
		"communicator":                        "none",
		"managed_image_resource_group_name":   "ignore",
		"managed_image_os_disk_snapshot_name": "ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject Managed Image build with OS disk snapshot name but without managed image name")
	}
}

func TestConfigShouldRejectManagedImageOSDiskSnapshotNameWithoutManagedImageResourceGroupName(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                         "ignore",
		"image_publisher":                     "ignore",
		"image_sku":                           "ignore",
		"location":                            "ignore",
		"subscription_id":                     "ignore",
		"communicator":                        "none",
		"managed_image_name":                  "ignore",
		"managed_image_os_disk_snapshot_name": "ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject Managed Image build with OS disk snapshot name but without managed image resource group name")
	}
}

func TestConfigShouldRejectImageDataDiskSnapshotPrefixWithoutManagedImageName(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"managed_image_resource_group_name": "ignore",
		"managed_image_data_disk_snapshot_prefix": "ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject Managed Image build with data disk snapshot prefix but without managed image name")
	}
}

func TestConfigShouldRejectImageDataDiskSnapshotPrefixWithoutManagedImageResourceGroupName(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":        "ignore",
		"image_publisher":    "ignore",
		"image_sku":          "ignore",
		"location":           "ignore",
		"subscription_id":    "ignore",
		"communicator":       "none",
		"managed_image_name": "ignore",
		"managed_image_data_disk_snapshot_prefix": "ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject Managed Image build with data disk snapshot prefix but without managed image resource group name")
	}
}

func TestConfigShouldAcceptManagedImageOSDiskSnapshotNameAndManagedImageDataDiskSnapshotPrefix(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                             "ignore",
		"image_publisher":                         "ignore",
		"image_sku":                               "ignore",
		"location":                                "ignore",
		"subscription_id":                         "ignore",
		"communicator":                            "none",
		"managed_image_resource_group_name":       "ignore",
		"managed_image_name":                      "ignore",
		"managed_image_os_disk_snapshot_name":     "ignore_ignore",
		"managed_image_data_disk_snapshot_prefix": "ignore_ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal("expected config to accept platform managed image build")
	}
}

func TestConfigShouldRejectManagedImageOSDiskSnapshotNameAndManagedImageDataDiskSnapshotPrefixWithCaptureContainerName(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                         "ignore",
		"image_publisher":                     "ignore",
		"image_sku":                           "ignore",
		"location":                            "ignore",
		"subscription_id":                     "ignore",
		"communicator":                        "none",
		"capture_container_name":              "ignore",
		"managed_image_os_disk_snapshot_name": "ignore_ignore",
		"managed_image_data_disk_snapshot_prefix": "ignore_ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject Managed Image build with data disk snapshot prefix and OS disk snapshot name with capture container name")
	}
}

func TestConfigShouldRejectManagedImageOSDiskSnapshotNameAndManagedImageDataDiskSnapshotPrefixWithCaptureNamePrefix(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                         "ignore",
		"image_publisher":                     "ignore",
		"image_sku":                           "ignore",
		"location":                            "ignore",
		"subscription_id":                     "ignore",
		"communicator":                        "none",
		"capture_name_prefix":                 "ignore",
		"managed_image_os_disk_snapshot_name": "ignore_ignore",
		"managed_image_data_disk_snapshot_prefix": "ignore_ignore",
		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject Managed Image build with data disk snapshot prefix and OS disk snapshot name with capture name prefix")
	}
}

func TestConfigShouldAcceptPlatformManagedImageBuild(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"managed_image_resource_group_name": "ignore",
		"managed_image_name":                "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatal("expected config to accept platform managed image build")
	}
}

// If the user specified a build for a VHD and a Managed Image it should be rejected.
func TestConfigShouldRejectVhdAndManagedImageOutput(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"capture_container_name":            "ignore",
		"capture_name_prefix":               "ignore",
		"managed_image_resource_group_name": "ignore",
		"managed_image_name":                "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject VHD and Managed Image build")
	}
}

// If the user specified a build of a VHD, but started with a managed image it should be rejected.
func TestConfigShouldRejectManagedImageSourceAndVhdOutput(t *testing.T) {
	config := map[string]interface{}{
		"image_url":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"managed_image_resource_group_name": "ignore",
		"managed_image_name":                "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject VHD and Managed Image build")
	}
}

func TestConfigShouldRejectCustomAndPlatformManagedImageBuild(t *testing.T) {
	config := map[string]interface{}{
		"custom_managed_image_resource_group_name": "ignore",
		"custom_managed_image_name":                "ignore",
		"image_offer":                              "ignore",
		"image_publisher":                          "ignore",
		"image_sku":                                "ignore",
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"communicator":                             "none",
		"managed_image_resource_group_name":        "ignore",
		"managed_image_name":                       "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject custom and platform input for a managed image build")
	}
}

func TestConfigShouldRejectCustomAndImageUrlForManagedImageBuild(t *testing.T) {
	config := map[string]interface{}{
		"image_url": "ignore",
		"custom_managed_image_resource_group_name": "ignore",
		"custom_managed_image_name":                "ignore",
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"communicator":                             "none",
		"managed_image_resource_group_name":        "ignore",
		"managed_image_name":                       "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject custom and platform input for a managed image build")
	}
}

func TestConfigShouldRejectMalformedManageImageStorageAccountTypes(t *testing.T) {
	config := map[string]interface{}{
		"custom_managed_image_resource_group_name": "ignore",
		"custom_managed_image_name":                "ignore",
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"communicator":                             "none",
		"managed_image_resource_group_name":        "ignore",
		"managed_image_name":                       "ignore",
		"managed_image_storage_account_type":       "--invalid--",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject custom and platform input for a managed image build")
	}
}

func TestConfigShouldRejectMalformedDiskCachingType(t *testing.T) {
	config := map[string]interface{}{
		"custom_managed_image_resource_group_name": "ignore",
		"custom_managed_image_name":                "ignore",
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"communicator":                             "none",
		"managed_image_resource_group_name":        "ignore",
		"managed_image_name":                       "ignore",
		"disk_caching_type":                        "--invalid--",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject custom and platform input for a managed image build")
	}
}

func TestConfigShouldAcceptManagedImageStorageAccountTypes(t *testing.T) {
	config := map[string]interface{}{
		"custom_managed_image_resource_group_name": "ignore",
		"custom_managed_image_name":                "ignore",
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"communicator":                             "none",
		"managed_image_resource_group_name":        "ignore",
		"managed_image_name":                       "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	storage_account_types := []string{"Premium_LRS", "Standard_LRS"}

	for _, x := range storage_account_types {
		config["managed_image_storage_account_type"] = x
		_, _, err := newConfig(config, getPackerConfiguration())
		if err != nil {
			t.Fatalf("expected config to accept a managed_image_storage_account_type of %q", x)
		}
	}
}

func TestConfigShouldAcceptDiskCachingTypes(t *testing.T) {
	config := map[string]interface{}{
		"custom_managed_image_resource_group_name": "ignore",
		"custom_managed_image_name":                "ignore",
		"location":                                 "ignore",
		"subscription_id":                          "ignore",
		"communicator":                             "none",
		"managed_image_resource_group_name":        "ignore",
		"managed_image_name":                       "ignore",

		// Does not matter for this test case, just pick one.
		"os_type": constants.Target_Linux,
	}

	storage_account_types := []string{"None", "ReadOnly", "ReadWrite"}

	for _, x := range storage_account_types {
		config["disk_caching_type"] = x
		_, _, err := newConfig(config, getPackerConfiguration())
		if err != nil {
			t.Fatalf("expected config to accept a disk_caching_type of %q", x)
		}
	}
}

func TestConfigShouldRejectTempAndBuildResourceGroupName(t *testing.T) {
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

		// custom may define one or the other, but not both
		"temp_resource_group_name":  "rgn00",
		"build_resource_group_name": "rgn00",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject the use of both temp_resource_group_name and build_resource_group_name")
	}
}

func TestConfigShouldRejectInvalidResourceGroupNames(t *testing.T) {
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
		"os_type":                "linux",
	}

	tests := []struct {
		name string
		ok   bool
	}{
		// The Good
		{"packer-Resource-Group-jt2j3fc", true},
		{"My", true},
		{"My-(with-parens)-Resource-Group", true},

		// The Bad
		{"My Resource Group", false},
		{"My-Resource-Group-", false},
		{"My.Resource.Group.", false},

		// The Ugly
		{"My!@#!@#%$%yM", false},
		{"   ", false},
		{"My10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false},
	}

	settings := []string{"temp_resource_group_name", "build_resource_group_name"}

	for _, x := range settings {
		for _, y := range tests {
			config[x] = y.name

			_, _, err := newConfig(config, getPackerConfiguration())
			if !y.ok && err == nil {
				t.Errorf("expected config to reject %q for setting %q", y.name, x)
			} else if y.ok && err != nil {
				t.Errorf("expected config to accept %q for setting %q", y.name, x)
			}
		}

		delete(config, "location") // not valid for build_resource_group_name
		delete(config, x)
	}
}

func TestConfigShouldRejectManagedDiskNames(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"os_type":                           "linux",
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
	}

	testsResourceGroupNames := []struct {
		name string
		ok   bool
	}{
		// The Good
		{"packer-Resource-Group-jt2j3fc", true},
		{"My", true},
		{"My-(with-parens)-Resource-Group", true},

		// The Bad
		{"My Resource Group", false},
		{"My-Resource-Group-", false},
		{"My.Resource.Group.", false},

		// The Ugly
		{"My!@#!@#%$%yM", false},
		{"   ", false},
		{"My10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false},
	}

	settingUnderTest := "managed_image_resource_group_name"
	for _, y := range testsResourceGroupNames {
		config[settingUnderTest] = y.name

		_, _, err := newConfig(config, getPackerConfiguration())
		if !y.ok && err == nil {
			t.Errorf("expected config to reject %q for setting %q", y.name, settingUnderTest)
		} else if y.ok && err != nil {
			t.Errorf("expected config to accept %q for setting %q", y.name, settingUnderTest)
		}
	}

	config["managed_image_resource_group_name"] = "ignored"

	testNames := []struct {
		name string
		ok   bool
	}{
		// The Good
		{"ManagedDiskName", true},
		{"Managed-Disk-Name", true},
		{"My33", true},

		// The Bad
		{"Managed Disk Name", false},
		{"Managed-Disk-Name-", false},
		{"Managed.Disk.Name.", false},

		// The Ugly
		{"My!@#!@#%$%yM", false},
		{"   ", false},
		{"My10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false},
	}

	settingUnderTest = "managed_image_name"
	for _, y := range testNames {
		config[settingUnderTest] = y.name

		_, _, err := newConfig(config, getPackerConfiguration())
		if !y.ok && err == nil {
			t.Logf("expected config to reject %q for setting %q", y.name, settingUnderTest)
		} else if y.ok && err != nil {
			t.Logf("expected config to accept %q for setting %q", y.name, settingUnderTest)
		}
	}
}

func TestConfigAdditionalDiskDefaultIsNil(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	if c.AdditionalDiskSize != nil {
		t.Errorf("Expected Config to not have a set of additional disks, but got a non nil value")
	}
}

func TestConfigAdditionalDiskOverrideDefault(t *testing.T) {
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

	diskconfig := map[string][]int32{
		"disk_additional_size": {32, 64},
	}

	c, _, _ := newConfig(config, diskconfig, getPackerConfiguration())
	if c.AdditionalDiskSize == nil {
		t.Errorf("Expected Config to have a set of additional disks, but got nil")
	}
	if len(c.AdditionalDiskSize) != 2 {
		t.Errorf("Expected Config to have a 2 additional disks, but got %d additional disks", len(c.AdditionalDiskSize))
	}
	if c.AdditionalDiskSize[0] != 32 {
		t.Errorf("Expected Config to have the first additional disks of size 32Gb, but got %dGb", c.AdditionalDiskSize[0])
	}
	if c.AdditionalDiskSize[1] != 64 {
		t.Errorf("Expected Config to have the second additional disks of size 64Gb, but got %dGb", c.AdditionalDiskSize[1])
	}
}

// Test that configuration handles plan info
//
// The use of plan info requires that the following three properties are set.
//
//  1. plan_name
//  2. plan_product
//  3. plan_publisher
func TestPlanInfoConfiguration(t *testing.T) {
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
		"os_type":                "linux",
		"communicator":           "none",
	}

	planInfo := map[string]string{
		"plan_name": "--plan-name--",
	}
	config["plan_info"] = planInfo

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject the use of plan_name without plan_product and plan_publisher")
	}

	planInfo["plan_product"] = "--plan-product--"
	_, _, err = newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject the use of plan_name and plan_product without plan_publisher")
	}

	planInfo["plan_publisher"] = "--plan-publisher--"
	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatalf("expected config to accept a complete plan configuration: %s", err)
	}

	if c.PlanInfo.PlanName != "--plan-name--" {
		t.Fatalf("Expected PlanName to be '--plan-name--', but got %q", c.PlanInfo.PlanName)
	}
	if c.PlanInfo.PlanProduct != "--plan-product--" {
		t.Fatalf("Expected PlanProduct to be '--plan-product--', but got %q", c.PlanInfo.PlanProduct)
	}
	if c.PlanInfo.PlanPublisher != "--plan-publisher--" {
		t.Fatalf("Expected PlanPublisher to be '--plan-publisher--, but got %q", c.PlanInfo.PlanPublisher)
	}
}

func TestPlanInfoPromotionCode(t *testing.T) {
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
		"os_type":                "linux",
		"communicator":           "none",
		"plan_info": map[string]string{
			"plan_name":           "--plan-name--",
			"plan_product":        "--plan-product--",
			"plan_publisher":      "--plan-publisher--",
			"plan_promotion_code": "--plan-promotion-code--",
		},
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Fatalf("expected config to accept plan_info configuration, but got %s", err)
	}

	if c.PlanInfo.PlanName != "--plan-name--" {
		t.Fatalf("Expected PlanName to be '--plan-name--', but got %q", c.PlanInfo.PlanName)
	}
	if c.PlanInfo.PlanProduct != "--plan-product--" {
		t.Fatalf("Expected PlanProduct to be '--plan-product--', but got %q", c.PlanInfo.PlanProduct)
	}
	if c.PlanInfo.PlanPublisher != "--plan-publisher--" {
		t.Fatalf("Expected PlanPublisher to be '--plan-publisher--, but got %q", c.PlanInfo.PlanPublisher)
	}
	if c.PlanInfo.PlanPromotionCode != "--plan-promotion-code--" {
		t.Fatalf("Expected PlanPublisher to be '--plan-promotion-code----, but got %q", c.PlanInfo.PlanPromotionCode)
	}
}

// plan_info defines 3 or 4 tags based on plan data.
// The user can define up to 15 tags.  If the combination of these two
// exceeds the max tag amount, the builder should reject the configuration.
func TestPlanInfoTooManyTagsErrors(t *testing.T) {
	exactMaxNumberOfTags := map[string]string{}
	for i := 0; i < 15; i++ {
		exactMaxNumberOfTags[fmt.Sprintf("tag%.2d", i)] = "ignored"
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
		"os_type":                "linux",
		"communicator":           "none",
		"azure_tags":             exactMaxNumberOfTags,
		"plan_info": map[string]string{
			"plan_name":           "--plan-name--",
			"plan_product":        "--plan-product--",
			"plan_publisher":      "--plan-publisher--",
			"plan_promotion_code": "--plan-promotion-code--",
		},
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Fatal("expected config to reject configuration due to excess tags")
	}
}

// The Azure builder creates temporary resources, but the user has some control over
// these values. This test asserts those values are controllable by the user.
func TestConfigShouldAllowTempNameOverrides(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"os_type":                           "linux",
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
		"temp_resource_group_name":          "myTempResourceGroupName",
		"temp_compute_name":                 "myTempComputeName",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Errorf("newConfig failed with %q", err)
	}

	if c.TempResourceGroupName != "myTempResourceGroupName" {
		t.Errorf("expected TempResourceGroupName to be %q, but got %q", "myTempResourceGroupName", c.TempResourceGroupName)
	}
	if c.tmpResourceGroupName != "myTempResourceGroupName" {
		t.Errorf("expected tmpResourceGroupName to be %q, but got %q", "myTempResourceGroupName", c.tmpResourceGroupName)
	}

	if c.TempComputeName != "myTempComputeName" {
		t.Errorf("expected TempComputeName to be %q, but got %q", "myTempComputeName", c.TempComputeName)
	}
	if c.tmpComputeName != "myTempComputeName" {
		t.Errorf("expected tmpComputeName to be %q, but got %q", "myTempComputeName", c.tmpResourceGroupName)
	}
}

func TestConfigShouldAllowAsyncResourceGroupOverride(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"os_type":                           "linux",
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
		"async_resourcegroup_delete":        "true",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Errorf("newConfig failed with %q", err)
	}

	if c.AsyncResourceGroupDelete != true {
		t.Errorf("expected async_resourcegroup_delete to be %q, but got %t", "async_resourcegroup_delete", c.AsyncResourceGroupDelete)
	}
}
func TestConfigShouldAllowAsyncResourceGroupOverrideNoValue(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"os_type":                           "linux",
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Errorf("newConfig failed with %q", err)
	}

	if c.AsyncResourceGroupDelete != false {
		t.Errorf("expected async_resourcegroup_delete to be %q, but got %t", "async_resourcegroup_delete", c.AsyncResourceGroupDelete)
	}
}
func TestConfigShouldAllowAsyncResourceGroupOverrideBadValue(t *testing.T) {
	config := map[string]interface{}{
		"image_offer":                       "ignore",
		"image_publisher":                   "ignore",
		"image_sku":                         "ignore",
		"location":                          "ignore",
		"subscription_id":                   "ignore",
		"communicator":                      "none",
		"os_type":                           "linux",
		"managed_image_name":                "ignore",
		"managed_image_resource_group_name": "ignore",
		"async_resourcegroup_delete":        "asdasda",
	}

	c, _, err := newConfig(config, getPackerConfiguration())
	if err != nil && c == nil {
		t.Log("newConfig failed  which is expected ", err)

	}

}
func TestConfigShouldAllowSharedImageGalleryOptions(t *testing.T) {
	config := map[string]interface{}{
		"location":        "ignore",
		"subscription_id": "ignore",
		"os_type":         "linux",
		"shared_image_gallery": map[string]string{
			"subscription":   "ignore",
			"resource_group": "ignore",
			"gallery_name":   "ignore",
			"image_name":     "ignore",
			"image_version":  "ignore",
		},
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err == nil {
		t.Log("expected config to accept Shared Image Gallery options", err)
	}

}

func TestConfigShouldRejectSharedImageGalleryWithVhdTarget(t *testing.T) {
	config := map[string]interface{}{
		"location":        "ignore",
		"subscription_id": "ignore",
		"os_type":         "linux",
		"shared_image_gallery": map[string]string{
			"subscription":   "ignore",
			"resource_group": "ignore",
			"gallery_name":   "ignore",
			"image_name":     "ignore",
			"image_version":  "ignore",
		},
		"resource_group_name":    "ignore",
		"storage_account":        "ignore",
		"capture_container_name": "ignore",
		"capture_name_prefix":    "ignore",
	}

	_, _, err := newConfig(config, getPackerConfiguration())
	if err != nil {
		t.Log("expected an error if Shared Image Gallery source is used with VHD target", err)
	}
}

func Test_GivenZoneNotSupportingResiliency_ConfigValidate_ShouldWarn(t *testing.T) {
	builderValues := getArmBuilderConfiguration()
	builderValues["managed_image_zone_resilient"] = "true"
	builderValues["location"] = "ukwest"

	c, _, err := newConfig(builderValues, getPackerConfiguration())
	if err != nil {
		t.Errorf("newConfig failed with %q", err)
	}

	var m = ""
	c.validateLocationZoneResiliency(func(s string) { m = s })

	if m != "WARNING: Zone resiliency may not be supported in ukwest, checkout the docs at https://docs.microsoft.com/en-us/azure/availability-zones/" {
		t.Errorf("warning message not as expected: %s", m)
	}
}

func Test_GivenZoneSupportingResiliency_ConfigValidate_ShouldNotWarn(t *testing.T) {
	builderValues := getArmBuilderConfiguration()
	builderValues["managed_image_zone_resilient"] = "true"
	builderValues["location"] = "westeurope"

	c, _, err := newConfig(builderValues, getPackerConfiguration())
	if err != nil {
		t.Errorf("newConfig failed with %q", err)
	}

	var m = ""
	c.validateLocationZoneResiliency(func(s string) { m = s })

	if m != "" {
		t.Errorf("warning message not as expected: %s", m)
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
