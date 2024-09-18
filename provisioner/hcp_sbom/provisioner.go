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
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/klauspost/compress/zstd"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{},
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

	// Download the file
	destPath, downloadErr := p.downloadSBOM(ui, comm)
	// defer os.Remove(destPath)
	if downloadErr != nil {
		return fmt.Errorf("failed to download file: %w", downloadErr)
	}

	// Validate the file
	ui.Say(fmt.Sprintf("Validating SBOM file %s", destPath))
	validationErr := p.validateSBOM(ui, destPath)
	if validationErr != nil {
		return fmt.Errorf("failed to validate SBOM file: %w", validationErr)
	}

	// Compress the file
	ui.Say(fmt.Sprintf("Compressing SBOM file %s", destPath))
	_, compessionErr := p.compressFile(ui, destPath)
	if compessionErr != nil {
		return fmt.Errorf("failed to compress file: %w", compessionErr)
	}

	// Future: send compressedData to the internal API as per RFC
	// ...

	return nil
}

// downloadSBOM downloads a Software Bill of Materials (SBOM) from a specified
// source to a local destination. It works with all communicators from packersdk.
// The method returns the path to the downloaded file or an error if any issues
// occur during the download process.
func (p *Provisioner) downloadSBOM(ui packersdk.Ui, comm packersdk.Communicator) (string, error) {
	src, err := interpolate.Render(p.config.Source, &p.config.ctx)
	if err != nil {
		return p.config.Destination, fmt.Errorf("error interpolating source: %s", err)
	}

	// Check if the source is a JSON file
	if filepath.Ext(src) != ".json" {
		return p.config.Destination, fmt.Errorf(
			"packer SBOM source file is not a JSON file: %s", src,
		)
	}

	// Determine the destination path
	dst := p.config.Destination
	if dst == "" {
		tmpFile, err := os.CreateTemp("", "packer-sbom-*.json")
		if err != nil {
			return dst, fmt.Errorf(
				"failed to create file for Packer SBOM: %s", err,
			)
		}
		dst = tmpFile.Name()
		tmpFile.Close()
	} else {
		dst, err = interpolate.Render(dst, &p.config.ctx)
		if err != nil {
			return dst, fmt.Errorf("error interpolating Packer SBOM destination: %s", err)
		}

		if strings.HasSuffix(dst, "/") {
			info, err := os.Stat(dst)
			if err != nil {
				return dst, fmt.Errorf("failed to stat destination for Packer SBOM: %s", err)
			}

			if info.IsDir() {
				tmpFile, err := os.CreateTemp(dst, "packer-sbom-*.json")
				if err != nil {
					return dst, fmt.Errorf("failed to create file for Packer SBOM: %s", err)
				}
				dst = tmpFile.Name()
				tmpFile.Close()
			}
		}
	}

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
	ui.Say(fmt.Sprintf("Downloading SBOM file %s => %s", src, dst))
	if err = comm.Download(src, pf); err != nil {
		ui.Error(fmt.Sprintf("download failed for SBOM file: %s", err))
		return dst, err
	}

	return dst, nil
}

func (p *Provisioner) compressFile(ui packersdk.Ui, filePath string) ([]byte, error) {
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer sourceFile.Close()

	data, err := io.ReadAll(sourceFile)
	if err != nil {
		return nil, err
	}

	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	compressedData := encoder.EncodeAll(data, nil)

	ui.Say(fmt.Sprintf("SBOM file compressed successfully. Size: %d bytes", len(compressedData)))
	return compressedData, nil
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
