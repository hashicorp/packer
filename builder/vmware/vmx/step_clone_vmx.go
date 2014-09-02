package vmx

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// StepCloneVMX takes a VMX file and clones the VM into the output directory.
type StepCloneVMX struct {
	OutputDir string
	Path      string
	VMName    string
}

func (s *StepCloneVMX) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
        src := s.Path
	vmxPath := filepath.Join(s.OutputDir, s.VMName+".vmx")

	ui.Say("Cloning source VM...")

        vagrant_box := common.NewVagrantBox(src)
        if vagrant_box != nil {
                var err error
                src, err = vagrant_box.Expand(".vmx")
                defer vagrant_box.Clean()
                if err != nil {
                        err := fmt.Errorf("Error expanding Vagrant Box: %s", err)
                        state.Put("error", err)
                        ui.Error(err.Error())
                        return multistep.ActionHalt
                }
        }

	log.Printf("Cloning from: %s", src)
	log.Printf("Cloning to: %s", vmxPath)
	if err := driver.Clone(vmxPath, src); err != nil {
	         state.Put("error", err)
	         return multistep.ActionHalt
	}

	vmxData, err := vmwcommon.ReadVMX(vmxPath)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	diskName, ok := vmxData["scsi0:0.filename"]
	if !ok {
		err := fmt.Errorf("Root disk filename could not be found!")
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("full_disk_path", filepath.Join(s.OutputDir, diskName))
	state.Put("vmx_path", vmxPath)
	return multistep.ActionContinue
}

func (s *StepCloneVMX) Cleanup(state multistep.StateBag) {
}
