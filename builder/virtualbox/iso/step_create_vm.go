package iso

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
)

// This step creates the actual virtual machine.
//
// Produces:
//   vmName string - The name of the VM
type stepCreateVM struct {
	vmName string
}

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	name := config.VMName

	commands := make([][]string, 14)
	commands[0] = []string{
		"createvm", "--name", name,
		"--ostype", config.GuestOSType, "--register",
	}
	commands[1] = []string{
		"modifyvm", name,
		"--boot1", "disk", "--boot2", "dvd", "--boot3", "none", "--boot4", "none",
	}
	commands[2] = []string{"modifyvm", name, "--cpus", strconv.Itoa(config.HWConfig.CpuCount)}
	commands[3] = []string{"modifyvm", name, "--memory", strconv.Itoa(config.HWConfig.MemorySize)}
	commands[4] = []string{"modifyvm", name, "--usb", map[bool]string{true: "on", false: "off"}[config.HWConfig.USB]}

	if strings.ToLower(config.HWConfig.Sound) == "none" {
		commands[5] = []string{"modifyvm", name, "--audio", config.HWConfig.Sound,
			"--audiocontroller", config.AudioController}
	} else {
		commands[5] = []string{"modifyvm", name, "--audio", config.HWConfig.Sound, "--audioin", "on", "--audioout", "on",
			"--audiocontroller", config.AudioController}
	}

	commands[6] = []string{"modifyvm", name, "--chipset", config.Chipset}
	commands[7] = []string{"modifyvm", name, "--firmware", config.Firmware}
	// Set the configured NIC type for all 8 possible NICs
	commands[8] = []string{"modifyvm", name,
		"--nictype1", config.NICType,
		"--nictype2", config.NICType,
		"--nictype3", config.NICType,
		"--nictype4", config.NICType,
		"--nictype5", config.NICType,
		"--nictype6", config.NICType,
		"--nictype7", config.NICType,
		"--nictype8", config.NICType}
	commands[9] = []string{"modifyvm", name, "--graphicscontroller", config.GfxController, "--vram", strconv.FormatUint(uint64(config.GfxVramSize), 10)}
	if config.RTCTimeBase == "UTC" {
		commands[10] = []string{"modifyvm", name, "--rtcuseutc", "on"}
	} else {
		commands[10] = []string{"modifyvm", name, "--rtcuseutc", "off"}
	}
	if config.NestedVirt == true {
		commands[11] = []string{"modifyvm", name, "--nested-hw-virt", "on"}
	} else {
		commands[11] = []string{"modifyvm", name, "--nested-hw-virt", "off"}
	}

	if config.GfxAccelerate3D {
		commands[12] = []string{"modifyvm", name, "--accelerate3d", "on"}
	} else {
		commands[12] = []string{"modifyvm", name, "--accelerate3d", "off"}
	}
	if config.GfxEFIResolution != "" {
		commands[13] = []string{"setextradata", name, "VBoxInternal2/EfiGraphicsResolution", config.GfxEFIResolution}
	}

	ui.Say("Creating virtual machine...")
	for _, command := range commands {
		err := driver.VBoxManage(command...)
		if err != nil {
			err := fmt.Errorf("Error creating VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the VM name property on the first command
		if s.vmName == "" {
			s.vmName = name
		}
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.vmName)

	return multistep.ActionContinue
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (config.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with VirtualBox host (keep_registered = true)")
		return
	}

	ui.Say("Deregistering and deleting VM...")
	if err := driver.Delete(s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %s", err))
	}
}
