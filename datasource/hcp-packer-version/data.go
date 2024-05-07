// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package hcp_packer_version

import (
	"context"
	"fmt"
	"log"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
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
	// The bucket name in the HCP Packer Registry.
	BucketName string `mapstructure:"bucket_name" required:"true"`
	// The channel name in the given bucket to use for retrieving the version.
	ChannelName string `mapstructure:"channel_name" required:"true"`
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
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the `bucket_name` must be specified"))
	}
	if d.config.ChannelName == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the `channel_name` must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

// DatasourceOutput is essentially a copy of []*models.HashicorpCloudPacker20230101Version, but without
// the build and ancestry details
type DatasourceOutput struct {
	// Name of the author who created this version.
	AuthorID string `mapstructure:"author_id"`

	// The name of the bucket that this version is associated with.
	BucketName string `mapstructure:"bucket_name"`

	// Current state of the version.
	Status string `mapstructure:"status"`

	// The date the version was created.
	CreatedAt string `mapstructure:"created_at"`

	// The fingerprint of the version; this is a  unique identifier set by the Packer build
	// that created this version.
	Fingerprint string `mapstructure:"fingerprint"`

	// The version ID. This is a ULID, which is a unique identifier similar
	// to a UUID. It is created by the HCP Packer Registry when a version is
	// first created, and is unique to this version.
	ID string `mapstructure:"id"`

	// The version name is created by the HCP Packer Registry once a version is
	// "complete". Incomplete or failed versions currently default to having a name "v0".
	Name string `mapstructure:"name"`

	// The date when this version was last updated.
	UpdatedAt string `mapstructure:"updated_at"`

	// The ID of the channel used to query this version.
	ChannelID string `mapstructure:"channel_id"`
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
	log.Printf(
		"[INFO] Reading HCP Packer Version info from HCP Packer Registry (%s) "+
			"[project_id=%s, organization_id=%s, channel=%s]",
		d.config.BucketName, cli.ProjectID, cli.OrganizationID, d.config.ChannelName,
	)

	channel, err := cli.GetChannel(ctx, d.config.BucketName, d.config.ChannelName)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"error retrieving HCP Packer Version from HCP Packer Registry: %s",
			err.Error(),
		)
	}
	if channel.Version == nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"there is no HCP Packer Version associated with the channel %s",
			d.config.ChannelName,
		)
	}

	version := channel.Version

	if *version.Status == hcpPackerModels.HashicorpCloudPacker20230101VersionStatusVERSIONREVOKED {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"the HCP Packer Version associated with the channel %s is revoked and can not be used on Packer builds",
			d.config.ChannelName,
		)
	}

	output := DatasourceOutput{
		AuthorID:    version.AuthorID,
		BucketName:  version.BucketName,
		Status:      string(*version.Status),
		CreatedAt:   version.CreatedAt.String(),
		Fingerprint: version.Fingerprint,
		ID:          version.ID,
		Name:        version.Name,
		UpdatedAt:   version.UpdatedAt.String(),
		ChannelID:   channel.ID,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
