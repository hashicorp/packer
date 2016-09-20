// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import "github.com/Azure/go-autorest/autorest/azure"

type Authenticate struct {
	env          azure.Environment
	clientID     string
	clientSecret string
	tenantID     string
}

func NewAuthenticate(env azure.Environment, clientID, clientSecret, tenantID string) *Authenticate {
	return &Authenticate{
		env:          env,
		clientID:     clientID,
		clientSecret: clientSecret,
		tenantID:     tenantID,
	}
}

func (a *Authenticate) getServicePrincipalToken() (*azure.ServicePrincipalToken, error) {
	return a.getServicePrincipalTokenWithResource(a.env.ResourceManagerEndpoint)
}

func (a *Authenticate) getServicePrincipalTokenWithResource(resource string) (*azure.ServicePrincipalToken, error) {
	oauthConfig, err := a.env.OAuthConfigForTenant(a.tenantID)
	if err != nil {
		return nil, err
	}

	spt, err := azure.NewServicePrincipalToken(
		*oauthConfig,
		a.clientID,
		a.clientSecret,
		resource)

	return spt, err
}
