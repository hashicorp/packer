//go:generate mapstructure-to-hcl2 -type Config

package wimexport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ImageName  string `mapstructure:"image_name"`
	OutputPath string `mapstructure:"output"`
	ctx        interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{"output"},
		},
	}, raws...)
	if err != nil {

		fmt.Println(err)

		return err
	}
	errs := new(packersdk.MultiError)

	if p.config.ImageName == "" {
		errs = packersdk.MultiErrorAppend(errs,
			fmt.Errorf("image_name not provided"))
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.BuilderType}}.wim"
	}

	if err = interpolate.Validate(p.config.OutputPath, &p.config.ctx); err != nil {
		fmt.Println(err)

		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	if len(errs.Errors) > 0 {
		fmt.Println(err)

		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	files := artifact.Files()

	var generatedData map[interface{}]interface{}
	stateData := artifact.State("generated_data")
	if stateData != nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = stateData.(map[interface{}]interface{})
	}
	// If stateData has a nil map generatedData will be nil
	// and we need to make sure it's not
	if generatedData == nil {
		generatedData = make(map[interface{}]interface{})
	}
	generatedData["BuildName"] = p.config.PackerBuildName
	generatedData["BuilderType"] = p.config.PackerBuilderType

	newartifact := NewArtifact(artifact.Files())

	for _, art := range files {
		art = filepath.ToSlash(art)

		if filepath.Ext(art) != ".vhd" && filepath.Ext(art) != ".vhdx" {
			continue
		}

		wimFile, err := interpolate.Render(p.config.OutputPath, &p.config.ctx)
		if err != nil {
			return nil, false, true, err
		}

		if _, err := os.Stat(wimFile); err != nil {
			newartifact.files = append(newartifact.files, wimFile)
		}
		if err := os.MkdirAll(filepath.Dir(wimFile), os.FileMode(0755)); err != nil {
			return nil, false, true, fmt.Errorf("unable to create dir: %s", err.Error())
		}

		driveLetter, err := p.mountVhd(art)
		if err != nil {
			return nil, false, true, fmt.Errorf("unable to mount %s: %s", art, err.Error())
		}

		defer p.dismountVhd(art)

		capturePath := fmt.Sprintf("%s:/", driveLetter)

		ui.Say(fmt.Sprintf("Capturing %s to %s", capturePath, wimFile))

		if p.newWindowsImage(wimFile, capturePath, p.config.ImageName); err != nil {
			return nil, false, true, fmt.Errorf("unable to capture %s: %s", art, err.Error())
		} else {
			// sets keep and forceOverride to true because we don't want to accidentally
			// delete the very artifact we're creating.
			return newartifact, true, true, nil
		}
	}

	return nil, false, true, fmt.Errorf("no vhd/vhdx found")
}

func (p *PostProcessor) mountVhd(path string) (string, error) {

	var script = `
param([string]$path)
$vhd = Mount-Vhd -Path $path -ReadOnly -PassThru
if ($vhd -ne $null) {
    $osVol = $vhd | Get-Disk | Get-Partition | Get-Volume | Where-Object { $_.DriveLetter -ne $null -AND $_.FileSystemType -eq "NTFS" }
	if ($osVol -ne $null) {
		return $osVol.DriveLetter
	}
}
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, path)
	if err != nil {
		return "", err
	}

	var driveLetter = strings.TrimSpace(cmdOut)
	return driveLetter, nil
}

func (p *PostProcessor) newWindowsImage(imagePath string, capturePath string, name string) error {

	var script = `
param([string]$imagePath,[string]$capturePath,[string]$name)
New-WindowsImage -ImagePath $imagePath -CapturePath $capturePath -Name $name
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, imagePath, capturePath, name)
	return err
}

func (p *PostProcessor) dismountVhd(path string) error {
	var script = `
param([string]$path)
$vhd = Dismount-Vhd -Path $path
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, path)
	return err
}
