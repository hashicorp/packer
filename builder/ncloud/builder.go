package ncloud

import (
	"context"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// Builder assume this implements packer.Builder
type Builder struct {
	config   Config
	stateBag multistep.StateBag
	runner   multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	b.stateBag = new(multistep.BasicStateBag)

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Message("Creating Naver Cloud Platform Connection ...")
	config := Config{
		AccessKey: b.config.AccessKey,
		SecretKey: b.config.SecretKey,
	}

	conn, err := config.Client()
	if err != nil {
		return nil, err
	}

	b.stateBag.Put("hook", hook)
	b.stateBag.Put("ui", ui)

	var steps []multistep.Step

	steps = []multistep.Step{}

	if b.config.Comm.Type == "ssh" {
		steps = []multistep.Step{
			NewStepValidateTemplate(conn, ui, &b.config),
			NewStepCreateLoginKey(conn, ui),
			NewStepCreateServerInstance(conn, ui, &b.config),
			NewStepCreateBlockStorageInstance(conn, ui, &b.config),
			NewStepGetRootPassword(conn, ui, &b.config),
			NewStepCreatePublicIPInstance(conn, ui, &b.config),
			&communicator.StepConnectSSH{
				Config: &b.config.Comm,
				Host: func(stateBag multistep.StateBag) (string, error) {
					return stateBag.Get("PublicIP").(string), nil
				},
				SSHConfig: b.config.Comm.SSHConfigFunc(),
			},
			&commonsteps.StepProvision{},
			&commonsteps.StepCleanupTempKeys{
				Comm: &b.config.Comm,
			},
			NewStepStopServerInstance(conn, ui),
			NewStepCreateServerImage(conn, ui, &b.config),
			NewStepDeleteBlockStorageInstance(conn, ui, &b.config),
			NewStepTerminateServerInstance(conn, ui),
		}
	} else if b.config.Comm.Type == "winrm" {
		steps = []multistep.Step{
			NewStepValidateTemplate(conn, ui, &b.config),
			NewStepCreateLoginKey(conn, ui),
			NewStepCreateServerInstance(conn, ui, &b.config),
			NewStepCreateBlockStorageInstance(conn, ui, &b.config),
			NewStepGetRootPassword(conn, ui, &b.config),
			NewStepCreatePublicIPInstance(conn, ui, &b.config),
			&communicator.StepConnectWinRM{
				Config: &b.config.Comm,
				Host: func(stateBag multistep.StateBag) (string, error) {
					return stateBag.Get("PublicIP").(string), nil
				},
				WinRMConfig: func(state multistep.StateBag) (*communicator.WinRMConfig, error) {
					return &communicator.WinRMConfig{
						Username: b.config.Comm.WinRMUser,
						Password: b.config.Comm.WinRMPassword,
					}, nil
				},
			},
			&commonsteps.StepProvision{},
			NewStepStopServerInstance(conn, ui),
			NewStepCreateServerImage(conn, ui, &b.config),
			NewStepDeleteBlockStorageInstance(conn, ui, &b.config),
			NewStepTerminateServerInstance(conn, ui),
		}
	}

	// Run!
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, b.stateBag)
	b.runner.Run(ctx, b.stateBag)

	// If there was an error, return that
	if rawErr, ok := b.stateBag.GetOk("Error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{
		StateData: map[string]interface{}{"generated_data": b.stateBag.Get("generated_data")},
	}

	if serverImage, ok := b.stateBag.GetOk("memberServerImage"); ok {
		artifact.MemberServerImage = serverImage.(*server.MemberServerImage)
	}

	return artifact, nil
}
