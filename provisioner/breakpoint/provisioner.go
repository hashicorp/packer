package breakpoint

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Note    string `mapstructure:"note"`
	Disable bool   `mapstructure:"disable"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	if p.config.Disable {
		if p.config.Note != "" {
			ui.Say(fmt.Sprintf(
				"Breakpoint provisioner with note \"%s\" disabled; continuing...",
				p.config.Note))
		} else {
			ui.Say("Breakpoint provisioner disabled; continuing...")
		}

		return nil
	}
	if p.config.Note != "" {
		ui.Say(fmt.Sprintf("Pausing at breakpoint provisioner with note \"%s\".", p.config.Note))
	} else {
		ui.Say("Pausing at breakpoint provisioner.")
	}

	message := fmt.Sprintf(
		"Press enter to continue.")

	result := make(chan string, 1)
	go func() {
		line, err := ui.Ask(message)
		if err != nil {
			log.Printf("Error asking for input: %s", err)
		}

		result <- line
	}()

	select {
	case <-result:
		return nil
	}
}

func (p *Provisioner) Cancel() {
	// Just hard quit.
	os.Exit(0)
}
