package proxmoxclone

import (
	proxmoxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/hcl/v2/hcldec"
	proxmox "github.com/hashicorp/packer/builder/proxmox/common"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"context"
	"fmt"
)

// The unique id for the builder
const BuilderID = "proxmox.clone"

type Builder struct {
	config Config
}

// Builder implements packer.Builder
var _ packer.Builder = &Builder{}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return b.config.Prepare(raws...)
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("clone-config", &b.config)

	preSteps := []multistep.Step{
		&StepSshKeyPair{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("%s.pem", b.config.PackerBuildName),
		},
	}
	postSteps := []multistep.Step{}

	sb := proxmox.NewSharedBuilder(BuilderID, b.config.Config, preSteps, postSteps, &cloneVMCreator{})
	return sb.Run(ctx, ui, hook, state)
}

type cloneVMCreator struct{}

func (*cloneVMCreator) Create(vmRef *proxmoxapi.VmRef, config proxmoxapi.ConfigQemu, state multistep.StateBag) error {
	client := state.Get("proxmoxClient").(*proxmoxapi.Client)
	c := state.Get("clone-config").(*Config)
	comm := state.Get("config").(*proxmox.Config).Comm

	fullClone := 1
	if c.FullClone.False() {
		fullClone = 0
	}

	config.FullClone = &fullClone
	config.CIuser = comm.SSHUsername
	config.Sshkeys = string(comm.SSHPublicKey)
	sourceVmr, err := client.GetVmRefByName(c.CloneVM)
	if err != nil {
		return err
	}
	err = config.CloneVm(sourceVmr, vmRef, client)
	if err != nil {
		return err
	}
	err = config.UpdateConfig(vmRef, client)
	if err != nil {
		return err
	}
	return nil
}
