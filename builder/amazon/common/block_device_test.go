package common

import (
	"github.com/mitchellh/goamz/ec2"
	"reflect"
	"testing"
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
				DeviceName:          "/dev/sdb",
				VirtualName:         "ephemeral0",
				SnapshotId:          "snap-1234",
				VolumeType:          "standard",
				VolumeSize:          8,
				DeleteOnTermination: true,
				IOPS:                1000,
			},
		},
	}

	for _, tc := range cases {
		blockDevices := BlockDevices{
			AMIMappings:    []BlockDevice{*tc.Config},
			LaunchMappings: []BlockDevice{*tc.Config},
		}

		expected := []ec2.BlockDeviceMapping{*tc.Result}

		if !reflect.DeepEqual(expected, blockDevices.BuildAMIDevices()) {
			t.Fatalf("bad: %#v", expected)
		}

		if !reflect.DeepEqual(expected, blockDevices.BuildLaunchDevices()) {
			t.Fatalf("bad: %#v", expected)
		}
	}
}
