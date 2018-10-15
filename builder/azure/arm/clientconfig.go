package arm

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
	packerAzureCommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/packer"
)

// ClientConfig allows for various ways to authenticate Azure clients
type ClientConfig struct {
	// Describes where API's are

	CloudEnvironmentName string `mapstructure:"cloud_environment_name"`
	cloudEnvironment     *azure.Environment

	// Authentication fields

	// Client ID
	ClientID string `mapstructure:"client_id"`
	// Client secret/password
	ClientSecret string `mapstructure:"client_secret"`
	// JWT bearer token for client auth (RFC 7523, Sec. 2.2)
	ClientJWT      string `mapstructure:"client_jwt"`
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

	if c.SubscriptionID == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A subscription_id must be specified"))
	}

	if c.useDeviceLogin() {
		// nothing else to check
		return
	}

	if c.ClientID == "" ||
		(c.ClientSecret == "" && c.ClientJWT == "") ||
		(c.ClientSecret != "" && c.ClientJWT != "") {
		// either client ID was not set, or neither or both secret and JWT are set
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("No valid set of authention methods specified: \n"+
			"specify either (client_id,client_secret) or (client_id,client_jwt) to use a service principal, \n"+
			"or specify none of these to use interactive user authentication."))
	}

	if c.ClientJWT != "" {
		// should be a JWT that is valid for at least 5 more minutes
		p := jwt.Parser{}
		claims := jwt.StandardClaims{}
		token, _, err := p.ParseUnverified(c.ClientJWT, &claims)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("client_jwt is not a JWT: %v", err))
		} else {
			if claims.ExpiresAt < time.Now().Add(5*time.Minute).Unix() {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("client_jwt will expire within 5 minutes, please use a JWT that is valid for at least 5 minutes"))
			}
			if t, ok := token.Header["x5t"]; !ok || t == "" {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("client_jwt is missing the x5t header value, which is required for bearer JWT client authentication to Azure"))
			}
		}
	}
}

func (c ClientConfig) useDeviceLogin() bool {
	return c.SubscriptionID != "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.ClientJWT == "" &&
		c.TenantID == ""
}

func (c ClientConfig) getServicePrincipalTokens(
	say func(string)) (
	servicePrincipalToken *adal.ServicePrincipalToken,
	servicePrincipalTokenVault *adal.ServicePrincipalToken,
	err error) {

	tenantID := c.TenantID

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

	} else if c.ClientSecret != "" {
		say("Getting tokens using client secret")
		auth := NewSecretOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientSecret, tenantID)

		servicePrincipalToken, err = auth.getServicePrincipalToken()
		if err != nil {
			return nil, nil, err
		}

		servicePrincipalTokenVault, err = auth.getServicePrincipalTokenWithResource(
			strings.TrimRight(c.cloudEnvironment.KeyVaultEndpoint, "/"))
		if err != nil {
			return nil, nil, err
		}
	} else {
		say("Getting tokens using client bearer JWT")
		auth := NewJWTOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientJWT, tenantID)

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
