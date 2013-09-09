package file

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
)

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The local path of the file to upload.
	Source string

	// The remote path where the local file will be uploaded to.
	Destination string

	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"source":      &p.config.Source,
		"destination": &p.config.Destination,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if _, err := os.Stat(p.config.Source); err != nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Bad source '%s': %s", p.config.Source, err))
	}

	if p.config.Destination == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Destination must be specified."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Uploading %s => %s", p.config.Source, p.config.Destination))
	info, err := os.Stat(p.config.Source)
	if err != nil {
		return err
	}

	// If we're uploading a directory, short circuit and do that
	if info.IsDir() {
		return comm.UploadDir(p.config.Destination, p.config.Source, nil)
	}

	// We're uploading a file...
	f, err := os.Open(p.config.Source)
	if err != nil {
		return err
	}
	defer f.Close()

	err = comm.Upload(p.config.Destination, f)
	if err != nil {
		ui.Error(fmt.Sprintf("Upload failed: %s", err))
	}
	return err
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
