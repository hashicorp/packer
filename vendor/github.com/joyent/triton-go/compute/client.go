package compute

import (
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type ComputeClient struct {
	Client *client.Client
}

func newComputeClient(client *client.Client) *ComputeClient {
	return &ComputeClient{
		Client: client,
	}
}

// NewClient returns a new client for working with Compute endpoints and
// resources within CloudAPI
func NewClient(config *triton.ClientConfig) (*ComputeClient, error) {
	// TODO: Utilize config interface within the function itself
	client, err := client.New(config.TritonURL, config.MantaURL, config.AccountName, config.Signers...)
	if err != nil {
		return nil, err
	}
	return newComputeClient(client), nil
}

// Datacenters returns a Compute client used for accessing functions pertaining
// to DataCenter functionality in the Triton API.
func (c *ComputeClient) Datacenters() *DataCentersClient {
	return &DataCentersClient{c.Client}
}

// Images returns a Compute client used for accessing functions pertaining to
// Images functionality in the Triton API.
func (c *ComputeClient) Images() *ImagesClient {
	return &ImagesClient{c.Client}
}

// Machine returns a Compute client used for accessing functions pertaining to
// machine functionality in the Triton API.
func (c *ComputeClient) Instances() *InstancesClient {
	return &InstancesClient{c.Client}
}

// Packages returns a Compute client used for accessing functions pertaining to
// Packages functionality in the Triton API.
func (c *ComputeClient) Packages() *PackagesClient {
	return &PackagesClient{c.Client}
}

// Services returns a Compute client used for accessing functions pertaining to
// Services functionality in the Triton API.
func (c *ComputeClient) Services() *ServicesClient {
	return &ServicesClient{c.Client}
}

// Snapshots returns a Compute client used for accessing functions pertaining to
// Snapshots functionality in the Triton API.
func (c *ComputeClient) Snapshots() *SnapshotsClient {
	return &SnapshotsClient{c.Client}
}
