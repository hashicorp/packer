package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"os"

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

func NewIteration(opts IterationOptions) *Iteration {
	i := Iteration{}

	if !opts.UseGitBackend {
		i.Author = os.Getenv("USER")
		i.Fingerprint = "dd5540f6d9d05614134da27c44062575b66e503d"
	}

	return &i
}

func NewIterationWithBucket(bucketSlug string, opts IterationOptions) *Iteration {
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

	bucketInput := &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug:  i.Bucket.Slug,
		Description: i.Bucket.Description,
		Labels:      i.Bucket.Labels,
	}

	err := UpsertBucket(ctx, i.client, bucketInput)
	if err != nil {
		return fmt.Errorf("failed to initialize iteration for bucket %q: %w", i.BucketPath(), err)
	}

	// Create/find iteration
	iterationInput := &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug: i.Bucket.Slug,
		Iteration: &models.HashicorpCloudPackerIteration{
			BucketSlug:  i.Bucket.Slug,
			AuthorID:    i.Author,
			Fingerprint: i.Fingerprint,
		},
	}
	iterationID, err := CreateIteration(ctx, i.client, iterationInput)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return fmt.Errorf("failed to CreateIteration for Bucket %s with error: %w", i.Bucket.Slug, err)
	}

	i.ID = iterationID

	return nil
}

func (i *Iteration) BucketPath() string {
	return i.Bucket.Slug
}
