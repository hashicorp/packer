package virtualbox

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

const BuilderId = "mitchellh.virtualbox"

type Builder struct {
	config config
	driver Driver
	runner multistep.Runner
}

type config struct {
	OutputDir string `mapstructure:"output_directory"`
}

func (b *Builder) Prepare(raw interface{}) error {
	var err error
	if err := mapstructure.Decode(raw, &b.config); err != nil {
		return err
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = "virtualbox"
	}

	errs := make([]error, 0)

	b.driver, err = b.newDriver()
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed creating VirtualBox driver: %s", err))
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) packer.Artifact {
	steps := []multistep.Step{
		new(stepPrepareOutputDir),
		new(stepSuppressMessages),
	}

	// Setup the state bag
	state := make(map[string]interface{})
	state["cache"] = cache
	state["config"] = &b.config
	state["driver"] = b.driver
	state["hook"] = hook
	state["ui"] = ui

	// Run
	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(state)

	return nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) newDriver() (Driver, error) {
	vboxmanagePath := "VBoxManage"
	driver := &VBox42Driver{vboxmanagePath}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}
