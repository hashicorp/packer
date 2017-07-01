package main

import (
	"errors"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/helper/communicator"
	gossh "golang.org/x/crypto/ssh"
	"github.com/hashicorp/packer/communicator/ssh"
	"github.com/vmware/govmomi"
	"context"
	"net/url"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"fmt"
)

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)
	ctx := context.TODO()
	state.Put("ctx", ctx)

	vcenter_url, err := url.Parse(fmt.Sprintf("https://%v/sdk", b.config.VCenterServer))
	if err != nil {
		return nil, err
	}
	vcenter_url.User = url.UserPassword(b.config.Username, b.config.Password)
	client, err := govmomi.NewClient(ctx, vcenter_url,true)
	if err != nil {
		return nil, err
	}
	state.Put("client", client)

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, b.config.Datacenter)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)
	state.Put("finder", finder)
	state.Put("datacenter", datacenter)

	vmSrc, err := finder.VirtualMachine(ctx, b.config.Template)
	if err != nil {
		return nil, err
	}
	state.Put("vmSrc", vmSrc)

	steps := []multistep.Step{
		&StepCloneVM{
			config: b.config,
		},
		&StepConfigureHW{
			config: b.config,
		},
		&StepRun{},
		&communicator.StepConnect{
			Config:    &b.config.Config,
			Host:      func(state multistep.StateBag) (string, error) {
				return state.Get("ip").(string), nil
			},
			SSHConfig: func(multistep.StateBag) (*gossh.ClientConfig, error) {
				return &gossh.ClientConfig{
					User: b.config.Config.SSHUsername,
					Auth: []gossh.AuthMethod{
						gossh.Password(b.config.Config.SSHPassword),
						gossh.KeyboardInteractive(
							ssh.PasswordKeyboardInteractive(b.config.Config.SSHPassword)),
					},
					// TODO: add a proper verification
					HostKeyCallback: gossh.InsecureIgnoreHostKey(),
				}, nil
			},
		},
		&common.StepProvision{},
		&StepShutdown{
			Command:    b.config.ShutdownCommand,
			ShutdownTimeout: b.config.ShutdownTimeout,
		},
		&StepCreateSnapshot{
			createSnapshot: b.config.CreateSnapshot,
		},
		&StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
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

	artifact := &Artifact{
		VMName: b.config.VMName,
		Conn: state.Get("vm").(*object.VirtualMachine),
	}
	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
