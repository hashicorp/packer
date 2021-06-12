package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
)

type Bucket struct {
	Slug        string
	Description string
	Labels      map[string]string
}

type Iteration struct {
	Bucket
	Fingerprint  string
	AncestorSlug string
	Author       string
	Labels       map[string]string
	Builds       []Build
	client       *Client
}

type Build struct {
	ComponentType string
	RunUUID       string
	PARtifacts    []PARtifact
}

type PARtifact struct {
	ID                           string
	ProviderName, ProviderRegion string
	Metadata                     map[string]string
}

func NewIteration(bucketSlug string, fingerprint string) *Iteration {
	b := Bucket{Slug: bucketSlug}
	i := Iteration{Bucket: b, Fingerprint: fingerprint}

	return &i
}

func (i *Iteration) Initialize(client *Client) error {
	if client == nil {
		return errors.New("unable to initialize an Iteration without a valid client")
	}
	i.client = client
	params := packer_service.NewGetBucketParamsWithContext(context.Background())
	params.BucketSlug = i.Slug
	params.LocationOrganizationID = i.client.Config.OrganizationID
	params.LocationProjectID = i.client.Config.ProjectID

	ib, err := i.client.Packer.GetBucket(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return fmt.Errorf("failed to GetImageBucket with error: %s", err)
	}

	log.Printf(`[DEBUG] Found an image with the name %s
LastUpdated: %v
Description: %s
Labels: %v
	\n`, i.Slug, ib.Payload.Image.UpdatedAt, ib.Payload.Image.Description, ib.Payload.Image.Labels)

	i.Slug = ib.Payload.Image.Slug

	return nil
}

func (i *Iteration) BucketPath() string {
	return strings.Join([]string{i.client.Config.OrganizationID, "projects", i.client.Config.ProjectID, i.Bucket.Slug}, "/")
}
