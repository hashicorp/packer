package common

import (
	"bytes"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// Define a mock struct to be used in unit tests for common aws steps.
type mockEC2ConnSpot struct {
	ec2iface.EC2API
	Config *aws.Config

	// Counters to figure out what code path was taken
	describeSpotPriceHistoryCount int
}

// Generates fake SpotPriceHistory data and returns it in the expected output
// format. Also increments a
func (m *mockEC2ConnSpot) DescribeSpotPriceHistory(copyInput *ec2.DescribeSpotPriceHistoryInput) (*ec2.DescribeSpotPriceHistoryOutput, error) {
	m.describeSpotPriceHistoryCount++
	testTime := time.Now().Add(-1 * time.Hour)
	sp := []*ec2.SpotPrice{
		{
			AvailabilityZone:   aws.String("us-east-1c"),
			InstanceType:       aws.String("t2.micro"),
			ProductDescription: aws.String("Linux/UNIX"),
			SpotPrice:          aws.String("0.003500"),
			Timestamp:          &testTime,
		},
		{
			AvailabilityZone:   aws.String("us-east-1f"),
			InstanceType:       aws.String("t2.micro"),
			ProductDescription: aws.String("Linux/UNIX"),
			SpotPrice:          aws.String("0.003500"),
			Timestamp:          &testTime,
		},
		{
			AvailabilityZone:   aws.String("us-east-1b"),
			InstanceType:       aws.String("t2.micro"),
			ProductDescription: aws.String("Linux/UNIX"),
			SpotPrice:          aws.String("0.003500"),
			Timestamp:          &testTime,
		},
	}
	output := &ec2.DescribeSpotPriceHistoryOutput{SpotPriceHistory: sp}

	return output, nil

}

func getMockConnSpot() ec2iface.EC2API {
	mockConn := &mockEC2ConnSpot{
		Config: aws.NewConfig(),
	}

	return mockConn
}

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
		SpotTags:                          TagMap(nil),
		Tags:                              TagMap{},
		VolumeTags:                        TagMap(nil),
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

	// Rerun, this time testing that we set security group IDs
	state.Put("subnet_id", "")
	template = stepRunSpotInstance.CreateTemplateData(aws.String("userdata"), "az", state,
		&ec2.LaunchTemplateInstanceMarketOptionsRequest{})
	if template.NetworkInterfaces != nil {
		t.Fatalf("Template shouldn't contain network interfaces object if subnet_id is unset.")
	}
}
