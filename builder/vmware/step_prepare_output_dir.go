package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepPrepareOutputDir struct {
	dir OutputDir
}

func (s *stepPrepareOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	dir := s.outputDir(state)
	dir.SetOutputDir(config.OutputDir)

	exists, err := dir.DirExists()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if exists {
		if config.PackerForce {
			ui.Say("Deleting previous output directory...")
			dir.RemoveAll()
		} else {
			state.Put("error", fmt.Errorf("Output directory '%s' already exists.", config.OutputDir))
			return multistep.ActionHalt
		}
	}

	if err := dir.MkdirAll(); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.dir = dir

	state.Put("dir", dir)

	return multistep.ActionContinue
}

func (s *stepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		ui := state.Get("ui").(packer.Ui)

		if s.dir != nil {
			ui.Say("Deleting output directory...")
			for i := 0; i < 5; i++ {
				err := s.dir.RemoveAll()
				if err == nil {
					break
				}

				log.Printf("Error removing output dir: %s", err)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func (s *stepPrepareOutputDir) outputDir(state multistep.StateBag) (dir OutputDir) {
	driver := state.Get("driver").(Driver)

	switch d := driver.(type) {
	case OutputDir:
		log.Printf("Using driver as the OutputDir implementation")
		dir = d
	default:
		log.Printf("Using localOutputDir implementation")
		dir = new(localOutputDir)
	}

	return
}
