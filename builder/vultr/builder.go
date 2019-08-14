package vultr

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vultr/govultr"
)

// Special OS IDs
const (
	AppOSID      = 186
	CustomOSID   = 159
	SnapshotOSID = 164
)

// BuilderID is the unique ID for the builder
const BuilderID = "packer.vultr"

// Builder ...
type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) (warnings []string, err error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (ret packer.Artifact, err error) {
	ui.Say("Running Vultr builder...")

	client := newVultrClient(b.config.APIKey)

	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&stepCreateSSHKey{
			client:       client,
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("vultr_%s.pem", b.config.PackerBuildName),
		},
		&stepCreateServer{client},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: sshConfig,
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepShutdown{client},
		&stepCreateSnapshot{client},
	}

	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("build was cancelled")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("build was halted")
	}

	if _, ok := state.GetOk("snapshot"); !ok {
		return nil, errors.New("cannot find snapshot in state")
	}

	snapshot := state.Get("snapshot").(*govultr.Snapshot)
	artifact := &Artifact{
		SnapshotID:  snapshot.SnapshotID,
		Description: snapshot.Description,
		client:      client,
	}

	return artifact, nil
}
