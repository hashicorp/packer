//go:generate mapstructure-to-hcl2 -type Config

package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The local path of the file to upload.
	Source  string
	Sources []string

	// The remote path where the local file will be uploaded to.
	Destination string

	// Direction
	Direction string

	// False if the sources have to exist.
	Generated bool

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

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

	if p.config.Direction == "" {
		p.config.Direction = "upload"
	}

	var errs *packer.MultiError

	if p.config.Direction != "download" && p.config.Direction != "upload" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Direction must be one of: download, upload."))
	}
	if p.config.Source != "" {
		p.config.Sources = append(p.config.Sources, p.config.Source)
	}

	if p.config.Direction == "upload" {
		for _, src := range p.config.Sources {
			if _, err := os.Stat(src); p.config.Generated == false && err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("Bad source '%s': %s", src, err))
			}
		}
	}

	if len(p.config.Sources) < 1 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Source must be specified."))
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

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	if p.config.Direction == "download" {
		return p.ProvisionDownload(ui, comm)
	} else {
		return p.ProvisionUpload(ui, comm)
	}
}

func (p *Provisioner) ProvisionDownload(ui packer.Ui, comm packer.Communicator) error {
	for _, src := range p.config.Sources {
		dst := p.config.Destination
		ui.Say(fmt.Sprintf("Downloading %s => %s", src, dst))
		// ensure destination dir exists.  p.config.Destination may either be a file or a dir.
		dir := dst
		// if it doesn't end with a /, set dir as the parent dir
		if !strings.HasSuffix(dst, "/") {
			dir = filepath.Dir(dir)
		} else if !strings.HasSuffix(src, "/") && !strings.HasSuffix(src, "*") {
			dst = filepath.Join(dst, filepath.Base(src))
		}
		if dir != "" {
			err := os.MkdirAll(dir, os.FileMode(0755))
			if err != nil {
				return err
			}
		}
		// if the src was a dir, download the dir
		if strings.HasSuffix(src, "/") || strings.ContainsAny(src, "*?[") {
			return comm.DownloadDir(src, dst, nil)
		}

		f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		// Create MultiWriter for the current progress
		pf := io.MultiWriter(f)

		// Download the file
		if err = comm.Download(src, pf); err != nil {
			ui.Error(fmt.Sprintf("Download failed: %s", err))
			return err
		}
	}
	return nil
}

func (p *Provisioner) ProvisionUpload(ui packer.Ui, comm packer.Communicator) error {
	for _, src := range p.config.Sources {
		dst := p.config.Destination

		ui.Say(fmt.Sprintf("Uploading %s => %s", src, dst))

		info, err := os.Stat(src)
		if err != nil {
			return err
		}

		// If we're uploading a directory, short circuit and do that
		if info.IsDir() {
			return comm.UploadDir(p.config.Destination, src, nil)
		}

		// We're uploading a file...
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		if strings.HasSuffix(dst, "/") {
			dst = dst + filepath.Base(src)
		}

		pf := ui.TrackProgress(filepath.Base(src), 0, info.Size(), f)
		defer pf.Close()

		// Upload the file
		if err = comm.Upload(dst, pf, &fi); err != nil {
			if strings.Contains(err.Error(), "Error restoring file") {
				ui.Error(fmt.Sprintf("Upload failed: %s; this can occur when "+
					"your file destination is a folder without a trailing "+
					"slash.", err))
			}
			ui.Error(fmt.Sprintf("Upload failed: %s", err))
			return err
		}
	}
	return nil
}
