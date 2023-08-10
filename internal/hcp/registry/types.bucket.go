// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/internal/hcp/api"
	"github.com/hashicorp/packer/internal/hcp/env"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc/codes"
)

// HeartbeatPeriod dictates how often a heartbeat is sent to HCP to signal a
// build is still alive.
const HeartbeatPeriod = 2 * time.Minute

// Bucket represents a single Image bucket on the HCP Packer registry.
type Bucket struct {
	Slug                           string
	Description                    string
	Destination                    string
	BucketLabels                   map[string]string
	BuildLabels                    map[string]string
	SourceImagesToParentIterations map[string]ParentIteration
	RunningBuilds                  map[string]chan struct{}
	Iteration                      *Iteration
	client                         *api.Client
}

type ParentIteration struct {
	IterationID string
	ChannelID   string
}

// NewBucketWithIteration initializes a simple Bucket that can be used publishing Packer build
// images to the HCP Packer registry.
func NewBucketWithIteration() *Bucket {
	b := Bucket{
		BucketLabels:                   make(map[string]string),
		BuildLabels:                    make(map[string]string),
		SourceImagesToParentIterations: make(map[string]ParentIteration),
		RunningBuilds:                  make(map[string]chan struct{}),
	}
	b.Iteration = NewIteration()

	return &b
}

func (b *Bucket) Validate() error {
	if b.Slug == "" {
		return fmt.Errorf("no Packer bucket name defined; either the environment variable %q is undefined or the HCL configuration has no build name", env.HCPPackerBucket)
	}
	return nil
}

// ReadFromHCLBuildBlock reads the information for initialising a Bucket from a HCL2 build block
func (b *Bucket) ReadFromHCLBuildBlock(build *hcl2template.BuildBlock) {
	if b == nil {
		return
	}

	registryBlock := build.HCPPackerRegistry
	if registryBlock == nil {
		return
	}

	b.Description = registryBlock.Description
	b.BucketLabels = registryBlock.BucketLabels
	b.BuildLabels = registryBlock.BuildLabels
	// If there's already a Slug this was set from env variable.
	// In Packer, env variable overrides config values so we keep it that way for consistency.
	if b.Slug == "" && registryBlock.Slug != "" {
		b.Slug = registryBlock.Slug
	}
}

// connect initializes a client connection to a remote HCP Packer Registry service on HCP.
// Upon a successful connection the initialized client is persisted on the Bucket b for later usage.
func (b *Bucket) connect() error {
	if b.client != nil {
		return nil
	}

	registryClient, err := api.NewClient()
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
func (b *Bucket) Initialize(ctx context.Context, templateType models.HashicorpCloudPackerIterationTemplateType) error {

	if err := b.connect(); err != nil {
		return err
	}

	b.Destination = fmt.Sprintf("%s/%s", b.client.OrganizationID, b.client.ProjectID)

	err := b.client.UpsertBucket(ctx, b.Slug, b.Description, b.BucketLabels)
	if err != nil {
		return fmt.Errorf("failed to initialize bucket %q: %w", b.Slug, err)
	}

	return b.initializeIteration(ctx, templateType)
}

func (b *Bucket) RegisterBuildForComponent(sourceName string) {
	if b == nil {
		return
	}

	if ok := b.Iteration.HasBuild(sourceName); ok {
		return
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, sourceName)
}

// CreateInitialBuildForIteration will create a build entry on the HCP Packer Registry for the named componentType.
// This initial creation is needed so that Packer can properly track when an iteration is complete.
func (b *Bucket) CreateInitialBuildForIteration(ctx context.Context, componentType string) error {
	status := models.HashicorpCloudPackerBuildStatusUNSET

	resp, err := b.client.CreateBuild(ctx,
		b.Slug,
		b.Iteration.RunUUID,
		b.Iteration.ID,
		b.Iteration.Fingerprint,
		componentType,
		status,
	)
	if err != nil {
		return err
	}

	build, err := NewBuildFromCloudPackerBuild(resp.Payload.Build)
	if err != nil {
		log.Printf("[TRACE] unable to load created build for %q: %v", componentType, err)
	}

	build.Labels = make(map[string]string)
	build.Images = make(map[string]registryimage.Image)

	// Initial build labels are only pushed to the registry when an actual Packer run is executed on the said build.
	// For example filtered builds (e.g --only or except) will not get the initial build labels until a build is executed on them.
	// Global build label updates to existing builds are handled in PopulateIteration.
	if len(b.BuildLabels) > 0 {
		build.MergeLabels(b.BuildLabels)
	}
	b.Iteration.StoreBuild(componentType, build)

	return nil
}

// UpdateBuildStatus updates the status of a build entry on the HCP Packer registry with its current local status.
// For updating a build status to DONE use CompleteBuild.
func (b *Bucket) UpdateBuildStatus(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {
	if status == models.HashicorpCloudPackerBuildStatusDONE {
		return fmt.Errorf("do not use UpdateBuildStatus for updating to DONE")
	}

	buildToUpdate, err := b.Iteration.Build(name)
	if err != nil {
		return err
	}

	if buildToUpdate.ID == "" {
		return fmt.Errorf("the build for the component %q does not have a valid id", name)
	}

	if buildToUpdate.Status == models.HashicorpCloudPackerBuildStatusDONE {
		return fmt.Errorf("cannot modify status of DONE build %s", name)
	}

	_, err = b.client.UpdateBuild(ctx,
		buildToUpdate.ID,
		buildToUpdate.RunUUID,
		"",
		"",
		"",
		"",
		nil,
		status,
		nil,
	)
	if err != nil {
		return err
	}
	buildToUpdate.Status = status
	b.Iteration.StoreBuild(name, buildToUpdate)
	return nil
}

// markBuildComplete should be called to set a build on the HCP Packer registry to DONE.
// Upon a successful call markBuildComplete will publish all images created by the named build,
// and set the registry build to done. A build with no images can not be set to DONE.
func (b *Bucket) markBuildComplete(ctx context.Context, name string) error {
	buildToUpdate, err := b.Iteration.Build(name)
	if err != nil {
		return err
	}

	if buildToUpdate.ID == "" {
		return fmt.Errorf("the build for the component %q does not have a valid id", name)
	}

	status := models.HashicorpCloudPackerBuildStatusDONE

	if buildToUpdate.Status == status {
		// let's no mess with anything that is already done
		return nil
	}

	if len(buildToUpdate.Images) == 0 {
		return fmt.Errorf("setting a build to DONE with no published images is not currently supported.")
	}

	var providerName, sourceID, sourceIterationID, sourceChannelID string
	images := make([]*models.HashicorpCloudPackerImageCreateBody, 0, len(buildToUpdate.Images))
	for _, image := range buildToUpdate.Images {
		// These values will always be the same for all images in a single build,
		// so we can just set it inside the loop without consequence
		if providerName == "" {
			providerName = image.ProviderName
		}
		if image.SourceImageID != "" {
			sourceID = image.SourceImageID
		}

		// Check if image is using some other HCP Packer image
		if v, ok := b.SourceImagesToParentIterations[image.SourceImageID]; ok {
			sourceIterationID = v.IterationID
			sourceChannelID = v.ChannelID
		}

		images = append(images, &models.HashicorpCloudPackerImageCreateBody{ImageID: image.ImageID, Region: image.ProviderRegion})
	}

	_, err = b.client.UpdateBuild(ctx,
		buildToUpdate.ID,
		buildToUpdate.RunUUID,
		buildToUpdate.CloudProvider,
		sourceID,
		sourceIterationID,
		sourceChannelID,
		buildToUpdate.Labels,
		status,
		images,
	)
	if err != nil {
		return err
	}

	buildToUpdate.Status = status
	b.Iteration.StoreBuild(name, buildToUpdate)
	return nil
}

// UpdateImageForBuild appends one or more images artifacts to the build referred to by componentType.
func (b *Bucket) UpdateImageForBuild(componentType string, images ...registryimage.Image) error {
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
func (b *Bucket) createIteration(templateType models.HashicorpCloudPackerIterationTemplateType) (*models.HashicorpCloudPackerIteration, error) {
	ctx := context.Background()

	if templateType == models.HashicorpCloudPackerIterationTemplateTypeTEMPLATETYPEUNSET {
		return nil, fmt.Errorf("packer error: template type should not be unset when creating an iteration. This is a Packer internal bug which should be reported to the development team for a fix.")
	}

	createIterationResp, err := b.client.CreateIteration(ctx, b.Slug, b.Iteration.Fingerprint, templateType)
	if err != nil {
		return nil, fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
	}

	if createIterationResp == nil {
		return nil, fmt.Errorf("failed to create Iteration for Bucket %s with error: %w", b.Slug, err)
	}

	log.Println("[TRACE] a valid iteration for build was created with the Id", createIterationResp.Payload.Iteration.ID)
	return createIterationResp.Payload.Iteration, nil
}

func (b *Bucket) initializeIteration(ctx context.Context, templateType models.HashicorpCloudPackerIterationTemplateType) error {
	// load existing iteration using fingerprint.
	iteration, err := b.client.GetIteration(ctx, b.Slug, api.GetIteration_byFingerprint(b.Iteration.Fingerprint))
	if api.CheckErrorCode(err, codes.Aborted) {
		// probably means Iteration doesn't exist need a way to check the error
		iteration, err = b.createIteration(templateType)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize iteration for fingerprint %s: %s", b.Iteration.Fingerprint, err)
	}

	if iteration == nil {
		return fmt.Errorf("failed to initialize iteration details for Bucket %s with error: %w", b.Slug, err)
	}

	if iteration.TemplateType != nil &&
		*iteration.TemplateType != models.HashicorpCloudPackerIterationTemplateTypeTEMPLATETYPEUNSET &&
		*iteration.TemplateType != templateType {
		return fmt.Errorf("This iteration was initially created with a %[2]s template. Changing from %[2]s to %[1]s is not supported.",
			templateType, *iteration.TemplateType)
	}

	log.Println("[TRACE] a valid iteration was retrieved with the id", iteration.ID)
	b.Iteration.ID = iteration.ID

	// If the iteration is completed and there are no new builds to add, Packer
	// should exit and inform the user that artifacts already exists for the
	// fingerprint associated with the iteration.
	if iteration.Complete {
		return fmt.Errorf("This iteration associated to the fingerprint %s is complete. "+
			"If you wish to add a new build to this image a new iteration must be created by changing the build fingerprint.", b.Iteration.Fingerprint)
	}

	return nil
}

// populateIteration populates the bucket iteration with the details needed for tracking builds for a Packer run.
// If an existing Packer registry iteration exists for the said iteration fingerprint, calling initialize on iteration
// that doesn't yet exist will call createIteration to create the entry on the HCP packer registry for the given bucket.
// All build details will be created (if they don't exists) and added to b.Iteration.builds for tracking during runtime.
func (b *Bucket) populateIteration(ctx context.Context) error {
	// list all this iteration's builds so we can figure out which ones
	// we want to run against. TODO: pagination?
	existingBuilds, err := b.client.ListBuilds(ctx, b.Slug, b.Iteration.ID)
	if err != nil {
		return fmt.Errorf("error listing builds for this existing iteration: %s", err)
	}

	var toCreate []string
	for _, expected := range b.Iteration.expectedBuilds {
		var found bool
		for _, existing := range existingBuilds {

			if existing.ComponentType == expected {
				found = true
				build, err := NewBuildFromCloudPackerBuild(existing)
				if err != nil {
					return fmt.Errorf("Unable to load existing build for %q: %v", existing.ComponentType, err)
				}

				// When running against an existing build the Packer RunUUID is most likely different.
				// We capture that difference here to know that the image was created in a different Packer run.
				build.RunUUID = b.Iteration.RunUUID

				// When bucket build labels represent some dynamic data set, possibly set via some user variable,
				//  we need to make sure that any newly executed builds get the labels at runtime.
				if build.IsNotDone() && len(b.BuildLabels) > 0 {
					build.MergeLabels(b.BuildLabels)
				}

				log.Printf("[TRACE] a build of component type %s already exists; skipping the create call", expected)
				b.Iteration.StoreBuild(existing.ComponentType, build)

				break
			}
		}

		if !found {
			missingbuild := expected
			toCreate = append(toCreate, missingbuild)
		}
	}

	if len(toCreate) == 0 {
		return nil
	}

	var errs *multierror.Error
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, buildName := range toCreate {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			log.Printf("[TRACE] registering build with iteration for %q.", name)
			err := b.CreateInitialBuildForIteration(ctx, name)

			if api.CheckErrorCode(err, codes.AlreadyExists) {
				log.Printf("[TRACE] build %s already exists in Packer registry, continuing...", name)
				return
			}

			if err != nil {
				mu.Lock()
				errs = multierror.Append(errs, err)
				mu.Unlock()
			}
		}(buildName)
	}
	wg.Wait()

	return errs.ErrorOrNil()
}

// IsExpectingBuildForComponent returns true if the component referenced by buildName is part of the iteration
// and is not marked as DONE on the HCP Packer registry.
func (b *Bucket) IsExpectingBuildForComponent(buildName string) bool {
	if !b.Iteration.HasBuild(buildName) {
		return false
	}

	build, err := b.Iteration.Build(buildName)
	if err != nil {
		return false
	}

	return build.IsNotDone()
}

// HeartbeatBuild periodically sends status updates for the build
//
// This lets HCP infer that a build is still running and should not be marked
// as cancelled by the HCP Packer registry service.
//
// Usage: defer (b.HeartbeatBuild(ctx, build, period))()
func (b *Bucket) HeartbeatBuild(ctx context.Context, build string) (func(), error) {
	buildToUpdate, err := b.Iteration.Build(build)
	if err != nil {
		return nil, err
	}

	heartbeatChan := make(chan struct{})
	go func() {
		log.Printf("[TRACE] starting heartbeats")

		tick := time.NewTicker(HeartbeatPeriod)

	outHeartbeats:
		for {
			select {
			case <-heartbeatChan:
				tick.Stop()
				break outHeartbeats
			case <-ctx.Done():
				tick.Stop()
				break outHeartbeats
			case <-tick.C:
				_, err = b.client.UpdateBuild(ctx,
					buildToUpdate.ID,
					buildToUpdate.RunUUID,
					"",
					"",
					"",
					"",
					nil,
					models.HashicorpCloudPackerBuildStatusRUNNING,
					nil,
				)
				if err != nil {
					log.Printf("[ERROR] failed to send heartbeat for build %q: %s", build, err)
				} else {
					log.Printf("[TRACE] updating build status for %q to running", build)
				}
			}
		}

		log.Printf("[TRACE] stopped heartbeating build %s", build)
	}()
	return func() {
		close(heartbeatChan)
	}, nil
}

func (b *Bucket) startBuild(ctx context.Context, buildName string) error {
	if !b.IsExpectingBuildForComponent(buildName) {
		return &ErrBuildAlreadyDone{
			Message: "build is already done",
		}
	}

	err := b.UpdateBuildStatus(ctx, buildName, models.HashicorpCloudPackerBuildStatusRUNNING)
	if err != nil {
		return fmt.Errorf("failed to update HCP Packer registry status for %q: %s", buildName, err)
	}

	cleanupHeartbeat, err := b.HeartbeatBuild(ctx, buildName)
	if err != nil {
		log.Printf("[ERROR] failed to start heartbeat function")
	}

	buildDone := make(chan struct{}, 1)
	go func() {
		log.Printf("[TRACE] waiting for heartbeat completion")
		select {
		case <-ctx.Done():
			cleanupHeartbeat()
			err := b.UpdateBuildStatus(
				context.Background(),
				buildName,
				models.HashicorpCloudPackerBuildStatusCANCELLED)
			if err != nil {
				log.Printf(
					"[ERROR] failed to update HCP Packer registry status for %q: %s",
					buildName,
					err)
			}
		case <-buildDone:
			cleanupHeartbeat()
		}
		log.Printf("[TRACE] done waiting for heartbeat completion")
	}()

	b.RunningBuilds[buildName] = buildDone

	return nil
}

func (b *Bucket) completeBuild(
	ctx context.Context,
	buildName string,
	artifacts []packer.Artifact,
	buildErr error,
) ([]packer.Artifact, error) {
	doneCh, ok := b.RunningBuilds[buildName]
	if !ok {
		log.Print("[ERROR] done build does not have an entry in the heartbeat table, state will be inconsistent.")

	} else {
		log.Printf("[TRACE] signal stopping heartbeats")
		// Stop heartbeating
		doneCh <- struct{}{}
		log.Printf("[TRACE] stopped heartbeats")
	}

	if buildErr != nil {
		status := models.HashicorpCloudPackerBuildStatusFAILED
		if ctx.Err() != nil {
			status = models.HashicorpCloudPackerBuildStatusCANCELLED
		}
		err := b.UpdateBuildStatus(context.Background(), buildName, status)
		if err != nil {
			log.Printf("[ERROR] failed to update build %q status to FAILED: %s", buildName, err)
		}
		return artifacts, fmt.Errorf("build failed, not uploading artifacts")
	}

	for _, art := range artifacts {
		var images []registryimage.Image
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           &images,
			WeaklyTypedInput: true,
			ErrorUnused:      false,
		})
		if err != nil {
			return artifacts, fmt.Errorf(
				"failed to create decoder for HCP Packer registry image: %w",
				err)
		}

		state := art.State(registryimage.ArtifactStateURI)
		err = decoder.Decode(state)
		if err != nil {
			return artifacts, fmt.Errorf(
				"failed to obtain HCP Packer registry image from post-processor artifact: %w",
				err)
		}
		log.Printf("[TRACE] updating artifacts for build %q", buildName)
		err = b.UpdateImageForBuild(buildName, images...)

		if err != nil {
			return artifacts, fmt.Errorf("failed to add image artifact for %q: %s", buildName, err)
		}
	}

	parErr := b.markBuildComplete(ctx, buildName)
	if parErr != nil {
		return artifacts, fmt.Errorf(
			"failed to update Packer registry with image artifacts for %q: %s",
			buildName,
			parErr)
	}

	return append(artifacts, &registryArtifact{
		BuildName:   buildName,
		BucketSlug:  b.Slug,
		IterationID: b.Iteration.ID,
	}), nil
}
