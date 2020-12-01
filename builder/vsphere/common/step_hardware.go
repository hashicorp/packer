//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type HardwareConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type HardwareConfig struct {
	// Number of CPU cores.
	CPUs int32 `mapstructure:"CPUs"`
	// Number of CPU cores per socket.
	CpuCores int32 `mapstructure:"cpu_cores"`
	// Amount of reserved CPU resources in MHz.
	CPUReservation int64 `mapstructure:"CPU_reservation"`
	// Upper limit of available CPU resources in MHz.
	CPULimit int64 `mapstructure:"CPU_limit"`
	// Enable CPU hot plug setting for virtual machine. Defaults to `false`.
	CpuHotAddEnabled bool `mapstructure:"CPU_hot_plug"`
	// Amount of RAM in MB.
	RAM int64 `mapstructure:"RAM"`
	// Amount of reserved RAM in MB.
	RAMReservation int64 `mapstructure:"RAM_reservation"`
	// Reserve all available RAM. Defaults to `false`. Cannot be used together
	// with `RAM_reservation`.
	RAMReserveAll bool `mapstructure:"RAM_reserve_all"`
	// Enable RAM hot plug setting for virtual machine. Defaults to `false`.
	MemoryHotAddEnabled bool `mapstructure:"RAM_hot_plug"`
	// Amount of video memory in KB.
	VideoRAM int64 `mapstructure:"video_ram"`
	// vGPU profile for accelerated graphics. See [NVIDIA GRID vGPU documentation](https://docs.nvidia.com/grid/latest/grid-vgpu-user-guide/index.html#configure-vmware-vsphere-vm-with-vgpu)
	// for examples of profile names. Defaults to none.
	VGPUProfile string `mapstructure:"vgpu_profile"`
	// Enable nested hardware virtualization for VM. Defaults to `false`.
	NestedHV bool `mapstructure:"NestedHV"`
	// Set the Firmware for virtual machine. Supported values: `bios`, `efi` or `efi-secure`. Defaults to `bios`.
	Firmware string `mapstructure:"firmware"`
	// During the boot, force entry into the BIOS setup screen. Defaults to `false`.
	ForceBIOSSetup bool `mapstructure:"force_bios_setup"`
}

func (c *HardwareConfig) Prepare() []error {
	var errs []error

	if c.RAMReservation > 0 && c.RAMReserveAll != false {
		errs = append(errs, fmt.Errorf("'RAM_reservation' and 'RAM_reserve_all' cannot be used together"))
	}

	if c.Firmware != "" && c.Firmware != "bios" && c.Firmware != "efi" && c.Firmware != "efi-secure" {
		errs = append(errs, fmt.Errorf("'firmware' must be '', 'bios', 'efi' or 'efi-secure'"))
	}

	return errs
}

type StepConfigureHardware struct {
	Config *HardwareConfig
}

func (s *StepConfigureHardware) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	if *s.Config != (HardwareConfig{}) {
		ui.Say("Customizing hardware...")

		err := vm.Configure(&driver.HardwareConfig{
			CPUs:                s.Config.CPUs,
			CpuCores:            s.Config.CpuCores,
			CPUReservation:      s.Config.CPUReservation,
			CPULimit:            s.Config.CPULimit,
			RAM:                 s.Config.RAM,
			RAMReservation:      s.Config.RAMReservation,
			RAMReserveAll:       s.Config.RAMReserveAll,
			NestedHV:            s.Config.NestedHV,
			CpuHotAddEnabled:    s.Config.CpuHotAddEnabled,
			MemoryHotAddEnabled: s.Config.MemoryHotAddEnabled,
			VideoRAM:            s.Config.VideoRAM,
			VGPUProfile:         s.Config.VGPUProfile,
			Firmware:            s.Config.Firmware,
			ForceBIOSSetup:      s.Config.ForceBIOSSetup,
		})
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigureHardware) Cleanup(multistep.StateBag) {}
