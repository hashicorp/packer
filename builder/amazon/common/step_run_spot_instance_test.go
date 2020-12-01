package common

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// Create statebag for running test
func tStateSpot() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("availability_zone", "us-east-1c")
	state.Put("securityGroupIds", []string{"sg-0b8984db72f213dc3"})
	state.Put("iamInstanceProfile", "packer-123")
	state.Put("subnet_id", "subnet-077fde4e")
	state.Put("source_image", "")
	return state
}

func getBasicStep() *StepRunSpotInstance {
	stepRunSpotInstance := StepRunSpotInstance{
		PollingConfig:            new(AWSPollingConfig),
		AssociatePublicIpAddress: false,
		LaunchMappings:           BlockDevices{},
		BlockDurationMinutes:     0,
		Debug:                    false,
		Comm: &communicator.Config{
			SSH: communicator.SSH{
				SSHKeyPairName: "foo",
			},
		},
		EbsOptimized:                      false,
		ExpectedRootDevice:                "ebs",
		InstanceInitiatedShutdownBehavior: "stop",
		InstanceType:                      "t2.micro",
		Region:                            "us-east-1",
		SourceAMI:                         "",
		SpotPrice:                         "auto",
		SpotTags:                          nil,
		Tags:                              map[string]string{},
		VolumeTags:                        nil,
		UserData:                          "",
		UserDataFile:                      "",
	}

	return &stepRunSpotInstance
}

func TestCreateTemplateData(t *testing.T) {
	state := tStateSpot()
	stepRunSpotInstance := getBasicStep()
	template := stepRunSpotInstance.CreateTemplateData(aws.String("userdata"), "az", state,
		&ec2.LaunchTemplateInstanceMarketOptionsRequest{})

	// expected := []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
	// 	&ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
	// 		DeleteOnTermination: aws.Bool(true),
	// 		DeviceIndex:         aws.Int64(0),
	// 		Groups:              aws.StringSlice([]string{"sg-0b8984db72f213dc3"}),
	// 		SubnetId:            aws.String("subnet-077fde4e"),
	// 	},
	// }
	// if expected != template.NetworkInterfaces {
	if template.NetworkInterfaces == nil {
		t.Fatalf("Template should have contained a networkInterface object: recieved %#v", template.NetworkInterfaces)
	}

	if *template.IamInstanceProfile.Name != state.Get("iamInstanceProfile") {
		t.Fatalf("Template should have contained a InstanceProfile name: recieved %#v", template.IamInstanceProfile.Name)
	}

	// Rerun, this time testing that we set security group IDs
	state.Put("subnet_id", "")
	template = stepRunSpotInstance.CreateTemplateData(aws.String("userdata"), "az", state,
		&ec2.LaunchTemplateInstanceMarketOptionsRequest{})
	if template.NetworkInterfaces != nil {
		t.Fatalf("Template shouldn't contain network interfaces object if subnet_id is unset.")
	}

	// Rerun, this time testing that instance doesn't have instance profile is iamInstanceProfile is unset
	state.Put("iamInstanceProfile", "")
	template = stepRunSpotInstance.CreateTemplateData(aws.String("userdata"), "az", state,
		&ec2.LaunchTemplateInstanceMarketOptionsRequest{})
	fmt.Println(template.IamInstanceProfile)
	if *template.IamInstanceProfile.Name != "" {
		t.Fatalf("Template shouldn't contain instance profile if iamInstanceProfile is unset.")
	}
}

func TestCreateTemplateData_NoEphemeral(t *testing.T) {
	state := tStateSpot()
	stepRunSpotInstance := getBasicStep()
	stepRunSpotInstance.NoEphemeral = true
	template := stepRunSpotInstance.CreateTemplateData(aws.String("userdata"), "az", state,
		&ec2.LaunchTemplateInstanceMarketOptionsRequest{})
	if len(template.BlockDeviceMappings) != 26 {
		t.Fatalf("Should have created 26 mappings to keep ephemeral drives from appearing.")
	}

	// Now check that noEphemeral doesn't mess with the mappings in real life.
	// state = tStateSpot()
	// stepRunSpotInstance = getBasicStep()
	// stepRunSpotInstance.NoEphemeral = true
	// mappings := []*ec2.InstanceBlockDeviceMapping{
	// 	&ec2.InstanceBlockDeviceMapping{
	// 		DeviceName: "xvda",
	// 		Ebs: {
	// 			DeleteOnTermination: true,
	// 			Status:              "attaching",
	// 			VolumeId:            "vol-044cd49c330f21c05",
	// 		},
	// 	},
	// 	&ec2.InstanceBlockDeviceMapping{
	// 		DeviceName: "/dev/xvdf",
	// 		Ebs: {
	// 			DeleteOnTermination: false,
	// 			Status:              "attaching",
	// 			VolumeId:            "vol-0eefaf2d6ae35827e",
	// 		},
	// 	},
	// }
	// template = stepRunSpotInstance.CreateTemplateData(aws.String("userdata"), "az", state,
	// 	&ec2.LaunchTemplateInstanceMarketOptionsRequest{})
	// if len(*template.BlockDeviceMappings) != 26 {
	// 	t.Fatalf("Should have created 26 mappings to keep ephemeral drives from appearing.")
	// }
}

type runSpotEC2ConnMock struct {
	ec2iface.EC2API

	CreateLaunchTemplateParams []*ec2.CreateLaunchTemplateInput
	CreateLaunchTemplateFn     func(*ec2.CreateLaunchTemplateInput) (*ec2.CreateLaunchTemplateOutput, error)

	CreateFleetParams []*ec2.CreateFleetInput
	CreateFleetFn     func(*ec2.CreateFleetInput) (*ec2.CreateFleetOutput, error)

	CreateTagsParams []*ec2.CreateTagsInput
	CreateTagsFn     func(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)

	DescribeInstancesParams []*ec2.DescribeInstancesInput
	DescribeInstancesFn     func(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

func (m *runSpotEC2ConnMock) CreateLaunchTemplate(req *ec2.CreateLaunchTemplateInput) (*ec2.CreateLaunchTemplateOutput, error) {
	m.CreateLaunchTemplateParams = append(m.CreateLaunchTemplateParams, req)
	resp, err := m.CreateLaunchTemplateFn(req)
	return resp, err
}

func (m *runSpotEC2ConnMock) CreateFleet(req *ec2.CreateFleetInput) (*ec2.CreateFleetOutput, error) {
	m.CreateFleetParams = append(m.CreateFleetParams, req)
	if m.CreateFleetFn != nil {
		resp, err := m.CreateFleetFn(req)
		return resp, err
	} else {
		return nil, nil
	}
}

func (m *runSpotEC2ConnMock) DescribeInstances(req *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	m.DescribeInstancesParams = append(m.DescribeInstancesParams, req)
	if m.DescribeInstancesFn != nil {
		resp, err := m.DescribeInstancesFn(req)
		return resp, err
	} else {
		return nil, nil
	}
}

func (m *runSpotEC2ConnMock) CreateTags(req *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	m.CreateTagsParams = append(m.CreateTagsParams, req)
	if m.CreateTagsFn != nil {
		resp, err := m.CreateTagsFn(req)
		return resp, err
	} else {
		return nil, nil
	}
}

func defaultEc2Mock(instanceId, spotRequestId, volumeId *string) *runSpotEC2ConnMock {
	instance := &ec2.Instance{
		InstanceId:            instanceId,
		SpotInstanceRequestId: spotRequestId,
		BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
			{
				Ebs: &ec2.EbsInstanceBlockDevice{
					VolumeId: volumeId,
				},
			},
		},
	}
	return &runSpotEC2ConnMock{
		CreateLaunchTemplateFn: func(in *ec2.CreateLaunchTemplateInput) (*ec2.CreateLaunchTemplateOutput, error) {
			return &ec2.CreateLaunchTemplateOutput{
				LaunchTemplate: nil,
				Warning:        nil,
			}, nil
		},
		CreateFleetFn: func(*ec2.CreateFleetInput) (*ec2.CreateFleetOutput, error) {
			return &ec2.CreateFleetOutput{
				Errors:  nil,
				FleetId: nil,
				Instances: []*ec2.CreateFleetInstance{
					{
						InstanceIds: []*string{instanceId},
					},
				},
			}, nil
		},
		DescribeInstancesFn: func(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				NextToken: nil,
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{instance},
					},
				},
			}, nil
		},
	}
}

func TestRun(t *testing.T) {
	instanceId := aws.String("test-instance-id")
	spotRequestId := aws.String("spot-id")
	volumeId := aws.String("volume-id")
	ec2Mock := defaultEc2Mock(instanceId, spotRequestId, volumeId)

	uiMock := packer.TestUi(t)

	state := tStateSpot()
	state.Put("ec2", ec2Mock)
	state.Put("ui", uiMock)
	state.Put("source_image", testImage())

	stepRunSpotInstance := getBasicStep()
	stepRunSpotInstance.Tags["Name"] = "Packer Builder"
	stepRunSpotInstance.Tags["test-tag"] = "test-value"
	stepRunSpotInstance.SpotTags = map[string]string{
		"spot-tag": "spot-tag-value",
	}
	stepRunSpotInstance.VolumeTags = map[string]string{
		"volume-tag": "volume-tag-value",
	}

	ctx := context.TODO()
	action := stepRunSpotInstance.Run(ctx, state)

	if err := state.Get("error"); err != nil {
		t.Fatalf("should not error, but: %v", err)
	}

	if action != multistep.ActionContinue {
		t.Fatalf("shoul continue, but: %v", action)
	}

	if len(ec2Mock.CreateLaunchTemplateParams) != 1 {
		t.Fatalf("createLaunchTemplate should be invoked once, but invoked %v", len(ec2Mock.CreateLaunchTemplateParams))
	}
	launchTemplateName := ec2Mock.CreateLaunchTemplateParams[0].LaunchTemplateName

	if len(ec2Mock.CreateLaunchTemplateParams[0].TagSpecifications) != 1 {
		t.Fatalf("exactly one launch template tag specification expected")
	}
	if *ec2Mock.CreateLaunchTemplateParams[0].TagSpecifications[0].ResourceType != "launch-template" {
		t.Fatalf("resource type 'launch-template' expected")
	}
	if len(ec2Mock.CreateLaunchTemplateParams[0].TagSpecifications[0].Tags) != 1 {
		t.Fatalf("1 launch template tag expected")
	}

	nameTag := ec2Mock.CreateLaunchTemplateParams[0].TagSpecifications[0].Tags[0]
	if *nameTag.Key != "spot-tag" || *nameTag.Value != "spot-tag-value" {
		t.Fatalf("expected spot-tag: spot-tag-value")
	}

	if len(ec2Mock.CreateFleetParams) != 1 {
		t.Fatalf("createFleet should be invoked once, but invoked %v", len(ec2Mock.CreateLaunchTemplateParams))
	}
	if *ec2Mock.CreateFleetParams[0].TargetCapacitySpecification.DefaultTargetCapacityType != "spot" {
		t.Fatalf("capacity type should be spot")
	}
	if *ec2Mock.CreateFleetParams[0].TargetCapacitySpecification.TotalTargetCapacity != 1 {
		t.Fatalf("target capacity should be 1")
	}
	if len(ec2Mock.CreateFleetParams[0].LaunchTemplateConfigs) != 1 {
		t.Fatalf("exactly one launch config template expected")
	}
	if *ec2Mock.CreateFleetParams[0].LaunchTemplateConfigs[0].LaunchTemplateSpecification.LaunchTemplateName != *launchTemplateName {
		t.Fatalf("launchTemplateName should match in createLaunchTemplate and createFleet requests")
	}

	if len(ec2Mock.DescribeInstancesParams) != 1 {
		t.Fatalf("describeInstancesParams should be invoked once, but invoked %v", len(ec2Mock.DescribeInstancesParams))
	}
	if *ec2Mock.DescribeInstancesParams[0].InstanceIds[0] != *instanceId {
		t.Fatalf("instanceId should match from createFleet response")
	}

	uiMock.Say(fmt.Sprintf("%v", ec2Mock.CreateTagsParams))
	if len(ec2Mock.CreateTagsParams) != 3 {
		t.Fatalf("createTags should be invoked 3 times")
	}
	if len(ec2Mock.CreateTagsParams[0].Resources) != 1 || *ec2Mock.CreateTagsParams[0].Resources[0] != *spotRequestId {
		t.Fatalf("should create tags for spot request")
	}
	if len(ec2Mock.CreateTagsParams[1].Resources) != 1 || *ec2Mock.CreateTagsParams[1].Resources[0] != *instanceId {
		t.Fatalf("should create tags for instance")
	}
	if len(ec2Mock.CreateTagsParams[2].Resources) != 1 || ec2Mock.CreateTagsParams[2].Resources[0] != volumeId {
		t.Fatalf("should create tags for volume")
	}
}

func TestRun_NoSpotTags(t *testing.T) {
	instanceId := aws.String("test-instance-id")
	spotRequestId := aws.String("spot-id")
	volumeId := aws.String("volume-id")
	ec2Mock := defaultEc2Mock(instanceId, spotRequestId, volumeId)

	uiMock := packer.TestUi(t)

	state := tStateSpot()
	state.Put("ec2", ec2Mock)
	state.Put("ui", uiMock)
	state.Put("source_image", testImage())

	stepRunSpotInstance := getBasicStep()
	stepRunSpotInstance.Tags["Name"] = "Packer Builder"
	stepRunSpotInstance.Tags["test-tag"] = "test-value"
	stepRunSpotInstance.VolumeTags = map[string]string{
		"volume-tag": "volume-tag-value",
	}

	ctx := context.TODO()
	action := stepRunSpotInstance.Run(ctx, state)

	if err := state.Get("error"); err != nil {
		t.Fatalf("should not error, but: %v", err)
	}

	if action != multistep.ActionContinue {
		t.Fatalf("shoul continue, but: %v", action)
	}

	if len(ec2Mock.CreateLaunchTemplateParams[0].TagSpecifications) != 0 {
		t.Fatalf("0 launch template tags expected")
	}
}
