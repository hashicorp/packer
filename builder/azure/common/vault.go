// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

// NOTE: vault APIs do not yet exist in the SDK, but once they do this code
// should be removed.

package common

import (
	"net/http"
	"strings"

	"github.com/Azure/go-autorest/autorest"
)

const (
	AzureVaultApiVersion     = "2015-06-01"
	AzureVaultScope          = "https://vault.azure.net"
	AzureVaultSecretTemplate = "https://{vault-name}.vault.azure.net/secrets/{secret-name}"
)

type VaultClient struct {
	autorest.Client
}

type Secret struct {
	ID         *string          `json:"id,omitempty"`
	Value      string           `json:"value"`
	Attributes SecretAttributes `json:"attributes"`
}

type SecretAttributes struct {
	Enabled bool    `json:"enabled"`
	Created *string `json:"created"`
	Updated *string `json:"updated"`
}

func (client *VaultClient) GetSecret(vaultName, secretName string) (*Secret, error) {
	p := map[string]interface{}{
		"secret-name": secretName,
	}
	q := map[string]interface{}{
		"api-version": AzureVaultApiVersion,
	}

	secretURL := strings.Replace(AzureVaultSecretTemplate, "{vault-name}", vaultName, -1)

	req, err := autorest.Prepare(&http.Request{},
		autorest.AsGet(),
		autorest.WithBaseURL(secretURL),
		autorest.WithPathParameters(p),
		autorest.WithQueryParameters(q))

	if err != nil {
		return nil, err
	}

	//resp, err := v.Send(req, http.StatusOK)
	resp, err := autorest.SendWithSender(client, req)
	if err != nil {
		return nil, err
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
