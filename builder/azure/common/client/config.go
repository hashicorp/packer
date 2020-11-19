//go:generate struct-markdown

package client

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// Config allows for various ways to authenticate Azure clients.
// When `client_id` and `subscription_id` are specified, Packer will use the
// specified Azure Active Directory (AAD) Service Principal (SP).
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
	cloudEnvironment     *azure.Environment

	// Authentication fields

	// The application ID of the AAD Service Principal.
	// Requires either `client_secret`, `client_cert_path` or `client_jwt` to be set as well.
	ClientID string `mapstructure:"client_id"`
	// A password/secret registered for the AAD SP.
	ClientSecret string `mapstructure:"client_secret"`
	// The path to a pem-encoded certificate that will be used to authenticate
	// as the specified AAD SP.
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

	// Flag to use Azure CLI authentication. Defaults to false.
	// CLI auth will use the information from an active `az login` session to connect to Azure and set the subscription id and tenant id associated to the signed in account.
	// If enabled, it will use the authentication provided by the `az` CLI.
	// Azure CLI authentication will use the credential marked as `isDefault` and can be verified using `az account show`.
	// Works with normal authentication (`az login`) and service principals (`az login --service-principal --username APP_ID --password PASSWORD --tenant TENANT_ID`).
	// Ignores all other configurations if enabled.
	UseAzureCLIAuth bool `mapstructure:"use_azure_cli_auth" required:"false"`
}

const (
	authTypeDeviceLogin     = "DeviceLogin"
	authTypeMSI             = "ManagedIdentity"
	authTypeClientSecret    = "ClientSecret"
	authTypeClientCert      = "ClientCertificate"
	authTypeClientBearerJWT = "ClientBearerJWT"
	authTypeAzureCLI        = "AzureCLI"
)

const DefaultCloudEnvironmentName = "Public"

func (c *Config) SetDefaultValues() error {
	if c.CloudEnvironmentName == "" {
		c.CloudEnvironmentName = DefaultCloudEnvironmentName
	}
	return c.setCloudEnvironment()
}

func (c *Config) CloudEnvironment() *azure.Environment {
	return c.cloudEnvironment
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
	c.cloudEnvironment = &env
	return err
}

func (c Config) Validate(errs *packersdk.MultiError) {
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

	if c.UseCLI() {
		return
	}

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
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("client_cert_path is not an accessible file: %v", err))
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
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("client_jwt is not a JWT: %v", err))
		} else {
			if claims.ExpiresAt < time.Now().Add(5*time.Minute).Unix() {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("client_jwt will expire within 5 minutes, please use a JWT that is valid for at least 5 minutes"))
			}
			if t, ok := token.Header["x5t"]; !ok || t == "" {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("client_jwt is missing the x5t header value, which is required for bearer JWT client authentication to Azure"))
			}
		}

		return
	}

	errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("No valid set of authentication values specified:\n"+
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

func (c Config) UseCLI() bool {
	return c.UseAzureCLIAuth == true
}

func (c Config) UseMSI() bool {
	return c.SubscriptionID == "" &&
		c.ClientID == "" &&
		c.ClientSecret == "" &&
		c.ClientJWT == "" &&
		c.ClientCertPath == "" &&
		c.TenantID == ""
}

func (c Config) GetServicePrincipalTokens(say func(string)) (
	servicePrincipalToken *adal.ServicePrincipalToken,
	servicePrincipalTokenVault *adal.ServicePrincipalToken,
	err error) {

	servicePrincipalToken, err = c.GetServicePrincipalToken(say,
		c.CloudEnvironment().ResourceManagerEndpoint)
	if err != nil {
		return nil, nil, err
	}
	servicePrincipalTokenVault, err = c.GetServicePrincipalToken(say,
		strings.TrimRight(c.CloudEnvironment().KeyVaultEndpoint, "/"))
	if err != nil {
		return nil, nil, err
	}
	return servicePrincipalToken, servicePrincipalTokenVault, nil
}

func (c Config) GetServicePrincipalToken(
	say func(string), forResource string) (
	servicePrincipalToken *adal.ServicePrincipalToken,
	err error) {

	var auth oAuthTokenProvider
	switch c.authType {
	case authTypeDeviceLogin:
		say("Getting tokens using device flow")
		auth = NewDeviceFlowOAuthTokenProvider(*c.cloudEnvironment, say, c.TenantID)
	case authTypeAzureCLI:
		say("Getting tokens using Azure CLI")
		auth = NewCliOAuthTokenProvider(*c.cloudEnvironment, say, c.TenantID)
	case authTypeMSI:
		say("Getting tokens using Managed Identity for Azure")
		auth = NewMSIOAuthTokenProvider(*c.cloudEnvironment)
	case authTypeClientSecret:
		say("Getting tokens using client secret")
		auth = NewSecretOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientSecret, c.TenantID)
	case authTypeClientCert:
		say("Getting tokens using client certificate")
		auth, err = NewCertOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientCertPath, c.TenantID)
		if err != nil {
			return nil, err
		}
	case authTypeClientBearerJWT:
		say("Getting tokens using client bearer JWT")
		auth = NewJWTOAuthTokenProvider(*c.cloudEnvironment, c.ClientID, c.ClientJWT, c.TenantID)
	default:
		panic("authType not set, call FillParameters, or set explicitly")
	}

	servicePrincipalToken, err = auth.getServicePrincipalTokenWithResource(forResource)
	if err != nil {
		return nil, err
	}

	err = servicePrincipalToken.EnsureFresh()
	if err != nil {
		return nil, err
	}

	return servicePrincipalToken, nil
}

// FillParameters capture the user intent from the supplied parameter set in authType, retrieves the TenantID and CloudEnvironment if not specified.
// The SubscriptionID is also retrieved in case MSI auth is requested.
func (c *Config) FillParameters() error {
	if c.authType == "" {
		if c.useDeviceLogin() {
			c.authType = authTypeDeviceLogin
		} else if c.UseCLI() {
			c.authType = authTypeAzureCLI
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

	if c.cloudEnvironment == nil {
		err := c.setCloudEnvironment()
		if err != nil {
			return err
		}
	}

	if c.authType == authTypeAzureCLI {
		tenantID, subscriptionID, err := getIDsFromAzureCLI()
		if err != nil {
			return fmt.Errorf("error fetching tenantID and subscriptionID from Azure CLI (are you logged on using `az login`?): %v", err)
		}

		c.TenantID = tenantID
		c.SubscriptionID = subscriptionID
	}

	if c.TenantID == "" {
		tenantID, err := findTenantID(*c.cloudEnvironment, c.SubscriptionID)
		if err != nil {
			return err
		}
		c.TenantID = tenantID
	}

	return nil
}

// allow override for unit tests
var findTenantID = FindTenantID
