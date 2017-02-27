package ebssurrogate

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

// StepRegisterAMI creates the AMI.
type StepRegisterAMI struct {
	RootDevice   RootBlockDevice
	BlockDevices []*ec2.BlockDeviceMapping
}

func (s *StepRegisterAMI) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	snapshotId := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Registering the AMI...")

	blockDevicesExcludingRoot := make([]*ec2.BlockDeviceMapping, 0, len(s.BlockDevices)-1)
	for _, blockDevice := range s.BlockDevices {
		if *blockDevice.DeviceName == s.RootDevice.SourceDeviceName {
			continue
		}

		blockDevicesExcludingRoot = append(blockDevicesExcludingRoot, blockDevice)
	}

	blockDevicesExcludingRoot = append(blockDevicesExcludingRoot, s.RootDevice.createBlockDeviceMapping(snapshotId))

	registerOpts := &ec2.RegisterImageInput{
		Name:                &config.AMIName,
		Architecture:        aws.String(ec2.ArchitectureValuesX8664),
		RootDeviceName:      aws.String(s.RootDevice.DeviceName),
		VirtualizationType:  aws.String(config.AMIVirtType),
		BlockDeviceMappings: blockDevicesExcludingRoot,
	}

	if config.AMIEnhancedNetworking {
		// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
		// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
		registerOpts.SriovNetSupport = aws.String("simple")

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

	// Wait for the image to become ready
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   awscommon.AMIStateRefreshFunc(ec2conn, *registerResp.ImageId),
		StepState: state,
	}

	ui.Say("Waiting for AMI to become ready...")
	if _, err := awscommon.WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(state multistep.StateBag) {}
