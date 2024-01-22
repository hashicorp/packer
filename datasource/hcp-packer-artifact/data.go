// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package hcp_packer_artifact

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	hcpapi "github.com/hashicorp/packer/internal/hcp/api"
)

type Datasource struct {
	config Config
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The name of the bucket your artifact is in.
	BucketName string `mapstructure:"bucket_name" required:"true"`

	// The name of the channel to use when retrieving your artifact.
	// Either `channel_name` or `version_fingerprint` MUST be set.
	// If using several artifacts from a single version, you may prefer sourcing a version first,
	// and referencing it for subsequent uses, as every `hcp_packer_artifact` with the channel set will generate a
	// potentially billable HCP Packer request, but if several `hcp_packer_artifact`s use a shared `hcp_packer_version`
	// that will only generate one potentially billable request.
	ChannelName string `mapstructure:"channel_name" required:"true"`

	// The fingerprint of the version to use when retrieving your artifact.
	// Either this or `channel_name` MUST be set.
	// Mutually exclusive with `channel_name`
	VersionFingerprint string `mapstructure:"version_fingerprint" required:"true"`

	// The name of the platform that your artifact is for.
	// For example, "aws", "azure", or "gce".
	Platform string `mapstructure:"platform" required:"true"`

	// The name of the region your artifact is in.
	// For example "us-east-1".
	Region string `mapstructure:"region" required:"true"`

	// The specific Packer builder used to create the artifact.
	// For example, "amazon-ebs.example"
	ComponentType string `mapstructure:"component_type" required:"false"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.BucketName == "" {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("the `bucket_name` must be specified"),
		)
	}

	// Ensure either channel_name or version_fingerprint is set, and not both at the same time.
	if d.config.ChannelName == "" && d.config.VersionFingerprint == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New(
			"`version_fingerprint` or `channel_name` must be specified",
		))
	}
	if d.config.ChannelName != "" && d.config.VersionFingerprint != "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New(
			"`version_fingerprint` and `channel_name` cannot be specified together",
		))
	}

	if d.config.Region == "" {
		errs = packersdk.MultiErrorAppend(errs,
			fmt.Errorf("the `region` must be specified"),
		)
	}

	if d.config.Platform == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"the `platform` must be specified",
		))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

// DatasourceOutput Information from []*hcpPackerModels.HashicorpCloudPacker20230101Artifact with some information
// from the parent []*hcpPackerModels.HashicorpCloudPacker20230101Build included where it seemed
// like it might be relevant. Need to copy so we can generate
type DatasourceOutput struct {
	// The name of the platform that the artifact exists in.
	// For example, "aws", "azure", or "gce".
	Platform string `mapstructure:"platform"`

	// The specific Packer builder or post-processor used to create the artifact.
	ComponentType string `mapstructure:"component_type"`

	// The date and time at which the artifact was created.
	CreatedAt string `mapstructure:"created_at"`

	// The ID of the build that created the artifact. This is a ULID, which is a
	// unique identifier similar to a UUID. It is created by the HCP Packer
	// Registry when a build is first created, and is unique to this build.
	BuildID string `mapstructure:"build_id"`

	// The version ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when a version is
	// first created, and is unique to this version.
	VersionID string `mapstructure:"version_id"`

	// The ID of the channel used to query the version. This value will be empty if the `version_fingerprint` was used
	// directly instead of a channel.
	ChannelID string `mapstructure:"channel_id"`

	// The UUID associated with the Packer run that created this artifact.
	PackerRunUUID string `mapstructure:"packer_run_uuid"`

	// Identifier or URL of the remote artifact as given by a build.
	// For example, ami-12345.
	ExternalIdentifier string `mapstructure:"external_identifier"`

	// The region as given by `packer build`. eg. "ap-east-1".
	// For locally managed clouds, this may map instead to a cluster, server or datastore.
	Region string `mapstructure:"region"`

	// The key:value metadata labels associated with this build.
	Labels map[string]string `mapstructure:"labels"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	ctx := context.TODO()

	cli, err := hcpapi.NewClient()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	var version *hcpPackerModels.HashicorpCloudPacker20230101Version
	var channelID string
	if d.config.VersionFingerprint != "" {
		log.Printf(
			"[INFO] Reading info from HCP Packer Registry (%s) "+
				"[project_id=%s, organization_id=%s, version_fingerprint=%s]",
			d.config.BucketName, cli.ProjectID, cli.OrganizationID, d.config.VersionFingerprint,
		)

		version, err = cli.GetVersion(ctx, d.config.BucketName, d.config.VersionFingerprint)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf(
				"error retrieving version from HCP Packer Registry: %s", err,
			)
		}
	} else {
		log.Printf(
			"[INFO] Reading info from HCP Packer Registry (%s) "+
				"[project_id=%s, organization_id=%s, channel=%s]",
			d.config.BucketName, cli.ProjectID, cli.OrganizationID, d.config.ChannelName,
		)

		var channel *hcpPackerModels.HashicorpCloudPacker20230101Channel
		channel, err = cli.GetChannel(ctx, d.config.BucketName, d.config.ChannelName)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf(
				"error retrieving channel from HCP Packer Registry: %s", err.Error(),
			)
		}

		if channel.Version == nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf(
				"there is no version associated with the channel %s", d.config.ChannelName,
			)
		}
		channelID = channel.ID
		version = channel.Version
	}

	if *version.Status == hcpPackerModels.HashicorpCloudPacker20230101VersionStatusVERSIONREVOKED {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"the version %s is revoked and can not be used on Packer builds", version.ID,
		)
	}

	var output DatasourceOutput

	cloudAndRegions := map[string][]string{}
	for _, build := range version.Builds {
		if build.Platform != d.config.Platform {
			continue
		}
		for _, artifact := range build.Artifacts {
			cloudAndRegions[build.Platform] = append(cloudAndRegions[build.Platform], artifact.Region)
			if artifact.Region == d.config.Region && filterBuildByComponentType(build, d.config.ComponentType) {
				// This is the desired artifact.
				output = DatasourceOutput{
					Platform:           build.Platform,
					ComponentType:      build.ComponentType,
					CreatedAt:          artifact.CreatedAt.String(),
					BuildID:            build.ID,
					VersionID:          build.VersionID,
					ChannelID:          channelID,
					PackerRunUUID:      build.PackerRunUUID,
					ExternalIdentifier: artifact.ExternalIdentifier,
					Region:             artifact.Region,
					Labels:             build.Labels,
				}
				return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
			}
		}
	}

	return cty.NullVal(cty.EmptyObject), fmt.Errorf(
		"could not find a build result matching "+
			"[region=%q, platform=%q, component_type=%q]. Available: %v ",
		d.config.Region, d.config.Platform, d.config.ComponentType, cloudAndRegions,
	)
}

func filterBuildByComponentType(build *hcpPackerModels.HashicorpCloudPacker20230101Build, componentType string) bool {
	// optional field is not specified, passthrough
	if componentType == "" {
		return true
	}
	// if specified, only the matched artifact metadata is returned by this effect
	return build.ComponentType == componentType
}
