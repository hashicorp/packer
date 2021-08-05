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

// Bucket represents a single Image bucket on the HCP Packer registry.
type Bucket struct {
	Slug        string
	Description string
	Destination string
	Labels      map[string]string
	Iteration   *Iteration
	client      *Client
}

// NewBucketWithIteration initializes a simple Bucket that can be used publishing Packer build
// images to the HCP Packer registry.
func NewBucketWithIteration(opts IterationOptions) (*Bucket, error) {
	b := Bucket{}

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

// Initialize registers the Bucket b with the configured HCP Packer Registry.
// Upon initialization a Bucket will be upserted to, and new iteration will be created for the build if the configured
// fingerprint has no associated iterations. Lastly, the initialization process with register the builds that need to be
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
		return fmt.Errorf("failed to initialize bucket %q: %w", b.Slug, err)
	}

	return b.initializeIteration(ctx)
}

func (b *Bucket) RegisterBuildForComponent(sourceName string) {
	if b == nil {
		return
	}

	if _, ok := b.Iteration.builds.Load(sourceName); ok {
		return
	}
	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, sourceName)
}

// CreateInitialBuildForIteration will create a build entry on the HCP Packer Registry for the named componentType.
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
		Labels:        make(map[string]string),
		Images:        make(map[string]Image),
	}

	log.Println("[TRACE] creating initial build for component", componentType)
	b.Iteration.builds.Store(componentType, build)

	return nil
}

// UpdateBuildStatus updates the status of a build entry on the HCP Packer registry with its current local status.
func (b *Bucket) UpdateBuildStatus(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {
	if status == models.HashicorpCloudPackerBuildStatusDONE {
		return b.markBuildComplete(ctx, name)
	}

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
			Labels:        buildToUpdate.Labels,
			Status:        status,
		},
	}

	_, err := UpdateBuild(ctx, b.client, buildInput)
	if err != nil {
		return err
	}
	buildToUpdate.Status = status
	b.Iteration.builds.Store(name, buildToUpdate)
	return nil
}

// markBuildComplete should be called to set a build on the HCP Packer registry to DONE.
// Upon a successful call markBuildComplete will publish all images created by the named build,
// and set the registry build to done. A build with no images can not be set to DONE.
func (b *Bucket) markBuildComplete(ctx context.Context, name string) error {
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

	status := models.HashicorpCloudPackerBuildStatusDONE

	if buildToUpdate.Status == status {
		// let's no mess with anything that is already done
		return nil
	}

	buildInput := &models.HashicorpCloudPackerUpdateBuildRequest{
		BuildID: buildToUpdate.ID,
		Updates: &models.HashicorpCloudPackerBuildUpdates{
			PackerRunUUID: buildToUpdate.RunUUID,
			Labels:        buildToUpdate.Labels,
			Status:        status,
		},
	}

	if len(buildToUpdate.Images) == 0 {
		return fmt.Errorf("setting a build to DONE with no published images is not currently support.")
	}

	var providerName string
	images := make([]*models.HashicorpCloudPackerImage, 0, len(buildToUpdate.Images))
	for _, image := range buildToUpdate.Images {
		if providerName == "" {
			providerName = image.ProviderName
		}
		images = append(images, &models.HashicorpCloudPackerImage{ImageID: image.ID, Region: image.ProviderRegion})
	}

	buildInput.Updates.CloudProvider = providerName
	buildInput.Updates.Images = images

	_, err := UpdateBuild(ctx, b.client, buildInput)
	if err != nil {
		return err
	}

	buildToUpdate.Status = status
	b.Iteration.builds.Store(name, buildToUpdate)
	return nil
}

// UpdateImageForBuild appends one or more images artifacts to the build referred to by componentType.
func (b *Bucket) UpdateImageForBuild(componentType string, images ...Image) error {
	return b.Iteration.AddImageToBuild(componentType, images...)
}

// UpdateLabelsForBuild merges the contents of data to the labels associated with the build referred to by componentType.
func (b *Bucket) UpdateLabelsForBuild(componentType string, data map[string]string) error {
	return b.Iteration.AddLabelsToBuild(componentType, data)
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

// createIteration creates an empty iteration for a give bucket on the HCP Packer registry.
// The iteration can then be stored locally and used for tracking build status and images for a running
// Packer build.
func (b *Bucket) createIteration() (*models.HashicorpCloudPackerIteration, error) {
	iterationInput := &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug:  b.Slug,
		Fingerprint: b.Iteration.Fingerprint,
	}

	iterationResp, err := CreateIteration(context.TODO(), b.client, iterationInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
	}

	if iterationResp == nil {
		return nil, fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
	}

	log.Println("[TRACE] a valid iteration for build was created with the Id", iterationResp.ID)
	return iterationResp, nil
}

// initializeIteration populates the bucket iteration with the details needed for tracking builds for a Packer run.
// If an existing Packer registry iteration exists for the said iteration fingerprint, calling initialize on iteration
// that doesn't yet exist will call createIteration to create the entry on the HCP packer registry for the given bucket.
// All build details will be created (if they don't exists) and added to b.Iteration.builds for tracking during runtime.
func (b *Bucket) initializeIteration(ctx context.Context) error {

	// load existing iteration using fingerprint.
	iterationResp, err := GetIteration(ctx, b.client, b.Slug, b.Iteration.Fingerprint)
	if checkErrorCode(err, codes.Aborted) {
		//probably means Iteration doesn't exist need a way to check the error
		iterationResp, err = b.createIteration()
	}

	if err != nil {
		return fmt.Errorf("failed to initialize iteration for fingerprint %s: %s", b.Iteration.Fingerprint, err)
	}

	if iterationResp == nil {
		return fmt.Errorf("failed to initialize iteration details for Bucket %s with error: %w", b.Slug, err)
	}

	log.Println("[TRACE] a valid iteration was retrieved with the id", iterationResp.ID)
	b.Iteration.ID = iterationResp.ID

	// If the iteration is completed and there are no new builds to add, Packer
	// should exit and inform the user that artifacts already exists for the
	// fingerprint associated with the iteration.
	if iterationResp.Complete {
		return fmt.Errorf("This iteration associated to the fingerprint %s is complete. "+
			"If you wish to add a new build to this image a new iteration must be created by changing the build fingerprint.", b.Iteration.Fingerprint)
	}

	// list all this iteration's builds so we can figure out which ones
	// we want to run against. TODO: pagination?
	existingBuilds, err := ListBuilds(ctx, b.client, b.Slug, iterationResp.ID)
	if err != nil {
		return fmt.Errorf("error listing builds for this existing iteration: %s", err)
	}

	var toCreate []string
	for _, expected := range b.Iteration.expectedBuilds {
		var found bool
		for _, existing := range existingBuilds {
			if existing.ComponentType == expected {
				found = true
				build := &Build{
					ID:            existing.ID,
					ComponentType: existing.ComponentType,
					RunUUID:       b.Iteration.RunUUID,
					Status:        existing.Status,
					Labels:        existing.Labels,
				}
				b.Iteration.builds.Store(existing.ComponentType, build)

				// TODO validate that this is safe. For builds that are DONE do we want to keep track of completed things
				// potential issue on updating the status of a build that is already DONE. Is this possible?
				for _, image := range existing.Images {

					err := b.UpdateImageForBuild(existing.ComponentType, Image{
						ID:             image.ImageID,
						ProviderRegion: image.Region,
					})

					if err != nil {
						log.Printf("[TRACE] unable to load existing images for %q: %v", existing.ComponentType, err)
					}
				}

				log.Printf("[TRACE] a build of component type %s already exists; skipping the create call", expected)
				break
			}
		}

		if !found {
			missingbuild := expected
			toCreate = append(toCreate, missingbuild)
		}
	}

	var errs *multierror.Error
	var wg sync.WaitGroup

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
				log.Printf("[TRACE] build %s already exists in Packer registry, continuing...", name)
				return
			}

			errs = multierror.Append(errs, err)
		}(buildName)
	}
	wg.Wait()

	return errs.ErrorOrNil()
}

// IsExpectingBuildForComponent returns true if the component referenced by buildName is part of the iteration
// and is not marked as DONE on the HCP Packer registry.
func (b *Bucket) IsExpectingBuildForComponent(buildName string) bool {

	v, ok := b.Iteration.builds.Load(buildName)
	if !ok {
		return false
	}

	build := v.(*Build)
	hasBuildID := build.ID != ""
	hasImages := len(build.Images) == 0
	isNotDone := build.Status != models.HashicorpCloudPackerBuildStatusDONE

	return hasBuildID && hasImages && isNotDone
}
