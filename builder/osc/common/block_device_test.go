package common

import (
	"reflect"
	"testing"

	"github.com/outscale/osc-go/oapi"
)

func TestBlockDevice(t *testing.T) {
	cases := []struct {
		Config *BlockDevice
		Result *oapi.BlockDeviceMapping
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				Bsu: oapi.Bsu{
					SnapshotId:         "snap-1234",
					VolumeType:         "standard",
					VolumeSize:         8,
					DeleteOnVmDeletion: true,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				VolumeSize: 8,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				Bsu: oapi.Bsu{
					VolumeSize:         8,
					DeleteOnVmDeletion: false,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "io1",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
				IOPS:               1000,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				Bsu: oapi.Bsu{
					VolumeType:         "io1",
					VolumeSize:         8,
					DeleteOnVmDeletion: true,
					Iops:               1000,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "gp2",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				Bsu: oapi.Bsu{
					VolumeType:         "gp2",
					VolumeSize:         8,
					DeleteOnVmDeletion: true,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "gp2",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				Bsu: oapi.Bsu{
					VolumeType:         "gp2",
					VolumeSize:         8,
					DeleteOnVmDeletion: true,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "standard",
				DeleteOnVmDeletion: true,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				Bsu: oapi.Bsu{
					VolumeType:         "standard",
					DeleteOnVmDeletion: true,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:  "/dev/sdb",
				VirtualName: "ephemeral0",
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName:        "/dev/sdb",
				VirtualDeviceName: "ephemeral0",
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				NoDevice:   true,
			},

			Result: &oapi.BlockDeviceMapping{
				DeviceName: "/dev/sdb",
				NoDevice:   "",
			},
		},
	}

	for _, tc := range cases {
		omiBlockDevices := OMIBlockDevices{
			OMIMappings: []BlockDevice{*tc.Config},
		}

		launchBlockDevices := LaunchBlockDevices{
			LaunchMappings: []BlockDevice{*tc.Config},
		}

		expected := []*oapi.BlockDeviceMapping{tc.Result}

		omiResults := omiBlockDevices.BuildOMIDevices()
		if !reflect.DeepEqual(expected, omiResults) {
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected, omiResults)
		}

		launchResults := launchBlockDevices.BuildLaunchDevices()
		if !reflect.DeepEqual(expected, launchResults) {
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected, launchResults)
		}
	}
}
