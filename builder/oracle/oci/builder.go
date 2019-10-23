// Package oci contains a packer.Builder implementation that builds Oracle
// Bare Metal Cloud Services (OCI) images.
package oci

import (
	"context"
	"fmt"

	ocommon "github.com/hashicorp/packer/builder/oracle/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/oracle/oci-go-sdk/core"
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

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
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
		&ocommon.StepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.Comm,
			DebugKeyPath: fmt.Sprintf("oci_%s.pem", b.config.PackerBuildName),
		},
		&stepCreateInstance{},
		&stepInstanceInfo{},
		&stepGetDefaultCredentials{
			Debug:     b.config.PackerDebug,
			Comm:      &b.config.Comm,
			BuildName: b.config.PackerBuildName,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost(b.config.Comm.SSHHost, "instance_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepImage{},
	}

	// Run the steps
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	region, err := b.config.configProvider.Region()
	if err != nil {
		return nil, err
	}

	image, ok := state.GetOk("image")
	if !ok {
		return nil, err
	}

	// Build the artifact and return it
	artifact := &Artifact{
		Image:  image.(core.Image),
		Region: region,
		driver: driver,
	}

	return artifact, nil
}

// Cancel terminates a running build.
