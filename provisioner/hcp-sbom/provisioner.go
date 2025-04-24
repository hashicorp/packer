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
	"regexp"
	"strings"

	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The file path or URL to the SBOM file in the Packer artifact.
	// This file must either be in the SPDX or CycloneDX format.
	Source string `mapstructure:"source" required:"true"`

	// The path on the local machine to store a copy of the SBOM file.
	// You can specify an absolute or a path relative to the working directory
	// when you execute the Packer build. If the file already exists on the
	// local machine, Packer overwrites the file. If the destination is a
	// directory, the directory must already exist.
	Destination string `mapstructure:"destination"`

	// The name of the SBOM file stored in HCP Packer.
	// If omitted, HCP Packer uses the build fingerprint as the file name.
	// This value must be between three and 36 characters from the following set: `[A-Za-z0-9_-]`.
	// You must specify a unique name for each build in an artifact version.
	SbomName string `mapstructure:"sbom_name"`
	ctx      interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

var sbomFormatRegexp = regexp.MustCompile("^[0-9A-Za-z-]{3,36}$")

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

	var errs error

	if p.config.Source == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("source must be specified"))
	}

	if p.config.SbomName != "" && !sbomFormatRegexp.MatchString(p.config.SbomName) {
		// Ugly but a bit of a problem with interpolation since Provisioners
		// are prepared twice in HCL2.
		//
		// If the information used for interpolating is populated in-between the
		// first call to Prepare (at the start of the build), and when the
		// Provisioner is actually called, the first call will fail, as
		// the value won't contain the actual interpolated value, but a
		// placeholder which doesn't match the regex.
		//
		// Since we don't have a way to discriminate between the calls
		// in the context of the provisioner, we ignore them, and later the
		// HCP Packer call will fail because of the broken regex.
		if strings.Contains(p.config.SbomName, "<no value>") {
			log.Printf("[WARN] interpolation incomplete for `sbom_name`, will possibly retry later with data populated into context, otherwise will fail when uploading to HCP Packer.")
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("`sbom_name` %q doesn't match the expected format, it must "+
				"contain between 3 and 36 characters, all from the following set: [A-Za-z0-9_-]", p.config.SbomName))
		}
	}

	return errs
}

// PackerSBOM is the type we write to the temporary JSON dump of the SBOM to
// be consumed by Packer core
type PackerSBOM struct {
	// RawSBOM is the raw data from the SBOM downloaded from the guest
	RawSBOM []byte `json:"raw_sbom"`
	// Format is the format detected by the provisioner
	//
	// Supported values: `SPDX` or `CYCLONEDX`
	Format hcpPackerModels.HashicorpCloudPacker20230101SbomFormat `json:"format"`
	// Name is the name of the SBOM to be set on HCP Packer
	//
	// If unset, HCP Packer will generate one
	Name string `json:"name,omitempty"`
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

	format, err := validateSBOM(buf.Bytes())
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
