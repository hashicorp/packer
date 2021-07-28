package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"github.com/hashicorp/packer/internal/packer_registry/env"
	"google.golang.org/grpc/codes"
)

type Bucket struct {
	Slug        string
	Description string
	Destination string
	Labels      map[string]string
	*Iteration
	client *Client
}

// NewBucketWithIteration initializes a simple Bucket that can be used for tracking Packer
// related build bits.
func NewBucketWithIteration(opts IterationOptions) (*Bucket, error) {
	b := Bucket{
		Labels: map[string]string{
			"CreatedBy": "Packer",
		},
	}

	i, err := NewIteration(opts)
	if err != nil {
		return nil, err
	}
	b.Iteration = i

	return &b, nil
}

func (b *Bucket) Validate() error {
	if b.Slug == "" {
		return fmt.Errorf("no Packer bucket name defined; either the environment variable %q is undefined or the HCL configuration has no build name", env.HCPPackerBucket)
	}
	return nil
}

// Initialize registers the Bucket b with the configured HCP Packer Registry.
// Upon initialization a Bucket will be upserted to, a new iteration will be created for the build if the configured
// fingerprint has not associated builds. Lastly, the initialization process with registered the builds that need to be
// completed before an iteration can be marked as DONE.
//
// b.Initialize() must be called before any data can be published to the configured HCP Packer Registry.
// TODO ensure initialize can only be called once
func (b *Bucket) Initialize(ctx context.Context) error {

	if err := b.connect(); err != nil {
		return err
	}

	b.Destination = fmt.Sprintf("%s/%s", b.client.OrganizationID, b.client.ProjectID)

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

	var existingBuilds []*models.HashicorpCloudPackerBuild

	id, err := CreateIteration(ctx, b.client, iterationInput)
	if err != nil {
		if !checkErrorCode(err, codes.AlreadyExists) {
			return fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
		}
		// load existing iteration using fingerprint.
		id, err = GetIteration(ctx, b.client, b.Slug, b.Iteration.Fingerprint)
		if err != nil {
			return fmt.Errorf("Error loading the existing iteration for fingerprint %s", b.Iteration.Fingerprint)
		}
		log.Println("[TRACE] a valid iteration was retrieved with the id", id)
		// list all this iteration's builds so we can figure out which ones
		// we want to run against. TODO: pagination?
		existingBuilds, err = ListBuilds(ctx, b.client, b.Slug, id)
		if err != nil {
			return fmt.Errorf("Error listing builds for this existing iteration: %s", err)
		}

	} else {
		log.Println("[TRACE] a valid iteration for build was created with the Id", id)
	}

	b.Iteration.ID = id

	var errs *multierror.Error
	var wg sync.WaitGroup

	//[]string{a,b}
	// remove the iteration's expected builds that already exist.
	var toCreate []string
	for _, expected := range b.Iteration.registeredBuilds {
		var found bool
		for _, existing := range existingBuilds {
			if existing.ComponentType == expected {
				found = true
				log.Printf("build of component type %s already exists; skipping the create call", expected)

				if existing.Status == models.HashicorpCloudPackerBuildStatusDONE {
					// We also need to remove the builds that are _complete_ from the
					// Iteration's expectedBuilds so we don't overwrite them.
					//b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds[:i], b.Iteration.expectedBuilds[i+1:]...)
					//log.Printf("build of component type %s is already marked DONE; removing the build from the HCP Packer Registry expected builds.", expected)
					break
				}

				// Lets create a build entry for any existing builds we want to update in this run
				b.Iteration.builds.Store(existing.ComponentType, &Build{
					ID:            existing.ID,
					ComponentType: existing.ComponentType,
					RunUUID:       b.Iteration.RunUUID,
					Status:        models.HashicorpCloudPackerBuildStatusUNSET,
					Metadata:      make(map[string]string),
					Images:        make([]Image, 0),
				})
				break
			}
		}
		if !found {
			missingbuild := expected
			toCreate = append(toCreate, missingbuild)
		}
	}

	//TODO handle the case were an iteration is completed. Packer should exit an inform the user
	// that artifacts already exists for the fingerprint associated with the iteration.

	for _, buildName := range toCreate {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			log.Printf("[TRACE] registering build with iteration for %q.", name)
			// Need a way to handle skipping builds that were already created.
			// TODO when we load an existing iteration we will probably have a build Id so we should skip.
			// we also need to bubble up the errors here.
			err := b.CreateInitialBuildForIteration(ctx, name)
			if checkErrorCode(err, codes.AlreadyExists) {
				// Check whether build is complete, and if so, skip it.
				log.Printf("[TRACE] build %s already exists in PAR, continuing...", name)
				return
			}

			errs = multierror.Append(errs, err)
		}(buildName)
	}
	wg.Wait()

	return errs.ErrorOrNil()
}

// connect initializes a client connection to a remote HCP Packer Registry service on HCP.
// Upon a successful connection the initialized client is persisted on the Bucket b for later usage.
func (b *Bucket) connect() error {
	if b.client != nil {
		return nil
	}

	registryClient, err := NewClient()
	if err != nil {
		return errors.New("Failed to create client connection to artifact registry: " + err.Error())
	}
	b.client = registryClient
	return nil
}

func (b *Bucket) RegisterBuildForComponent(sourceName string) {
	if b == nil {
		return
	}

	if _, ok := b.Iteration.builds.Load(sourceName); ok {
		return
	}
	b.Iteration.registeredBuilds = append(b.Iteration.registeredBuilds, sourceName)
}

func (b *Bucket) PublishBuildStatus(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {
	// Lets check if we have something already for this build
	build, ok := b.Iteration.builds.Load(name)
	if !ok {
		return fmt.Errorf("no build for the component %q associated to the iteration %q", name, b.Iteration.ID)
	}

	buildToUpdate, ok := build.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", name)
	}

	if buildToUpdate.ID == "" {
		return fmt.Errorf("the build for the component %q does not have a valid id", name)
	}

	buildInput := &models.HashicorpCloudPackerUpdateBuildRequest{
		BuildID: buildToUpdate.ID,
		Updates: &models.HashicorpCloudPackerBuildUpdates{
			PackerRunUUID: buildToUpdate.RunUUID,
			Labels:        buildToUpdate.Metadata,
			Status:        status,
		},
	}

	// Possible bug of being able to set DONE with no RunUUID being set.
	if status == models.HashicorpCloudPackerBuildStatusDONE {
		images := make([]*models.HashicorpCloudPackerImage, 0, len(buildToUpdate.Images))
		var providerName string
		for _, partifact := range buildToUpdate.Images {
			if providerName == "" {
				providerName = partifact.ProviderName
			}
			images = append(images, &models.HashicorpCloudPackerImage{ImageID: partifact.ID, Region: partifact.ProviderRegion})
		}
		buildInput.Updates.CloudProvider = providerName
		buildInput.Updates.Images = images
	}

	_, err := UpdateBuild(ctx, b.client, buildInput)
	if err != nil {
		return err
	}
	buildToUpdate.Status = status
	b.Iteration.builds.Store(name, buildToUpdate)
	return nil
}

// CreateInitialBuildForIteration will create a build record on the HCP Packer Registry for named componentType.
// This initial creation is needed so that Packer can properly track when an iteration is complete.
func (b *Bucket) CreateInitialBuildForIteration(ctx context.Context, componentType string) error {

	status := models.HashicorpCloudPackerBuildStatusUNSET
	buildInput := &models.HashicorpCloudPackerCreateBuildRequest{
		BucketSlug:  b.Slug,
		Fingerprint: b.Iteration.Fingerprint,
		Build: &models.HashicorpCloudPackerBuild{
			ComponentType: componentType,
			IterationID:   b.Iteration.ID,
			PackerRunUUID: b.Iteration.RunUUID,
			Status:        status,
		},
	}

	id, err := CreateBuild(ctx, b.client, buildInput)
	if err != nil {
		return err
	}

	build := &Build{
		ID:            id,
		ComponentType: componentType,
		RunUUID:       b.Iteration.RunUUID,
		Status:        status,
		Metadata:      make(map[string]string),
		Images:        make([]Image, 0),
	}

	log.Println("[TRACE] creating initial build for component", componentType)
	b.Iteration.builds.Store(componentType, build)

	return nil
}

// AddImageToBuild appends one or more images artifacts to the build referred to by componentType.
func (b *Bucket) AddImageToBuild(componentType string, images ...Image) error {
	existingBuild, ok := b.Iteration.builds.Load(componentType)
	if !ok {
		return errors.New("no associated build found for the name " + componentType)
	}

	build, ok := existingBuild.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", componentType)
	}

	for _, artifact := range images {
		if build.CloudProvider == "" {
			build.CloudProvider = artifact.ProviderName
		}
		build.Images = append(build.Images, artifact)
	}

	b.Iteration.builds.Store(componentType, build)

	return nil
}

// AddBuildMetadata merges the contents of data to the labels associated with the build referred to by componentType.
func (b *Bucket) AddBuildMetadata(componentType string, data map[string]string) error {
	existingBuild, ok := b.Iteration.builds.Load(componentType)
	if !ok {
		return errors.New("no associated build found for the name " + componentType)
	}

	build, ok := existingBuild.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", componentType)
	}

	for k, v := range data {
		if _, ok := build.Metadata[k]; ok {
			continue
		}
		build.Metadata[k] = v
	}

	b.Iteration.builds.Store(componentType, build)

	return nil
}

// Load defaults from environment variables
func (b *Bucket) LoadDefaultSettingsFromEnv() {
	// Configure HCP Packer Registry destination
	if b.Slug == "" {
		b.Slug = os.Getenv(env.HCPPackerBucket)
	}

	// Set some iteration values. For Packer RunUUID should always be set.
	// Creating an iteration differently? Let's not overwrite a UUID that might be set.
	if b.Iteration.RunUUID == "" {
		b.Iteration.RunUUID = os.Getenv("PACKER_RUN_UUID")
	}

}
