package common

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepExportVm struct {
	OutputDir  string
	SkipExport bool
}

func (s *StepExportVm) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

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

	// Store the path to the export directory for later steps
	exportPath := filepath.Join(s.OutputDir, vmName)
	state.Put("export_path", exportPath)

	return multistep.ActionContinue
}

func (s *StepExportVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
