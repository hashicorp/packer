package common

import (
	"bytes"
	"strconv"
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
	state.Put("subnet_id", "subnet-077fde4e")
	state.Put("source_image", "")
	return state
}

func getBasicStep() *StepRunSpotInstance {
	stepRunSpotInstance := StepRunSpotInstance{
		AssociatePublicIpAddress: false,
		BlockDevices: BlockDevices{
			AMIBlockDevices: AMIBlockDevices{
				AMIMappings: []BlockDevice(nil),
			},
			LaunchBlockDevices: LaunchBlockDevices{
				LaunchMappings: []BlockDevice(nil),
			},
		},
		BlockDurationMinutes: 0,
		Debug:                false,
		Comm: &communicator.Config{
			SSHKeyPairName: "foo",
		},
		EbsOptimized:                      false,
		ExpectedRootDevice:                "ebs",
		IamInstanceProfile:                "",
		InstanceInitiatedShutdownBehavior: "stop",
		InstanceType:                      "t2.micro",
		SourceAMI:                         "",
		SpotPrice:                         "auto",
		SpotPriceProduct:                  "Linux/UNIX",
		SpotTags:                          TagMap(nil),
		Tags:                              TagMap{},
		VolumeTags:                        TagMap(nil),
		UserData:                          "",
		UserDataFile:                      "",
	}

	return &stepRunSpotInstance
}
func TestCalculateSpotPrice(t *testing.T) {
	stepRunSpotInstance := getBasicStep()
	// Set spot price and spot price product
	stepRunSpotInstance.SpotPrice = "auto"
	stepRunSpotInstance.SpotPriceProduct = "Linux/UNIX"
	ec2conn := getMockConnSpot()
	// state := tStateSpot()
	spotPrice, err := stepRunSpotInstance.CalculateSpotPrice("", ec2conn)
	if err != nil {
		t.Fatalf("Should not have had an error calculating spot price")
	}
	sp, _ := strconv.ParseFloat(spotPrice, 64)
	expected := 0.008500
	if sp != expected { // 0.003500 (from spot history) + .005
		t.Fatalf("Expected spot price of \"0.008500\", not %s", spotPrice)
	}
}
