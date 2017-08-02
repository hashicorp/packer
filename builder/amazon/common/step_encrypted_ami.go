package common

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepCreateEncryptedAMICopy struct {
	image             *ec2.Image
	KeyID             string
	EncryptBootVolume bool
	Name              string
	AMIMappings       []BlockDevice
}

func (s *StepCreateEncryptedAMICopy) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	kmsKeyId := s.KeyID

	// Encrypt boot not set, so skip step
	if !s.EncryptBootVolume {
		if kmsKeyId != "" {
			log.Printf("Ignoring KMS Key ID: %s, encrypted=false", kmsKeyId)
		}
		return multistep.ActionContinue
	}

	ui.Say("Creating Encrypted AMI Copy")

	amis := state.Get("amis").(map[string]string)
	var region, id string
	if amis != nil {
		for region, id = range amis {
			break // There is only ever one region:ami pair in this map
		}
	}

	ui.Say(fmt.Sprintf("Copying AMI: %s(%s)", region, id))

	if kmsKeyId != "" {
		ui.Say(fmt.Sprintf("Encrypting with KMS Key ID: %s", kmsKeyId))
	}

	copyOpts := &ec2.CopyImageInput{
		Name:          &s.Name, // Try to overwrite existing AMI
		SourceImageId: aws.String(id),
		SourceRegion:  aws.String(region),
		Encrypted:     aws.Bool(true),
		KmsKeyId:      aws.String(kmsKeyId),
	}

	copyResp, err := ec2conn.CopyImage(copyOpts)
	if err != nil {
		err := fmt.Errorf("Error copying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the copy to become ready
	stateChange := StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   AMIStateRefreshFunc(ec2conn, *copyResp.ImageId),
		StepState: state,
	}

	ui.Say("Waiting for AMI copy to become ready...")
	if _, err := WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for AMI Copy: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Get the encrypted AMI image, we need the new snapshot id's
	encImagesResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{aws.String(*copyResp.ImageId)}})
	if err != nil {
		err := fmt.Errorf("Error searching for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	encImage := encImagesResp.Images[0]
	var encSnapshots []string
	for _, blockDevice := range encImage.BlockDeviceMappings {
		if blockDevice.Ebs != nil && blockDevice.Ebs.SnapshotId != nil {
			encSnapshots = append(encSnapshots, *blockDevice.Ebs.SnapshotId)
		}
	}

	// Get the unencrypted AMI image
	unencImagesResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{aws.String(id)}})
	if err != nil {
		err := fmt.Errorf("Error searching for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	unencImage := unencImagesResp.Images[0]

	// Remove unencrypted AMI
	ui.Say("Deregistering unencrypted AMI")
	deregisterOpts := &ec2.DeregisterImageInput{ImageId: aws.String(id)}
	if _, err := ec2conn.DeregisterImage(deregisterOpts); err != nil {
		ui.Error(fmt.Sprintf("Error deregistering AMI, may still be around: %s", err))
		return multistep.ActionHalt
	}

	// Remove associated unencrypted snapshot(s)
	ui.Say("Deleting unencrypted snapshots")
	snapshots := state.Get("snapshots").(map[string][]string)

	for _, blockDevice := range unencImage.BlockDeviceMappings {
		if blockDevice.Ebs != nil && blockDevice.Ebs.SnapshotId != nil {
			// If this packer run didn't create it, then don't delete it
			doDelete := true
			for _, origDevice := range s.AMIMappings {
				if origDevice.SnapshotId == *blockDevice.Ebs.SnapshotId {
					doDelete = false
				}
			}
			if doDelete == false {
				ui.Message(fmt.Sprintf("Keeping Snapshot ID: %s", *blockDevice.Ebs.SnapshotId))
				continue
			}
			ui.Message(fmt.Sprintf("Deleting Snapshot ID: %s", *blockDevice.Ebs.SnapshotId))
			deleteSnapOpts := &ec2.DeleteSnapshotInput{
				SnapshotId: aws.String(*blockDevice.Ebs.SnapshotId),
			}
			if _, err := ec2conn.DeleteSnapshot(deleteSnapOpts); err != nil {
				ui.Error(fmt.Sprintf("Error deleting snapshot, may still be around: %s", err))
				return multistep.ActionHalt
			}
		}
	}

	// Replace original AMI ID with Encrypted ID in state
	amis[region] = *copyResp.ImageId
	snapshots[region] = encSnapshots
	state.Put("amis", amis)
	state.Put("snapshots", snapshots)

	imagesResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{copyResp.ImageId}})
	if err != nil {
		err := fmt.Errorf("Error searching for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = imagesResp.Images[0]

	return multistep.ActionContinue
}

func (s *StepCreateEncryptedAMICopy) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deregistering the AMI because cancellation or error...")
	deregisterOpts := &ec2.DeregisterImageInput{ImageId: s.image.ImageId}
	if _, err := ec2conn.DeregisterImage(deregisterOpts); err != nil {
		ui.Error(fmt.Sprintf("Error deregistering AMI, may still be around: %s", err))
		return
	}
}
