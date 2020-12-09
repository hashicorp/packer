package qemu

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"
)

func Test_buildCreateCommand(t *testing.T) {
	type testCase struct {
		Step     *stepCreateDisk
		I        int
		Expected []string
		Reason   string
	}
	testcases := []testCase{
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: false,
			},
			0,
			[]string{"create", "-f", "qcow2", "target.qcow2", "1234M"},
			"Basic, happy path, no backing store, no extra args",
		},
		{
			&stepCreateDisk{
				Format:             "qcow2",
				DiskImage:          true,
				UseBackingFile:     true,
				AdditionalDiskSize: []string{"1M", "2M"},
			},
			0,
			[]string{"create", "-f", "qcow2", "-b", "source.qcow2", "target.qcow2", "1234M"},
			"Basic, happy path, backing store, additional disks",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
				DiskImage:      true,
			},
			1,
			[]string{"create", "-f", "qcow2", "target.qcow2", "1234M"},
			"Basic, happy path, backing store set but not at first index, no extra args",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
				DiskImage:      true,
				QemuImgArgs: QemuImgArgs{
					Create: []string{"-foo", "bar"},
				},
			},
			0,
			[]string{"create", "-f", "qcow2", "-b", "source.qcow2", "-foo", "bar", "target.qcow2", "1234M"},
			"Basic, happy path, backing store set, extra args",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
				QemuImgArgs: QemuImgArgs{
					Create: []string{"-foo", "bar"},
				},
			},
			1,
			[]string{"create", "-f", "qcow2", "-foo", "bar", "target.qcow2", "1234M"},
			"Basic, happy path, backing store set but not at first index, extra args",
		},
	}

	for _, tc := range testcases {
		state := new(multistep.BasicStateBag)
		state.Put("iso_path", "source.qcow2")
		command := tc.Step.buildCreateCommand("target.qcow2", "1234M", tc.I, state)

		assert.Equal(t, command, tc.Expected,
			fmt.Sprintf("%s. Expected %#v", tc.Reason, tc.Expected))
	}
}

func Test_StepCreateCalled(t *testing.T) {
	type testCase struct {
		Step     *stepCreateDisk
		Expected []string
		Reason   string
	}
	testcases := []testCase{
		{
			&stepCreateDisk{
				Format:         "qcow2",
				DiskImage:      true,
				DiskSize:       "1M",
				VMName:         "target",
				UseBackingFile: true,
			},
			[]string{
				"create", "-f", "qcow2", "-b", "source.qcow2", "target", "1M",
			},
			"Basic, happy path, backing store, no additional disks",
		},
		{
			&stepCreateDisk{
				Format:         "raw",
				DiskImage:      false,
				DiskSize:       "4M",
				VMName:         "target",
				UseBackingFile: false,
			},
			[]string{
				"create", "-f", "raw", "target", "4M",
			},
			"Basic, happy path, raw, no additional disks",
		},
		{
			&stepCreateDisk{
				Format:             "qcow2",
				DiskImage:          true,
				DiskSize:           "4M",
				VMName:             "target",
				UseBackingFile:     false,
				AdditionalDiskSize: []string{"3M", "8M"},
			},
			[]string{
				"create", "-f", "qcow2", "target-1", "3M",
				"create", "-f", "qcow2", "target-2", "8M",
			},
			"Skips disk creation when disk can be copied",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				DiskImage:      true,
				DiskSize:       "4M",
				VMName:         "target",
				UseBackingFile: false,
			},
			nil,
			"Skips disk creation when disk can be copied",
		},
		{
			&stepCreateDisk{
				Format:             "qcow2",
				DiskImage:          true,
				DiskSize:           "1M",
				VMName:             "target",
				UseBackingFile:     true,
				AdditionalDiskSize: []string{"3M", "8M"},
			},
			[]string{
				"create", "-f", "qcow2", "-b", "source.qcow2", "target", "1M",
				"create", "-f", "qcow2", "target-1", "3M",
				"create", "-f", "qcow2", "target-2", "8M",
			},
			"Basic, happy path, backing store, additional disks",
		},
	}

	for _, tc := range testcases {
		d := new(DriverMock)
		state := copyTestState(t, d)
		state.Put("iso_path", "source.qcow2")
		action := tc.Step.Run(context.TODO(), state)
		if action != multistep.ActionContinue {
			t.Fatalf("Should have gotten an ActionContinue")
		}

		assert.Equal(t, d.QemuImgCalls, tc.Expected,
			fmt.Sprintf("%s. Expected %#v", tc.Reason, tc.Expected))
	}
}
