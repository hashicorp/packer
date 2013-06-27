// vagrant implements the packer.PostProcessor interface and adds a
// post-processor that turns artifacts of known builders into Vagrant
// boxes.
package vagrant

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
)

var builtins = map[string]string{
	"mitchellh.amazonebs": "aws",
}

type Config struct {
	OutputPath string `mapstructure:"output"`
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raw interface{}) error {
	err := mapstructure.Decode(raw, &p.config)
	if err != nil {
		return err
	}

	if p.config.OutputPath == "" {
		return fmt.Errorf("`output` must be specified.")
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, error) {
	ppName, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	// Get the actual PostProcessor implementation for this type
	var pp packer.PostProcessor
	switch ppName {
	case "aws":
		pp = new(AWSBoxPostProcessor)
	default:
		return nil, fmt.Errorf("Vagrant box post-processor not found: %s", ppName)
	}

	// Prepare and run the post-processor
	config := map[string]string{"output": p.config.OutputPath}
	if err := pp.Configure(config); err != nil {
		return nil, err
	}

	return pp.PostProcess(ui, artifact)
}
