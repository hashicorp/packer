package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type stepCloneVMX struct{}

func (s stepCloneVMX) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	vmxPath := filepath.Join(config.OutputDir, config.VMName+".vmx")

	ui.Say("Cloning and writing VMX file")

	vmxData, err := s.readVMX(config.SourceVMXPath)
	if err != nil {
		err := fmt.Errorf("Error reading Source VMX file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// clear mac address (should do a test for all mac addresses)
	delete(vmxData, "ethernet0.generatedAddress")
	delete(vmxData, "ethernet0.generatedAddressOffset")
	vmxData["uuid.action"] = "create"

	if config.VMXData != nil {
		log.Println("Setting custom VMX data...")
		for k, v := range config.VMXData {
			log.Printf("Setting VMX: '%s' = '%s'", k, v)
			vmxData[k] = v
		}
	}

	if floppyPathRaw, ok := state["floppy_path"]; ok {
		log.Println("Floppy path present, setting in VMX")
		vmxData["floppy0.present"] = "TRUE"
		vmxData["floppy0.fileType"] = "file"
		vmxData["floppy0.fileName"] = floppyPathRaw.(string)
	}

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error creating Destination VMX file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["vmx_path"] = vmxPath

	return multistep.ActionContinue
}

func (stepCloneVMX) Cleanup(map[string]interface{}) {}

func (stepCloneVMX) readVMX(vmxPath string) (map[string]string, error) {
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
