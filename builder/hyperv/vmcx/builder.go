package vmcx

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/multistep"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
	"github.com/mitchellh/packer/common"
	powershell "github.com/mitchellh/packer/common/powershell"
	"github.com/mitchellh/packer/common/powershell/hyperv"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const (
	DefaultRamSize                 = 1 * 1024  // 1GB
	MinRamSize                     = 32        // 32MB
	MaxRamSize                     = 32 * 1024 // 32GB
	MinNestedVirtualizationRamSize = 4 * 1024  // 4GB

	LowRam = 256 // 256MB

	DefaultUsername = ""
	DefaultPassword = ""
)

// Builder implements packer.Builder and builds the actual Hyperv
// images.
type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig         `mapstructure:",squash"`
	common.HTTPConfig           `mapstructure:",squash"`
	common.ISOConfig            `mapstructure:",squash"`
	hypervcommon.FloppyConfig   `mapstructure:",squash"`
	hypervcommon.OutputConfig   `mapstructure:",squash"`
	hypervcommon.SSHConfig      `mapstructure:",squash"`
	hypervcommon.RunConfig      `mapstructure:",squash"`
	hypervcommon.ShutdownConfig `mapstructure:",squash"`

	// The size, in megabytes, of the computer memory in the VM.
	// By default, this is 1024 (about 1 GB).
	RamSize uint `mapstructure:"ram_size"`
	// A list of files to place onto a floppy disk that is attached when the
	// VM is booted. This is most useful for unattended Windows installs,
	// which look for an Autounattend.xml file on removable media. By default,
	// no floppy will be attached. All files listed in this setting get
	// placed into the root directory of the floppy and the floppy is attached
	// as the first floppy device. Currently, no support exists for creating
	// sub-directories on the floppy. Wildcard characters (*, ?, and [])
	// are allowed. Directory names are also allowed, which will add all
	// the files found in the directory to the floppy.
	FloppyFiles []string `mapstructure:"floppy_files"`
	//
	SecondaryDvdImages []string `mapstructure:"secondary_iso_images"`

	// Should integration services iso be mounted
	GuestAdditionsMode string `mapstructure:"guest_additions_mode"`

	// The path to the integration services iso
	GuestAdditionsPath string `mapstructure:"guest_additions_path"`

	// This is the name of the virtual machine to clone from.
	CloneFromVMName string `mapstructure:"clone_from_vm_name"`

	// This is the name of the snapshot to clone from. A blank snapshot name will use the latest snapshot.
	CloneFromSnapshotName string `mapstructure:"clone_from_snapshot_name"`

	// This will clone all snapshots if true. It will clone latest snapshot if false.
	CloneAllSnapshots bool `mapstructure:"clone_all_snapshots"`

	// This is the name of the new virtual machine.
	// By default this is "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name"`

	BootCommand                    []string `mapstructure:"boot_command"`
	SwitchName                     string   `mapstructure:"switch_name"`
	SwitchVlanId                   string   `mapstructure:"switch_vlan_id"`
	VlanId                         string   `mapstructure:"vlan_id"`
	Cpu                            uint     `mapstructure:"cpu"`
	Generation                     uint     `mapstructure:"generation"`
	EnableMacSpoofing              bool     `mapstructure:"enable_mac_spoofing"`
	EnableDynamicMemory            bool     `mapstructure:"enable_dynamic_memory"`
	EnableSecureBoot               bool     `mapstructure:"enable_secure_boot"`
	EnableVirtualizationExtensions bool     `mapstructure:"enable_virtualization_extensions"`

	Communicator string `mapstructure:"communicator"`

	SkipCompaction bool `mapstructure:"skip_compaction"`

	ctx interpolate.Context
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)

	err = b.checkRamSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	log.Println(fmt.Sprintf("%s: %v", "VMName", b.config.VMName))

	if b.config.SwitchName == "" {
		b.config.SwitchName = b.detectSwitchName()
	}

	if b.config.Cpu < 1 {
		b.config.Cpu = 1
	}

	if b.config.Generation != 2 {
		b.config.Generation = 1
	}

	if b.config.Generation == 2 {
		if len(b.config.FloppyFiles) > 0 {
			err = errors.New("Generation 2 vms don't support floppy drives. Use ISO image instead.")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	log.Println(fmt.Sprintf("Using switch %s", b.config.SwitchName))
	log.Println(fmt.Sprintf("%s: %v", "SwitchName", b.config.SwitchName))

	// Errors
	if b.config.GuestAdditionsMode == "" {
		if b.config.GuestAdditionsPath != "" {
			b.config.GuestAdditionsMode = "attach"
		} else {
			b.config.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

			if _, err := os.Stat(b.config.GuestAdditionsPath); os.IsNotExist(err) {
				if err != nil {
					b.config.GuestAdditionsPath = ""
					b.config.GuestAdditionsMode = "none"
				} else {
					b.config.GuestAdditionsMode = "attach"
				}
			}
		}
	}

	if b.config.GuestAdditionsPath == "" && b.config.GuestAdditionsMode == "attach" {
		b.config.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

		if _, err := os.Stat(b.config.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				b.config.GuestAdditionsPath = ""
			}
		}
	}

	for _, isoPath := range b.config.SecondaryDvdImages {
		if _, err := os.Stat(isoPath); os.IsNotExist(err) {
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Secondary Dvd image does not exist: %s", err))
			}
		}
	}

	numberOfIsos := len(b.config.SecondaryDvdImages)

	if b.config.GuestAdditionsMode == "attach" {
		if _, err := os.Stat(b.config.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Guest additions iso does not exist: %s", err))
			}
		}

		numberOfIsos = numberOfIsos + 1
	}

	if b.config.Generation < 2 && numberOfIsos > 2 {
		if b.config.GuestAdditionsMode == "attach" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are only 2 ide controllers available, so we can't support guest additions and these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are only 2 ide controllers available, so we can't support these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		}
	} else if b.config.Generation > 1 && len(b.config.SecondaryDvdImages) > 16 {
		if b.config.GuestAdditionsMode == "attach" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are not enough drive letters available for scsi (limited to 16), so we can't support guest additions and these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are not enough drive letters available for scsi (limited to 16), so we can't support these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		}
	}

	if b.config.EnableVirtualizationExtensions {
		hasVirtualMachineVirtualizationExtensions, err := powershell.HasVirtualMachineVirtualizationExtensions()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed detecting virtual machine virtualization extensions support: %s", err))
		} else {
			if !hasVirtualMachineVirtualizationExtensions {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("This version of Hyper-V does not support virtual machine virtualization extension. Please use Windows 10 or Windows Server 2016 or newer."))
			}
		}
	}

	virtualMachineExists, err := powershell.DoesVirtualMachineExist(b.config.CloneFromVMName)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed detecting if virtual machine to clone from exists: %s", err))
	} else {
		if !virtualMachineExists {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Virtual machine '%s' to clone from does not exist.", b.config.CloneFromVMName))
		} else {
			if b.config.CloneFromSnapshotName != "" {
				virtualMachineSnapshotExists, err := powershell.DoesVirtualMachineSnapshotExist(b.config.CloneFromVMName, b.config.CloneFromSnapshotName)
				if err != nil {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed detecting if virtual machine snapshot to clone from exists: %s", err))
				} else {
					if !virtualMachineSnapshotExists {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf("Virtual machine snapshot '%s' on virtual machine '%s' to clone from does not exist.", b.config.CloneFromSnapshotName, b.config.CloneFromVMName))
					}
				}
			}

			virtualMachineOn, err := powershell.IsVirtualMachineOn(b.config.CloneFromVMName)
			if err != nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed detecting if virtual machine to clone is running: %s", err))
			} else {
				if virtualMachineOn {
					warning := fmt.Sprintf("Cloning from a virtual machine that is running.")
					warnings = appendWarnings(warnings, warning)
				}
			}
		}
	}

	// Warnings

	if b.config.ShutdownCommand == "" {
		warnings = appendWarnings(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	warning := b.checkHostAvailableMemory()
	if warning != "" {
		warnings = appendWarnings(warnings, warning)
	}

	if b.config.EnableVirtualizationExtensions {
		if b.config.EnableDynamicMemory {
			warning = fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, dynamic memory should not be allowed.")
			warnings = appendWarnings(warnings, warning)
		}

		if !b.config.EnableMacSpoofing {
			warning = fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, mac spoofing should be allowed.")
			warnings = appendWarnings(warnings, warning)
		}

		if b.config.RamSize < MinNestedVirtualizationRamSize {
			warning = fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, there should be 4GB or more memory set for the vm, otherwise Hyper-V may fail to start any nested VMs.")
			warnings = appendWarnings(warnings, warning)
		}
	}

	if b.config.SwitchVlanId != "" {
		if b.config.SwitchVlanId != b.config.VlanId {
			warning = fmt.Sprintf("Switch network adaptor vlan should match virtual machine network adaptor vlan. The switch will not be able to see traffic from the VM.")
			warnings = appendWarnings(warnings, warning)
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a Hyperv appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Hyperv
	driver, err := hypervcommon.NewHypervPS4Driver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Hyper-V driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&hypervcommon.StepCreateTempDir{},
		&hypervcommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			ResultKey:    "iso_path",
			Url:          b.config.ISOUrls,
			Extension:    b.config.TargetExtension,
			TargetPath:   b.config.TargetPath,
		},
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&hypervcommon.StepCreateSwitch{
			SwitchName: b.config.SwitchName,
		},
		&hypervcommon.StepCloneVM{
			CloneFromVMName:                b.config.CloneFromVMName,
			CloneFromSnapshotName:          b.config.CloneFromSnapshotName,
			CloneAllSnapshots:              b.config.CloneAllSnapshots,
			VMName:                         b.config.VMName,
			SwitchName:                     b.config.SwitchName,
			RamSize:                        b.config.RamSize,
			Cpu:                            b.config.Cpu,
			EnableMacSpoofing:              b.config.EnableMacSpoofing,
			EnableDynamicMemory:            b.config.EnableDynamicMemory,
			EnableSecureBoot:               b.config.EnableSecureBoot,
			EnableVirtualizationExtensions: b.config.EnableVirtualizationExtensions,
		},

		&hypervcommon.StepEnableIntegrationService{},

		&hypervcommon.StepMountDvdDrive{
			Generation: b.config.Generation,
		},
		&hypervcommon.StepMountFloppydrive{
			Generation: b.config.Generation,
		},

		&hypervcommon.StepMountGuestAdditions{
			GuestAdditionsMode: b.config.GuestAdditionsMode,
			GuestAdditionsPath: b.config.GuestAdditionsPath,
			Generation:         b.config.Generation,
		},

		&hypervcommon.StepMountSecondaryDvdImages{
			IsoPaths:   b.config.SecondaryDvdImages,
			Generation: b.config.Generation,
		},

		&hypervcommon.StepConfigureVlan{
			VlanId:       b.config.VlanId,
			SwitchVlanId: b.config.SwitchVlanId,
		},

		&hypervcommon.StepRun{
			BootWait: b.config.BootWait,
		},

		&hypervcommon.StepTypeBootCommand{
			BootCommand: b.config.BootCommand,
			SwitchName:  b.config.SwitchName,
			Ctx:         b.config.ctx,
		},

		// configure the communicator ssh, winrm
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      hypervcommon.CommHost,
			SSHConfig: hypervcommon.SSHConfigFunc(&b.config.SSHConfig),
		},

		// provision requires communicator to be setup
		&common.StepProvision{},

		&hypervcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},

		// wait for the vm to be powered off
		&hypervcommon.StepWaitForPowerOff{},

		// remove the secondary dvd images
		// after we power down
		&hypervcommon.StepUnmountSecondaryDvdImages{},
		&hypervcommon.StepUnmountGuestAdditions{},
		&hypervcommon.StepUnmountDvdDrive{},
		&hypervcommon.StepUnmountFloppyDrive{
			Generation: b.config.Generation,
		},
		&hypervcommon.StepExportVm{
			OutputDir:      b.config.OutputDir,
			SkipCompaction: b.config.SkipCompaction,
		},

		// the clean up actions for each step will be executed reverse order
	}

	// Run the steps.
	if b.config.PackerDebug {
		pauseFn := common.MultistepDebugFn(ui)
		state.Put("pauseFn", pauseFn)
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: pauseFn,
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return hypervcommon.NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func appendWarnings(slice []string, data ...string) []string {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]string, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func (b *Builder) checkRamSize() error {
	if b.config.RamSize == 0 {
		b.config.RamSize = DefaultRamSize
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", b.config.RamSize))

	if b.config.RamSize < MinRamSize {
		return fmt.Errorf("ram_size: Virtual machine requires memory size >= %v MB, but defined: %v", MinRamSize, b.config.RamSize)
	} else if b.config.RamSize > MaxRamSize {
		return fmt.Errorf("ram_size: Virtual machine requires memory size <= %v MB, but defined: %v", MaxRamSize, b.config.RamSize)
	}

	return nil
}

func (b *Builder) checkHostAvailableMemory() string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		freeMB := powershell.GetHostAvailableMemory()

		if (freeMB - float64(b.config.RamSize)) < LowRam {
			return fmt.Sprintf("Hyper-V might fail to create a VM if there is not enough free memory in the system.")
		}
	}

	return ""
}

func (b *Builder) detectSwitchName() string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		// no switch name, try to get one attached to a online network adapter
		onlineSwitchName, err := hyperv.GetExternalOnlineVirtualSwitch()
		if onlineSwitchName != "" && err == nil {
			return onlineSwitchName
		}
	}

	return fmt.Sprintf("packer-%s", b.config.PackerBuildName)
}
