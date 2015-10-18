package common

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
				SnapshotId:          "snap-1234",
				VolumeType:          "standard",
				VolumeSize:          8,
				DeleteOnTermination: true,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sdb"),
				Ebs: &ec2.EbsBlockDevice{
					SnapshotId:          aws.String("snap-1234"),
					VolumeType:          aws.String("standard"),
					VolumeSize:          aws.Int64(8),
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				VolumeSize: 8,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sdb"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize:          aws.Int64(8),
					DeleteOnTermination: aws.Bool(false),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:          "/dev/sdb",
				VolumeType:          "io1",
				VolumeSize:          8,
				DeleteOnTermination: true,
				IOPS:                1000,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sdb"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeType:          aws.String("io1"),
					VolumeSize:          aws.Int64(8),
					DeleteOnTermination: aws.Bool(true),
					Iops:                aws.Int64(1000),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:          "/dev/sdb",
				VolumeType:          "gp2",
				VolumeSize:          8,
				DeleteOnTermination: true,
				Encrypted:           true,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sdb"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeType:          aws.String("gp2"),
					VolumeSize:          aws.Int64(8),
					DeleteOnTermination: aws.Bool(true),
					Encrypted:           aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:          "/dev/sdb",
				VolumeType:          "standard",
				DeleteOnTermination: true,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sdb"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeType:          aws.String("standard"),
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:  "/dev/sdb",
				VirtualName: "ephemeral0",
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName:  aws.String("/dev/sdb"),
				VirtualName: aws.String("ephemeral0"),
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				NoDevice:   true,
			},

			Result: &ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sdb"),
				NoDevice:   aws.String(""),
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
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected, got)
		}

		if !reflect.DeepEqual(expected, blockDevices.BuildLaunchDevices()) {
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected,
				blockDevices.BuildLaunchDevices())
		}
	}
}
