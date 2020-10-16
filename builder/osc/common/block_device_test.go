package common

import (
	"reflect"
	"testing"

	"github.com/outscale/osc-sdk-go/osc"
)

func TestBlockDevice_LaunchDevices(t *testing.T) {
	cases := []struct {
		Config *BlockDevice
		Result osc.BlockDeviceMappingVmCreation
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName:        "/dev/sdb",
				VirtualDeviceName: "ephemeral0",
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				NoDevice:   true,
			},

			Result: osc.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				NoDevice:   "",
			},
		},
	}

	for _, tc := range cases {

		launchBlockDevices := LaunchBlockDevices{
			LaunchMappings: []BlockDevice{*tc.Config},
		}

		expected := []osc.BlockDeviceMappingVmCreation{tc.Result}

		launchResults := launchBlockDevices.BuildOSCLaunchDevices()
		if !reflect.DeepEqual(expected, launchResults) {
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected, launchResults)
		}
	}
}

func TestBlockDevice_OMI(t *testing.T) {
	cases := []struct {
		Config *BlockDevice
		Result osc.BlockDeviceMappingImage
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: osc.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
					SnapshotId:         "snap-1234",
					VolumeType:         "standard",
					VolumeSize:         8,
					DeleteOnVmDeletion: true,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: osc.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
					VolumeSize:         8,
					DeleteOnVmDeletion: true,
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

			Result: osc.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: osc.BsuToCreate{
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

			Result: osc.BlockDeviceMappingImage{
				DeviceName:        "/dev/sdb",
				VirtualDeviceName: "ephemeral0",
			},
		},
	}

	for i, tc := range cases {
		omiBlockDevices := OMIBlockDevices{
			OMIMappings: []BlockDevice{*tc.Config},
		}

		expected := []osc.BlockDeviceMappingImage{tc.Result}

		omiResults := omiBlockDevices.BuildOscOMIDevices()
		if !reflect.DeepEqual(expected, omiResults) {
			t.Fatalf("%d - Bad block device, \nexpected: %+#v\n\ngot: %+#v",
				i, expected, omiResults)
		}
	}
}
