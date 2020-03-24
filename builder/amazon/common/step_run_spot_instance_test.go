package common

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// Create statebag for running test
func tStateSpot() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
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
