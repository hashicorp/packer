// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package api provides access to the HCP Packer Registry API.
package api

import (
	"fmt"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	organizationSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	projectSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
	"github.com/hashicorp/packer/version"
)

// DeprecatedClient is an HCP client capable of making requests on behalf of a service principal
type DeprecatedClient struct {
	Packer         packerSvc.ClientService
	Organization   organizationSvc.ClientService
	Project        projectSvc.ClientService
	OrganizationID string
	ProjectID      string
}

// NewDeprecatedClient returns an authenticated client to a HCP Packer Registry.
// Client authentication requires the following environment variables be set HCP_CLIENT_ID and HCP_CLIENT_SECRET.
// Upon error a HCPClientError will be returned.
func NewDeprecatedClient() (*DeprecatedClient, error) {
	// Use NewClient to validate HCP configuration provided by user.
	tempClient, err := NewClient()
	if err != nil {
		return nil, err
	}

	hcpClientCfg := httpclient.Config{
		SourceChannel: fmt.Sprintf("packer/%s", version.PackerVersion.FormattedVersion()),
	}
	cl, err := httpclient.New(hcpClientCfg)
	if err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}

	client := DeprecatedClient{
		Packer:         packerSvc.New(cl, nil),
		Organization:   organizationSvc.New(cl, nil),
		Project:        projectSvc.New(cl, nil),
		OrganizationID: tempClient.OrganizationID,
		ProjectID:      tempClient.ProjectID,
	}
	return &client, nil
}
