package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step configures a VMX by setting some default settings as well
// as taking in custom data to set, attaching a floppy if it exists, etc.
//
// Uses:
//   vmx_path string
type StepConfigureVMX struct {
	CustomData map[string]string
	SkipFloppy bool
}

func (s *StepConfigureVMX) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	vmxContents, err := ioutil.ReadFile(vmxPath)
	if err != nil {
		err := fmt.Errorf("Error reading VMX file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vmxData := ParseVMX(string(vmxContents))

	// Set this so that no dialogs ever appear from Packer.
	vmxData["msg.autoanswer"] = "true"

	// Create a new UUID for this VM, since it is a new VM
	vmxData["uuid.action"] = "create"

	// Delete any generated addresses since we want to regenerate
	// them. Conflicting MAC addresses is a bad time.
	addrRegex := regexp.MustCompile(`(?i)^ethernet\d+\.generatedAddress`)
	for k, _ := range vmxData {
		if addrRegex.MatchString(k) {
			delete(vmxData, k)
		}
	}

	// Set custom data
	for k, v := range s.CustomData {
		log.Printf("Setting VMX: '%s' = '%s'", k, v)
		k = strings.ToLower(k)
		vmxData[k] = v
	}

	// Set a floppy disk, but only if we should
	if !s.SkipFloppy {
		// Set a floppy disk if we have one
		if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
			log.Println("Floppy path present, setting in VMX")
			vmxData["floppy0.present"] = "TRUE"
			vmxData["floppy0.filetype"] = "file"
			vmxData["floppy0.filename"] = floppyPathRaw.(string)
		}
	}

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error writing VMX file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigureVMX) Cleanup(state multistep.StateBag) {
}
