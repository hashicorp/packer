package iso

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type vmxTemplateData struct {
	Name    string
	GuestOS string
	ISOPath string
	Version string

	SCSI_Present         string
	SCSI_diskAdapterType string
	SATA_Present         string
	NVME_Present         string

	DiskName              string
	DiskType              string
	CDROMType             string
	CDROMType_MasterSlave string

	Network_Type    string
	Network_Device  string
	Network_Adapter string

	Sound_Present string
	Usb_Present   string

	Serial_Present  string
	Serial_Type     string
	Serial_Endpoint string
	Serial_Host     string
	Serial_Yield    string
	Serial_Filename string
	Serial_Auto     string

	Parallel_Present       string
	Parallel_Bidirectional string
	Parallel_Filename      string
	Parallel_Auto          string
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
	host     string
	yield    string
}

type serialConfigFile struct {
	filename string
	yield    string
}

type serialConfigDevice struct {
	devicename string
	yield      string
}

type serialConfigAuto struct {
	devicename string
	yield      string
}

type serialUnion struct {
	serialType interface{}
	pipe       *serialConfigPipe
	file       *serialConfigFile
	device     *serialConfigDevice
	auto       *serialConfigAuto
}

func unformat_serial(config string) (*serialUnion, error) {
	var defaultSerialPort string
	if runtime.GOOS == "windows" {
		defaultSerialPort = "COM1"
	} else {
		defaultSerialPort = "/dev/ttyS0"
	}

	input := strings.SplitN(config, ":", 2)
	if len(input) < 1 {
		return nil, fmt.Errorf("Unexpected format for serial port: %s", config)
	}

	var formatType, formatOptions string
	formatType = input[0]
	if len(input) == 2 {
		formatOptions = input[1]
	} else {
		formatOptions = ""
	}

	switch strings.ToUpper(formatType) {
	case "PIPE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) < 3 || len(comp) > 4 {
			return nil, fmt.Errorf("Unexpected format for serial port : pipe : %s", config)
		}
		if res := strings.ToLower(comp[1]); res != "client" && res != "server" {
			return nil, fmt.Errorf("Unexpected format for serial port : pipe : endpoint : %s : %s", res, config)
		}
		if res := strings.ToLower(comp[2]); res != "app" && res != "vm" {
			return nil, fmt.Errorf("Unexpected format for serial port : pipe : host : %s : %s", res, config)
		}
		res := &serialConfigPipe{
			filename: comp[0],
			endpoint: comp[1],
			host:     map[string]string{"app": "TRUE", "vm": "FALSE"}[strings.ToLower(comp[2])],
			yield:    "FALSE",
		}
		if len(comp) == 4 {
			res.yield = strings.ToUpper(comp[3])
		}
		if res.yield != "TRUE" && res.yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for serial port : pipe : yield : %s : %s", res.yield, config)
		}
		return &serialUnion{serialType: res, pipe: res}, nil

	case "FILE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) > 2 {
			return nil, fmt.Errorf("Unexpected format for serial port : file : %s", config)
		}

		res := &serialConfigFile{yield: "FALSE"}

		res.filename = filepath.FromSlash(comp[0])

		res.yield = map[bool]string{true: strings.ToUpper(comp[0]), false: "FALSE"}[len(comp) > 1]
		if res.yield != "TRUE" && res.yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for serial port : file : yield : %s : %s", res.yield, config)
		}

		return &serialUnion{serialType: res, file: res}, nil

	case "DEVICE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) > 2 {
			return nil, fmt.Errorf("Unexpected format for serial port : device : %s", config)
		}

		res := new(serialConfigDevice)

		if len(comp) == 2 {
			res.devicename = map[bool]string{true: filepath.FromSlash(comp[0]), false: defaultSerialPort}[len(comp[0]) > 0]
			res.yield = strings.ToUpper(comp[1])
		} else if len(comp) == 1 {
			res.devicename = map[bool]string{true: filepath.FromSlash(comp[0]), false: defaultSerialPort}[len(comp[0]) > 0]
			res.yield = "FALSE"
		} else if len(comp) == 0 {
			res.devicename = defaultSerialPort
			res.yield = "FALSE"
		}

		if res.yield != "TRUE" && res.yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for serial port : device : yield : %s : %s", res.yield, config)
		}

		return &serialUnion{serialType: res, device: res}, nil

	case "AUTO":
		res := new(serialConfigAuto)
		res.devicename = defaultSerialPort

		if len(formatOptions) > 0 {
			res.yield = strings.ToUpper(formatOptions)
		} else {
			res.yield = "FALSE"
		}

		if res.yield != "TRUE" && res.yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for serial port : auto : yield : %s : %s", res.yield, config)
		}

		return &serialUnion{serialType: res, auto: res}, nil

	case "NONE":
		return &serialUnion{serialType: nil}, nil

	default:
		return nil, fmt.Errorf("Unknown serial type : %s : %s", strings.ToUpper(formatType), config)
	}
}

/* parallel port */
type parallelUnion struct {
	parallelType interface{}
	file         *parallelPortFile
	device       *parallelPortDevice
	auto         *parallelPortAuto
}
type parallelPortFile struct {
	filename string
}
type parallelPortDevice struct {
	bidirectional string
	devicename    string
}
type parallelPortAuto struct {
	bidirectional string
}

func unformat_parallel(config string) (*parallelUnion, error) {
	input := strings.SplitN(config, ":", 2)
	if len(input) < 1 {
		return nil, fmt.Errorf("Unexpected format for parallel port: %s", config)
	}

	var formatType, formatOptions string
	formatType = input[0]
	if len(input) == 2 {
		formatOptions = input[1]
	} else {
		formatOptions = ""
	}

	switch strings.ToUpper(formatType) {
	case "FILE":
		res := &parallelPortFile{filename: filepath.FromSlash(formatOptions)}
		return &parallelUnion{parallelType: res, file: res}, nil
	case "DEVICE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) < 1 || len(comp) > 2 {
			return nil, fmt.Errorf("Unexpected format for parallel port: %s", config)
		}
		res := new(parallelPortDevice)
		res.bidirectional = "FALSE"
		res.devicename = filepath.FromSlash(comp[0])
		if len(comp) > 1 {
			switch strings.ToUpper(comp[1]) {
			case "BI":
				res.bidirectional = "TRUE"
			case "UNI":
				res.bidirectional = "FALSE"
			default:
				return nil, fmt.Errorf("Unknown parallel port direction : %s : %s", strings.ToUpper(comp[0]), config)
			}
		}
		return &parallelUnion{parallelType: res, device: res}, nil

	case "AUTO":
		res := new(parallelPortAuto)
		switch strings.ToUpper(formatOptions) {
		case "":
			fallthrough
		case "UNI":
			res.bidirectional = "FALSE"
		case "BI":
			res.bidirectional = "TRUE"
		default:
			return nil, fmt.Errorf("Unknown parallel port direction : %s : %s", strings.ToUpper(formatOptions), config)
		}
		return &parallelUnion{parallelType: res, auto: res}, nil

	case "NONE":
		return &parallelUnion{parallelType: nil}, nil
	}

	return nil, fmt.Errorf("Unexpected format for parallel port: %s", config)
}

/* regular steps */
func (s *stepCreateVMX) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)

	// Convert the iso_path into a path relative to the .vmx file if possible
	if relativeIsoPath, err := filepath.Rel(config.VMXTemplatePath, filepath.FromSlash(isoPath)); err == nil {
		isoPath = relativeIsoPath
	}

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

		SCSI_Present:         "FALSE",
		SCSI_diskAdapterType: "lsilogic",
		SATA_Present:         "FALSE",
		NVME_Present:         "FALSE",

		DiskType:              "scsi",
		CDROMType:             "ide",
		CDROMType_MasterSlave: "0",

		Network_Adapter: "e1000",

		Sound_Present: map[bool]string{true: "TRUE", false: "FALSE"}[bool(config.Sound)],
		Usb_Present:   map[bool]string{true: "TRUE", false: "FALSE"}[bool(config.USB)],

		Serial_Present:   "FALSE",
		Parallel_Present: "FALSE",
	}

	/// Use the disk adapter type that the user specified to tweak the .vmx
	//  Also sync the cdrom adapter type according to what's common for that disk type.
	diskAdapterType := strings.ToLower(config.DiskAdapterType)
	switch diskAdapterType {
	case "ide":
		templateData.DiskType = "ide"
		templateData.CDROMType = "ide"
		templateData.CDROMType_MasterSlave = "1"
	case "sata":
		templateData.SATA_Present = "TRUE"
		templateData.DiskType = "sata"
		templateData.CDROMType = "sata"
		templateData.CDROMType_MasterSlave = "1"
	case "nvme":
		templateData.NVME_Present = "TRUE"
		templateData.DiskType = "nvme"
		templateData.SATA_Present = "TRUE"
		templateData.CDROMType = "sata"
		templateData.CDROMType_MasterSlave = "0"
	case "scsi":
		diskAdapterType = "lsilogic"
		fallthrough
	default:
		templateData.SCSI_Present = "TRUE"
		templateData.SCSI_diskAdapterType = diskAdapterType
		templateData.DiskType = "scsi"
		templateData.CDROMType = "ide"
		templateData.CDROMType_MasterSlave = "0"
	}

	/// Handle the cdrom adapter type. If the disk adapter type and the
	//  cdrom adapter type are the same, then ensure that the cdrom is the
	//  slave device on whatever bus the disk adapter is on.
	cdromAdapterType := strings.ToLower(config.CdromAdapterType)
	if cdromAdapterType == "" {
		cdromAdapterType = templateData.CDROMType
	} else if cdromAdapterType == diskAdapterType {
		templateData.CDROMType_MasterSlave = "1"
	} else {
		templateData.CDROMType_MasterSlave = "0"
	}

	switch cdromAdapterType {
	case "ide":
		templateData.CDROMType = "ide"
	case "sata":
		templateData.SATA_Present = "TRUE"
		templateData.CDROMType = "sata"
	case "scsi":
		templateData.SCSI_Present = "TRUE"
		templateData.CDROMType = "scsi"
	default:
		err := fmt.Errorf("Error procesing VMX template: %s", cdromAdapterType)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	/// Assign the network adapter type into the template if one was specified.
	network_adapter := strings.ToLower(config.NetworkAdapterType)
	if network_adapter != "" {
		templateData.Network_Adapter = network_adapter
	}

	/// Check the network type that the user specified
	network := config.Network
	driver := state.Get("driver").(vmwcommon.Driver).GetVmwareDriver()

	// check to see if the driver implements a network mapper for mapping
	// the network-type to its device-name.
	if driver.NetworkMapper != nil {

		// read network map configuration into a NetworkNameMapper.
		netmap, err := driver.NetworkMapper()
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// try and convert the specified network to a device.
		device, err := netmap.NameIntoDevice(network)

		if err == nil {
			// success. so we know that it's an actual network type inside netmap.conf
			templateData.Network_Type = network
			templateData.Network_Device = device
		} else {
			// otherwise, we were unable to find the type, so assume its a custom device.
			templateData.Network_Type = "custom"
			templateData.Network_Device = network
		}

		// if NetworkMapper is nil, then we're using something like ESX, so fall
		// back to the previous logic of using "nat" despite it not mattering to ESX.
	} else {
		templateData.Network_Type = "nat"
		templateData.Network_Device = network

		network = "nat"
	}

	// store the network so that we can later figure out what ip address to bind to
	state.Put("vmnetwork", network)

	/// check if serial port has been configured
	if config.Serial == "" {
		templateData.Serial_Present = "FALSE"
	} else {
		serial, err := unformat_serial(config.Serial)
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
		templateData.Serial_Auto = "FALSE"

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
		case *serialConfigAuto:
			templateData.Serial_Type = "device"
			templateData.Serial_Filename = filepath.FromSlash(serial.auto.devicename)
			templateData.Serial_Yield = serial.auto.yield
			templateData.Serial_Auto = "TRUE"
		case nil:
			templateData.Serial_Present = "FALSE"
			break

		default:
			err := fmt.Errorf("Error procesing VMX template: %v", serial)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	/// check if parallel port has been configured
	if config.Parallel == "" {
		templateData.Parallel_Present = "FALSE"
	} else {
		parallel, err := unformat_parallel(config.Parallel)
		if err != nil {
			err := fmt.Errorf("Error procesing VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		templateData.Parallel_Auto = "FALSE"
		switch parallel.parallelType.(type) {
		case *parallelPortFile:
			templateData.Parallel_Present = "TRUE"
			templateData.Parallel_Filename = filepath.FromSlash(parallel.file.filename)
		case *parallelPortDevice:
			templateData.Parallel_Present = "TRUE"
			templateData.Parallel_Bidirectional = parallel.device.bidirectional
			templateData.Parallel_Filename = filepath.FromSlash(parallel.device.devicename)
		case *parallelPortAuto:
			templateData.Parallel_Present = "TRUE"
			templateData.Parallel_Auto = "TRUE"
			templateData.Parallel_Bidirectional = parallel.auto.bidirectional
		case nil:
			templateData.Parallel_Present = "FALSE"
			break

		default:
			err := fmt.Errorf("Error procesing VMX template: %v", parallel)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ctx.Data = &templateData

	/// render the .vmx template
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
ethernet0.connectionType = "{{ .Network_Type }}"
ethernet0.vnet = "{{ .Network_Device }}"
ethernet0.displayName = "Ethernet"
ethernet0.linkStatePropagation.enable = "FALSE"
ethernet0.pciSlotNumber = "33"
ethernet0.present = "TRUE"
ethernet0.virtualDev = "{{ .Network_Adapter }}"
ethernet0.wakeOnPcktRcv = "FALSE"
extendedConfigFile = "{{ .Name }}.vmxf"
floppy0.present = "FALSE"
guestOS = "{{ .GuestOS }}"
gui.fullScreenAtPowerOn = "FALSE"
gui.viewModeAtPowerOn = "windowed"
hgfs.linkRootShare = "TRUE"
hgfs.mapRootShare = "TRUE"

scsi0.present = "{{ .SCSI_Present }}"
scsi0.virtualDev = "{{ .SCSI_diskAdapterType }}"
scsi0.pciSlotNumber = "16"
scsi0:0.redo = ""
sata0.present = "{{ .SATA_Present }}"
nvme0.present = "{{ .NVME_Present }}"

{{ .DiskType }}0:0.present = "TRUE"
{{ .DiskType }}0:0.fileName = "{{ .DiskName }}.vmdk"

{{ .CDROMType }}0:{{ .CDROMType_MasterSlave }}.present = "TRUE"
{{ .CDROMType }}0:{{ .CDROMType_MasterSlave }}.fileName = "{{ .ISOPath }}"
{{ .CDROMType }}0:{{ .CDROMType_MasterSlave }}.deviceType = "cdrom-image"

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

// Serial
serial0.present = "{{ .Serial_Present }}"
serial0.startConnected = "{{ .Serial_Present }}"
serial0.fileName = "{{ .Serial_Filename }}"
serial0.autodetect = "{{ .Serial_Auto }}"
serial0.fileType = "{{ .Serial_Type }}"
serial0.yieldOnMsrRead = "{{ .Serial_Yield }}"
serial0.pipe.endPoint = "{{ .Serial_Endpoint }}"
serial0.tryNoRxLoss = "{{ .Serial_Host }}"

// Parallel
parallel0.present = "{{ .Parallel_Present }}"
parallel0.startConnected = "{{ .Parallel_Present }}"
parallel0.fileName = "{{ .Parallel_Filename }}"
parallel0.autodetect = "{{ .Parallel_Auto }}"
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
