package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCollateArtifacts struct {
	OutputDir  string
	SkipExport bool
}

// Runs the step required to collate all build artifacts under the
// specified output directory
func (s *StepCollateArtifacts) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
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

	return multistep.ActionContinue
}

// Cleanup does nothing
func (s *StepCollateArtifacts) Cleanup(state multistep.StateBag) {}
