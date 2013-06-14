package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// MultistepDebugFn will return a proper multistep.DebugPauseFn to
// use for debugging if you're using multistep in your builder.
func MultistepDebugFn(ui packer.Ui) multistep.DebugPauseFn {
	return func(loc multistep.DebugLocation, name string, state map[string]interface{}) {
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
			"Pausing %s step '%s'. Press any key to continue.\n",
			locationString, name)

		ui.Say(message)
		ui.Ask("")
	}
}
