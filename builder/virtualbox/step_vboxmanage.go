package virtualbox

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
	"text/template"
)

type commandTemplate struct {
	Name string
}

// This step executes additional VBoxManage commands as specified by the
// template.
//
// Uses:
//
// Produces:
type stepVBoxManage struct{}

func (s *stepVBoxManage) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	if len(config.VBoxManage) > 0 {
		ui.Say("Executing custom VBoxManage commands...")
	}

	tplData := &commandTemplate{
		Name: vmName,
	}

	for _, originalCommand := range config.VBoxManage {
		command := make([]string, len(originalCommand))
		copy(command, originalCommand)

		for i, arg := range command {
			var buf bytes.Buffer
			t := template.Must(template.New("arg").Parse(arg))
			t.Execute(&buf, tplData)
			command[i] = buf.String()
		}

		ui.Say(fmt.Sprintf("Executing: %s", strings.Join(command, " ")))
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error executing command: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepVBoxManage) Cleanup(state map[string]interface{}) {}
