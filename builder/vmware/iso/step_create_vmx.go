package iso

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type vmxTemplateData struct {
	Name     string
	GuestOS  string
	DiskName string
	ISOPath  string
	Version  string
}

type additionalDiskTemplateData struct {
	DiskNumber int
	DiskName   string
}

// This step creates the VMX file for the VM.
//
// Uses:
//   config *config
//   iso_path string
//   ui     packer.Ui
//
// Produces:
//   vmx_path string - The path to the VMX file.
type stepCreateVMX struct {
	tempDir string
}

func (s *stepCreateVMX) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Building and writing VMX file")

	vmxTemplate := DefaultVMXTemplate
	if config.VMXTemplatePath != "" {
		f, err := os.Open(config.VMXTemplatePath)
		if err != nil {
			err := fmt.Errorf("Error reading VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		defer f.Close()

		rawBytes, err := ioutil.ReadAll(f)
		if err != nil {
			err := fmt.Errorf("Error reading VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		vmxTemplate = string(rawBytes)
	}

	ctx := config.ctx

	if len(config.AdditionalDiskSize) > 0 {
		for i, _ := range config.AdditionalDiskSize {
			ctx.Data = &additionalDiskTemplateData{
				DiskNumber: i + 1,
				DiskName:   config.DiskName,
			}

			diskTemplate := DefaultAdditionalDiskTemplate
			if config.VMXDiskTemplatePath != "" {
				f, err := os.Open(config.VMXDiskTemplatePath)
				if err != nil {
					err := fmt.Errorf("Error reading VMX disk template: %s", err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				defer f.Close()

				rawBytes, err := ioutil.ReadAll(f)
				if err != nil {
					err := fmt.Errorf("Error reading VMX disk template: %s", err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}

				diskTemplate = string(rawBytes)
			}

			diskContents, err := interpolate.Render(diskTemplate, &ctx)
			if err != nil {
				err := fmt.Errorf("Error preparing VMX template for additional disk: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			vmxTemplate += diskContents
		}
	}

	ctx.Data = &vmxTemplateData{
		Name:     config.VMName,
		GuestOS:  config.GuestOSType,
		DiskName: config.DiskName,
		Version:  config.Version,
		ISOPath:  isoPath,
	}

	vmxContents, err := interpolate.Render(vmxTemplate, &ctx)
	if err != nil {
		err := fmt.Errorf("Error procesing VMX template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vmxDir := config.OutputDir
	if config.RemoteType != "" {
		// For remote builds, we just put the VMX in a temporary
		// directory since it just gets uploaded anyways.
		vmxDir, err = ioutil.TempDir("", "packer-vmx")
		if err != nil {
			err := fmt.Errorf("Error preparing VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the tempDir so we clean it up
		s.tempDir = vmxDir
	}

	vmxPath := filepath.Join(vmxDir, config.VMName+".vmx")
	if err := vmwcommon.WriteVMX(vmxPath, vmwcommon.ParseVMX(vmxContents)); err != nil {
		err := fmt.Errorf("Error creating VMX file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vmx_path", vmxPath)

	return multistep.ActionContinue
}

func (s *stepCreateVMX) Cleanup(multistep.StateBag) {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// This is the default VMX template used if no other template is given.
// This is hardcoded here. If you wish to use a custom template please
// do so by specifying in the builder configuration.
const DefaultVMXTemplate = `
.encoding = "UTF-8"
bios.bootOrder = "hdd,CDROM"
checkpoint.vmState = ""
cleanShutdown = "TRUE"
config.version = "8"
displayName = "{{ .Name }}"
ehci.pciSlotNumber = "34"
ehci.present = "TRUE"
ethernet0.addressType = "generated"
ethernet0.bsdName = "en0"
ethernet0.connectionType = "nat"
ethernet0.displayName = "Ethernet"
ethernet0.linkStatePropagation.enable = "FALSE"
ethernet0.pciSlotNumber = "33"
ethernet0.present = "TRUE"
ethernet0.virtualDev = "e1000"
ethernet0.wakeOnPcktRcv = "FALSE"
extendedConfigFile = "{{ .Name }}.vmxf"
floppy0.present = "FALSE"
guestOS = "{{ .GuestOS }}"
gui.fullScreenAtPowerOn = "FALSE"
gui.viewModeAtPowerOn = "windowed"
hgfs.linkRootShare = "TRUE"
hgfs.mapRootShare = "TRUE"
ide1:0.present = "TRUE"
ide1:0.fileName = "{{ .ISOPath }}"
ide1:0.deviceType = "cdrom-image"
isolation.tools.hgfs.disable = "FALSE"
memsize = "512"
nvram = "{{ .Name }}.nvram"
pciBridge0.pciSlotNumber = "17"
pciBridge0.present = "TRUE"
pciBridge4.functions = "8"
pciBridge4.pciSlotNumber = "21"
pciBridge4.present = "TRUE"
pciBridge4.virtualDev = "pcieRootPort"
pciBridge5.functions = "8"
pciBridge5.pciSlotNumber = "22"
pciBridge5.present = "TRUE"
pciBridge5.virtualDev = "pcieRootPort"
pciBridge6.functions = "8"
pciBridge6.pciSlotNumber = "23"
pciBridge6.present = "TRUE"
pciBridge6.virtualDev = "pcieRootPort"
pciBridge7.functions = "8"
pciBridge7.pciSlotNumber = "24"
pciBridge7.present = "TRUE"
pciBridge7.virtualDev = "pcieRootPort"
powerType.powerOff = "soft"
powerType.powerOn = "soft"
powerType.reset = "soft"
powerType.suspend = "soft"
proxyApps.publishToHost = "FALSE"
replay.filename = ""
replay.supported = "FALSE"
scsi0.pciSlotNumber = "16"
scsi0.present = "TRUE"
scsi0.virtualDev = "lsilogic"
scsi0:0.fileName = "{{ .DiskName }}.vmdk"
scsi0:0.present = "TRUE"
scsi0:0.redo = ""
sound.startConnected = "FALSE"
tools.syncTime = "TRUE"
tools.upgrade.policy = "upgradeAtPowerCycle"
usb.pciSlotNumber = "32"
usb.present = "FALSE"
virtualHW.productCompatibility = "hosted"
virtualHW.version = "{{ .Version }}"
vmci0.id = "1861462627"
vmci0.pciSlotNumber = "35"
vmci0.present = "TRUE"
vmotion.checkpointFBSize = "65536000"
`

const DefaultAdditionalDiskTemplate = `
scsi0:{{ .DiskNumber }}.fileName = "{{ .DiskName}}-{{ .DiskNumber }}.vmdk"
scsi0:{{ .DiskNumber }}.present = "TRUE"
scsi0:{{ .DiskNumber }}.redo = ""
`
