// The linode package contains a packer.Builder implementation
// that builds Linode images.
package linode

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/linode/linodego"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// The unique ID for this builder.
const BuilderID = "packer.linode"

// Builder represents a Packer Builder.
type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}
	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (ret packersdk.Artifact, err error) {
	ui.Say("Running builder ...")

	client := newLinodeClient(b.config.PersonalAccessToken)

	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("linode_%s.pem", b.config.PackerBuildName),
		},
		&stepCreateLinode{client},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost(b.config.Comm.Host()),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepShutdownLinode{client},
		&stepCreateImage{client},
	}

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	if _, ok := state.GetOk("image"); !ok {
		return nil, errors.New("Cannot find image in state.")
	}

	image := state.Get("image").(*linodego.Image)
	artifact := Artifact{
		ImageLabel: image.Label,
		ImageID:    image.ID,
		Driver:     &client,
		StateData:  map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}

		instance := state.Get("instance").(*linodego.Instance)
		if len(instance.IPv4) == 0 {
			return "", fmt.Errorf("Linode instance %d has no IPv4 addresses!", instance.ID)
		}
		return instance.IPv4[0].String(), nil
	}
}
