package common

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-amazon/builder/common/awserrors"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type StepCreateTags struct {
	AMISkipCreateImage bool

	Tags         map[string]string
	SnapshotTags map[string]string
	Ctx          interpolate.Context
}

func (s *StepCreateTags) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	session := state.Get("awsSession").(*session.Session)
	ui := state.Get("ui").(packersdk.Ui)

	if s.AMISkipCreateImage {
		ui.Say("Skipping AMI create tags...")
		return multistep.ActionContinue
	}

	amis := state.Get("amis").(map[string]string)

	if len(s.Tags) == 0 && len(s.SnapshotTags) == 0 {
		return multistep.ActionContinue
	}

	// Adds tags to AMIs and snapshots
	for region, ami := range amis {
		ui.Say(fmt.Sprintf("Adding tags to AMI (%s)...", ami))

		regionConn := ec2.New(session, &aws.Config{
			Region: aws.String(region),
		})

		// Retrieve image list for given AMI
		resourceIds := []*string{&ami}
		imageResp, err := regionConn.DescribeImages(&ec2.DescribeImagesInput{
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

		// Convert tags to ec2.Tag format
		ui.Say("Creating AMI tags")
		amiTags, err := TagMap(s.Tags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		amiTags.Report(ui)

		ui.Say("Creating snapshot tags")
		snapshotTags, err := TagMap(s.SnapshotTags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		snapshotTags.Report(ui)

		// Retry creating tags for about 2.5 minutes
		err = retry.Config{Tries: 11, ShouldRetry: func(error) bool {
			if awserrors.Matches(err, "InvalidAMIID.NotFound", "") || awserrors.Matches(err, "InvalidSnapshot.NotFound", "") {
				return true
			}
			return false
		},
			RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			// Tag images and snapshots

			var err error
			if len(amiTags) > 0 {
				_, err = regionConn.CreateTags(&ec2.CreateTagsInput{
					Resources: resourceIds,
					Tags:      amiTags,
				})
				if err != nil {
					return err
				}
			}

			// Override tags on snapshots
			if len(snapshotTags) > 0 {
				_, err = regionConn.CreateTags(&ec2.CreateTagsInput{
					Resources: snapshotIds,
					Tags:      snapshotTags,
				})
			}
			return err
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
