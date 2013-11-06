package qemu

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// stepWaitForShutdown waits for the shutdown of the currently running
// qemu VM.
type stepWaitForShutdown struct {
	Message string
}

func (s *stepWaitForShutdown) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	stopCh := make(chan struct{})
	defer close(stopCh)

	cancelCh := make(chan struct{})
	go func() {
		for {
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				close(cancelCh)
				return
			}

			select {
			case <-stopCh:
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}()

	ui.Say(s.Message)
	driver.WaitForShutdown(cancelCh)
	return multistep.ActionContinue
}

func (s *stepWaitForShutdown) Cleanup(state multistep.StateBag) {}
