package hyperone

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepMountChroot struct{}

func (s *stepMountChroot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	device := state.Get("device").(string)

	log.Printf("Mount path: %s", config.ChrootMountPath)

	ui.Say(fmt.Sprintf("Creating mount directory: %s", config.ChrootMountPath))

	opts := ""
	if len(config.MountOptions) > 0 {
		opts = "-o " + strings.Join(config.MountOptions, " -o ")
	}

	deviceMount := device
	if config.MountPartition != "" {
		deviceMount = fmt.Sprintf("%s%s", device, config.MountPartition)
	}

	commands := []string{
		fmt.Sprintf("mkdir -m 755 -p %s", config.ChrootMountPath),
		fmt.Sprintf("mount %s %s %s", opts, deviceMount, config.ChrootMountPath),
	}

	err := runCommands(commands, config.ctx, state)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("mount_path", config.ChrootMountPath)

	return multistep.ActionContinue
}

func (s *stepMountChroot) Cleanup(state multistep.StateBag) {}
