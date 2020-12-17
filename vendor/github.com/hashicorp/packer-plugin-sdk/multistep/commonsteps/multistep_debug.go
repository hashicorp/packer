package commonsteps

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// MultistepDebugFn will return a proper multistep.DebugPauseFn to
// use for debugging if you're using multistep in your builder.
func MultistepDebugFn(ui packersdk.Ui) multistep.DebugPauseFn {
	return func(loc multistep.DebugLocation, name string, state multistep.StateBag) {
		var locationString string
		switch loc {
		case multistep.DebugLocationAfterRun:
			locationString = "after run of"
		case multistep.DebugLocationBeforeCleanup:
			locationString = "before cleanup of"
		default:
			locationString = "at"
		}

		message := fmt.Sprintf(
			"Pausing %s step '%s'. Press enter to continue.",
			locationString, name)

		result := make(chan string, 1)
		go func() {
			line, err := ui.Ask(message)
			if err != nil {
				log.Printf("Error asking for input: %s", err)
			}

			result <- line
		}()

		for {
			select {
			case <-result:
				return
			case <-time.After(100 * time.Millisecond):
				if _, ok := state.GetOk(multistep.StateCancelled); ok {
					return
				}
			}
		}
	}
}
