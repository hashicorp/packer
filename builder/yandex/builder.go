package yandex

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// The unique ID for this builder.
const BuilderID = "packer.yandex"

// Builder represents a Packer Builder.
type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c
	return warnings, nil
}

// Run executes a yandex Packer build and returns a packer.Artifact
// representing a Yandex Cloud machine image.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	driver, err := NewDriverYandexCloud(ui, b.config)

	if err != nil {
		return nil, err
	}

	// Set up the state
	state := &multistep.BasicStateBag{}
	state.Put("config", b.config)
	state.Put("driver", driver)
	state.Put("sdk", driver.SDK())
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("yc_%s.pem", b.config.PackerBuildName),
		},
		&stepCreateInstance{
			Debug:         b.config.PackerDebug,
			SerialLogFile: b.config.SerialLogFile,
		},
		&stepInstanceInfo{},
		&communicator.StepConnect{
			Config:    &b.config.Communicator,
			Host:      commHost,
			SSHConfig: b.config.Communicator.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Communicator,
		},
		&stepShutdown{
			Debug: b.config.PackerDebug,
		},
		&stepCreateImage{},
	}

	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// Report any errors
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	if _, ok := state.GetOk("image_id"); !ok {
		log.Println("Failed to find image_id in state. Bug?")
		return nil, nil
	}

	artifact := &ArtifactMini{
		imageID:     state.Get("image_id").(string),
		imageName:   state.Get("image_name").(string),
		imageFamily: state.Get("image_family").(string),
		config:      b.config,
	}
	return artifact, nil
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
