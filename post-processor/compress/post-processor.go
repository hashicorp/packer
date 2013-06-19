// compress implements the packer.PostProcessor interface and adds a
// post-processor for compressing output.
package compress

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
)

type Config struct {
	Format string
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raw interface{}) error {
	if err := mapstructure.Decode(raw, &p.config); err != nil {
		return err
	}

	if p.config.Format == "" {
		p.config.Format = "tar.gz"
	}

	if p.config.Format != "tar.gz" {
		return errors.New("only 'tar.gz' is a supported format right now")
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, error) {
	ui.Say("We made it to here.")
	return nil, nil
}
