package proxmoxclone

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/proxmox/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// The unique id for the builder
const BuilderID = "proxmox.clone"

type Builder struct {
	config Config
}

// Builder implements packer.Builder
var _ packer.Builder = &Builder{}

var pluginVersion = "1.0.0"

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return b.config.Prepare(raws...)
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("clone-config", &b.config)

	steps := []multistep.Step{}
	postSteps := []multistep.Step{}

	sb := proxmox.NewSharedBuilder(BuilderID, b.config.Config, steps, postSteps)
	return sb.Run(ctx, ui, hook, state)
}
