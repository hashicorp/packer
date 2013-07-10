package vmware

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"path/filepath"
	"text/template"
)

type vmxTemplateData struct {
	Name     string
	GuestOS  string
	DiskName string
	ISOPath  string
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
type stepCreateVMX struct{}

func (stepCreateVMX) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	isoPath := state["iso_path"].(string)
	ui := state["ui"].(packer.Ui)

	ui.Say("Building and writing VMX file")

	tplData := &vmxTemplateData{
		Name:     config.VMName,
		GuestOS:  config.GuestOSType,
		DiskName: config.DiskName,
		ISOPath:  isoPath,
	}

	var buf bytes.Buffer
	t := template.Must(template.New("vmx").Parse(DefaultVMXTemplate))
	t.Execute(&buf, tplData)

	vmxData := ParseVMX(buf.String())
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

	vmxPath := filepath.Join(config.OutputDir, config.VMName+".vmx")
	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error creating VMX file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["vmx_path"] = vmxPath

	return multistep.ActionContinue
}

func (stepCreateVMX) Cleanup(map[string]interface{}) {
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
virtualHW.version = "9"
vmci0.id = "1861462627"
vmci0.pciSlotNumber = "35"
vmci0.present = "TRUE"
vmotion.checkpointFBSize = "65536000"
`
