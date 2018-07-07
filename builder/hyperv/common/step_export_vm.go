package common

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepExportVm struct {
	OutputDir  string
	SkipExport bool
}

func (s *StepExportVm) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	ui.Say("Exporting virtual machine...")

	// The VM name is needed for the export command
	var vmName string
	if v, ok := state.GetOk("vmName"); ok {
		vmName = v.(string)
	}

	// The export process exports the VM to a folder named 'vmName' under
	// the output directory. This contains the usual 'Snapshots', 'Virtual
	// Hard Disks' and 'Virtual Machines' directories.
	err := driver.ExportVirtualMachine(vmName, s.OutputDir)
	if err != nil {
		err = fmt.Errorf("Error exporting vm: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Shuffle around the exported folders to maintain backwards
	// compatibility. This moves the 'Snapshots', 'Virtual Hard Disks' and
	// 'Virtual Machines' directories from <output directory>/<vm name> so
	// they appear directly under <output directory>. The empty '<output
	// directory>/<vm name>' directory is removed when complete.
	// The 'Snapshots' folder will not be moved into the output directory
	// if it is empty.
	exportPath := filepath.Join(s.OutputDir, vmName)
	err = driver.PreserveLegacyExportBehaviour(exportPath, s.OutputDir)
	if err != nil {
		// No need to halt here; Just warn the user instead
		err = fmt.Errorf("WARNING: Error restoring legacy export dir structure: %s", err)
		ui.Error(err.Error())
	}

	return multistep.ActionContinue
}

func (s *StepExportVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
