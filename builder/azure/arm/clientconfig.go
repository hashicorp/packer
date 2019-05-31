//go:generate struct-markdown

package arm

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/packer/packer"
)

// ClientConfig allows for various ways to authenticate Azure clients
type ClientConfig struct {
	// One of Public, China, Germany, or
    // USGovernment. Defaults to Public. Long forms such as
    // USGovernmentCloud and AzureUSGovernmentCloud are also supported.
	CloudEnvironmentName string `mapstructure:"cloud_environment_name" required:"false"`
	cloudEnvironment     *azure.Environment

	// Authentication fields

	// Client ID
	ClientID string `mapstructure:"client_id"`
	// Client secret/password
	ClientSecret string `mapstructure:"client_secret"`
	// Certificate path for client auth
	ClientCertPath string `mapstructure:"client_cert_path"`
	// JWT bearer token for client auth (RFC 7523, Sec. 2.2)
	ClientJWT      string `mapstructure:"client_jwt"`
	ObjectID       string `mapstructure:"object_id"`
	// The account identifier with which your client_id and
    // subscription_id are associated. If not specified, tenant_id will be
    // looked up using subscription_id.
	TenantID       string `mapstructure:"tenant_id" required:"false"`
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

	if c.useMSI() {
		return
	}

	if c.useDeviceLogin() {
		return
	}

	if c.SubscriptionID != "" && c.ClientID != "" &&
		c.ClientSecret != "" &&
		c.ClientCertPath == "" &&
		c.ClientJWT == "" {
		// Service principal using secret
		return
	}

	if c.SubscriptionID != "" && c.ClientID != "" &&
		c.ClientSecret == "" &&
		c.ClientCertPath != "" &&
		c.ClientJWT == "" {
		// Service principal using certificate

		if _, err := os.Stat(c.ClientCertPath); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("client_cert_path is not an accessible file: %v", err))
		}
		return
	}

	if c.SubscriptionID != "" && c.ClientID != "" &&
		c.ClientSecret == "" &&
		c.ClientCertPath == "" &&
		c.ClientJWT != "" {
		// Service principal using JWT
		// Check that JWT is valid for at least 5 more minutes

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

		return
	}

	errs = packer.MultiErrorAppend(errs, fmt.Errorf("No valid set of authentication values specified:\n"+
		"  to use the Managed Identity of the current machine, do not specify any of the fields below\n"+
		"  to use interactive user authentication, specify only subscription_id\n"+
		"  to use an Azure Active Directory service principal, specify either:\n"+
		"  - subscription_id, client_id and client_secret\n"+
		"  - subscription_id, client_id and client_cert_path\n"+
		"  - subscription_id, client_id and client_jwt."))
}

func (c ClientConfig) useDeviceLogin() bool {
	return c.SubscriptionID != "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.ClientJWT == "" &&
		c.ClientCertPath == ""
}

func (c ClientConfig) useMSI() bool {
	return c.SubscriptionID == "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.ClientJWT == "" &&
		c.ClientCertPath == "" &&
		c.TenantID == ""
}

func (c ClientConfig) getServicePrincipalTokens(
	say func(string)) (
	servicePrincipalToken *adal.ServicePrincipalToken,
	servicePrincipalTokenVault *adal.ServicePrincipalToken,
	err error) {

	tenantID := c.TenantID

	var auth oAuthTokenProvider

	if c.useDeviceLogin() {
		say("Getting tokens using device flow")
		auth = NewDeviceFlowOAuthTokenProvider(*c.cloudEnvironment, say, tenantID)
	} else if c.useMSI() {
		say("Getting tokens using Managed Identity for Azure")
		auth = NewMSIOAuthTokenProvider(*c.cloudEnvironment)
	} else if c.ClientSecret != "" {
		say("Getting tokens using client secret")
		auth = NewSecretOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientSecret, tenantID)
	} else if c.ClientCertPath != "" {
		say("Getting tokens using client certificate")
		auth, err = NewCertOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientCertPath, tenantID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		say("Getting tokens using client bearer JWT")
		auth = NewJWTOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientJWT, tenantID)
	}

	servicePrincipalToken, err = auth.getServicePrincipalToken()
	if err != nil {
		return nil, nil, err
	}

	err = servicePrincipalToken.EnsureFresh()
	if err != nil {
		return nil, nil, err
	}

	servicePrincipalTokenVault, err = auth.getServicePrincipalTokenWithResource(
		strings.TrimRight(c.cloudEnvironment.KeyVaultEndpoint, "/"))
	if err != nil {
		return nil, nil, err
	}

	err = servicePrincipalTokenVault.EnsureFresh()
	if err != nil {
		return nil, nil, err
	}

	return servicePrincipalToken, servicePrincipalTokenVault, nil
}
