package googlecompute

import (
	"code.google.com/p/google-api-go-client/compute/v1beta16"
)

// GoogleComputeClient represents a GCE client.
type GoogleComputeClient struct {
	ProjectId     string
	Service       *compute.Service
	Zone          string
	clientSecrets *clientSecrets
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
