package common

import (
	"github.com/mitchellh/goamz/ec2"
	"reflect"
	"testing"
)

func TestBlockDevice(t *testing.T) {
	ec2Mapping := []ec2.BlockDeviceMapping{
		ec2.BlockDeviceMapping{
			DeviceName:          "/dev/sdb",
			VirtualName:         "ephemeral0",
			SnapshotId:          "snap-1234",
			VolumeType:          "standard",
			VolumeSize:          8,
			DeleteOnTermination: true,
			IOPS:                1000,
		},
	}

	blockDevice := BlockDevice{
		DeviceName:          "/dev/sdb",
		VirtualName:         "ephemeral0",
		SnapshotId:          "snap-1234",
		VolumeType:          "standard",
		VolumeSize:          8,
		DeleteOnTermination: true,
		IOPS:                1000,
	}

	blockDevices := BlockDevices{
		AMIMappings:    []BlockDevice{blockDevice},
		LaunchMappings: []BlockDevice{blockDevice},
	}

	if !reflect.DeepEqual(ec2Mapping, blockDevices.BuildAMIDevices()) {
		t.Fatalf("bad: %#v", ec2Mapping)
	}

	if !reflect.DeepEqual(ec2Mapping, blockDevices.BuildLaunchDevices()) {
		t.Fatalf("bad: %#v", ec2Mapping)
	}
}
