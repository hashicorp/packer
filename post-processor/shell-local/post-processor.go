package shell_local

import (
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
)

type PostProcessor struct {
	config sl.Config
}

type ExecuteCommandTemplate struct {
	Vars   string
	Script string
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}

	return sl.Validate(&p.config)
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	// this particular post-processor doesn't do anything with the artifact
	// except to return it.

	retBool, retErr := sl.Run(ui, &p.config)
	if !retBool {
		return nil, retBool, retErr
	}

	return artifact, retBool, retErr
}
