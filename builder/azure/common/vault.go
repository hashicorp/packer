// NOTE: vault APIs do not yet exist in the SDK, but once they do this code
// should be removed.

package common

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
)

const (
	AzureVaultApiVersion = "2016-10-01"
)

// Enables us to test steps that access this cli
type AZVaultClientIface interface {
	GetSecret(string, string) (*Secret, error)
	SetSecret(string, string, string) error
}

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
		autorest.WithQueryParameters(q),
	)

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

func (client *VaultClient) SetSecret(vaultName, secretName string, secretValue string) error {
	p := map[string]interface{}{
		"secret-name": autorest.Encode("path", secretName),
	}
	q := map[string]interface{}{
		"api-version": AzureVaultApiVersion,
	}

	jsonBody := fmt.Sprintf(`{"value": "%s"}`, secretValue)

	req, err := autorest.Prepare(
		&http.Request{},
		autorest.AsPut(),
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.WithBaseURL(client.getVaultUrl(vaultName)),
		autorest.WithPathParameters("/secrets/{secret-name}", p),
		autorest.WithQueryParameters(q),
		autorest.WithString(jsonBody),
	)

	if err != nil {
		return err
	}

	resp, err := autorest.SendWithSender(client, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"Failed to set secret to %s/%s, HTTP status code=%d (%s)",
			vaultName,
			secretName,
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
	}

	return nil
}

func (client *VaultClient) getVaultUrl(vaultName string) string {
	return fmt.Sprintf("%s://%s.%s/", client.keyVaultEndpoint.Scheme, vaultName, client.keyVaultEndpoint.Host)
}
