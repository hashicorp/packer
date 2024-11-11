// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package hcp_sbom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Source is a required field that specifies the path to the SBOM file that
	// needs to be downloaded.
	// It can be a file path or a URL.
	Source string `mapstructure:"source" required:"true"`
	// Destination is an optional field that specifies the path where the SBOM
	// file will be downloaded to for the user.
	// The 'Destination' must be a writable location. If the destination is a file,
	// the SBOM will be saved or overwritten at that path. If the destination is
	// a directory, a file will be created within the directory to store the SBOM.
	// Any parent directories for the destination must already exist and be
	// writable by the provisioning user (generally not root), otherwise,
	// a "Permission Denied" error will occur. If the source path is a file,
	// it is recommended that the destination path be a file as well.
	Destination string `mapstructure:"destination"`
	ctx         interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "hcp-sbom",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError
	if p.config.Source == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("source must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// PackerSBOM is the type we write to the temporary JSON dump of the SBOM to
// be consumed by Packer core
type PackerSBOM struct {
	// RawSBOM is the raw data from the SBOM downloaded from the guest
	RawSBOM []byte `json:"raw_sbom"`
	// Format is the format detected by the provisioner
	//
	// Supported values: `spdx` or `cyclonedx`
	Format string `json:"format"`
}

func (p *Provisioner) Provision(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	log.Println("Starting to provision with `hcp-sbom` provisioner")

	if generatedData == nil {
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	src := p.config.Source

	pkrDst := generatedData["dst"].(string)
	if pkrDst == "" {
		return fmt.Errorf("packer destination path missing from configs: this is an internal error, which should be reported to be fixed.")
	}

	var buf bytes.Buffer
	if err := comm.Download(src, &buf); err != nil {
		ui.Errorf("download failed for SBOM file: %s", err)
		return err
	}

	format, err := ValidateSBOM(buf.Bytes())
	if err != nil {
		return fmt.Errorf("validation failed for SBOM file: %s", err)
	}

	outFile, err := os.Create(pkrDst)
	if err != nil {
		return fmt.Errorf("failed to open/create output file %q: %s", pkrDst, err)
	}
	defer outFile.Close()

	err = json.NewEncoder(outFile).Encode(PackerSBOM{
		RawSBOM: buf.Bytes(),
		Format:  format,
		Name:    p.config.SbomName,
	})
	if err != nil {
		return fmt.Errorf("failed to write sbom file to %q: %s", pkrDst, err)
	}

	if p.config.Destination == "" {
		return nil
	}

	// SBOM for User
	usrDst, err := p.getUserDestination()
	if err != nil {
		return fmt.Errorf("failed to compute destination path %q: %s", p.config.Destination, err)
	}
	err = os.WriteFile(usrDst, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write SBOM to destination %q: %s", usrDst, err)
	}

	return nil
}

// getUserDestination determines and returns the destination path for the user SBOM file.
func (p *Provisioner) getUserDestination() (string, error) {
	dst := p.config.Destination

	// Check if the destination exists and determine its type
	info, err := os.Stat(dst)
	if err == nil {
		if info.IsDir() {
			// If the destination is a directory, create a temporary file inside it
			tmpFile, err := os.CreateTemp(dst, "packer-user-sbom-*.json")
			if err != nil {
				return "", fmt.Errorf("failed to create temporary file in user SBOM directory %s: %s", dst, err)
			}
			dst = tmpFile.Name()
			tmpFile.Close()
		}
		return dst, nil
	}

	outDir := filepath.Dir(dst)
	// In case the destination does not exist, we'll get the dirpath,
	// and create it if it doesn't already exist
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create destination directory for user SBOM: %s\n", err)
	}

	// Check if the destination is a directory after the previous step.
	//
	// This happens if the path specified ends with a `/`, in which case the
	// destination is a directory, and we must create a temporary file in
	// this destination directory.
	destStat, statErr := os.Stat(dst)
	if statErr == nil && destStat.IsDir() {
		tmpFile, err := os.CreateTemp(outDir, "packer-user-sbom-*.json")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file in user SBOM directory %s: %s", dst, err)
		}
		dst = tmpFile.Name()
		tmpFile.Close()
	}

	return dst, nil
}
