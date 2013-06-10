package vmware

import (
	"github.com/mitchellh/multistep"
	"log"
)

// This step downloads the ISO specified.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   iso_path string
type stepDownloadISO struct{}

func (stepDownloadISO) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)

	log.Printf("Acquiring lock to download the ISO.")

	state["iso_path"] = config.ISOUrl

	return multistep.ActionContinue
}

func (stepDownloadISO) Cleanup(map[string]interface{}) {}
