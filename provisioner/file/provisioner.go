package file

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	Checksum string
	ChecksumType string `mapstructure:"checksum_type"`

	// The remote path where the file will be uploaded/downloaded to.
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
		// Create tempdir for downloads
		cache := &packer.FileCache{CacheDir: "packer_cache"}

		// Download all downloadable files
		for i, src := range p.config.Sources {
			// Convert src to valid URL if possible
			uploadURL, err := common.DownloadableURL(src)
			if err != nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf(
					"Failed to parse source: %s", src))
			}

			if !strings.HasPrefix(uploadURL, "file://") {
				// Prepare download location
				cacheKey := filepath.Base(uploadURL)
				targetPath := cache.Lock(cacheKey)
				newPath := filepath.Join(filepath.Dir(targetPath), cacheKey)
				defer cache.Unlock(cacheKey)

				var checksum []byte
				if p.config.Checksum != "" {
					var err error
					checksum, err = hex.DecodeString(p.config.Checksum)
					if err != nil {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf(
							"Error parsing checksum: %s", err))
					}
				}

				// Prepare download config
				client := common.NewDownloadClient(&common.DownloadConfig{
					Url: uploadURL,
					TargetPath: targetPath,
					Checksum: checksum,
					Hash: common.HashForType(p.config.ChecksumType),
					UserAgent: "Packer",
				})

				// Download if local file doesn't exist for checksum invalid
				if verified, _ := client.VerifyChecksum(newPath); !verified {
					path, err, _ := download(client)
					if err != nil {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf(
							"Failed to download %s: %s", uploadURL, err))
					}

					// Rename file back to original file name
					err = os.Rename(path, newPath)

					if err != nil {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf(
							"Failed to rename downloaded %s: %s", targetPath, err))
					}
				}

				// Set new path for current src value
				p.config.Sources[i] = newPath
			}

			// Verify that the given source file exists
			if _, err := os.Stat(p.config.Sources[i]); err != nil {
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
		dst := p.config.Destination

		ui.Say(fmt.Sprintf("Uploading %s => %s", src, dst))

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

		if strings.HasSuffix(dst, "/") {
			dst = filepath.Join(dst, filepath.Base(src))
		}

		err = comm.Upload(dst, f, &fi)
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

func download(download *common.DownloadClient) (string, error, bool) {
	// Blatantly stolen from common/step_download.go
	var path string

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
