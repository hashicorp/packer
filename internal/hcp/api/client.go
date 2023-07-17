// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package api provides access to the HCP Packer Registry API.
package api

import (
	"fmt"
	"log"
	"time"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	organizationSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	projectSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
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

	if client.ProjectID != "" && client.OrganizationID == "" {
		getProjParams := projectSvc.NewProjectServiceGetParams()
		getProjParams.ID = client.ProjectID
		project, err := RetryProjectServiceGet(client, getProjParams)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch project %q: %v", client.ProjectID, err)
		}

		client.ProjectID = project.Payload.Project.ID
		client.OrganizationID = project.Payload.Project.Parent.ID

	}

	if client.ProjectID == "" {
		// For the initial release of the HCP TFP, since only one project was allowed per organization at the time,
		// the provider handled used the single organization's single project by default, instead of requiring the
		// user to set it. Once multiple projects are available, this helper issues a warning: when multiple projects exist within the org,
		// a project ID should be set on the provider or on each resource. Otherwise, the oldest project will be used by default.
		// This helper will eventually be deprecated after a migration period.
		project, err := getProjectFromCredentials(client)
		if err != nil {
			return nil, fmt.Errorf("unable to get project from credentials: %v", err)
		}

		client.ProjectID = project.ID
		client.OrganizationID = project.Parent.ID
	}

	return client, nil
}

const (
	retryCount   = 10
	retryDelay   = 10
	counterStart = 1
)

var errorCodesToRetry = [...]int{502, 503, 504}

// Helper to check what requests to retry based on the response HTTP code
func shouldRetryErrorCode(errorCode int, errorCodesToRetry []int) bool {
	for i := range errorCodesToRetry {
		if errorCodesToRetry[i] == errorCode {
			return true
		}
	}
	return false
}

// RetryProjectServiceGet wraps the ProjectServiceGet function in a loop that supports retrying the GET request
func RetryProjectServiceGet(client *Client, params *projectSvc.ProjectServiceGetParams) (*projectSvc.ProjectServiceGetOK, error) {
	resp, err := client.Project.ProjectServiceGet(params, nil)

	if err != nil {
		serviceErr, ok := err.(*projectSvc.ProjectServiceGetDefault)
		if !ok {
			return nil, err
		}

		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), errorCodesToRetry[:]) && counter < retryCount {
			resp, err = client.Project.ProjectServiceGet(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to get configured project. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}

// RetryOrganizationServiceList wraps the OrganizationServiceList function in a loop that supports retrying the GET request
func RetryOrganizationServiceList(client *Client, params *organizationSvc.OrganizationServiceListParams) (*organizationSvc.OrganizationServiceListOK, error) {
	resp, err := client.Organization.OrganizationServiceList(params, nil)

	if err != nil {
		serviceErr, ok := err.(*organizationSvc.OrganizationServiceListDefault)
		if !ok {
			return nil, err
		}
		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), errorCodesToRetry[:]) && counter < retryCount {
			resp, err = client.Organization.OrganizationServiceList(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to get list of organizations. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}

// RetryProjectServiceList wraps the ProjectServiceList function in a loop that supports retrying the GET request
func RetryProjectServiceList(client *Client, params *projectSvc.ProjectServiceListParams) (*projectSvc.ProjectServiceListOK, error) {
	resp, err := client.Project.ProjectServiceList(params, nil)

	if err != nil {
		serviceErr, ok := err.(*projectSvc.ProjectServiceListDefault)
		if !ok {
			return nil, err
		}

		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), errorCodesToRetry[:]) && counter < retryCount {
			resp, err = client.Project.ProjectServiceList(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to get list of projects. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}

// getProjectFromCredentials uses the configured client credentials to
// fetch the associated organization and returns that organization's
// single project.
func getProjectFromCredentials(client *Client) (project *models.ResourcemanagerProject, err error) {
	if client.OrganizationID == "" {
		// Get the organization ID.
		listOrgParams := organizationSvc.NewOrganizationServiceListParams()
		listOrgResp, err := RetryOrganizationServiceList(client, listOrgParams)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch organization list: %v", err)
		}
		orgLen := len(listOrgResp.Payload.Organizations)
		if orgLen != 1 {
			return nil, fmt.Errorf("unexpected number of organizations: expected 1, actual: %v", orgLen)
		}
		client.OrganizationID = listOrgResp.Payload.Organizations[0].ID
	}

	// Get the project using the organization ID.
	listProjParams := projectSvc.NewProjectServiceListParams()
	listProjParams.ScopeID = &client.OrganizationID
	scopeType := string(models.ResourceIDResourceTypeORGANIZATION)
	listProjParams.ScopeType = &scopeType
	listProjResp, err := RetryProjectServiceList(client, listProjParams)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch project id: %v", err)
	}
	if len(listProjResp.Payload.Projects) > 1 {
		log.Printf("[WARNING] Multiple HCP projects found, will pick the oldest one by default\n" +
			"To specify which project to use, set the HCP_PROJECT_ID environment variable to the one you want to use.")
		return getOldestProject(listProjResp.Payload.Projects), nil
	}
	project = listProjResp.Payload.Projects[0]
	return project, nil
}

// getOldestProject retrieves the oldest project from a list based on its created_at time.
func getOldestProject(projects []*models.ResourcemanagerProject) (oldestProj *models.ResourcemanagerProject) {
	oldestTime := time.Now()

	for _, proj := range projects {
		projTime := time.Time(proj.CreatedAt)
		if projTime.Before(oldestTime) {
			oldestProj = proj
			oldestTime = projTime
		}
	}
	return oldestProj
}
