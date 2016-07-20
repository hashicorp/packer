// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

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
	AzureVaultApiVersion = "2015-06-01"
)

type VaultClient struct {
	autorest.Client
	keyVaultEndpoint url.URL
}

func NewVaultClient(keyVaultEndpoint url.URL) VaultClient {
	return VaultClient{
		keyVaultEndpoint: keyVaultEndpoint,
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

func (client *VaultClient) getVaultUrl(vaultName string) string {
	return fmt.Sprintf("%s://%s.%s/", client.keyVaultEndpoint.Scheme, vaultName, client.keyVaultEndpoint.Host)
}
