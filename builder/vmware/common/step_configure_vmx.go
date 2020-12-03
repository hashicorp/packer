package common

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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
	CustomData       map[string]string
	DisplayName      string
	SkipFloppy       bool
	VMName           string
	DiskAdapterType  string
	CDROMAdapterType string
}

func (s *StepConfigureVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	log.Printf("Configuring VMX...\n")

	var err error
	ui := state.Get("ui").(packersdk.Ui)

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
		// Grab list of temporary builder devices so we can append the floppy to it
		tmpBuildDevices := state.Get("temporaryDevices").([]string)

		// Set a floppy disk if we have one
		if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
			log.Println("Floppy path present, setting in VMX")
			vmxData["floppy0.present"] = "TRUE"
			vmxData["floppy0.filetype"] = "file"
			vmxData["floppy0.filename"] = floppyPathRaw.(string)

			// Add it to our list of build devices to later remove
			tmpBuildDevices = append(tmpBuildDevices, "floppy0")
		}

		// Build the list back in our statebag
		state.Put("temporaryDevices", tmpBuildDevices)
	}

	// Add our custom CD, if it exists
	if cdPath, ok := state.GetOk("cd_path"); ok {
		if cdPath != "" {
			diskAndCDConfigData := DefaultDiskAndCDROMTypes(s.DiskAdapterType, s.CDROMAdapterType)
			cdromPrefix := diskAndCDConfigData.CDROMType + "1:" + diskAndCDConfigData.CDROMType_PrimarySecondary
			vmxData[cdromPrefix+".present"] = "TRUE"
			vmxData[cdromPrefix+".fileName"] = cdPath.(string)
			vmxData[cdromPrefix+".deviceType"] = "cdrom-image"
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

	// Set the extendedConfigFile setting for the .vmxf filename to the VMName
	// if displayName is not set. This is needed so that when VMWare creates
	// the .vmxf file it matches the displayName if it is set. When just using
	// the sisplayName if it was empty VMWare would make a file named ".vmxf".
	// The ".vmxf" file would not get deleted when the VM got deleted.
	if s.DisplayName != "" {
		vmxData["extendedconfigfile"] = fmt.Sprintf("%s.vmxf", s.DisplayName)
	} else {
		vmxData["extendedconfigfile"] = fmt.Sprintf("%s.vmxf", s.VMName)
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

type DiskAndCDConfigData struct {
	SCSI_Present         string
	SCSI_diskAdapterType string
	SATA_Present         string
	NVME_Present         string

	DiskType                   string
	CDROMType                  string
	CDROMType_PrimarySecondary string
	CDROM_PATH                 string
}

// DefaultDiskAndCDROMTypes takes the disk adapter type and cdrom adapter type from the config and converts them
// into template interpolation data for creating or configuring a vmx.
func DefaultDiskAndCDROMTypes(diskAdapterType string, cdromAdapterType string) DiskAndCDConfigData {
	diskData := DiskAndCDConfigData{
		SCSI_Present:         "FALSE",
		SCSI_diskAdapterType: "lsilogic",
		SATA_Present:         "FALSE",
		NVME_Present:         "FALSE",

		DiskType:                   "scsi",
		CDROMType:                  "ide",
		CDROMType_PrimarySecondary: "0",
	}
	/// Use the disk adapter type that the user specified to tweak the .vmx
	//  Also sync the cdrom adapter type according to what's common for that disk type.
	//  XXX: If the cdrom type is modified, make sure to update common/step_clean_vmx.go
	//       so that it will regex the correct cdrom device for removal.
	diskAdapterType = strings.ToLower(diskAdapterType)
	switch diskAdapterType {
	case "ide":
		diskData.DiskType = "ide"
		diskData.CDROMType = "ide"
		diskData.CDROMType_PrimarySecondary = "1"
	case "sata":
		diskData.SATA_Present = "TRUE"
		diskData.DiskType = "sata"
		diskData.CDROMType = "sata"
		diskData.CDROMType_PrimarySecondary = "1"
	case "nvme":
		diskData.NVME_Present = "TRUE"
		diskData.DiskType = "nvme"
		diskData.SATA_Present = "TRUE"
		diskData.CDROMType = "sata"
		diskData.CDROMType_PrimarySecondary = "0"
	case "scsi":
		diskAdapterType = "lsilogic"
		fallthrough
	default:
		diskData.SCSI_Present = "TRUE"
		diskData.SCSI_diskAdapterType = diskAdapterType // defaults to lsilogic
		diskData.DiskType = "scsi"
		diskData.CDROMType = "ide"
		diskData.CDROMType_PrimarySecondary = "0"
	}

	/// Handle the cdrom adapter type. If the disk adapter type and the
	//  cdrom adapter type are the same, then ensure that the cdrom is the
	//  secondary device on whatever bus the disk adapter is on.
	if cdromAdapterType == "" {
		cdromAdapterType = diskData.CDROMType
	} else if cdromAdapterType == diskAdapterType {
		diskData.CDROMType_PrimarySecondary = "1"
	} else {
		diskData.CDROMType_PrimarySecondary = "0"
	}

	switch cdromAdapterType {
	case "ide":
		diskData.CDROMType = "ide"
	case "sata":
		diskData.SATA_Present = "TRUE"
		diskData.CDROMType = "sata"
	case "scsi":
		diskData.SCSI_Present = "TRUE"
		diskData.CDROMType = "scsi"
	}
	return diskData
}

func (s *StepConfigureVMX) Cleanup(state multistep.StateBag) {
}
