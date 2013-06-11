package virtualbox

import (
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

const BuilderId = "mitchellh.virtualbox"

type Builder struct {
	config config
	runner multistep.Runner
}

type config struct {
	OutputDir string `mapstructure:"output_directory"`
}

func (b *Builder) Prepare(raw interface{}) error {
	if err := mapstructure.Decode(raw, &b.config); err != nil {
		return err
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) packer.Artifact {
	return nil
}

func (b *Builder) Cancel() {
}
