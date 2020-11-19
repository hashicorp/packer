package proxmoxiso

import (
	"context"

	proxmoxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/hcl/v2/hcldec"
	proxmox "github.com/hashicorp/packer/builder/proxmox/common"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// The unique id for the builder
const BuilderID = "proxmox.iso"

type Builder struct {
	config Config
}

// Builder implements packer.Builder
var _ packer.Builder = &Builder{}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return b.config.Prepare(raws...)
}

const downloadPathKey = "downloaded_iso_path"

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("iso-config", &b.config)

	preSteps := []multistep.Step{
		&commonsteps.StepDownload{
			Checksum:    b.config.ISOChecksum,
			Description: "ISO",
			Extension:   b.config.TargetExtension,
			ResultKey:   downloadPathKey,
			TargetPath:  b.config.TargetPath,
			Url:         b.config.ISOUrls,
		},
	}
	for idx := range b.config.AdditionalISOFiles {
		preSteps = append(preSteps, &commonsteps.StepDownload{
			Checksum:    b.config.AdditionalISOFiles[idx].ISOChecksum,
			Description: "additional ISO",
			Extension:   b.config.AdditionalISOFiles[idx].TargetExtension,
			ResultKey:   b.config.AdditionalISOFiles[idx].DownloadPathKey,
			TargetPath:  b.config.AdditionalISOFiles[idx].DownloadPathKey,
			Url:         b.config.AdditionalISOFiles[idx].ISOUrls,
		})
	}
	preSteps = append(preSteps,
		&stepUploadISO{},
		&stepUploadAdditionalISOs{},
	)
	postSteps := []multistep.Step{
		&stepFinalizeISOTemplate{},
	}

	sb := proxmox.NewSharedBuilder(BuilderID, b.config.Config, preSteps, postSteps, &isoVMCreator{})
	return sb.Run(ctx, ui, hook, state)
}

type isoVMCreator struct{}

func (*isoVMCreator) Create(vmRef *proxmoxapi.VmRef, config proxmoxapi.ConfigQemu, state multistep.StateBag) error {
	isoFile := state.Get("iso_file").(string)
	config.QemuIso = isoFile

	client := state.Get("proxmoxClient").(*proxmoxapi.Client)
	return config.CreateVm(vmRef, client)
}
