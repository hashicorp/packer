package file

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"os"
	"sort"
	"strings"
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
	var md mapstructure.Metadata
	decoderConfig := &mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &p.config,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	for _, raw := range raws {
		err := decoder.Decode(raw)
		if err != nil {
			return err
		}
	}

	// Accumulate any errors
	errs := make([]error, 0)

	// Unused keys are errors
	if len(md.Unused) > 0 {
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			if unused != "type" && !strings.HasPrefix(unused, "packer_") {
				errs = append(
					errs, fmt.Errorf("Unknown configuration key: %s", unused))
			}
		}
	}

	if _, err := os.Stat(p.config.Source); err != nil {
		errs = append(errs,
			fmt.Errorf("Bad source '%s': %s", p.config.Source, err))
	}

	if p.config.Destination == "" {
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
	defer f.Close()

	return comm.Upload(p.config.Destination, f)
}
