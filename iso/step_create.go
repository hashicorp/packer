package iso

import (
	"fmt"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
)

type CreateConfig struct {
	common.HardwareConfig `mapstructure:",squash"`

	DiskThinProvisioned bool   `mapstructure:"disk_thin_provisioned"`
	DiskControllerType  string `mapstructure:"disk_controller_type"`

	VMName        string `mapstructure:"vm_name"`
	Folder        string `mapstructure:"folder"`
	Host          string `mapstructure:"host"`
	ResourcePool  string `mapstructure:"resource_pool"`
	Datastore     string `mapstructure:"datastore"`
	GuestOSType   string `mapstructure:"guest_os_type"`
	Network       string `mapstructure:"network"`
	NetworkCard   string `mapstructure:"network_card"`
	USBController bool   `mapstructure:"usb_controller"`
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	// needed to avoid changing the original config in case of errors
	tmp := *c

	// do recursive calls
	errs = append(errs, tmp.HardwareConfig.Prepare()...)

	// check for errors
	if tmp.VMName == "" {
		errs = append(errs, fmt.Errorf("Target VM name is required"))
	}
	if tmp.Host == "" {
		errs = append(errs, fmt.Errorf("vSphere host is required"))
	}

	if len(errs) > 0 {
		return errs
	}

	// set default values
	if tmp.GuestOSType == "" {
		tmp.GuestOSType = "otherGuest"
	}

	// change the original config
	*c = tmp

	return []error{}
}

type StepCreateVM struct {
	Config *CreateConfig
}

func (s *StepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	ui.Say("Creating VM...")

	vm, err := d.CreateVM(&driver.CreateConfig{
		HardwareConfig: s.Config.HardwareConfig.ToDriverHardwareConfig(),

		DiskThinProvisioned: s.Config.DiskThinProvisioned,
		DiskControllerType:  s.Config.DiskControllerType,
		Name:                s.Config.VMName,
		Folder:              s.Config.Folder,
		Host:                s.Config.Host,
		ResourcePool:        s.Config.ResourcePool,
		Datastore:           s.Config.Datastore,
		GuestOS:             s.Config.GuestOSType,
		Network:             s.Config.Network,
		NetworkCard:         s.Config.NetworkCard,
		USBController:       s.Config.USBController,
	})

	if err != nil {
		state.Put("error", fmt.Errorf("error creating vm: %v", err))
		return multistep.ActionHalt
	}

	state.Put("vm", vm)
	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	st := state.Get("vm")
	if st == nil {
		return
	}
	vm := st.(*driver.VirtualMachine)

	ui.Say("Destroying VM...")

	err := vm.Destroy()
	if err != nil {
		ui.Error(err.Error())
	}
}
