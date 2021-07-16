package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

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

// NewBucketWithIteration initializes a simple Bucket that can be used for tracking Packer
// related build bits.
func NewBucketWithIteration(opts IterationOptions) *Bucket {
	b := Bucket{
		Labels: map[string]string{
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

// Initialize registers the Bucket b with the configured HCP Packer Registry.
// Upon initialization a Bucket will be upserted to, a new iteration will be created for the build if the configured
// fingerprint has not associated builds. Lastly, the initialization process with registered the builds that need to be
// completed before an iteration can be marked as DONE.
//
// b.Initialize() must be called before any data can be published to the configured HCP Packer Registry.
func (b *Bucket) Initialize(ctx context.Context) error {
	// NOOP
	if b == nil {
		return nil
	}

	// Bucket already initialized
	if b.client != nil {
		return nil
	}

	if b.client == nil {
		if err := b.connect(); err != nil {
			return err
		}
	}

	bucketInput := &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug:  b.Slug,
		Description: b.Description,
		Labels:      b.Labels,
	}

	err := UpsertBucket(ctx, b.client, bucketInput)
	if err != nil {
		return fmt.Errorf("failed to initialize iteration for bucket %q: %w", b.Slug, err)
	}

	// Create/find iteration logic to be added

	iterationInput := &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug:  b.Slug,
		Fingerprint: b.Iteration.Fingerprint,
	}

	id, err := CreateIteration(ctx, b.client, iterationInput)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
	}

	b.Iteration.ID = id

	var wg sync.WaitGroup
	for k := range b.Iteration.builds.m {
		log.Printf("[TRACE] initialize running build for %q", k)
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			b.PublishBuildStatus(ctx, name, models.HashicorpCloudPackerBuildStatusUNSET)
		}(k)
	}
	wg.Wait()
	return nil
}

// connect initializes a client connection to a remote HCP Packer Registry service on HCP.
// Upon a successful connection the initialized client is persisted on the Bucket b for later usage.
func (b *Bucket) connect() error {
	// NOOP
	if b == nil {
		return nil
	}

	registryClient, err := NewClient(b.Config)
	if err != nil {
		return errors.New("Failed to create client connection to artifact registry: " + err.Error())
	}
	b.client = registryClient
	return nil
}

func (b *Bucket) AddBuildForSource(sourceName string) {
	if b == nil {
		return
	}

	if _, ok := b.Iteration.builds.m[sourceName]; ok {
		return
	}

	build := &Build{
		ComponentType: sourceName,
		RunUUID:       b.Iteration.RunUUID,
		PARtifacts:    make([]PARtifact, 0),
	}
	b.Iteration.builds.Lock()
	b.Iteration.builds.m[sourceName] = build
	b.Iteration.builds.Unlock()
}

func (b *Bucket) PublishBuildStatus(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {
	// NOOP
	if b == nil {
		return nil
	}

	// Lets check if we have something already for this build
	existingBuild, ok := b.Iteration.builds.m[name]
	if ok && existingBuild.ID != "" {
		buildInput := &models.HashicorpCloudPackerUpdateBuildRequest{
			BuildID: existingBuild.ID,
			Updates: &models.HashicorpCloudPackerBuildUpdates{
				PackerRunUUID: existingBuild.RunUUID,
				Status:        &status,
			},
		}

		// Possible bug of being able to set DONE with no RunUUID being set.
		if status == models.HashicorpCloudPackerBuildStatusDONE {
			images := make([]*models.HashicorpCloudPackerImage, 0, len(existingBuild.PARtifacts))
			for _, partifact := range existingBuild.PARtifacts {
				images = append(images, &models.HashicorpCloudPackerImage{ImageID: partifact.ID, Region: partifact.ProviderRegion})
			}
			//TODO pass along cloud provider
			buildInput.Updates.Images = images
		}

		_, err := UpdateBuild(ctx, b.client, buildInput)
		if err != nil {
			return err
		}
		b.Iteration.builds.Lock()
		existingBuild.Status = status
		b.Iteration.builds.m[name] = existingBuild
		b.Iteration.builds.Unlock()
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

	b.Iteration.builds.Lock()
	b.Iteration.builds.m[name] = build
	b.Iteration.builds.Unlock()
	return nil
}

func (b *Bucket) AddBuildArtifact(ctx context.Context, name string, partifacts ...PARtifact) error {
	// NOOP
	if b == nil {
		return nil
	}

	build, ok := b.Iteration.builds.m[name]
	if !ok {
		return errors.New("no associated build found for the name " + name)
	}

	for _, artifact := range partifacts {
		if build.CloudProvider == "" {
			build.CloudProvider = artifact.ProviderName
		}
		build.PARtifacts = append(build.PARtifacts, artifact)
	}

	b.Iteration.builds.Lock()
	b.Iteration.builds.m[name] = build
	b.Iteration.builds.Unlock()

	return nil
}

// Load defaults from environment variables
func (b *Bucket) Canonicalize() {
	// NOOP
	if b == nil {
		return
	}

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

	// Set some iteration values. For Packer RunUUID should always be set.
	// Creating a bucket differently? Let's not overwrite a UUID that might be set.
	if b.Iteration.RunUUID == "" {
		b.Iteration.RunUUID = os.Getenv("PACKER_RUN_UUID")
	}

}
