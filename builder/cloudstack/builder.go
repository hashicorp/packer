package cloudstack

import (
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&stepKeypair{
			Debug:                b.config.PackerDebug,
			DebugKeyPath:         fmt.Sprintf("cs_%s.pem", b.config.PackerBuildName),
			KeyPair:              b.config.Keypair,
			PrivateKeyFile:       b.config.Comm.SSHPrivateKey,
			SSHAgentAuth:         b.config.Comm.SSHAgentAuth,
			TemporaryKeyPairName: b.config.TemporaryKeypairName,
		},
		&stepCreateSecurityGroup{},
		&stepCreateInstance{
			Ctx:   b.config.ctx,
			Debug: b.config.PackerDebug,
		},
		&stepSetupNetworking{},
		&communicator.StepConnect{
			Config: &b.config.Comm,
			Host:   commHost,
			SSHConfig: sshConfig(
				b.config.Comm.SSHAgentAuth,
				b.config.Comm.SSHUsername,
				b.config.Comm.SSHPassword),
			SSHPort:   commPort,
			WinRMPort: commPort,
		},
		&common.StepProvision{},
		&stepShutdownInstance{},
		&stepCreateTemplate{},
	}

	// Configure the runner and run the steps.
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		ui.Error(rawErr.(error).Error())
		return nil, rawErr.(error)
	}

	// If there was no template created, just return
	if _, ok := state.GetOk("template"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		client:   client,
		config:   b.config,
		template: state.Get("template").(*cloudstack.CreateTemplateResponse),
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
