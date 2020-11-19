// Package oci contains a packer.Builder implementation that builds Oracle
// Bare Metal Cloud Services (OCI) images.
package oci

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	ocommon "github.com/hashicorp/packer/builder/oracle/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/oracle/oci-go-sdk/core"
)

// BuilderId uniquely identifies the builder
const BuilderId = "packer.oracle.oci"

// OCI API version
const ociAPIVersion = "20160918"

// Builder is a builder implementation that creates Oracle OCI custom images.
type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := b.config.Prepare(raws...)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {
	driver, err := NewDriverOCI(&b.config)
	if err != nil {
		return nil, err
	}

	// Populate the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
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
			Host:      communicator.CommHost(b.config.Comm.Host(), "instance_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepImage{},
	}

	// Run the steps
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
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
		Image:     image.(core.Image),
		Region:    region,
		driver:    driver,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}

// Cancel terminates a running build.
