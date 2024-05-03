// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package hcp_packer_iteration

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
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
	// The name of the channel to use when retrieving your image
	Channel string `mapstructure:"channel" required:"true"`
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
	if d.config.Channel == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`channel` is currently a required field."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

// DatasourceOutput is essentially a copy of []*models.HashicorpCloudPackerIteration, but without the
// []Builds or ancestor id.
type DatasourceOutput struct {
	// who created the iteration
	AuthorID string `mapstructure:"author_id"`
	// Name of the bucket that the iteration was retrieved from
	BucketName string `mapstructure:"bucket_name"`
	// If true, this iteration is considered "ready to use" and will be
	// returned even if the include_incomplete flag is "false" in the
	// list iterations request. Note that if you are retrieving an iteration
	// using a channel, this will always be "true"; channels cannot be assigned
	// to incomplete iterations.
	Complete bool `mapstructure:"complete"`
	// The date the iteration was created.
	CreatedAt string `mapstructure:"created_at"`
	// The fingerprint of the build; this could be a git sha or other unique
	// identifier as set by the Packer build that created this iteration.
	Fingerprint string `mapstructure:"fingerprint"`
	// The iteration ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when an iteration is
	// first created, and is unique to this iteration.
	ID string `mapstructure:"id"`
	// The version number assigned to an iteration. This number is an integer,
	// and is created by the HCP Packer Registry once an iteration is
	// marked "complete". If a new iteration is marked "complete", the version
	// that HCP Packer assigns to it will always be the highest previous
	// iteration version plus one.
	IncrementalVersion int32 `mapstructure:"incremental_version"`
	// The date when this iteration was last updated.
	UpdatedAt string `mapstructure:"updated_at"`
	// The ID of the channel used to query the image iteration.
	ChannelID string `mapstructure:"channel_id"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	log.Printf("[WARN] Deprecation: `hcp-packer-iteration` datasource has been deprecated. " +
		"Please use `hcp-packer-version` datasource instead.")
	ctx := context.TODO()

	cli, err := hcpapi.NewDeprecatedClient()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}
	// Load channel.
	log.Printf("[INFO] Reading iteration info from HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s]",
		d.config.Bucket, cli.ProjectID, cli.OrganizationID, d.config.Channel)

	channel, err := cli.GetChannel(ctx, d.config.Bucket, d.config.Channel)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error retrieving "+
			"iteration from HCP Packer registry: %s", err.Error())
	}
	if channel.Iteration == nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("there is no iteration associated with the channel %s",
			d.config.Channel)
	}

	iteration := channel.Iteration

	revokeAt := time.Time(iteration.RevokeAt)
	if !revokeAt.IsZero() && revokeAt.Before(time.Now().UTC()) {
		// If RevokeAt is not a zero date and is before NOW, it means this iteration is revoked and should not be used
		// to build new images.
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("the iteration associated with the channel %s is revoked and can not be used on Packer builds",
			d.config.Channel)
	}

	output := DatasourceOutput{
		AuthorID:           iteration.AuthorID,
		BucketName:         iteration.BucketSlug,
		Complete:           iteration.Complete,
		CreatedAt:          iteration.CreatedAt.String(),
		Fingerprint:        iteration.Fingerprint,
		ID:                 iteration.ID,
		IncrementalVersion: iteration.IncrementalVersion,
		UpdatedAt:          iteration.UpdatedAt.String(),
		ChannelID:          channel.ID,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
