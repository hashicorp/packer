// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package amazonebs

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type config struct {
	AccessKey string `json:"access_key"`
	AMIName   string `json:"ami_name"`
	Region    string
	SecretKey string `json:"secret_key"`
	SourceAmi string `json:"source_ami"`
}

type Builder struct {
	config config
}

func (b *Builder) Prepare(raw interface{}) (err error) {
	// Marshal and unmarshal the raw configuration as a way to get it
	// into our "config" struct.
	// TODO: Use the reflection package and provide this as an API for
	// better error messages
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonBytes, &b.config)
	if err != nil {
		return
	}

	log.Printf("Config: %+v\n", b.config)

	// TODO: Validate the configuration
	return
}

func (b *Builder) Run(build packer.Build, ui packer.Ui) {
	auth := aws.Auth{b.config.AccessKey, b.config.SecretKey}
	region := aws.Regions[b.config.Region]
	ec2conn := ec2.New(auth, region)

	runOpts := &ec2.RunInstances{
		ImageId:      b.config.SourceAmi,
		InstanceType: "m1.small",
		MinCount:     0,
		MaxCount:     0,
	}

	ui.Say("Launching a source AWS instance...\n")
	runResp, err := ec2conn.RunInstances(runOpts)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	instance := &runResp.Instances[0]
	log.Printf("instance id: %s\n", instance.InstanceId)
	ui.Say("Waiting for instance to become ready...\n")
	err = waitForState(ec2conn, instance, "running")
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Stop the instance so we can create an AMI from it
	_, err = ec2conn.StopInstances(instance.InstanceId)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Wait for the instance to actual stop
	// TODO: Handle diff source states, i.e. this force state sucks
	ui.Say("Waiting for the instance to stop...\n")
	instance.State.Name = "stopping"
	err = waitForState(ec2conn, instance, "stopped")
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Create the image
	ui.Say("Creating the AMI...\n")
	createOpts := &ec2.CreateImage{
		InstanceId: instance.InstanceId,
		Name:       b.config.AMIName,
	}

	createResp, err := ec2conn.CreateImage(createOpts)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	ui.Say("AMI: %s\n", createResp.ImageId)

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...\n")
	for {
		imageResp, err := ec2conn.Images([]string{createResp.ImageId}, ec2.NewFilter())
		if err != nil {
			ui.Error("%s\n", err.Error())
			return
		}

		if imageResp.Images[0].State == "available" {
			break
		}
	}

	// Make sure we clean up the instance by terminating it, no matter what
	defer func() {
		// TODO: error handling
		ui.Say("Terminating the source AWS instance...\n")
		ec2conn.TerminateInstances([]string{instance.InstanceId})
	}()
}

func waitForState(ec2conn *ec2.EC2, i *ec2.Instance, target string) (err error) {
	log.Printf("Waiting for instance state to become: %s\n", target)

	original := i.State.Name
	for i.State.Name == original {
		var resp *ec2.InstancesResp
		resp, err = ec2conn.Instances([]string{i.InstanceId}, ec2.NewFilter())
		if err != nil {
			return
		}

		i = &resp.Reservations[0].Instances[0]

		time.Sleep(2 * time.Second)
	}

	if i.State.Name != target {
		return fmt.Errorf("unexpected target state '%s', wanted '%s'", i.State.Name, target)
	}

	return
}
