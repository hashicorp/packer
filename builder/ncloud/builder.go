package ncloud

import (
	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// Builder assume this implements packer.Builder
type Builder struct {
	config   *Config
	stateBag multistep.StateBag
	runner   multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	b.stateBag = new(multistep.BasicStateBag)

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	ui.Message("Creating Naver Cloud Platform Connection ...")
	conn := ncloud.NewConnection(b.config.AccessKey, b.config.SecretKey)

	b.stateBag.Put("hook", hook)
	b.stateBag.Put("ui", ui)

	var steps []multistep.Step

	steps = []multistep.Step{}

	if b.config.Comm.Type == "ssh" {
		steps = []multistep.Step{
			NewStepValidateTemplate(conn, ui, b.config),
			NewStepCreateLoginKey(conn, ui),
			NewStepCreateServerInstance(conn, ui, b.config),
			NewStepCreateBlockStorageInstance(conn, ui, b.config),
			NewStepGetRootPassword(conn, ui),
			NewStepCreatePublicIPInstance(conn, ui, b.config),
			&communicator.StepConnectSSH{
				Config: &b.config.Comm,
				Host: func(stateBag multistep.StateBag) (string, error) {
					return stateBag.Get("PublicIP").(string), nil
				},
				SSHConfig: SSHConfig(b.config.Comm.SSHUsername),
			},
			&common.StepProvision{},
			NewStepStopServerInstance(conn, ui),
			NewStepCreateServerImage(conn, ui, b.config),
			NewStepDeleteBlockStorageInstance(conn, ui, b.config),
			NewStepTerminateServerInstance(conn, ui),
		}
	} else if b.config.Comm.Type == "winrm" {
		steps = []multistep.Step{
			NewStepValidateTemplate(conn, ui, b.config),
			NewStepCreateLoginKey(conn, ui),
			NewStepCreateServerInstance(conn, ui, b.config),
			NewStepCreateBlockStorageInstance(conn, ui, b.config),
			NewStepGetRootPassword(conn, ui),
			NewStepCreatePublicIPInstance(conn, ui, b.config),
			&communicator.StepConnectWinRM{
				Config: &b.config.Comm,
				Host: func(stateBag multistep.StateBag) (string, error) {
					return stateBag.Get("PublicIP").(string), nil
				},
				WinRMConfig: func(state multistep.StateBag) (*communicator.WinRMConfig, error) {
					return &communicator.WinRMConfig{
						Username: b.config.Comm.WinRMUser,
						Password: state.Get("Password").(string),
					}, nil
				},
			},
			&common.StepProvision{},
			NewStepStopServerInstance(conn, ui),
			NewStepCreateServerImage(conn, ui, b.config),
			NewStepDeleteBlockStorageInstance(conn, ui, b.config),
			NewStepTerminateServerInstance(conn, ui),
		}
	}

	// Run!
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, b.stateBag)
	b.runner.Run(b.stateBag)

	// If there was an error, return that
	if rawErr, ok := b.stateBag.GetOk("Error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{}

	if serverImage, ok := b.stateBag.GetOk("memberServerImage"); ok {
		artifact.ServerImage = serverImage.(*ncloud.ServerImage)
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	b.runner.Cancel()
}
