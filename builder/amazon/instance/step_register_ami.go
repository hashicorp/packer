package instance

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strconv"
	"text/template"
	"time"
)

type amiNameData struct {
	CreateTime string
}

type StepRegisterAMI struct{}

func (s *StepRegisterAMI) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*Config)
	manifestPath := state["remote_manifest_path"].(string)
	ui := state["ui"].(packer.Ui)

	// Parse the name of the AMI
	amiNameBuf := new(bytes.Buffer)
	tData := amiNameData{
		strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	t := template.Must(template.New("ami").Parse(config.AMIName))
	t.Execute(amiNameBuf, tData)
	amiName := amiNameBuf.String()

	ui.Say("Registering the AMI...")
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf(
			"ec2-register %s -n '%s' -O '%s' -W '%s'",
			manifestPath,
			amiName,
			config.AccessKey,
			config.SecretKey),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		state["error"] = fmt.Errorf("Error registering AMI: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	if cmd.ExitStatus != 0 {
		state["error"] = fmt.Errorf(
			"AMI registration failed. Please see the output above for more\n" +
				"details on what went wrong.")
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(map[string]interface{}) {}
