package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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


	if len(p.config.Sources) < 1 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Source must be specified."))
	}

	if p.config.Destination == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Destination must be specified."))
	}

	if p.config.Direction == "upload" {
		// Download all downloadable files
		for i, src := range p.config.Sources {
			// Convert src to valid URL if possible
			uploadURL, err := common.DownloadableURL(src)
			if err != nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf(
					"Failed to parse source: %s", src))
			}

			if !strings.HasPrefix(uploadURL, "file://") {
				uploadFile, err := ioutil.TempFile("", "packer")
				if err != nil {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf(
						"Failed to create temporary file to store source: %s", err))
				}

				defer os.Remove(uploadFile.Name())
				uploadFile.Close()

				// Download file to local temp path
				config := &common.DownloadConfig{
					Url: uploadURL,
					TargetPath: uploadFile.Name(),
					UserAgent: "Packer",
				}
				path, err, _ := download(config)
				if err != nil {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf(
						"Failed to download: %s", uploadURL))
				}

				// Set new path for current src value
				p.config.Sources[i] = path
				src = path
			}
			if _, err := os.Stat(src); err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("Bad source '%s': %s", src, err))
			}
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	if p.config.Direction == "download" {
		return p.ProvisionDownload(ui, comm)
	} else {
		return p.ProvisionUpload(ui, comm)
	}
}

func (p *Provisioner) ProvisionDownload(ui packer.Ui, comm packer.Communicator) error {
	for _, src := range p.config.Sources {
		ui.Say(fmt.Sprintf("Downloading %s => %s", src, p.config.Destination))
		// ensure destination dir exists.  p.config.Destination may either be a file or a dir.
		dir := p.config.Destination
		// if it doesn't end with a /, set dir as the parent dir
		if !strings.HasSuffix(p.config.Destination, "/") {
			dir = filepath.Dir(dir)
		}
		if dir != "" {
			err := os.MkdirAll(dir, os.FileMode(0755))
			if err != nil {
				return err
			}
		}
		// if the config.Destination was a dir, download the dir
		if strings.HasSuffix(p.config.Destination, "/") {
			return comm.DownloadDir(src, p.config.Destination, nil)
		}

		f, err := os.OpenFile(p.config.Destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		err = comm.Download(src, f)
		if err != nil {
			ui.Error(fmt.Sprintf("Download failed: %s", err))
			return err
		}
	}
	return nil
}

func (p *Provisioner) ProvisionUpload(ui packer.Ui, comm packer.Communicator) error {
	// Upload files and directories
	for _, src := range p.config.Sources {
		ui.Say(fmt.Sprintf("Uploading %s => %s", src, p.config.Destination))

		// Stat the path to determine whether it's a directory or file
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

		err = comm.Upload(p.config.Destination, f, &fi)
		if err != nil {
			ui.Error(fmt.Sprintf("Upload failed: %s", err))
			return err
		}
	}
	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}

func download(config *common.DownloadConfig) (string, error, bool) {
	// Blatantly stolen from common/step_download.go
	var path string
	download := common.NewDownloadClient(config)

	downloadCompleteCh := make(chan error, 1)
	go func() {
		var err error
		path, err = download.Get()
		downloadCompleteCh <- err
	}()

	for {
		select {
		case err := <-downloadCompleteCh:
			if err != nil {
				return "", err, true
			}
			return path, nil, true
		}
	}
}
