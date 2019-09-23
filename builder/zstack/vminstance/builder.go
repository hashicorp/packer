// The zstack package contains a packer.Builder implementation that
// builds images for ZStack Engine.
package vminstance

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/communicator"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder.
const BuilderId = "packer.zstack"

// Builder represents a Packer Builder.
type Builder struct {
	config Config
	runner multistep.Runner
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	errs := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)

	b.config.ctx.EnableEnv = true

	err := b.config.Check()

	if len(err) > 0 {
		errs = packer.MultiErrorAppend(errs, err...)
	}

	return nil, errs
}

// Run executes a zstack Packer build and returns a packer.Artifact
// representing a zstack image.
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Message(fmt.Sprintf("[DEBUG] hello zstack builder"))
	secret := "******"
	if b.config.ShowSecret {
		secret = b.config.KeySecret
	}
	ui.Message(fmt.Sprintf("[INFO] access_key: %s, key_secret: %s, base_url: %s, create timeout: %v ms",
		b.config.AccessKey, secret, b.config.BaseUrl, b.config.stateTimeout.Nanoseconds()/1000/1000))
	driver, err := NewDriverZStack(b.config, ui)
	if err != nil {
		return nil, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&StepPreValidate{},
		&StepGetSSHKey{
			Publicfile: b.config.SSHPublicKeyFile,
		},
	}

	steps = append(steps, &StepCreateVmInstance{})

	if b.config.DataVolumeImage != "" || b.config.DataVolumeSize != "" {
		steps = append(steps, &StepCreateDataVolume{}, &StepAttachDataVolume{})
	}

	if b.config.SkipProvisionMod {
		ui.Message("skip provision mode on")
	} else {
		steps = append(steps,
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      getHostIp,
				SSHConfig: b.config.Comm.SSHConfigFunc(),
			},
			&StepMkfsMount{}, &common.StepProvision{}, &StepStopVmInstance{},
		)
	}
	steps = append(steps, &StepCreateImage{})

	if b.config.ExportImage {
		steps = append(steps, &StepExportImage{})
	}

	// Run the steps.
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	i, ok := state.GetOk(Image)
	if !ok {
		log.Println("Failed to find image in state. Bug?")
		return nil, nil
	}

	p, _ := state.GetOk(ExportPath)
	if p == nil {
		p = []string{}
	}
	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	artifact := &Artifact{
		builderIdValue: BuilderId,
		driver:         driver,
		config:         b.config,
		images:         i.([]*zstacktype.Image),
		exportPath:     p.([]string),
	}
	return artifact, nil
}
