package wim

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hypervcommon "github.com/hashicorp/packer/builder/hyperv/common"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

type StepCollateArtifacts struct {
	OutputDir  string
	SkipExport bool

	// It's set when the vhd is mounted, unset when dismounted.
	vhdPath string
}

func (s *StepCollateArtifacts) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Collating build artifacts...")

	if s.SkipExport {
		// Get the path to the main build directory from the statebag
		var buildDir string
		if v, ok := state.GetOk("build_dir"); ok {
			buildDir = v.(string)
		}
		// If the user has chosen to skip a full export of the VM the only
		// artifacts that they are interested in will be the VHDs. The
		// called function searches for all disks under the given source
		// directory and moves them to a 'Virtual Hard Disks' folder under
		// the destination directory
		err := driver.MoveCreatedVHDsToOutputDir(buildDir, s.OutputDir)
		if err != nil {
			err = fmt.Errorf("Error moving VHDs from build dir to output dir: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		// Get the full path to the export directory from the statebag
		var exportPath string
		if v, ok := state.GetOk("export_path"); ok {
			exportPath = v.(string)
		}
		// The export process exports the VM into a folder named 'vm name'
		// under the output directory. However, to maintain backwards
		// compatibility, we now need to shuffle around the exported folders
		// so the 'Snapshots', 'Virtual Hard Disks' and 'Virtual Machines'
		// directories appear *directly* under <output directory>.
		// The empty '<output directory>/<vm name>' directory is removed
		// when complete.
		// The 'Snapshots' folder will not be moved into the output
		// directory if it is empty.
		err := driver.PreserveLegacyExportBehaviour(exportPath, s.OutputDir)
		if err != nil {
			// No need to halt here; Just warn the user instead
			err = fmt.Errorf("WARNING: Error restoring legacy export dir structure: %s", err)
			ui.Error(err.Error())
		}
	}

	// Get the dir used to store the VMs files during the build process
	var buildDir string
	if v, ok := state.GetOk("build_dir"); ok {
		buildDir = v.(string)
	}

	var files []string
	filepath.Walk(buildDir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString("\\.vhd$|\\.vhdx$", f.Name())
			if err == nil && r {
				files = append(files, path)
			}
		}
		return nil
	})

	if len(files) != 1 {
		err := fmt.Errorf("Error finding Vhd/x")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	driveLetter, err := s.mountVhd(files[0])
	if err != nil {
		err = fmt.Errorf("Error mounting Vhd: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vhdPath = files[0]

	defer s.dismountVhd(s.vhdPath)

	wimPath := filepath.Join(s.OutputDir, "packer.wim")
	capturePath := fmt.Sprintf("%s:/", driveLetter)

	if s.newWindowsImage(wimPath, capturePath, "packer"); err != nil {
		err = fmt.Errorf("Error exporting WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCollateArtifacts) Cleanup(state multistep.StateBag) {
	_ = s.dismountVhd(s.vhdPath)
}

func (s *StepCollateArtifacts) mountVhd(path string) (string, error) {

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

func (s *StepCollateArtifacts) newWindowsImage(imagePath string, capturePath string, name string) error {

	var script = `
param([string]$imagePath,[string]$capturePath,[string]$name)
New-WindowsImage -ImagePath $imagePath -CapturePath $capturePath -Name $name
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, imagePath, capturePath, name)
	return err
}

func (s *StepCollateArtifacts) dismountVhd(path string) error {
	if s.vhdPath == "" {
		return nil
	}

	var script = `
param([string]$path)
$vhd = Dismount-Vhd -Path $path
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, path)
	if err == nil {
		s.vhdPath = ""
	}

	return err
}
