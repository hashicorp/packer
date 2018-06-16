package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepExportVm struct {
	OutputDir      string
	SkipCompaction bool
	SkipExport     bool
}

func (s *StepExportVm) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// Get the VM name; Get the temp directory used to store the VMs files
	// during the build process
	var vmName, tmpPath string
	if v, ok := state.GetOk("vmName"); ok {
		vmName = v.(string)
	}
	if v, ok := state.GetOk("packerTempDir"); ok {
		tmpPath = v.(string)
	}

	// Compact disks first so the export process has less to do
	if s.SkipCompaction {
		ui.Say("Skipping disk compaction...")
	} else {
		ui.Say("Compacting disks...")
		err := driver.CompactDisks(tmpPath)
		if err != nil {
			err := fmt.Errorf("Error compacting disks: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	ui.Say("Exporting virtual machine...")
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

	return multistep.ActionContinue
}

func (s *StepExportVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
