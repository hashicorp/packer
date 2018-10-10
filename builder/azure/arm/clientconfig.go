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
	ClientID       string `mapstructure:"client_id"`
	ClientSecret   string `mapstructure:"client_secret"`
	ObjectID       string `mapstructure:"object_id"`
	TenantID       string `mapstructure:"tenant_id"`
	SubscriptionID string `mapstructure:"subscription_id"`
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
	cloudEnvironment azure.Environment,
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
		servicePrincipalToken, err = packerAzureCommon.Authenticate(cloudEnvironment, tenantID, say, cloudEnvironment.ServiceManagementEndpoint)
		if err != nil {
			return nil, nil, err
		}
		say("Getting token for Vault resource")
		servicePrincipalTokenVault, err = packerAzureCommon.Authenticate(cloudEnvironment, tenantID, say, strings.TrimRight(cloudEnvironment.KeyVaultEndpoint, "/"))
		if err != nil {
			return nil, nil, err
		}

	} else {
		auth := NewAuthenticate(cloudEnvironment, c.ClientID, c.ClientSecret, tenantID)

		servicePrincipalToken, err = auth.getServicePrincipalToken()
		if err != nil {
			return nil, nil, err
		}

		servicePrincipalTokenVault, err = auth.getServicePrincipalTokenWithResource(
			strings.TrimRight(cloudEnvironment.KeyVaultEndpoint, "/"))
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
