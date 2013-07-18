package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
        "os"
	"io/ioutil"
)

// This step creates the virtual disks for the VM.
//
// Uses:
//   config *config
//   driver Driver
//   ui     packer.Ui
//
// Produces:
//   full_disk_path (string) - The full path to the created disk.
type stepCloneDisk struct{}

func (s stepCloneDisk) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	ui.Say("Cloning virtual machine disk")
	sourcevmxpath := config.SourceVMXPath
	full_disk_path := filepath.Join(config.OutputDir, config.DiskName+".vmdk")


        vmxData, err := s.readVMX(config.SourceVMXPath)
        if err != nil {
                err := fmt.Errorf("Error reading Source VMX file: %s", err)
                state["error"] = err
                ui.Error(err.Error())
                return multistep.ActionHalt
        }
	//ui.Say(fmt.Sprintf("Dir %s", filepath.Dir(sourcevmxpath)))
	//ui.Say(fmt.Sprintf("%s", vmxData["scsi0:0.fileName"]))
	full_source_disk_path := filepath.Join(filepath.Dir(sourcevmxpath), vmxData["scsi0:0.fileName"])
	// the full source disk pack is assuming vmxData is returning a relative path.  we need to check that assumption.

	if err := driver.CloneDisk(full_source_disk_path, full_disk_path, fmt.Sprintf("%dM", config.DiskSize)); err != nil {
		err := fmt.Errorf("Error cloning disk: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["full_disk_path"] = full_disk_path

	return multistep.ActionContinue
}

func (stepCloneDisk) Cleanup(map[string]interface{}) {}

func (stepCloneDisk) readVMX(vmxPath string) (map[string]string, error) {
        vmxF, err := os.Open(vmxPath)
        if err != nil {
                return nil, err
        }
        defer vmxF.Close()

        vmxBytes, err := ioutil.ReadAll(vmxF)
        if err != nil {
                return nil, err
        }

        return ParseVMX(string(vmxBytes)), nil
}
