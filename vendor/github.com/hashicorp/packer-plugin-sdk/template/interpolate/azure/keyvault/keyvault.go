package keyvault

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	keyvaultmgmt "github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2018-02-14/keyvault"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"
)

type Client struct {
	VaultsClient     *keyvault.VaultsClient
	ManagementClient *keyvaultmgmt.BaseClient
}

func New() (*Client, error) {
	c := &Client{}

	vaultsClient, managementClient, err := newVaultsClients()
	if err != nil {
		return nil, err
	}

	c.VaultsClient = vaultsClient
	c.ManagementClient = managementClient

	return c, nil
}

func (c *Client) GetKeyVaultId(resourceGroup string, keyVaultName string) (string, error) {
	resp, err := c.VaultsClient.Get(context.Background(), resourceGroup, keyVaultName)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (c *Client) GetSecret(keyVaultId string, key string) (string, error) {
	keyVaultBaseUri, err := GetKeyVaultBaseUrlFromID(context.Background(), c.VaultsClient, keyVaultId)
	if err != nil {
		return "", err
	}

	resp, err := c.ManagementClient.GetSecret(context.Background(), keyVaultBaseUri, key, "")
	if err != nil {
		return "", err
	}

	return *resp.Value, nil
}

func newVaultsClients() (*keyvault.VaultsClient, *keyvaultmgmt.BaseClient, error) {
	builder := &authentication.Builder{
		SupportsClientCertAuth:   true,
		SupportsClientSecretAuth: true,
		SupportsAzureCliToken:    true,
	}

	config, err := builder.Build()
	if err != nil {
		return nil, nil, err
	}

	env, err := authentication.DetermineEnvironment(config.Environment)
	if err != nil {
		return nil, nil, err
	}

	oauthConfig, err := config.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, nil, err
	}

	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, nil, fmt.Errorf("Unable to configure OAuthConfig for tenant %s", config.TenantID)
	}

	sender := sender.BuildSender("AzureRM")
	keyVaultAuth, err := config.GetAuthorizationToken(sender, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, nil, err
	}
	endpoint := env.ResourceManagerEndpoint

	VaultsClient := keyvault.NewVaultsClientWithBaseURI(endpoint, config.SubscriptionID)
	VaultsClient.Authorizer = keyVaultAuth

	ManagementClient := keyvaultmgmt.New()
	MgmtAuthorizer := config.BearerAuthorizerCallback(sender, oauthConfig)
	ManagementClient.Client.Authorizer = MgmtAuthorizer

	return &VaultsClient, &ManagementClient, nil
}

func GetKeyVaultBaseUrlFromID(ctx context.Context, client *keyvault.VaultsClient, keyVaultId string) (string, error) {
	resourceGroup, vaultName, err := getKeyVaultResourceGroupAndName(keyVaultId)
	if err != nil {
		return "", err
	}

	resp, err := client.Get(ctx, resourceGroup, vaultName)
	if err != nil {
		if resp.Response.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("Error unable to find KeyVault %q (Resource Group %q): %+v", vaultName, resourceGroup, err)
		}
		return "", fmt.Errorf("Error making Read request on KeyVault %q (Resource Group %q): %+v", vaultName, resourceGroup, err)
	}

	if resp.Properties == nil || resp.Properties.VaultURI == nil {
		return "", fmt.Errorf("vault (%s) response properties or VaultURI is nil", keyVaultId)
	}

	return *resp.Properties.VaultURI, nil
}

// A KeyVaultId has the following structure: /subscriptions/<subscription_id>/resourceGroups/<resourceGroup_name>/providers/Microsoft.KeyVault/vaults/<vault_name>
func getKeyVaultResourceGroupAndName(keyVaultId string) (string, string, error) {
	if keyVaultId == "" {
		return "", "", fmt.Errorf("keyVaultId is empty")
	}

	// Trim / from beginning and end and split KeyVaultId into pairs as structure suggests
	components := strings.Split(strings.TrimSuffix(strings.TrimPrefix(keyVaultId, "/"), "/"), "/")

	// We should have an even number of key-value pairs.
	if len(components)%2 != 0 {
		return "", "", fmt.Errorf("keyVaultId malformed: The number of path segments is not divisible by 2 in %s", keyVaultId)
	}

	// We iterate the structure looking for the known keys after we can find the desired values
	var resourceGroup, vaultName string = "", ""
	for current := 0; current < len(components); current += 2 {
		key := components[current]
		value := components[current+1]

		// Check key/value for empty strings.
		if key == "" || value == "" {
			return "", "", fmt.Errorf("keyVaultId malformed: Key/Value cannot be empty strings. Key: '%s', Value: '%s'", key, value)
		}

		switch key {
		case "resourceGroups":
			resourceGroup = value
		case "vaults":
			vaultName = value
		default:
			continue
		}
	}

	if resourceGroup == "" || vaultName == "" {
		return "", "", fmt.Errorf("keyVaultId malformed: Could not find resourceGroup and vaultName in %s", keyVaultId)
	}

	return resourceGroup, vaultName, nil
}
