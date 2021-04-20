package ncloud

import (
	"context"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// Builder assume this implements packersdk.Builder
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

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	ui.Message("Creating NAVER CLOUD PLATFORM Connection ...")
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

	steps := []multistep.Step{
		NewStepValidateTemplate(conn, ui, &b.config),
		NewStepCreateLoginKey(conn, ui, &b.config),
		multistep.If(b.config.SupportVPC, NewStepCreateInitScript(conn, ui, &b.config)),
		multistep.If(b.config.SupportVPC, NewStepCreateAccessControlGroup(conn, ui, &b.config)),
		NewStepCreateServerInstance(conn, ui, &b.config),
		NewStepCreateBlockStorage(conn, ui, &b.config),
		NewStepGetRootPassword(conn, ui, &b.config),
		NewStepCreatePublicIP(conn, ui, &b.config),
		multistep.If(b.config.Comm.Type == "ssh", &communicator.StepConnectSSH{
			Config: &b.config.Comm,
			Host: func(stateBag multistep.StateBag) (string, error) {
				return stateBag.Get("public_ip").(string), nil
			},
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		}),
		multistep.If(b.config.Comm.Type == "winrm", &communicator.StepConnectWinRM{
			Config: &b.config.Comm,
			Host: func(stateBag multistep.StateBag) (string, error) {
				return stateBag.Get("public_ip").(string), nil
			},
			WinRMConfig: func(state multistep.StateBag) (*communicator.WinRMConfig, error) {
				return &communicator.WinRMConfig{
					Username: b.config.Comm.WinRMUser,
					Password: b.config.Comm.WinRMPassword,
				}, nil
			},
		}),
		&commonsteps.StepProvision{},
		multistep.If(b.config.Comm.Type == "ssh", &commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		}),
		NewStepStopServerInstance(conn, ui, &b.config),
		NewStepCreateServerImage(conn, ui, &b.config),
		NewStepDeleteBlockStorage(conn, ui, &b.config),
		NewStepTerminateServerInstance(conn, ui, &b.config),
	}

	// Run!
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, b.stateBag)
	b.runner.Run(ctx, b.stateBag)

	// If there was an error, return that
	if rawErr, ok := b.stateBag.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{
		StateData: map[string]interface{}{"generated_data": b.stateBag.Get("generated_data")},
	}

	if serverImage, ok := b.stateBag.GetOk("member_server_image"); ok {
		artifact.MemberServerImage = serverImage.(*server.MemberServerImage)
	}

	return artifact, nil
}
