package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"log"
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

	id, err := CreateIteration(ctx, b.client, iterationInput)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return fmt.Errorf("failed to CreateIteration for Bucket %s with error: %w", b.Slug, err)
	}

	b.Iteration.ID = id

	return nil
}

func (b *Bucket) UpdateBuild(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {

	// Lets check if we have something already for this build
	existingBuild, ok := b.Iteration.Builds.m[name]
	if ok && existingBuild.ID != "" {
		buildInput := &models.HashicorpCloudPackerUpdateBuildRequest{
			BuildID: existingBuild.ID,
			Updates: &models.HashicorpCloudPackerBuildUpdates{
				Status: &status,
			},
		}
		if status == models.HashicorpCloudPackerBuildStatusDONE {
			images := make([]*models.HashicorpCloudPackerImage, 0, len(existingBuild.PARtifacts))
			log.Println("WILKEN we setting some image details for now: " + name)
			for _, partifact := range existingBuild.PARtifacts {
				images = append(images, &models.HashicorpCloudPackerImage{ImageID: partifact.ID, Region: partifact.ProviderRegion})
				log.Printf("WILKEN adding image details for %#v\n", partifact)
			}
			buildInput.Updates.Images = images
		}
		log.Printf("WILKEN calling build update with %#v\n", buildInput)

		_, err := UpdateBuild(ctx, b.client, buildInput)
		if err != nil {
			return err
		}
		b.Iteration.Builds.Lock()
		existingBuild.Status = status
		b.Iteration.Builds.m[name] = existingBuild
		b.Iteration.Builds.Unlock()
		return nil
	}

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

	/*
		switch name {
		case "debian.null.example2":
			buildInput.Build.ID = "01FADSXNA9JA4TKX67CQW09JNW"
		case "debian.null.example":
			buildInput.Build.ID = "01FADSSYVBP01Y1CXJ4JR3H98V"
		}
	*/

	id, err := CreateBuild(ctx, b.client, buildInput)
	if err != nil {
		return err
	}

	build := &Build{
		ID:            id,
		ComponentType: name,
		RunUUID:       b.Iteration.RunUUID,
		Status:        status,
		PARtifacts:    make([]PARtifact, 0),
	}

	b.Iteration.Builds.Lock()
	b.Iteration.Builds.m[name] = build
	b.Iteration.Builds.Unlock()
	return nil
}

func (b *Bucket) AddBuildArtifact(ctx context.Context, name string, partifacts ...PARtifact) error {
	build, ok := b.Iteration.Builds.m[name]
	if !ok {
		return errors.New("no associated build found for the name " + name)
	}

	for _, artifact := range partifacts {
		log.Printf("WILKEN adding a partifact %q => %v\n", name, artifact)
		if build.CloudProvider == "" {
			build.CloudProvider = artifact.ProviderName
		}
		build.PARtifacts = append(build.PARtifacts, artifact)
	}

	b.Iteration.Builds.Lock()
	b.Iteration.Builds.m[name] = build
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
