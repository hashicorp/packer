// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package hcp_sbom

import (
	"bytes"
	"context"
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

func (p *Provisioner) Provision(
	ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator,
	generatedData map[string]interface{},
) error {
	log.Println("Starting to provision with `hcp_sbom` provisioner")

	if generatedData == nil {
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	downloadErr := p.downloadAndValidateSBOM(ui, comm, generatedData)
	if downloadErr != nil {
		return fmt.Errorf("failed to download SBOM file: %w", downloadErr)
	}

	return nil
}

// downloadAndValidateSBOM handles downloading SBOM files for the User and Packer.
func (p *Provisioner) downloadAndValidateSBOM(
	ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{},
) error {
	src, err := interpolate.Render(p.config.Source, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	var buf bytes.Buffer
	if err = comm.Download(src, &buf); err != nil {
		ui.Error(fmt.Sprintf("download failed for SBOM file: %s", err))
		return err
	}

	pkrBuf := bytes.NewBuffer(buf.Bytes())
	usrBuf := bytes.NewBuffer(buf.Bytes())
	if _, err = ValidateSBOM(&buf); err != nil {
		ui.Error(fmt.Sprintf("validation failed for SBOM file: %s", err))
		return err
	}

	// SBOM for Packer
	pkrDst, err := p.getPackerDestination(generatedData)
	if err != nil {
		return fmt.Errorf("failed to get Packer SBOM destination: %s", err)
	}

	err = p.writeToFile(pkrBuf, pkrDst)
	if err != nil {
		return fmt.Errorf("failed to download Packer SBOM: %s", err)
	}
	log.Printf("Packer SBOM file successfully downloaded to: %s\n", pkrDst)

	// SBOM for User
	usrDst, err := p.getUserDestination()
	if err != nil {
		return fmt.Errorf("failed to determine user SBOM destination: %s", err)
	}

	if usrDst != "" {
		err = p.writeToFile(usrBuf, usrDst)
		if err != nil {
			return fmt.Errorf("failed to download User SBOM: %s", err)
		}
		log.Printf("User SBOM file successfully downloaded to: %s\n", usrDst)
	}
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

func (p *Provisioner) writeToFile(buf *bytes.Buffer, dst string) error {
	// Open the destination file
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination file for SBOM: %s", err)
	}
	defer f.Close()

	// Write the buffer content to the destination file
	if _, err = buf.WriteTo(f); err != nil {
		return err
	}

	return nil
}
