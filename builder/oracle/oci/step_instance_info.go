package oci

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		driver = state.Get("driver").(Driver)
		ui     = state.Get("ui").(packer.Ui)
		id     = state.Get("instance_id").(string)
	)

	ip, err := driver.GetInstanceIP(id)
	if err != nil {
		err = fmt.Errorf("Error getting instance's public IP: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_ip", ip)

	ui.Say(fmt.Sprintf("Instance has public IP: %s.", ip))

	return multistep.ActionContinue
}

func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
