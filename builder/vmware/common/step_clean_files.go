package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
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
type StepCleanFiles struct{}

func (StepCleanFiles) Run(state multistep.StateBag) multistep.StepAction {
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
				// Only report the error if the file still exists. We do this
				// because sometimes the files naturally get removed on their
				// own as VMware does its own cleanup.
				if _, serr := os.Stat(path); serr == nil || !os.IsNotExist(serr) {
					state.Put("error", err)
					return multistep.ActionHalt
				}
			}
		}
	}

	return multistep.ActionContinue
}

func (StepCleanFiles) Cleanup(multistep.StateBag) {}
