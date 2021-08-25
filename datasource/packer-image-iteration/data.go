//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config,ParBuild,ParImage
package packer_image_iteration

import (
	"context"
	"fmt"
	"log"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
)

type Datasource struct {
	config Config
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The name of the bucket your image is in.
	Bucket string `mapstructure:"bucket_name"`
	// The name of the channel to use when retrieving your image
	Channel string `mapstructure:"channel"`
	// TODO: Version          string `mapstructure:"version"`
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
	if d.config.Channel == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`channel` is currently a required field."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

// Copy of []*models.HashicorpCloudPackerBuild. Need to copy so we can generate
// the HCL spec.
type ParBuild struct {
	// aws
	CloudProvider string `mapstructure:"cloud_provider"`
	// builder or post-processor used to build this
	ComponentType string `mapstructure:"component_type"`
	// created at
	// Format: date-time
	CreatedAt string `mapstructure:"created_at"`
	// ULID
	ID string `mapstructure:"id"`
	// images
	Images []ParImage `mapstructure:"images"`
	// ULID of the iteration
	IterationID string `mapstructure:"iteration_id"`
	// unstructured metadata
	Labels map[string]string `mapstructure:"labels"`
	// packer run uuid
	PackerRunUUID string `mapstructure:"packer_run_uuid"`
	// complete
	Status string `mapstructure:"status"`
	// updated at
	// Format: date-time
	UpdatedAt string `mapstructure:"updated_at"`
}

// Copy of []*models.HashicorpCloudPackerImage Need to copy so we can generate
// the HCL spec.
type ParImage struct {
	// Timestamp at which this image was created
	// Format: date-time
	CreatedAt string `mapstructure:"created_at,omitempty"`
	// ULID for the image
	ID string `mapstructure:"id,omitempty"`
	// ID or URL of the remote cloud image as given by a build.
	ImageID string `mapstructure:"image_id,omitempty"`
	// region as given by `packer build`. eg. "ap-east-1"
	Region string `mapstructure:"region,omitempty"`
}

type DatasourceOutput struct {
	Id                 string     `mapstructure:"Id"`
	IncrementalVersion int32      `mapstructure:"incremental_version"`
	CreatedAt          string     `mapstructure:"created_at"`
	Builds             []ParBuild `mapstructure:"builds"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	cli, err := packerregistry.NewClient()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}
	// Load channel.
	log.Printf("[INFO] Reading info from HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s]",
		d.config.Bucket, cli.ProjectID, cli.OrganizationID, d.config.Channel)

	iteration, err := packerregistry.GetIterationFromChannel(context.TODO(), cli, d.config.Bucket, d.config.Channel)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error retrieving "+
			"image iteration from HCP Packer registry: %s", err.Error())
	}
	output := DatasourceOutput{
		IncrementalVersion: iteration.IncrementalVersion,
		CreatedAt:          iteration.CreatedAt.String(),
		Builds:             convertPackerBuildList(iteration.Builds),
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func convertPackerBuildList(builds []*models.HashicorpCloudPackerBuild) (flattened []ParBuild) {
	for _, build := range builds {
		out := ParBuild{
			CloudProvider: build.CloudProvider,
			ComponentType: build.ComponentType,
			CreatedAt:     build.CreatedAt.String(),
			ID:            build.ID,
			Images:        convertPackerBuildImagesList(build.Images),
			Labels:        build.Labels,
			PackerRunUUID: build.PackerRunUUID,
			Status:        string(build.Status),
			UpdatedAt:     build.UpdatedAt.String(),
		}
		flattened = append(flattened, out)
	}
	return
}

func convertPackerBuildImagesList(images []*models.HashicorpCloudPackerImage) (flattened []ParImage) {
	for _, image := range images {
		out := ParImage{
			CreatedAt: image.CreatedAt.String(),
			ID:        image.ID,
			ImageID:   image.ImageID,
			Region:    image.Region,
		}
		flattened = append(flattened, out)
	}
	return
}
