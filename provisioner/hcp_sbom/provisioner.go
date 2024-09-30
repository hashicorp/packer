// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package hcp_sbom

import (
	"context"
	"errors"

	"fmt"
	"log"
	"os"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"

	"path/filepath"
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

func (p *Provisioner) Provision(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	log.Printf(
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
		return fmt.Errorf("failed to download Packer SBOM file: %w", downloadErr)
	}

	// Download the file for user
	downloadErr = p.downloadSBOMForUser(ui, comm)
	if downloadErr != nil {
		return fmt.Errorf("failed to download User SBOM file: %w", downloadErr)
	}

	// Validate the file
	log.Printf(fmt.Sprintf("Validating SBOM file: %s\n", destPath))
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

	// Download the file for Packer
	dst, ok := generatedData["dst"].(string) // this has been set by HCPSBOMInternalProvisioner.Provision
	if !ok || dst == "" {
		return "", fmt.Errorf("destination path for Packer SBOM file is not valid")
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return dst, fmt.Errorf("failed to create destination directory for Packer SBOM: %w", err)
	}

	// Open the destination file
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return dst, fmt.Errorf("failed to open destination file for Packer SBOM: %s", err)
	}
	defer f.Close()

	// Download the file
	ui.Say(fmt.Sprintf("Downloading SBOM file %s for Packer => %s", src, dst))
	if err = comm.Download(src, f); err != nil {
		ui.Error(fmt.Sprintf("download failed for Packer SBOM file: %s", err))
		return dst, err
	}

	return dst, nil
}

// downloadSBOMForUser downloads a Software Bill of Materials (SBOM) file from a specified source
// to a local destination path on the machine.
func (p *Provisioner) downloadSBOMForUser(
	ui packersdk.Ui, comm packersdk.Communicator,
) error {
	dst := p.config.Destination
	if dst == "" {
		log.Println("skipped downloading user SBOM file because 'Destination' is not provided")
		return nil
	}

	dst, err := interpolate.Render(dst, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating SBOM file destination from user: %s\n", err)
	}

	src, err := interpolate.Render(p.config.Source, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	// Check if the destination exists and determine its type
	info, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			// If destination doesn't exist, assume it's a file path and ensure parent directories are created
			dir := filepath.Dir(dst)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory for user SBOM: %s\n", err)
			}
		} else {
			return fmt.Errorf("failed to stat destination for user SBOM: %s\n", err)
		}
	} else if info.IsDir() {
		// If the destination is a directory, create a temporary file inside it
		tmpFile, err := os.CreateTemp(dst, "packer-user-sbom-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file in user SBOM directory %s: %s", dst, err)
		}
		dst = tmpFile.Name()
		tmpFile.Close()
	}

	// Open the destination file
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination file for user SBOM: %s", err)
	}
	defer f.Close()

	// Download the file
	ui.Say(fmt.Sprintf("Downloading SBOM file for user %s => %s", src, dst))
	if err = comm.Download(src, f); err != nil {
		return fmt.Errorf("download failed for user SBOM file: %s", err)
	}

	ui.Say(fmt.Sprintf("User SBOM file successfully downloaded to: %s\n", dst))
	return nil
}

// validateSBOM validates CycloneDX SBOM files
func (p *Provisioner) validateSBOM(ui packersdk.Ui, filePath string) error {
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer sourceFile.Close()

	decoder := cyclonedx.NewBOMDecoder(sourceFile, cyclonedx.BOMFileFormatJSON)
	bom := new(cyclonedx.BOM)
	if err := decoder.Decode(bom); err != nil {
		return fmt.Errorf("failed to decode CycloneDX SBOM: %w", err)
	}

	if bom.BOMFormat != "CycloneDX" {
		return fmt.Errorf("invalid bomFormat: %s, expected CycloneDX", bom.BOMFormat)
	}
	if bom.SpecVersion.String() == "" {
		return fmt.Errorf("specVersion is required")
	}

	return nil
}
