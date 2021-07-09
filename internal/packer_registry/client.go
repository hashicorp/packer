package packer_registry

import (
	"errors"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
)

// ClientConfig specifies configuration for the client that interacts with HCP
type ClientConfig struct {
	ClientID     string
	ClientSecret string

	// OrganizationID  is the organization unique identifier to launch resources in.
	OrganizationID string

	// ProjectID  is the project unique identifier to launch resources in.
	ProjectID string
}

func (cfg ClientConfig) Validate() error {
	if cfg.OrganizationID == "" {
		return errors.New(`no valid HCP Organization ID found, check that HCP_PACKER_REGISTRY is in the format "HCP_ORG_ID/HCP_PROJ_ID"`)
	}

	if cfg.ProjectID == "" {
		return errors.New(`no valid HCP Project ID found, check that HCP_PACKER_REGISTRY is in the format "HCP_ORG_ID/HCP_PROJ_ID"`)
	}

	return nil
}

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Config ClientConfig

	Organization organization_service.ClientService
	Project      project_service.ClientService
	Packer       packerSvc.ClientService
}

// NewClient returns an authenticated client to a HCP Packer Artifact Registry.
// Client authentication requires the following environment variables be set HCP_CLIENT_ID, HCP_CLIENT_SECRET, and HCP_PACKER_REGISTRY.
// if not explicitly provided via a valid ClientConfig cfg.
// Upon error a HCPClientError will be returned.
func NewClient(cfg ClientConfig) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}

	cl, err := httpclient.New(httpclient.Config{})
	if err != nil {
		return nil, &ClientError{
			StatusCode: InvalidClientConfig,
			Err:        err,
		}
	}

	svc := packerSvc.New(cl, nil)
	return &Client{
		Packer: svc,
		Config: cfg,
	}, nil

}
