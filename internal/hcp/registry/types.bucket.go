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
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	packerSDK "github.com/hashicorp/packer-plugin-sdk/packer"
	packerSDKRegistry "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer/hcl2template"
	hcpPackerAPI "github.com/hashicorp/packer/internal/hcp/api"
	"github.com/hashicorp/packer/internal/hcp/env"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc/codes"
)

// HeartbeatPeriod dictates how often a heartbeat is sent to HCP to signal a
// build is still alive.
const HeartbeatPeriod = 2 * time.Minute

// Bucket represents a single bucket on the HCP Packer registry.
type Bucket struct {
	Name                                     string
	Description                              string
	Destination                              string
	BucketLabels                             map[string]string
	BuildLabels                              map[string]string
	SourceExternalIdentifierToParentVersions map[string]ParentVersion
	RunningBuilds                            map[string]chan struct{}
	Version                                  *Version
	client                                   *hcpPackerAPI.Client
}

type ParentVersion struct {
	VersionID string
	ChannelID string
}

// NewBucketWithVersion initializes a simple Bucket that can be used for publishing Packer build artifacts
// to the HCP Packer registry.
func NewBucketWithVersion() *Bucket {
	b := Bucket{
		BucketLabels:                             make(map[string]string),
		BuildLabels:                              make(map[string]string),
		SourceExternalIdentifierToParentVersions: make(map[string]ParentVersion),
		RunningBuilds:                            make(map[string]chan struct{}),
	}
	b.Version = NewVersion()

	return &b
}

func (bucket *Bucket) Validate() error {
	if bucket.Name == "" {
		return fmt.Errorf(
			"no Packer bucket name defined; either the environment variable %q is undefined or "+
				"the HCL configuration has no build name",
			env.HCPPackerBucket,
		)
	}
	return nil
}

// ReadFromHCLBuildBlock reads the information for initialising a Bucket from a HCL2 build block
func (bucket *Bucket) ReadFromHCLBuildBlock(build *hcl2template.BuildBlock) {
	if bucket == nil {
		return
	}

	registryBlock := build.HCPPackerRegistry
	if registryBlock == nil {
		return
	}

	bucket.Description = registryBlock.Description
	bucket.BucketLabels = registryBlock.BucketLabels
	bucket.BuildLabels = registryBlock.BuildLabels
	// If there's already a Name this was set from env variable.
	// In Packer, env variable overrides config values so we keep it that way for consistency.
	if bucket.Name == "" && registryBlock.Slug != "" {
		bucket.Name = registryBlock.Slug
	}
}

// connect initializes a client connection to a remote HCP Packer Registry service on HCP.
// Upon a successful connection the initialized client is persisted on the Bucket b for later usage.
func (bucket *Bucket) connect() error {
	if bucket.client != nil {
		return nil
	}

	registryClient, err := hcpPackerAPI.NewClient()
	if err != nil {
		return errors.New("Failed to create client connection to artifact registry: " + err.Error())
	}
	bucket.client = registryClient
	return nil
}

// Initialize registers the bucket with the configured HCP Packer Registry.
// Upon initialization a Bucket will be upserted to, and new version will be created for the build if the configured
// fingerprint has no associated versions. Lastly, the initialization process with register the builds that need to be
// completed before an version can be marked as DONE.
//
// b.Initialize() must be called before any data can be published to the configured HCP Packer Registry.
// TODO ensure initialize can only be called once
func (bucket *Bucket) Initialize(
	ctx context.Context, templateType hcpPackerModels.HashicorpCloudPacker20230101TemplateType,
) error {

	if err := bucket.connect(); err != nil {
		return err
	}

	bucket.Destination = fmt.Sprintf("%s/%s", bucket.client.OrganizationID, bucket.client.ProjectID)

	err := bucket.client.UpsertBucket(ctx, bucket.Name, bucket.Description, bucket.BucketLabels)
	if err != nil {
		return fmt.Errorf("failed to initialize bucket %q: %w", bucket.Name, err)
	}

	return bucket.initializeVersion(ctx, templateType)
}

func (bucket *Bucket) RegisterBuildForComponent(sourceName string) {
	if bucket == nil {
		return
	}

	if ok := bucket.Version.HasBuild(sourceName); ok {
		return
	}

	bucket.Version.expectedBuilds = append(bucket.Version.expectedBuilds, sourceName)
}

// CreateInitialBuildForVersion will create a build entry on the HCP Packer Registry for the named componentType.
// This initial creation is needed so that Packer can properly track when an version is complete.
func (bucket *Bucket) CreateInitialBuildForVersion(ctx context.Context, componentType string) error {
	status := hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET

	resp, err := bucket.client.CreateBuild(ctx,
		bucket.Name,
		bucket.Version.RunUUID,
		bucket.Version.Fingerprint,
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
	build.Artifacts = make(map[string]packerSDKRegistry.Image)

	// Initial build labels are only pushed to the registry when an actual Packer run is executed on the said build.
	// For example filtered builds (e.g --only or except) will not get the initial build labels until a build is
	// executed on them.
	// Global build label updates to existing builds are handled in PopulateVersion.
	if len(bucket.BuildLabels) > 0 {
		build.MergeLabels(bucket.BuildLabels)
	}
	bucket.Version.StoreBuild(componentType, build)

	return nil
}

// UpdateBuildStatus updates the status of a build entry on the HCP Packer registry with its current local status.
// For updating a build status to DONE use CompleteBuild.
func (bucket *Bucket) UpdateBuildStatus(
	ctx context.Context, name string, status hcpPackerModels.HashicorpCloudPacker20230101BuildStatus,
) error {
	if status == hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE {
		return fmt.Errorf("do not use UpdateBuildStatus for updating to DONE")
	}

	buildToUpdate, err := bucket.Version.Build(name)
	if err != nil {
		return err
	}

	if buildToUpdate.ID == "" {
		return fmt.Errorf("the build for the component %q does not have a valid id", name)
	}

	if buildToUpdate.Status == hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE {
		return fmt.Errorf("cannot modify status of DONE build %s", name)
	}

	_, err = bucket.client.UpdateBuild(ctx,
		bucket.Name,
		bucket.Version.Fingerprint,
		buildToUpdate.ID,
		buildToUpdate.RunUUID,
		"",
		"",
		"",
		"",
		nil,
		status,
		nil,
		nil,
	)
	if err != nil {
		return err
	}
	buildToUpdate.Status = status
	bucket.Version.StoreBuild(name, buildToUpdate)
	return nil
}

// markBuildComplete should be called to set a build on the HCP Packer registry to DONE.
// Upon a successful call markBuildComplete will publish all artifacts created by the named build,
// and set the build to done. A build with no artifacts can not be set to DONE.
func (bucket *Bucket) markBuildComplete(ctx context.Context, name string) error {
	buildToUpdate, err := bucket.Version.Build(name)
	if err != nil {
		return err
	}

	if buildToUpdate.ID == "" {
		return fmt.Errorf("the build for the component %q does not have a valid id", name)
	}

	status := hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE

	if buildToUpdate.Status == status {
		// let's no mess with anything that is already done
		return nil
	}

	if len(buildToUpdate.Artifacts) == 0 {
		return fmt.Errorf("setting a build to DONE with no published artifacts is not currently supported")
	}

	var platformName, sourceID, parentVersionID, parentChannelID string
	artifacts := make([]*hcpPackerModels.HashicorpCloudPacker20230101ArtifactCreateBody, 0, len(buildToUpdate.Artifacts))
	for _, artifact := range buildToUpdate.Artifacts {
		// These values will always be the same for all artifacts in a single build,
		// so we can just set it inside the loop without consequence
		if platformName == "" {
			platformName = artifact.ProviderName
		}
		if artifact.SourceImageID != "" {
			sourceID = artifact.SourceImageID
		}

		// Check if artifact is using some other HCP Packer artifact
		if v, ok := bucket.SourceExternalIdentifierToParentVersions[artifact.SourceImageID]; ok {
			parentVersionID = v.VersionID
			parentChannelID = v.ChannelID
		}

		artifacts = append(
			artifacts,
			&hcpPackerModels.HashicorpCloudPacker20230101ArtifactCreateBody{
				ExternalIdentifier: artifact.ImageID,
				Region:             artifact.ProviderRegion,
			},
		)
	}

	_, err = bucket.client.UpdateBuild(ctx,
		bucket.Name,
		bucket.Version.Fingerprint,
		buildToUpdate.ID,
		buildToUpdate.RunUUID,
		buildToUpdate.Platform,
		sourceID,
		parentVersionID,
		parentChannelID,
		buildToUpdate.Labels,
		status,
		artifacts,
		&buildToUpdate.Metadata,
	)
	if err != nil {
		return err
	}

	buildToUpdate.Status = status
	bucket.Version.StoreBuild(name, buildToUpdate)
	return nil
}

// UpdateArtifactForBuild appends one or more artifacts to the build referred to by componentType.
func (bucket *Bucket) UpdateArtifactForBuild(componentType string, artifacts ...packerSDKRegistry.Image) error {
	return bucket.Version.AddArtifactToBuild(componentType, artifacts...)
}

// UpdateLabelsForBuild merges the contents of data to the labels associated with the build referred to by componentType.
func (bucket *Bucket) UpdateLabelsForBuild(componentType string, data map[string]string) error {
	return bucket.Version.AddLabelsToBuild(componentType, data)
}

// LoadDefaultSettingsFromEnv loads defaults from environment variables
func (bucket *Bucket) LoadDefaultSettingsFromEnv() {
	// Configure HCP Packer Registry destination
	if bucket.Name == "" {
		bucket.Name = os.Getenv(env.HCPPackerBucket)
	}

	// Set some version values. For Packer RunUUID should always be set.
	// Creating an version differently? Let's not overwrite a UUID that might be set.
	if bucket.Version.RunUUID == "" {
		bucket.Version.RunUUID = os.Getenv("PACKER_RUN_UUID")
	}

}

// createVersion creates an empty version for a given bucket on the HCP Packer registry.
// The version can then be stored locally and used for tracking build status and artifacts for a running
// Packer build.
func (bucket *Bucket) createVersion(
	templateType hcpPackerModels.HashicorpCloudPacker20230101TemplateType,
) (*hcpPackerModels.HashicorpCloudPacker20230101Version, error) {
	ctx := context.Background()

	if templateType == hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeTEMPLATETYPEUNSET {
		return nil, fmt.Errorf(
			"packer error: template type should not be unset when creating a version. " +
				"This is a Packer internal bug which should be reported to the development team for a fix",
		)
	}

	createVersionResp, err := bucket.client.CreateVersion(
		ctx, bucket.Name, bucket.Version.Fingerprint, templateType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Version for Bucket %s with error: %w", bucket.Name, err)
	}

	if createVersionResp == nil {
		return nil, fmt.Errorf("failed to create Version for Bucket %s with error: %w", bucket.Name, err)
	}

	log.Println(
		"[TRACE] a valid version for build was created with the Id", createVersionResp.Payload.Version.ID,
	)
	return createVersionResp.Payload.Version, nil
}

func (bucket *Bucket) initializeVersion(
	ctx context.Context, templateType hcpPackerModels.HashicorpCloudPacker20230101TemplateType,
) error {
	// load existing version using fingerprint.
	version, err := bucket.client.GetVersion(ctx, bucket.Name, bucket.Version.Fingerprint)
	if hcpPackerAPI.CheckErrorCode(err, codes.Aborted) {
		// probably means Version doesn't exist need a way to check the error
		version, err = bucket.createVersion(templateType)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize version for fingerprint %s: %s", bucket.Version.Fingerprint, err)
	}

	if version == nil {
		return fmt.Errorf("failed to initialize version details for Bucket %s with error: %w", bucket.Name, err)
	}

	if version.TemplateType != nil &&
		*version.TemplateType != hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeTEMPLATETYPEUNSET &&
		*version.TemplateType != templateType {
		return fmt.Errorf(
			"This version was initially created with a %[2]s template. "+
				"Changing from %[2]s to %[1]s is not supported",
			templateType, *version.TemplateType,
		)
	}

	log.Println(
		"[TRACE] a valid version was retrieved with the id", version.ID,
	)
	bucket.Version.ID = version.ID

	// If the version is completed and there are no new builds to add, Packer
	// should exit and inform the user that artifacts already exists for the
	// fingerprint associated with the version.
	if bucket.client.IsVersionComplete(version) {
		return fmt.Errorf(
			"The version associated to the fingerprint %v is complete. If you wish to add a new build to "+
				"this bucket a new version must be created by changing the fingerprint.",
			bucket.Version.Fingerprint,
		)
	}

	return nil
}

// populateVersion populates the version with the details needed for tracking builds for a Packer run.
// If a version exists for the said fingerprint, calling initialize on version that doesn't yet exist will call
// createVersion to create the entry on the HCP packer registry for the given bucket.
// All build details will be created (if they don't exist) and added to b.Version.builds for tracking during runtime.
func (bucket *Bucket) populateVersion(ctx context.Context) error {
	// list all this version's builds so we can figure out which ones
	// we want to run against. TODO: pagination?
	existingBuilds, err := bucket.client.ListBuilds(ctx, bucket.Name, bucket.Version.Fingerprint)
	if err != nil {
		return fmt.Errorf("error listing builds for this existing version: %s", err)
	}

	var toCreate []string
	for _, expected := range bucket.Version.expectedBuilds {
		var found bool
		for _, existing := range existingBuilds {

			if existing.ComponentType == expected {
				found = true
				build, err := NewBuildFromCloudPackerBuild(existing)
				if err != nil {
					return fmt.Errorf("Unable to load existing build for %q: %v", existing.ComponentType, err)
				}

				// When running against an existing build the Packer RunUUID is most likely different.
				// We capture that difference here to know that the artifacts were created in a different Packer run.
				build.RunUUID = bucket.Version.RunUUID

				// When bucket build labels represent some dynamic data set, possibly set via some user variable,
				//  we need to make sure that any newly executed builds get the labels at runtime.
				if build.IsNotDone() && len(bucket.BuildLabels) > 0 {
					build.MergeLabels(bucket.BuildLabels)
				}

				log.Printf(
					"[TRACE] a build of component type %s already exists; skipping the create call", expected,
				)
				bucket.Version.StoreBuild(existing.ComponentType, build)

				break
			}
		}

		if !found {
			missingBuild := expected
			toCreate = append(toCreate, missingBuild)
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

			log.Printf("[TRACE] registering build with version for %q.", name)
			err := bucket.CreateInitialBuildForVersion(ctx, name)

			if hcpPackerAPI.CheckErrorCode(err, codes.AlreadyExists) {
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

// IsExpectingBuildForComponent returns true if the component referenced by buildName is part of the version
// and is not marked as DONE on the HCP Packer registry.
func (bucket *Bucket) IsExpectingBuildForComponent(buildName string) bool {
	if !bucket.Version.HasBuild(buildName) {
		return false
	}

	build, err := bucket.Version.Build(buildName)
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
func (bucket *Bucket) HeartbeatBuild(ctx context.Context, build string) (func(), error) {
	buildToUpdate, err := bucket.Version.Build(build)
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
				_, err = bucket.client.UpdateBuild(ctx,
					bucket.Name,
					bucket.Version.Fingerprint,
					buildToUpdate.ID,
					buildToUpdate.RunUUID,
					"",
					"",
					"",
					"",
					nil,
					hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING,
					nil,
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

func (bucket *Bucket) startBuild(ctx context.Context, buildName string) error {
	if !bucket.IsExpectingBuildForComponent(buildName) {
		return &ErrBuildAlreadyDone{
			Message: "build is already done",
		}
	}

	err := bucket.UpdateBuildStatus(ctx, buildName, hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING)
	if err != nil {
		return fmt.Errorf("failed to update HCP Packer Build status for %q: %s", buildName, err)
	}

	cleanupHeartbeat, err := bucket.HeartbeatBuild(ctx, buildName)
	if err != nil {
		log.Printf("[ERROR] failed to start heartbeat function")
	}

	buildDone := make(chan struct{}, 1)
	go func() {
		log.Printf("[TRACE] waiting for heartbeat completion")
		select {
		case <-ctx.Done():
			cleanupHeartbeat()
			err := bucket.UpdateBuildStatus(
				context.Background(),
				buildName,
				hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDCANCELLED)
			if err != nil {
				log.Printf(
					"[ERROR] failed to update HCP Packer Build status for %q: %s",
					buildName,
					err)
			}
		case <-buildDone:
			cleanupHeartbeat()
		}
		log.Printf("[TRACE] done waiting for heartbeat completion")
	}()

	bucket.RunningBuilds[buildName] = buildDone

	return nil
}

type NotAHCPArtifactError struct {
	error
}

func (bucket *Bucket) completeBuild(
	ctx context.Context,
	buildName string,
	packerSDKArtifacts []packerSDK.Artifact,
	buildErr error,
) ([]packerSDK.Artifact, error) {
	doneCh, ok := bucket.RunningBuilds[buildName]
	if !ok {
		log.Print("[ERROR] done build does not have an entry in the heartbeat table, state will be inconsistent.")

	} else {
		log.Printf("[TRACE] signal stopping heartbeats")
		// Stop heartbeating
		doneCh <- struct{}{}
		log.Printf("[TRACE] stopped heartbeats")
	}

	if buildErr != nil {
		status := hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDFAILED
		if ctx.Err() != nil {
			status = hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDCANCELLED
		}
		err := bucket.UpdateBuildStatus(context.Background(), buildName, status)
		if err != nil {
			log.Printf("[ERROR] failed to update build %q status to FAILED: %s", buildName, err)
		}
		return packerSDKArtifacts, fmt.Errorf("build failed, not uploading artifacts")
	}

	for _, art := range packerSDKArtifacts {
		var sdkImages []packerSDKRegistry.Image
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           &sdkImages,
			WeaklyTypedInput: true,
			ErrorUnused:      false,
		})
		if err != nil {
			return packerSDKArtifacts, fmt.Errorf(
				"failed to create decoder for HCP Packer artifact: %w",
				err)
		}

		state := art.State(packerSDKRegistry.ArtifactStateURI)
		if state == nil {
			log.Printf("[WARN] - artifact %q returned a nil value for the HCP state, ignoring", art.BuilderId())
			continue
		}

		err = decoder.Decode(state)
		if err != nil {
			log.Printf("[WARN] - artifact %q failed to be decoded to an HCP artifact, this is probably because it is not compatible: %s", art.BuilderId(), err)
			continue
		}

		err = bucket.UpdateArtifactForBuild(buildName, sdkImages...)
		if err != nil {
			return packerSDKArtifacts, fmt.Errorf("failed to add artifact for %q: %s", buildName, err)
		}
	}

	build, err := bucket.Version.Build(buildName)
	if err != nil {
		return packerSDKArtifacts, fmt.Errorf(
			"failed to get build %q from version being built. This is a Packer bug.",
			buildName)
	}
	if len(build.Artifacts) == 0 {
		return packerSDKArtifacts, &NotAHCPArtifactError{
			fmt.Errorf("No HCP Packer-compatible artifacts were found for the build"),
		}
	}

	parErr := bucket.markBuildComplete(ctx, buildName)
	if parErr != nil {
		return packerSDKArtifacts, fmt.Errorf(
			"failed to update HCP Packer artifacts for %q: %s",
			buildName,
			parErr)
	}

	return append(packerSDKArtifacts, &registryArtifact{
		BuildName:  buildName,
		BucketName: bucket.Name,
		VersionID:  bucket.Version.ID,
	}), nil
}
