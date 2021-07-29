package packer_registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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
	if b.client == nil {
		if err := b.connect(); err != nil {
			return err
		}
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

	var existingBuilds []*models.HashicorpCloudPackerBuild

	// First try retrieving iteration using the current fingerprint, to make
	// sure it doesn't already exist.
	var toCreate []string
	iterationResp, err := GetIteration(ctx, b.client, b.Slug, b.Iteration.Fingerprint)
	if err == nil {
		log.Println("[TRACE] a valid iteration was retrieved from the HCP Packer registry with the id", iterationResp.ID)
		b.Iteration.ID = iterationResp.ID

		// list all this iteration's builds so we can figure out which ones
		// we want to run against. TODO: pagination?
		existingBuilds, err = ListBuilds(ctx, b.client, b.Slug, iterationResp.ID)
		if err != nil {
			return fmt.Errorf("Error listing builds for this existing iteration: %s", err)
		}

		// remove the iteration's expected builds that already exist.
		for _, expected := range b.Iteration.expectedBuilds {
			var found bool
			for i, existing := range existingBuilds {
				if existing.ComponentType == expected {
					found = true
					log.Printf("build of component type %s already exists; skipping the create call", expected)

					if existing.Status == models.HashicorpCloudPackerBuildStatusDONE {
						// We also need to remove the builds that are _complete_ from the
						// Iteration's expectedBuilds so we don't overwrite them.
						b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds[:i], b.Iteration.expectedBuilds[i+1:]...)
						log.Printf("build of component type %s is already marked "+
							"DONE; removing the build from the HCP Packer Registry expected builds.", expected)
						break
					}

					// Lets create a build entry for any existing builds we want to update in this run
					b.Iteration.builds.Store(existing.ComponentType, &Build{
						ID:            existing.ID,
						ComponentType: existing.ComponentType,
						RunUUID:       b.Iteration.RunUUID,
						Status:        models.HashicorpCloudPackerBuildStatusUNSET,
						Metadata:      make(map[string]string),
						PARtifacts:    make([]PARtifact, 0),
					})
					break
				}
			}
			if !found {
				missingbuild := expected
				toCreate = append(toCreate, missingbuild)
			}
		}

		// If the iteration is completed and there are no new builds to add, Packer
		// should exit and inform the user that artifacts already exists for the
		// fingerprint associated with the iteration.
		if iterationResp.Complete && len(toCreate) == 0 {
			return fmt.Errorf("This iteration is complete and this Packer build " +
				"adds no new components. exiting.")
		}
	} else if strings.Contains(err.Error(), "record not found") {
		// Iteration doesn't already exist; create a new one.
		iterationInput := &models.HashicorpCloudPackerCreateIterationRequest{
			BucketSlug:  b.Slug,
			Fingerprint: b.Iteration.Fingerprint,
		}

		id, err := CreateIteration(ctx, b.client, iterationInput)
		log.Println("[TRACE] Created an iteration in the HCP Packer registry with the id", id)
		if err != nil {
			if !checkErrorCode(err, codes.AlreadyExists) {
				return fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
			}
		}
		toCreate = b.Iteration.expectedBuilds
		b.Iteration.ID = id
	} else {
		return fmt.Errorf("Error retrieving iteration with fingerprint %s for build.", b.Iteration.Fingerprint)
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
				// This shouldn't happen because of the logic above.
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
	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, sourceName)
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
		images := make([]*models.HashicorpCloudPackerImage, 0, len(buildToUpdate.PARtifacts))
		var providerName string
		for _, partifact := range buildToUpdate.PARtifacts {
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

// CreateInitialBuildForIteration will create a build record on the Packer registry for named component.
// This initial creation is needed so that Packer can properly track when an iteration is complete.
func (b *Bucket) CreateInitialBuildForIteration(ctx context.Context, name string) error {

	status := models.HashicorpCloudPackerBuildStatusUNSET
	buildInput := &models.HashicorpCloudPackerCreateBuildRequest{
		BucketSlug:  b.Slug,
		Fingerprint: b.Iteration.Fingerprint,
		Build: &models.HashicorpCloudPackerBuild{
			ComponentType: name,
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
		ComponentType: name,
		RunUUID:       b.Iteration.RunUUID,
		Status:        status,
		Metadata:      make(map[string]string),
		PARtifacts:    make([]PARtifact, 0),
	}

	log.Println("[TRACE] creating initial build for component", name)
	b.Iteration.builds.Store(name, build)

	return nil
}

func (b *Bucket) AddBuildArtifact(name string, partifacts ...PARtifact) error {
	existingBuild, ok := b.Iteration.builds.Load(name)
	if !ok {
		return errors.New("no associated build found for the name " + name)
	}

	build, ok := existingBuild.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", name)
	}

	for _, artifact := range partifacts {
		if build.CloudProvider == "" {
			build.CloudProvider = artifact.ProviderName
		}
		build.PARtifacts = append(build.PARtifacts, artifact)
	}

	b.Iteration.builds.Store(name, build)

	return nil
}

func (b *Bucket) AddBuildMetadata(name string, data map[string]string) error {
	existingBuild, ok := b.Iteration.builds.Load(name)
	if !ok {
		return errors.New("no associated build found for the name " + name)
	}

	build, ok := existingBuild.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", name)
	}

	for k, v := range data {
		if _, ok := build.Metadata[k]; ok {
			continue
		}
		build.Metadata[k] = v
	}

	b.Iteration.builds.Store(name, build)

	return nil
}

// Load defaults from environment variables
func (b *Bucket) LoadDefaultSettingsFromEnv() {
	// Configure HCP registry destination
	if b.Slug == "" {
		b.Slug = os.Getenv(env.HCPPackerBucket)
	}

	// Set some iteration values. For Packer RunUUID should always be set.
	// Creating a bucket differently? Let's not overwrite a UUID that might be set.
	if b.Iteration.RunUUID == "" {
		b.Iteration.RunUUID = os.Getenv("PACKER_RUN_UUID")
	}

}
