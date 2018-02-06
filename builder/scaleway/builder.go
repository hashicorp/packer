// The scaleway package contains a packer.Builder implementation
// that builds Scaleway images (snapshots).

package scaleway

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

// The unique id for the builder
const BuilderId = "hashicorp.scaleway"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = *c

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	client, err := api.NewScalewayAPI(b.config.Organization, b.config.Token, b.config.UserAgent, b.config.Region)

	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&stepCreateSSHKey{
			Debug:          b.config.PackerDebug,
			DebugKeyPath:   fmt.Sprintf("scw_%s.pem", b.config.PackerBuildName),
			PrivateKeyFile: b.config.Comm.SSHPrivateKey,
		},
		new(stepCreateServer),
		new(stepServerInfo),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: sshConfig,
		},
		new(common.StepProvision),
		new(stepShutdown),
		new(stepSnapshot),
		new(stepImage),
	}

	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

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

	if _, ok := state.GetOk("snapshot_name"); !ok {
		return nil, errors.New("Cannot find snapshot_name in state.")
	}

	artifact := &Artifact{
		imageName:    state.Get("image_name").(string),
		imageID:      state.Get("image_id").(string),
		snapshotName: state.Get("snapshot_name").(string),
		snapshotID:   state.Get("snapshot_id").(string),
		regionName:   state.Get("region").(string),
		client:       client,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
