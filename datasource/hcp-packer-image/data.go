//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package hcp_packer_image

import (
	"context"
	"fmt"
	"log"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	packerregistry "github.com/hashicorp/packer/internal/registry"
)

type Datasource struct {
	config Config
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The name of the bucket your image is in.
	Bucket string `mapstructure:"bucket_name" required:"true"`
	// The name of the iteration Id to use when retrieving your image
	IterationID string `mapstructure:"iteration_id" required:"true"`
	// The name of the cloud provider that your image is for. For example,
	// "aws" or "gce".
	CloudProvider string `mapstructure:"cloud_provider" required:"true"`
	// The name of the cloud region your image is in. For example "us-east-1".
	Region string `mapstructure:"region" required:"true"`
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
	if d.config.IterationID == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The `iteration_id`"+
			" must be specified. If you do not know your iteration_id, you "+
			"can retrieve it using the bucket name and desired channel using"+
			" the hcp-packer-iteration data source."))
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

// Information from []*models.HashicorpCloudPackerImage with some information
// from the parent []*models.HashicorpCloudPackerBuild included where it seemed
// like it might be relevant. Need to copy so we can generate
type DatasourceOutput struct {
	// The name of the cloud provider that the image exists in. For example,
	// "aws", "azure", or "gce".
	CloudProvider string `mapstructure:"cloud_provider"`
	// The specific Packer builder or post-processor used to create the image.
	ComponentType string `mapstructure:"component_type"`
	// The date and time at which the image was created.
	CreatedAt string `mapstructure:"created_at"`
	// The id of the build that created the image. This is a ULID, which is a
	// unique identifier similar to a UUID. It is created by the HCP Packer
	// Registry when an build is first created, and is unique to this build.
	BuildID string `mapstructure:"build_id"`
	// The iteration id. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an iteration is
	// first created, and is unique to this iteration.
	IterationID string `mapstructure:"iteration_id"`
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
	ctx := context.TODO()

	cli, err := packerregistry.NewClient()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	// Load channel.
	log.Printf("[INFO] Reading info from HCP Packer registry (%s) [project_id=%s, organization_id=%s, iteration_id=%s]",
		d.config.Bucket, cli.ProjectID, cli.OrganizationID, d.config.IterationID)

	iteration, err := cli.GetIteration(ctx, d.config.Bucket, packerregistry.GetIteration_byID(d.config.IterationID))
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error retrieving "+
			"image iteration from HCP Packer registry: %s", err.Error())
	}

	output := DatasourceOutput{}

	cloudAndRegions := map[string][]string{}
	for _, build := range iteration.Builds {
		if build.CloudProvider != d.config.CloudProvider {
			continue
		}
		for _, image := range build.Images {
			cloudAndRegions[build.CloudProvider] = append(cloudAndRegions[build.CloudProvider], image.Region)
			if image.Region == d.config.Region {
				// This is the desired image.
				output = DatasourceOutput{
					CloudProvider: build.CloudProvider,
					ComponentType: build.ComponentType,
					CreatedAt:     image.CreatedAt.String(),
					BuildID:       build.ID,
					IterationID:   build.IterationID,
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
		"region (%q) and cloud provider (%q). Available: %v ",
		d.config.Region, d.config.CloudProvider, cloudAndRegions)
}
