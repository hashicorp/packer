package iso

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type vmxTemplateData struct {
	Name     string
	GuestOS  string
	DiskName string
	ISOPath  string
	Version  string

	Network  string
	Sound_Present string
	Usb_Present string

	Serial_Present string
	Serial_Type string
	Serial_Endpoint string
	Serial_Host string
	Serial_Yield string
	Serial_Filename string

	Parallel_Present string
	Parallel_Bidirectional string
	Parallel_Filename string
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

/* serial conversions */
type serialConfigPipe struct {
	filename string
	endpoint string
	host string
	yield string
}

type serialConfigFile struct {
	filename string
}

type serialConfigDevice struct {
	devicename string
}

type serialUnion struct {
	serialType interface{}
	pipe *serialConfigPipe
	file *serialConfigFile
	device *serialConfigDevice
}

func unformat_serial(config string) (*serialUnion,error) {
	comptype := strings.SplitN(config, ":", 2)
	if len(comptype) < 1 {
		return nil,fmt.Errorf("Unexpected format for serial port: %s", config)
	}
	switch strings.ToUpper(comptype[0]) {
		case "PIPE":
			comp := strings.Split(comptype[1], ",")
			if len(comp) < 3 || len(comp) > 4 {
				return nil,fmt.Errorf("Unexpected format for serial port : pipe : %s", config)
			}
			if res := strings.ToLower(comp[1]); res != "client" && res != "server" {
				return nil,fmt.Errorf("Unexpected format for serial port : pipe : endpoint : %s : %s", res, config)
			}
			if res := strings.ToLower(comp[2]); res != "app" && res != "vm" {
				return nil,fmt.Errorf("Unexpected format for serial port : pipe : host : %s : %s", res, config)
			}
			res := &serialConfigPipe{
				filename : comp[0],
				endpoint : comp[1],
				host : map[string]string{"app":"TRUE","vm":"FALSE"}[strings.ToLower(comp[2])],
				yield : "FALSE",
			}
			if len(comp) == 4 {
				res.yield = strings.ToUpper(comp[3])
			}
			if res.yield != "TRUE" && res.yield != "FALSE" {
				return nil,fmt.Errorf("Unexpected format for serial port : pipe : yield : %s : %s", res.yield, config)
			}
			return &serialUnion{serialType:res, pipe:res},nil

		case "FILE":
			res := &serialConfigFile{ filename : comptype[1] }
			return &serialUnion{serialType:res, file:res},nil

		case "DEVICE":
			res := new(serialConfigDevice)
			res.devicename = map[bool]string{true:strings.ToUpper(comptype[1]), false:"COM1"}[len(comptype[1]) > 0]
			return &serialUnion{serialType:res, device:res},nil

		default:
			return nil,fmt.Errorf("Unknown serial type : %s : %s", strings.ToUpper(comptype[0]), config)
	}
}

/* parallel port */
type parallelUnion struct {
	parallelType interface{}
	file *parallelPortFile
	device *parallelPortDevice
}
type parallelPortFile struct {
	filename string
}
type parallelPortDevice struct {
	bidirectional string
	devicename string
}

func unformat_parallel(config string) (*parallelUnion,error) {
	comptype := strings.SplitN(config, ":", 2)
	if len(comptype) < 1 {
		return nil,fmt.Errorf("Unexpected format for parallel port: %s", config)
	}
	switch strings.ToUpper(comptype[0]) {
		case "FILE":
			res := &parallelPortFile{ filename: comptype[1] }
			return &parallelUnion{ parallelType:res, file: res},nil
		case "DEVICE":
			comp := strings.Split(comptype[1], ",")
			if len(comp) < 1 || len(comp) > 2 {
				return nil,fmt.Errorf("Unexpected format for parallel port: %s", config)
			}
			res := new(parallelPortDevice)
			res.bidirectional = "FALSE"
			res.devicename = strings.ToUpper(comp[0])
			if len(comp) > 1 {
				switch strings.ToUpper(comp[1]) {
					case "BI":
						res.bidirectional = "TRUE"
					case "UNI":
						res.bidirectional = "FALSE"
					default:
						return nil,fmt.Errorf("Unknown parallel port direction : %s : %s", strings.ToUpper(comp[0]), config)
				}
			}
			return &parallelUnion{ parallelType:res, device: res},nil
	}
	return nil,fmt.Errorf("Unexpected format for parallel port: %s", config)
}

/* regular steps */
func (s *stepCreateVMX) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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
		for i := range config.AdditionalDiskSize {
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

	templateData := vmxTemplateData{
		Name:     config.VMName,
		GuestOS:  config.GuestOSType,
		DiskName: config.DiskName,
		Version:  config.Version,
		ISOPath:  isoPath,

		Network:  		config.Network,
		Sound_Present:	map[bool]string{true:"TRUE",false:"FALSE"}[bool(config.Sound)],
		Usb_Present:	map[bool]string{true:"TRUE",false:"FALSE"}[bool(config.USB)],

		Serial_Present:		"FALSE",
		Parallel_Present:	"FALSE",
	}

	// store the network so that we can later figure out what ip address to bind to
	state.Put("vmnetwork", config.Network)

	// check if serial port has been configured
	if config.Serial != "" {
		serial,err := unformat_serial(config.Serial)
		if err != nil {
			err := fmt.Errorf("Error procesing VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		templateData.Serial_Present = "TRUE"
		templateData.Serial_Filename = ""
		templateData.Serial_Yield = ""
		templateData.Serial_Endpoint = ""
		templateData.Serial_Host = ""

		switch serial.serialType.(type) {
			case *serialConfigPipe:
				templateData.Serial_Type = "pipe"
				templateData.Serial_Endpoint = serial.pipe.endpoint
				templateData.Serial_Host = serial.pipe.host
				templateData.Serial_Yield = serial.pipe.yield
				templateData.Serial_Filename = filepath.FromSlash(serial.pipe.filename)
			case *serialConfigFile:
				templateData.Serial_Type = "file"
				templateData.Serial_Filename = filepath.FromSlash(serial.file.filename)
			case *serialConfigDevice:
				templateData.Serial_Type = "device"
				templateData.Serial_Filename = filepath.FromSlash(serial.device.devicename)
			default:
				err := fmt.Errorf("Error procesing VMX template: %v", serial)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
		}
	}

	// check if parallel port has been configured
	if config.Parallel != "" {
		parallel,err := unformat_parallel(config.Parallel)
		if err != nil {
			err := fmt.Errorf("Error procesing VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		switch parallel.parallelType.(type) {
			case *parallelPortFile:
				templateData.Parallel_Present = "TRUE"
				templateData.Parallel_Filename = filepath.FromSlash(parallel.file.filename)
			case *parallelPortDevice:
				templateData.Parallel_Present = "TRUE"
				templateData.Parallel_Bidirectional = parallel.device.bidirectional
				templateData.Parallel_Filename = filepath.FromSlash(parallel.device.devicename)
			default:
				err := fmt.Errorf("Error procesing VMX template: %v", parallel)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
		}
	}

	ctx.Data = &templateData

	// render the .vmx template
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
ethernet0.connectionType = "{{ .Network }}"
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

// Sound
sound.startConnected = "{{ .Sound_Present }}"
sound.present = "{{ .Sound_Present }}"
sound.fileName = "-1"
sound.autodetect = "TRUE"

tools.syncTime = "TRUE"
tools.upgrade.policy = "upgradeAtPowerCycle"

// USB
usb.pciSlotNumber = "32"
usb.present = "{{ .Usb_Present }}"
usb_xhci.present = "TRUE"

// Serial
serial0.present = "{{ .Serial_Present }}"
serial0.startConnected = "{{ .Serial_Present }}"
serial0.fileName = "{{ .Serial_Filename }}"
serial0.autodetect = "TRUE"
serial0.fileType = "{{ .Serial_Type }}"
serial0.yieldOnMsrRead = "{{ .Serial_Yield }}"
serial0.pipe.endPoint = "{{ .Serial_Endpoint }}"
serial0.tryNoRxLoss = "{{ .Serial_Host }}"

// Parallel
parallel0.present = "{{ .Parallel_Present }}"
parallel0.startConnected = "{{ .Parallel_Present }}"
parallel0.fileName = "{{ .Parallel_Filename }}"
parallel0.autodetect = "TRUE"
parallel0.bidirectional = "{{ .Parallel_Bidirectional }}"

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
