package file

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"os"
)

type config struct {
	// The local path of the file to upload.
	Source string

	// The remote path where the local file will be uploaded to.
	Destination string
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	for _, raw := range raws {
		if err := mapstructure.Decode(raw, &p.config); err != nil {
			return err
		}
	}

	errs := []error{}

	if _, err := os.Stat(p.config.Source); err != nil {
		errs = append(errs, fmt.Errorf("Bad source file '%s': %s", p.config.Source, err))
	}

	if len(p.config.Destination) == 0 {
		errs = append(errs, errors.New("Destination must be specified."))
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}
	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Uploading %s => %s", p.config.Source, p.config.Destination))
	f, err := os.Open(p.config.Source)
	if err != nil {
		return err
	}
	return comm.Upload(p.config.Destination, f)
}
