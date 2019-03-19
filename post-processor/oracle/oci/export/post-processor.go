package export

import (
	"github.com/hashicorp/packer/packer"
)

const BuilderId = "packer.post-processor.oracle-oci-export"

// Configuration of this post processor

type PostProcessor struct {
	config *Config
}

// Entry point for configuration parsing when we've defined
func (p *PostProcessor) Configure(raws ...interface{}) error {

	config, err := NewConfig(raws...)
	if err != nil {
		return err
	}
	p.config = config
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	ui.Say(p.config.CompartmentID)
	cli, err := NewExportClient(p.config)
	if err != nil {
		return artifact, false, err
	}
	_, err = cli.ExportImage(artifact.Id())
	if err != nil {
		return artifact, false, err
	}
	ui.Say("Image successfully exported to Object storage")
	return artifact, true, nil
}
