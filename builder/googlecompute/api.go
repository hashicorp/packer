package googlecompute

import (
	"errors"
	"net/http"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/compute/v1beta16"
)

// GoogleComputeClient represents a GCE client.
type GoogleComputeClient struct {
	ProjectId     string
	Service       *compute.Service
	Zone          string
	clientSecrets *clientSecrets
}

// InstanceConfig represents a GCE instance configuration.
// Used for creating machine instances.
/*
type InstanceConfig struct {
	Description       string
	Image             string
	MachineType       string
	Metadata          *compute.Metadata
	Name              string
	NetworkInterfaces []*compute.NetworkInterface
	ServiceAccounts   []*compute.ServiceAccount
	Tags              *compute.Tags
}
*/

// New initializes and returns a *GoogleComputeClient.
//
// The projectId must be the project name, i.e. myproject, not the project
// number.
func New(projectId string, zone string, c *clientSecrets, pemKey []byte) (*GoogleComputeClient, error) {
	googleComputeClient := &GoogleComputeClient{
		ProjectId: projectId,
		Zone:      zone,
	}
	// Get the access token.
	t := jwt.NewToken(c.Web.ClientEmail, "", pemKey)
	t.ClaimSet.Aud = c.Web.TokenURI
	httpClient := &http.Client{}
	token, err := t.Assert(httpClient)
	if err != nil {
		return nil, err
	}
	config := &oauth.Config{
		ClientId: c.Web.ClientId,
		Scope:    "",
		TokenURL: c.Web.TokenURI,
		AuthURL:  c.Web.AuthURI,
	}
	transport := &oauth.Transport{Config: config}
	transport.Token = token
	s, err := compute.New(transport.Client())
	if err != nil {
		return nil, err
	}
	googleComputeClient.Service = s
	return googleComputeClient, nil
}

// InstanceStatus returns a string representing the status of the named instance.
// Status will be one of: "PROVISIONING", "STAGING", "RUNNING", "STOPPING",
// "STOPPED", "TERMINATED".
func (g *GoogleComputeClient) InstanceStatus(zone, name string) (string, error) {
	instanceGetCall := g.Service.Instances.Get(g.ProjectId, zone, name)
	instance, err := instanceGetCall.Do()
	if err != nil {
		return "", err
	}
	return instance.Status, nil
}

// CreateImage registers a GCE Image with a project.
func (g *GoogleComputeClient) CreateImage(name, description, sourceURL string) (*compute.Operation, error) {
	imageRawDisk := &compute.ImageRawDisk{
		ContainerType: "TAR",
		Source:        sourceURL,
	}
	image := &compute.Image{
		Description: description,
		Name:        name,
		RawDisk:     imageRawDisk,
		SourceType:  "RAW",
	}
	imageInsertCall := g.Service.Images.Insert(g.ProjectId, image)
	operation, err := imageInsertCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// GetNatIp returns the public IPv4 address for named GCE instance.
func (g *GoogleComputeClient) GetNatIP(zone, name string) (string, error) {
	instanceGetCall := g.Service.Instances.Get(g.ProjectId, zone, name)
	instance, err := instanceGetCall.Do()
	if err != nil {
		return "", err
	}
	for _, ni := range instance.NetworkInterfaces {
		if ni.AccessConfigs == nil {
			continue
		}
		for _, ac := range ni.AccessConfigs {
			if ac.NatIP != "" {
				return ac.NatIP, nil
			}
		}
	}
	return "", nil
}

// ZoneOperationStatus returns the status for the named zone operation.
func (g *GoogleComputeClient) ZoneOperationStatus(zone, name string) (string, error) {
	zoneOperationsGetCall := g.Service.ZoneOperations.Get(g.ProjectId, zone, name)
	operation, err := zoneOperationsGetCall.Do()
	if err != nil {
		return "", err
	}
	if operation.Status == "DONE" {
		err = processOperationStatus(operation)
		if err != nil {
			return operation.Status, err
		}
	}
	return operation.Status, nil
}

// GlobalOperationStatus returns the status for the named global operation.
func (g *GoogleComputeClient) GlobalOperationStatus(name string) (string, error) {
	globalOperationsGetCall := g.Service.GlobalOperations.Get(g.ProjectId, name)
	operation, err := globalOperationsGetCall.Do()
	if err != nil {
		return "", err
	}
	if operation.Status == "DONE" {
		err = processOperationStatus(operation)
		if err != nil {
			return operation.Status, err
		}
	}
	return operation.Status, nil
}

// processOperationStatus extracts errors from the specified operation.
func processOperationStatus(o *compute.Operation) error {
	if o.Error != nil {
		messages := make([]string, len(o.Error.Errors))
		for _, e := range o.Error.Errors {
			messages = append(messages, e.Message)
		}
		return errors.New(strings.Join(messages, "\n"))
	}
	return nil
}

// DeleteImage deletes the named image. Returns a Global Operation.
func (g *GoogleComputeClient) DeleteImage(name string) (*compute.Operation, error) {
	imagesDeleteCall := g.Service.Images.Delete(g.ProjectId, name)
	operation, err := imagesDeleteCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// DeleteInstance deletes the named instance. Returns a Zone Operation.
func (g *GoogleComputeClient) DeleteInstance(zone, name string) (*compute.Operation, error) {
	instanceDeleteCall := g.Service.Instances.Delete(g.ProjectId, zone, name)
	operation, err := instanceDeleteCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}
