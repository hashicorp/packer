package common

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step configures a VMX by setting some default settings as well
// as taking in custom data to set, attaching a floppy if it exists, etc.
//
// Uses:
//   vmx_path string
//
// Produces:
//   display_name string - Value of the displayName key set in the VMX file
type StepConfigureVMX struct {
	CustomData  map[string]string
	DisplayName string
	SkipFloppy  bool
	VMName      string
}

func (s *StepConfigureVMX) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	log.Printf("Configuring VMX...\n")

	var err error
	ui := state.Get("ui").(packer.Ui)

	vmxPath := state.Get("vmx_path").(string)
	vmxData, err := ReadVMX(vmxPath)
	if err != nil {
		err := fmt.Errorf("Error reading VMX file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set this so that no dialogs ever appear from Packer.
	vmxData["msg.autoanswer"] = "true"

	// Create a new UUID for this VM, since it is a new VM
	vmxData["uuid.action"] = "create"

	// Delete any generated addresses since we want to regenerate
	// them. Conflicting MAC addresses is a bad time.
	addrRegex := regexp.MustCompile(`(?i)^ethernet\d+\.generatedAddress`)
	for k := range vmxData {
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

	// If the build is taking place on a remote ESX server, the displayName
	// will be needed for discovery of the VM's IP address and for export
	// of the VM. The displayName key should always be set in the VMX file,
	// so error if we don't find it and the user has not set it in the config.
	if s.DisplayName != "" {
		vmxData["displayname"] = s.DisplayName
		state.Put("display_name", s.DisplayName)
	} else {
		displayName, ok := vmxData["displayname"]
		if !ok { // Packer converts key names to lowercase!
			err := fmt.Errorf("Error: Could not get value of displayName from VMX data")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
			state.Put("display_name", displayName)
		}
	}

	err = WriteVMX(vmxPath, vmxData)

	if err != nil {
		err := fmt.Errorf("Error writing VMX file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigureVMX) Cleanup(state multistep.StateBag) {
}
