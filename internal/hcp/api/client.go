// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package api provides access to the HCP Packer Registry API.
package api

import (
	"fmt"
	"log"
	"os"
	"time"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	organizationSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	projectSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/models"
	rmmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
	"github.com/hashicorp/packer/internal/hcp/env"
	"github.com/hashicorp/packer/version"
)

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Packer       packerSvc.ClientService
	Organization organizationSvc.ClientService
	Project      projectSvc.ClientService

	// OrganizationID  is the organization unique identifier on HCP.
	OrganizationID string

	// ProjectID  is the project unique identifier on HCP.
	ProjectID string
}

// NewClient returns an authenticated client to a HCP Packer Registry.
// Client authentication requires the following environment variables be set HCP_CLIENT_ID and HCP_CLIENT_SECRET.
// Upon error a HCPClientError will be returned.
func NewClient() (*Client, error) {
	if !env.HasHCPCredentials() {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        fmt.Errorf("the client authentication requires both %s and %s environment variables to be set", env.HCPClientID, env.HCPClientSecret),
		}
	}

	cl, err := httpclient.New(httpclient.Config{
		SourceChannel: fmt.Sprintf("packer/%s", version.PackerVersion.FormattedVersion()),
	})
	if err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}

	client := &Client{
		Packer:       packerSvc.New(cl, nil),
		Organization: organizationSvc.New(cl, nil),
		Project:      projectSvc.New(cl, nil),
	}

	if err := client.loadOrganizationID(); err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}
	if err := client.loadProjectID(); err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}

	return client, nil
}

func (c *Client) loadOrganizationID() error {
	// Get the organization ID.
	listOrgParams := organizationSvc.NewOrganizationServiceListParams()
	listOrgResp, err := c.Organization.OrganizationServiceList(listOrgParams, nil)
	if err != nil {
		return fmt.Errorf("unable to fetch organization list: %v", err)
	}
	orgLen := len(listOrgResp.Payload.Organizations)
	if orgLen != 1 {
		return fmt.Errorf("unexpected number of organizations: expected 1, actual: %v", orgLen)
	}
	c.OrganizationID = listOrgResp.Payload.Organizations[0].ID
	return nil
}

func (c *Client) loadProjectID() error {
	// Get the project using the organization ID.
	listProjParams := projectSvc.NewProjectServiceListParams()
	listProjParams.ScopeID = &c.OrganizationID
	scopeType := string(rmmodels.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	listProjParams.ScopeType = &scopeType
	listProjResp, err := c.Project.ProjectServiceList(listProjParams, nil)
	if err != nil {
		return fmt.Errorf("unable to fetch project id: %v", err)
	}

	if env.HasProjectID() {
		proj, err := findProjectByID(os.Getenv(env.HCPProjectID), listProjResp.Payload.Projects)
		if err != nil {
			return err
		}

		c.ProjectID = proj.ID
	} else {
		if len(listProjResp.Payload.Projects) > 1 {
			log.Printf("[WARNING] Multiple HCP projects found, will pick the oldest one by default\n" +
				"To specify which project to use, set the HCP_PROJECT_ID environment variable to the one you want to use.")
		}

		proj, err := findOldestProject(listProjResp.Payload.Projects)
		if err != nil {
			return err
		}

		c.ProjectID = proj.ID
	}

	return nil
}

func findOldestProject(projs []*models.HashicorpCloudResourcemanagerProject) (*models.HashicorpCloudResourcemanagerProject, error) {
	if len(projs) == 0 {
		return nil, fmt.Errorf("no project found")
	}

	proj := projs[0]
	for i := 1; i < len(projs); i++ {
		nxtProj := projs[i]

		if time.Time(nxtProj.CreatedAt).Before(time.Time(proj.CreatedAt)) {
			proj = nxtProj
		}
	}

	return proj, nil
}

func findProjectByID(projID string, projs []*models.HashicorpCloudResourcemanagerProject) (*models.HashicorpCloudResourcemanagerProject, error) {
	for _, proj := range projs {
		if proj.ID == projID {
			return proj, nil
		}
	}

	return nil, fmt.Errorf("No project %q found", projID)
}
