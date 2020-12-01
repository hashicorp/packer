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
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

type vmxTemplateData struct {
	Name    string
	GuestOS string
	ISOPath string
	Version string

	CpuCount   string
	MemorySize string

	DiskName string
	vmwcommon.DiskAndCDConfigData

	Network_Type    string
	Network_Device  string
	Network_Adapter string
	Network_Name    string

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
//   ui     packersdk.Ui
//
// Produces:
//   vmx_path string - The path to the VMX file.
type stepCreateVMX struct {
	tempDir string
}

/* regular steps */
func (s *stepCreateVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packersdk.Ui)

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

	ictx := config.ctx

	// Mount extra vmdks we created earlier.
	if len(config.AdditionalDiskSize) > 0 {
		incrementer := 1
		for i := range config.AdditionalDiskSize {
			// slot 7 is special and reserved, so we need to skip that index.
			if i+1 == 7 {
				incrementer = 2
			}
			ictx.Data = &additionalDiskTemplateData{
				DiskNumber: i + incrementer,
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

			diskContents, err := interpolate.Render(diskTemplate, &ictx)
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

		Network_Adapter: "e1000",

		Sound_Present: map[bool]string{true: "TRUE", false: "FALSE"}[bool(config.HWConfig.Sound)],
		Usb_Present:   map[bool]string{true: "TRUE", false: "FALSE"}[bool(config.HWConfig.USB)],

		Serial_Present:   "FALSE",
		Parallel_Present: "FALSE",
	}

	templateData.DiskAndCDConfigData = vmwcommon.DefaultDiskAndCDROMTypes(config.DiskAdapterType, config.CdromAdapterType)

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
	if config.HWConfig.NetworkName != "" {
		templateData.Network_Name = config.HWConfig.NetworkName
	}
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

	ictx.Data = &templateData

	/// render the .vmx template
	vmxContents, err := interpolate.Render(vmxTemplate, &ictx)
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
{{if .Network_Name }}ethernet0.networkName = "{{ .Network_Name }}"{{end}}

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
