package ovf

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
        "github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// This step imports an OVF VM into VirtualBox.
type StepImport struct {
	Name       string
	SourcePath string
	ImportOpts string

	vmName string
}

func (s *StepImport) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
        src := s.SourcePath

	ui.Say(fmt.Sprintf("Importing VM: %s", src))

        vagrant_box := common.NewVagrantBox(src)
        if vagrant_box != nil {
                var err error
                src, err = vagrant_box.Expand(".ovf")
                defer vagrant_box.Clean()
                if err != nil {
	                err := fmt.Errorf("Error expanding Vagrant Box: %s", err)
	                state.Put("error", err)
	                ui.Error(err.Error())
	                return multistep.ActionHalt
                }
        }

	if err := driver.Import(s.Name, src, s.ImportOpts); err != nil {
	         err := fmt.Errorf("Error importing VM: %s", err)
	         state.Put("error", err)
	         ui.Error(err.Error())
	         return multistep.ActionHalt
        }

	s.vmName = s.Name
	state.Put("vmName", s.Name)
	return multistep.ActionContinue
}

func (s *StepImport) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering and deleting imported VM...")
	if err := driver.Delete(s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %s", err))
	}
}
