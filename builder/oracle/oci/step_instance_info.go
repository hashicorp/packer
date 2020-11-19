package oci

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		driver = state.Get("driver").(Driver)
		ui     = state.Get("ui").(packersdk.Ui)
		id     = state.Get("instance_id").(string)
	)

	ip, err := driver.GetInstanceIP(ctx, id)
	if err != nil {
		err = fmt.Errorf("Error getting instance's IP: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_ip", ip)

	ui.Say(fmt.Sprintf("Instance has IP: %s.", ip))

	return multistep.ActionContinue
}

func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
