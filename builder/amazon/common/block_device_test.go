package common

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/goamz/ec2"
	"testing"
)

func TestBlockDevice(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

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

	assert.Equal(ec2Mapping, blockDevices.BuildAMIDevices(), "should match output")
	assert.Equal(ec2Mapping, blockDevices.BuildLaunchDevices(), "should match output")
}
