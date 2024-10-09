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

	// Download the files
	downloadErr := p.downloadSBOM(ui, comm, generatedData)
	if downloadErr != nil {
		return fmt.Errorf("failed to download SBOM file: %w", downloadErr)
	}

	return nil
}

// downloadSBOM handles downloading SBOM files for the User and Packer.
func (p *Provisioner) downloadSBOM(
	ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{},
) error {
	// Interpolate the source path
	src, err := interpolate.Render(p.config.Source, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	// Attempt to download SBOM for User
	dst, err := p.getUserDestination()
	if err != nil {
		return fmt.Errorf("failed to determine user SBOM destination: %s", err)
	}

	// If User SBOM destination is valid, try to download the SBOM file
	if dst != "" {
		ui.Say(fmt.Sprintf("Attempting to download SBOM file for User: %s", src))
		err = p.downloadToFile(ui, comm, src, dst)
		if err != nil {
			return fmt.Errorf("user SBOM download failed: %s", err)
		}
		ui.Say(fmt.Sprintf("User SBOM file successfully downloaded to: %s", dst))
	}

	// Attempt to download SBOM for Packer
	dst, err = p.getPackerDestination(generatedData)
	if err != nil {
		return fmt.Errorf("failed to get Packer SBOM destination: %s", err)
	}

	err = p.downloadToFile(ui, comm, src, dst)
	if err != nil {
		return fmt.Errorf("failed to download Packer SBOM: %s", err)
	}

	ui.Say(fmt.Sprintf("Packer SBOM file successfully downloaded to: %s", dst))
	return nil
}

// getUserDestination determines and returns the destination path for the user SBOM file.
func (p *Provisioner) getUserDestination() (string, error) {
	dst := p.config.Destination
	if dst == "" {
		log.Println("skipped downloading user SBOM file because 'Destination' is not provided")
		return "", nil
	}

	dst, err := interpolate.Render(dst, &p.config.ctx)
	if err != nil {
		return "", fmt.Errorf("error interpolating SBOM file destination for user: %s", err)
	}

	// Check if the destination exists and determine its type
	info, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			// If destination doesn't exist, assume it's a file path and ensure parent directories are created
			dir := filepath.Dir(dst)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("failed to create destination directory for user SBOM: %s\n", err)
			}
		} else {
			return "", fmt.Errorf("failed to stat destination for user SBOM: %s\n", err)
		}
	} else if info.IsDir() {
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

// getPackerDestination retrieves the destination path for the Packer SBOM file.
func (p *Provisioner) getPackerDestination(generatedData map[string]interface{}) (string, error) {
	dst, ok := generatedData["dst"].(string) // This has been set by HCPSBOMInternalProvisioner.Provision
	if !ok || dst == "" {
		return "", fmt.Errorf("destination path for Packer SBOM file is not valid")
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory for Packer SBOM: %w", err)
	}

	return dst, nil
}

// downloadToFile performs the actual download operation to the specified file destination.
func (p *Provisioner) downloadToFile(ui packersdk.Ui, comm packersdk.Communicator, src, dst string) error {
	// Open the destination file
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination file for SBOM: %s", err)
	}
	defer f.Close()

	// Download the file
	ui.Say(fmt.Sprintf("Downloading SBOM file %s => %s", src, dst))
	if err = comm.Download(src, f); err != nil {
		ui.Error(fmt.Sprintf("download failed for SBOM file: %s", err))
		return err
	}

	return nil
}