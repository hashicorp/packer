package chroot

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

// StepAttachVolume attaches the previously created volume to an
// available device location.
//
// Produces:
//   device string - The location where the volume was attached.
//   attach_cleanup CleanupFunc
type StepAttachVolume struct {
	attached bool
	volumeId string
}

func (s *StepAttachVolume) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	device := state.Get("device").(string)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)
	volumeId := state.Get("volume_id").(string)

	// For the API call, it expects "sd" prefixed devices.
	attachVolume := strings.Replace(device, "/xvd", "/sd", 1)

	ui.Say(fmt.Sprintf("Attaching the root volume to %s", attachVolume))
	_, err := ec2conn.AttachVolume(&ec2.AttachVolumeInput{
		InstanceId: instance.InstanceId,
		VolumeId:   &volumeId,
		Device:     &attachVolume,
	})
	if err != nil {
		err := fmt.Errorf("Error attaching volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Mark that we attached it so we can detach it later
	s.attached = true
	s.volumeId = volumeId

	// Wait for the volume to become attached
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"attaching"},
		StepState: state,
		Target:    "attached",
		Refresh: func() (interface{}, string, error) {
			attempts := 0
			for attempts < 30 {
				resp, err := ec2conn.DescribeVolumes(&ec2.DescribeVolumesInput{VolumeIds: []*string{&volumeId}})
				if err != nil {
					return nil, "", err
				}
				if len(resp.Volumes[0].Attachments) > 0 {
					a := resp.Volumes[0].Attachments[0]
					return a, *a.State, nil
				}
				// When Attachment on volume is not present sleep for 2s and retry
				attempts += 1
				ui.Say(fmt.Sprintf(
					"Volume %s show no attachments. Attempt %d/30. Sleeping for 2s and will retry.",
					volumeId, attempts))
				time.Sleep(2 * time.Second)
			}

			// Attachment on volume is not present after all attempts
			return nil, "", errors.New("No attachments on volume.")
		},
	}

	_, err = awscommon.WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("attach_cleanup", s)
	return multistep.ActionContinue
}

func (s *StepAttachVolume) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepAttachVolume) CleanupFunc(state multistep.StateBag) error {
	if !s.attached {
		return nil
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Detaching EBS volume...")
	_, err := ec2conn.DetachVolume(&ec2.DetachVolumeInput{VolumeId: &s.volumeId})
	if err != nil {
		return fmt.Errorf("Error detaching EBS volume: %s", err)
	}

	s.attached = false

	// Wait for the volume to detach
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"attaching", "attached", "detaching"},
		StepState: state,
		Target:    "detached",
		Refresh: func() (interface{}, string, error) {
			resp, err := ec2conn.DescribeVolumes(&ec2.DescribeVolumesInput{VolumeIds: []*string{&s.volumeId}})
			if err != nil {
				return nil, "", err
			}

			v := resp.Volumes[0]
			if len(v.Attachments) > 0 {
				return v, *v.Attachments[0].State, nil
			} else {
				return v, "detached", nil
			}
		},
	}

	_, err = awscommon.WaitForState(&stateChange)
	if err != nil {
		return fmt.Errorf("Error waiting for volume: %s", err)
	}

	return nil
}
