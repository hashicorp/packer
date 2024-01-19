// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package hcp_packer_image

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
	hcpPackerDeprecatedModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
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
	// The name of the bucket your image is in.
	Bucket string `mapstructure:"bucket_name" required:"true"`
	// The name of the channel to use when retrieving your image.
	// Either this or `iteration_id` MUST be set.
	// Mutually exclusive with `iteration_id`.
	// If using several images from a single iteration, you may prefer
	// sourcing an iteration first, and referencing it for subsequent uses,
	// as every `hcp-packer-image` with the channel set will generate a
	// potentially billable HCP Packer request, but if several
	// `hcp-packer-image`s use a shared `hcp-packer-iteration` that will
	// only generate one potentially billable request.
	Channel string `mapstructure:"channel" required:"true"`
	// The ID of the iteration to use when retrieving your image
	// Either this or `channel` MUST be set.
	// Mutually exclusive with `channel`
	IterationID string `mapstructure:"iteration_id" required:"true"`
	// The name of the cloud provider that your image is for. For example,
	// "aws" or "gce".
	CloudProvider string `mapstructure:"cloud_provider" required:"true"`
	// The name of the cloud region your image is in. For example "us-east-1".
	Region string `mapstructure:"region" required:"true"`
	// The specific Packer builder used to create the image.
	// For example, "amazon-ebs.example"
	ComponentType string `mapstructure:"component_type" required:"false"`
	// TODO: Version          string `mapstructure:"version"`
	// TODO: Fingerprint          string `mapstructure:"fingerprint"`
	// TODO: Label          string `mapstructure:"label"`
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

	if d.config.Bucket == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The `bucket_name` must be specified"))
	}

	// Ensure either channel or iteration_id are set, and not both at the same time
	if d.config.Channel == "" &&
		d.config.IterationID == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New(
			"The `iteration_id` or `channel` must be specified."))
	}
	if d.config.Channel != "" &&
		d.config.IterationID != "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New(
			"`iteration_id` and `channel` cannot both be specified."))
	}

	if d.config.Region == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`region` is "+
			"currently a required field."))
	}
	if d.config.CloudProvider == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`cloud_provider` is "+
			"currently a required field."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

// DatasourceOutput Information from []*hcpPackerDeprecatedModels.HashicorpCloudPackerImage with some information
// from the parent []*hcpPackerDeprecatedModels.HashicorpCloudPackerBuild included where it seemed
// like it might be relevant. Need to copy so we can generate
type DatasourceOutput struct {
	// The name of the cloud provider that the image exists in. For example,
	// "aws", "azure", or "gce".
	CloudProvider string `mapstructure:"cloud_provider"`
	// The specific Packer builder or post-processor used to create the image.
	ComponentType string `mapstructure:"component_type"`
	// The date and time at which the image was created.
	CreatedAt string `mapstructure:"created_at"`
	// The ID of the build that created the image. This is a ULID, which is a
	// unique identifier similar to a UUID. It is created by the HCP Packer
	// Registry when an build is first created, and is unique to this build.
	BuildID string `mapstructure:"build_id"`
	// The iteration ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an iteration is
	// first created, and is unique to this iteration.
	IterationID string `mapstructure:"iteration_id"`
	// The ID of the channel used to query the image iteration. This value will be empty if the `iteration_id` was used
	// directly instead of a channel.
	ChannelID string `mapstructure:"channel_id"`
	// The UUID associated with the Packer run that created this image.
	PackerRunUUID string `mapstructure:"packer_run_uuid"`
	// ID or URL of the remote cloud image as given by a build.
	ID string `mapstructure:"id"`
	// The cloud region as given by `packer build`. eg. "ap-east-1".
	// For locally managed clouds, this may map instead to a cluster, server
	// or datastore.
	Region string `mapstructure:"region"`
	// The key:value metadata labels associated with this build.
	Labels map[string]string `mapstructure:"labels"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	log.Printf("[WARN] Deprecation: `hcp-packer-image` datasource has been deprecated. " +
		"Please use `hcp-packer-artifact` datasource instead.")
	ctx := context.TODO()

	cli, err := hcpapi.NewDeprecatedClient()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	var iteration *hcpPackerDeprecatedModels.HashicorpCloudPackerIteration
	var channelID string
	if d.config.IterationID != "" {
		log.Printf("[INFO] Reading info from HCP Packer registry (%s) [project_id=%s, organization_id=%s, iteration_id=%s]",
			d.config.Bucket, cli.ProjectID, cli.OrganizationID, d.config.IterationID)

		iter, err := cli.GetIteration(ctx, d.config.Bucket, hcpapi.GetIteration_byID(d.config.IterationID))
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf(
				"error retrieving image iteration from HCP Packer registry: %s",
				err)
		}
		iteration = iter
	} else {
		log.Printf("[INFO] Reading info from HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s]",
			d.config.Bucket, cli.ProjectID, cli.OrganizationID, d.config.Channel)

		channel, err := cli.GetChannel(ctx, d.config.Bucket, d.config.Channel)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("error retrieving "+
				"channel from HCP Packer registry: %s", err.Error())
		}

		if channel.Iteration == nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("there is no iteration associated with the channel %s",
				d.config.Channel)
		}
		channelID = channel.ID
		iteration = channel.Iteration
	}

	revokeAt := time.Time(iteration.RevokeAt)
	if !revokeAt.IsZero() && revokeAt.Before(time.Now().UTC()) {
		// If RevokeAt is not a zero date and is before NOW, it means this iteration is revoked and should not be used
		// to build new images.
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("the iteration %s is revoked and can not be used on Packer builds",
			iteration.ID)
	}

	output := DatasourceOutput{}

	cloudAndRegions := map[string][]string{}
	for _, build := range iteration.Builds {
		if build.CloudProvider != d.config.CloudProvider {
			continue
		}
		for _, image := range build.Images {
			cloudAndRegions[build.CloudProvider] = append(cloudAndRegions[build.CloudProvider], image.Region)
			if image.Region == d.config.Region && filterBuildByComponentType(build, d.config.ComponentType) {
				// This is the desired image.
				output = DatasourceOutput{
					CloudProvider: build.CloudProvider,
					ComponentType: build.ComponentType,
					CreatedAt:     image.CreatedAt.String(),
					BuildID:       build.ID,
					IterationID:   build.IterationID,
					ChannelID:     channelID,
					PackerRunUUID: build.PackerRunUUID,
					ID:            image.ImageID,
					Region:        image.Region,
					Labels:        build.Labels,
				}
				return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
			}
		}
	}

	return cty.NullVal(cty.EmptyObject), fmt.Errorf("could not find a build result matching "+
		"[region=%q, cloud_provider=%q, component_type=%q]. Available: %v ",
		d.config.Region, d.config.CloudProvider, d.config.ComponentType, cloudAndRegions)
}

func filterBuildByComponentType(build *hcpPackerDeprecatedModels.HashicorpCloudPackerBuild, componentType string) bool {
	// optional field is not specified, passthrough
	if componentType == "" {
		return true
	}
	// if specified, only the matched image metadata is returned by this effect
	return build.ComponentType == componentType
}
