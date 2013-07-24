package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

// This step cleans up the VMX by removing or changing this prior to
// being ready for use.
//
// Uses:
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type stepCleanVMX struct{}

func (s stepCleanVMX) Run(state map[string]interface{}) multistep.StepAction {
	isoPath, isoExists := state["iso_path"]
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	ui.Say("Cleaning VMX prior to finishing up...")

	vmxData, err := s.readVMX(vmxPath)
	if err != nil {
		state["error"] = fmt.Errorf("Error reading VMX: %s", err)
		return multistep.ActionHalt
	}

	if _, ok := state["floppy_path"]; ok {
		// Delete the floppy0 entries so the floppy is no longer mounted
		ui.Message("Unmounting floppy from VMX...")
		for k, _ := range vmxData {
			if strings.HasPrefix(k, "floppy0.") {
				log.Printf("Deleting key: %s", k)
				delete(vmxData, k)
			}
		}
		vmxData["floppy0.present"] = "FALSE"
	}

	if isoExists {
		ui.Message("Detatching ISO from CD-ROM device...")
		devRe := regexp.MustCompile(`^ide\d:\d\.`)
		for k, _ := range vmxData {
			match := devRe.FindString(k)
			if match == "" {
				continue
			}

			filenameKey := match + ".filename"
			if filename, ok := vmxData[filenameKey]; ok {
				if filename == isoPath.(string) {
					// Change the CD-ROM device back to auto-detect to eject
					vmxData[filenameKey] = "auto detect"
					vmxData[match+".deviceType"] = "cdrom-raw"
				}
			}
		}
	}

	// Rewrite the VMX
	if err := WriteVMX(vmxPath, vmxData); err != nil {
		state["error"] = fmt.Errorf("Error writing VMX: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCleanVMX) Cleanup(map[string]interface{}) {}

func (stepCleanVMX) readVMX(vmxPath string) (map[string]string, error) {
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
