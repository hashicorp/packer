package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"github.com/hashicorp/packer/internal/packer_registry/env"
	"google.golang.org/grpc/codes"
)

type Bucket struct {
	Slug        string
	Description string
	Destination string
	Labels      map[string]string
	Config      ClientConfig
	*Iteration
	client *Client
}

func NewBucketWithIteration(opts IterationOptions) *Bucket {
	b := Bucket{
		Description: "Base alpine to rule all clouds.",
		Labels: map[string]string{
			"Team":      "Dev",
			"ManagedBy": "Packer",
		},
	}

	i := NewIteration(opts)
	b.Iteration = i

	return &b
}

func (b *Bucket) Validate() error {
	if b.Slug == "" {
		return fmt.Errorf("no Packer bucket name defined; either the environment variable %q is undefined or the HCL configuration has no build name", env.HCPPackerBucket)
	}

	if b.Destination == "" {
		return fmt.Errorf("no Packer registry defined; either the environment variable %q is undefined or the HCL configuration has no build name", env.HCPPackerRegistry)
	}

	return nil
}

func (b *Bucket) Connect() error {
	registryClient, err := NewClient(b.Config)
	if err != nil {
		return errors.New("Failed to create client connection to artifact registry: " + err.Error())
	}
	b.client = registryClient
	return nil
}

func (b *Bucket) Initialize(ctx context.Context) error {
	bucketInput := &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug:  b.Slug,
		Description: b.Description,
		Labels:      b.Labels,
	}

	err := UpsertBucket(ctx, b.client, bucketInput)
	if err != nil {
		return fmt.Errorf("failed to initialize iteration for bucket %q: %w", b.Slug, err)
	}

	// Create/find iteration
	iterationInput := &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug:  b.Slug,
		Fingerprint: b.Iteration.Fingerprint,
	}

	iterationID, err := CreateIteration(ctx, b.client, iterationInput)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return fmt.Errorf("failed to CreateIteration for Bucket %s with error: %w", b.Slug, err)
	}

	b.Iteration.ID = iterationID

	return nil
}

func (b *Bucket) UpsertBuild(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {
	buildInput := &models.HashicorpCloudPackerCreateBuildRequest{
		BucketSlug:  b.Slug,
		Fingerprint: b.Iteration.Fingerprint,
		Build: &models.HashicorpCloudPackerBuild{
			ComponentType: name,
			IterationID:   b.Iteration.ID,
			PackerRunUUID: b.Iteration.RunUUID,
			Status:        &status,
		},
	}

	err := UpsertBuild(ctx, b.client, buildInput)
	if err != nil {
		return err
	}

	build := Build{
		ComponentType: name,
		RunUUID:       b.Iteration.RunUUID,
		Status:        status,
		PARtifacts:    NewBuildPARtifacts(),
	}

	b.Iteration.Builds.Lock()
	b.Iteration.Builds.m[name] = &build
	b.Iteration.Builds.Unlock()
	return nil
}

// Load defaults from environment variables
func (b *Bucket) Canonicalize() {

	if b.Config.ClientID == "" {
		b.Config.ClientID = os.Getenv(env.HCPClientID)
	}

	if b.Config.ClientSecret == "" {
		b.Config.ClientSecret = os.Getenv(env.HCPClientSecret)
	}

	// Configure HCP registry destination
	if b.Slug == "" {
		b.Slug = os.Getenv(env.HCPPackerBucket)
	}

	loc := os.Getenv(env.HCPPackerRegistry)
	locParts := strings.Split(loc, "/")
	if len(locParts) != 2 {
		// we want an error here. Or at least when we try to create the registry client we fail
		return
	}
	orgID, projID := locParts[0], locParts[1]

	if b.Destination == "" {
		b.Destination = loc
	}

	if b.Config.OrganizationID == "" {
		b.Config.OrganizationID = orgID
	}

	if b.Config.ProjectID == "" {
		b.Config.ProjectID = projID
	}

}
