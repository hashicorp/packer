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
	log.Println("[Debug] BSU Builder Prerare function")
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	log.Println("[Debug] BSU Builder Run function")
	return nil, nil
}

func (b *Builder) Cancel() {
	log.Println("[Debug] BSU Builder Run function")
}
