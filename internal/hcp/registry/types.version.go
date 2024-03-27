// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	packerSDKRegistry "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer/internal/hcp/env"
	"github.com/hashicorp/packer/packer"
	"github.com/oklog/ulid"
)

type Version struct {
	ID             string
	Fingerprint    string
	RunUUID        string
	builds         sync.Map
	expectedBuilds []string
}

type VersionOptions struct {
	TemplateBaseDir string
}

// NewVersion returns a pointer to a Version that can be used for storing Packer build details needed.
func NewVersion() *Version {
	i := Version{
		expectedBuilds: make([]string, 0),
	}

	return &i
}

// Initialize prepares the version to be used with HCP Packer.
func (version *Version) Initialize() error {
	if version == nil {
		return errors.New("Unexpected call to initialize for a nil Version")
	}

	// Bydefault we try to load a Fingerprint from the environment variable.
	// If no variable is defined we generate a new fingerprint.
	version.Fingerprint = os.Getenv(env.HCPPackerBuildFingerprint)

	if version.Fingerprint != "" {
		return nil
	}

	fp, err := ulid.New(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0))
	if err != nil {
		return fmt.Errorf("Failed to generate a fingerprint: %s", err)
	}
	version.Fingerprint = fp.String()

	return nil
}

// StoreBuild stores a build for buildName to an active version.
func (version *Version) StoreBuild(buildName string, build *Build) {
	version.builds.Store(buildName, build)
}

// Build gets the store build associated with buildName in the active version.
func (version *Version) Build(buildName string) (*Build, error) {
	build, ok := version.builds.Load(buildName)
	if !ok {
		return nil, errors.New("no associated build found for the name " + buildName)
	}

	b, ok := build.(*Build)
	if !ok {
		return nil, fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", buildName)
	}

	return b, nil
}

// HasBuild checks if version has a stored build associated with buildName.
func (version *Version) HasBuild(buildName string) bool {
	_, ok := version.builds.Load(buildName)

	return ok
}

// AddArtifactToBuild appends one or more artifacts to the build referred to by buildName.
func (version *Version) AddArtifactToBuild(buildName string, artifacts ...packerSDKRegistry.Image) error {
	build, err := version.Build(buildName)
	if err != nil {
		return fmt.Errorf("AddArtifactToBuild: %w", err)
	}

	err = build.AddArtifacts(artifacts...)
	if err != nil {
		return fmt.Errorf("AddArtifactToBuild: %w", err)
	}

	version.StoreBuild(buildName, build)
	return nil
}

// AddLabelsToBuild merges the contents of data to the labels associated with the build referred to by buildName.
func (version *Version) AddLabelsToBuild(buildName string, data map[string]string) error {
	build, err := version.Build(buildName)
	if err != nil {
		return fmt.Errorf("AddLabelsToBuild: %w", err)
	}

	build.MergeLabels(data)

	version.StoreBuild(buildName, build)
	return nil
}

// AddSHAToBuildLabels adds the Git SHA for the current version (if set) as a label for all the builds of the version
func (version *Version) AddSHAToBuildLabels(sha string) {
	version.builds.Range(func(_, v any) bool {
		b, ok := v.(*Build)
		if !ok {
			return true
		}

		b.MergeLabels(map[string]string{
			"git_sha": sha,
		})

		return true
	})
}

// RemainingBuilds returns the list of builds that are not in a DONE status
func (version *Version) RemainingBuilds() []*Build {
	var todo []*Build

	version.builds.Range(func(k, v any) bool {
		build, ok := v.(*Build)
		if !ok {
			// Unlikely since the builds map contains only Build instances
			return true
		}

		if build.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE {
			todo = append(todo, build)
		}
		return true
	})

	return todo
}

func (version *Version) statusSummary(ui sdkpacker.Ui) {
	rem := version.RemainingBuilds()
	if rem == nil {
		return
	}

	buf := &strings.Builder{}

	buf.WriteString(fmt.Sprintf(
		"\nVersion %q is incomplete, the following builds are missing artifact metadata:\n\n",
		version.Fingerprint))
	for _, b := range rem {
		buf.WriteString(fmt.Sprintf("* %q: %s\n", b.ComponentType, b.Status))
	}
	buf.WriteString("\nYou may resume work on this version in further Packer builds by defining the following variable in your environment:\n")
	buf.WriteString(fmt.Sprintf("HCP_PACKER_BUILD_FINGERPRINT=%q", version.Fingerprint))

	ui.Say(buf.String())
}

// AddMetadataToBuild adds metadata to a build in the HCP Packer registry.
func (version *Version) AddMetadataToBuild(
	ctx context.Context, buildName string, metadata packer.BuildMetadata,
) error {
	buildToUpdate, err := version.Build(buildName)
	if err != nil {
		return err
	}

	packerMetadata := make(map[string]interface{})
	packerMetadata["version"] = metadata.PackerVersion

	var pluginsMetadata []map[string]interface{}
	for _, plugin := range metadata.Plugins {
		pluginMetadata := map[string]interface{}{
			"version": plugin.Description.Version,
			"name":    plugin.Name,
		}
		pluginsMetadata = append(pluginsMetadata, pluginMetadata)
	}
	packerMetadata["plugins"] = pluginsMetadata

	buildToUpdate.Metadata.Packer = packerMetadata
	return nil
}
