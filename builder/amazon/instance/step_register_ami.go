package instance

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
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
	config := state["config"].(*Config)
	ec2conn := state["ec2"].(*ec2.EC2)
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
	registerOpts := &ec2.RegisterImage{
		ImageLocation: manifestPath,
		Name:          amiName,
	}

	registerResp, err := ec2conn.RegisterImage(registerOpts)
	if err != nil {
		state["error"] = fmt.Errorf("Error registering AMI: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", registerResp.ImageId))
	amis := make(map[string]string)
	amis[config.Region] = registerResp.ImageId
	state["amis"] = amis

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	if err := awscommon.WaitForAMI(ec2conn, registerResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(map[string]interface{}) {}
