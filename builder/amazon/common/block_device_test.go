package common

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestBlockDevice(t *testing.T) {
	cases := []struct {
		Config *BlockDevice
		Result *ec2.BlockDeviceMapping
	}{
		{
			Config: &BlockDevice{
				DeviceName:          "/dev/sdb",
				VirtualName:         "ephemeral0",
				SnapshotId:          "snap-1234",
				VolumeType:          "standard",
				VolumeSize:          8,
				DeleteOnTermination: true,
				IOPS:                1000,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName:  aws.String("/dev/sdb"),
				VirtualName: aws.String("ephemeral0"),
				EBS: &ec2.EBSBlockDevice{
					SnapshotID:          aws.String("snap-1234"),
					VolumeType:          aws.String("standard"),
					VolumeSize:          aws.Long(8),
					DeleteOnTermination: aws.Boolean(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				VolumeSize: 8,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName:  aws.String("/dev/sdb"),
				VirtualName: aws.String(""),
				EBS: &ec2.EBSBlockDevice{
					VolumeType:          aws.String(""),
					VolumeSize:          aws.Long(8),
					DeleteOnTermination: aws.Boolean(false),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:          "/dev/sdb",
				VirtualName:         "ephemeral0",
				VolumeType:          "io1",
				VolumeSize:          8,
				DeleteOnTermination: true,
				IOPS:                1000,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName:  aws.String("/dev/sdb"),
				VirtualName: aws.String("ephemeral0"),
				EBS: &ec2.EBSBlockDevice{
					VolumeType:          aws.String("io1"),
					VolumeSize:          aws.Long(8),
					DeleteOnTermination: aws.Boolean(true),
					IOPS:                aws.Long(1000),
				},
			},
		},
	}

	for _, tc := range cases {
		blockDevices := BlockDevices{
			AMIMappings:    []BlockDevice{*tc.Config},
			LaunchMappings: []BlockDevice{*tc.Config},
		}

		expected := []*ec2.BlockDeviceMapping{tc.Result}
		got := blockDevices.BuildAMIDevices()
		if !reflect.DeepEqual(expected, got) {
			t.Fatalf("Bad block device, \nexpected: %s\n\ngot: %s", awsutil.StringValue(expected), awsutil.StringValue(got))
		}

		if !reflect.DeepEqual(expected, blockDevices.BuildLaunchDevices()) {
			t.Fatalf("Bad block device, \nexpected: %s\n\ngot: %s", awsutil.StringValue(expected), awsutil.StringValue(blockDevices.BuildLaunchDevices()))
		}
	}
}
