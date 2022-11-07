//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config,ParBuild,ParImage
package packer_image_iteration

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	hcpapi "github.com/hashicorp/packer/internal/hcp/api"
)

// Type for Packer datasource has been renamed temporarily to prevent it from being
// automatically registered as a viable datasource plugin in command/plugin.go.
// In the future this type will be renamed to allow for the use of the datasource.
type DeactivatedDatasource struct {
	config Config
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The name of the bucket your image is in.
	Bucket string `mapstructure:"bucket_name" required:"true"`
	// The name of the channel to use when retrieving your image
	Channel string `mapstructure:"channel" required:"true"`
	// TODO: Version          string `mapstructure:"version"`
	// TODO: Label          string `mapstructure:"label"`
}

func (d *DeactivatedDatasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *DeactivatedDatasource) Configure(raws ...interface{}) error {
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
	// The name of the cloud provider that the build exists in. For example,
	// "aws", "azure", or "gce".
	CloudProvider string `mapstructure:"cloud_provider"`
	// The specific Packer builder or post-processor used to create the build.
	ComponentType string `mapstructure:"component_type"`
	// The date and time at which the build was run.
	CreatedAt string `mapstructure:"created_at"`
	// The build ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an build is
	// first created, and is unique to this build.
	ID string `mapstructure:"id"`
	// A list of images as stored in the HCP Packer registry. See the ParImage
	// docs for more information.
	Images []ParImage `mapstructure:"images"`
	// The iteration ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an iteration is
	// first created, and is unique to this iteration.
	IterationID string `mapstructure:"iteration_id"`
	// Unstructured key:value metadata associated with the build.
	Labels map[string]string `mapstructure:"labels"`
	// The UUID associated with the Packer run that created this build.
	PackerRunUUID string `mapstructure:"packer_run_uuid"`
	// Whether the build is considered "complete" (the Packer build ran
	// successfully and created an artifact), or "incomplete" (the Packer
	// build did not finish, and there is no uploaded artifact).
	Status string `mapstructure:"status"`
	// The date and time at which the build was last updated.
	UpdatedAt string `mapstructure:"updated_at"`
}

// Copy of []*models.HashicorpCloudPackerImage Need to copy so we can generate
// the HCL spec.
type ParImage struct {
	// The date and time at which the build was last updated.
	CreatedAt string `mapstructure:"created_at,omitempty"`
	// The iteration ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an iteration is
	// first created, and is unique to this iteration.
	ID string `mapstructure:"id,omitempty"`
	// ID or URL of the remote cloud image as given by a build.
	ImageID string `mapstructure:"image_id,omitempty"`
	// The cloud region as given by `packer build`. eg. "ap-east-1".
	// For locally managed clouds, this may map instead to a cluster, server
	// or datastore.
	Region string `mapstructure:"region,omitempty"`
}

type DatasourceOutput struct {
	// The iteration ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an iteration is
	// first created, and is unique to this iteration.
	Id string `mapstructure:"Id"`
	// The version number assigned to an iteration. This number is an integer,
	// and is created by the HCP Packer Registry once an iteration is
	// marked "complete". If a new iteration is marked "complete", the version
	// that HCP Packer assigns to it will always be the highest previous
	// iteration version plus one.
	IncrementalVersion int32 `mapstructure:"incremental_version"`
	// The date the iteration was created.
	CreatedAt string `mapstructure:"created_at"`
	// A list of builds that are stored in the iteration. These builds can be
	// parsed using HCL to find individual image IDs for specific providers.
	Builds []ParBuild `mapstructure:"builds"`
}

func (d *DeactivatedDatasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *DeactivatedDatasource) Execute() (cty.Value, error) {
	ctx := context.TODO()

	cli, err := hcpapi.NewClient()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}
	// Load channel.
	log.Printf("[INFO] Reading info from HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s]",
		d.config.Bucket, cli.ProjectID, cli.OrganizationID, d.config.Channel)

	channel, err := cli.GetChannel(ctx, d.config.Bucket, d.config.Channel)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error retrieving "+
			"channel from HCP Packer registry: %s", err.Error())
	}

	var iteration *models.HashicorpCloudPackerIteration
	if channel != nil {
		if channel.Iteration != nil {
			iteration = channel.Iteration
		}
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("there is no iteration associated with the channel %s",
			d.config.Channel)
	}

	revokeAt := time.Time(iteration.RevokeAt)
	if !revokeAt.IsZero() && revokeAt.Before(time.Now().UTC()) {
		// If RevokeAt is not a zero date and is before NOW, it means this iteration is revoked and should not be used
		// to build new images.
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("the iteration associated with the channel %s is revoked and can not be used on Packer builds",
			d.config.Channel)
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
