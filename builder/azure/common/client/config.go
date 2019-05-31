//go:generate struct-markdown

package client

import (
	"fmt"
	"github.com/hashicorp/packer/builder/azure/common"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/packer/packer"
)

// Config allows for various ways to authenticate Azure clients.
// When `client_id` and `subscription_id` are specified, Packer will use the
// specified Azure Active Directoty (AAD) Service Principal (SP).
// If only `subscription_id` is specified, Packer will try to interactively
// log on the current user (tokens will be cached).
// If none of these options are specified, Packer will attempt to use the
// Managed Identity and subscription of the VM that Packer is running on.
// This will only work if Packer is running on an Azure VM.
type Config struct {
	// One of Public, China, Germany, or
	// USGovernment. Defaults to Public. Long forms such as
	// USGovernmentCloud and AzureUSGovernmentCloud are also supported.
	CloudEnvironmentName string `mapstructure:"cloud_environment_name" required:"false"`
	CloudEnvironment     *azure.Environment

	// Authentication fields

	// The application ID of the AAD Service Principal.
	// Requires either `client_secret`, `client_cert_path` or `client_jwt` to be set as well.
	ClientID string `mapstructure:"client_id"`
	// A password/secret registered for the AAD SP.
	ClientSecret string `mapstructure:"client_secret"`
	// The path to a certificate that will be used to authenticate as the specified AAD SP.
	ClientCertPath string `mapstructure:"client_cert_path"`
	// A JWT bearer token for client auth (RFC 7523, Sec. 2.2) that will be used
	// to authenticate the AAD SP. Provides more control over token the expiration
	// when using certificate authentication than when using `client_cert_path`.
	ClientJWT string `mapstructure:"client_jwt"`
	// The object ID for the AAD SP. Optional, will be derived from the oAuth token if left empty.
	ObjectID string `mapstructure:"object_id"`

	// The Active Directory tenant identifier with which your `client_id` and
	// `subscription_id` are associated. If not specified, `tenant_id` will be
	// looked up using `subscription_id`.
	TenantID string `mapstructure:"tenant_id" required:"false"`
	// The subscription to use.
	SubscriptionID string `mapstructure:"subscription_id"`

	authType string
}

const (
	authTypeDeviceLogin     = "DeviceLogin"
	authTypeMSI             = "ManagedIdentity"
	authTypeClientSecret    = "ClientSecret"
	authTypeClientCert      = "ClientCertificate"
	authTypeClientBearerJWT = "ClientBearerJWT"
)

const DefaultCloudEnvironmentName = "Public"

func (c *Config) SetDefaultValues() error {
	if c.CloudEnvironmentName == "" {
		c.CloudEnvironmentName = DefaultCloudEnvironmentName
	}
	return c.setCloudEnvironment()
}

func (c *Config) setCloudEnvironment() error {
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
	c.CloudEnvironment = &env
	return err
}

func (c Config) Validate(errs *packer.MultiError) {
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

	if c.UseMSI() {
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

func (c Config) useDeviceLogin() bool {
	return c.SubscriptionID != "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.ClientJWT == "" &&
		c.ClientCertPath == ""
}

func (c Config) UseMSI() bool {
	return c.SubscriptionID == "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.ClientJWT == "" &&
		c.ClientCertPath == "" &&
		c.TenantID == ""
}

func (c Config) GetServicePrincipalTokens(
	say func(string)) (
	servicePrincipalToken *adal.ServicePrincipalToken,
	servicePrincipalTokenVault *adal.ServicePrincipalToken,
	err error) {

	tenantID := c.TenantID

	var auth oAuthTokenProvider
	switch c.authType {
	case authTypeDeviceLogin:
		say("Getting tokens using device flow")
		auth = NewDeviceFlowOAuthTokenProvider(*c.CloudEnvironment, say, tenantID)
	case authTypeMSI:
		say("Getting tokens using Managed Identity for Azure")
		auth = NewMSIOAuthTokenProvider(*c.CloudEnvironment)
	case authTypeClientSecret:
		say("Getting tokens using client secret")
		auth = NewSecretOAuthTokenProvider(*c.CloudEnvironment, c.ClientID, c.ClientSecret, tenantID)
	case authTypeClientCert:
		say("Getting tokens using client certificate")
		auth, err = NewCertOAuthTokenProvider(*c.CloudEnvironment, c.ClientID, c.ClientCertPath, tenantID)
		if err != nil {
			return nil, nil, err
		}
	case authTypeClientBearerJWT:
		say("Getting tokens using client bearer JWT")
		auth = NewJWTOAuthTokenProvider(*c.CloudEnvironment, c.ClientID, c.ClientJWT, tenantID)
	default:
		panic("authType not set, call FillParameters, or set explicitly")
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
		strings.TrimRight(c.CloudEnvironment.KeyVaultEndpoint, "/"))
	if err != nil {
		return nil, nil, err
	}

	err = servicePrincipalTokenVault.EnsureFresh()
	if err != nil {
		return nil, nil, err
	}

	return servicePrincipalToken, servicePrincipalTokenVault, nil
}

func (c *Config) FillParameters() error {
	if c.authType == "" {
		if c.useDeviceLogin() {
			c.authType = authTypeDeviceLogin
		} else if c.UseMSI() {
			c.authType = authTypeMSI
		} else if c.ClientSecret != "" {
			c.authType = authTypeClientSecret
		} else if c.ClientCertPath != "" {
			c.authType = authTypeClientCert
		} else {
			c.authType = authTypeClientBearerJWT
		}
	}

	if c.authType == authTypeMSI && c.SubscriptionID == "" {

		subscriptionID, err := getSubscriptionFromIMDS()
		if err != nil {
			return fmt.Errorf("error fetching subscriptionID from VM metadata service for Managed Identity authentication: %v", err)
		}
		c.SubscriptionID = subscriptionID
	}

	if c.TenantID == "" {
		tenantID, err := common.FindTenantID(*c.CloudEnvironment, c.SubscriptionID)
		if err != nil {
			return err
		}
		c.TenantID = tenantID
	}

	if c.CloudEnvironment == nil {
		err := c.setCloudEnvironment()
		if err != nil {
			return err
		}
	}

	return nil
}

// allow override for unit tests
var findTenantID = common.FindTenantID
