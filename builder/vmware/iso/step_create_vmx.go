package iso

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
	"github.com/hashicorp/packer/template/interpolate"
)

type vmxTemplateData struct {
	Name    string
	GuestOS string
	ISOPath string
	Version string

	CpuCount   string
	MemorySize string

	SCSI_Present         string
	SCSI_diskAdapterType string
	SATA_Present         string
	NVME_Present         string

	DiskName                   string
	DiskType                   string
	CDROMType                  string
	CDROMType_PrimarySecondary string

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

		DiskType:                   "scsi",
		CDROMType:                  "ide",
		CDROMType_PrimarySecondary: "0",

		Network_Adapter: "e1000",

		Sound_Present: map[bool]string{true: "TRUE", false: "FALSE"}[bool(config.HWConfig.Sound)],
		Usb_Present:   map[bool]string{true: "TRUE", false: "FALSE"}[bool(config.HWConfig.USB)],

		Serial_Present:   "FALSE",
		Parallel_Present: "FALSE",
	}

	/// Use the disk adapter type that the user specified to tweak the .vmx
	//  Also sync the cdrom adapter type according to what's common for that disk type.
	//  XXX: If the cdrom type is modified, make sure to update common/step_clean_vmx.go
	//       so that it will regex the correct cdrom device for removal.
	diskAdapterType := strings.ToLower(config.DiskAdapterType)
	switch diskAdapterType {
	case "ide":
		templateData.DiskType = "ide"
		templateData.CDROMType = "ide"
		templateData.CDROMType_PrimarySecondary = "1"
	case "sata":
		templateData.SATA_Present = "TRUE"
		templateData.DiskType = "sata"
		templateData.CDROMType = "sata"
		templateData.CDROMType_PrimarySecondary = "1"
	case "nvme":
		templateData.NVME_Present = "TRUE"
		templateData.DiskType = "nvme"
		templateData.SATA_Present = "TRUE"
		templateData.CDROMType = "sata"
		templateData.CDROMType_PrimarySecondary = "0"
	case "scsi":
		diskAdapterType = "lsilogic"
		fallthrough
	default:
		templateData.SCSI_Present = "TRUE"
		templateData.SCSI_diskAdapterType = diskAdapterType
		templateData.DiskType = "scsi"
		templateData.CDROMType = "ide"
		templateData.CDROMType_PrimarySecondary = "0"
	}

	/// Handle the cdrom adapter type. If the disk adapter type and the
	//  cdrom adapter type are the same, then ensure that the cdrom is the
	//  secondary device on whatever bus the disk adapter is on.
	cdromAdapterType := strings.ToLower(config.CdromAdapterType)
	if cdromAdapterType == "" {
		cdromAdapterType = templateData.CDROMType
	} else if cdromAdapterType == diskAdapterType {
		templateData.CDROMType_PrimarySecondary = "1"
	} else {
		templateData.CDROMType_PrimarySecondary = "0"
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
		err := fmt.Errorf("Error processing VMX template: %s", cdromAdapterType)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	/// Now that we figured out the CDROM device to add, store it
	/// to the list of temporary build devices in our statebag
	tmpBuildDevices := state.Get("temporaryDevices").([]string)
	tmpCdromDevice := fmt.Sprintf("%s0:%s", templateData.CDROMType, templateData.CDROMType_PrimarySecondary)
	tmpBuildDevices = append(tmpBuildDevices, tmpCdromDevice)
	state.Put("temporaryDevices", tmpBuildDevices)

	/// Assign the network adapter type into the template if one was specified.
	network_adapter := strings.ToLower(config.HWConfig.NetworkAdapterType)
	if network_adapter != "" {
		templateData.Network_Adapter = network_adapter
	}

	/// Check the network type that the user specified
	network := config.HWConfig.Network
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
		devices, err := netmap.NameIntoDevices(network)

		if err == nil && len(devices) > 0 {
			// If multiple devices exist, for example for network "nat", VMware chooses
			// the actual device. Only type "custom" allows the exact choice of a
			// specific virtual network (see below). We allow VMware to choose the device
			// and for device-specific operations like GuestIP, try to go over all
			// devices that match a name (e.g. "nat").
			// https://pubs.vmware.com/workstation-9/index.jsp?topic=%2Fcom.vmware.ws.using.doc%2FGUID-3B504F2F-7A0B-415F-AE01-62363A95D052.html
			templateData.Network_Type = network
			templateData.Network_Device = ""
		} else {
			// otherwise, we were unable to find the type, so assume it's a custom device
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
	if !config.HWConfig.HasSerial() {
		templateData.Serial_Present = "FALSE"
	} else {
		// FIXME
		serial, err := config.HWConfig.ReadSerial()
		if err != nil {
			err := fmt.Errorf("Error processing VMX template: %s", err)
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

		// Set the number of cpus if it was specified
		if config.HWConfig.CpuCount > 0 {
			templateData.CpuCount = strconv.Itoa(config.HWConfig.CpuCount)
		}

		// Apply the memory size that was specified
		if config.HWConfig.MemorySize > 0 {
			templateData.MemorySize = strconv.Itoa(config.HWConfig.MemorySize)
		} else {
			templateData.MemorySize = "512"
		}

		switch serial.Union.(type) {
		case *vmwcommon.SerialConfigPipe:
			templateData.Serial_Type = "pipe"
			templateData.Serial_Endpoint = serial.Pipe.Endpoint
			templateData.Serial_Host = serial.Pipe.Host
			templateData.Serial_Yield = serial.Pipe.Yield
			templateData.Serial_Filename = filepath.FromSlash(serial.Pipe.Filename)
		case *vmwcommon.SerialConfigFile:
			templateData.Serial_Type = "file"
			templateData.Serial_Filename = filepath.FromSlash(serial.File.Filename)
		case *vmwcommon.SerialConfigDevice:
			templateData.Serial_Type = "device"
			templateData.Serial_Filename = filepath.FromSlash(serial.Device.Devicename)
		case *vmwcommon.SerialConfigAuto:
			templateData.Serial_Type = "device"
			templateData.Serial_Filename = filepath.FromSlash(serial.Auto.Devicename)
			templateData.Serial_Yield = serial.Auto.Yield
			templateData.Serial_Auto = "TRUE"
		case nil:
			templateData.Serial_Present = "FALSE"
			break

		default:
			err := fmt.Errorf("Error processing VMX template: %v", serial)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	/// check if parallel port has been configured
	if !config.HWConfig.HasParallel() {
		templateData.Parallel_Present = "FALSE"
	} else {
		// FIXME
		parallel, err := config.HWConfig.ReadParallel()
		if err != nil {
			err := fmt.Errorf("Error processing VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		templateData.Parallel_Auto = "FALSE"
		switch parallel.Union.(type) {
		case *vmwcommon.ParallelPortFile:
			templateData.Parallel_Present = "TRUE"
			templateData.Parallel_Filename = filepath.FromSlash(parallel.File.Filename)
		case *vmwcommon.ParallelPortDevice:
			templateData.Parallel_Present = "TRUE"
			templateData.Parallel_Bidirectional = parallel.Device.Bidirectional
			templateData.Parallel_Filename = filepath.FromSlash(parallel.Device.Devicename)
		case *vmwcommon.ParallelPortAuto:
			templateData.Parallel_Present = "TRUE"
			templateData.Parallel_Auto = "TRUE"
			templateData.Parallel_Bidirectional = parallel.Auto.Bidirectional
		case nil:
			templateData.Parallel_Present = "FALSE"
			break

		default:
			err := fmt.Errorf("Error processing VMX template: %v", parallel)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ctx.Data = &templateData

	/// render the .vmx template
	vmxContents, err := interpolate.Render(vmxTemplate, &ctx)
	if err != nil {
		err := fmt.Errorf("Error processing VMX template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vmxDir := config.OutputDir
	if config.RemoteType != "" {
		// For remote builds, we just put the VMX in a temporary
		// directory since it just gets uploaded anyways.
		vmxDir, err = tmp.Dir("vmw-iso")
		if err != nil {
			err := fmt.Errorf("Error preparing VMX template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the tempDir so we clean it up
		s.tempDir = vmxDir
	}

	/// Now to handle options that will modify the template without using "vmxTemplateData"
	vmxData := vmwcommon.ParseVMX(vmxContents)

	// If no cpus were specified, then remove the entry to use the default
	if vmxData["numvcpus"] == "" {
		delete(vmxData, "numvcpus")
	}

	// If some number of cores were specified, then update "cpuid.coresPerSocket" with the requested value
	if config.HWConfig.CoreCount > 0 {
		vmxData["cpuid.corespersocket"] = strconv.Itoa(config.HWConfig.CoreCount)
	}

	/// Write the vmxData to the vmxPath
	vmxPath := filepath.Join(vmxDir, config.VMName+".vmx")
	if err := vmwcommon.WriteVMX(vmxPath, vmxData); err != nil {
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

displayName = "{{ .Name }}"

// Hardware
numvcpus = "{{ .CpuCount }}"
memsize = "{{ .MemorySize }}"

config.version = "8"
virtualHW.productCompatibility = "hosted"
virtualHW.version = "{{ .Version }}"

// Bootup
nvram = "{{ .Name }}.nvram"

floppy0.present = "FALSE"
bios.bootOrder = "hdd,cdrom"

// Configuration
extendedConfigFile = "{{ .Name }}.vmxf"
gui.fullScreenAtPowerOn = "FALSE"
gui.viewModeAtPowerOn = "windowed"
hgfs.linkRootShare = "TRUE"
hgfs.mapRootShare = "TRUE"
isolation.tools.hgfs.disable = "FALSE"
proxyApps.publishToHost = "FALSE"
replay.filename = ""
replay.supported = "FALSE"

checkpoint.vmState = ""
vmotion.checkpointFBSize = "65536000"

// Power control
cleanShutdown = "TRUE"
powerType.powerOff = "soft"
powerType.powerOn = "soft"
powerType.reset = "soft"
powerType.suspend = "soft"

// Tools
guestOS = "{{ .GuestOS }}"
tools.syncTime = "TRUE"
tools.upgrade.policy = "upgradeAtPowerCycle"

// Bus
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

ehci.present = "TRUE"
ehci.pciSlotNumber = "34"

vmci0.present = "TRUE"
vmci0.id = "1861462627"
vmci0.pciSlotNumber = "35"

// Network Adapter
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

// Hard disks
scsi0.present = "{{ .SCSI_Present }}"
scsi0.virtualDev = "{{ .SCSI_diskAdapterType }}"
scsi0.pciSlotNumber = "16"
scsi0:0.redo = ""
sata0.present = "{{ .SATA_Present }}"
nvme0.present = "{{ .NVME_Present }}"

{{ .DiskType }}0:0.present = "TRUE"
{{ .DiskType }}0:0.fileName = "{{ .DiskName }}.vmdk"

{{ .CDROMType }}0:{{ .CDROMType_PrimarySecondary }}.present = "TRUE"
{{ .CDROMType }}0:{{ .CDROMType_PrimarySecondary }}.fileName = "{{ .ISOPath }}"
{{ .CDROMType }}0:{{ .CDROMType_PrimarySecondary }}.deviceType = "cdrom-image"

// Sound
sound.startConnected = "{{ .Sound_Present }}"
sound.present = "{{ .Sound_Present }}"
sound.fileName = "-1"
sound.autodetect = "TRUE"

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
`

const DefaultAdditionalDiskTemplate = `
scsi0:{{ .DiskNumber }}.fileName = "{{ .DiskName}}-{{ .DiskNumber }}.vmdk"
scsi0:{{ .DiskNumber }}.present = "TRUE"
scsi0:{{ .DiskNumber }}.redo = ""
`
