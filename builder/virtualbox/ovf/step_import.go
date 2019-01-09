package ovf

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step imports an OVF VM into VirtualBox.
type StepImport struct {
	Name        string
	ImportFlags []string

	vmName     string
	extractDir string
}

func (s *StepImport) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmPath := state.Get("vm_path").(string)

	// if it's a vagrant box, extract the box and read the ovf inside.
	if strings.HasSuffix(vmPath, ".box") {
		// Use Dir of the vm_path to extract box to
		vmDir := filepath.Dir(vmPath)
		s.extractDir = filepath.Join(vmDir, s.Name)
		err := os.Mkdir(s.extractDir, 0777)
		if err != nil {
			err = fmt.Errorf("Trouble extracting vagrant box for use: %s", err.Error())
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		commonhelper.UntarBox(s.extractDir, vmPath)
		vmPath = filepath.Join(s.extractDir, "box.ovf")

		log.Printf("vmPath is %s", vmPath)
		if _, err := os.Stat(vmPath); os.IsNotExist(err) {
			err = fmt.Errorf("Could not find a box.ovf inside of the given " +
				"vagrant box; please make sure that this box has virtualbox " +
				"as its provider")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say(fmt.Sprintf("Importing VM: %s", vmPath))
	if err := driver.Import(s.Name, vmPath, s.ImportFlags); err != nil {
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
	config := state.Get("config").(*Config)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (config.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with VirtualBox host (keep_registered = true)")
		return
	}

	ui.Say("Deregistering and deleting imported VM...")
	if err := driver.Delete(s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %s", err))
	}

	// clean up extracted version of vagrant box
	if s.extractDir != "" {
		os.RemoveAll(s.extractDir)
	}
}
