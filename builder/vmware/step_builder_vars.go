package vmware

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
	httpPort := state["http_port"].(int)

	config.template.BuilderVars["flavor"] = config.ToolsUploadFlavor
	config.template.BuilderVars["http_ip"] = "10.0.2.2"
	config.template.BuilderVars["http_port"] = strconv.FormatInt(int64(httpPort), 10)

	// Pre-process the VM name so we can use that as a builder var
	if err := config.template.ProcessSingle("vm_name"); err != nil {
		state["error"] = fmt.Errorf("Error processing vm_name: %s", err)
		return multistep.ActionHalt
	}

	config.template.BuilderVars["vm_name"] = config.VMName

	return multistep.ActionContinue
}

func (s *stepBuilderVars) Cleanup(map[string]interface{}) {}
