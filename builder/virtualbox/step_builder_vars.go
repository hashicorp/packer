package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"strconv"
)

// This step sets the various builder variables available for templates.
//
// Uses:
//   http_port int
type stepBuilderVars struct{}

func (s *stepBuilderVars) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	httpPort := state["http_port"].(int)

	version, err := driver.Version()
	if err != nil {
		state["error"] = fmt.Errorf("Error reading VirtualBox version: %s", err)
		return multistep.ActionHalt
	}

	config.template.BuilderVars["http_ip"] = "10.0.2.2"
	config.template.BuilderVars["http_port"] = strconv.FormatInt(int64(httpPort), 10)
	config.template.BuilderVars["vbox_version"] = version

	// TODO(mitchellh): recursion problem here or something. Wah.
	config.template.BuilderVars["vm_name"] = config.VMName

	return multistep.ActionContinue
}

func (s *stepBuilderVars) Cleanup(map[string]interface{}) {}
