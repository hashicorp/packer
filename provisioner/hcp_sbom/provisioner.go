// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package hcp_sbom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Source              string `mapstructure:"source" required:"true"`
	Destination         string `mapstructure:"destination"`
	ctx                 interpolate.Context
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

func (p *Provisioner) Provision(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	ui.Say(
		fmt.Sprintf("Starting to provision with hcp-sbom using source: %s",
			p.config.Source,
		),
	)

	if generatedData == nil {
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	// Download the file for Packer
	destPath, downloadErr := p.downloadSBOMForPacker(ui, comm, generatedData)
	if downloadErr != nil {
		return fmt.Errorf("failed to download file: %w", downloadErr)
	}

	// Download the file for user
	p.downloadSBOMForUser(ui, comm)

	// Validate the file
	ui.Say(fmt.Sprintf("Validating SBOM file %s", destPath))
	validationErr := p.validateSBOM(ui, destPath)
	if validationErr != nil {
		return fmt.Errorf("failed to validate SBOM file: %w", validationErr)
	}

	return nil
}

// downloadSBOMForPacker downloads SBOM from a specified source to a local
// destination set by internal SBOM provisioner. It works with all communicators
// from packersdk.
func (p *Provisioner) downloadSBOMForPacker(
	ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{},
) (string, error) {
	src, err := interpolate.Render(p.config.Source, &p.config.ctx)
	if err != nil {
		return p.config.Destination, fmt.Errorf("error interpolating source: %s", err)
	}

	// FIXME:: Do we really need this?
	// Check if the source is a JSON file
	if filepath.Ext(src) != ".json" {
		return p.config.Destination, fmt.Errorf(
			"packer SBOM source file is not a JSON file: %s", src,
		)
	}

	// Download the file for Packer
	desti, ok := generatedData["dst"] // this has been set by HCPSBOMInternalProvisioner.Provision
	if !ok {
		return "", fmt.Errorf("failed to find location for Packer SBOM file")
	}

	dst := fmt.Sprintf("%v", desti)
	// Ensure the destination directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
		return dst, fmt.Errorf("failed to create destination directory for Packer SBOM: %s", err)
	}

	// Open the destination file
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return dst, fmt.Errorf("failed to open destination file: %s", err)
	}
	defer f.Close()

	// Create MultiWriter for the current progress
	pf := io.MultiWriter(f)

	// Download the file
	ui.Say(fmt.Sprintf("Downloading SBOM file %s for Packer => %s", src, dst))
	if err = comm.Download(src, pf); err != nil {
		ui.Error(fmt.Sprintf("download failed for Packer SBOM file: %s", err))
		return dst, err
	}

	return dst, nil
}

// downloadSBOMForUser downloads a SBOM from a specified source to a local
// destination given by user. It works with all communicators from packersdk.
func (p *Provisioner) downloadSBOMForUser(
	ui packersdk.Ui, comm packersdk.Communicator,
) {
	src, err := interpolate.Render(p.config.Source, &p.config.ctx)
	if err != nil {
		ui.Say(fmt.Sprintf("error interpolating source: %s", err))
		return
	}

	// Determine the destination path
	dst := p.config.Destination
	if dst == "" {
		ui.Say("skipped downloading SBOM file for user because 'Destination' is not provided")
		return
	}

	dst, err = interpolate.Render(dst, &p.config.ctx)
	if err != nil {
		ui.Say(fmt.Sprintf("error interpolating SBOM file destination: %s", err))
		return
	}

	if strings.HasSuffix(dst, "/") {
		info, err := os.Stat(dst)
		if err != nil {
			ui.Say(fmt.Sprintf("failed to stat destination for SBOM: %s", err))
			return
		}

		if info.IsDir() {
			tmpFile, err := os.CreateTemp(dst, "packer-user-sbom-*.json")
			if err != nil {
				ui.Say(fmt.Sprintf("failed to create file for Packer SBOM: %s", err))
				return
			}
			dst = tmpFile.Name()
			tmpFile.Close()
		}
	}

	// Ensure the destination directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
		ui.Say(fmt.Sprintf("failed to create destination directory for Packer SBOM: %s", err))
		return
	}

	// Open the destination file
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		ui.Say(fmt.Sprintf("failed to open destination file: %s", err))
		return
	}
	defer f.Close()

	// Create MultiWriter for the current progress
	pf := io.MultiWriter(f)

	// Download the file
	ui.Say(fmt.Sprintf("Downloading SBOM file for user %s => %s", src, dst))
	if err = comm.Download(src, pf); err != nil {
		ui.Error(fmt.Sprintf("download failed for user SBOM file: %s", err))
		return
	}
}

type SBOM struct {
	BomFormat   string `json:"bomFormat"`
	SpecVersion string `json:"specVersion"`
}

func (p *Provisioner) validateSBOM(ui packersdk.Ui, filePath string) error {
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	data, err := io.ReadAll(sourceFile)
	if err != nil {
		return err
	}

	var sbom SBOM
	if err := json.Unmarshal(data, &sbom); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if sbom.BomFormat != "CycloneDX" {
		return fmt.Errorf("invalid bomFormat: %s", sbom.BomFormat)
	}

	if sbom.SpecVersion == "" {
		return fmt.Errorf("specVersion is required")
	}

	return nil
}
