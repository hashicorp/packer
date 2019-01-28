// The bsusurrogate package contains a packer.Builder implementation that
// builds a new EBS-backed AMI using an ephemeral instance.
package bsusurrogate

import (
	"log"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const BuilderId = "digitalonus.osc.bsusurrogate"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	osccommon.AccessConfig `mapstructure:",squash"`
	osccommon.RunConfig    `mapstructure:",squash"`
	ctx                    interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	log.Println("Preparing Outscale Builder...")
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	log.Println("Running Outscale Builder...")
	return nil, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
