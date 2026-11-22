// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// This is the content to copy to `destination`. If destination is a file,
	// content will be written to that file, in case of a directory a file named
	// `pkr-file-content` is created. It's recommended to use a file as the
	// destination. The `templatefile` function might be used here, or any
	// interpolation syntax. This attribute cannot be specified with source or
	// sources.
	Content string `mapstructure:"content" required:"true"`
	// The path to a local file or directory to upload to the
	// machine. The path can be absolute or relative. If it is relative, it is
	// relative to the working directory when Packer is executed. If this is a
	// directory, the existence of a trailing slash is important. Read below on
	// uploading directories. Mandatory unless `sources` is set.
	Source string `mapstructure:"source" required:"true"`
	// A list of sources to upload. This can be used in place of the `source`
	// option if you have several files that you want to upload to the same
	// place. Note that the destination must be a directory with a trailing
	// slash, and that all files listed in `sources` will be uploaded to the
	// same directory with their file names preserved.
	Sources []string `mapstructure:"sources" required:"false"`
	// The path where the file will be uploaded to in the machine. This value
	// must be a writable location and any parent directories
	// must already exist. If the provisioning user (generally not root) cannot
	// write to this directory, you will receive a "Permission Denied" error.
	// If the source is a file, it's a good idea to make the destination a file
	// as well, but if you set your destination as a directory, at least make
	// sure that the destination ends in a trailing slash so that Packer knows
	// to use the source's basename in the final upload path. Failure to do so
	// may cause Packer to fail on file uploads. If the destination file
	// already exists, it will be overwritten.
	Destination string `mapstructure:"destination" required:"true"`
	// The direction of the file transfer. This defaults to "upload". If it is
	// set to "download" then the file "source" in the machine will be
	// downloaded locally to "destination"
	Direction string `mapstructure:"direction" required:"false"`
	// A list of glob patterns (matched against the slash-separated path
	// relative to the source, and against the base name) selecting which
	// files to transfer. When a source is a directory, only matching files
	// inside it are transferred. When empty, everything is transferred.
	// Cannot be used together with `content`.
	Include []string `mapstructure:"include" required:"false"`
	// A list of glob patterns (matched against the slash-separated path
	// relative to the source, and against the base name) selecting which
	// files to skip. Exclude wins over include. When a source is a
	// directory, the patterns are also passed to the communicator's
	// UploadDir/DownloadDir. Cannot be used together with `content`.
	Exclude []string `mapstructure:"exclude" required:"false"`
	// For advanced users only. If true, check the file existence only before
	// uploading, rather than upon pre-build validation. This allows users to
	// upload files created on-the-fly. This defaults to false. We
	// don't recommend using this feature, since it can cause Packer to become
	// dependent on system state. We would prefer you generate your files before
	// the Packer run, but realize that there are situations where this may be
	// unavoidable.
	Generated bool `mapstructure:"generated" required:"false"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "file",
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

	var errs *packersdk.MultiError

	if p.config.Direction != "download" && p.config.Direction != "upload" {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Direction must be one of: download, upload."))
	}
	if p.config.Source != "" {
		p.config.Sources = append(p.config.Sources, p.config.Source)
	}

	if p.config.Direction == "upload" {
		for _, src := range p.config.Sources {
			if _, err := os.Stat(src); p.config.Generated == false && err != nil {
				errs = packersdk.MultiErrorAppend(errs,
					fmt.Errorf("Bad source '%s': %s", src, err))
			}
		}
	}

	if len(p.config.Sources) > 0 && p.config.Content != "" {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("source(s) conflicts with content."))
	}

	if p.config.Content != "" && (len(p.config.Include) > 0 || len(p.config.Exclude) > 0) {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("include/exclude cannot be used with content."))
	}

	if p.config.Destination == "" {
		errs = packersdk.MultiErrorAppend(errs,
			errors.New("Destination must be specified."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// matchesAnyPattern reports whether rel (a slash-separated path) or its base
// name matches any of the glob patterns.
func matchesAnyPattern(patterns []string, rel string) bool {
	base := path.Base(rel)
	for _, pat := range patterns {
		if ok, err := path.Match(pat, rel); err == nil && ok {
			return true
		}
		if ok, err := path.Match(pat, base); err == nil && ok {
			return true
		}
	}
	return false
}

// passesFilters reports whether a file identified by rel (slash-separated,
// relative to the source root, or the source itself) should be transferred
// given the include/exclude globs. Exclude wins over include; an empty
// include list means "everything".
func passesFilters(include, exclude []string, rel string) bool {
	rel = path.Clean(rel)
	if len(exclude) > 0 && matchesAnyPattern(exclude, rel) {
		return false
	}
	if len(include) > 0 && !matchesAnyPattern(include, rel) {
		return false
	}
	return true
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	if generatedData == nil {
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	if p.config.Content != "" {
		file, err := tmp.File("pkr-file-content")
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := file.WriteString(p.config.Content); err != nil {
			return err
		}
		p.config.Content = ""
		p.config.Sources = append(p.config.Sources, file.Name())
	}

	if p.config.Direction == "download" {
		return p.ProvisionDownload(ui, comm)
	} else {
		return p.ProvisionUpload(ui, comm)
	}
}

// uploadDirFiltered uploads the files of srcDir that pass the include/exclude
// filters, one by one, preserving the directory layout under dst.
func (p *Provisioner) uploadDirFiltered(ui packersdk.Ui, comm packersdk.Communicator, dst, srcDir string) error {
	return filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return err
		}
		if !passesFilters(p.config.Include, p.config.Exclude, filepath.ToSlash(rel)) {
			return nil
		}
		ui.Say(fmt.Sprintf("Uploading %s => %s", filePath, filepath.Join(dst, rel)))
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		fi := info
		return comm.Upload(filepath.Join(dst, rel), f, &fi)
	})
}

func (p *Provisioner) ProvisionDownload(ui packersdk.Ui, comm packersdk.Communicator) error {
	dst, err := interpolate.Render(p.config.Destination, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("Error interpolating destination: %s", err)
	}
	for _, src := range p.config.Sources {
		dst := dst
		src, err := interpolate.Render(src, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Error interpolating source: %s", err)
		}

		isDirOrGlob := strings.HasSuffix(src, "/") || strings.ContainsAny(src, "*?[")
		if !isDirOrGlob && !passesFilters(p.config.Include, p.config.Exclude, filepath.ToSlash(src)) {
			ui.Say(fmt.Sprintf("Skipping %s (filtered by include/exclude)", src))
			continue
		}

		// ensure destination dir exists.  p.config.Destination may either be a file or a dir.
		dir := dst
		// if it doesn't end with a /, set dir as the parent dir
		if !strings.HasSuffix(dst, "/") {
			dir = filepath.Dir(dir)
		} else if !strings.HasSuffix(src, "/") && !strings.HasSuffix(src, "*") {
			dst = filepath.Join(dst, filepath.Base(src))
		}
		ui.Say(fmt.Sprintf("Downloading %s => %s", src, dst))

		if dir != "" {
			err := os.MkdirAll(dir, os.FileMode(0755))
			if err != nil {
				return err
			}
		}
		// if the src was a dir or glob, download the dir; the communicator
		// applies the exclude patterns to the remote listing.
		if isDirOrGlob {
			return comm.DownloadDir(src, dst, p.config.Exclude)
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

func (p *Provisioner) ProvisionUpload(ui packersdk.Ui, comm packersdk.Communicator) error {
	dst, err := interpolate.Render(p.config.Destination, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("Error interpolating destination: %s", err)
	}
	for _, src := range p.config.Sources {
		src, err := interpolate.Render(src, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Error interpolating source: %s", err)
		}

		ui.Say(fmt.Sprintf("Uploading %s => %s", src, dst))

		info, err := os.Stat(src)
		if err != nil {
			return err
		}

		// If we're uploading a directory, short circuit and do that
		if info.IsDir() {
			if len(p.config.Include) > 0 {
				// UploadDir cannot express an allow-list, so expand the
				// directory locally and upload each included file.
				if err := p.uploadDirFiltered(ui, comm, dst, src); err != nil {
					ui.Error(fmt.Sprintf("Upload failed: %s", err))
					return err
				}
			} else if err = comm.UploadDir(dst, src, p.config.Exclude); err != nil {
				ui.Error(fmt.Sprintf("Upload failed: %s", err))
				return err
			}
			continue
		}

		// File (or glob-free) source: apply include/exclude to the entry itself.
		if !passesFilters(p.config.Include, p.config.Exclude, filepath.ToSlash(src)) {
			ui.Say(fmt.Sprintf("Skipping %s (filtered by include/exclude)", src))
			continue
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

		filedst := dst
		if strings.HasSuffix(dst, "/") {
			filedst = dst + filepath.Base(src)
		}

		pf := ui.TrackProgress(filepath.Base(src), 0, info.Size(), f)
		defer pf.Close()

		// Upload the file
		if err = comm.Upload(filedst, pf, &fi); err != nil {
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
