package vmware

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
	"text/template"
)

type toolsUploadPathTemplate struct {
	Flavor string
}

type stepUploadTools struct{}

func (*stepUploadTools) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	if config.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	comm := state["communicator"].(packer.Communicator)
	tools_source := state["tools_upload_source"].(string)
	ui := state["ui"].(packer.Ui)

	ui.Say(fmt.Sprintf("Uploading the '%s' VMware Tools", config.ToolsUploadFlavor))
	f, err := os.Open(tools_source)
	if err != nil {
		state["error"] = fmt.Errorf("Error opening VMware Tools ISO: %s", err)
		return multistep.ActionHalt
	}
	defer f.Close()

	tplData := &toolsUploadPathTemplate{Flavor: config.ToolsUploadFlavor}
	var processedPath bytes.Buffer
	t := template.Must(template.New("path").Parse(config.ToolsUploadPath))
	t.Execute(&processedPath, tplData)

	if err := comm.Upload(processedPath.String(), f); err != nil {
		state["error"] = fmt.Errorf("Error uploading VMware Tools: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*stepUploadTools) Cleanup(map[string]interface{}) {}
