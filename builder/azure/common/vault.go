// NOTE: vault APIs do not yet exist in the SDK, but once they do this code
// should be removed.

package common

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	AzureVaultApiVersion = "2016-10-01"
)

type VaultClient struct {
	autorest.Client
	keyVaultEndpoint url.URL
	SubscriptionID   string
	baseURI          string
}

func NewVaultClient(keyVaultEndpoint url.URL) VaultClient {
	return VaultClient{
		keyVaultEndpoint: keyVaultEndpoint,
	}
}

func NewVaultClientWithBaseURI(baseURI, subscriptionID string) VaultClient {
	return VaultClient{
		baseURI:        baseURI,
		SubscriptionID: subscriptionID,
	}
}

type Secret struct {
	ID    *string `json:"id,omitempty"`
	Value string  `json:"value"`
}

func (client *VaultClient) GetSecret(vaultName, secretName string) (*Secret, error) {
	p := map[string]interface{}{
		"secret-name": autorest.Encode("path", secretName),
	}
	q := map[string]interface{}{
		"api-version": AzureVaultApiVersion,
	}

	req, err := autorest.Prepare(
		&http.Request{},
		autorest.AsGet(),
		autorest.WithBaseURL(client.getVaultUrl(vaultName)),
		autorest.WithPathParameters("/secrets/{secret-name}", p),
		autorest.WithQueryParameters(q))

	if err != nil {
		return nil, err
	}

	resp, err := autorest.SendWithSender(client, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Failed to fetch secret from %s/%s, HTTP status code=%d (%s)",
			vaultName,
			secretName,
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
	}

	var secret Secret

	err = autorest.Respond(
		resp,
		autorest.ByUnmarshallingJSON(&secret))
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

// Delete deletes the specified Azure key vault.
//
// resourceGroupName is the name of the Resource Group to which the vault belongs. vaultName is the name of the vault
// to delete
func (client *VaultClient) Delete(resourceGroupName string, vaultName string) (result autorest.Response, err error) {
	req, err := client.DeletePreparer(resourceGroupName, vaultName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "keyvault.VaultsClient", "Delete", nil, "Failure preparing request")
		return
	}

	resp, err := client.DeleteSender(req)
	if err != nil {
		result.Response = resp
		err = autorest.NewErrorWithError(err, "keyvault.VaultsClient", "Delete", resp, "Failure sending request")
		return
	}

	result, err = client.DeleteResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "keyvault.VaultsClient", "Delete", resp, "Failure responding to request")
	}

	return
}

// DeletePreparer prepares the Delete request.
func (client *VaultClient) DeletePreparer(resourceGroupName string, vaultName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"SubscriptionID":    autorest.Encode("path", client.SubscriptionID),
		"vaultName":         autorest.Encode("path", vaultName),
	}

	queryParameters := map[string]interface{}{
		"api-version": AzureVaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsDelete(),
		autorest.WithBaseURL(client.baseURI),
		autorest.WithPathParameters("/subscriptions/{SubscriptionID}/resourceGroups/{resourceGroupName}/providers/Microsoft.KeyVault/vaults/{vaultName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// DeleteSender sends the Delete request. The method will close the
// http.Response Body if it receives an error.
func (client *VaultClient) DeleteSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoPollForAsynchronous(client.PollingDelay))
}

// DeleteResponder handles the response to the Delete request. The method always
// closes the http.Response Body.
func (client *VaultClient) DeleteResponder(resp *http.Response) (result autorest.Response, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByClosing())
	result.Response = resp
	return
}

func (client *VaultClient) getVaultUrl(vaultName string) string {
	return fmt.Sprintf("%s://%s.%s/", client.keyVaultEndpoint.Scheme, vaultName, client.keyVaultEndpoint.Host)
}
