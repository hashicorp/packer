package common

import (
	"reflect"
	"testing"

	"github.com/outscale/osc-go/oapi"
)

func TestBlockDevice_LaunchDevices(t *testing.T) {
	tr := new(bool)
	f := new(bool)

	*tr = true
	*f = false

	cases := []struct {
		Config *BlockDevice
		Result oapi.BlockDeviceMappingVmCreation
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					SnapshotId:         "snap-1234",
					VolumeType:         "standard",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				VolumeSize: 8,
			},

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeSize:         8,
					DeleteOnVmDeletion: f,
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

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "io1",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
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

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "gp2",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
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

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "gp2",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "standard",
				DeleteOnVmDeletion: true,
			},

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "standard",
					DeleteOnVmDeletion: tr,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:  "/dev/sdb",
				VirtualName: "ephemeral0",
			},

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName:        "/dev/sdb",
				VirtualDeviceName: "ephemeral0",
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				NoDevice:   true,
			},

			Result: oapi.BlockDeviceMappingVmCreation{
				DeviceName: "/dev/sdb",
				NoDevice:   "",
			},
		},
	}

	for _, tc := range cases {

		launchBlockDevices := LaunchBlockDevices{
			LaunchMappings: []BlockDevice{*tc.Config},
		}

		expected := []oapi.BlockDeviceMappingVmCreation{tc.Result}

		launchResults := launchBlockDevices.BuildLaunchDevices()
		if !reflect.DeepEqual(expected, launchResults) {
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected, launchResults)
		}
	}
}

func TestBlockDevice_OMI(t *testing.T) {
	tr := new(bool)
	f := new(bool)

	*tr = true
	*f = false

	cases := []struct {
		Config *BlockDevice
		Result oapi.BlockDeviceMappingImage
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oapi.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					SnapshotId:         "snap-1234",
					VolumeType:         "standard",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				VolumeSize: 8,
			},

			Result: oapi.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeSize:         8,
					DeleteOnVmDeletion: f,
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

			Result: oapi.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "io1",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
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

			Result: oapi.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "gp2",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
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

			Result: oapi.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "gp2",
					VolumeSize:         8,
					DeleteOnVmDeletion: tr,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "standard",
				DeleteOnVmDeletion: true,
			},

			Result: oapi.BlockDeviceMappingImage{
				DeviceName: "/dev/sdb",
				Bsu: oapi.BsuToCreate{
					VolumeType:         "standard",
					DeleteOnVmDeletion: tr,
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:  "/dev/sdb",
				VirtualName: "ephemeral0",
			},

			Result: oapi.BlockDeviceMappingImage{
				DeviceName:        "/dev/sdb",
				VirtualDeviceName: "ephemeral0",
			},
		},
	}

	for _, tc := range cases {
		omiBlockDevices := OMIBlockDevices{
			OMIMappings: []BlockDevice{*tc.Config},
		}

		expected := []oapi.BlockDeviceMappingImage{tc.Result}

		omiResults := omiBlockDevices.BuildOMIDevices()
		if !reflect.DeepEqual(expected, omiResults) {
			t.Fatalf("Bad block device, \nexpected: %+#v\n\ngot: %+#v",
				expected, omiResults)
		}
	}
}
