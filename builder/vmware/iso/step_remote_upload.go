package iso

import (
	"context"
	"fmt"
	"log"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// stepRemoteUpload uploads some thing from the state bag to a remote driver
// (if it can) and stores that new remote path into the state bag.
type stepRemoteUpload struct {
	Key       string
	Message   string
	DoCleanup bool
}

func (s *stepRemoteUpload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	// Check if the driver is a remote driver (it should never be)
	remote, ok := driver.(vmwcommon.RemoteDriver)
	if !ok {
		return multistep.ActionContinue
	}

	// Inform the user that this component has been de-fanged
	ui.Say("The regular vmware builders do not have the ability to be uploaded. Please use the vmware-esx builders.")
	return multistep.ActionHalt
}

func (s *stepRemoteUpload) Cleanup(state multistep.StateBag) {}
