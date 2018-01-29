// Package oci contains a packer.Builder implementation that builds Oracle
// Bare Metal Cloud Services (OCI) images.
package oci

import (
	"fmt"
	"log"

	client "github.com/hashicorp/packer/builder/oracle/oci/client"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// BuilderId uniquely identifies the builder
const BuilderId = "packer.oracle.oci"

// OCI API version
const ociAPIVersion = "20160918"

// Builder is a builder implementation that creates Oracle OCI custom images.
type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(rawConfig ...interface{}) ([]string, error) {
	config, err := NewConfig(rawConfig...)
	if err != nil {
		return nil, err
	}
	b.config = config

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	driver, err := NewDriverOCI(b.config)
	if err != nil {
		return nil, err
	}

	// Populate the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepKeyPair{
			Debug:          b.config.PackerDebug,
			DebugKeyPath:   fmt.Sprintf("oci_%s.pem", b.config.PackerBuildName),
			PrivateKeyFile: b.config.Comm.SSHPrivateKey,
		},
		&stepCreateInstance{},
		&stepInstanceInfo{},
		&communicator.StepConnect{
			Config: &b.config.Comm,
			Host:   commHost,
			SSHConfig: SSHConfig(
				b.config.Comm.SSHUsername,
				b.config.Comm.SSHPassword),
		},
		&common.StepProvision{},
		&stepImage{},
	}

	// Run the steps
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{
		Image:  state.Get("image").(client.Image),
		Region: b.config.AccessCfg.Region,
		driver: driver,
	}

	return artifact, nil
}

// Cancel terminates a running build.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
