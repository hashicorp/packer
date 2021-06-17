package packer_registry

import (
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

	Project      project_service.ClientService
	Organization organization_service.ClientService
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

func NewClient() (*Client, error) {
	cl, err := httpclient.New(httpclient.Config{})
	if err != nil {
		return nil, err
	}

	loc := os.Getenv("PACKER_ARTIFACT_REGISTRY")
	if loc == "" {
		return nil, fmt.Errorf("error encountered when configuring PAR connection: no PACKER_ARTIFACT_REGISTRY defined")
	}

	locParts := strings.Split(loc, "/")
	if len(locParts) != 2 {
		return nil, fmt.Errorf(`error Artifact Registry location %q is not in the expected format "HCP_ORG_ID/HCP_PROJ_ID"`, loc)
	}
	orgID, projID := locParts[0], locParts[1]
	// Configure registry bits
	svc := packerSvc.New(cl, nil)
	return &Client{
		Packer: svc,
		Config: ClientConfig{
			OrganizationID: orgID,
			ProjectID:      projID,
		},
	}, nil

}
