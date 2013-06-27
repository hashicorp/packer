package vagrant

import (
	"github.com/mitchellh/packer/packer"
)

type AWSBoxPostProcessor struct {
}

func (p *AWSBoxPostProcessor) Configure(raw interface{}) error {
	return nil
}

func (p *AWSBoxPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, error) {
	return nil, nil
}
