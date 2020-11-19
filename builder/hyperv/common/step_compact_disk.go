package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCompactDisk struct {
	SkipCompaction bool
}

// Run runs a compaction/optimisation process on attached VHD/VHDX disks
func (s *StepCompactDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if s.SkipCompaction {
		ui.Say("Skipping disk compaction...")
		return multistep.ActionContinue
	}

	// Get the dir used to store the VMs files during the build process
	var buildDir string
	if v, ok := state.GetOk("build_dir"); ok {
		buildDir = v.(string)
	}

	ui.Say("Compacting disks...")
	// CompactDisks searches for all VHD/VHDX files under the supplied
	// path and runs the compacting process on each of them. If no disks
	// are found under the supplied path this is treated as a 'soft' error
	// and a warning message is printed. All other errors halt the build.
	result, err := driver.CompactDisks(buildDir)
	if err != nil {
		err := fmt.Errorf("Error compacting disks: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// Report disk compaction results/warn if no disks were found
	ui.Message(result)

	return multistep.ActionContinue
}

// Cleanup does nothing
func (s *StepCompactDisk) Cleanup(state multistep.StateBag) {}
