package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"time"
)

type OutputDir interface {
	FileExists(path string) bool
	MkdirAll(path string) error
	RemoveAll(path string) error
	DirType() string
}

type localOutputDir struct{}

func (localOutputDir) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (localOutputDir) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func (localOutputDir) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (localOutputDir) DirType() string {
	return "local"
}

type stepPrepareOutputDir struct{}

func (s *stepPrepareOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	for _, dir := range s.outputDirs(state) {
		if dir.FileExists(config.OutputDir) && config.PackerForce {
			ui.Say(fmt.Sprintf("Deleting previous %s output directory...", dir.DirType()))
			dir.RemoveAll(config.OutputDir)
		}

		if err := dir.MkdirAll(config.OutputDir); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		config := state.Get("config").(*config)
		ui := state.Get("ui").(packer.Ui)

		for _, dir := range s.outputDirs(state) {
			ui.Say(fmt.Sprintf("Deleting %s output directory...", dir.DirType()))
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
}

func (s *stepPrepareOutputDir) outputDirs(state multistep.StateBag) []OutputDir {
	driver := state.Get("driver").(Driver)
	dirs := []OutputDir{
		localOutputDir{},
	}

	if dir, ok := driver.(OutputDir); ok {
		dirs = append(dirs, dir)
	}

	return dirs
}
