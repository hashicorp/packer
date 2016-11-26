package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	retry "github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

type StepCreateTags struct {
	Tags         map[string]string
	SnapshotTags map[string]string
}

func (s *StepCreateTags) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)

	if len(s.Tags) == 0 && len(s.SnapshotTags) == 0 {
		return multistep.ActionContinue
	}

	// Adds tags to AMIs and snapshots
	for region, ami := range amis {
		ui.Say(fmt.Sprintf("Adding tags to AMI (%s)...", ami))

		// Convert tags to ec2.Tag format
		amiTags := ConvertToEC2Tags(s.Tags, ui)
		ui.Say(fmt.Sprintf("Snapshot tags:"))
		snapshotTags := ConvertToEC2Tags(s.SnapshotTags, ui)

		// Declare list of resources to tag
		awsConfig := aws.Config{
			Credentials: ec2conn.Config.Credentials,
			Region:      aws.String(region),
		}
		session, err := session.NewSession(&awsConfig)
		if err != nil {
			err := fmt.Errorf("Error creating AWS session: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		regionconn := ec2.New(session)

		// Retrieve image list for given AMI
		resourceIds := []*string{&ami}
		imageResp, err := regionconn.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: resourceIds,
		})

		if err != nil {
			err := fmt.Errorf("Error retrieving details for AMI (%s): %s", ami, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(imageResp.Images) == 0 {
			err := fmt.Errorf("Error retrieving details for AMI (%s), no images found", ami)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		image := imageResp.Images[0]
		snapshotIds := []*string{}

		// Add only those with a Snapshot ID, i.e. not Ephemeral
		for _, device := range image.BlockDeviceMappings {
			if device.Ebs != nil && device.Ebs.SnapshotId != nil {
				ui.Say(fmt.Sprintf("Tagging snapshot: %s", *device.Ebs.SnapshotId))
				resourceIds = append(resourceIds, device.Ebs.SnapshotId)
				snapshotIds = append(snapshotIds, device.Ebs.SnapshotId)
			}
		}

		// Retry creating tags for about 2.5 minutes
		err = retry.Retry(0.2, 30, 11, func() (bool, error) {
			// Tag images and snapshots
			_, err := regionconn.CreateTags(&ec2.CreateTagsInput{
				Resources: resourceIds,
				Tags:      amiTags,
			})
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidAMIID.NotFound" ||
					awsErr.Code() == "InvalidSnapshot.NotFound" {
					return false, nil
				}
			}

			// Override tags on snapshots
			_, err = regionconn.CreateTags(&ec2.CreateTagsInput{
				Resources: snapshotIds,
				Tags:      snapshotTags,
			})
			if err == nil {
				return true, nil
			}
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidSnapshot.NotFound" {
					return false, nil
				}
			}
			return true, err
		})

		if err != nil {
			err := fmt.Errorf("Error adding tags to Resources (%#v): %s", resourceIds, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateTags) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

func ConvertToEC2Tags(tags map[string]string, ui packer.Ui) []*ec2.Tag {
	var amiTags []*ec2.Tag
	for key, value := range tags {
		ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"", key, value))
		amiTags = append(amiTags, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return amiTags
}
