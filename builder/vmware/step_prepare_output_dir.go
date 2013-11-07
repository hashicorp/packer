package vmware

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepPrepareOutputDir struct{}

func (s *stepPrepareOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	dir := s.outputDir(state)
	exists, err := dir.DirExists(config.OutputDir)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if exists && config.PackerForce {
		ui.Say("Deleting previous output directory...")
		dir.RemoveAll(config.OutputDir)
	}

	if err := dir.MkdirAll(config.OutputDir); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		config := state.Get("config").(*config)
		ui := state.Get("ui").(packer.Ui)

		dir := s.outputDir(state)
		ui.Say("Deleting output directory...")
		for i := 0; i < 5; i++ {
			err := dir.RemoveAll(config.OutputDir)
			if err == nil {
				break
			}

			log.Printf("Error removing output dir: %s", err)
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *stepPrepareOutputDir) outputDir(state multistep.StateBag) (dir OutputDir) {
	driver := state.Get("driver").(Driver)

	switch d := driver.(type) {
	case OutputDir:
		dir = d
	default:
		dir = new(localOutputDir)
	}

	return
}
