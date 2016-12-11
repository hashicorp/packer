package common

import (
	"fmt"
	"os"

	"github.com/mitchellh/multistep"
)

// StepPrepareParallelsTools is a step that prepares parameters related
// to Parallels Tools.
//
// Uses:
//   driver Driver
//
// Produces:
//   parallels_tools_path string
type StepPrepareParallelsTools struct {
	ParallelsToolsFlavor string
	ParallelsToolsMode   string
}

func (s *StepPrepareParallelsTools) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	if s.ParallelsToolsMode == ParallelsToolsModeDisable {
		return multistep.ActionContinue
	}

	path, err := driver.ToolsISOPath(s.ParallelsToolsFlavor)

	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if _, err := os.Stat(path); err != nil {
		state.Put("error", fmt.Errorf(
			"Couldn't find Parallels Tools for the '%s' flavor! Please, check the\n"+
				"value of 'parallels_tools_flavor'. Valid flavors are: 'win', 'lin',\n"+
				"'mac', 'os2' and 'other'", s.ParallelsToolsFlavor))
		return multistep.ActionHalt
	}

	state.Put("parallels_tools_path", path)
	return multistep.ActionContinue
}

func (s *StepPrepareParallelsTools) Cleanup(multistep.StateBag) {}
