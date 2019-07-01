package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/common/random"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepRunSpotInstance struct {
	AssociatePublicIpAddress          bool
	BlockDevices                      BlockDevices
	BlockDurationMinutes              int64
	Debug                             bool
	Comm                              *communicator.Config
	EbsOptimized                      bool
	ExpectedRootDevice                string
	IamInstanceProfile                string
	InstanceInitiatedShutdownBehavior string
	InstanceType                      string
	SourceAMI                         string
	SpotPrice                         string
	SpotPriceProduct                  string
	SpotTags                          TagMap
	SpotInstanceTypes                 []string
	Tags                              TagMap
	VolumeTags                        TagMap
	UserData                          string
	UserDataFile                      string
	Ctx                               interpolate.Context

	instanceId  string
	spotRequest *ec2.SpotInstanceRequest
}

func (s *StepRunSpotInstance) CalculateSpotPrice(az string, ec2conn ec2iface.EC2API) (string, error) {
	// Calculate the spot price for a given availability zone
	spotPrice := s.SpotPrice

	if spotPrice == "auto" {
		// Detect the spot price
		startTime := time.Now().Add(-1 * time.Hour)
		resp, err := ec2conn.DescribeSpotPriceHistory(&ec2.DescribeSpotPriceHistoryInput{
			InstanceTypes:       []*string{&s.InstanceType},
			ProductDescriptions: []*string{&s.SpotPriceProduct},
			AvailabilityZone:    &az,
			StartTime:           &startTime,
		})
		if err != nil {
			return "", fmt.Errorf("Error finding spot price: %s", err)
		}

		var price float64
		for _, history := range resp.SpotPriceHistory {
			log.Printf("[INFO] Candidate spot price: %s", *history.SpotPrice)
			current, err := strconv.ParseFloat(*history.SpotPrice, 64)
			if err != nil {
				log.Printf("[ERR] Error parsing spot price: %s", err)
				continue
			}
			if price == 0 || current < price {
				price = current
				if az == "" {
					az = *history.AvailabilityZone
				}
			}
		}
		if price == 0 {
			return "", fmt.Errorf("No candidate spot prices found!")
		} else {
			// Add 0.5 cents to minimum spot bid to ensure capacity will be available
			// Avoids price-too-low error in active markets which can fluctuate
			price = price + 0.005
		}

		spotPrice = strconv.FormatFloat(price, 'f', -1, 64)
	}

	s.SpotPrice = spotPrice

	return spotPrice, nil

}

func (s *StepRunSpotInstance) CreateTemplateData(userData *string, az string,
	state multistep.StateBag, marketOptions *ec2.LaunchTemplateInstanceMarketOptionsRequest) *ec2.RequestLaunchTemplateData {
	// Convert the BlockDeviceMapping into a
	// LaunchTemplateBlockDeviceMappingRequest. These structs are identical,
	// except for the EBS field -- on one, that field contains a
	// LaunchTemplateEbsBlockDeviceRequest, and on the other, it contains an
	// EbsBlockDevice. The EbsBlockDevice and
	// LaunchTemplateEbsBlockDeviceRequest structs are themselves
	// identical except for the struct's name, so you can cast one directly
	// into the other.
	blockDeviceMappings := s.BlockDevices.BuildLaunchDevices()
	var launchMappingRequests []*ec2.LaunchTemplateBlockDeviceMappingRequest
	for _, mapping := range blockDeviceMappings {
		launchRequest := &ec2.LaunchTemplateBlockDeviceMappingRequest{
			DeviceName:  mapping.DeviceName,
			Ebs:         (*ec2.LaunchTemplateEbsBlockDeviceRequest)(mapping.Ebs),
			NoDevice:    mapping.NoDevice,
			VirtualName: mapping.VirtualName,
		}
		launchMappingRequests = append(launchMappingRequests, launchRequest)
	}

	// Create a launch template.
	templateData := ec2.RequestLaunchTemplateData{
		BlockDeviceMappings:   launchMappingRequests,
		DisableApiTermination: aws.Bool(false),
		EbsOptimized:          &s.EbsOptimized,
		IamInstanceProfile:    &ec2.LaunchTemplateIamInstanceProfileSpecificationRequest{Name: &s.IamInstanceProfile},
		ImageId:               &s.SourceAMI,
		InstanceMarketOptions: marketOptions,
		Placement: &ec2.LaunchTemplatePlacementRequest{
			AvailabilityZone: &az,
		},
		UserData: userData,
	}
	// Create a network interface
	securityGroupIds := aws.StringSlice(state.Get("securityGroupIds").([]string))
	subnetId := state.Get("subnet_id").(string)

	if subnetId != "" {
		// Set up a full network interface
		networkInterface := ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
			Groups:              securityGroupIds,
			DeleteOnTermination: aws.Bool(true),
			DeviceIndex:         aws.Int64(0),
			SubnetId:            aws.String(subnetId),
		}
		if s.AssociatePublicIpAddress {
			networkInterface.SetAssociatePublicIpAddress(s.AssociatePublicIpAddress)
		}
		templateData.SetNetworkInterfaces([]*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{&networkInterface})
	} else {
		templateData.SetSecurityGroupIds(securityGroupIds)

	}

	// If instance type is not set, we'll just pick the lowest priced instance
	// available.
	if s.InstanceType != "" {
		templateData.SetInstanceType(s.InstanceType)
	}

	if s.Comm.SSHKeyPairName != "" {
		templateData.SetKeyName(s.Comm.SSHKeyPairName)
	}

	return &templateData
}

func (s *StepRunSpotInstance) LoadUserData() (string, error) {
	userData := s.UserData
	if s.UserDataFile != "" {
		contents, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", fmt.Errorf("Problem reading user data file: %s", err)
		}

		userData = string(contents)
	}

	// Test if it is encoded already, and if not, encode it
	if _, err := base64.StdEncoding.DecodeString(userData); err != nil {
		log.Printf("[DEBUG] base64 encoding user data...")
		userData = base64.StdEncoding.EncodeToString([]byte(userData))
	}
	return userData, nil
}

func (s *StepRunSpotInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Launching a spot AWS instance...")

	// Get and validate the source AMI
	image, ok := state.Get("source_image").(*ec2.Image)
	if !ok {
		state.Put("error", fmt.Errorf("source_image type assertion failed"))
		return multistep.ActionHalt
	}
	s.SourceAMI = *image.ImageId

	if s.ExpectedRootDevice != "" && *image.RootDeviceType != s.ExpectedRootDevice {
		state.Put("error", fmt.Errorf(
			"The provided source AMI has an invalid root device type.\n"+
				"Expected '%s', got '%s'.",
			s.ExpectedRootDevice, *image.RootDeviceType))
		return multistep.ActionHalt
	}

	azConfig := ""
	if azRaw, ok := state.GetOk("availability_zone"); ok {
		azConfig = azRaw.(string)
	}
	az := azConfig

	ui.Message(fmt.Sprintf("Finding spot price for %s %s...",
		s.SpotPriceProduct, s.InstanceType))
	spotPrice, err := s.CalculateSpotPrice(az, ec2conn)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Determined spot instance price of: %s.", spotPrice))

	var instanceId string

	ui.Say("Interpolating tags for spot instance...")
	// s.Tags will tag the eventually launched instance
	// s.SpotTags apply to the spot request itself, and do not automatically
	// get applied to the spot instance that is launched once the request is
	// fulfilled
	if _, exists := s.Tags["Name"]; !exists {
		s.Tags["Name"] = "Packer Builder"
	}

	// Convert tags from the tag map provided by the user into *ec2.Tag s
	ec2Tags, err := s.Tags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error generating tags for source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// This prints the tags to the ui; it doesn't actually add them to the
	// instance yet
	ec2Tags.Report(ui)

	spotOptions := ec2.LaunchTemplateSpotMarketOptionsRequest{
		MaxPrice: &s.SpotPrice,
	}
	if s.BlockDurationMinutes != 0 {
		spotOptions.BlockDurationMinutes = &s.BlockDurationMinutes
	}
	marketOptions := &ec2.LaunchTemplateInstanceMarketOptionsRequest{
		SpotOptions: &spotOptions,
	}
	marketOptions.SetMarketType(ec2.MarketTypeSpot)

	// Create a launch template for the instance
	ui.Message("Loading User Data File...")

	// Generate a random name to avoid conflicting with other
	// instances of packer running in this AWS account
	launchTemplateName := fmt.Sprintf(
		"packer-fleet-launch-template-%s",
		random.AlphaNum(7))
	state.Put("launchTemplateName", launchTemplateName) // For the cleanup step

	userData, err := s.LoadUserData()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	ui.Message("Creating Spot Fleet launch template...")
	templateData := s.CreateTemplateData(&userData, az, state, marketOptions)
	launchTemplate := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateData: templateData,
		LaunchTemplateName: aws.String(launchTemplateName),
		VersionDescription: aws.String("template generated by packer for launching spot instances"),
	}

	// Tell EC2 to create the template
	_, err = ec2conn.CreateLaunchTemplate(launchTemplate)
	if err != nil {
		err := fmt.Errorf("Error creating launch template for spot instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Add overrides for each user-provided instance type
	var overrides []*ec2.FleetLaunchTemplateOverridesRequest
	for _, instanceType := range s.SpotInstanceTypes {
		override := ec2.FleetLaunchTemplateOverridesRequest{
			InstanceType: aws.String(instanceType),
		}
		overrides = append(overrides, &override)
	}

	createFleetInput := &ec2.CreateFleetInput{
		LaunchTemplateConfigs: []*ec2.FleetLaunchTemplateConfigRequest{
			{
				LaunchTemplateSpecification: &ec2.FleetLaunchTemplateSpecificationRequest{
					LaunchTemplateName: aws.String(launchTemplateName),
					Version:            aws.String("1"),
				},
				Overrides: overrides,
			},
		},
		ReplaceUnhealthyInstances: aws.Bool(false),
		TargetCapacitySpecification: &ec2.TargetCapacitySpecificationRequest{
			TotalTargetCapacity:       aws.Int64(1),
			DefaultTargetCapacityType: aws.String("spot"),
		},
		Type: aws.String("instant"),
	}

	// Create the request for the spot instance.
	req, createOutput := ec2conn.CreateFleetRequest(createFleetInput)
	ui.Message(fmt.Sprintf("Sending spot request (%s)...", req.RequestID))

	// Tag the spot instance request (not the eventual spot instance)
	spotTags, err := s.SpotTags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error generating tags for spot request: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	spotTags.Report(ui)

	if len(spotTags) > 0 && s.SpotTags.IsSet() {
		err = retry.Config{
			Tries:       11,
			ShouldRetry: func(error) bool { return false },
			RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
				Tags:      spotTags,
				Resources: []*string{aws.String(req.RequestID)},
			})
			return err
		})
		if err != nil {
			err := fmt.Errorf("Error tagging spot request: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Actually send the spot connection request.
	err = req.Send()
	if err != nil {
		err := fmt.Errorf("Error waiting for spot request (%s) to become ready: %s", req.RequestID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(createOutput.Errors) > 0 {
		err := fmt.Errorf("error sending spot request: %s", *createOutput.Errors[0])
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	instanceId = *createOutput.Instances[0].InstanceIds[0]

	// Set the instance ID so that the cleanup works properly
	s.instanceId = instanceId

	ui.Message(fmt.Sprintf("Instance ID: %s", instanceId))

	r, err := ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	})
	if err != nil || len(r.Reservations) == 0 || len(r.Reservations[0].Instances) == 0 {
		err := fmt.Errorf("Error finding source instance.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	instance := r.Reservations[0].Instances[0]

	// Retry creating tags for about 2.5 minutes
	err = retry.Config{
		Tries: 11,
		ShouldRetry: func(error) bool {
			if awsErr, ok := err.(awserr.Error); ok {
				switch awsErr.Code() {
				case "InvalidInstanceID.NotFound":
					return true
				}
			}
			return false
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
			Tags:      ec2Tags,
			Resources: []*string{instance.InstanceId},
		})
		return err
	})

	if err != nil {
		err := fmt.Errorf("Error tagging source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	volumeIds := make([]*string, 0)
	for _, v := range instance.BlockDeviceMappings {
		if ebs := v.Ebs; ebs != nil {
			volumeIds = append(volumeIds, ebs.VolumeId)
		}
	}

	if len(volumeIds) > 0 && s.VolumeTags.IsSet() {
		ui.Say("Adding tags to source EBS Volumes")

		volumeTags, err := s.VolumeTags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
		if err != nil {
			err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		volumeTags.Report(ui)

		_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
			Resources: volumeIds,
			Tags:      volumeTags,
		})

		if err != nil {
			err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	}

	if s.Debug {
		if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
			ui.Message(fmt.Sprintf("Public DNS: %s", *instance.PublicDnsName))
		}

		if instance.PublicIpAddress != nil && *instance.PublicIpAddress != "" {
			ui.Message(fmt.Sprintf("Public IP: %s", *instance.PublicIpAddress))
		}

		if instance.PrivateIpAddress != nil && *instance.PrivateIpAddress != "" {
			ui.Message(fmt.Sprintf("Private IP: %s", *instance.PrivateIpAddress))
		}
	}

	state.Put("instance", instance)

	return multistep.ActionContinue
}

func (s *StepRunSpotInstance) Cleanup(state multistep.StateBag) {

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	launchTemplateName := state.Get("launchTemplateName").(string)

	// Cancel the spot request if it exists
	if s.spotRequest != nil {
		ui.Say("Cancelling the spot request...")
		input := &ec2.CancelSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{s.spotRequest.SpotInstanceRequestId},
		}
		if _, err := ec2conn.CancelSpotInstanceRequests(input); err != nil {
			ui.Error(fmt.Sprintf("Error cancelling the spot request, may still be around: %s", err))
			return
		}

		err := WaitUntilSpotRequestFulfilled(aws.BackgroundContext(), ec2conn, *s.spotRequest.SpotInstanceRequestId)
		if err != nil {
			ui.Error(err.Error())
		}

	}

	// Terminate the source instance if it exists
	if s.instanceId != "" {
		ui.Say("Terminating the source AWS instance...")
		if _, err := ec2conn.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: []*string{&s.instanceId}}); err != nil {
			ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
			return
		}

		if err := WaitUntilInstanceTerminated(aws.BackgroundContext(), ec2conn, s.instanceId); err != nil {
			ui.Error(err.Error())
		}
	}

	// Delete the launch template used to create the spot fleet
	deleteInput := &ec2.DeleteLaunchTemplateInput{
		LaunchTemplateName: aws.String(launchTemplateName),
	}
	if _, err := ec2conn.DeleteLaunchTemplate(deleteInput); err != nil {
		ui.Error(err.Error())
	}
}
