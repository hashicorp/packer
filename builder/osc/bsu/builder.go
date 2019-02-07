// The amazonebs package contains a packer.Builder implementation that
// builds OMIs for Outscale OAPI.
//
// In general, there are two types of OMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package bsu

import (
	"log"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "oapi.outscale.bsu"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	osccommon.AccessConfig `mapstructure:",squash"`
	osccommon.OMIConfig    `mapstructure:",squash"`
	osccommon.BlockDevices `mapstructure:",squash"`
	osccommon.RunConfig    `mapstructure:",squash"`
	VolumeRunTags          osccommon.TagMap `mapstructure:"run_volume_tags"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = osccommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"omi_description",
				"run_tags",
				"run_volume_tags",
				"spot_tags",
				"snapshot_tags",
				"tags",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.OMIForceDeregister = true
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.OMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.BlockDevices.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	log.Println("[Debug] BSU Builder Run function")
	return nil, nil
}

func (b *Builder) Cancel() {
	log.Println("[Debug] BSU Builder Run function")
}
