// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"fmt"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	packerSDKRegistry "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

// Build represents a build of a given component type for some bucket on the HCP Packer Registry.
type Build struct {
	ID            string
	Platform      string
	ComponentType string
	RunUUID       string
	Labels        map[string]string
	Artifacts     map[string]packerSDKRegistry.Image
	Status        hcpPackerModels.HashicorpCloudPacker20230101BuildStatus
	Metadata      hcpPackerModels.HashicorpCloudPacker20230101BuildMetadata
}

// NewBuildFromCloudPackerBuild converts a HashicorpCloudPackerBuild to a local build that can be tracked and
// published to the HCP Packer.
// Any existing labels or artifacts associated to src will be copied to the returned Build.
func NewBuildFromCloudPackerBuild(src *hcpPackerModels.HashicorpCloudPacker20230101Build) (*Build, error) {

	build := Build{
		ID:            src.ID,
		ComponentType: src.ComponentType,
		Platform:      src.Platform,
		RunUUID:       src.PackerRunUUID,
		Status:        *src.Status,
		Labels:        src.Labels,
	}

	var err error
	for _, artifact := range src.Artifacts {
		err = build.AddArtifacts(packerSDKRegistry.Image{
			ImageID:        artifact.ExternalIdentifier,
			ProviderName:   build.Platform,
			ProviderRegion: artifact.Region,
		})

		if err != nil {
			return nil, fmt.Errorf("NewBuildFromCloudPackerBuild: %w", err)
		}
	}

	return &build, nil
}

// MergeLabels merges the contents of data to the labels associated with the build.
// Duplicate keys will be updated to reflect the new value.
func (build *Build) MergeLabels(data map[string]string) {
	if data == nil {
		return
	}

	if build.Labels == nil {
		build.Labels = make(map[string]string)
	}

	for k, v := range data {
		// TODO: (nywilken) Determine why we skip labels already set
		// if _, ok := build.Labels[k]; ok {
		// continue
		// }
		build.Labels[k] = v
	}

}

// AddArtifacts appends one or more artifacts to the build.
func (build *Build) AddArtifacts(artifacts ...packerSDKRegistry.Image) error {

	if build.Artifacts == nil {
		build.Artifacts = make(map[string]packerSDKRegistry.Image)
	}

	for _, artifact := range artifacts {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("AddArtifacts: failed to add artifact to build %q: %w", build.ComponentType, err)
		}

		if build.Platform == "" {
			build.Platform = artifact.ProviderName
		}

		build.MergeLabels(artifact.Labels)
		build.Artifacts[artifact.String()] = artifact
	}

	return nil
}

// IsNotDone returns true if build does not satisfy all requirements of a completed build.
// A completed build must have a valid ID, one or more Artifacts, and its Status is HashicorpCloudPacker20230101BuildStatusBUILDDONE.
func (build *Build) IsNotDone() bool {
	hasBuildID := build.ID != ""
	hasNoArtifacts := len(build.Artifacts) == 0
	isNotDone := build.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE

	return hasBuildID && hasNoArtifacts && isNotDone
}
