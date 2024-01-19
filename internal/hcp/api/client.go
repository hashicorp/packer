// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package api provides access to the HCP Packer Registry API.
package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	organizationSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	projectSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	rmmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
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

	hcpClientCfg := httpclient.Config{
		SourceChannel: fmt.Sprintf("packer/%s", version.PackerVersion.FormattedVersion()),
	}
	if err := hcpClientCfg.Canonicalize(); err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}

	cl, err := httpclient.New(hcpClientCfg)
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
	// A client.Config.hcpConfig is set when calling Canonicalize on basic HCP httpclient, as on line 52.
	// If a user sets HCP_* env. variables they will be loaded into the client via the SDK and used for any client calls.
	// For HCP_ORGANIZATION_ID and HCP_PROJECT_ID if they are both set via env. variables the call to hcpClientCfg.Connicalize()
	// will automatically loaded them using the FromEnv configOption.
	//
	// If both values are set we should have all that we need to continue so we can returned the configured client.
	if hcpClientCfg.Profile().OrganizationID != "" && hcpClientCfg.Profile().ProjectID != "" {
		client.OrganizationID = hcpClientCfg.Profile().OrganizationID
		client.ProjectID = hcpClientCfg.Profile().ProjectID

		return client, nil
	}

	if client.OrganizationID == "" {
		err := client.loadOrganizationID()
		if err != nil {
			return nil, &ClientError{
				StatusCode: InvalidClientConfig,
				Err:        err,
			}
		}
	}

	if client.ProjectID == "" {
		err := client.loadProjectID()
		if err != nil {
			return nil, &ClientError{
				StatusCode: InvalidClientConfig,
				Err:        err,
			}
		}
	}

	return client, nil
}

func (c *Client) loadOrganizationID() error {
	if env.HasOrganizationID() {
		c.OrganizationID = os.Getenv(env.HCPOrganizationID)
		return nil
	}
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
	if env.HasProjectID() {
		c.ProjectID = os.Getenv(env.HCPProjectID)
		err := c.ValidateRegistryForProject()
		if err != nil {
			return fmt.Errorf("project validation for id %q responded in error: %v", c.ProjectID, err)
		}
		return nil
	}
	// Get the project using the organization ID.
	listProjParams := projectSvc.NewProjectServiceListParams()
	listProjParams.ScopeID = &c.OrganizationID
	scopeType := string(rmmodels.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	listProjParams.ScopeType = &scopeType
	listProjResp, err := c.Project.ProjectServiceList(listProjParams, nil)

	if err != nil {
		//For permission errors, our service principal may not have the ability
		// to see all projects for an Org; this is the case for project-level service principals.
		serviceErr, ok := err.(*projectSvc.ProjectServiceListDefault)
		if !ok {
			return fmt.Errorf("unable to fetch project list: %v", err)
		}
		if serviceErr.Code() == http.StatusForbidden {
			return fmt.Errorf("unable to fetch project\n\n"+
				"If the provided credentials are tied to a specific project try setting the %s environment variable to one you want to use.", env.HCPProjectID)
		}
	}

	if len(listProjResp.Payload.Projects) > 1 {
		log.Printf("[WARNING] Multiple HCP projects found, will pick the oldest one by default\n"+
			"To specify which project to use, set the %s environment variable to the one you want to use.", env.HCPProjectID)
	}

	proj, err := getOldestProject(listProjResp.Payload.Projects)
	if err != nil {
		return err
	}
	c.ProjectID = proj.ID
	return nil
}

// getOldestProject retrieves the oldest project from a list based on its created_at time.
func getOldestProject(projects []*rmmodels.HashicorpCloudResourcemanagerProject) (*rmmodels.HashicorpCloudResourcemanagerProject, error) {
	if len(projects) == 0 {
		return nil, fmt.Errorf("no project found")
	}

	oldestTime := time.Now()
	var oldestProj *rmmodels.HashicorpCloudResourcemanagerProject
	for _, proj := range projects {
		projTime := time.Time(proj.CreatedAt)
		if projTime.Before(oldestTime) {
			oldestProj = proj
			oldestTime = projTime
		}
	}
	return oldestProj, nil
}

// ValidateRegistryForProject validates that there is an active registry associated to the configured organization and project ids.
// A successful validation will result in a nil response. All other response represent an invalid registry error request or a registry not found error.
func (c *Client) ValidateRegistryForProject() error {
	params := packerSvc.NewPackerServiceGetRegistryParams()
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID

	resp, err := c.Packer.PackerServiceGetRegistry(params, nil)
	if err != nil {
		return err
	}

	if resp.GetPayload().Registry == nil {
		return fmt.Errorf("No active HCP Packer registry was found for the organization %q and project %q", c.OrganizationID, c.ProjectID)
	}

	return nil

}
