package virtualbox

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepPrepareOutputDir struct{}

func (stepPrepareOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	if _, err := os.Stat(config.OutputDir); err == nil && config.PackerForce {
		ui.Say("Deleting previous output directory...")
		os.RemoveAll(config.OutputDir)
	}

	// Create the directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Make sure we can write in the directory
	f, err := os.Create(filepath.Join(config.OutputDir, "_packer_perm_check"))
	if err != nil {
		err = fmt.Errorf("Couldn't write to output directory: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}
	f.Close()
	os.Remove(f.Name())

	return multistep.ActionContinue
}

func (stepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		config := state.Get("config").(*config)
		ui := state.Get("ui").(packer.Ui)

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
