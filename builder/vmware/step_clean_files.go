package vmware

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
//   config *config
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepCleanFiles struct{}

func (stepCleanFiles) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting unnecessary VMware files...")
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
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
				return os.Remove(path)
			}
		}

		return nil
	}

	if err := filepath.Walk(config.OutputDir, visit); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCleanFiles) Cleanup(multistep.StateBag) {}
