package arm

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	packerAzureCommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/packer"
)

// ClientConfig allows for various ways to authenticate Azure clients
type ClientConfig struct {
	// Describes where API's are

	CloudEnvironmentName string `mapstructure:"cloud_environment_name"`
	cloudEnvironment     *azure.Environment

	// Authentication fields

	ClientID       string `mapstructure:"client_id"`
	ClientSecret   string `mapstructure:"client_secret"`
	ObjectID       string `mapstructure:"object_id"`
	TenantID       string `mapstructure:"tenant_id"`
	SubscriptionID string `mapstructure:"subscription_id"`
}

const DefaultCloudEnvironmentName = "Public"

func (c *ClientConfig) provideDefaultValues() {
	if c.CloudEnvironmentName == "" {
		c.CloudEnvironmentName = DefaultCloudEnvironmentName
	}
}

func (c *ClientConfig) setCloudEnvironment() error {
	lookup := map[string]string{
		"CHINA":           "AzureChinaCloud",
		"CHINACLOUD":      "AzureChinaCloud",
		"AZURECHINACLOUD": "AzureChinaCloud",

		"GERMAN":           "AzureGermanCloud",
		"GERMANCLOUD":      "AzureGermanCloud",
		"AZUREGERMANCLOUD": "AzureGermanCloud",

		"GERMANY":           "AzureGermanCloud",
		"GERMANYCLOUD":      "AzureGermanCloud",
		"AZUREGERMANYCLOUD": "AzureGermanCloud",

		"PUBLIC":           "AzurePublicCloud",
		"PUBLICCLOUD":      "AzurePublicCloud",
		"AZUREPUBLICCLOUD": "AzurePublicCloud",

		"USGOVERNMENT":           "AzureUSGovernmentCloud",
		"USGOVERNMENTCLOUD":      "AzureUSGovernmentCloud",
		"AZUREUSGOVERNMENTCLOUD": "AzureUSGovernmentCloud",
	}

	name := strings.ToUpper(c.CloudEnvironmentName)
	envName, ok := lookup[name]
	if !ok {
		return fmt.Errorf("There is no cloud environment matching the name '%s'!", c.CloudEnvironmentName)
	}

	env, err := azure.EnvironmentFromName(envName)
	c.cloudEnvironment = &env
	return err
}

func (c ClientConfig) assertRequiredParametersSet(errs *packer.MultiError) {
	/////////////////////////////////////////////
	// Authentication via OAUTH

	// Check if device login is being asked for, and is allowed.
	//
	// Device login is enabled if the user only defines SubscriptionID and not
	// ClientID, ClientSecret, and TenantID.
	//
	// Device login is not enabled for Windows because the WinRM certificate is
	// readable by the ObjectID of the App.  There may be another way to handle
	// this case, but I am not currently aware of it - send feedback.

	if !c.useDeviceLogin() {
		if c.ClientID == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A client_id must be specified"))
		}

		if c.ClientSecret == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A client_secret must be specified"))
		}

		if c.SubscriptionID == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A subscription_id must be specified"))
		}
	}
}

func (c ClientConfig) useDeviceLogin() bool {
	return c.SubscriptionID != "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.TenantID == ""
}

func (c ClientConfig) getServicePrincipalTokens(
	say func(string)) (
	servicePrincipalToken *adal.ServicePrincipalToken,
	servicePrincipalTokenVault *adal.ServicePrincipalToken,
	err error) {

	tenantID := c.TenantID
	if tenantID == "" {
		tenantID = "common"
	}

	if c.useDeviceLogin() {
		say("Getting auth token for Service management endpoint")
		servicePrincipalToken, err = packerAzureCommon.Authenticate(*c.cloudEnvironment, tenantID, say, c.cloudEnvironment.ServiceManagementEndpoint)
		if err != nil {
			return nil, nil, err
		}
		say("Getting token for Vault resource")
		servicePrincipalTokenVault, err = packerAzureCommon.Authenticate(*c.cloudEnvironment, tenantID, say, strings.TrimRight(c.cloudEnvironment.KeyVaultEndpoint, "/"))
		if err != nil {
			return nil, nil, err
		}

	} else {
		auth := NewAuthenticate(*c.cloudEnvironment, c.ClientID, c.ClientSecret, tenantID)

		servicePrincipalToken, err = auth.getServicePrincipalToken()
		if err != nil {
			return nil, nil, err
		}

		servicePrincipalTokenVault, err = auth.getServicePrincipalTokenWithResource(
			strings.TrimRight(c.cloudEnvironment.KeyVaultEndpoint, "/"))
		if err != nil {
			return nil, nil, err
		}
	}

	err = servicePrincipalToken.EnsureFresh()

	if err != nil {
		return nil, nil, err
	}

	err = servicePrincipalTokenVault.EnsureFresh()

	if err != nil {
		return nil, nil, err
	}

	return servicePrincipalToken, servicePrincipalTokenVault, nil
}
