package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

// These are the extensions of files that are important for the function
// of a VMware virtual machine. Any other file is discarded as part of the
// build.
var KeepFileExtensions = []string{".nvram", ".vmdk", ".vmsd", ".vmx", ".vmxf"}

// This step removes unnecessary files from the final result.
//
// Uses:
//   dir    OutputDir
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepCleanFiles struct{}

func (stepCleanFiles) Run(state multistep.StateBag) multistep.StepAction {
	dir := state.Get("dir").(OutputDir)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting unnecessary VMware files...")
	files, err := dir.ListFiles()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	for _, path := range files {
		// If the file isn't critical to the function of the
		// virtual machine, we get rid of it.
		keep := false
		ext := filepath.Ext(path)
		for _, goodExt := range KeepFileExtensions {
			if goodExt == ext {
				keep = true
				break
			}
		}

		if !keep {
			ui.Message(fmt.Sprintf("Deleting: %s", path))
			if err = dir.Remove(path); err != nil {
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (stepCleanFiles) Cleanup(multistep.StateBag) {}
