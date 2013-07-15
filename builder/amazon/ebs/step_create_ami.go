package ebs

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"strconv"
	"text/template"
	"time"
)

type stepCreateAMI struct{}

type amiNameData struct {
	CreateTime string
}

func (s *stepCreateAMI) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	ec2conn := state["ec2"].(*ec2.EC2)
	instance := state["instance"].(*ec2.Instance)
	ui := state["ui"].(packer.Ui)

	// Parse the name of the AMI
	amiNameBuf := new(bytes.Buffer)
	tData := amiNameData{
		strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	t := template.Must(template.New("ami").Parse(config.AMIName))
	t.Execute(amiNameBuf, tData)
	amiName := amiNameBuf.String()

	// Create the image
	ui.Say(fmt.Sprintf("Creating the AMI: %s", amiName))
	createOpts := &ec2.CreateImage{
		InstanceId: instance.InstanceId,
		Name:       amiName,
	}

	createResp, err := ec2conn.CreateImage(createOpts)
	if err != nil {
		err := fmt.Errorf("Error creating AMI: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", createResp.ImageId))
	amis := make(map[string]string)
	amis[config.Region] = createResp.ImageId
	state["amis"] = amis

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	for {
		imageResp, err := ec2conn.Images([]string{createResp.ImageId}, ec2.NewFilter())
		if err != nil {
			err := fmt.Errorf("Error querying images: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if imageResp.Images[0].State == "available" {
			break
		}

		log.Printf("Image in state %s, sleeping 2s before checking again",
			imageResp.Images[0].State)

		time.Sleep(2 * time.Second)
	}

	return multistep.ActionContinue
}

func (s *stepCreateAMI) Cleanup(map[string]interface{}) {
	// No cleanup...
}
