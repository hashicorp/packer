package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"google.golang.org/grpc/codes"
)

type Iteration struct {
	Bucket
	ID           string
	Fingerprint  string
	AncestorSlug string
	Author       string
	Labels       map[string]string
	Builds       []Build
	client       *Client
}

type IterationOptions struct {
	UseGitBackend bool
}

func NewIteration(bucketSlug string, opts IterationOptions) *Iteration {
	b := Bucket{
		Slug:        bucketSlug,
		Description: "Base debian image to rule all clouds.",
		Labels: map[string]string{
			"Team":      "Dev",
			"ManagedBy": "Packer",
		},
	}

	i := Iteration{Bucket: b}

	if !opts.UseGitBackend {
		i.Author = os.Getenv("USER")
		i.Fingerprint = "dd5540f6d9d05614134da27c44062575b66e503d"
	}

	return &i
}

func (i *Iteration) Initialize(ctx context.Context, client *Client) error {
	if client == nil {
		return errors.New("unable to initialize an Iteration without a valid client")
	}
	i.client = client

	// Create bucket if exist we continue as is, eventually we want to treat this like an upsert

	params := packer_service.NewCreateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = i.client.Config.OrganizationID
	params.LocationProjectID = i.client.Config.ProjectID
	params.Body = &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug:  i.Bucket.Slug,
		Description: i.Bucket.Description,
		Labels:      i.Bucket.Labels,
	}

	/*
		params := packer_service.NewGetBucketParamsWithContext(context.Background())
		params.BucketSlug = i.Slug
		params.LocationOrganizationID = i.client.Config.OrganizationID
		params.LocationProjectID = i.client.Config.ProjectID
	*/
	_, err := i.client.Packer.CreateBucket(params, nil, func(*runtime.ClientOperation) {})

	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return fmt.Errorf("failed to CreateImageBucket with error: %w", err)
	}

	// Create/find iteration
	{
		params := packer_service.NewCreateIterationParamsWithContext(ctx)
		params.LocationOrganizationID = i.client.Config.OrganizationID
		params.LocationProjectID = i.client.Config.ProjectID
		params.Body = &models.HashicorpCloudPackerCreateIterationRequest{
			BucketSlug: i.Bucket.Slug,
		}
		it, err := i.client.Packer.CreateIteration(params, nil, func(*runtime.ClientOperation) {})

		if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
			return fmt.Errorf("failed to CreateIteration for Bucket %s with error: %w", i.Bucket.Slug, err)
		}

		i.ID = it.Payload.Iteration.ID
	}

	return nil
}

func (i *Iteration) BucketPath() string {
	return strings.Join([]string{i.client.Config.OrganizationID, "projects", i.client.Config.ProjectID, i.Bucket.Slug}, "/")
}
