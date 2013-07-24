package ebs

import (
	"bytes"
	"errors"
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
	imageId := createResp.ImageId
	ui.Say(fmt.Sprintf("AMI: %s", imageId))
	amis := make(map[string]string)
	amis[config.Region] = imageId
	state["amis"] = amis

	// To simplify testing, we abstract the call to ec2.Images into a function, imageFetch
	imageFetcher := func() (*ec2.ImagesResp, error) {
		return ec2conn.Images([]string{imageId}, ec2.NewFilter())
	}

	// Wait for the image to become ready
	timeout := 5 * time.Minute                        // maximum time to wait for
	wait := 2 * time.Second                           // time to wait between checks
	err = waitForAMI(ui, imageFetcher, timeout, wait) // Wait for the AMI to be ready
	if err != nil {
		message := fmt.Errorf(":1s: %s", err)
		state["error"] = message
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Wait for the image to become ready
//
// imageFetcher - function to encapsulate the call to ec2.Images.  simplifies testing
// timeout - if the image is not available after this duration, blow up
// wait - the amount of time to wait between checks to ec2
func waitForAMI(ui packer.Ui, imageFetcher func() (*ec2.ImagesResp, error), timeout, wait time.Duration) error {
	// encapsulates the for loop from #Run into a separate, easily testable function
	ui.Say("Waiting for AMI to become ready...")
	timeoutTimer := time.After(timeout)

	for {
		imageResp, err := imageFetcher()
		if err != nil {
			if ec2err := err.(*ec2.Error); ec2err.Code != "InvalidAMIID.NotFound" {
				return err
			}
		}

		if imageResp != nil { // imageResp check is required since it's possible we could be here in err (e.g. InvalidAMIID.NotFound)
			if imageResp.Images[0].State == "available" {
				break
			}

			log.Printf("Image in state %s, sleeping 2s before checking again",
				imageResp.Images[0].State)
		}

		select {
		case <-time.After(wait):
		case <-timeoutTimer:
			return errors.New("AMI not available after 5 minutes")
		}
	}

	return nil
}

func (s *stepCreateAMI) Cleanup(map[string]interface{}) {
	// No cleanup...
}
