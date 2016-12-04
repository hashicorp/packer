package cloudstack

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

const BuilderId = "packer.cloudstack"

// Builder represents the CloudStack builder.
type Builder struct {
	config *Config
	runner multistep.Runner
	ui     packer.Ui
}

// Prepare implements the packer.Builder interface.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	config, errs := NewConfig(raws...)
	if errs != nil {
		return nil, errs
	}
	b.config = config

	return nil, nil
}

// Run implements the packer.Builder interface.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	b.ui = ui

	// Create a CloudStack API client.
	client := cloudstack.NewAsyncClient(
		b.config.APIURL,
		b.config.APIKey,
		b.config.SecretKey,
		!b.config.SSLNoVerify,
	)

	// Set the time to wait before timing out
	client.AsyncTimeout(int64(b.config.AsyncTimeout.Seconds()))

	// Some CloudStack service providers only allow HTTP GET calls.
	client.HTTPGETOnly = b.config.HTTPGetOnly

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("client", client)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&stepPrepareConfig{},
		&stepCreateInstance{},
		&stepSetupNetworking{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: sshConfig,
		},
		&common.StepProvision{},
		&stepShutdownInstance{},
		&stepCreateTemplate{},
	}

	// Configure the runner.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	// Run the steps.
	b.runner.Run(state)

	// If there are no templates, then just return
	template, ok := state.Get("template").(*cloudstack.CreateTemplateResponse)
	if !ok || template == nil {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		client:   client,
		config:   b.config,
		template: template,
	}

	return artifact, nil
}

// Cancel the step runner.
func (b *Builder) Cancel() {
	if b.runner != nil {
		b.ui.Say("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
