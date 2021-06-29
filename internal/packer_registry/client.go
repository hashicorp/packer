package packer_registry

import (
	"errors"
	"fmt"
	"os"
	"strings"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
)

// Client is an HCP client capable of making requests on behalf of a service principal
type Client struct {
	Config ClientConfig

	Organization organization_service.ClientService
	Project      project_service.ClientService
	Packer       packerSvc.ClientService
}

// ClientConfig specifies configuration for the client that interacts with HCP
type ClientConfig struct {
	ClientID     string
	ClientSecret string

	// OrganizationID (optional) is the organization unique identifier to launch resources in.
	OrganizationID string

	// ProjectID (optional) is the project unique identifier to launch resources in.
	ProjectID string
}

// NewClient returns an authenticated client to a HCP Packer Artifact Registry.
// Client authentication requires the following environment variables be set HCP_CLIENT_ID, HCP_CLIENT_SECRET, and PACKER_ARTIFACT_REGISTRY.
// if not explicitly provided via a valid ClientConfig cfg.
// Upon error a HCPClientError will be returned.
func NewClient(_ ClientConfig) (*Client, error) {

	/*
		Not all Packer builds will publish image artifacts to a Packer Artifact Registry.
		To prevent premature HCP client errors we return immediately if no PACKER_ARTIFACT_REGISTRY environment variables is set.

		TODO when using a build block configuration for PAR, input ClientConfig is configured, this should fail hard.
	*/
	if _, ok := os.LookupEnv("PACKER_ARTIFACT_REGISTRY"); !ok {
		return nil, NewNonRegistryEnabledError()
	}

	loc := os.Getenv("PACKER_ARTIFACT_REGISTRY")
	locParts := strings.Split(loc, "/")
	if len(locParts) != 2 {
		return nil, &ClientError{
			Err: errors.New(fmt.Sprintf(`error PACKER_ARTIFACT_REGISTRY %q is not in the expected format "HCP_ORG_ID/HCP_PROJ_ID"`, loc)),
		}
	}
	orgID, projID := locParts[0], locParts[1]

	// Configure registry bits
	cl, err := httpclient.New(httpclient.Config{})
	if err != nil {
		return nil, &ClientError{
			StatusCode: InvalidHCPConfig,
			Err:        err,
		}
	}

	svc := packerSvc.New(cl, nil)
	return &Client{
		Packer: svc,
		Config: ClientConfig{
			OrganizationID: orgID,
			ProjectID:      projID,
		},
	}, nil

}
