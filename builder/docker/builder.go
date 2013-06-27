// Docker builder
package docker

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

const BuilderId = "geetarista.docker"

type config struct {
	Repository string `mapstructure:"repository"`
	BuildPath  string `mapstructure:"build_path"`
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) error {
	for _, raw := range raws {
		err := mapstructure.Decode(raw, &b.config)
		if err != nil {
			return err
		}
	}

	errs := make([]error, 0)

	if b.config.Repository == "" {
		errs = append(errs, errors.New("a repository is required"))
	}

	if b.config.BuildPath == "" {
		b.config.BuildPath = "."
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	steps := []multistep.Step{
		new(stepBuild),
	}

	state := map[string]interface{}{
		"config": b.config,
		"ui":     ui,
	}

	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(state)

	log.Println("Docker builder complete. Returning artifact.")

	return &Artifact{b.config.Repository}, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
