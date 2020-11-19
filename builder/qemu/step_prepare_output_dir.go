package qemu

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepPrepareOutputDir struct{}

func (stepPrepareOutputDir) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	if _, err := os.Stat(config.OutputDir); err == nil && config.PackerForce {
		ui.Say("Deleting previous output directory...")
		os.RemoveAll(config.OutputDir)
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		config := state.Get("config").(*Config)
		ui := state.Get("ui").(packersdk.Ui)

		ui.Say("Deleting output directory...")
		for i := 0; i < 5; i++ {
			err := os.RemoveAll(config.OutputDir)
			if err == nil {
				break
			}

			log.Printf("Error removing output dir: %s", err)
			time.Sleep(2 * time.Second)
		}
	}
}
