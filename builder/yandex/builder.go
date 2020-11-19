package yandex

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/packerbuilderdata"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
)

// The unique ID for this builder.
const BuilderID = "packer.yandex"

// Builder represents a Packer Builder.
type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}
	generatedData := []string{
		"ImageID",
		"ImageName",
		"ImageFamily",
		"ImageDescription",
		"ImageFolderID",
		"SourceImageID",
		"SourceImageName",
		"SourceImageDescription",
		"SourceImageFamily",
		"SourceImageFolderID",
	}
	return generatedData, warnings, nil
}

// Run executes a yandex Packer build and returns a packersdk.Artifact
// representing a Yandex.Cloud compute image.
func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {
	driver, err := NewDriverYC(ui, &b.config.AccessConfig)
	ctx = requestid.ContextWithClientTraceID(ctx, uuid.New().String())

	if err != nil {
		return nil, err
	}

	// Set up the state
	state := &multistep.BasicStateBag{}
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("sdk", driver.SDK())
	state.Put("hook", hook)
	state.Put("ui", ui)
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	// Build the steps
	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("yc_%s.pem", b.config.PackerBuildName),
		},
		&StepCreateInstance{
			Debug:         b.config.PackerDebug,
			SerialLogFile: b.config.SerialLogFile,
			GeneratedData: generatedData,
		},
		&stepInstanceInfo{},
		&communicator.StepConnect{
			Config:    &b.config.Communicator,
			Host:      commHost,
			SSHConfig: b.config.Communicator.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Communicator,
		},
		&StepTeardownInstance{},
		&stepCreateImage{
			GeneratedData: generatedData,
		},
	}

	// Run the steps
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// Report any errors
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	image, ok := state.GetOk("image")
	if !ok {
		return nil, fmt.Errorf("Failed to find 'image' in state. Bug?")
	}

	artifact := &Artifact{
		Image:     image.(*compute.Image),
		config:    &b.config,
		driver:    driver,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
