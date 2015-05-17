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

	// Direction
	Direction string

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

	if p.config.Direction == "" {
		p.config.Direction = "upload"
	}

	if p.config.Direction != "download" && p.config.Direction != "upload" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Direction must be one of: download, upload."))
	}

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

	if p.config.Direction == "upload" {
		if _, err := os.Stat(p.config.Source); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Bad source '%s': %s", p.config.Source, err))
		}
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
	if p.config.Direction == "download" {
		return p.ProvisionDownload( ui, comm)
	} else {
		return p.ProvisionUpload( ui, comm)
	}
}

func (p *Provisioner) ProvisionDownload(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Downloading %s => %s", p.config.Source, p.config.Destination))

	f, err := os.OpenFile(p.config.Destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	err = comm.Download(p.config.Source, f)
	if err != nil {
		ui.Error(fmt.Sprintf("Download failed: %s", err))
	}
	return err
}

func (p *Provisioner) ProvisionUpload(ui packer.Ui, comm packer.Communicator) error {
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

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	err = comm.Upload(p.config.Destination, f, &fi)
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
