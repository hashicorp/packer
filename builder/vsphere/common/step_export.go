package common

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepExport struct {
	Format         string
	OutputPath     string
	SkipExport     bool
	OVFToolOptions []string
}

func (s *StepExport) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// Skip export if requested
	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	ui.Say("Exporting virtual machine...")
	log.Printf("Exporting %s in %s with these options: %#v", s.Format, s.OutputPath, s.OVFToolOptions)
	if err := driver.ExportVirtualMachine(s.OutputPath, s.Format, s.OVFToolOptions); err != nil {
		err := fmt.Errorf("Error exporting VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("exportPath", s.OutputPath)
	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
