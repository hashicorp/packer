package chroot

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/random"
	confighelper "github.com/hashicorp/packer/packer-plugin-sdk/template/config"
)

// StepRegisterAMI creates the AMI.
type StepRegisterAMI struct {
	PollingConfig            *awscommon.AWSPollingConfig
	RootVolumeSize           int64
	EnableAMIENASupport      confighelper.Trilean
	EnableAMISriovNetSupport bool
	AMISkipBuildRegion       bool
}

func (s *StepRegisterAMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	snapshotID := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Registering the AMI...")

	var registerOpts *ec2.RegisterImageInput

	// Create the image
	amiName := config.AMIName
	state.Put("intermediary_image", false)
	if config.AMIEncryptBootVolume.True() || s.AMISkipBuildRegion {
		state.Put("intermediary_image", true)

		// From AWS SDK docs: You can encrypt a copy of an unencrypted snapshot,
		// but you cannot use it to create an unencrypted copy of an encrypted
		// snapshot. Your default CMK for EBS is used unless you specify a
		// non-default key using KmsKeyId.

		// If encrypt_boot is nil or true, we need to create a temporary image
		// so that in step_region_copy, we can copy it with the correct
		// encryption
		amiName = random.AlphaNum(7)
	}

	// Source Image is only required to be passed if the image is not from scratch
	if config.FromScratch {
		registerOpts = buildBaseRegisterOpts(config, nil, s.RootVolumeSize, snapshotID, amiName)
	} else {
		image := state.Get("source_image").(*ec2.Image)
		registerOpts = buildBaseRegisterOpts(config, image, s.RootVolumeSize, snapshotID, amiName)
	}

	if s.EnableAMISriovNetSupport {
		// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
		// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
		registerOpts.SriovNetSupport = aws.String("simple")
	}
	if s.EnableAMIENASupport.True() {
		// Set EnaSupport to true
		// As of February 2017, this applies to C5, I3, P2, R4, X1, and m4.16xlarge
		registerOpts.EnaSupport = aws.Bool(true)
	}

	registerResp, err := ec2conn.RegisterImage(registerOpts)
	if err != nil {
		state.Put("error", fmt.Errorf("Error registering AMI: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", *registerResp.ImageId))
	amis := make(map[string]string)
	amis[*ec2conn.Config.Region] = *registerResp.ImageId
	state.Put("amis", amis)

	ui.Say("Waiting for AMI to become ready...")
	if err := s.PollingConfig.WaitUntilAMIAvailable(ctx, ec2conn, *registerResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(state multistep.StateBag) {}

// Builds the base register opts with architecture, name, root block device, mappings, virtualizationtype
func buildBaseRegisterOpts(config *Config, sourceImage *ec2.Image, rootVolumeSize int64, snapshotID string, amiName string) *ec2.RegisterImageInput {
	var (
		mappings       []*ec2.BlockDeviceMapping
		rootDeviceName string
	)

	generatingNewBlockDeviceMappings := config.FromScratch || len(config.AMIMappings) > 0
	if generatingNewBlockDeviceMappings {
		mappings = config.AMIMappings.BuildEC2BlockDeviceMappings()
		rootDeviceName = config.RootDeviceName
	} else {
		// If config.FromScratch is false, source image must be set
		mappings = sourceImage.BlockDeviceMappings
		rootDeviceName = *sourceImage.RootDeviceName
	}

	newMappings := make([]*ec2.BlockDeviceMapping, len(mappings))
	for i, device := range mappings {
		newDevice := device
		if *newDevice.DeviceName == rootDeviceName {
			if newDevice.Ebs != nil {
				newDevice.Ebs.SnapshotId = aws.String(snapshotID)
			} else {
				newDevice.Ebs = &ec2.EbsBlockDevice{SnapshotId: aws.String(snapshotID)}
			}

			if generatingNewBlockDeviceMappings || rootVolumeSize > *newDevice.Ebs.VolumeSize {
				newDevice.Ebs.VolumeSize = aws.Int64(rootVolumeSize)
			}
		}

		// assume working from a snapshot, so we unset the Encrypted field if set,
		// otherwise AWS API will return InvalidParameter
		if newDevice.Ebs != nil && newDevice.Ebs.Encrypted != nil {
			newDevice.Ebs.Encrypted = nil
		}

		newMappings[i] = newDevice
	}

	if config.FromScratch {
		return &ec2.RegisterImageInput{
			Name:                &amiName,
			Architecture:        aws.String(config.Architecture),
			RootDeviceName:      aws.String(rootDeviceName),
			VirtualizationType:  aws.String(config.AMIVirtType),
			BlockDeviceMappings: newMappings,
		}
	}

	return buildRegisterOptsFromExistingImage(config, sourceImage, newMappings, rootDeviceName, amiName)
}

func buildRegisterOptsFromExistingImage(config *Config, image *ec2.Image, mappings []*ec2.BlockDeviceMapping, rootDeviceName string, amiName string) *ec2.RegisterImageInput {
	registerOpts := &ec2.RegisterImageInput{
		Name:                &amiName,
		Architecture:        image.Architecture,
		RootDeviceName:      &rootDeviceName,
		BlockDeviceMappings: mappings,
		VirtualizationType:  image.VirtualizationType,
	}

	if config.AMIVirtType != "" {
		registerOpts.VirtualizationType = aws.String(config.AMIVirtType)
	}

	if config.AMIVirtType != "hvm" {
		registerOpts.KernelId = image.KernelId
		registerOpts.RamdiskId = image.RamdiskId
	}
	return registerOpts
}
